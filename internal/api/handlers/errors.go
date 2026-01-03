package handlers

import (
	"errors"
	"net/http"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
)

// handleServiceError translates service errors to HTTP responses.
// Uses the shared Response wrapper for consistent API format.
func handleServiceError(w http.ResponseWriter, err error) {
	status := errorToStatus(err)
	shared.WriteErrorWithStatus(w, status, err)
}

// errorToStatus maps service errors to HTTP status codes.
func errorToStatus(err error) int {
	switch {
	case errors.Is(err, services.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, services.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, services.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, services.ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, services.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
