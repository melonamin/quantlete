//go:build js && wasm

package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/matrix-org/go-sqlite3-js" // SQLite driver for WASM (bridges to sql.js)
)

const driverName = "sqlite3"

// OpenWasm opens an in-memory SQLite database for WASM mode.
// The database is backed by sql.js in the browser.
func OpenWasm() (*DB, error) {
	conn, err := sql.Open(driverName, ":memory:")
	if err != nil {
		return nil, fmt.Errorf("opening wasm database: %w", err)
	}

	conn.SetMaxOpenConns(1)
	conn.SetMaxIdleConns(1)

	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("connecting to wasm database: %w", err)
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

// RunMigrations runs all pending migrations on the given connection.
//
// This is a convenience function for WASM mode only (build tag: js && wasm).
// It wraps the database connection and calls the standard Migrate() method.
//
// In server mode, migrations are run directly via db.Migrate() after opening
// the database file. In WASM mode, this function is called from JavaScript
// after initializing the sql.js database.
//
// The function is exposed here rather than as a method because the WASM bridge
// may need to call it before fully constructing the DB instance.
func RunMigrations(conn *sql.DB) error {
	db := &DB{conn: conn, path: ":memory:"}
	return db.Migrate()
}
