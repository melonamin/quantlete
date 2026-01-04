// Code generation script for HTTP and WASM adapters.
// Parses service files for //adapter: comments and generates
// adapter code that calls service methods.
//
// Usage:
//
//	go run ./scripts/generate-adapters
//
// The generator looks for service methods with //adapter: comments:
//
//	//adapter:wasm getActivities
//	//adapter:http GET /api/v1/activities
//	func (s *ActivityService) List(ctx context.Context, in ListActivitiesInput) (*ListActivitiesOutput, error)
//
// Input types should have appropriate struct tags:
//
//	type ListActivitiesInput struct {
//	    AthleteID  int64    `json:"athlete_id" adapter:"context"`  // from global/strava
//	    SportTypes []string `json:"sport_types" adapter:"query,name=sport_type,split=,"`
//	    Page       int      `json:"page" adapter:"query"`
//	}
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// AdapterMethod represents a service method to generate adapters for
type AdapterMethod struct {
	ServiceName string
	MethodName  string
	InputType   string
	OutputType  string
	WasmName    string // JS function name for WASM
	HTTPMethod  string // GET, POST, PUT, DELETE
	HTTPPath    string // /api/v1/...
	Category    string // For grouping in generated code
}

// InputField represents a field in an input struct
type InputField struct {
	Name        string
	Type        string
	JSONName    string
	AdapterTag  string // adapter tag value
	Source      string // context, query, path, body
	QueryName   string // override name for query param
	SplitChar   string // character to split comma-separated values
	PathParam   string // path parameter name
	Default     string // default value for the field
	IsPointer   bool
	IsSlice     bool
	ElementType string // for slices
}

// InputType represents a parsed input struct
type InputType struct {
	Name   string
	Fields []InputField
}

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	servicesDir := filepath.Join(root, "internal", "services")

	// Parse all service files for adapter methods and input types
	methods, inputTypes, err := parseServices(servicesDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing services: %v\n", err)
		os.Exit(1)
	}

	if len(methods) == 0 {
		fmt.Println("No //adapter: methods found. To use this generator:")
		fmt.Println("  Add //adapter:wasm jsName comment above service methods")
		fmt.Println("  Add //adapter:http METHOD /path comment above service methods")
		os.Exit(0)
	}

	// Generate WASM adapters
	wasmMethods := filterWasmMethods(methods)
	if len(wasmMethods) > 0 {
		wasmPath := filepath.Join(root, "cmd", "wasm", "adapters.gen.go")
		if err := generateWasmAdapters(wasmMethods, inputTypes, wasmPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating WASM adapters: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated WASM: %s (%d methods)\n", wasmPath, len(wasmMethods))
	}

	// Generate HTTP handlers
	httpMethods := filterHTTPMethods(methods)
	if len(httpMethods) > 0 {
		httpPath := filepath.Join(root, "internal", "api", "handlers", "adapters.gen.go")
		if err := generateHTTPAdapters(httpMethods, inputTypes, httpPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating HTTP adapters: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated HTTP: %s (%d methods)\n", httpPath, len(httpMethods))
	}
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
	// Matches: //adapter:wasm jsName category=CategoryName
	wasmRe = regexp.MustCompile(`^//adapter:wasm\s+(\w+)(?:\s+category=(\S+))?$`)
	// Matches: //adapter:http METHOD /path
	httpRe = regexp.MustCompile(`^//adapter:http\s+(GET|POST|PUT|DELETE|PATCH)\s+(/\S+)$`)
)

func parseServices(servicesDir string) ([]AdapterMethod, map[string]InputType, error) {
	entries, err := os.ReadDir(servicesDir)
	if err != nil {
		return nil, nil, fmt.Errorf("reading directory: %w", err)
	}

	var methods []AdapterMethod
	inputTypes := make(map[string]InputType)

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") {
			continue
		}
		if strings.HasSuffix(name, "_test.go") {
			continue
		}

		filePath := filepath.Join(servicesDir, name)
		fileMethods, fileInputTypes, err := parseServiceFile(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing %s: %w", name, err)
		}

		methods = append(methods, fileMethods...)
		for k, v := range fileInputTypes {
			inputTypes[k] = v
		}
	}

	return methods, inputTypes, nil
}

func parseServiceFile(filePath string) ([]AdapterMethod, map[string]InputType, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing file: %w", err)
	}

	var methods []AdapterMethod
	inputTypes := make(map[string]InputType)

	// First pass: collect input types
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if strings.HasSuffix(typeSpec.Name.Name, "Input") {
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				inputType := parseInputStruct(typeSpec.Name.Name, structType)
				inputTypes[typeSpec.Name.Name] = inputType
			}
		}
	}

	// Second pass: collect adapter methods from comments
	// Read file line by line to find comments before func declarations
	fileData, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return nil, nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(fileData))
	var pendingWasm, pendingHTTPMethod, pendingHTTPPath, pendingCategory string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Check for wasm adapter comment
		if matches := wasmRe.FindStringSubmatch(line); matches != nil {
			pendingWasm = matches[1]
			if len(matches) > 2 && matches[2] != "" {
				pendingCategory = matches[2]
			}
			continue
		}

		// Check for http adapter comment
		if matches := httpRe.FindStringSubmatch(line); matches != nil {
			pendingHTTPMethod = matches[1]
			pendingHTTPPath = matches[2]
			continue
		}

		// Check for func declaration with receiver
		if strings.HasPrefix(line, "func (") && (pendingWasm != "" || pendingHTTPMethod != "") {
			method := parseMethodDecl(line)
			if method != nil {
				method.WasmName = pendingWasm
				method.HTTPMethod = pendingHTTPMethod
				method.HTTPPath = pendingHTTPPath
				method.Category = pendingCategory
				methods = append(methods, *method)
			}
			pendingWasm = ""
			pendingHTTPMethod = ""
			pendingHTTPPath = ""
			pendingCategory = ""
		}

		// Reset pending on non-comment, non-func lines
		if !strings.HasPrefix(line, "//") && line != "" && !strings.HasPrefix(line, "func ") {
			pendingWasm = ""
			pendingHTTPMethod = ""
			pendingHTTPPath = ""
			pendingCategory = ""
		}
	}

	return methods, inputTypes, scanner.Err()
}

func parseInputStruct(name string, structType *ast.StructType) InputType {
	inputType := InputType{Name: name}

	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			continue // embedded field
		}

		for _, fieldName := range field.Names {
			f := InputField{
				Name: fieldName.Name,
				Type: typeToString(field.Type),
			}

			// Parse struct tags
			if field.Tag != nil {
				tag := strings.Trim(field.Tag.Value, "`")
				f.JSONName = getTagValue(tag, "json")
				f.AdapterTag = getTagValue(tag, "adapter")
				parseAdapterTag(&f)
			}

			// Determine if pointer or slice
			if _, ok := field.Type.(*ast.StarExpr); ok {
				f.IsPointer = true
			}
			if arr, ok := field.Type.(*ast.ArrayType); ok {
				f.IsSlice = true
				f.ElementType = typeToString(arr.Elt)
			}

			inputType.Fields = append(inputType.Fields, f)
		}
	}

	return inputType
}

func parseAdapterTag(f *InputField) {
	if f.AdapterTag == "" {
		f.Source = "body" // default
		return
	}

	parts := strings.Split(f.AdapterTag, ",")
	f.Source = parts[0]

	for i, part := range parts[1:] {
		if strings.HasPrefix(part, "name=") {
			f.QueryName = strings.TrimPrefix(part, "name=")
		} else if strings.HasPrefix(part, "split=") {
			val := strings.TrimPrefix(part, "split=")
			if val == "" && i+2 < len(parts) && parts[i+2] == "" {
				// Handle split=, where comma is both delimiter and value
				val = ","
			}
			f.SplitChar = val
		} else if strings.HasPrefix(part, "param=") {
			f.PathParam = strings.TrimPrefix(part, "param=")
		} else if strings.HasPrefix(part, "default=") {
			f.Default = strings.TrimPrefix(part, "default=")
		}
	}
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		return "[]" + typeToString(t.Elt)
	case *ast.SelectorExpr:
		return typeToString(t.X) + "." + t.Sel.Name
	default:
		return "interface{}"
	}
}

func getTagValue(tag, key string) string {
	// Simple tag parser: looks for key:"value"
	idx := strings.Index(tag, key+`:"`)
	if idx == -1 {
		return ""
	}
	start := idx + len(key) + 2
	end := strings.Index(tag[start:], `"`)
	if end == -1 {
		return ""
	}
	value := tag[start : start+end]
	// For json tags, handle omitempty by trimming at comma
	// For adapter tags, keep the full value (commas are used as separators within)
	if key == "json" {
		if comma := strings.Index(value, ","); comma != -1 {
			value = value[:comma]
		}
	}
	return value
}

// Matches service methods like:
// func (s *ServiceName) MethodName(ctx context.Context, in InputType) (*OutputType, error)
// func (s *ServiceName) MethodName(ctx context.Context, in InputType) ([]OutputType, error)
var methodDeclRe = regexp.MustCompile(`^func \((\w+) \*(\w+)\) (\w+)\(ctx context\.Context, in (\w+)\) \((\[?\]?\*?[\w.]+), error\)`)

func parseMethodDecl(line string) *AdapterMethod {
	matches := methodDeclRe.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	return &AdapterMethod{
		ServiceName: matches[2],
		MethodName:  matches[3],
		InputType:   matches[4],
		OutputType:  matches[5],
	}
}

func filterWasmMethods(methods []AdapterMethod) []AdapterMethod {
	var result []AdapterMethod
	for _, m := range methods {
		if m.WasmName != "" {
			result = append(result, m)
		}
	}
	return result
}

func filterHTTPMethods(methods []AdapterMethod) []AdapterMethod {
	var result []AdapterMethod
	for _, m := range methods {
		if m.HTTPMethod != "" {
			result = append(result, m)
		}
	}
	return result
}

func generateWasmAdapters(methods []AdapterMethod, inputTypes map[string]InputType, outputPath string) error {
	// Group by category
	byCategory := make(map[string][]AdapterMethod)
	for _, m := range methods {
		cat := m.Category
		if cat == "" {
			cat = m.ServiceName
		}
		byCategory[cat] = append(byCategory[cat], m)
	}

	var categories []string
	for cat := range byCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		sort.Slice(byCategory[cat], func(i, j int) bool {
			return byCategory[cat][i].WasmName < byCategory[cat][j].WasmName
		})
	}

	var buf bytes.Buffer
	buf.WriteString(`// Code generated by scripts/generate-adapters. DO NOT EDIT.

//go:build js && wasm

package main

import (
	"fmt"

	"github.com/melonamin/quantlete/internal/services"
)

// Ensure imports are used
var (
	_ = fmt.Sprintf
	_ = services.ListActivitiesInput{}
)

`)

	for _, cat := range categories {
		buf.WriteString(fmt.Sprintf("// ============================================================================\n"))
		buf.WriteString(fmt.Sprintf("// %s\n", cat))
		buf.WriteString(fmt.Sprintf("// ============================================================================\n\n"))

		for _, m := range byCategory[cat] {
			inputType := inputTypes[m.InputType]
			generateWasmMethod(&buf, m, inputType)
		}
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0o644)
}

func generateWasmMethod(buf *bytes.Buffer, m AdapterMethod, inputType InputType) {
	// Function comment
	buf.WriteString(fmt.Sprintf("// gen%s wraps %s.%s\n", capitalize(m.WasmName), m.ServiceName, m.MethodName))
	buf.WriteString(fmt.Sprintf("//wasm:export %s\n", m.WasmName))
	if m.Category != "" {
		buf.WriteString(fmt.Sprintf("//wasm:category %s\n", m.Category))
	}

	// Function signature using wrapper pattern
	buf.WriteString(fmt.Sprintf("var gen%s = wrapWasmAthlete(%q, func(wc *WasmContext) interface{} {\n", capitalize(m.WasmName), m.WasmName))

	// Initialize input
	buf.WriteString(fmt.Sprintf("\tvar input services.%s\n", m.InputType))

	// Check if we need to parse JSON (any non-context fields)
	hasNonContextFields := false
	for _, f := range inputType.Fields {
		if f.Source != "context" {
			hasNonContextFields = true
			break
		}
	}

	// Parse JSON from args if there are non-context fields
	if hasNonContextFields {
		buf.WriteString("\n\tif wc.HasArg(0) && wc.ArgString(0) != \"\" {\n")
		buf.WriteString("\t\tif err := wc.ArgJSON(0, &input); err != nil {\n")
		buf.WriteString("\t\t\treturn errorJSON(fmt.Errorf(\"parsing input: %w\", err))\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t}\n")
	}

	// Apply defaults for fields with zero values after parsing
	for _, f := range inputType.Fields {
		if f.Default != "" && f.Source != "context" && !f.IsPointer {
			switch f.Type {
			case "int":
				buf.WriteString(fmt.Sprintf("\tif input.%s == 0 {\n", f.Name))
				buf.WriteString(fmt.Sprintf("\t\tinput.%s = %s\n", f.Name, f.Default))
				buf.WriteString("\t}\n")
			case "string":
				buf.WriteString(fmt.Sprintf("\tif input.%s == \"\" {\n", f.Name))
				buf.WriteString(fmt.Sprintf("\t\tinput.%s = %q\n", f.Name, f.Default))
				buf.WriteString("\t}\n")
			case "bool":
				// For bool defaults, only apply if default is "true" (false is zero value)
				if f.Default == "true" {
					// We can't distinguish "not set" from "set to false", so skip bool defaults
					// Users should explicitly pass false if they want false
				}
			}
		}
	}

	// Set context fields (athleteID) AFTER parsing JSON so it's not overwritten
	for _, f := range inputType.Fields {
		if f.Source == "context" {
			if f.Name == "AthleteID" {
				buf.WriteString("\tinput.AthleteID = wc.AthleteID\n")
			}
		}
	}

	// Call service via registry using wrapper context
	buf.WriteString(fmt.Sprintf("\n\tresult, err := wc.Registry.%s.%s(wc.Ctx, input)\n", m.ServiceName, m.MethodName))
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\treturn errorJSON(err)\n")
	buf.WriteString("\t}\n\n")

	// Return response
	buf.WriteString("\treturn dataJSON(result)\n")
	buf.WriteString("})\n\n")
}

func generateHTTPAdapters(methods []AdapterMethod, inputTypes map[string]InputType, outputPath string) error {
	var buf bytes.Buffer
	buf.WriteString(`// Code generated by scripts/generate-adapters. DO NOT EDIT.

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

// Ensure imports are used
var (
	_ = json.Marshal
	_ = fmt.Errorf
	_ = strconv.Atoi
	_ = strings.Split
	_ = chi.URLParam
	_ = services.ListActivitiesInput{}
	_ = shared.ParseDateParam
	_ = (*strava.Client)(nil)
)

`)

	// Group by service
	byService := make(map[string][]AdapterMethod)
	for _, m := range methods {
		byService[m.ServiceName] = append(byService[m.ServiceName], m)
	}

	var serviceNames []string
	for svc := range byService {
		serviceNames = append(serviceNames, svc)
	}
	sort.Strings(serviceNames)

	for _, svc := range serviceNames {
		for _, m := range byService[svc] {
			inputType := inputTypes[m.InputType]
			generateHTTPMethod(&buf, m, inputType)
		}
	}

	return os.WriteFile(outputPath, buf.Bytes(), 0o644)
}

func generateHTTPMethod(buf *bytes.Buffer, m AdapterMethod, inputType InputType) {
	handlerName := fmt.Sprintf("Gen%s%s", m.ServiceName, m.MethodName)

	buf.WriteString(fmt.Sprintf("// %s handles %s %s\n", handlerName, m.HTTPMethod, m.HTTPPath))
	buf.WriteString(fmt.Sprintf("func %s(svc *services.%s, strava *strava.Client) http.HandlerFunc {\n", handlerName, m.ServiceName))
	buf.WriteString("\treturn func(w http.ResponseWriter, r *http.Request) {\n")

	// Auth check
	buf.WriteString("\t\tathlete := strava.GetAthlete()\n")
	buf.WriteString("\t\tif athlete == nil {\n")
	buf.WriteString("\t\t\tshared.WriteErrorWithStatus(w, http.StatusUnauthorized, fmt.Errorf(\"not authenticated\"))\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n\n")

	// Initialize input
	buf.WriteString(fmt.Sprintf("\t\tvar input services.%s\n", m.InputType))

	// Set context fields
	for _, f := range inputType.Fields {
		if f.Source == "context" {
			if f.Name == "AthleteID" {
				buf.WriteString("\t\tinput.AthleteID = athlete.ID\n")
			}
		}
	}

	// Parse path params
	for _, f := range inputType.Fields {
		if f.Source == "path" {
			paramName := f.PathParam
			if paramName == "" {
				// Default to lowercase field name
				paramName = strings.ToLower(f.Name)
			}
			if f.Type == "string" {
				buf.WriteString(fmt.Sprintf("\t\tinput.%s = chi.URLParam(r, %q)\n", f.Name, paramName))
			} else if f.Type == "int64" {
				buf.WriteString(fmt.Sprintf("\t\tif idStr := chi.URLParam(r, %q); idStr != \"\" {\n", paramName))
				buf.WriteString(fmt.Sprintf("\t\t\tif parsed, err := strconv.ParseInt(idStr, 10, 64); err == nil {\n"))
				buf.WriteString(fmt.Sprintf("\t\t\t\tinput.%s = parsed\n", f.Name))
				buf.WriteString("\t\t\t}\n")
				buf.WriteString("\t\t}\n")
			}
		}
	}

	// Parse query params or body
	if m.HTTPMethod == "GET" || m.HTTPMethod == "DELETE" {
		// Generate query param parsing into temp buffer first
		var queryBuf bytes.Buffer
		for _, f := range inputType.Fields {
			if f.Source == "context" || f.Source == "path" {
				continue
			}
			generateQueryParamParsing(&queryBuf, f)
		}
		// Only declare q if we actually have parsing code that uses it
		if queryBuf.Len() > 0 {
			buf.WriteString("\n\t\tq := r.URL.Query()\n")
			buf.Write(queryBuf.Bytes())
		}
	} else {
		// Parse body for POST/PUT/PATCH
		buf.WriteString("\n\t\tif err := json.NewDecoder(r.Body).Decode(&input); err != nil {\n")
		buf.WriteString("\t\t\tshared.WriteErrorWithStatus(w, http.StatusBadRequest, err)\n")
		buf.WriteString("\t\t\treturn\n")
		buf.WriteString("\t\t}\n")
	}

	// Call service
	buf.WriteString("\n\t\tresult, err := svc." + m.MethodName + "(r.Context(), input)\n")
	buf.WriteString("\t\tif err != nil {\n")
	buf.WriteString("\t\t\thandleServiceError(w, err)\n")
	buf.WriteString("\t\t\treturn\n")
	buf.WriteString("\t\t}\n\n")

	// Write response
	buf.WriteString("\t\tshared.WriteSuccess(w, result)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n\n")
}

func generateQueryParamParsing(buf *bytes.Buffer, f InputField) {
	queryName := f.QueryName
	if queryName == "" {
		queryName = f.JSONName
		if queryName == "" {
			queryName = strings.ToLower(f.Name)
		}
	}

	switch {
	case f.Type == "int" || f.Type == "*int":
		// Apply default first if specified
		if f.Default != "" && !f.IsPointer {
			buf.WriteString(fmt.Sprintf("\t\tinput.%s = %s\n", f.Name, f.Default))
		}
		buf.WriteString(fmt.Sprintf("\t\tif v := q.Get(%q); v != \"\" {\n", queryName))
		buf.WriteString("\t\t\tif parsed, err := strconv.Atoi(v); err == nil {\n")
		if f.IsPointer {
			buf.WriteString(fmt.Sprintf("\t\t\t\tinput.%s = &parsed\n", f.Name))
		} else {
			buf.WriteString(fmt.Sprintf("\t\t\t\tinput.%s = parsed\n", f.Name))
		}
		buf.WriteString("\t\t\t}\n")
		buf.WriteString("\t\t}\n")

	case f.Type == "bool" || f.Type == "*bool":
		// Apply default first if specified
		if f.Default != "" && !f.IsPointer {
			buf.WriteString(fmt.Sprintf("\t\tinput.%s = %s\n", f.Name, f.Default))
		}
		buf.WriteString(fmt.Sprintf("\t\tif v := q.Get(%q); v != \"\" {\n", queryName))
		buf.WriteString("\t\t\tparsed := v == \"true\" || v == \"1\"\n")
		if f.IsPointer {
			buf.WriteString(fmt.Sprintf("\t\t\tinput.%s = &parsed\n", f.Name))
		} else {
			buf.WriteString(fmt.Sprintf("\t\t\tinput.%s = parsed\n", f.Name))
		}
		buf.WriteString("\t\t}\n")

	case f.Type == "string":
		if f.Default != "" {
			// Use query value if present, otherwise default
			buf.WriteString(fmt.Sprintf("\t\tif v := q.Get(%q); v != \"\" {\n", queryName))
			buf.WriteString(fmt.Sprintf("\t\t\tinput.%s = v\n", f.Name))
			buf.WriteString("\t\t} else {\n")
			buf.WriteString(fmt.Sprintf("\t\t\tinput.%s = %q\n", f.Name, f.Default))
			buf.WriteString("\t\t}\n")
		} else {
			buf.WriteString(fmt.Sprintf("\t\tinput.%s = q.Get(%q)\n", f.Name, queryName))
		}

	case f.IsSlice && f.SplitChar != "":
		buf.WriteString(fmt.Sprintf("\t\tif v := q.Get(%q); v != \"\" {\n", queryName))
		buf.WriteString(fmt.Sprintf("\t\t\tinput.%s = strings.Split(v, %q)\n", f.Name, f.SplitChar))
		buf.WriteString("\t\t}\n")

	case strings.HasPrefix(f.Type, "*time.Time"):
		buf.WriteString(fmt.Sprintf("\t\tif t, ok := shared.ParseDateParam(q.Get(%q)); ok {\n", queryName))
		buf.WriteString(fmt.Sprintf("\t\t\tinput.%s = &t\n", f.Name))
		buf.WriteString("\t\t}\n")
	}
}

func serviceVarName(serviceName string) string {
	// Map service names to their global variable names
	switch serviceName {
	case "ActivityService":
		return "activityService"
	case "GearService":
		return "gearService"
	case "StatsService":
		return "statsService"
	case "DashboardService":
		return "dashboardService"
	case "SegmentsService":
		return "segmentsService"
	case "PhotosService":
		return "photosService"
	case "MaintenanceService":
		return "maintenanceService"
	case "ChallengesService":
		return "challengesService"
	default:
		// Default: lowercase first letter
		return strings.ToLower(serviceName[:1]) + serviceName[1:]
	}
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
