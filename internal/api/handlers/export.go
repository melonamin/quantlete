package handlers

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// ExportHandler handles data export endpoints.
type ExportHandler struct {
	activities   *services.ActivityService
	stravaClient *strava.Client
}

// NewExportHandler creates a new export handler.
func NewExportHandler(activities *services.ActivityService, stravaClient *strava.Client) *ExportHandler {
	return &ExportHandler{
		activities:   activities,
		stravaClient: stravaClient,
	}
}

// ExportActivitiesCSV exports activities as CSV.
func (h *ExportHandler) ExportActivitiesCSV(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	input := parseExportActivitiesInput(r, athlete.ID)
	writer := csv.NewWriter(w)
	started := false
	start := func() error {
		if started {
			return nil
		}
		filename := fmt.Sprintf("quantlete-activities-%s.csv", time.Now().Format("2006-01-02"))
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		started = true
		return writer.Write(exportCSVHeader)
	}

	err := h.activities.StreamExport(ctx, input, func(activities []storage.Activity) error {
		if err := start(); err != nil {
			return err
		}
		for i := range activities {
			if err := writer.Write(exportCSVRow(&activities[i])); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	})
	if err != nil {
		if !started {
			shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to fetch activities"))
		}
		return
	}
	if err := start(); err != nil {
		return
	}
	writer.Flush()
}

// ExportActivitiesJSON exports activities as JSON.
func (h *ExportHandler) ExportActivitiesJSON(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	input := parseExportActivitiesInput(r, athlete.ID)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	started := false
	first := true
	start := func() error {
		if started {
			return nil
		}
		filename := fmt.Sprintf("quantlete-activities-%s.json", time.Now().Format("2006-01-02"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		started = true
		_, err := io.WriteString(w, "[\n")
		return err
	}

	err := h.activities.StreamExport(ctx, input, func(activities []storage.Activity) error {
		if err := start(); err != nil {
			return err
		}
		for i := range activities {
			if !first {
				if _, err := io.WriteString(w, ",\n"); err != nil {
					return err
				}
			}
			if err := encoder.Encode(&activities[i]); err != nil {
				return err
			}
			first = false
		}
		return nil
	})
	if err != nil {
		if !started {
			shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to fetch activities"))
		}
		return
	}
	if err := start(); err != nil {
		return
	}
	_, _ = io.WriteString(w, "]\n")
}

func parseExportActivitiesInput(r *http.Request, athleteID int64) services.ExportActivitiesInput {
	input := services.ExportActivitiesInput{AthleteID: athleteID}
	if sportTypes := r.URL.Query().Get("sport_type"); sportTypes != "" {
		input.SportTypes = strings.Split(sportTypes, ",")
	}
	if after := r.URL.Query().Get("after"); after != "" {
		if parsed, err := time.Parse("2006-01-02", after); err == nil {
			input.StartAfter = &parsed
		}
	}
	if before := r.URL.Query().Get("before"); before != "" {
		if parsed, err := time.Parse("2006-01-02", before); err == nil {
			input.StartBefore = &parsed
		}
	}
	return input
}

var exportCSVHeader = []string{
	"ID", "Name", "Sport Type", "Start Date", "Timezone",
	"Distance (m)", "Moving Time (s)", "Elapsed Time (s)",
	"Elevation Gain (m)", "Average Speed (m/s)", "Max Speed (m/s)",
	"Average Heartrate", "Max Heartrate", "Average Watts", "Max Watts",
	"Weighted Average Watts", "Kilojoules", "Calories",
	"Average Cadence", "Commute", "Trainer",
	"Start Lat", "Start Lng", "City", "Country",
	"Gear ID", "Description",
}

func exportCSVRow(a *storage.Activity) []string {
	return []string{
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
		formatOptionalFloat(a.AverageHeartrate, 1),
		formatOptionalFloat(a.MaxHeartrate, 1),
		formatOptionalFloat(a.AverageWatts, 1),
		formatOptionalFloat(a.MaxWatts, 1),
		formatOptionalFloat(a.WeightedAverageWatts, 1),
		formatOptionalFloat(a.Kilojoules, 1),
		formatOptionalFloat(a.Calories, 1),
		formatOptionalFloat(a.AverageCadence, 1),
		strconv.FormatBool(a.Commute),
		strconv.FormatBool(a.Trainer),
		formatOptionalFloat(a.StartLat, 6),
		formatOptionalFloat(a.StartLng, 6),
		a.LocationCity,
		a.LocationCountry,
		a.GearID,
		a.Description,
	}
}

func formatOptionalFloat(value *float64, precision int) string {
	if value == nil {
		return ""
	}
	return strconv.FormatFloat(*value, 'f', precision, 64)
}
