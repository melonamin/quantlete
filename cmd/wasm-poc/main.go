//go:build js && wasm

// Package main provides a WASM entry point that exposes the Go storage layer
// to JavaScript, using sql.js via go-sqlite3-js.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"syscall/js"

	_ "github.com/matrix-org/go-sqlite3-js"
)

var db *sql.DB

func main() {
	fmt.Println("Quantlete Go WASM Storage Layer")

	// Register JavaScript-callable functions
	js.Global().Set("goInitDB", js.FuncOf(initDB))
	js.Global().Set("goExecSQL", js.FuncOf(execSQL))
	js.Global().Set("goQuerySQL", js.FuncOf(querySQL))
	js.Global().Set("goGetDashboardStats", js.FuncOf(getDashboardStats))
	js.Global().Set("goGetActivities", js.FuncOf(getActivities))
	js.Global().Set("goTest", js.FuncOf(testFunc))

	fmt.Println("Go WASM functions registered")

	// Keep the Go program running
	select {}
}

// Simple test function that doesn't touch the database
func testFunc(this js.Value, args []js.Value) interface{} {
	fmt.Println("testFunc called")
	return `{"ok":true,"message":"Test function works!"}`
}

// initDB initializes the database connection
// Called from JS: goInitDB()
func initDB(this js.Value, args []js.Value) interface{} {
	var err error
	db, err = sql.Open("sqlite3", ":memory:")
	if err != nil {
		return jsError(err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return jsError(err)
	}

	return jsSuccess("Database initialized")
}

// execSQL executes a SQL statement (INSERT, UPDATE, DELETE, CREATE, etc.)
// Called from JS: goExecSQL(sql, ...params)
func execSQL(this js.Value, args []js.Value) interface{} {
	if db == nil {
		return jsError(fmt.Errorf("database not initialized"))
	}
	if len(args) < 1 {
		return jsError(fmt.Errorf("missing SQL statement"))
	}

	sqlStr := args[0].String()
	params := make([]interface{}, len(args)-1)
	for i := 1; i < len(args); i++ {
		params[i-1] = jsValueToGo(args[i])
	}

	result, err := db.ExecContext(context.Background(), sqlStr, params...)
	if err != nil {
		return jsError(err)
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertId, _ := result.LastInsertId()

	return toJSValue(map[string]interface{}{
		"ok":             true,
		"rows_affected":  rowsAffected,
		"last_insert_id": lastInsertId,
	})
}

// querySQL executes a SQL query and returns results
// Called from JS: goQuerySQL(sql, ...params)
func querySQL(this js.Value, args []js.Value) (result interface{}) {
	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("querySQL panic: %v\n", r)
			result = jsError(fmt.Errorf("panic: %v", r))
		}
	}()

	fmt.Println("querySQL called")

	if db == nil {
		return jsError(fmt.Errorf("database not initialized"))
	}
	if len(args) < 1 {
		return jsError(fmt.Errorf("missing SQL statement"))
	}

	sqlStr := args[0].String()
	fmt.Printf("querySQL: %s\n", sqlStr)

	params := make([]interface{}, len(args)-1)
	for i := 1; i < len(args); i++ {
		params[i-1] = jsValueToGo(args[i])
	}

	fmt.Println("querySQL: executing query...")
	rows, err := db.QueryContext(context.Background(), sqlStr, params...)
	if err != nil {
		fmt.Printf("querySQL error: %v\n", err)
		return jsError(err)
	}
	defer rows.Close()
	fmt.Println("querySQL: query executed, getting columns...")

	columns, err := rows.Columns()
	if err != nil {
		fmt.Printf("querySQL columns error: %v\n", err)
		return jsError(err)
	}
	fmt.Printf("querySQL: columns = %v\n", columns)

	var results []map[string]interface{}
	rowCount := 0
	for rows.Next() {
		fmt.Printf("querySQL: processing row %d\n", rowCount)
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Printf("querySQL scan error: %v\n", err)
			return jsError(err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
		rowCount++
	}

	fmt.Printf("querySQL: returning %d rows\n", len(results))
	return toJSValue(map[string]interface{}{
		"ok":      true,
		"columns": columns,
		"rows":    results,
	})
}

// getDashboardStats returns dashboard statistics for an athlete
// Called from JS: goGetDashboardStats(athleteId)
func getDashboardStats(this js.Value, args []js.Value) (result interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("getDashboardStats panic: %v\n", r)
			result = jsError(fmt.Errorf("panic: %v", r))
		}
	}()

	fmt.Println("getDashboardStats called")

	if db == nil {
		return jsError(fmt.Errorf("database not initialized"))
	}
	if len(args) < 1 {
		return jsError(fmt.Errorf("missing athleteId"))
	}

	athleteId := args[0].Int()
	fmt.Printf("getDashboardStats: athleteId=%d\n", athleteId)

	var totalActivities int
	var totalDistance, totalElevation float64
	var totalTime int

	err := db.QueryRowContext(context.Background(), `
		SELECT
			COUNT(*) as total_activities,
			COALESCE(SUM(distance), 0) as total_distance,
			COALESCE(SUM(moving_time), 0) as total_time,
			COALESCE(SUM(total_elevation_gain), 0) as total_elevation
		FROM activities
		WHERE athlete_id = ?
	`, athleteId).Scan(&totalActivities, &totalDistance, &totalTime, &totalElevation)

	if err != nil {
		return jsError(err)
	}

	return toJSValue(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"total_activities": totalActivities,
			"total_distance":   totalDistance,
			"total_time":       totalTime,
			"total_elevation":  totalElevation,
		},
	})
}

// getActivities returns paginated activities for an athlete
// Called from JS: goGetActivities(athleteId, limit, offset, sportType?)
func getActivities(this js.Value, args []js.Value) (result interface{}) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("getActivities panic: %v\n", r)
			result = jsError(fmt.Errorf("panic: %v", r))
		}
	}()

	fmt.Println("getActivities called")

	if db == nil {
		return jsError(fmt.Errorf("database not initialized"))
	}
	if len(args) < 3 {
		return jsError(fmt.Errorf("missing required arguments: athleteId, limit, offset"))
	}

	athleteId := args[0].Int()
	limit := args[1].Int()
	offset := args[2].Int()
	fmt.Printf("getActivities: athleteId=%d, limit=%d, offset=%d\n", athleteId, limit, offset)

	var sportType string
	if len(args) > 3 && !args[3].IsNull() && !args[3].IsUndefined() {
		sportType = args[3].String()
	}

	// Build query
	query := `
		SELECT id, name, sport_type, distance, moving_time, total_elevation_gain, start_date
		FROM activities
		WHERE athlete_id = ?
	`
	params := []interface{}{athleteId}

	if sportType != "" {
		query += " AND sport_type = ?"
		params = append(params, sportType)
	}

	query += " ORDER BY start_date DESC LIMIT ? OFFSET ?"
	params = append(params, limit, offset)

	rows, err := db.QueryContext(context.Background(), query, params...)
	if err != nil {
		return jsError(err)
	}
	defer rows.Close()

	var activities []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, sportTypeVal, startDate string
		var distance, elevation float64
		var movingTime int

		if err := rows.Scan(&id, &name, &sportTypeVal, &distance, &movingTime, &elevation, &startDate); err != nil {
			return jsError(err)
		}

		activities = append(activities, map[string]interface{}{
			"id":                   id,
			"name":                 name,
			"sport_type":           sportTypeVal,
			"distance":             distance,
			"moving_time":          movingTime,
			"total_elevation_gain": elevation,
			"start_date":           startDate,
		})
	}

	return toJSValue(map[string]interface{}{
		"ok":   true,
		"data": activities,
	})
}

// Helper functions

// toJSValue converts a Go value to a JavaScript-compatible return value
// Returns a JSON string that JavaScript must parse
func toJSValue(v interface{}) string {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return `{"ok":false,"error":"JSON marshal error: ` + err.Error() + `"}`
	}
	return string(jsonBytes)
}

func jsError(err error) string {
	return toJSValue(map[string]interface{}{
		"ok":    false,
		"error": err.Error(),
	})
}

func jsSuccess(message string) string {
	return toJSValue(map[string]interface{}{
		"ok":      true,
		"message": message,
	})
}

func jsValueToGo(v js.Value) interface{} {
	switch v.Type() {
	case js.TypeNull, js.TypeUndefined:
		return nil
	case js.TypeBoolean:
		return v.Bool()
	case js.TypeNumber:
		return v.Float()
	case js.TypeString:
		return v.String()
	case js.TypeObject:
		// Check if it's a Uint8Array (for binary data)
		if v.Get("constructor").Get("name").String() == "Uint8Array" {
			length := v.Get("length").Int()
			data := make([]byte, length)
			js.CopyBytesToGo(data, v)
			return data
		}
		// Try to convert to JSON
		jsonStr := js.Global().Get("JSON").Call("stringify", v).String()
		var result interface{}
		json.Unmarshal([]byte(jsonStr), &result)
		return result
	default:
		return v.String()
	}
}
