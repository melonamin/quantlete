// Package storage provides database access and persistence for the application.
package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/marcboeker/go-duckdb" // DuckDB driver
)

// DB wraps a DuckDB database connection.
type DB struct {
	conn *sql.DB
	path string
	mu   sync.RWMutex
}

// Open opens or creates a DuckDB database at the given path.
func Open(dataDir string) (*DB, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("creating data directory: %w", err)
	}

	dbPath := filepath.Join(dataDir, "stata.duckdb")

	conn, err := sql.Open("duckdb", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(1) // DuckDB works best with a single connection
	conn.SetMaxIdleConns(1)

	// Test connection
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("connecting to database: %w", err)
	}

	db := &DB{
		conn: conn,
		path: dbPath,
	}

	return db, nil
}

// OpenInMemory opens an in-memory DuckDB database (for testing).
func OpenInMemory() (*DB, error) {
	conn, err := sql.Open("duckdb", "")
	if err != nil {
		return nil, fmt.Errorf("opening in-memory database: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("connecting to in-memory database: %w", err)
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

// Query executes a query that returns rows.
func (db *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// QueryRow executes a query that returns a single row.
func (db *DB) QueryRow(query string, args ...any) *sql.Row {
	return db.conn.QueryRow(query, args...)
}
