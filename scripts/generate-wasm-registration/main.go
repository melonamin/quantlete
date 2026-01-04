// Code generation script for WASM function registration.
// Parses cmd/wasm/*.go files for //wasm:export comments and generates
// the registration map in registration.gen.go.
//
// Usage:
//
//	go run ./scripts/generate-wasm-registration
//
// The generator looks for functions with //wasm:export comments:
//
//	//wasm:export getActivities
//	//wasm:category Activities
//	func getActivities(this js.Value, args []js.Value) interface{} { ... }
//
// If the JS name is omitted, it defaults to the Go function name:
//
//	//wasm:export
//	func saveActivity(this js.Value, args []js.Value) interface{} { ... }
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

// WasmFunc represents a WASM function to be registered
type WasmFunc struct {
	JSName   string // JavaScript name
	GoName   string // Go function name
	Category string // Category for grouping
}

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	wasmDir := filepath.Join(root, "cmd", "wasm")
	outputPath := filepath.Join(wasmDir, "registration.gen.go")

	// Parse all wasm/*.go files for //wasm:export comments
	funcs, err := parseExportedFunctions(wasmDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing WASM functions: %v\n", err)
		os.Exit(1)
	}

	if len(funcs) == 0 {
		fmt.Println("No //wasm:export functions found. To use this generator:")
		fmt.Println("  Add //wasm:export [jsName] comment above WASM functions")
		fmt.Println("  Add //wasm:category Name comment to set category")
		os.Exit(0)
	}

	// Generate registration code
	if err := generateRegistration(funcs, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating registration: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated: %s (%d functions)\n", outputPath, len(funcs))
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
	// Matches: //wasm:export [jsName]
	exportRe = regexp.MustCompile(`^//wasm:export(?:\s+(\w+))?$`)
	// Matches: //wasm:category Name
	categoryRe = regexp.MustCompile(`^//wasm:category\s+(.+)$`)
	// Matches: func funcName(this js.Value, args []js.Value) interface{}
	funcDefRe = regexp.MustCompile(`^func (\w+)\(this js\.Value, args \[\]js\.Value\)`)
	// Matches: var funcName = wrapWasm*("jsName", ...) or var funcName = wrapWasm*(...)
	wrapperDefRe = regexp.MustCompile(`^var (\w+) = wrap\w+\(`)
)

func parseExportedFunctions(wasmDir string) ([]WasmFunc, error) {
	entries, err := os.ReadDir(wasmDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	// Use map to deduplicate: generated adapters override manual implementations
	funcMap := make(map[string]WasmFunc)

	// First pass: parse manual files (non-generated)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".gen.go") {
			continue
		}

		fileFuncs, err := parseFile(filepath.Join(wasmDir, name))
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		for _, f := range fileFuncs {
			funcMap[f.JSName] = f
		}
	}

	// Second pass: parse adapters.gen.go (these override manual implementations)
	adaptersPath := filepath.Join(wasmDir, "adapters.gen.go")
	if _, err := os.Stat(adaptersPath); err == nil {
		adapterFuncs, err := parseFile(adaptersPath)
		if err != nil {
			return nil, fmt.Errorf("parsing adapters.gen.go: %w", err)
		}
		for _, f := range adapterFuncs {
			funcMap[f.JSName] = f // Override manual implementation
		}
	}

	// Convert map to slice
	funcs := make([]WasmFunc, 0, len(funcMap))
	for _, f := range funcMap {
		funcs = append(funcs, f)
	}

	return funcs, nil
}

func parseFile(filePath string) ([]WasmFunc, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var funcs []WasmFunc
	var pendingExport string     // JS name from //wasm:export
	var currentCategory string   // Current category from //wasm:category
	var hasExport bool

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Check for category directive
		if matches := categoryRe.FindStringSubmatch(line); matches != nil {
			currentCategory = strings.TrimSpace(matches[1])
			continue
		}

		// Check for export directive
		if matches := exportRe.FindStringSubmatch(line); matches != nil {
			hasExport = true
			pendingExport = matches[1] // May be empty, will use function name
			continue
		}

		// Check for function definition (old style: func name(...))
		if matches := funcDefRe.FindStringSubmatch(line); matches != nil && hasExport {
			goName := matches[1]
			jsName := pendingExport
			if jsName == "" {
				jsName = goName // Default to Go function name
			}

			funcs = append(funcs, WasmFunc{
				JSName:   jsName,
				GoName:   goName,
				Category: currentCategory,
			})

			hasExport = false
			pendingExport = ""
		}

		// Check for wrapper definition (new style: var name = wrapWasm*(...))
		if matches := wrapperDefRe.FindStringSubmatch(line); matches != nil && hasExport {
			goName := matches[1]
			jsName := pendingExport
			if jsName == "" {
				jsName = goName // Default to Go variable name
			}

			funcs = append(funcs, WasmFunc{
				JSName:   jsName,
				GoName:   goName,
				Category: currentCategory,
			})

			hasExport = false
			pendingExport = ""
		}

		// Reset on non-comment, non-empty lines (except func/var defs)
		if !strings.HasPrefix(line, "//") && line != "" && !strings.HasPrefix(line, "func ") && !strings.HasPrefix(line, "var ") {
			hasExport = false
			pendingExport = ""
		}
	}

	return funcs, scanner.Err()
}

func generateRegistration(funcs []WasmFunc, outputPath string) error {
	// Group functions by category
	byCategory := make(map[string][]WasmFunc)
	for _, f := range funcs {
		cat := f.Category
		if cat == "" {
			cat = "Uncategorized"
		}
		byCategory[cat] = append(byCategory[cat], f)
	}

	// Sort categories and functions within each category
	var categories []string
	for cat := range byCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		sort.Slice(byCategory[cat], func(i, j int) bool {
			return byCategory[cat][i].JSName < byCategory[cat][j].JSName
		})
	}

	// Generate code
	var buf bytes.Buffer
	buf.WriteString(`// Code generated by scripts/generate-wasm-registration. DO NOT EDIT.

//go:build js && wasm

package main

import "syscall/js"

// wasmFunctions contains all WASM functions to be registered with JavaScript.
var wasmFunctions = map[string]interface{}{
`)

	for _, cat := range categories {
		buf.WriteString(fmt.Sprintf("\t// %s\n", cat))
		for _, f := range byCategory[cat] {
			buf.WriteString(fmt.Sprintf("\t%q: js.FuncOf(%s),\n", f.JSName, f.GoName))
		}
		buf.WriteString("\n")
	}

	buf.WriteString(`}

// RegisterAll registers all WASM functions under the goStorage namespace.
func RegisterAll() {
	js.Global().Set("goStorage", wasmFunctions)
}
`)

	return os.WriteFile(outputPath, buf.Bytes(), 0o644)
}
