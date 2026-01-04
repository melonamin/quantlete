//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
)

// Maximum allowed array sizes to prevent memory exhaustion
const (
	maxStreamDataSize     = 100000 // Max data points in a stream
	maxTrainingLoadDays   = 10000  // Max days of training load data
	maxBestEffortsPerSave = 1000   // Max best efforts per activity
	maxAlgorithmArraySize = 100000 // Max array size for algorithm functions
)

// recoverPanic catches panics and logs them
func recoverPanic(funcName string) {
	if r := recover(); r != nil {
		fmt.Printf("[%s] panic: %v\n", funcName, r)
	}
}

// validateArraySize checks that an array doesn't exceed the maximum allowed size
func validateArraySize(length, maxSize int, name string) error {
	if length > maxSize {
		return fmt.Errorf("%s exceeds maximum size: %d > %d", name, length, maxSize)
	}
	return nil
}

// toJSON converts a Go value to a JSON string.
// Used for custom response structures that don't fit the standard Response format.
// Prefer dataJSON/errorJSON/successJSON for standard responses.
func toJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return shared.ErrorMessage("JSON marshal error").ToJSON()
	}
	return string(b)
}

// dataJSON wraps data in a success response and returns as JSON string.
// Uses shared.Response for consistent response format.
func dataJSON(v interface{}) string {
	return shared.SuccessResponse(v).ToJSON()
}

// errorJSON returns an error response as JSON string.
// Uses shared.Response for consistent response format.
func errorJSON(err error) string {
	return shared.ErrorResponse(err).ToJSON()
}

// successJSON returns a success response with a message as JSON string.
// Uses shared.Response for consistent response format.
func successJSON(message string) string {
	return shared.SuccessMessage(message).ToJSON()
}

// jsArrayToFloat64 converts a JS array to []float64 with size validation
func jsArrayToFloat64(jsArr js.Value) ([]float64, error) {
	return jsArrayToFloat64WithLimit(jsArr, maxAlgorithmArraySize)
}

// jsArrayToFloat64WithLimit converts a JS array to []float64 with a custom size limit
func jsArrayToFloat64WithLimit(jsArr js.Value, maxSize int) ([]float64, error) {
	if jsArr.Type() != js.TypeObject {
		return nil, fmt.Errorf("expected array, got %s", jsArr.Type().String())
	}

	length := jsArr.Length()
	if length > maxSize {
		return nil, fmt.Errorf("array too large: %d elements exceeds maximum of %d", length, maxSize)
	}

	result := make([]float64, length)
	for i := 0; i < length; i++ {
		val := jsArr.Index(i)
		switch val.Type() {
		case js.TypeNumber:
			result[i] = val.Float()
		case js.TypeNull, js.TypeUndefined:
			result[i] = 0
		default:
			return nil, fmt.Errorf("element %d: expected number, got %s", i, val.Type().String())
		}
	}
	return result, nil
}

// ============================================================================
// Validation Functions
// ============================================================================

// ensureInitialized checks if the database is initialized
func ensureInitialized() error {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.DB() == nil {
		return fmt.Errorf("storage not initialized - call init() first")
	}
	return nil
}

// ensureAthleteID checks if athleteID is set
func ensureAthleteID() error {
	b := getBridge()
	if b == nil || b.athleteID == 0 {
		return fmt.Errorf("athlete ID not set - call setAthleteId() first")
	}
	return nil
}

// ============================================================================
// WASM Function Wrappers
// ============================================================================
//
// These wrappers reduce boilerplate in WASM functions by handling common
// concerns: panic recovery, bridge access, initialization checks, and context.
//
// Usage:
//   func myHandler(wc *WasmContext) interface{} {
//       result, err := wc.Registry.SomeService.DoThing(wc.Ctx, wc.AthleteID)
//       if err != nil {
//           return errorJSON(err)
//       }
//       return dataJSON(result)
//   }
//
//   var myFunction = wrapWasmAthlete("myFunction", myHandler)

// WasmContext provides context for WASM function execution.
// It encapsulates the bridge, registry, athlete context, and arguments.
type WasmContext struct {
	Bridge    *WasmBridge
	Registry  *services.ServiceRegistry
	AthleteID int64
	Ctx       context.Context
	Args      []js.Value
}

// WasmHandler handles a WASM call with context (no athlete required).
type WasmHandler func(*WasmContext) interface{}

// WasmHandlerAthlete handles a WASM call that requires athlete context.
type WasmHandlerAthlete func(*WasmContext) interface{}

// wrapWasm creates a wrapper for WASM functions that only need basic initialization.
// Use for functions that don't require athlete ID.
func wrapWasm(name string, fn WasmHandler) func(js.Value, []js.Value) interface{} {
	return func(this js.Value, args []js.Value) interface{} {
		defer recoverPanic(name)

		b := getBridge()
		if b == nil || b.registry == nil {
			return errorJSON(fmt.Errorf("storage not initialized"))
		}

		if err := ensureInitialized(); err != nil {
			return errorJSON(err)
		}

		wc := &WasmContext{
			Bridge:   b,
			Registry: b.registry,
			Ctx:      context.Background(),
			Args:     args,
		}

		return fn(wc)
	}
}

// wrapWasmAthlete creates a wrapper for WASM functions that require athlete context.
// This is the most common wrapper - use for functions that need athlete ID.
func wrapWasmAthlete(name string, fn WasmHandlerAthlete) func(js.Value, []js.Value) interface{} {
	return func(this js.Value, args []js.Value) interface{} {
		defer recoverPanic(name)

		b := getBridge()
		if b == nil || b.registry == nil {
			return errorJSON(fmt.Errorf("storage not initialized"))
		}

		if err := ensureInitialized(); err != nil {
			return errorJSON(err)
		}
		if err := ensureAthleteID(); err != nil {
			return errorJSON(err)
		}

		wc := &WasmContext{
			Bridge:    b,
			Registry:  b.registry,
			AthleteID: b.athleteID,
			Ctx:       context.Background(),
			Args:      args,
		}

		return fn(wc)
	}
}

// wrapWasmRaw creates a wrapper that only handles panic recovery.
// Use for functions that manage their own initialization (e.g., init, setAthleteId).
func wrapWasmRaw(name string, fn func(js.Value, []js.Value) interface{}) func(js.Value, []js.Value) interface{} {
	return func(this js.Value, args []js.Value) interface{} {
		defer recoverPanic(name)
		return fn(this, args)
	}
}

// ============================================================================
// Argument Parsing Helpers
// ============================================================================

// ArgString returns args[index] as string, or empty string if missing.
func (wc *WasmContext) ArgString(index int) string {
	if index >= len(wc.Args) {
		return ""
	}
	return wc.Args[index].String()
}

// ArgInt returns args[index] as int, or 0 if missing/invalid.
func (wc *WasmContext) ArgInt(index int) int {
	if index >= len(wc.Args) {
		return 0
	}
	return wc.Args[index].Int()
}

// ArgInt64 returns args[index] as int64, or 0 if missing/invalid.
func (wc *WasmContext) ArgInt64(index int) int64 {
	if index >= len(wc.Args) {
		return 0
	}
	// JS numbers are float64, so we need to convert
	return int64(wc.Args[index].Float())
}

// ArgFloat64 returns args[index] as float64, or 0 if missing/invalid.
func (wc *WasmContext) ArgFloat64(index int) float64 {
	if index >= len(wc.Args) {
		return 0
	}
	return wc.Args[index].Float()
}

// ArgBool returns args[index] as bool, or false if missing.
func (wc *WasmContext) ArgBool(index int) bool {
	if index >= len(wc.Args) {
		return false
	}
	return wc.Args[index].Bool()
}

// ArgJSON parses args[index] as JSON into the provided pointer.
// Returns error if missing or invalid JSON.
func (wc *WasmContext) ArgJSON(index int, v interface{}) error {
	if index >= len(wc.Args) {
		return fmt.Errorf("missing argument at index %d", index)
	}
	jsonStr := wc.Args[index].String()
	return json.Unmarshal([]byte(jsonStr), v)
}

// HasArg returns true if an argument exists at the given index.
func (wc *WasmContext) HasArg(index int) bool {
	return index < len(wc.Args)
}
