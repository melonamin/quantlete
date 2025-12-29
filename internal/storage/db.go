// Package storage provides database access and persistence for the application.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "modernc.org/sqlite" // SQLite driver (pure Go)
)

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
	path string
	mu   sync.RWMutex
}

// Open opens or creates a SQLite database at the given path.
func Open(dataDir, dbFile string) (*DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	filename := dbFile
	if strings.TrimSpace(filename) == "" {
		filename = "stata.db"
	}
	dbPath := filepath.Join(dataDir, filename)

	dsn := dbPath
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(1) // SQLite works best with a single connection
	conn.SetMaxIdleConns(1)

	// Test connection
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	// Enable foreign keys (disabled by default in SQLite)
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	return db, nil
}

// OpenInMemory opens an in-memory SQLite database (for testing).
func OpenInMemory() (*DB, error) {
	conn, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("opening in-memory database: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("connecting to in-memory database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	return &DB{
		conn: conn,
		path: ":memory:",
	}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Conn returns the underlying sql.DB connection.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// Path returns the database file path.
func (db *DB) Path() string {
	return db.path
}

// Exec executes a query without returning any rows.
func (db *DB) Exec(query string, args ...any) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// ExecContext executes a query without returning any rows, honoring context cancellation.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows.
func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryContext executes a query that returns rows, honoring context cancellation.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row.
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.conn.QueryRow(query, args...)
}

// QueryRowContext executes a query that returns a single row, honoring context cancellation.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.conn.QueryRowContext(ctx, query, args...)
}
