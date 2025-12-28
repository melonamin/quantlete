package storage

import (
	"database/sql/driver"
	"fmt"
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
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time.Format("2006-01-02 15:04:05-07:00"), nil
}
