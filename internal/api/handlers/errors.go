package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
)

// handleServiceError translates service errors to HTTP responses.
// Uses the shared Response wrapper for consistent API format.
func handleServiceError(w http.ResponseWriter, err error) {
	status := errorToStatus(err)
	if status == http.StatusInternalServerError {
		slog.Error("internal server error", "error", err)
	}
	shared.WriteJSONResponse(w, status, shared.ErrorMessage(services.ClientMessage(err)))
}

// errorToStatus maps service errors to HTTP status codes.
func errorToStatus(err error) int {
	switch services.ErrorKindFor(err) {
	case services.ErrorKindNotFound:
		return http.StatusNotFound
	case services.ErrorKindUnauthorized:
		return http.StatusUnauthorized
	case services.ErrorKindForbidden:
		return http.StatusForbidden
	case services.ErrorKindBadRequest:
		return http.StatusBadRequest
	case services.ErrorKindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// writeJSON writes a raw JSON response (used by webhooks and import handlers).
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}
