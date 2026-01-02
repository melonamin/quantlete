//go:build js && wasm

package storage

import (
	_ "github.com/matrix-org/go-sqlite3-js" // SQLite driver for WASM (bridges to sql.js)
)

const driverName = "sqlite3"
