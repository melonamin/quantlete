package storage

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// JoinStrings is a small helper used across packages for building SQL placeholders.
func JoinStrings(strs []string, sep string) string {
	return joinStrings(strs, sep)
}

// SQLiteTime wraps time.Time to handle SQLite TEXT timestamp scanning.
type SQLiteTime struct {
	time.Time
}

// Scan implements sql.Scanner for SQLiteTime.
func (t *SQLiteTime) Scan(value any) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	case string:
		return t.parseString(v)
	case []byte:
		return t.parseString(string(v))
	default:
		return fmt.Errorf("cannot scan %T into SQLiteTime", value)
	}
}

func (t *SQLiteTime) parseString(s string) error {
	if s == "" {
		t.Time = time.Time{}
		return nil
	}

	// Strip monotonic clock reading if present (Go's default time.String() format)
	// Format: "2006-01-02 15:04:05.999999999 -0700 MST m=+0.000000001"
	if idx := strings.Index(s, " m="); idx != -1 {
		s = s[:idx]
	}

	// Try common formats
	formats := []string{
		"2006-01-02 15:04:05-07:00",               // SQLite with timezone
		"2006-01-02 15:04:05+00:00",               // SQLite with UTC
		"2006-01-02T15:04:05Z",                    // RFC3339 UTC
		"2006-01-02T15:04:05-07:00",               // RFC3339 with timezone
		"2006-01-02 15:04:05",                     // SQLite without timezone
		"2006-01-02",                              // SQLite date() function output
		"2006-01-02T15:04:05.000Z",                // RFC3339 with milliseconds
		"2006-01-02 15:04:05.999999999 -0700 MST", // Go default format (after stripping m=)
		time.RFC3339,
		time.RFC3339Nano,
	}

	var err error
	for _, format := range formats {
		t.Time, err = time.Parse(format, s)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("cannot parse %q as time", s)
}

// Value implements driver.Valuer for SQLiteTime.
func (t SQLiteTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Format("2006-01-02 15:04:05-07:00"), nil
}

// SQLiteTimePtr wraps a *time.Time into an SQLiteTime value suitable for SQL parameters.
// Returns nil if the pointer is nil, avoiding nil pointer dereference.
func SQLiteTimePtr(t *time.Time) any {
	if t == nil {
		return nil
	}
	return SQLiteTime{Time: *t}
}

// AddTimeRangeFilter appends time range conditions to a WHERE clause builder.
// It handles both start (>=) and end (<=) bounds, wrapping times in SQLiteTime.
//
// Example usage:
//
//	var conditions []string
//	var args []any
//	conditions, args = AddTimeRangeFilter(conditions, args, "start_date", filters.StartAfter, filters.StartBefore)
func AddTimeRangeFilter(conditions []string, args []any, column string, start, end *time.Time) ([]string, []any) {
	if start != nil {
		conditions = append(conditions, column+" >= ?")
		args = append(args, SQLiteTime{Time: *start})
	}
	if end != nil {
		conditions = append(conditions, column+" <= ?")
		args = append(args, SQLiteTime{Time: *end})
	}
	return conditions, args
}

// MaxSliceParamLength is the maximum number of elements allowed in a slice parameter.
// This prevents DoS attacks from excessively large IN clauses.
const MaxSliceParamLength = 100

// sliceParamPattern matches slice placeholders: /*SLICE:name*/?N
// The pattern is designed to only match explicitly marked placeholders in SQL queries,
// preventing any SQL injection through user input (comments are static in queries).
var sliceParamPattern = regexp.MustCompile(`/\*SLICE:\w+\*/\?(\d+)`)

// expandSliceParams expands slice parameters in SQL queries.
//
// This function handles placeholders like /*SLICE:name*/?N by replacing them with
// the appropriate number of ? placeholders and flattening the args array. It's used
// to support IN clauses with dynamic slice lengths while maintaining SQL safety.
//
// # Marker Format
//
// Slice parameters must be marked in SQL with the format: /*SLICE:name*/?N
//   - "name" is a descriptive identifier (alphanumeric + underscore)
//   - N is the 1-based parameter index in the args array
//
// # Supported Slice Types
//
//   - []string
//   - []int, []int64
//   - []float64
//   - []any
//
// # Error Conditions
//
//   - Empty slice: Returns error (IN () is invalid SQL syntax)
//   - Slice exceeds MaxSliceParamLength (100): Returns error to prevent DoS
//   - Reusing the same ?N index for multiple slices: Returns error
//   - Non-slice argument for a slice placeholder: Returns error
//   - Parameter index out of range: Returns error
//
// # Caller Recommendations
//
// When an empty slice is possible, callers should either:
//   - Pre-filter and short-circuit (e.g., return early if slice is empty)
//   - Use a different query structure (e.g., WHERE 1=0 for empty results)
//
// # Example
//
//	sql: "WHERE id IN (/*SLICE:ids*/?1)"
//	args: []any{[]int{1, 2, 3}}
//	returns: "WHERE id IN (?, ?, ?)", []any{1, 2, 3}, nil
//
// # Security
//
// This function is safe from SQL injection because:
//   - Only ? placeholders are inserted (never user values)
//   - The /*SLICE:...*/ markers are static in queries, not user-controlled
//   - All slice elements are added as bound parameters
func expandSliceParams(sql string, args []any) (string, []any, error) {
	matches := sliceParamPattern.FindAllStringSubmatchIndex(sql, -1)

	if len(matches) == 0 {
		return sql, args, nil
	}

	// Build new SQL and args
	var newSQL strings.Builder
	var newArgs []any
	lastEnd := 0

	// Track which arg positions are slices and their expansions
	slicePositions := make(map[int]int) // position -> length

	// First pass: identify slice positions and validate
	for _, match := range matches {
		posStr := sql[match[2]:match[3]]
		pos, err := strconv.Atoi(posStr)
		if err != nil {
			return "", nil, fmt.Errorf("invalid slice parameter index %q: %w", posStr, err)
		}

		if pos < 1 || pos > len(args) {
			return "", nil, fmt.Errorf("slice parameter ?%d out of range (have %d args)", pos, len(args))
		}

		// Reject repeated use of the same slice parameter
		if _, exists := slicePositions[pos]; exists {
			return "", nil, fmt.Errorf("slice parameter ?%d used more than once; this is not supported", pos)
		}

		arg := args[pos-1]
		length, err := sliceLength(arg)
		if err != nil {
			return "", nil, fmt.Errorf("parameter ?%d: %w", pos, err)
		}
		if length > MaxSliceParamLength {
			return "", nil, fmt.Errorf("parameter ?%d: slice length %d exceeds maximum %d", pos, length, MaxSliceParamLength)
		}
		if length == 0 {
			return "", nil, fmt.Errorf("parameter ?%d: empty slice not allowed in IN clause", pos)
		}
		slicePositions[pos] = length
	}

	// Second pass: build new SQL
	for _, match := range matches {
		// Append SQL before this match
		newSQL.WriteString(sql[lastEnd:match[0]])

		// Get position and length
		posStr := sql[match[2]:match[3]]
		pos, _ := strconv.Atoi(posStr) // Already validated in first pass
		length := slicePositions[pos]

		// Write expanded placeholders
		placeholders := make([]string, length)
		for i := range placeholders {
			placeholders[i] = "?"
		}
		newSQL.WriteString(strings.Join(placeholders, ", "))

		lastEnd = match[1]
	}
	newSQL.WriteString(sql[lastEnd:])

	// Build new args array with slices expanded
	for i, arg := range args {
		pos := i + 1
		if _, isSlice := slicePositions[pos]; isSlice {
			// Expand slice elements
			expanded, _ := expandSlice(arg)
			newArgs = append(newArgs, expanded...)
		} else {
			newArgs = append(newArgs, arg)
		}
	}

	return newSQL.String(), newArgs, nil
}

// sliceLength returns the length of a slice argument.
func sliceLength(arg any) (int, error) {
	switch v := arg.(type) {
	case []string:
		return len(v), nil
	case []int:
		return len(v), nil
	case []int64:
		return len(v), nil
	case []float64:
		return len(v), nil
	case []any:
		return len(v), nil
	default:
		return 0, fmt.Errorf("expected slice, got %T", arg)
	}
}

// expandSlice converts a slice to []any.
func expandSlice(arg any) ([]any, error) {
	switch v := arg.(type) {
	case []string:
		result := make([]any, len(v))
		for i, s := range v {
			result[i] = s
		}
		return result, nil
	case []int:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	case []int64:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	case []float64:
		result := make([]any, len(v))
		for i, n := range v {
			result[i] = n
		}
		return result, nil
	case []any:
		return v, nil
	default:
		return nil, fmt.Errorf("expected slice, got %T", arg)
	}
}
