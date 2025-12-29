// Code generation script for SQL queries.
// Parses schema/queries/*.sql files and generates:
// - internal/storage/queries.gen.go (Go query functions)
// - web/src/lib/wasm/queries.gen.ts (TypeScript query functions)
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"
)

// Query represents a parsed SQL query.
type Query struct {
	Name       string   // Function name (e.g., GetEddingtonDays)
	ReturnType string   // :one, :many, :exec
	Comment    string   // Documentation comment
	SQL        string   // The SQL query
	Params     []string // Parameter placeholders (?1, ?2, etc.)
	HasSlice   bool     // Whether query has a SLICE placeholder
	SliceName  string   // Name of the slice parameter
	File       string   // Source file name
}

// QueryFile represents a parsed query file.
type QueryFile struct {
	Name    string
	Queries []Query
}

func main() {
	// Find project root (where go.mod is)
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	queriesDir := filepath.Join(root, "schema", "queries")
	goOutput := filepath.Join(root, "internal", "storage", "queries.gen.go")
	tsOutput := filepath.Join(root, "web", "src", "lib", "wasm", "queries.gen.ts")

	// Parse all query files
	files, err := parseQueryDir(queriesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing queries: %v\n", err)
		os.Exit(1)
	}

	// Generate Go code
	if err := generateGoCode(files, goOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating Go code: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s\n", goOutput)

	// Generate TypeScript code
	if err := generateTSCode(files, tsOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating TypeScript code: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s\n", tsOutput)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod")
		}
		dir = parent
	}
}

func parseQueryDir(dir string) ([]QueryFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var files []QueryFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		queries, err := parseQueryFile(path)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		files = append(files, QueryFile{
			Name:    strings.TrimSuffix(entry.Name(), ".sql"),
			Queries: queries,
		})
	}

	// Sort for deterministic output
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	return files, nil
}

var (
	nameRe  = regexp.MustCompile(`^--\s*name:\s*(\w+)\s+(:one|:many|:exec)`)
	sliceRe = regexp.MustCompile(`/\*SLICE:(\w+)\*/`)
	paramRe = regexp.MustCompile(`\?(\d+)`)
)

func parseQueryFile(path string) ([]Query, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var queries []Query
	var current *Query
	var sqlLines []string
	var comments []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for name directive
		if matches := nameRe.FindStringSubmatch(line); matches != nil {
			// Save previous query
			if current != nil && len(sqlLines) > 0 {
				current.SQL = strings.TrimSpace(strings.Join(sqlLines, "\n"))
				current.Params = extractParams(current.SQL)
				if match := sliceRe.FindStringSubmatch(current.SQL); match != nil {
					current.HasSlice = true
					current.SliceName = match[1]
				}
				queries = append(queries, *current)
			}

			current = &Query{
				Name:       matches[1],
				ReturnType: matches[2],
				Comment:    strings.TrimSpace(strings.Join(comments, " ")),
				File:       filepath.Base(path),
			}
			sqlLines = nil
			comments = nil
			continue
		}

		// Collect comments before name directive
		if current == nil && strings.HasPrefix(strings.TrimSpace(line), "--") {
			comment := strings.TrimPrefix(strings.TrimSpace(line), "--")
			comment = strings.TrimSpace(comment)
			if comment != "" && !strings.HasPrefix(comment, "name:") {
				comments = append(comments, comment)
			}
			continue
		}

		// Collect SQL lines
		if current != nil {
			// Skip empty lines at the start
			if len(sqlLines) == 0 && strings.TrimSpace(line) == "" {
				continue
			}
			// Skip comment lines within SQL
			if strings.HasPrefix(strings.TrimSpace(line), "--") {
				continue
			}
			sqlLines = append(sqlLines, line)
		}
	}

	// Save last query
	if current != nil && len(sqlLines) > 0 {
		current.SQL = strings.TrimSpace(strings.Join(sqlLines, "\n"))
		current.Params = extractParams(current.SQL)
		if match := sliceRe.FindStringSubmatch(current.SQL); match != nil {
			current.HasSlice = true
			current.SliceName = match[1]
		}
		queries = append(queries, *current)
	}

	return queries, scanner.Err()
}

func extractParams(sql string) []string {
	matches := paramRe.FindAllStringSubmatch(sql, -1)
	seen := make(map[string]bool)
	var params []string
	for _, m := range matches {
		if !seen[m[1]] {
			seen[m[1]] = true
			params = append(params, m[1])
		}
	}
	sort.Strings(params)
	return params
}

// Go code generation

var goTemplate = template.Must(template.New("go").Parse(`// Code generated by scripts/generate-sql. DO NOT EDIT.

package storage

import (
	"database/sql"
)

// Queries provides generated query functions.
type Queries struct {
	db *sql.DB
}

// NewQueries creates a new Queries instance.
func NewQueries(db *sql.DB) *Queries {
	return &Queries{db: db}
}

{{range .Files}}
// ============================================================================
// {{.Name}} queries
// ============================================================================
{{range .Queries}}
{{if .Comment}}// {{.Name}} - {{.Comment}}{{end}}
const {{.Name}}SQL = ` + "`" + `{{.SQL}}` + "`" + `
{{end}}
{{end}}
`))

func generateGoCode(files []QueryFile, output string) error {
	var buf bytes.Buffer
	if err := goTemplate.Execute(&buf, map[string]interface{}{
		"Files": files,
	}); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	// Format the Go code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Write unformatted for debugging
		if err := os.WriteFile(output+".unformatted", buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("writing unformatted: %w", err)
		}
		return fmt.Errorf("formatting Go code: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(output, formatted, 0644)
}

// TypeScript code generation

func generateTSCode(files []QueryFile, output string) error {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by scripts/generate-sql. DO NOT EDIT.

import type { WasmDatabase } from './db'

/**
 * Generated SQL queries from schema/queries/*.sql
 */
export const queries = {
`)

	for i, file := range files {
		if i > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(fmt.Sprintf("  // ============================================================================\n"))
		buf.WriteString(fmt.Sprintf("  // %s queries\n", file.Name))
		buf.WriteString(fmt.Sprintf("  // ============================================================================\n"))

		for _, q := range file.Queries {
			// Write function
			buf.WriteString(fmt.Sprintf("\n  /**\n"))
			if q.Comment != "" {
				buf.WriteString(fmt.Sprintf("   * %s\n", q.Comment))
			}
			buf.WriteString(fmt.Sprintf("   * @generated from %s\n", q.File))
			buf.WriteString(fmt.Sprintf("   */\n"))

			// Function signature
			funcName := toLowerCamelCase(q.Name)
			buf.WriteString(fmt.Sprintf("  %s<T extends Record<string, unknown>>(db: WasmDatabase", funcName))

			// Parameters
			for _, p := range q.Params {
				buf.WriteString(fmt.Sprintf(", p%s: unknown", p))
			}
			buf.WriteString("): ")

			// Return type
			switch q.ReturnType {
			case ":one":
				buf.WriteString("T | null")
			case ":many":
				buf.WriteString("T[]")
			default:
				buf.WriteString("void")
			}

			buf.WriteString(" {\n")

			// SQL constant
			sql := escapeForTS(q.SQL)
			buf.WriteString(fmt.Sprintf("    const sql = `%s`\n", sql))

			// Params array
			if len(q.Params) > 0 {
				buf.WriteString("    const params = [")
				for i, p := range q.Params {
					if i > 0 {
						buf.WriteString(", ")
					}
					buf.WriteString(fmt.Sprintf("p%s", p))
				}
				buf.WriteString("]\n")
			}

			// Execute
			switch q.ReturnType {
			case ":one":
				if len(q.Params) > 0 {
					buf.WriteString("    return db.queryOne<T>(sql, params)\n")
				} else {
					buf.WriteString("    return db.queryOne<T>(sql)\n")
				}
			case ":many":
				if len(q.Params) > 0 {
					buf.WriteString("    return db.query<T>(sql, params)\n")
				} else {
					buf.WriteString("    return db.query<T>(sql)\n")
				}
			default:
				if len(q.Params) > 0 {
					buf.WriteString("    db.exec(sql, params)\n")
				} else {
					buf.WriteString("    db.exec(sql)\n")
				}
			}

			buf.WriteString("  },\n")
		}
	}

	buf.WriteString("}\n")

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(output, buf.Bytes(), 0644)
}

func toLowerCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func escapeForTS(s string) string {
	// Escape backticks and ${} for template literals
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "${", "\\${")
	return s
}
