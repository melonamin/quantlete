// Code generation script for go-storage TypeScript declarations.
// Parses cmd/wasm/*.go files and generates TypeScript type declarations
// for the goStorage global object.
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// WasmFunc represents a WASM function exposed to JavaScript
type WasmFunc struct {
	JSName     string   // JavaScript name (e.g., "getActivities")
	GoName     string   // Go function name (e.g., "getActivities")
	Comment    string   // Documentation comment
	Params     []Param  // Parameters from "Called from JS:" comment
	ReturnType string   // TypeScript return type
	Category   string   // Category from comment section
}

// Param represents a function parameter
type Param struct {
	Name string
	Type string // TypeScript type
}

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	wasmDir := filepath.Join(root, "cmd", "wasm")
	tsOutput := filepath.Join(root, "web", "src", "lib", "wasm", "go-storage.gen.ts")

	// Parse main.go for function registrations
	funcs, err := parseWasmFunctions(wasmDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing WASM functions: %v\n", err)
		os.Exit(1)
	}

	// Generate TypeScript
	if err := generateTypeScript(funcs, tsOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating TypeScript: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated: %s (%d functions)\n", tsOutput, len(funcs))
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

var (
	// Matches: "jsName": js.FuncOf(goFuncName),
	registrationRe = regexp.MustCompile(`"(\w+)":\s*js\.FuncOf\((\w+)\)`)
	// Matches: // Called from JS: goStorage.funcName(params)
	calledFromRe = regexp.MustCompile(`// Called from JS: goStorage\.(\w+)\((.*?)\)`)
	// Matches: func funcName(this js.Value, args []js.Value) interface{}
	funcDefRe = regexp.MustCompile(`func (\w+)\(this js\.Value, args \[\]js\.Value\)`)
	// Matches category comments: // ============ Category ============
	categoryRe = regexp.MustCompile(`^// =+\s*([^=]+?)\s*=+\s*$`)
)

func parseWasmFunctions(wasmDir string) ([]WasmFunc, error) {
	// First, parse main.go for registrations
	registrations, err := parseRegistrations(filepath.Join(wasmDir, "main.go"))
	if err != nil {
		return nil, fmt.Errorf("parsing registrations: %w", err)
	}

	// Then parse all .go files for function details
	funcDetails := make(map[string]WasmFunc)
	entries, err := os.ReadDir(wasmDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		details, err := parseFunctionDetails(filepath.Join(wasmDir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		for goName, detail := range details {
			funcDetails[goName] = detail
		}
	}

	// Combine registrations with function details
	var result []WasmFunc
	for jsName, goName := range registrations {
		if detail, ok := funcDetails[goName]; ok {
			detail.JSName = jsName
			detail.GoName = goName
			result = append(result, detail)
		} else {
			// Function registered but details not found - use defaults
			result = append(result, WasmFunc{
				JSName:     jsName,
				GoName:     goName,
				ReturnType: "string",
			})
		}
	}

	// Sort by JS name for consistent output
	sort.Slice(result, func(i, j int) bool {
		return result[i].JSName < result[j].JSName
	})

	return result, nil
}

func parseRegistrations(mainPath string) (map[string]string, error) {
	file, err := os.Open(mainPath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	registrations := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := registrationRe.FindStringSubmatch(line); matches != nil {
			jsName := matches[1]
			goName := matches[2]
			registrations[jsName] = goName
		}
	}

	return registrations, scanner.Err()
}

func parseFunctionDetails(filePath string) (map[string]WasmFunc, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	result := make(map[string]WasmFunc)
	scanner := bufio.NewScanner(file)

	var currentCategory string
	var commentLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Track category headers
		if matches := categoryRe.FindStringSubmatch(line); matches != nil {
			currentCategory = strings.TrimSpace(matches[1])
			commentLines = nil
			continue
		}

		// Collect comment lines
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			commentLines = append(commentLines, line)
			continue
		}

		// Check for function definition
		if matches := funcDefRe.FindStringSubmatch(line); matches != nil {
			goName := matches[1]

			// Parse the collected comments
			var comment string
			var params []Param
			returnType := "string" // Default

			for _, cl := range commentLines {
				cl = strings.TrimSpace(cl)
				cl = strings.TrimPrefix(cl, "//")
				cl = strings.TrimSpace(cl)

				// Check for "Called from JS:" pattern
				if callMatch := calledFromRe.FindStringSubmatch("// " + cl); callMatch != nil {
					params = parseParamsFromSignature(callMatch[2])
				} else if !strings.HasPrefix(cl, "=") && comment == "" {
					// First non-category comment is the description
					comment = cl
				}
			}

			// Infer return type from function name
			returnType = inferReturnType(goName)

			result[goName] = WasmFunc{
				GoName:     goName,
				Comment:    comment,
				Params:     params,
				ReturnType: returnType,
				Category:   currentCategory,
			}

			commentLines = nil
		} else if !strings.HasPrefix(strings.TrimSpace(line), "//") && strings.TrimSpace(line) != "" {
			// Non-comment, non-function line - reset comments
			commentLines = nil
		}
	}

	return result, scanner.Err()
}

func parseParamsFromSignature(paramStr string) []Param {
	paramStr = strings.TrimSpace(paramStr)
	if paramStr == "" {
		return nil
	}

	var params []Param
	parts := strings.Split(paramStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Check for optional marker (year? -> year with optional: true)
		name := part
		optional := strings.HasSuffix(name, "?")
		if optional {
			name = strings.TrimSuffix(name, "?")
		}

		// Infer type from parameter name (pass original with ? for type lookup)
		tsType := inferParamType(part)

		// Add optional marker to name if present
		paramName := name
		if optional {
			paramName += "?"
		}

		params = append(params, Param{
			Name: paramName,
			Type: tsType,
		})
	}

	return params
}

func inferParamType(name string) string {
	lowerName := strings.ToLower(name)

	// JSON suffix indicates string containing JSON
	if strings.HasSuffix(lowerName, "json") {
		return "string"
	}

	// ID fields (except gear IDs which are strings like "b12345")
	if strings.HasSuffix(lowerName, "id") && !strings.Contains(lowerName, "gear") {
		return "number"
	}

	// Specific numeric fields
	numericNames := []string{
		"limit", "year", "month", "page", "windowseconds",
		"durationseconds", "np", "ftp", "initialatl", "initialctl",
		"plannedtss", "targettsb", "ctltau", "atltau",
		"currentctl", "currentatl", "stepstocalculate", "currente",
		"segmentid", "componentid",
	}
	// Handle optional parameters (year? -> year)
	cleanName := strings.TrimSuffix(lowerName, "?")
	for _, n := range numericNames {
		if cleanName == n {
			return "number"
		}
	}

	// Boolean fields
	if strings.HasPrefix(lowerName, "is") || strings.HasPrefix(lowerName, "has") ||
		strings.HasPrefix(lowerName, "include") || strings.HasPrefix(lowerName, "force") {
		return "boolean"
	}

	// Array fields (passed as JS arrays, not JSON)
	arrayNames := []string{"watts", "values", "distances", "dailytss"}
	for _, n := range arrayNames {
		if lowerName == n {
			return "number[]"
		}
	}

	// Default to string
	return "string"
}

func inferReturnType(goName string) string {
	lowerName := strings.ToLower(goName)

	// Special cases - functions that return non-JSON data
	if lowerName == "exportdatabase" || lowerName == "exportdb" {
		return "Uint8Array"
	}

	// Most functions return JSON string
	return "string"
}

func generateTypeScript(funcs []WasmFunc, output string) error {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by scripts/generate-go-storage. DO NOT EDIT.
//
// This file contains TypeScript declarations for the Go WASM storage layer.
// The goStorage global object is registered by the Go WASM binary at runtime.
//
// To regenerate: just generate-go-storage

/**
 * Go WASM storage interface.
 * All methods are synchronous and return JSON strings (except exportDb).
 * Parse results with JSON.parse() and check for errors.
 */
export interface GoStorageInterface {
`)

	// Group functions by category
	categories := make(map[string][]WasmFunc)
	var categoryOrder []string
	seen := make(map[string]bool)

	for _, f := range funcs {
		cat := f.Category
		if cat == "" {
			cat = "Other"
		}
		if !seen[cat] {
			seen[cat] = true
			categoryOrder = append(categoryOrder, cat)
		}
		categories[cat] = append(categories[cat], f)
	}

	for _, cat := range categoryOrder {
		fns := categories[cat]
		buf.WriteString(fmt.Sprintf("\n  // %s\n", cat))

		for _, f := range fns {
			// Write JSDoc comment
			if f.Comment != "" {
				buf.WriteString(fmt.Sprintf("  /** %s */\n", f.Comment))
			}

			// Write method signature
			paramStr := ""
			for i, p := range f.Params {
				if i > 0 {
					paramStr += ", "
				}
				paramStr += fmt.Sprintf("%s: %s", p.Name, p.Type)
			}

			buf.WriteString(fmt.Sprintf("  %s(%s): %s\n", f.JSName, paramStr, f.ReturnType))
		}
	}

	buf.WriteString(`}

/**
 * Global goStorage object registered by Go WASM.
 * Available after WASM initialization.
 */
declare const goStorage: GoStorageInterface

export { goStorage }
`)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil { //nolint:gosec
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(output, buf.Bytes(), 0o644) //nolint:gosec
}
