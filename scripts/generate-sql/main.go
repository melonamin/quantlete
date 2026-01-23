// Code generation script for SQL queries.
// Parses schema/queries/*.sql files and generates:
// - internal/storage/queries.gen.go (Go query functions)
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

// Column represents a column in a SELECT query.
type Column struct {
	SQLName  string // Original column name (e.g., "athlete_id")
	GoName   string // PascalCase name (e.g., "AthleteID")
	GoType   string // Go type (e.g., "int64")
	JSONName string // JSON tag name (e.g., "athlete_id")
}

// Param represents a query parameter.
type Param struct {
	Position string // Parameter position (e.g., "1", "2")
	Name     string // Inferred parameter name (e.g., "athleteID")
	GoType   string // Go type (e.g., "int64")
}

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
	Columns    []Column // Parsed columns from SELECT (for Go generation)
	ParamInfos []Param  // Inferred parameter info (for Go generation)
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
	// gearQueryRe matches queries that operate on the gear table (gear IDs are strings like "b12345")
	// Uses word boundaries to avoid false positives (e.g., "gearbox" or "some_gear_table")
	gearQueryRe = regexp.MustCompile(`(?i)\b(from\s+gear|from\s+v_gear|update\s+gear)\b`)
)

func parseQueryFile(path string) ([]Query, error) {
	file, err := os.Open(path) //nolint:gosec // G304: input path is from controlled source
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

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
				current.Columns = parseSelectColumns(current.SQL)
				current.ParamInfos = inferParamInfo(current.SQL, current.Params)
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
		current.Columns = parseSelectColumns(current.SQL)
		current.ParamInfos = inferParamInfo(current.SQL, current.Params)
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
	// Sort numerically, not alphabetically (otherwise "10" < "2")
	sort.Slice(params, func(i, j int) bool {
		var a, b int
		fmt.Sscanf(params[i], "%d", &a)
		fmt.Sscanf(params[j], "%d", &b)
		return a < b
	})
	return params
}

// parseSelectColumns extracts columns from a SELECT statement or RETURNING clause.
func parseSelectColumns(sql string) []Column {
	upperSQL := strings.ToUpper(sql)

	// First, try to parse SELECT ... FROM
	selectIdx := strings.Index(upperSQL, "SELECT")
	if selectIdx != -1 {
		fromIdx := findFromOutsideParens(upperSQL, selectIdx+6)
		if fromIdx != -1 {
			columnPart := strings.TrimSpace(sql[selectIdx+6 : fromIdx])
			return parseColumnList(columnPart, sql)
		}
	}

	// If no SELECT, try to parse RETURNING clause (for UPDATE/INSERT/DELETE ... RETURNING)
	returningIdx := strings.Index(upperSQL, "RETURNING")
	if returningIdx != -1 {
		// RETURNING clause goes to end of statement (or semicolon)
		columnPart := strings.TrimSpace(sql[returningIdx+9:])
		// Remove trailing semicolon if present
		columnPart = strings.TrimSuffix(strings.TrimSpace(columnPart), ";")
		return parseColumnList(columnPart, sql)
	}

	return nil
}

// parseColumnList parses a comma-separated list of columns.
func parseColumnList(columnPart, fullSQL string) []Column {
	columns := splitColumns(columnPart)

	var result []Column
	for _, col := range columns {
		col = strings.TrimSpace(col)
		if col == "" {
			continue
		}

		c := parseColumn(col, fullSQL)
		if c.SQLName != "" {
			result = append(result, c)
		}
	}

	return result
}

// splitColumns splits a column list by commas, respecting parentheses.
func splitColumns(s string) []string {
	var result []string
	var current strings.Builder
	depth := 0

	for _, r := range s {
		switch r {
		case '(':
			depth++
			current.WriteRune(r)
		case ')':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				result = append(result, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}

// parseColumn parses a single column expression.
// fullSQL is passed for context (e.g., to detect gear table queries).
func parseColumn(col string, fullSQL string) Column {
	col = strings.TrimSpace(col)

	// Normalize whitespace (replace newlines and multiple spaces with single space)
	col = strings.Join(strings.Fields(col), " ")

	// Check for AS alias (must be outside of parentheses)
	asIdx := findASOutsideParens(col)
	if asIdx != -1 {
		alias := strings.TrimSpace(col[asIdx+4:])
		// Remove any trailing comments or extra stuff
		if spaceIdx := strings.IndexAny(alias, " \t\n"); spaceIdx != -1 {
			alias = alias[:spaceIdx]
		}
		return Column{
			SQLName:  alias,
			GoName:   toPascalCase(alias),
			GoType:   inferGoType(alias, col, fullSQL),
			JSONName: alias,
		}
	}

	// Check for table.column format (but not inside parentheses)
	if !strings.HasPrefix(col, "(") {
		if dotIdx := strings.LastIndex(col, "."); dotIdx != -1 {
			name := strings.TrimSpace(col[dotIdx+1:])
			return Column{
				SQLName:  name,
				GoName:   toPascalCase(name),
				GoType:   inferGoType(name, col, fullSQL),
				JSONName: name,
			}
		}
	}

	// Simple column name
	return Column{
		SQLName:  col,
		GoName:   toPascalCase(col),
		GoType:   inferGoType(col, col, fullSQL),
		JSONName: col,
	}
}

// findFromOutsideParens finds "FROM" that is not inside parentheses.
// Returns the index relative to the start of the string, or -1 if not found.
func findFromOutsideParens(upperSQL string, startIdx int) int {
	depth := 0
	for i := startIdx; i < len(upperSQL)-4; i++ {
		switch upperSQL[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		// Look for FROM preceded by whitespace/newline (not part of another word)
		if depth == 0 && i+4 <= len(upperSQL) {
			if (i == startIdx || isWhitespace(upperSQL[i-1])) && upperSQL[i:i+4] == "FROM" {
				// Check it's followed by whitespace (not FROMAGE or something)
				if i+4 == len(upperSQL) || isWhitespace(upperSQL[i+4]) {
					return i
				}
			}
		}
	}
	return -1
}

func isWhitespace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// findASOutsideParens finds the position of " AS " that is not inside parentheses.
// Returns -1 if not found.
func findASOutsideParens(col string) int {
	upperCol := strings.ToUpper(col)
	depth := 0
	for i := 0; i < len(upperCol)-3; i++ {
		switch upperCol[i] {
		case '(':
			depth++
		case ')':
			depth--
		}
		if depth == 0 && i+4 <= len(upperCol) && upperCol[i:i+4] == " AS " {
			return i
		}
	}
	return -1
}

// inferGoType determines the Go type for a column based on its name and expression.
// fullSQL is provided for additional context (e.g., to detect gear table queries).
func inferGoType(name, expr, fullSQL string) string {
	lowerName := strings.ToLower(name)
	lowerExpr := strings.ToLower(expr)

	// Check if wrapped in COALESCE - makes it non-nullable
	hasCoalesce := strings.Contains(lowerExpr, "coalesce(")

	// Check if this is a gear table query (gear IDs are strings like "b12345")
	isGearQuery := gearQueryRe.MatchString(fullSQL)

	// ID fields - gear IDs are strings (Strava format: "b12345")
	if lowerName == "id" {
		if isGearQuery {
			return "string"
		}
		return "int64"
	}
	if lowerName == "gear_id" {
		return "string"
	}
	// zone_def_id is a composite text key like "Run:2024-01-01"
	if lowerName == "zone_def_id" {
		return "string"
	}
	if strings.HasSuffix(lowerName, "_id") {
		return "int64"
	}

	// Zone distribution seconds fields use int (not int64) because per-activity
	// and per-week durations fit comfortably within 32-bit range (~68 years max).
	if strings.HasPrefix(lowerName, "seconds_z") || lowerName == "total_seconds" {
		return "int"
	}

	// Count fields - athlete_effort_count is nullable
	if lowerName == "count" || lowerName == "activity_count" || lowerName == "segment_count" ||
		strings.HasPrefix(lowerName, "total_") && !strings.Contains(lowerName, "distance") && !strings.Contains(lowerName, "elevation") && !strings.Contains(lowerName, "time") {
		return "int"
	}
	// Nullable count fields
	if strings.HasSuffix(lowerName, "_count") {
		if hasCoalesce {
			return "int"
		}
		return "*int"
	}

	// Rank fields - nullable integers
	if strings.HasSuffix(lowerName, "_rank") || lowerName == "pr_rank" {
		if hasCoalesce {
			return "int"
		}
		return "*int"
	}

	// Climb category
	if lowerName == "climb_category" {
		return "int"
	}

	// Workout type
	if lowerName == "workout_type" {
		return "*int"
	}

	// Time duration fields (seconds) - some are nullable
	if lowerName == "moving_time" || lowerName == "elapsed_time" || lowerName == "total_time" {
		return "int"
	}
	// Nullable time fields
	if strings.HasSuffix(lowerName, "_elapsed_time") || strings.HasSuffix(lowerName, "_time") && !strings.Contains(lowerName, "date") {
		if hasCoalesce {
			return "int"
		}
		return "*int"
	}

	// Date/time fields - some are nullable
	if strings.HasSuffix(lowerName, "_at") || lowerName == "date" || lowerName == "day" ||
		lowerName == "month" || lowerName == "year" || lowerName == "week" || lowerName == "week_start" {
		return "string" // Dates as strings in SQLite
	}
	// Nullable date fields (athlete_pr_date, etc.)
	if strings.HasSuffix(lowerName, "_date") && lowerName != "start_date" && lowerName != "start_date_local" {
		if hasCoalesce {
			return "string"
		}
		return "*string"
	}
	// Non-nullable start_date fields
	if strings.HasPrefix(lowerName, "start_date") {
		return "string"
	}

	// Boolean fields
	if lowerName == "commute" || lowerName == "private" || lowerName == "trainer" ||
		lowerName == "starred" || lowerName == "retired" || lowerName == "is_primary" ||
		strings.HasPrefix(lowerName, "is_") || strings.HasPrefix(lowerName, "has_") {
		return "bool"
	}

	// Coordinate fields - nullable unless COALESCE
	if strings.HasSuffix(lowerName, "_lat") || strings.HasSuffix(lowerName, "_lng") ||
		lowerName == "start_lat" || lowerName == "start_lng" ||
		lowerName == "end_lat" || lowerName == "end_lng" {
		if hasCoalesce {
			return "float64"
		}
		return "*float64"
	}

	// Nullable numeric fields - power, HR, cadence, price, etc.
	if strings.Contains(lowerName, "watts") || strings.Contains(lowerName, "heartrate") ||
		strings.Contains(lowerName, "cadence") || lowerName == "kilojoules" ||
		lowerName == "calories" || lowerName == "suffer_score" ||
		lowerName == "average_grade" || lowerName == "maximum_grade" ||
		lowerName == "elev_high" || lowerName == "elev_low" ||
		lowerName == "max_heartrate" || lowerName == "purchase_price" {
		if hasCoalesce {
			return "float64"
		}
		return "*float64"
	}

	// Distance and elevation - often with COALESCE
	if strings.Contains(lowerName, "distance") || strings.Contains(lowerName, "elevation") ||
		lowerName == "total_elevation_gain" || lowerName == "elevation_high" || lowerName == "elevation_low" {
		return "float64"
	}

	// Training load metrics
	if lowerName == "tss" || lowerName == "ctl" || lowerName == "atl" || lowerName == "tsb" ||
		lowerName == "normalized_power" || lowerName == "intensity_factor" || lowerName == "ftp_used" {
		return "float64"
	}

	// Speed fields
	if strings.Contains(lowerName, "speed") {
		return "float64"
	}

	// Numeric aggregates (SUM, AVG, etc.)
	if strings.Contains(lowerExpr, "sum(") || strings.Contains(lowerExpr, "avg(") ||
		strings.Contains(lowerExpr, "min(") || strings.Contains(lowerExpr, "max(") {
		return "float64"
	}

	// COUNT
	if strings.Contains(lowerExpr, "count(") {
		return "int"
	}

	// Default to string
	return "string"
}

// inferParamInfo creates parameter info with inferred names and types.
func inferParamInfo(sql string, params []string) []Param {
	result := make([]Param, len(params))

	// Build a map of param position to context
	paramContexts := make(map[string]string)
	lines := strings.Split(sql, "\n")
	for _, line := range lines {
		for _, p := range params {
			placeholder := "?" + p
			if strings.Contains(line, placeholder) {
				paramContexts[p] = line
			}
		}
	}

	for i, p := range params {
		ctx := paramContexts[p]
		result[i] = Param{
			Position: p,
			Name:     inferParamName(p, ctx),
			GoType:   inferParamType(p, ctx, sql),
		}
	}

	return result
}

// inferParamName determines a parameter name from its context.
func inferParamName(pos, context string) string {
	lowerCtx := strings.ToLower(context)
	placeholder := "?" + pos

	// Common patterns
	patterns := []struct {
		contains string
		name     string
	}{
		{"athlete_id = " + placeholder, "athleteID"},
		{"athlete_id=" + placeholder, "athleteID"},
		{"sport_type = " + placeholder, "sportType"},
		{"sport_type=" + placeholder, "sportType"},
		{"gear_id = " + placeholder, "gearID"},
		{"gear_id=" + placeholder, "gearID"},
		{"activity_id = " + placeholder, "activityID"},
		{"activity_id=" + placeholder, "activityID"},
		{"segment_id = " + placeholder, "segmentID"},
		{"segment_id=" + placeholder, "segmentID"},
		{"start_date >= " + placeholder, "after"},
		{"start_date <= " + placeholder, "before"},
		{"start_date > " + placeholder, "after"},
		{"start_date < " + placeholder, "before"},
		{"day >= " + placeholder, "after"},
		{"day <= " + placeholder, "before"},
		{"limit " + placeholder, "limit"},
		{"offset " + placeholder, "offset"},
		{"year = " + placeholder, "year"},
		{"month = " + placeholder, "month"},
		{"country = " + placeholder, "country"},
		{"activity_type = " + placeholder, "activityType"},
	}

	for _, p := range patterns {
		if strings.Contains(lowerCtx, p.contains) {
			return p.name
		}
	}

	// Default to param + position
	return "arg" + pos
}

// inferParamType determines a parameter type from its context.
// fullSQL is provided for additional context (e.g., to detect gear table queries).
func inferParamType(pos, context, fullSQL string) string {
	lowerCtx := strings.ToLower(context)
	placeholder := "?" + pos

	// Check if this is a gear table query (gear IDs are strings like "b12345")
	isGearQuery := gearQueryRe.MatchString(fullSQL)

	// For INSERT statements, infer type from column name at corresponding position
	if insertType := inferInsertParamType(pos, fullSQL); insertType != "" {
		return insertType
	}

	// Gear ID is a string (Strava format: "b12345")
	if strings.Contains(lowerCtx, "gear_id = "+placeholder) || strings.Contains(lowerCtx, "gear_id="+placeholder) {
		return "string"
	}

	// Other _id fields (athlete_id, activity_id, etc.) are always int64
	if strings.Contains(lowerCtx, "_id = "+placeholder) || strings.Contains(lowerCtx, "_id="+placeholder) {
		return "int64"
	}

	// In gear queries, bare id (g.id or gear.id) is a string
	if isGearQuery {
		if strings.Contains(lowerCtx, "id = "+placeholder) || strings.Contains(lowerCtx, "id="+placeholder) {
			return "string"
		}
	}

	// Other bare id fields are int64
	if strings.Contains(lowerCtx, "id = "+placeholder) {
		return "int64"
	}

	// Pagination
	if strings.Contains(lowerCtx, "limit "+placeholder) || strings.Contains(lowerCtx, "offset "+placeholder) {
		return "int64"
	}

	// Year
	if strings.Contains(lowerCtx, "year = "+placeholder) || strings.Contains(lowerCtx, "year="+placeholder) {
		return "int"
	}

	// Default to string (dates, sport_type, etc.)
	return "string"
}

// inferInsertParamType infers parameter type from INSERT statement column names.
// Returns empty string if not an INSERT or column not found.
func inferInsertParamType(pos, fullSQL string) string {
	upperSQL := strings.ToUpper(fullSQL)

	// Check if this is an INSERT statement
	insertIdx := strings.Index(upperSQL, "INSERT INTO")
	if insertIdx == -1 {
		return ""
	}

	// Find the column list: INSERT INTO table_name (col1, col2, ...)
	openParen := strings.Index(fullSQL[insertIdx:], "(")
	if openParen == -1 {
		return ""
	}
	openParen += insertIdx

	closeParen := strings.Index(fullSQL[openParen:], ")")
	if closeParen == -1 {
		return ""
	}
	closeParen += openParen

	// Extract column names
	columnPart := fullSQL[openParen+1 : closeParen]
	columns := strings.Split(columnPart, ",")
	for i := range columns {
		columns[i] = strings.TrimSpace(columns[i])
	}

	// Find which column this parameter corresponds to by counting placeholders in VALUES
	valuesIdx := strings.Index(upperSQL, "VALUES")
	if valuesIdx == -1 {
		return ""
	}

	valuesParen := strings.Index(fullSQL[valuesIdx:], "(")
	if valuesParen == -1 {
		return ""
	}
	valuesParen += valuesIdx

	valuesCloseParen := strings.Index(fullSQL[valuesParen:], ")")
	if valuesCloseParen == -1 {
		return ""
	}
	valuesCloseParen += valuesParen

	// Extract values and find parameter position
	valuesPart := fullSQL[valuesParen+1 : valuesCloseParen]
	values := strings.Split(valuesPart, ",")

	placeholder := "?" + pos
	for i, v := range values {
		if strings.TrimSpace(v) == placeholder && i < len(columns) {
			return inferTypeFromColumnName(columns[i])
		}
	}

	return ""
}

// inferTypeFromColumnName infers Go type from a column name.
func inferTypeFromColumnName(colName string) string {
	lowerName := strings.ToLower(strings.TrimSpace(colName))

	// ID fields
	if lowerName == "athlete_id" || lowerName == "activity_id" || lowerName == "segment_id" {
		return "int64"
	}
	if lowerName == "gear_id" {
		return "string"
	}

	// Zone distribution seconds fields
	if strings.HasPrefix(lowerName, "seconds_z") || lowerName == "total_seconds" {
		return "int"
	}

	// Timestamp fields
	if strings.HasSuffix(lowerName, "_at") {
		return "SQLiteTime"
	}

	// Default to string
	return "string"
}

// toPascalCase converts snake_case to PascalCase.
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			// Handle common abbreviations
			upper := strings.ToUpper(p)
			if upper == "ID" || upper == "URL" || upper == "API" || upper == "SQL" ||
				upper == "HTTP" || upper == "JSON" || upper == "XML" || upper == "TSS" ||
				upper == "CTL" || upper == "ATL" || upper == "TSB" || upper == "FTP" ||
				upper == "HR" || upper == "NP" || upper == "PR" || upper == "KOM" {
				parts[i] = upper
			} else {
				parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
			}
		}
	}
	return strings.Join(parts, "")
}

// Go code generation

// goQueryData holds the data for generating a single query.
type goQueryData struct {
	Query
	SQLConstName   string
	ParamSignature string
	ParamArgs      string
	ScanFields     string
	HasColumns     bool
}

// prepareGoQueryData prepares query data for template execution.
func prepareGoQueryData(q Query) goQueryData {
	data := goQueryData{
		Query:        q,
		SQLConstName: toLowerFirst(q.Name) + "SQL",
		HasColumns:   len(q.Columns) > 0 && q.ReturnType != ":exec",
	}

	// Build parameter signature
	var sigParts []string
	for _, p := range q.ParamInfos {
		sigParts = append(sigParts, fmt.Sprintf("%s %s", p.Name, p.GoType))
	}
	if len(sigParts) > 0 {
		data.ParamSignature = ", " + strings.Join(sigParts, ", ")
	}

	// Build parameter args
	var argParts []string
	for _, p := range q.ParamInfos {
		argParts = append(argParts, p.Name)
	}
	if len(argParts) > 0 {
		data.ParamArgs = ", " + strings.Join(argParts, ", ")
	}

	// Build scan fields
	var scanParts []string
	for _, c := range q.Columns {
		scanParts = append(scanParts, "&i."+c.GoName)
	}
	data.ScanFields = strings.Join(scanParts, ", ")

	return data
}

func toLowerFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

var goTemplate = template.Must(template.New("go").Parse(`// Code generated by scripts/generate-sql. DO NOT EDIT.

package storage

import (
	"context"
	"database/sql"
)

// DBTX is the interface for database operations (supports both *sql.DB and *sql.Tx).
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Queries provides generated query methods.
type Queries struct {
	db DBTX
}

// NewQueries creates a new Queries instance.
func NewQueries(db DBTX) *Queries {
	return &Queries{db: db}
}

// WithTx returns a new Queries instance that uses the provided transaction.
func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{db: tx}
}

{{range .Queries}}
// ============================================================================
// {{.Query.Name}}
// ============================================================================
{{if .HasColumns}}
// {{.Query.Name}}Row represents a row returned by {{.Query.Name}}.
type {{.Query.Name}}Row struct {
{{range .Query.Columns}}	{{.GoName}} {{.GoType}} ` + "`" + `json:"{{.JSONName}}"` + "`" + `
{{end}}}
{{end}}
{{if .Query.Comment}}// {{.Query.Name}} - {{.Query.Comment}}
{{end}}const {{.SQLConstName}} = ` + "`" + `{{.Query.SQL}}` + "`" + `
{{if eq .Query.ReturnType ":many"}}
// {{.Query.Name}} executes the query and returns all rows.
func (q *Queries) {{.Query.Name}}(ctx context.Context{{.ParamSignature}}) ([]{{.Query.Name}}Row, error) {
	rows, err := q.db.QueryContext(ctx, {{.SQLConstName}}{{.ParamArgs}})
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []{{.Query.Name}}Row
	for rows.Next() {
		var i {{.Query.Name}}Row
		if err := rows.Scan({{.ScanFields}}); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
{{else if eq .Query.ReturnType ":one"}}
// {{.Query.Name}} executes the query and returns a single row, or nil if not found.
func (q *Queries) {{.Query.Name}}(ctx context.Context{{.ParamSignature}}) (*{{.Query.Name}}Row, error) {
	row := q.db.QueryRowContext(ctx, {{.SQLConstName}}{{.ParamArgs}})
	var i {{.Query.Name}}Row
	err := row.Scan({{.ScanFields}})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}
{{else}}
// {{.Query.Name}} executes the query.
func (q *Queries) {{.Query.Name}}(ctx context.Context{{.ParamSignature}}) error {
	_, err := q.db.ExecContext(ctx, {{.SQLConstName}}{{.ParamArgs}})
	return err
}
{{end}}
{{end}}
`))

func generateGoCode(files []QueryFile, output string) error {
	// Prepare query data with computed fields
	var allQueries []goQueryData
	for _, file := range files {
		for _, q := range file.Queries {
			allQueries = append(allQueries, prepareGoQueryData(q))
		}
	}

	var buf bytes.Buffer
	if err := goTemplate.Execute(&buf, map[string]interface{}{
		"Queries": allQueries,
	}); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	// Format the Go code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Write unformatted for debugging
		if writeErr := os.WriteFile(output+".unformatted", buf.Bytes(), 0o644); writeErr != nil { //nolint:gosec // G306: debug file, permissions are fine
			return fmt.Errorf("writing unformatted: %w", writeErr)
		}
		return fmt.Errorf("formatting Go code: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil { //nolint:gosec // G301: directory for generated code
		return fmt.Errorf("creating directory: %w", err)
	}

	return os.WriteFile(output, formatted, 0o644) //nolint:gosec // G306: generated code file
}
