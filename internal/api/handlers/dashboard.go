package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/strava"
)

// DashboardHandler handles dashboard-related endpoints.
type DashboardHandler struct {
	svc    *services.DashboardService
	strava *strava.Client
}

// NewDashboardHandler creates a new dashboard handler.
func NewDashboardHandler(svc *services.DashboardService, stravaClient *strava.Client) *DashboardHandler {
	return &DashboardHandler{
		svc:    svc,
		strava: stravaClient,
	}
}

// GetDashboardConfig handles GET /api/v1/dashboard/config
func (h *DashboardHandler) GetDashboardConfig(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	cfg, err := h.svc.GetConfig(r.Context(), services.GetDashboardInput{
		AthleteID: athlete.ID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}

// UpdateDashboardConfig handles PUT /api/v1/dashboard/config
func (h *DashboardHandler) UpdateDashboardConfig(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var cfg services.DashboardConfigOutput
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	result, err := h.svc.UpdateConfig(r.Context(), services.UpdateDashboardConfigInput{
		AthleteID: athlete.ID,
		Config:    cfg,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetDashboard handles GET /api/v1/dashboard
// Returns all dashboard data in a single response.
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	result, err := h.svc.GetDashboard(r.Context(), services.GetDashboardInput{
		AthleteID: athlete.ID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// GetStats handles GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	stats, err := h.svc.GetStats(r.Context(), services.GetDashboardInput{
		AthleteID: athlete.ID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// GetWeeklyStats handles GET /api/v1/dashboard/weekly
func (h *DashboardHandler) GetWeeklyStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	stats, err := h.svc.GetWeeklyStats(r.Context(), services.GetDashboardInput{
		AthleteID: athlete.ID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// GetRecentActivities handles GET /api/v1/dashboard/recent
func (h *DashboardHandler) GetRecentActivities(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	limit := 5
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 20 {
			limit = parsed
		}
	}

	activities, err := h.svc.GetRecentActivities(r.Context(), services.GetRecentActivitiesInput{
		AthleteID: athlete.ID,
		Limit:     limit,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, activities)
}

// GetSportTypeStats handles GET /api/v1/dashboard/sports
func (h *DashboardHandler) GetSportTypeStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	stats, err := h.svc.GetSportTypeStats(r.Context(), services.GetDashboardInput{
		AthleteID: athlete.ID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// GetMonthlyStats handles GET /api/v1/dashboard/monthly
func (h *DashboardHandler) GetMonthlyStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	year := 0
	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed > 2000 && parsed < 2100 {
			year = parsed
		}
	}

	stats, err := h.svc.GetMonthlyStats(r.Context(), services.GetMonthlyStatsInput{
		AthleteID: athlete.ID,
		Year:      year,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// GetYearlyStats handles GET /api/v1/dashboard/yearly
func (h *DashboardHandler) GetYearlyStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	stats, err := h.svc.GetYearlyStats(r.Context(), services.GetDashboardInput{
		AthleteID: athlete.ID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// GetCalendarData handles GET /api/v1/dashboard/calendar
func (h *DashboardHandler) GetCalendarData(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	year := time.Now().Year()
	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed > 2000 && parsed < 2100 {
			year = parsed
		}
	}

	data, err := h.svc.GetCalendarData(r.Context(), services.GetCalendarDataInput{
		AthleteID: athlete.ID,
		Year:      year,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// Heatmap and Eddington endpoints are handled by StatsHandler.

// GetCalendarActivities handles GET /api/v1/dashboard/calendar/activities
func (h *DashboardHandler) GetCalendarActivities(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed > 2000 && parsed < 2100 {
			year = parsed
		}
	}

	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	activities, err := h.svc.GetCalendarActivities(r.Context(), services.GetCalendarActivitiesInput{
		AthleteID: athlete.ID,
		Year:      year,
		Month:     month,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	// The service already returns an empty slice if nil
	writeJSON(w, http.StatusOK, activities)
}

// GetCalendarSummary handles GET /api/v1/dashboard/calendar/summary
func (h *DashboardHandler) GetCalendarSummary(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	if y := r.URL.Query().Get("year"); y != "" {
		if parsed, err := strconv.Atoi(y); err == nil && parsed > 2000 && parsed < 2100 {
			year = parsed
		}
	}

	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed >= 1 && parsed <= 12 {
			month = parsed
		}
	}

	summary, err := h.svc.GetCalendarSummary(r.Context(), services.GetCalendarActivitiesInput{
		AthleteID: athlete.ID,
		Year:      year,
		Month:     month,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// parseSportTypes parses sport_type query parameter into a slice.
func parseSportTypes(r *http.Request) []string {
	if st := r.URL.Query().Get("sport_type"); st != "" {
		return strings.Split(st, ",")
	}
	return nil
}
