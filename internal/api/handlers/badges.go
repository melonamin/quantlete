package handlers

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sasha/stata/internal/badges"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// BadgesHandler handles public badge endpoints.
type BadgesHandler struct {
	stats    *storage.StatsRepository
	settings *storage.SettingsRepository
	strava   *strava.Client
}

// NewBadgesHandler creates a new badges handler.
func NewBadgesHandler(stats *storage.StatsRepository, settings *storage.SettingsRepository, stravaClient *strava.Client) *BadgesHandler {
	return &BadgesHandler{
		stats:    stats,
		settings: settings,
		strava:   stravaClient,
	}
}

// getAthleteID returns the athlete ID for badge generation.
// First checks for authenticated user, then falls back to public badges setting.
func (h *BadgesHandler) getAthleteID(r *http.Request) (int64, error) {
	// Check if user is authenticated
	if h.strava != nil {
		if athlete := h.strava.GetAthlete(); athlete != nil {
			return athlete.ID, nil
		}
	}

	// Check if public badges are enabled for any athlete
	if h.settings == nil {
		return 0, fmt.Errorf("public badges not enabled")
	}
	athleteID, err := h.settings.GetPublicBadgesAthleteID(r.Context())
	if err != nil {
		return 0, err
	}
	if athleteID == 0 {
		return 0, fmt.Errorf("public badges not enabled")
	}

	return athleteID, nil
}

// parseParams extracts theme, size, background, and unit from query parameters.
func parseParams(r *http.Request) (badges.Theme, badges.Size, badges.Background, badges.UnitSystem) {
	theme := badges.GetTheme(r.URL.Query().Get("theme"))
	size := badges.GetSize(r.URL.Query().Get("size"))
	bg := badges.GetBackground(r.URL.Query().Get("bg"))
	unit := badges.UnitMetric
	if r.URL.Query().Get("unit") == "imperial" {
		unit = badges.UnitImperial
	}
	return theme, size, bg, unit
}

// writeSVG writes an SVG response with appropriate headers.
func writeSVG(w http.ResponseWriter, svg string) {
	// Generate ETag from content
	etag := fmt.Sprintf(`"%x"`, md5.Sum([]byte(svg)))

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600") // 1 hour
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(svg))
}

// writeBadgeError writes an error badge SVG.
func writeBadgeError(w http.ResponseWriter, message string) {
	theme := badges.DefaultTheme()
	size := badges.DefaultSize()
	bg := badges.DefaultBackground()

	svg, _ := badges.RenderSVG(badges.BadgeData{
		Title: "Error",
		Value: message,
	}, theme, size, bg)

	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK) // Return 200 so the image displays
	w.Write([]byte(svg))
}

// GetDistanceBadge handles GET /badges/distance.svg
func (h *BadgesHandler) GetDistanceBadge(w http.ResponseWriter, r *http.Request) {
	athleteID, err := h.getAthleteID(r)
	if err != nil {
		writeBadgeError(w, "Not Available")
		return
	}

	theme, size, bg, unit := parseParams(r)

	stats, err := h.stats.GetDashboardStats(r.Context(), athleteID)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	svg, err := badges.RenderSVG(badges.BadgeData{
		Title: "Total Distance",
		Value: badges.FormatDistance(stats.TotalDistance, unit),
	}, theme, size, bg)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	writeSVG(w, svg)
}

// GetTimeBadge handles GET /badges/time.svg
func (h *BadgesHandler) GetTimeBadge(w http.ResponseWriter, r *http.Request) {
	athleteID, err := h.getAthleteID(r)
	if err != nil {
		writeBadgeError(w, "Not Available")
		return
	}

	theme, size, bg, _ := parseParams(r)

	stats, err := h.stats.GetDashboardStats(r.Context(), athleteID)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	svg, err := badges.RenderSVG(badges.BadgeData{
		Title: "Total Time",
		Value: badges.FormatDuration(stats.TotalMovingTime),
	}, theme, size, bg)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	writeSVG(w, svg)
}

// GetElevationBadge handles GET /badges/elevation.svg
func (h *BadgesHandler) GetElevationBadge(w http.ResponseWriter, r *http.Request) {
	athleteID, err := h.getAthleteID(r)
	if err != nil {
		writeBadgeError(w, "Not Available")
		return
	}

	theme, size, bg, unit := parseParams(r)

	stats, err := h.stats.GetDashboardStats(r.Context(), athleteID)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	svg, err := badges.RenderSVG(badges.BadgeData{
		Title: "Total Elevation",
		Value: badges.FormatElevation(stats.TotalElevationGain, unit),
	}, theme, size, bg)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	writeSVG(w, svg)
}

// GetActivitiesBadge handles GET /badges/activities.svg
func (h *BadgesHandler) GetActivitiesBadge(w http.ResponseWriter, r *http.Request) {
	athleteID, err := h.getAthleteID(r)
	if err != nil {
		writeBadgeError(w, "Not Available")
		return
	}

	theme, size, bg, _ := parseParams(r)

	stats, err := h.stats.GetDashboardStats(r.Context(), athleteID)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	svg, err := badges.RenderSVG(badges.BadgeData{
		Title: "Activities",
		Value: badges.FormatNumber(stats.TotalActivities),
	}, theme, size, bg)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	writeSVG(w, svg)
}

// GetEddingtonBadge handles GET /badges/eddington.svg
func (h *BadgesHandler) GetEddingtonBadge(w http.ResponseWriter, r *http.Request) {
	athleteID, err := h.getAthleteID(r)
	if err != nil {
		writeBadgeError(w, "Not Available")
		return
	}

	theme, size, bg, unit := parseParams(r)

	// Get eddington data (nil sportTypes = all activities)
	result, err := h.stats.GetEddingtonData(r.Context(), athleteID, nil)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	unitLabel := "km"
	if unit == badges.UnitImperial {
		unitLabel = "mi"
	}

	svg, err := badges.RenderSVG(badges.BadgeData{
		Title:    "Eddington Number",
		Value:    fmt.Sprintf("E%d", result.Number),
		Subtitle: fmt.Sprintf("%d days with %d+ %s", result.Number, result.Number, unitLabel),
	}, theme, size, bg)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	writeSVG(w, svg)
}

// GetYearBadge handles GET /badges/year-{year}.svg
func (h *BadgesHandler) GetYearBadge(w http.ResponseWriter, r *http.Request) {
	athleteID, err := h.getAthleteID(r)
	if err != nil {
		writeBadgeError(w, "Not Available")
		return
	}

	yearStr := chi.URLParam(r, "year")
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeBadgeError(w, "Invalid Year")
		return
	}

	theme, size, bg, unit := parseParams(r)

	// Get yearly stats
	stats, err := h.stats.GetYearlyStats(r.Context(), athleteID)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	// Find the requested year
	var yearStat *storage.YearStat
	for _, s := range stats {
		if s.Year == year {
			yearStat = &s
			break
		}
	}

	if yearStat == nil {
		writeBadgeError(w, "No Data")
		return
	}

	svg, err := badges.RenderSVG(badges.BadgeData{
		Title:    fmt.Sprintf("Year %d", year),
		Value:    badges.FormatDistance(yearStat.TotalDistance, unit),
		Subtitle: fmt.Sprintf("%d activities", yearStat.ActivityCount),
	}, theme, size, bg)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	writeSVG(w, svg)
}

// GetMonthBadge handles GET /badges/month-{month}.svg
func (h *BadgesHandler) GetMonthBadge(w http.ResponseWriter, r *http.Request) {
	athleteID, err := h.getAthleteID(r)
	if err != nil {
		writeBadgeError(w, "Not Available")
		return
	}

	monthStr := chi.URLParam(r, "month") // Format: 2024-12
	if len(monthStr) < 7 || !strings.Contains(monthStr, "-") {
		writeBadgeError(w, "Invalid Month")
		return
	}

	theme, size, bg, unit := parseParams(r)

	// Parse year from month string
	yearStr := monthStr[:4]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		writeBadgeError(w, "Invalid Month")
		return
	}

	// Get monthly stats for the year
	stats, err := h.stats.GetMonthlyStats(r.Context(), athleteID, year)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	// Find the requested month
	var monthStat *storage.MonthlyStat
	for _, s := range stats {
		if s.Month == monthStr {
			monthStat = &s
			break
		}
	}

	if monthStat == nil {
		writeBadgeError(w, "No Data")
		return
	}

	// Format month name
	monthName := formatMonthName(monthStr)

	svg, err := badges.RenderSVG(badges.BadgeData{
		Title:    monthName,
		Value:    badges.FormatDistance(monthStat.TotalDistance, unit),
		Subtitle: fmt.Sprintf("%d activities", monthStat.ActivityCount),
	}, theme, size, bg)
	if err != nil {
		writeBadgeError(w, "Error")
		return
	}

	writeSVG(w, svg)
}

// formatMonthName converts "2024-12" to "December 2024"
func formatMonthName(monthStr string) string {
	parts := strings.Split(monthStr, "-")
	if len(parts) != 2 {
		return monthStr
	}

	monthNum, err := strconv.Atoi(parts[1])
	if err != nil {
		return monthStr
	}

	months := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}

	if monthNum < 1 || monthNum > 12 {
		return monthStr
	}

	return fmt.Sprintf("%s %s", months[monthNum-1], parts[0])
}
