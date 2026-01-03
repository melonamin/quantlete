package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

// ActivitiesHandler handles activity-related endpoints.
type ActivitiesHandler struct {
	svc    *services.ActivityService
	strava *strava.Client
}

// NewActivitiesHandler creates a new activities handler.
func NewActivitiesHandler(svc *services.ActivityService, stravaClient *strava.Client) *ActivitiesHandler {
	return &ActivitiesHandler{
		svc:    svc,
		strava: stravaClient,
	}
}

// List returns a paginated list of activities.
func (h *ActivitiesHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	// Parse query parameters
	input, err := parseListInput(r, athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	result, err := h.svc.List(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetByID returns a single activity by ID.
func (h *ActivitiesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid activity ID"})
		return
	}

	result, err := h.svc.GetByID(r.Context(), services.GetActivityInput{
		AthleteID:  athlete.ID,
		ActivityID: id,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetStreams handles GET /api/v1/activities/:id/streams
func (h *ActivitiesHandler) GetStreams(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid activity ID"})
		return
	}

	result, err := h.svc.GetStreams(r.Context(), services.GetActivityStreamsInput{
		AthleteID:  athlete.ID,
		ActivityID: id,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// parseListInput extracts filter and pagination params from the request.
func parseListInput(r *http.Request, athleteID int64) (services.ListActivitiesInput, error) {
	q := r.URL.Query()

	input := services.ListActivitiesInput{
		AthleteID: athleteID,
		Page:      1,
		PerPage:   50,
	}

	// Parse sport types
	if sportTypes := q.Get("sport_type"); sportTypes != "" {
		input.SportTypes = strings.Split(sportTypes, ",")
	}

	// Parse date range
	if t, ok := shared.ParseDateParam(q.Get("after")); ok {
		input.StartAfter = &t
	}
	if t, ok := shared.ParseDateParam(q.Get("before")); ok {
		input.StartBefore = &t
	}

	// Parse other filters
	input.GearID = q.Get("gear_id")
	input.Search = q.Get("search")

	if commute := q.Get("commute"); commute != "" {
		v := commute == "true" || commute == "1"
		input.Commute = &v
	}

	if trainer := q.Get("trainer"); trainer != "" {
		v := trainer == "true" || trainer == "1"
		input.Trainer = &v
	}

	// Parse pagination
	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			input.Page = p
		}
	}

	if perPageStr := q.Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 200 {
			input.PerPage = pp
		}
	}

	// Parse sorting
	input.OrderBy = q.Get("order_by")
	input.OrderDir = q.Get("order_dir")

	return input, nil
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
		// Note: status header already written, this just writes the body
		_, _ = w.Write([]byte(`{"error":"internal server error"}`))
	}
}
