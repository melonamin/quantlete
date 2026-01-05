// Code generation script for TypeScript types from Go structs using AST parsing.
//
// This script parses Go source files from:
//   - cmd/wasm/*.go (anonymous request structs)
//   - internal/services/*.go (service layer input/output types)
//   - internal/storage/*.go (database-facing structs)
//   - internal/api/handlers/*.go (API response structs)
//   - internal/strava/*.go (external API types)
//
// And generates TypeScript interfaces in web/src/lib/wasm/types.gen.ts that
// match the JSON serialization format used by the WASM bridge.
//
// Capabilities:
//   - Multi-line struct definitions
//   - Pointer types and optionality (json omitempty)
//   - Type aliases to basic types
//   - Slice and map types
//   - Package-qualified types (e.g., time.Time)
//
// Limitations:
//   - Generic types are not supported (will become 'unknown')
//   - Embedded struct fields are not flattened
//   - Inline anonymous structs become 'unknown'
//   - Only file-scope var declarations are detected for anonymous structs
//
// Run with: just generate-ts-types
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
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
	Package  string // package path for collision detection
	Comment  string
}

// ParserContext holds state for parsing and generation
type ParserContext struct {
	typeAliases   map[string]string   // Go type aliases to their base types
	knownStructs  map[string]bool     // all parsed struct names for cross-referencing
	unknownTypes  map[string][]string // unknown types mapped to their usage locations
	malformedTags []string            // malformed struct tags encountered during parsing
}

// NewParserContext creates a new parser context
func NewParserContext() *ParserContext {
	return &ParserContext{
		typeAliases:  make(map[string]string),
		knownStructs: make(map[string]bool),
		unknownTypes: make(map[string][]string),
	}
}

// skipSuffixes defines type name suffixes to exclude from generation
// These are typically internal implementation types, not API types
var skipSuffixes = []string{
	"Repository",
	"Handler",
	"Router",
	"Service",
}

// allowedConfigTypes are Config types that ARE part of the public API
// and should be generated despite the general "Config" exclusion rule
var allowedConfigTypes = map[string]bool{
	"TrainingGoalsConfig":   true,
	"GoalsSportConfig":      true,
	"HRZoneConfig":          true,
	"DashboardConfig":       true,
	"DashboardWidgetConfig": true,
	"NotificationConfig":    true,
	"ServiceConfig":         true,
	"EventConfig":           true,
}

// shouldSkipType determines if a type should be excluded from generation
func shouldSkipType(name string) bool {
	for _, suffix := range skipSuffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}

	// Skip Config types unless explicitly allowed
	if strings.HasSuffix(name, "Config") && !allowedConfigTypes[name] {
		return true
	}

	return false
}

// DirConfig specifies how to parse a directory
type DirConfig struct {
	path     string
	anonOnly bool
}

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project root: %v\n", err)
		os.Exit(1)
	}

	dirs := []DirConfig{
		{filepath.Join(root, "cmd", "wasm"), true},
		{filepath.Join(root, "internal", "services"), false},
		{filepath.Join(root, "internal", "api", "handlers"), false},
		{filepath.Join(root, "internal", "storage"), false},
		{filepath.Join(root, "internal", "notifications"), false},
		// NOTE: internal/strava is NOT included - those are Strava API wire types,
		// not our API output types. The frontend uses handler output types.
	}

	tsOutput := filepath.Join(root, "web", "src", "lib", "wasm", "types.gen.ts")

	ctx := NewParserContext()

	// Phase 1: Collect ALL type aliases from all directories first
	// This ensures aliases are resolved correctly regardless of parse order
	for _, dir := range dirs {
		if err := collectTypeAliases(dir.path, ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error collecting aliases from %s: %v\n", dir.path, err)
			os.Exit(1)
		}
	}

	// Phase 2: Parse structs with complete alias map
	var allStructs []GoStruct
	for _, dir := range dirs {
		structs, err := parseDir(dir.path, dir.anonOnly, ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", dir.path, err)
			os.Exit(1)
		}
		allStructs = append(allStructs, structs...)
	}

	// Sort by (Name, Package) with services preferred over storage for deterministic deduplication.
	// When types exist in both services and storage, we prefer services (the API layer).
	sort.Slice(allStructs, func(i, j int) bool {
		if allStructs[i].Name != allStructs[j].Name {
			return allStructs[i].Name < allStructs[j].Name
		}
		// Prefer services over storage
		iIsServices := allStructs[i].Package == "services"
		jIsServices := allStructs[j].Package == "services"
		if iIsServices != jIsServices {
			return iIsServices // services comes first
		}
		return allStructs[i].Package < allStructs[j].Package
	})

	// Detect collisions but only fail for non-services/storage pairs
	collisions := detectCollisions(allStructs)
	realCollisions := make(map[string][]string)
	for name, packages := range collisions {
		// Allow services/storage collisions (services takes precedence)
		if len(packages) == 2 {
			isServiceStorageCollision := (packages[0] == "services" && packages[1] == "storage") ||
				(packages[0] == "storage" && packages[1] == "services")
			if isServiceStorageCollision {
				continue // This is expected - services wraps storage types
			}
		}
		realCollisions[name] = packages
	}
	if len(realCollisions) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: Name collisions detected - types with same name in different packages:")
		for name, packages := range realCollisions {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", name, packages)
		}
		fmt.Fprintln(os.Stderr, "\nFix by renaming types to be unique, or exclude a directory from generation.")
		os.Exit(1)
	}

	// Deduplicate - first occurrence wins (deterministic due to sorting)
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
		ctx.knownStructs[s.Name] = true
	}

	if err := generateTypeScript(unique, tsOutput, ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating TypeScript: %v\n", err)
		os.Exit(1)
	}

	// Fail build if any malformed struct tags were encountered during parsing
	if len(ctx.malformedTags) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: Malformed struct tags encountered during parsing:")
		for _, tag := range ctx.malformedTags {
			fmt.Fprintf(os.Stderr, "  %s\n", tag)
		}
		fmt.Fprintln(os.Stderr, "\nFix the struct tag syntax in the source file.")
		os.Exit(1)
	}

	// Fail build if any unknown types were encountered
	if len(ctx.unknownTypes) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: Unknown types encountered (would generate 'unknown' in TypeScript):")
		for goType, usages := range ctx.unknownTypes {
			fmt.Fprintf(os.Stderr, "  %s (referenced by: %v)\n", goType, usages)
		}
		fmt.Fprintln(os.Stderr, "\nFix by adding the type's source directory to dirs, or marking the field as json:\"-\".")
		os.Exit(1)
	}

	fmt.Printf("Generated: %s (%d types)\n", tsOutput, len(unique))
}

// detectCollisions finds struct names that appear in multiple packages
func detectCollisions(structs []GoStruct) map[string][]string {
	nameToPackages := make(map[string][]string)
	for _, s := range structs {
		nameToPackages[s.Name] = append(nameToPackages[s.Name], s.Package)
	}

	collisions := make(map[string][]string)
	for name, packages := range nameToPackages {
		if len(packages) > 1 {
			// Deduplicate package list
			seen := make(map[string]bool)
			var unique []string
			for _, pkg := range packages {
				if !seen[pkg] {
					seen[pkg] = true
					unique = append(unique, pkg)
				}
			}
			if len(unique) > 1 {
				collisions[name] = unique
			}
		}
	}
	return collisions
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

// collectTypeAliases scans a directory for type aliases without parsing structs
func collectTypeAliases(dir string, ctx *ParserContext) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading directory: %w", err)
	}

	// Sort entries for deterministic processing
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if err := collectTypeAliasesFromFile(path, ctx); err != nil {
			return fmt.Errorf("collecting aliases from %s: %w", entry.Name(), err)
		}
	}

	return nil
}

func collectTypeAliasesFromFile(path string, ctx *ParserContext) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return fmt.Errorf("parsing file: %w", err)
	}

	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			// Check for type alias to basic type (e.g., type WidgetWidth int)
			if ident, ok := typeSpec.Type.(*ast.Ident); ok {
				if isBasicType(ident.Name) {
					ctx.typeAliases[typeSpec.Name.Name] = ident.Name
				}
			}
		}
	}

	return nil
}

func parseDir(dir string, anonOnly bool, ctx *ParserContext) ([]GoStruct, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	// Sort entries for deterministic output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var allStructs []GoStruct
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		structs, err := parseGoFile(path, anonOnly, ctx)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}
		allStructs = append(allStructs, structs...)
	}

	return allStructs, nil
}

func parseGoFile(path string, anonOnly bool, ctx *ParserContext) ([]GoStruct, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing file: %w", err)
	}

	fileName := filepath.Base(path)
	packageName := node.Name.Name
	var structs []GoStruct

	// Collect structs (aliases already collected in Phase 1)
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.GenDecl:
			// Handle type declarations: type StructName struct { ... }
			if x.Tok == token.TYPE && !anonOnly {
				for _, spec := range x.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					structType, ok := typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					name := typeSpec.Name.Name
					// Skip unexported types
					if !ast.IsExported(name) {
						continue
					}
					// Skip utility types using the extracted function
					if shouldSkipType(name) {
						continue
					}

					fields := extractFields(structType, ctx)
					if len(fields) > 0 {
						comment := extractComment(x.Doc)
						structs = append(structs, GoStruct{
							Name:     name,
							Fields:   fields,
							FileName: fileName,
							Package:  packageName,
							Comment:  comment,
						})
					}
				}
			}

			// Handle anonymous structs: var req struct { ... }
			if x.Tok == token.VAR {
				for _, spec := range x.Specs {
					valueSpec, ok := spec.(*ast.ValueSpec)
					if !ok || len(valueSpec.Names) == 0 {
						continue
					}

					structType, ok := valueSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					varName := valueSpec.Names[0].Name
					fields := extractFields(structType, ctx)
					if len(fields) > 0 {
						// Derive type name from context
						typeName := deriveTypeName(varName, findEnclosingFunc(node, fset, x.Pos()))
						structs = append(structs, GoStruct{
							Name:     typeName,
							Fields:   fields,
							FileName: fileName,
							Package:  packageName,
						})
					}
				}
			}

		}
		return true
	})

	return structs, nil
}

func extractFields(structType *ast.StructType, ctx *ParserContext) []GoField {
	var fields []GoField

	if structType.Fields == nil {
		return fields
	}

	for _, field := range structType.Fields.List {
		// Get JSON tag, tracking any malformed tags
		jsonName, hasOmitempty := parseJSONTag(field.Tag, ctx)
		if jsonName == "" || jsonName == "-" {
			continue
		}

		// Get type string
		goType := typeToString(field.Type)
		isPointer := strings.HasPrefix(goType, "*")

		// Get field name (handle embedded fields)
		var fieldName string
		if len(field.Names) > 0 {
			fieldName = field.Names[0].Name
		} else {
			// Embedded field - use type name
			fieldName = goType
			if idx := strings.LastIndex(fieldName, "."); idx >= 0 {
				fieldName = fieldName[idx+1:]
			}
			fieldName = strings.TrimPrefix(fieldName, "*")
		}

		fields = append(fields, GoField{
			GoName:   fieldName,
			GoType:   goType,
			JSONName: jsonName,
			Optional: isPointer || hasOmitempty,
		})
	}

	return fields
}

func parseJSONTag(tag *ast.BasicLit, ctx *ParserContext) (jsonName string, hasOmitempty bool) {
	if tag == nil {
		return "", false
	}

	// Recover from malformed struct tags and track them for error reporting
	defer func() {
		if r := recover(); r != nil {
			ctx.malformedTags = append(ctx.malformedTags, tag.Value)
			jsonName = ""
			hasOmitempty = false
		}
	}()

	// Remove backticks
	tagValue := strings.Trim(tag.Value, "`")

	// Parse struct tag
	structTag := reflect.StructTag(tagValue)
	jsonTag := structTag.Get("json")
	if jsonTag == "" {
		return "", false
	}

	// Split on comma for options
	parts := strings.Split(jsonTag, ",")
	name := parts[0]
	for _, part := range parts[1:] {
		if part == "omitempty" {
			hasOmitempty = true
		}
	}

	return name, hasOmitempty
}

func typeToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			// Slice
			return "[]" + typeToString(t.Elt)
		}
		// Array - treat as slice for TS purposes
		return "[]" + typeToString(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", typeToString(t.Key), typeToString(t.Value))
	case *ast.SelectorExpr:
		// Package.Type (e.g., time.Time, storage.Activity)
		return typeToString(t.X) + "." + t.Sel.Name
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		// Inline struct - cannot be modeled in TS
		return "struct{}"
	default:
		return "unknown"
	}
}

func isBasicType(name string) bool {
	switch name {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "string", "bool", "byte", "rune":
		return true
	}
	return false
}

func extractComment(doc *ast.CommentGroup) string {
	if doc == nil || len(doc.List) == 0 {
		return ""
	}
	// Get first line of comment
	text := doc.List[0].Text
	text = strings.TrimPrefix(text, "//")
	text = strings.TrimPrefix(text, "/*")
	text = strings.TrimSuffix(text, "*/")
	return strings.TrimSpace(text)
}

func findEnclosingFunc(file *ast.File, fset *token.FileSet, pos token.Pos) string {
	var funcName string
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Pos() < pos && pos < fn.End() {
				funcName = fn.Name.Name
			}
		}
		return true
	})
	return funcName
}

func deriveTypeName(varName, funcName string) string {
	if funcName != "" {
		name := toPascalCase(funcName)
		if varName == "req" {
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

	name := toPascalCase(varName)
	if name == "Req" {
		return "RequestInput"
	}
	return name + "Input"
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	var result []string

	for _, p := range parts {
		if len(p) == 0 {
			continue
		}

		upper := strings.ToUpper(p)
		if upper == "ID" || upper == "URL" || upper == "API" || upper == "HR" {
			result = append(result, upper)
			continue
		}

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

func goTypeToTS(goType string, ctx *ParserContext) string {
	isPointer := strings.HasPrefix(goType, "*")
	baseType := strings.TrimLeft(goType, "*")

	// Resolve type aliases
	if aliasBase, ok := ctx.typeAliases[baseType]; ok {
		baseType = aliasBase
	}

	var tsType string
	switch baseType {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "byte", "rune":
		tsType = "number"
	case "string":
		tsType = "string"
	case "bool":
		tsType = "boolean"
	case "time.Time", "SQLiteTime", "FlexTime":
		tsType = "string"
	case "interface{}", "any", "struct{}":
		tsType = "unknown"
	case "json.RawMessage":
		tsType = "unknown"
	default:
		if strings.HasPrefix(baseType, "[]") {
			elemType := strings.TrimPrefix(baseType, "[]")
			return goTypeToTS(elemType, ctx) + "[]"
		}
		if strings.HasPrefix(baseType, "map[") {
			closeBracket := strings.Index(baseType, "]")
			if closeBracket > 4 {
				keyType := baseType[4:closeBracket]
				valueType := baseType[closeBracket+1:]

				var tsKeyType string
				switch keyType {
				case "string":
					tsKeyType = "string"
				case "int", "int8", "int16", "int32", "int64",
					"uint", "uint8", "uint16", "uint32", "uint64":
					tsKeyType = "number"
				default:
					tsKeyType = "string"
				}
				return fmt.Sprintf("Record<%s, %s>", tsKeyType, goTypeToTS(valueType, ctx))
			}
		}

		typeName := baseType
		if idx := strings.LastIndex(baseType, "."); idx >= 0 {
			typeName = baseType[idx+1:]
		}
		if ctx.knownStructs[typeName] {
			tsType = typeName
		} else {
			// Track unknown types for error reporting at the end
			ctx.unknownTypes[goType] = append(ctx.unknownTypes[goType], typeName)
			tsType = "unknown"
		}
	}

	if isPointer {
		return tsType + " | null"
	}
	return tsType
}

func needsQuoting(name string) bool {
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_') {
			return true
		}
	}
	return false
}

func generateTypeScript(structs []GoStruct, output string, ctx *ParserContext) error {
	var buf bytes.Buffer

	buf.WriteString(`// Code generated by scripts/generate-ts-types (AST-based). DO NOT EDIT.
//
// This file contains TypeScript interfaces that match the JSON serialization
// format used by the Go WASM bridge. These types should be used for all
// communication between TypeScript and Go WASM.
//
// To regenerate: just generate-ts-types

`)

	// Group by file
	fileGroups := make(map[string][]GoStruct)
	for _, s := range structs {
		fileGroups[s.FileName] = append(fileGroups[s.FileName], s)
	}

	var fileNames []string
	for name := range fileGroups {
		fileNames = append(fileNames, name)
	}
	sort.Strings(fileNames)

	for _, fileName := range fileNames {
		fileStructs := fileGroups[fileName]

		buf.WriteString("// ============================================================================\n")
		buf.WriteString(fmt.Sprintf("// From %s\n", fileName))
		buf.WriteString("// ============================================================================\n\n")

		for _, s := range fileStructs {
			if s.Comment != "" {
				buf.WriteString(fmt.Sprintf("/** %s */\n", s.Comment))
			}
			buf.WriteString(fmt.Sprintf("export interface %s {\n", s.Name))

			for _, f := range s.Fields {
				tsType := goTypeToTS(f.GoType, ctx)
				optionalMark := ""
				if f.Optional {
					optionalMark = "?"
				}
				fieldName := f.JSONName
				if needsQuoting(fieldName) {
					fieldName = fmt.Sprintf("'%s'", fieldName)
				}
				buf.WriteString(fmt.Sprintf("  %s%s: %s\n", fieldName, optionalMark, tsType))
			}

			buf.WriteString("}\n\n")
		}
	}

	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(output, buf.Bytes(), 0o644) //nolint:gosec
}
