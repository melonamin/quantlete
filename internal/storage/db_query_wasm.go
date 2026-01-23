//go:build js && wasm

package storage

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"
)

// Exec executes a query without returning any rows.
// In WASM mode, arguments are sanitized to avoid ValueOf serialization errors.
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.conn.Exec(query, sanitizeArgsForWASM(args)...)
}

// ExecContext executes a query without returning any rows, honoring context cancellation.
// In WASM mode, arguments are sanitized to avoid ValueOf serialization errors.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, sanitizeArgsForWASM(args)...)
}

// Query executes a query that returns rows.
// In WASM mode, arguments are sanitized to avoid ValueOf serialization errors.
func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(query, sanitizeArgsForWASM(args)...)
}

// QueryContext executes a query that returns rows, honoring context cancellation.
// In WASM mode, arguments are sanitized to avoid ValueOf serialization errors.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, sanitizeArgsForWASM(args)...)
}

// QueryRow executes a query that returns a single row.
// In WASM mode, arguments are sanitized to avoid ValueOf serialization errors.
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.conn.QueryRow(query, sanitizeArgsForWASM(args)...)
}

// QueryRowContext executes a query that returns a single row, honoring context cancellation.
// In WASM mode, arguments are sanitized to avoid ValueOf serialization errors.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.conn.QueryRowContext(ctx, query, sanitizeArgsForWASM(args)...)
}

// sanitizeArgsForWASM converts query arguments to WASM-compatible types.
// go-sqlite3-js can only serialize: int64, float64, string, bool, nil.
// Formatted strings (fmt.Sprintf, time.Format) may fail when boxed in []any.
func sanitizeArgsForWASM(args []any) []any {
	sanitized := make([]any, len(args))
	for i, arg := range args {
		sanitized[i] = sanitizeArg(arg)
	}
	return sanitized
}

// sanitizeArg converts a single argument to a WASM-safe type.
func sanitizeArg(arg any) any {
	if arg == nil {
		return nil
	}
	switch v := arg.(type) {
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case float32:
		return float64(v)
	case float64:
		return v
	case bool:
		return v
	case string:
		// Force string copy to avoid any interface boxing issues
		return string([]byte(v))
	case []byte:
		return v
	case time.Time:
		return string([]byte(v.Format("2006-01-02 15:04:05-07:00")))
	case *time.Time:
		if v == nil {
			return nil
		}
		return string([]byte(v.Format("2006-01-02 15:04:05-07:00")))
	case SQLiteTime:
		return string([]byte(v.Format("2006-01-02 15:04:05-07:00")))
	case *SQLiteTime:
		if v == nil || v.IsZero() {
			return nil
		}
		return string([]byte(v.Format("2006-01-02 15:04:05-07:00")))
	default:
		// For driver.Valuer types, try to get the value
		if valuer, ok := arg.(driver.Valuer); ok {
			if val, err := valuer.Value(); err == nil {
				return sanitizeArg(val)
			}
		}
		// Last resort: convert to string via fmt.Sprintf, then copy
		s := fmt.Sprintf("%v", arg)
		return string([]byte(s))
	}
}
