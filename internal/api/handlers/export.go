package handlers

import (
	"encoding/csv"
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

// MaxExportActivities is the maximum number of activities that can be exported at once.
const MaxExportActivities = 50000

// ExportHandler handles data export endpoints.
type ExportHandler struct {
	activityRepo *storage.ActivityRepository
	stravaClient *strava.Client
}

// NewExportHandler creates a new export handler.
func NewExportHandler(activityRepo *storage.ActivityRepository, stravaClient *strava.Client) *ExportHandler {
	return &ExportHandler{
		activityRepo: activityRepo,
		stravaClient: stravaClient,
	}
}

// ExportActivitiesCSV exports activities as CSV.
func (h *ExportHandler) ExportActivitiesCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	// Parse optional filters
	filters := storage.ActivityFilters{}

	if sportTypes := r.URL.Query().Get("sport_type"); sportTypes != "" {
		filters.SportTypes = strings.Split(sportTypes, ",")
	}

	if after := r.URL.Query().Get("after"); after != "" {
		t, err := time.Parse("2006-01-02", after)
		if err == nil {
			filters.StartAfter = &t
		}
	}

	if before := r.URL.Query().Get("before"); before != "" {
		t, err := time.Parse("2006-01-02", before)
		if err == nil {
			filters.StartBefore = &t
		}
	}

	// Fetch activities with reasonable limit to prevent DoS
	activities, _, err := h.activityRepo.List(ctx, filters, storage.Pagination{Page: 1, PerPage: MaxExportActivities})
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	// Set headers for CSV download
	filename := fmt.Sprintf("stata-activities-%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID", "Name", "Sport Type", "Start Date", "Timezone",
		"Distance (m)", "Moving Time (s)", "Elapsed Time (s)",
		"Elevation Gain (m)", "Average Speed (m/s)", "Max Speed (m/s)",
		"Average Heartrate", "Max Heartrate", "Average Watts", "Max Watts",
		"Weighted Average Watts", "Kilojoules", "Calories",
		"Average Cadence", "Commute", "Trainer",
		"Start Lat", "Start Lng", "City", "Country",
		"Gear ID", "Description",
	}
	if err := writer.Write(header); err != nil {
		return
	}

	// Write data rows
	for _, a := range activities {
		avgHR := ""
		if a.AverageHeartrate != nil {
			avgHR = fmt.Sprintf("%.1f", *a.AverageHeartrate)
		}
		maxHR := ""
		if a.MaxHeartrate != nil {
			maxHR = fmt.Sprintf("%.1f", *a.MaxHeartrate)
		}
		avgWatts := ""
		if a.AverageWatts != nil {
			avgWatts = fmt.Sprintf("%.1f", *a.AverageWatts)
		}
		maxWatts := ""
		if a.MaxWatts != nil {
			maxWatts = fmt.Sprintf("%.1f", *a.MaxWatts)
		}
		weightedWatts := ""
		if a.WeightedAverageWatts != nil {
			weightedWatts = fmt.Sprintf("%.1f", *a.WeightedAverageWatts)
		}
		kj := ""
		if a.Kilojoules != nil {
			kj = fmt.Sprintf("%.1f", *a.Kilojoules)
		}
		cal := ""
		if a.Calories != nil {
			cal = fmt.Sprintf("%.1f", *a.Calories)
		}
		avgCadence := ""
		if a.AverageCadence != nil {
			avgCadence = fmt.Sprintf("%.1f", *a.AverageCadence)
		}
		startLat := ""
		if a.StartLat != nil {
			startLat = fmt.Sprintf("%.6f", *a.StartLat)
		}
		startLng := ""
		if a.StartLng != nil {
			startLng = fmt.Sprintf("%.6f", *a.StartLng)
		}

		row := []string{
			strconv.FormatInt(a.ID, 10),
			a.Name,
			a.SportType,
			a.StartDate.Format(time.RFC3339),
			a.Timezone,
			fmt.Sprintf("%.2f", a.Distance),
			strconv.Itoa(a.MovingTime),
			strconv.Itoa(a.ElapsedTime),
			fmt.Sprintf("%.2f", a.TotalElevationGain),
			fmt.Sprintf("%.2f", a.AverageSpeed),
			fmt.Sprintf("%.2f", a.MaxSpeed),
			avgHR,
			maxHR,
			avgWatts,
			maxWatts,
			weightedWatts,
			kj,
			cal,
			avgCadence,
			strconv.FormatBool(a.Commute),
			strconv.FormatBool(a.Trainer),
			startLat,
			startLng,
			a.LocationCity,
			a.LocationCountry,
			a.GearID,
			a.Description,
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}
}

// ExportActivitiesJSON exports activities as JSON.
func (h *ExportHandler) ExportActivitiesJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	// Parse optional filters
	filters := storage.ActivityFilters{}

	if sportTypes := r.URL.Query().Get("sport_type"); sportTypes != "" {
		filters.SportTypes = strings.Split(sportTypes, ",")
	}

	if after := r.URL.Query().Get("after"); after != "" {
		t, err := time.Parse("2006-01-02", after)
		if err == nil {
			filters.StartAfter = &t
		}
	}

	if before := r.URL.Query().Get("before"); before != "" {
		t, err := time.Parse("2006-01-02", before)
		if err == nil {
			filters.StartBefore = &t
		}
	}

	// Fetch activities with reasonable limit to prevent DoS
	activities, _, err := h.activityRepo.List(ctx, filters, storage.Pagination{Page: 1, PerPage: MaxExportActivities})
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	// Set headers for JSON download
	filename := fmt.Sprintf("stata-activities-%s.json", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(activities); err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
	}
}

// ExportStats returns export statistics (counts, date ranges).
func (h *ExportHandler) ExportStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	filters := storage.ActivityFilters{}

	// Get most recent activity (also gives us total count)
	activities, total, err := h.activityRepo.List(ctx, filters, storage.Pagination{
		Page:     1,
		PerPage:  1,
		OrderBy:  "start_date",
		OrderDir: "DESC",
	})
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	var lastActivity *string
	if len(activities) > 0 {
		s := activities[0].StartDate.Format(time.RFC3339)
		lastActivity = &s
	}

	// Get oldest activity
	oldest, _, err := h.activityRepo.List(ctx, filters, storage.Pagination{
		Page:     1,
		PerPage:  1,
		OrderBy:  "start_date",
		OrderDir: "ASC",
	})
	if err != nil {
		http.Error(w, "Failed to fetch oldest activity", http.StatusInternalServerError)
		return
	}

	var firstActivity *string
	if len(oldest) > 0 {
		s := oldest[0].StartDate.Format(time.RFC3339)
		firstActivity = &s
	}

	response := map[string]interface{}{
		"total_activities": total,
		"first_activity":   firstActivity,
		"last_activity":    lastActivity,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode export stats response", "error", err)
	}
}
