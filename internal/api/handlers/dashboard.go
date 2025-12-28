package handlers

import (
	"net/http"
	"strconv"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// DashboardHandler handles dashboard-related endpoints.
type DashboardHandler struct {
	stats  *storage.StatsRepository
	strava *strava.Client
}

// NewDashboardHandler creates a new dashboard handler.
func NewDashboardHandler(stats *storage.StatsRepository, stravaClient *strava.Client) *DashboardHandler {
	return &DashboardHandler{
		stats:  stats,
		strava: stravaClient,
	}
}

// DashboardResponse combines all dashboard data.
type DashboardResponse struct {
	Stats            *storage.DashboardStats    `json:"stats"`
	WeeklyStats      []storage.WeeklyStat       `json:"weekly_stats"`
	RecentActivities []storage.RecentActivity   `json:"recent_activities"`
	SportTypeStats   []storage.SportTypeStat    `json:"sport_type_stats"`
}

// GetDashboard handles GET /api/v1/dashboard
// Returns all dashboard data in a single response.
func (h *DashboardHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	ctx := r.Context()

	// Fetch all dashboard data
	stats, err := h.stats.GetDashboardStats(ctx, athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get stats"})
		return
	}

	weeklyStats, err := h.stats.GetWeeklyStats(ctx, athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get weekly stats"})
		return
	}

	recentActivities, err := h.stats.GetRecentActivities(ctx, athlete.ID, 5)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get recent activities"})
		return
	}

	sportTypeStats, err := h.stats.GetStatsBySportType(ctx, athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get sport type stats"})
		return
	}

	resp := DashboardResponse{
		Stats:            stats,
		WeeklyStats:      weeklyStats,
		RecentActivities: recentActivities,
		SportTypeStats:   sportTypeStats,
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetStats handles GET /api/v1/dashboard/stats
func (h *DashboardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	stats, err := h.stats.GetDashboardStats(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get stats"})
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

	stats, err := h.stats.GetWeeklyStats(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get weekly stats"})
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

	activities, err := h.stats.GetRecentActivities(r.Context(), athlete.ID, limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get recent activities"})
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

	stats, err := h.stats.GetStatsBySportType(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get sport type stats"})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
