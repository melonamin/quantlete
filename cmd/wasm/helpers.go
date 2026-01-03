//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

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
	if bridge == nil || bridge.db == nil {
		return fmt.Errorf("storage not initialized - call init() first")
	}
	return nil
}

// ensureAthleteID checks if athleteID is set
func ensureAthleteID() error {
	if bridge == nil || bridge.athleteID == 0 {
		return fmt.Errorf("athlete ID not set - call setAthleteId() first")
	}
	return nil
}
