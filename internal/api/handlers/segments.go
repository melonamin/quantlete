package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/geo"
	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type SegmentsHandler struct {
	segments *storage.SegmentRepository
	strava   *strava.Client
}

func NewSegmentsHandler(segments *storage.SegmentRepository, stravaClient *strava.Client) *SegmentsHandler {
	return &SegmentsHandler{segments: segments, strava: stravaClient}
}

type segmentResponse struct {
	ID                   int64    `json:"id"`
	Name                 string   `json:"name"`
	ActivityType         string   `json:"activity_type"`
	Distance             float64  `json:"distance"`
	AverageGrade         float64  `json:"average_grade"`
	MaximumGrade         float64  `json:"maximum_grade"`
	ElevationHigh        float64  `json:"elevation_high"`
	ElevationLow         float64  `json:"elevation_low"`
	ClimbCategory        int      `json:"climb_category"`
	StartLat             *float64 `json:"start_lat,omitempty"`
	StartLng             *float64 `json:"start_lng,omitempty"`
	EndLat               *float64 `json:"end_lat,omitempty"`
	EndLng               *float64 `json:"end_lng,omitempty"`
	Starred              bool     `json:"starred"`
	Polyline             string   `json:"polyline,omitempty"`
	AthleteKOMRank       *int     `json:"athlete_kom_rank,omitempty"`
	AthleteEffortCount   *int     `json:"athlete_effort_count,omitempty"`
	AthletePRElapsedTime *int     `json:"athlete_pr_elapsed_time,omitempty"`
	AthletePRDate        *string  `json:"athlete_pr_date,omitempty"`
}

func segmentToResponse(s *storage.Segment) segmentResponse {
	var prDate *string
	if s.AthletePRDate != nil && !s.AthletePRDate.IsZero() {
		v := s.AthletePRDate.Format(time.RFC3339)
		prDate = &v
	}
	return segmentResponse{
		ID:                   s.ID,
		Name:                 s.Name,
		ActivityType:         s.ActivityType,
		Distance:             s.Distance,
		AverageGrade:         s.AverageGrade,
		MaximumGrade:         s.MaximumGrade,
		ElevationHigh:        s.ElevationHigh,
		ElevationLow:         s.ElevationLow,
		ClimbCategory:        s.ClimbCategory,
		StartLat:             s.StartLat,
		StartLng:             s.StartLng,
		EndLat:               s.EndLat,
		EndLng:               s.EndLng,
		Starred:              s.Starred,
		Polyline:             s.Polyline,
		AthleteKOMRank:       s.AthleteKOMRank,
		AthleteEffortCount:   s.AthleteEffortCount,
		AthletePRElapsedTime: s.AthletePRElapsedTime,
		AthletePRDate:        prDate,
	}
}

type segmentListItemResponse struct {
	segmentResponse
	TimesCompleted  int     `json:"times_completed"`
	LastEffortDate  *string `json:"last_effort_date,omitempty"`
	BestElapsedTime *int    `json:"best_elapsed_time,omitempty"`
}

func segmentListItemToResponse(it storage.SegmentListItem) segmentListItemResponse {
	seg := segmentToResponse(&it.Segment)
	seg.Polyline = ""

	var last *string
	if it.LastEffortDate != nil && !it.LastEffortDate.IsZero() {
		v := it.LastEffortDate.Format(time.RFC3339)
		last = &v
	}

	return segmentListItemResponse{
		segmentResponse: seg,
		TimesCompleted:  it.TimesCompleted,
		LastEffortDate:  last,
		BestElapsedTime: it.BestElapsedTime,
	}
}

type segmentEffortResponse struct {
	ID               int64    `json:"id"`
	SegmentID        int64    `json:"segment_id"`
	ActivityID       int64    `json:"activity_id"`
	AthleteID        int64    `json:"athlete_id"`
	Name             string   `json:"name,omitempty"`
	ElapsedTime      int      `json:"elapsed_time"`
	MovingTime       int      `json:"moving_time"`
	StartDate        *string  `json:"start_date,omitempty"`
	StartDateLocal   *string  `json:"start_date_local,omitempty"`
	Distance         float64  `json:"distance"`
	AverageWatts     *float64 `json:"average_watts,omitempty"`
	AverageHeartrate *float64 `json:"average_heartrate,omitempty"`
	MaxHeartrate     *int     `json:"max_heartrate,omitempty"`
	PRRank           *int     `json:"pr_rank,omitempty"`
	Country          string   `json:"country,omitempty"`
}

func segmentEffortToResponse(e storage.SegmentEffort) segmentEffortResponse {
	var startDate *string
	if e.StartDate != nil && !e.StartDate.IsZero() {
		v := e.StartDate.Format(time.RFC3339)
		startDate = &v
	}
	var startLocal *string
	if e.StartDateLocal != nil && !e.StartDateLocal.IsZero() {
		v := e.StartDateLocal.Format(time.RFC3339)
		startLocal = &v
	}
	return segmentEffortResponse{
		ID:               e.ID,
		SegmentID:        e.SegmentID,
		ActivityID:       e.ActivityID,
		AthleteID:        e.AthleteID,
		Name:             e.Name,
		ElapsedTime:      e.ElapsedTime,
		MovingTime:       e.MovingTime,
		StartDate:        startDate,
		StartDateLocal:   startLocal,
		Distance:         e.Distance,
		AverageWatts:     e.AverageWatts,
		AverageHeartrate: e.AverageHeartrate,
		MaxHeartrate:     e.MaxHeartrate,
		PRRank:           e.PRRank,
		Country:          e.Country,
	}
}

type segmentsListResponse struct {
	Data       []segmentListItemResponse `json:"data"`
	Total      int                       `json:"total"`
	Page       int                       `json:"page"`
	PerPage    int                       `json:"per_page"`
	TotalPages int                       `json:"total_pages"`
}

type segmentEffortsListResponse struct {
	Data       []segmentEffortResponse `json:"data"`
	Total      int                     `json:"total"`
	Page       int                     `json:"page"`
	PerPage    int                     `json:"per_page"`
	TotalPages int                     `json:"total_pages"`
}

// List handles GET /api/v1/segments
func (h *SegmentsHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	q := r.URL.Query()
	f := storage.SegmentFilters{
		ActivityType: q.Get("activity_type"),
		Country:      q.Get("country"),
		Search:       q.Get("search"),
		KOMOnly:      q.Get("kom_only") == "true" || q.Get("kom_only") == "1",
		QueryParams:  pagination.ParseQueryParams(q),
	}
	if v := q.Get("starred"); v != "" {
		b := v == "true" || v == "1"
		f.Starred = &b
	}

	result, err := h.segments.List(r.Context(), athlete.ID, f)
	if err != nil {
		slog.Error("failed to list segments", "error", err, "athlete_id", athlete.ID, "filters", f)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to list segments"})
		return
	}

	out := make([]segmentListItemResponse, 0, len(result.Items))
	for _, it := range result.Items {
		out = append(out, segmentListItemToResponse(it))
	}

	writeJSON(w, http.StatusOK, segmentsListResponse{
		Data:       out,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
}

// GetByID handles GET /api/v1/segments/:id
func (h *SegmentsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid segment id"})
		return
	}

	seg, err := h.segments.GetByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch segment"})
		return
	}
	if seg == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "segment not found"})
		return
	}

	// Include a short effort history by default.
	efforts, err := h.segments.ListEfforts(r.Context(), athlete.ID, id, 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch efforts"})
		return
	}

	outEfforts := make([]segmentEffortResponse, 0, len(efforts))
	for _, e := range efforts {
		outEfforts = append(outEfforts, segmentEffortToResponse(e))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"segment": segmentToResponse(seg),
		"efforts": outEfforts,
	})
}

// ListEfforts handles GET /api/v1/segments/:id/efforts
func (h *SegmentsHandler) ListEfforts(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid segment id"})
		return
	}

	q := r.URL.Query()
	f := storage.SegmentEffortFilters{
		QueryParams: pagination.ParseQueryParams(q),
	}

	result, err := h.segments.ListEffortsPaginated(r.Context(), athlete.ID, id, f)
	if err != nil {
		slog.Error("failed to list segment efforts", "error", err, "athlete_id", athlete.ID, "segment_id", id, "filters", f)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch efforts"})
		return
	}

	out := make([]segmentEffortResponse, 0, len(result.Items))
	for _, e := range result.Items {
		out = append(out, segmentEffortToResponse(e))
	}

	writeJSON(w, http.StatusOK, segmentEffortsListResponse{
		Data:       out,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
}

// Countries handles GET /api/v1/segments/countries
func (h *SegmentsHandler) Countries(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	stats, err := h.segments.ListCountryStats(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to list countries"})
		return
	}

	type outItem struct {
		Country string `json:"country"`
		ISO2    string `json:"iso2,omitempty"`
		Count   int    `json:"count"`
	}

	out := make([]outItem, 0, len(stats))
	for _, s := range stats {
		item := outItem{Country: s.Country, Count: s.Count}
		if iso2, ok := geo.ISO2FromCountry(s.Country); ok {
			item.ISO2 = iso2
		}
		out = append(out, item)
	}

	writeJSON(w, http.StatusOK, out)
}
