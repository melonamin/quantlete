package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

// NotificationsHandler handles notification-related HTTP requests.
type NotificationsHandler struct {
	notifications *services.NotificationService
	strava        *strava.Client
}

// NewNotificationsHandler creates a new notifications handler.
func NewNotificationsHandler(notifications *services.NotificationService, stravaClient *strava.Client) *NotificationsHandler {
	return &NotificationsHandler{
		notifications: notifications,
		strava:        stravaClient,
	}
}

// TestRequest is the request body for testing a single notification service.
type TestRequest struct {
	ServiceID string `json:"serviceId"`
}

// TestResult is the response for a notification test.
type TestResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// Test handles POST /api/v1/notifications/test
// Tests a single notification service by sending a test message.
func (h *NotificationsHandler) Test(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	if req.ServiceID == "" {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("serviceId is required"))
		return
	}

	err := h.notifications.TestNotification(r.Context(), athlete.ID, req.ServiceID)
	if err != nil {
		// Check if it's a business error (bad request, not found, etc.)
		if errors.Is(err, services.ErrBadRequest) || errors.Is(err, services.ErrNotFound) {
			shared.WriteSuccess(w, TestResult{
				Success: false,
				Error:   err.Error(),
			})
			return
		}
		// Internal error
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to test notification"))
		return
	}

	shared.WriteSuccess(w, TestResult{Success: true})
}

// TestAll handles POST /api/v1/notifications/test-all
// Tests all enabled notification services.
func (h *NotificationsHandler) TestAll(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	results, err := h.notifications.TestAllNotifications(r.Context(), athlete.ID)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to test notifications"))
		return
	}

	shared.WriteSuccess(w, results)
}
