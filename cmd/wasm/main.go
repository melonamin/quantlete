//go:build js && wasm

// Package main provides a production WASM entry point that exposes the Go storage layer
// to JavaScript, using sql.js via go-sqlite3-js.
package main

import (
	"fmt"
	"sync"
	"syscall/js"

	_ "github.com/matrix-org/go-sqlite3-js"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// WasmBridge holds the service registry and runtime state for the WASM bridge
type WasmBridge struct {
	registry  *services.ServiceRegistry
	athleteID int64
}

// bridgeMu protects access to the global bridge variable.
// Use getBridge() and setBridge() for thread-safe access.
var bridgeMu sync.RWMutex

// bridge is the global bridge instance - single source of truth for all state.
// Access via getBridge() for reads and setBridge() for writes.
var bridge *WasmBridge

// getBridge returns the current bridge instance in a thread-safe manner.
func getBridge() *WasmBridge {
	bridgeMu.RLock()
	defer bridgeMu.RUnlock()
	return bridge
}

// setBridge sets the global bridge instance in a thread-safe manner.
func setBridge(b *WasmBridge) {
	bridgeMu.Lock()
	defer bridgeMu.Unlock()
	bridge = b
}

func main() {
	fmt.Println("Quantlete Go WASM Storage Layer (Production)")

	// Register all WASM functions (generated from //wasm:export comments)
	RegisterAll()

	fmt.Println("Go WASM storage functions registered")

	// Keep the Go program running
	select {}
}

// ============================================================================
// Initialization
// ============================================================================

//wasm:category Initialization

// initStorage initializes the database connection
// Called from JS: goStorage.init()
//
//wasm:export init
func initStorage(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("initStorage")

	// Open in-memory database (sql.js handles persistence externally)
	db, err := storage.OpenWasm()
	if err != nil {
		return errorJSON(err)
	}

	// Run migrations
	if err := storage.RunMigrations(db.Conn()); err != nil {
		return errorJSON(fmt.Errorf("running migrations: %w", err))
	}

	// Initialize bridge with service registry
	setBridge(&WasmBridge{
		registry: services.NewServiceRegistry(db),
	})

	return successJSON("Database initialized")
}

// setAthleteID sets the current athlete ID for queries
// Called from JS: goStorage.setAthleteId(id)
//
//wasm:export setAthleteId
func setAthleteID(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("setAthleteID")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing athlete ID"))
	}

	b := getBridge()
	if b == nil {
		return errorJSON(fmt.Errorf("storage not initialized - call init() first"))
	}

	id := int64(args[0].Int())
	// Note: This is technically a race, but acceptable since athlete ID is set once
	// during initialization and read-only thereafter. A full mutex would be overkill.
	b.athleteID = id
	return successJSON(fmt.Sprintf("Athlete ID set to %d", id))
}

// exportDatabase exports the database for OPFS persistence
// Called from JS: goStorage.exportDb()
//
//wasm:export exportDb
func exportDatabase(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("exportDatabase")

	// Get the sql.js database and call export()
	dbMap := js.Global().Get("_go_sqlite_dbs")
	jsDb := dbMap.Call("get", ":memory:")
	if !jsDb.Truthy() {
		return errorJSON(fmt.Errorf("database not found"))
	}

	// Call db.export() to get Uint8Array
	data := jsDb.Call("export")
	return data // Return the Uint8Array directly
}
