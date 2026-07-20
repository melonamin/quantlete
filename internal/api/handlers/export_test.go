package handlers

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

func TestExportActivitiesCSVScopedAndComplete(t *testing.T) {
	const (
		requestingAthleteID = int64(101)
		otherAthleteID      = int64(202)
		activityCount       = 250
	)

	db := exportTestDB(t)
	seedExportAthlete(t, db, requestingAthleteID)
	seedExportAthlete(t, db, otherAthleteID)
	seedExportActivities(t, db, requestingAthleteID, 1000, activityCount, "requester")
	seedExportActivities(t, db, otherAthleteID, 9000, 1, "other")

	handler := exportTestHandler(db, requestingAthleteID)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/activities/csv", http.NoBody)
	recorder := httptest.NewRecorder()
	handler.ExportActivitiesCSV(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("ExportActivitiesCSV() status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	records, err := csv.NewReader(strings.NewReader(recorder.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse CSV export: %v", err)
	}
	if len(records) != activityCount+1 {
		t.Fatalf("CSV record count = %d, want %d", len(records), activityCount+1)
	}

	wantHeader := []string{
		"ID", "Name", "Sport Type", "Start Date", "Timezone",
		"Distance (m)", "Moving Time (s)", "Elapsed Time (s)",
		"Elevation Gain (m)", "Average Speed (m/s)", "Max Speed (m/s)",
		"Average Heartrate", "Max Heartrate", "Average Watts", "Max Watts",
		"Weighted Average Watts", "Kilojoules", "Calories",
		"Average Cadence", "Commute", "Trainer",
		"Start Lat", "Start Lng", "City", "Country",
		"Gear ID", "Description",
	}
	if !reflect.DeepEqual(records[0], wantHeader) {
		t.Errorf("CSV header = %v, want %v", records[0], wantHeader)
	}
	for _, record := range records[1:] {
		if strings.HasPrefix(record[1], "other-") {
			t.Fatalf("CSV export included another athlete's activity: %v", record)
		}
	}
}

func TestExportActivitiesJSONScopedAndComplete(t *testing.T) {
	const (
		requestingAthleteID = int64(303)
		otherAthleteID      = int64(404)
		activityCount       = 250
	)

	db := exportTestDB(t)
	seedExportAthlete(t, db, requestingAthleteID)
	seedExportAthlete(t, db, otherAthleteID)
	seedExportActivities(t, db, requestingAthleteID, 2000, activityCount, "requester")
	seedExportActivities(t, db, otherAthleteID, 9500, 1, "other")

	handler := exportTestHandler(db, requestingAthleteID)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/export/activities/json", http.NoBody)
	recorder := httptest.NewRecorder()
	handler.ExportActivitiesJSON(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("ExportActivitiesJSON() status = %d, want %d; body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var activities []struct {
		ID        int64
		AthleteID int64
		Name      string
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &activities); err != nil {
		t.Fatalf("parse JSON export: %v; body = %s", err, recorder.Body.String())
	}
	if len(activities) != activityCount {
		t.Fatalf("JSON activity count = %d, want %d", len(activities), activityCount)
	}
	for _, activity := range activities {
		if activity.AthleteID != requestingAthleteID {
			t.Fatalf("JSON export included athlete %d activity %d (%q)", activity.AthleteID, activity.ID, activity.Name)
		}
	}
}

func exportTestDB(t *testing.T) *storage.DB {
	t.Helper()
	db, err := storage.OpenInMemory()
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	if err := db.Migrate(); err != nil {
		_ = db.Close()
		t.Fatalf("migrate in-memory database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedExportAthlete(t *testing.T, db *storage.DB, athleteID int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO athletes (id, username, firstname, lastname)
		VALUES (?, ?, 'Export', 'Test')
	`, athleteID, fmt.Sprintf("export-%d", athleteID))
	if err != nil {
		t.Fatalf("seed athlete %d: %v", athleteID, err)
	}
}

func seedExportActivities(t *testing.T, db *storage.DB, athleteID, firstActivityID int64, count int, namePrefix string) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin activity seed transaction: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	start := time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)
	for i := range count {
		activityTime := start.Add(time.Duration(i) * time.Hour).Format(time.RFC3339)
		_, err := tx.ExecContext(ctx, `
			INSERT INTO activities (
				id, athlete_id, name, description, sport_type, start_date, start_date_local,
				timezone, distance, moving_time, elapsed_time, device_name, gear_id,
				polyline, summary_polyline
			) VALUES (?, ?, ?, '', 'Ride', ?, ?, '', 1000, 100, 120, '', '', '', '')
		`, firstActivityID+int64(i), athleteID, fmt.Sprintf("%s-%03d", namePrefix, i), activityTime, activityTime)
		if err != nil {
			t.Fatalf("seed activity %d for athlete %d: %v", i, athleteID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit activity seed transaction: %v", err)
	}
}

func exportTestHandler(db *storage.DB, athleteID int64) *ExportHandler {
	repo := storage.NewActivityRepository(db)
	service := services.NewActivityService(repo, nil, nil)
	stravaClient := strava.NewClient(&config.StravaConfig{})
	stravaClient.SetToken(nil, &strava.Athlete{ID: athleteID})
	return NewExportHandler(service, stravaClient)
}
