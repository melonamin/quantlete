package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// DashboardHandler handles dashboard-related endpoints.
type DashboardHandler struct {
	stats  *storage.StatsRepository
	config *storage.DashboardConfigRepository
	strava *strava.Client
}

// NewDashboardHandler creates a new dashboard handler.
func NewDashboardHandler(stats *storage.StatsRepository, configRepo *storage.DashboardConfigRepository, stravaClient *strava.Client) *DashboardHandler {
	return &DashboardHandler{
		stats:  stats,
		config: configRepo,
		strava: stravaClient,
	}
}

// DashboardResponse combines all dashboard data.
type DashboardResponse struct {
	Stats            *storage.DashboardStats  `json:"stats"`
	WeeklyStats      []storage.WeeklyStat     `json:"weekly_stats"`
	RecentActivities []storage.RecentActivity `json:"recent_activities"`
	SportTypeStats   []storage.SportTypeStat  `json:"sport_type_stats"`
}

// GetDashboardConfig handles GET /api/v1/dashboard/config
func (h *DashboardHandler) GetDashboardConfig(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	cfg, err := h.config.Get(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get dashboard config"})
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

	var cfg storage.DashboardConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	if err := validateDashboardConfig(cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if err := h.config.Upsert(r.Context(), athlete.ID, cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to save dashboard config"})
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}

func validateDashboardConfig(cfg storage.DashboardConfig) error {
	seen := make(map[string]bool)
	for _, w := range cfg.Widgets {
		if strings.TrimSpace(w.ID) == "" {
			return fmt.Errorf("widget id is required")
		}
		if seen[w.ID] {
			return fmt.Errorf("duplicate widget id: %s", w.ID)
		}
		seen[w.ID] = true

		switch w.Width {
		case storage.WidgetWidthOneThird, storage.WidgetWidthHalf, storage.WidgetWidthTwoThird, storage.WidgetWidthFull:
		default:
			return fmt.Errorf("invalid widget width for %s", w.ID)
		}
	}
	return nil
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
		slog.Error("failed to get dashboard stats", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get stats"})
		return
	}

	weeklyStats, err := h.stats.GetWeeklyStats(ctx, athlete.ID)
	if err != nil {
		slog.Error("failed to get weekly stats", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get weekly stats"})
		return
	}

	recentActivities, err := h.stats.GetRecentActivities(ctx, athlete.ID, 5)
	if err != nil {
		slog.Error("failed to get recent activities", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get recent activities"})
		return
	}

	sportTypeStats, err := h.stats.GetStatsBySportType(ctx, athlete.ID)
	if err != nil {
		slog.Error("failed to get sport type stats", "error", err, "athlete_id", athlete.ID)
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

	stats, err := h.stats.GetMonthlyStats(r.Context(), athlete.ID, year)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get monthly stats"})
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

	stats, err := h.stats.GetYearlyStats(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get yearly stats"})
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

	data, err := h.stats.GetCalendarData(r.Context(), athlete.ID, year)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get calendar data"})
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

	activities, err := h.stats.GetCalendarActivities(r.Context(), athlete.ID, year, month)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get calendar activities"})
		return
	}

	if activities == nil {
		activities = []storage.CalendarActivity{}
	}

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

	summary, err := h.stats.GetCalendarMonthSummary(r.Context(), athlete.ID, year, month)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get calendar summary"})
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
