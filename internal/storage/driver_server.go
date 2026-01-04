//go:build !js || !wasm

package storage

import (
	_ "modernc.org/sqlite" // SQLite driver for server (pure Go)
)

const driverName = "sqlite"
