// Code generation script for TypeScript types from Go WASM bridge structs.
//
// This script parses Go source files from:
//   - cmd/wasm/*.go (anonymous request structs)
//   - internal/storage/*.go (database-facing structs)
//   - internal/api/handlers/*.go (API response structs)
//
// And generates TypeScript interfaces in web/src/lib/wasm/types.gen.ts that
// match the JSON serialization format used by the WASM bridge.
//
// # Parser Limitations (regex-based, intentionally simple)
//
// This parser uses regular expressions rather than Go's AST parser for simplicity.
// The following limitations apply:
//
//   - Only fields with `json:"..."` tags are emitted
//   - The json tag must be the first (or only) struct tag
//   - Embedded/anonymous struct fields are not supported
//   - Multi-line field declarations are not supported
//   - Only basic type aliases (type X <primitive>) are handled
//   - Complex generic types are not supported
//
// If you encounter issues with specific struct definitions, consider adding
// manual type definitions in the appropriate web/src/lib/api/*.ts file.
//
// Run with: just generate-ts-types
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// GoField represents a parsed struct field
type GoField struct {
	GoName   string
	GoType   string
	JSONName string
	Optional bool // true if pointer type or has omitempty
}

// GoStruct represents a parsed Go struct
type GoStruct struct {
	Name     string
	Fields   []GoField
	FileName string
	Comment  string
}

func main() {
	// Find project root (where go.mod is)
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	wasmDir := filepath.Join(root, "cmd", "wasm")
	storageDir := filepath.Join(root, "internal", "storage")
	handlersDir := filepath.Join(root, "internal", "api", "handlers")
	tsOutput := filepath.Join(root, "web", "src", "lib", "wasm", "types.gen.ts")

	// Parse WASM Go files (anonymous structs for inputs)
	wasmStructs, err := parseDir(wasmDir, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing WASM files: %v\n", err)
		os.Exit(1)
	}

	// Parse storage Go files (named structs for responses)
	storageStructs, err := parseDir(storageDir, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing storage files: %v\n", err)
		os.Exit(1)
	}

	// Parse handlers Go files (named structs for API responses)
	handlerStructs, err := parseDir(handlersDir, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing handler files: %v\n", err)
		os.Exit(1)
	}

	// Combine all structs
	allStructs := append(wasmStructs, storageStructs...)
	allStructs = append(allStructs, handlerStructs...)

	// Sort and deduplicate
	sort.Slice(allStructs, func(i, j int) bool {
		return allStructs[i].Name < allStructs[j].Name
	})

	seen := make(map[string]bool)
	var unique []GoStruct
	for _, s := range allStructs {
		if !seen[s.Name] {
			seen[s.Name] = true
			unique = append(unique, s)
		}
	}

	// Populate knownStructs for cross-referencing types
	for _, s := range unique {
		knownStructs[s.Name] = true
	}

	// Generate TypeScript code
	if err := generateTypeScript(unique, tsOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating TypeScript: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Generated: %s (%d types)\n", tsOutput, len(unique))
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

// parseDir parses Go files in a directory.
// If anonOnly is true, only anonymous structs (var x struct{}) are parsed.
// If false, named structs (type X struct{}) are also parsed.
func parseDir(dir string, anonOnly bool) ([]GoStruct, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	var allStructs []GoStruct
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		// Skip test files
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		structs, err := parseGoFile(path, anonOnly)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		allStructs = append(allStructs, structs...)
	}

	return allStructs, nil
}

var (
	// Match anonymous struct: var req struct { ... }
	anonStructRe = regexp.MustCompile(`var\s+(\w+)\s+struct\s*\{`)
	// Match named struct: type StructName struct { ... }
	namedStructRe = regexp.MustCompile(`^type\s+(\w+)\s+struct\s*\{`)
	// Match type alias: type TypeName BaseType
	typeAliasRe = regexp.MustCompile(`^type\s+(\w+)\s+(int|int8|int16|int32|int64|uint|uint8|uint16|uint32|uint64|float32|float64|string|bool)\s*$`)
	// Match field with JSON tag: FieldName Type `json:"json_name"`
	fieldRe = regexp.MustCompile(`^\s*(\w+)\s+(\S+)\s+` + "`" + `json:"([^"]+)"` + "`")
	// Match function definition: func funcName(...)
	funcRe = regexp.MustCompile(`^func\s+(\w+)\s*\(`)
	// Match comment before function/struct
	commentRe = regexp.MustCompile(`^//\s*(.*)$`)
)

// typeAliases maps Go type aliases to their base types (e.g., WidgetWidth -> int).
// This is a process-global variable populated during parsing and used during
// type conversion. It is NOT concurrency-safe and is intended for single-threaded
// CLI execution only. If parsing is ever parallelized, this would need synchronization.
var typeAliases = make(map[string]string)

func parseGoFile(path string, anonOnly bool) ([]GoStruct, error) {
	file, err := os.Open(path) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var structs []GoStruct
	var current *GoStruct
	var lastComment string
	var currentFunc string
	inStruct := false
	braceDepth := 0

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Track comments
		if matches := commentRe.FindStringSubmatch(line); matches != nil {
			lastComment = matches[1]
			continue
		}

		// Track current function
		if matches := funcRe.FindStringSubmatch(line); matches != nil {
			currentFunc = matches[1]
			lastComment = ""
			continue
		}

		// Track type aliases (e.g., type WidgetWidth int)
		if matches := typeAliasRe.FindStringSubmatch(line); matches != nil {
			typeAliases[matches[1]] = matches[2]
			lastComment = ""
			continue
		}

		// Check for anonymous struct start
		if matches := anonStructRe.FindStringSubmatch(line); matches != nil {
			structName := deriveTypeName(matches[1], currentFunc, lastComment)
			current = &GoStruct{
				Name:     structName,
				FileName: filepath.Base(path),
				Comment:  lastComment,
			}
			inStruct = true
			braceDepth = 1
			lastComment = ""
			continue
		}

		// Check for named struct start (only if not anonOnly)
		if !anonOnly {
			if matches := namedStructRe.FindStringSubmatch(line); matches != nil {
				structName := matches[1]
				// Skip non-exported types (lowercase first letter)
				if len(structName) > 0 && structName[0] >= 'a' && structName[0] <= 'z' {
					lastComment = ""
					continue
				}
				// Skip repository/handler types (they're not data types)
				if strings.HasSuffix(structName, "Repository") ||
					strings.HasSuffix(structName, "Handler") ||
					strings.HasSuffix(structName, "Router") {
					lastComment = ""
					continue
				}
				current = &GoStruct{
					Name:     structName,
					FileName: filepath.Base(path),
					Comment:  lastComment,
				}
				inStruct = true
				braceDepth = 1
				lastComment = ""
				continue
			}
		}

		// Inside struct, parse fields
		if inStruct && current != nil {
			// Count braces
			braceDepth += strings.Count(line, "{")
			braceDepth -= strings.Count(line, "}")

			// Check for end of struct
			if braceDepth <= 0 {
				if len(current.Fields) > 0 {
					structs = append(structs, *current)
				}
				current = nil
				inStruct = false
				continue
			}

			// Parse field
			if matches := fieldRe.FindStringSubmatch(line); matches != nil {
				jsonName := parseJSONTag(matches[3])
				// Skip ignored fields (json:"-")
				if isIgnoredField(jsonName) {
					continue
				}
				field := GoField{
					GoName:   matches[1],
					GoType:   matches[2],
					JSONName: jsonName,
					Optional: isOptionalType(matches[2]) || hasOmitempty(matches[3]),
				}
				current.Fields = append(current.Fields, field)
			}
		}

		// Reset comment if non-comment line
		if !strings.HasPrefix(strings.TrimSpace(line), "//") {
			lastComment = ""
		}
	}

	return structs, scanner.Err()
}

// deriveTypeName creates a TypeScript type name from the variable name and function context
func deriveTypeName(varName, funcName, comment string) string {
	// Use the function name as the primary source for type naming
	if funcName != "" {
		// Convert function name to type name
		// e.g., "saveActivity" -> "SaveActivityInput"
		// e.g., "getActivities" -> "GetActivitiesRequest"
		name := toPascalCase(funcName)

		// Determine suffix based on the variable name
		if varName == "req" {
			// For "save*" functions, use "Input"
			// For "get*" functions, use "Request"
			if strings.HasPrefix(strings.ToLower(funcName), "save") ||
				strings.HasPrefix(strings.ToLower(funcName), "create") ||
				strings.HasPrefix(strings.ToLower(funcName), "update") ||
				strings.HasPrefix(strings.ToLower(funcName), "compute") {
				return name + "Input"
			}
			return name + "Request"
		}
		return name + toPascalCase(varName)
	}

	// Fallback: use PascalCase of variable name
	name := toPascalCase(varName)
	if name == "Req" {
		return "RequestInput"
	}
	return name + "Input"
}

// parseJSONTag extracts the field name from a JSON tag
func parseJSONTag(tag string) string {
	// Handle "name,omitempty" format
	parts := strings.Split(tag, ",")
	return parts[0]
}

// isIgnoredField checks if a field should be skipped (json:"-")
func isIgnoredField(jsonName string) bool {
	return jsonName == "-"
}

// needsQuoting checks if a field name needs to be quoted in TypeScript
func needsQuoting(name string) bool {
	// Field names with dots, dashes, or other special chars need quoting
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_') {
			return true
		}
	}
	return false
}

// hasOmitempty checks if JSON tag has omitempty
func hasOmitempty(tag string) bool {
	return strings.Contains(tag, "omitempty")
}

// isOptionalType checks if a Go type is a pointer (optional)
func isOptionalType(goType string) bool {
	return strings.HasPrefix(goType, "*")
}

// knownStructs holds all parsed struct names for cross-referencing during
// type conversion. When a field references another struct (e.g., storage.BestEffort),
// this map is checked to determine if the type should be rendered as a known
// interface name or as "unknown".
//
// This is a process-global variable populated in main() after parsing is complete.
// It is NOT concurrency-safe and is intended for single-threaded CLI execution only.
// If parsing is ever parallelized, this would need synchronization.
var knownStructs = make(map[string]bool)

// goTypeToTS converts a Go type to TypeScript
func goTypeToTS(goType string) string {
	// Handle pointer types
	isPointer := strings.HasPrefix(goType, "*")
	baseType := strings.TrimPrefix(goType, "*")

	// Resolve type aliases (e.g., WidgetWidth -> int -> number)
	if aliasBase, ok := typeAliases[baseType]; ok {
		baseType = aliasBase
	}

	var tsType string
	switch baseType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64":
		tsType = "number"
	case "string":
		tsType = "string"
	case "bool":
		tsType = "boolean"
	case "time.Time", "SQLiteTime":
		// time.Time and SQLiteTime serialize to ISO 8601 string in JSON
		tsType = "string"
	case "interface{}":
		tsType = "unknown"
	case "json.RawMessage":
		tsType = "unknown"
	default:
		// Check for slice types
		if strings.HasPrefix(baseType, "[]") {
			elemType := strings.TrimPrefix(baseType, "[]")
			return goTypeToTS(elemType) + "[]"
		}
		// Check for map types: map[K]V -> Record<K, V>
		if strings.HasPrefix(baseType, "map[") {
			// Find the closing bracket for the key type
			closeBracket := strings.Index(baseType, "]")
			if closeBracket > 4 { // "map[" is 4 chars
				keyType := baseType[4:closeBracket]
				valueType := baseType[closeBracket+1:]

				// Convert Go key type to TS
				var tsKeyType string
				switch keyType {
				case "string":
					tsKeyType = "string"
				case "int", "int8", "int16", "int32", "int64",
					"uint", "uint8", "uint16", "uint32", "uint64":
					tsKeyType = "number"
				default:
					tsKeyType = "string" // Fallback for unknown key types
				}

				return fmt.Sprintf("Record<%s, %s>", tsKeyType, goTypeToTS(valueType))
			}
		}
		// Strip package prefix (e.g., "storage.DailyTrainingLoadPoint" -> "DailyTrainingLoadPoint")
		typeName := baseType
		if idx := strings.LastIndex(baseType, "."); idx >= 0 {
			typeName = baseType[idx+1:]
		}
		// Check if it's a known struct type
		if knownStructs[typeName] {
			tsType = typeName
		} else {
			// Unknown type - log warning for visibility
			log.Printf("WARNING: unknown Go type %q (base: %q) - falling back to 'unknown'", goType, typeName)
			tsType = "unknown"
		}
	}

	if isPointer {
		return tsType + " | null"
	}
	return tsType
}

func toPascalCase(s string) string {
	// First handle snake_case by splitting on underscore
	parts := strings.Split(s, "_")

	var result []string
	for _, p := range parts {
		if len(p) == 0 {
			continue
		}

		// Handle common abbreviations
		upper := strings.ToUpper(p)
		if upper == "ID" || upper == "URL" || upper == "API" || upper == "HR" {
			result = append(result, upper)
			continue
		}

		// Handle camelCase by finding uppercase letters
		// e.g., "saveActivity" -> ["save", "Activity"]
		var words []string
		var currentWord strings.Builder
		for i, r := range p {
			if i > 0 && r >= 'A' && r <= 'Z' {
				if currentWord.Len() > 0 {
					words = append(words, currentWord.String())
					currentWord.Reset()
				}
			}
			currentWord.WriteRune(r)
		}
		if currentWord.Len() > 0 {
			words = append(words, currentWord.String())
		}

		// Capitalize each word
		for _, word := range words {
			if len(word) > 0 {
				upperWord := strings.ToUpper(word)
				if upperWord == "ID" || upperWord == "URL" || upperWord == "API" || upperWord == "HR" {
					result = append(result, upperWord)
				} else {
					result = append(result, strings.ToUpper(word[:1])+strings.ToLower(word[1:]))
				}
			}
		}
	}
	return strings.Join(result, "")
}

func generateTypeScript(structs []GoStruct, output string) error {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by scripts/generate-ts-types. DO NOT EDIT.
//
// This file contains TypeScript interfaces that match the JSON serialization
// format used by the Go WASM bridge. These types should be used for all
// communication between TypeScript and Go WASM.
//
// To regenerate: just generate-ts-types

`)

	// Group structs by source file for better organization
	fileGroups := make(map[string][]GoStruct)
	for _, s := range structs {
		fileGroups[s.FileName] = append(fileGroups[s.FileName], s)
	}

	// Sort file names
	var fileNames []string
	for name := range fileGroups {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		fileStructs := fileGroups[fileName]

		buf.WriteString(fmt.Sprintf("// ============================================================================\n"))
		buf.WriteString(fmt.Sprintf("// From %s\n", fileName))
		buf.WriteString(fmt.Sprintf("// ============================================================================\n\n"))

		for _, s := range fileStructs {
			if s.Comment != "" {
				buf.WriteString(fmt.Sprintf("/** %s */\n", s.Comment))
			}
			buf.WriteString(fmt.Sprintf("export interface %s {\n", s.Name))

			for _, f := range s.Fields {
				tsType := goTypeToTS(f.GoType)
				optionalMark := ""
				if f.Optional {
					optionalMark = "?"
				}
				// Quote field names with special characters
				fieldName := f.JSONName
				if needsQuoting(fieldName) {
					fieldName = fmt.Sprintf("'%s'", fieldName)
				}
				buf.WriteString(fmt.Sprintf("  %s%s: %s\n", fieldName, optionalMark, tsType))
			}

			buf.WriteString("}\n\n")
		}
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil { //nolint:gosec
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(output, buf.Bytes(), 0o644) //nolint:gosec
}
