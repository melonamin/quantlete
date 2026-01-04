// Package shared provides common utilities used across multiple packages.
package shared

import (
	"time"
)

// DateParamFormat is the simple date format for query parameters (YYYY-MM-DD).
const DateParamFormat = "2006-01-02"

// ParseDateParam parses a date string from a query parameter.
// It tries RFC3339 first (e.g., "2024-01-15T10:30:00Z"), then falls back
// to the simple date format (e.g., "2024-01-15").
//
// Returns the parsed time and true if successful, or zero time and false if
// the string is empty or unparseable.
func ParseDateParam(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}

	// Try RFC3339 first (full timestamp with timezone)
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}

	// Fall back to simple date format
	if t, err := time.Parse(DateParamFormat, s); err == nil {
		return t, true
	}

	return time.Time{}, false
}
