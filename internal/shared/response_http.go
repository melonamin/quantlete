//go:build !js || !wasm

package shared

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// WriteSuccess writes a successful JSON response to the HTTP response writer.
func WriteSuccess(w http.ResponseWriter, data any) {
	WriteJSONResponse(w, http.StatusOK, SuccessResponse(data))
}

// WriteMessage writes a successful message response to the HTTP response writer.
func WriteMessage(w http.ResponseWriter, msg string) {
	WriteJSONResponse(w, http.StatusOK, SuccessMessage(msg))
}

// WriteErrorWithStatus writes an error response with the given HTTP status code.
func WriteErrorWithStatus(w http.ResponseWriter, status int, err error) {
	errMsg := err.Error()

	// For internal errors, log the actual error and return a generic message
	if status == http.StatusInternalServerError {
		slog.Error("internal server error", "error", err)
		errMsg = "internal server error"
	}

	WriteJSONResponse(w, status, Response{OK: false, Error: errMsg})
}

// WriteJSONResponse writes a JSON response with the given status code.
func WriteJSONResponse(w http.ResponseWriter, status int, resp Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}
