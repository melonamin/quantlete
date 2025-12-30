package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// ActivitiesHandler handles activity-related endpoints.
type ActivitiesHandler struct {
	repo    *storage.ActivityRepository
	streams *storage.StreamRepository
	strava  *strava.Client
}

// NewActivitiesHandler creates a new activities handler.
func NewActivitiesHandler(repo *storage.ActivityRepository, streams *storage.StreamRepository, stravaClient *strava.Client) *ActivitiesHandler {
	return &ActivitiesHandler{
		repo:    repo,
		streams: streams,
		strava:  stravaClient,
	}
}

// ListResponse represents a paginated list response.
type ListResponse struct {
	Data       []ActivityResponse `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PerPage    int                `json:"per_page"`
	TotalPages int                `json:"total_pages"`
}

// ActivityResponse represents an activity in API responses.
type ActivityResponse struct {
	ID                   int64    `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description,omitempty"`
	SportType            string   `json:"sport_type"`
	StartDate            string   `json:"start_date"`
	StartDateLocal       string   `json:"start_date_local"`
	Timezone             string   `json:"timezone,omitempty"`
	Distance             float64  `json:"distance"`
	MovingTime           int      `json:"moving_time"`
	ElapsedTime          int      `json:"elapsed_time"`
	TotalElevationGain   float64  `json:"total_elevation_gain"`
	ElevHigh             *float64 `json:"elev_high,omitempty"`
	ElevLow              *float64 `json:"elev_low,omitempty"`
	AverageSpeed         float64  `json:"average_speed"`
	MaxSpeed             float64  `json:"max_speed"`
	AverageHeartrate     *float64 `json:"average_heartrate,omitempty"`
	MaxHeartrate         *float64 `json:"max_heartrate,omitempty"`
	AverageWatts         *float64 `json:"average_watts,omitempty"`
	MaxWatts             *float64 `json:"max_watts,omitempty"`
	WeightedAverageWatts *float64 `json:"weighted_average_watts,omitempty"`
	Kilojoules           *float64 `json:"kilojoules,omitempty"`
	AverageCadence       *float64 `json:"average_cadence,omitempty"`
	Calories             *float64 `json:"calories,omitempty"`
	KudosCount           int      `json:"kudos_count"`
	CommentCount         int      `json:"comment_count"`
	PhotoCount           int      `json:"photo_count"`
	Commute              bool     `json:"commute"`
	Private              bool     `json:"private"`
	Trainer              bool     `json:"trainer"`
	WorkoutType          *int     `json:"workout_type,omitempty"`
	DeviceName           string   `json:"device_name,omitempty"`
	GearID               string   `json:"gear_id,omitempty"`
	StartLat             *float64 `json:"start_lat,omitempty"`
	StartLng             *float64 `json:"start_lng,omitempty"`
	EndLat               *float64 `json:"end_lat,omitempty"`
	EndLng               *float64 `json:"end_lng,omitempty"`
	SummaryPolyline      string   `json:"summary_polyline,omitempty"`
	LocationCity         string   `json:"location_city,omitempty"`
	LocationCountry      string   `json:"location_country,omitempty"`
}

// List returns a paginated list of activities.
func (h *ActivitiesHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filters, page, err := parseListParams(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	// Query activities
	activities, total, err := h.repo.List(ctx, filters, page)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch activities"})
		return
	}

	// Convert to response format
	data := make([]ActivityResponse, len(activities))
	for i, a := range activities {
		data[i] = activityToResponse(&a)
	}

	totalPages := (total + page.PerPage - 1) / page.PerPage
	response := ListResponse{
		Data:       data,
		Total:      total,
		Page:       page.Page,
		PerPage:    page.PerPage,
		TotalPages: totalPages,
	}

	writeJSON(w, http.StatusOK, response)
}

// GetByID returns a single activity by ID.
func (h *ActivitiesHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	activity, err := h.repo.GetByID(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch activity"})
		return
	}

	if activity == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "activity not found"})
		return
	}

	// Verify the activity belongs to the authenticated athlete
	if activity.AthleteID != athlete.ID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "access denied"})
		return
	}

	writeJSON(w, http.StatusOK, activityToResponse(activity))
}

type ActivityStreamResponse struct {
	ActivityID   int64           `json:"activity_id"`
	StreamType   string          `json:"stream_type"`
	OriginalSize int             `json:"original_size"`
	Resolution   string          `json:"resolution"`
	SeriesType   string          `json:"series_type"`
	Data         json.RawMessage `json:"data"`
}

// GetStreams handles GET /api/v1/activities/:id/streams
func (h *ActivitiesHandler) GetStreams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	// Verify the activity belongs to the authenticated athlete
	activity, err := h.repo.GetByID(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch activity"})
		return
	}
	if activity == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "activity not found"})
		return
	}
	if activity.AthleteID != athlete.ID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "access denied"})
		return
	}

	streams, err := h.streams.GetByActivityID(ctx, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch streams"})
		return
	}

	resp := make([]ActivityStreamResponse, 0, len(streams))
	for _, s := range streams {
		resp = append(resp, ActivityStreamResponse{
			ActivityID:   s.ActivityID,
			StreamType:   s.StreamType,
			OriginalSize: s.OriginalSize,
			Resolution:   s.Resolution,
			SeriesType:   s.SeriesType,
			Data:         s.Data,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// parseListParams extracts filter and pagination params from the request.
func parseListParams(r *http.Request) (storage.ActivityFilters, storage.Pagination, error) {
	q := r.URL.Query()

	filters := storage.ActivityFilters{}
	page := storage.Pagination{
		Page:    1,
		PerPage: 50,
	}

	// Parse sport types
	if sportTypes := q.Get("sport_type"); sportTypes != "" {
		filters.SportTypes = strings.Split(sportTypes, ",")
	}

	// Parse date range
	if after := q.Get("after"); after != "" {
		t, err := time.Parse(time.RFC3339, after)
		if err != nil {
			t, err = time.Parse("2006-01-02", after)
		}
		if err == nil {
			filters.StartAfter = &t
		}
	}

	if before := q.Get("before"); before != "" {
		t, err := time.Parse(time.RFC3339, before)
		if err != nil {
			t, err = time.Parse("2006-01-02", before)
		}
		if err == nil {
			filters.StartBefore = &t
		}
	}

	// Parse other filters
	filters.GearID = q.Get("gear_id")
	filters.Search = q.Get("search")

	if commute := q.Get("commute"); commute != "" {
		v := commute == "true" || commute == "1"
		filters.Commute = &v
	}

	if trainer := q.Get("trainer"); trainer != "" {
		v := trainer == "true" || trainer == "1"
		filters.Trainer = &v
	}

	// Parse pagination
	if pageStr := q.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page.Page = p
		}
	}

	if perPageStr := q.Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 200 {
			page.PerPage = pp
		}
	}

	// Parse sorting
	page.OrderBy = q.Get("order_by")
	page.OrderDir = q.Get("order_dir")

	return filters, page, nil
}

// activityToResponse converts a storage activity to an API response.
func activityToResponse(a *storage.Activity) ActivityResponse {
	return ActivityResponse{
		ID:                   a.ID,
		Name:                 a.Name,
		Description:          a.Description,
		SportType:            a.SportType,
		StartDate:            a.StartDate.Format(time.RFC3339),
		StartDateLocal:       a.StartDateLocal.Format(time.RFC3339),
		Timezone:             a.Timezone,
		Distance:             a.Distance,
		MovingTime:           a.MovingTime,
		ElapsedTime:          a.ElapsedTime,
		TotalElevationGain:   a.TotalElevationGain,
		ElevHigh:             a.ElevHigh,
		ElevLow:              a.ElevLow,
		AverageSpeed:         a.AverageSpeed,
		MaxSpeed:             a.MaxSpeed,
		AverageHeartrate:     a.AverageHeartrate,
		MaxHeartrate:         a.MaxHeartrate,
		AverageWatts:         a.AverageWatts,
		MaxWatts:             a.MaxWatts,
		WeightedAverageWatts: a.WeightedAverageWatts,
		Kilojoules:           a.Kilojoules,
		AverageCadence:       a.AverageCadence,
		Calories:             a.Calories,
		KudosCount:           a.KudosCount,
		CommentCount:         a.CommentCount,
		PhotoCount:           a.PhotoCount,
		Commute:              a.Commute,
		Private:              a.Private,
		Trainer:              a.Trainer,
		WorkoutType:          a.WorkoutType,
		DeviceName:           a.DeviceName,
		GearID:               a.GearID,
		StartLat:             a.StartLat,
		StartLng:             a.StartLng,
		EndLat:               a.EndLat,
		EndLng:               a.EndLng,
		SummaryPolyline:      a.SummaryPolyline,
		LocationCity:         a.LocationCity,
		LocationCountry:      a.LocationCountry,
	}
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
