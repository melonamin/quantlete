// Code generation script for go-storage TypeScript declarations and wrappers.
// Parses cmd/wasm/*.go files and generates TypeScript type declarations
// for the goStorage global object, plus wrapper implementations.
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
	JSName       string  // JavaScript name (e.g., "getActivities")
	GoName       string  // Go function name (e.g., "getActivities")
	Comment      string  // Documentation comment
	Params       []Param // Parameters from "Called from JS:" comment
	ReturnType   string  // TypeScript return type
	Category     string  // Category from comment section
	ResponseType string  // How the response should be handled (data, value, direct, void)
	TSReturnType string  // TypeScript return type for the wrapper function
}

// Param represents a function parameter
type Param struct {
	Name     string
	Type     string // TypeScript type
	Optional bool   // Whether the parameter is optional
}

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	wasmDir := filepath.Join(root, "cmd", "wasm")
	tsOutput := filepath.Join(root, "web", "src", "lib", "wasm", "go-storage.gen.ts")
	wrappersOutput := filepath.Join(root, "web", "src", "lib", "wasm", "go-storage-wrappers.gen.ts")

	// Parse main.go for function registrations
	funcs, err := parseWasmFunctions(wasmDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing WASM functions: %v\n", err)
		os.Exit(1)
	}

	// Generate TypeScript interface
	if err := generateTypeScript(funcs, tsOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating TypeScript: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated: %s (%d functions)\n", tsOutput, len(funcs))

	// Generate TypeScript wrappers
	if err := generateWrappers(funcs, wrappersOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating TypeScript wrappers: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated: %s (%d wrappers)\n", wrappersOutput, len(funcs))
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
	// Matches: var funcName = wrapWasm*("jsName", ...) or wrapWasm*(...)
	wrapperDefRe = regexp.MustCompile(`var (\w+) = wrap\w+\(`)
	// Matches category comments: // ============ Category ============
	categoryRe = regexp.MustCompile(`^// =+\s*([^=]+?)\s*=+\s*$`)
	// Matches: //wasm:export jsName
	wasmExportRe = regexp.MustCompile(`//wasm:export\s+(\w+)`)
	// Matches: //wasm:category CategoryName
	wasmCategoryRe = regexp.MustCompile(`//wasm:category\s+(.+)`)
	// Matches: json.Unmarshal([]byte(args[0].String()), &input)
	jsonInputRe = regexp.MustCompile(`json\.Unmarshal\(\[\]byte\(args\[0\]\.String\(\)\)`)
	// Matches: args[0].Int() or args[0].Float() for direct numeric params
	directNumericRe = regexp.MustCompile(`args\[(\d+)\]\.(Int|Float)\(\)`)
	// Matches: wc.ArgJSON(0, &input) - new wrapper style JSON input
	wcArgJSONRe = regexp.MustCompile(`wc\.ArgJSON\((\d+),`)
	// Matches: wc.ArgInt(0) or wc.ArgString(0) - new wrapper style direct input
	wcArgDirectRe = regexp.MustCompile(`wc\.Arg(Int|String|Float)\((\d+)\)`)
)

func parseWasmFunctions(wasmDir string) ([]WasmFunc, error) {
	// First, try registration.gen.go, then fall back to main.go
	registrations, err := parseRegistrations(filepath.Join(wasmDir, "registration.gen.go"))
	if err != nil {
		// Try main.go as fallback
		registrations, err = parseRegistrations(filepath.Join(wasmDir, "main.go"))
		if err != nil {
			return nil, fmt.Errorf("parsing registrations: %w", err)
		}
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
	// Read entire file to allow look-ahead for function body analysis
	content, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	result := make(map[string]WasmFunc)

	var currentCategory string
	var commentLines []string
	var wasmExportName string
	var wasmCategory string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Track category headers (old style: // ====== Category ======)
		if matches := categoryRe.FindStringSubmatch(line); matches != nil {
			currentCategory = strings.TrimSpace(matches[1])
			commentLines = nil
			continue
		}

		// Track //wasm:category directive
		if matches := wasmCategoryRe.FindStringSubmatch(line); matches != nil {
			wasmCategory = strings.TrimSpace(matches[1])
			continue
		}

		// Track //wasm:export directive
		if matches := wasmExportRe.FindStringSubmatch(line); matches != nil {
			wasmExportName = matches[1]
			continue
		}

		// Collect comment lines
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			commentLines = append(commentLines, line)
			continue
		}

		// Check for function definition (old style or wrapper style)
		var goName string
		if matches := funcDefRe.FindStringSubmatch(line); matches != nil {
			goName = matches[1]
		} else if matches := wrapperDefRe.FindStringSubmatch(line); matches != nil {
			goName = matches[1]
		}

		if goName != "" {
			// Parse the collected comments
			var comment string
			var params []Param
			returnType := "string" // Default

			for _, cl := range commentLines {
				cl = strings.TrimSpace(cl)
				cl = strings.TrimPrefix(cl, "//")
				cl = strings.TrimSpace(cl)

				// Check for "Called from JS:" pattern (old style)
				if callMatch := calledFromRe.FindStringSubmatch("// " + cl); callMatch != nil {
					params = parseParamsFromSignature(callMatch[2])
				} else if !strings.HasPrefix(cl, "=") && !strings.HasPrefix(cl, "wasm:") && comment == "" {
					// First non-category, non-directive comment is the description
					comment = cl
				}
			}

			// If no params from "Called from JS:", analyze function body
			if len(params) == 0 {
				params = analyzeFunctionBody(lines, i)
			}

			// Infer return type from function name
			returnType = inferReturnType(goName)

			// Use wasmCategory if set, otherwise use currentCategory
			category := currentCategory
			if wasmCategory != "" {
				category = wasmCategory
			}

			result[goName] = WasmFunc{
				GoName:     goName,
				Comment:    comment,
				Params:     params,
				ReturnType: returnType,
				Category:   category,
			}

			// Also store by wasmExportName if different from goName
			// This helps when registration uses jsName -> genGoName mapping
			if wasmExportName != "" && wasmExportName != goName {
				// Store reference so we can look up by either name
				result[wasmExportName+"_export"] = result[goName]
			}

			commentLines = nil
			wasmExportName = ""
			wasmCategory = ""
		} else if !strings.HasPrefix(strings.TrimSpace(line), "//") && strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "var ") {
			// Non-comment, non-function, non-var line - reset comments
			commentLines = nil
		}
	}

	return result, nil
}

// analyzeFunctionBody looks at the function body to detect input patterns
func analyzeFunctionBody(lines []string, funcLineIdx int) []Param {
	// Look at the next ~30 lines to find the function body patterns
	braceCount := 0
	started := false

	for i := funcLineIdx; i < len(lines) && i < funcLineIdx+50; i++ {
		line := lines[i]

		// Track braces to stay within function
		braceCount += strings.Count(line, "{") - strings.Count(line, "}")
		if strings.Contains(line, "{") {
			started = true
		}
		if started && braceCount <= 0 {
			break
		}

		// Check for JSON input pattern: json.Unmarshal([]byte(args[0].String())
		if jsonInputRe.MatchString(line) {
			// This function accepts a JSON string input
			return []Param{{
				Name:     "dataJSON",
				Type:     "string",
				Optional: true, // Most functions check for empty string
			}}
		}

		// Check for wrapper-style JSON input: wc.ArgJSON(0, &input)
		if wcArgJSONRe.MatchString(line) {
			// This function accepts a JSON string input (wrapper style)
			return []Param{{
				Name:     "dataJSON",
				Type:     "string",
				Optional: true, // Usually checked with wc.HasArg(0)
			}}
		}

		// Check for direct numeric args like args[0].Int()
		if matches := directNumericRe.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			// Found direct numeric parameter access
			var params []Param
			for _, match := range matches {
				idx := match[1]
				paramType := "number"
				paramName := fmt.Sprintf("arg%s", idx)

				// Try to infer better name from context
				if strings.Contains(line, "limit") {
					paramName = "limit"
				} else if strings.Contains(line, "year") {
					paramName = "year"
				} else if strings.Contains(line, "month") {
					paramName = "month"
				} else if strings.Contains(line, "id") || strings.Contains(line, "Id") || strings.Contains(line, "ID") {
					paramName = "id"
				}

				params = append(params, Param{
					Name:     paramName,
					Type:     paramType,
					Optional: true, // Usually checked with len(args) > 0
				})
			}
			if len(params) > 0 {
				return params
			}
		}

		// Check for wrapper-style direct args: wc.ArgInt(0), wc.ArgString(0), wc.ArgFloat(0)
		if matches := wcArgDirectRe.FindAllStringSubmatch(line, -1); len(matches) > 0 {
			var params []Param
			for _, match := range matches {
				argType := match[1] // Int, String, Float
				idx := match[2]
				paramType := "number"
				if argType == "String" {
					paramType = "string"
				}
				paramName := fmt.Sprintf("arg%s", idx)

				// Try to infer better name from context
				if strings.Contains(line, "limit") {
					paramName = "limit"
				} else if strings.Contains(line, "year") {
					paramName = "year"
				} else if strings.Contains(line, "month") {
					paramName = "month"
				} else if strings.Contains(line, "id") || strings.Contains(line, "Id") || strings.Contains(line, "ID") {
					paramName = "id"
				}

				params = append(params, Param{
					Name:     paramName,
					Type:     paramType,
					Optional: true,
				})
			}
			if len(params) > 0 {
				return params
			}
		}
	}

	return nil
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

		params = append(params, Param{
			Name:     name,
			Type:     tsType,
			Optional: optional,
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
				optMarker := ""
				if p.Optional {
					optMarker = "?"
				}
				paramStr += fmt.Sprintf("%s%s: %s", p.Name, optMarker, p.Type)
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

// generateWrappers creates TypeScript wrapper functions
func generateWrappers(funcs []WasmFunc, output string) error {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by scripts/generate-go-storage. DO NOT EDIT.
//
// This file contains TypeScript wrapper utilities and generated functions
// for the Go WASM storage layer. These handle JSON serialization and error handling.
//
// To regenerate: just generate-go-storage

let initialized = false

export interface GoStorageResult<T = unknown> {
  ok: boolean
  error?: string
  message?: string
  data?: T
}

/**
 * Set initialization state (called by go-storage.ts after WASM init)
 */
export function setInitialized(value: boolean): void {
  initialized = value
}

/**
 * Check if storage is initialized
 */
export function isStorageInitialized(): boolean {
  return initialized
}

/**
 * Parse Go WASM result - all Go functions return JSON strings.
 */
export function parseGoResult<T>(result: string, operation?: string): GoStorageResult<T> & T {
  try {
    return JSON.parse(result)
  } catch {
    const ctx = operation ? ` + "`" + `(${operation})` + "`" + ` : ''
    return {
      ok: false,
      error: ` + "`" + `Failed to parse Go result${ctx}: ${result}` + "`" + `,
    } as GoStorageResult<T> & T
  }
}

/**
 * Helper for calling Go storage functions that return data directly.
 * Handles initialization check, parsing, and error handling.
 * Note: Go WASM wraps responses in {"ok": true, "data": <result>},
 * so we extract .data when present.
 */
export function callGoStorage<T>(
  fn: () => string,
  operation: string
): T {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<T>(fn(), operation)
  if (!result.ok) {
    throw new Error(result.error || ` + "`" + `Failed to ${operation}` + "`" + `)
  }
  // Go WASM wraps responses in {ok, data}, extract .data if present
  // Otherwise return the result directly (for backwards compatibility with inline fields)
  if (result.data !== undefined) {
    return result.data as T
  }
  return result as T
}

/**
 * Helper for calling Go storage functions that return arrays in .data.
 */
export function callGoStorageArray<T>(
  fn: () => string,
  operation: string
): T[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<T[]>(fn(), operation)
  if (!result.ok) {
    throw new Error(result.error || ` + "`" + `Failed to ${operation}` + "`" + `)
  }
  return result.data ?? []
}

/**
 * Helper for calling Go storage functions that return a single value.
 */
export function callGoStorageValue<T extends number | string | boolean>(
  fn: () => string,
  operation: string
): T {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ value: T }>(fn(), operation)
  if (!result.ok) {
    throw new Error(result.error || ` + "`" + `Failed to ${operation}` + "`" + `)
  }
  return (result as { value: T }).value
}

/**
 * Helper for calling Go storage functions that don't return data.
 */
export function callGoStorageVoid(
  fn: () => string,
  operation: string
): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<void>(fn(), operation)
  if (!result.ok) {
    throw new Error(result.error || ` + "`" + `Failed to ${operation}` + "`" + `)
  }
}

`)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil { //nolint:gosec
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(output, buf.Bytes(), 0o644) //nolint:gosec
}
