package storage

import (
	"context"
	"encoding/json"
	"math"
	"testing"

	algorithms "github.com/melonamin/quantlete/algorithms/go"
	"github.com/melonamin/quantlete/internal/shared"
)

func encodeFloat64Array(data []float64) ([]byte, error) {
	return json.Marshal(data)
}

func TestTrainingLoadRepository_GetDailySeries_PadsRestDays(t *testing.T) {
	db := testDB(t)
	athleteID := int64(101)
	createTestAthlete(t, db, athleteID)

	insertActivityWithTSS(t, db, 1, athleteID, "2024-01-01", 50)
	insertActivityWithTSS(t, db, 2, athleteID, "2024-01-03", 100)

	repo := NewTrainingLoadRepository(db, nil, nil, nil)
	series, err := repo.GetDailySeries(context.Background(), athleteID, nil, nil)
	if err != nil {
		t.Fatalf("GetDailySeries error: %v", err)
	}
	if len(series) != 3 {
		t.Fatalf("expected 3 days, got %d", len(series))
	}

	expectedDays := []string{"2024-01-01", "2024-01-02", "2024-01-03"}
	expectedTSS := []float64{50, 0, 100}
	expectedLoad := algorithms.CalculateTrainingLoad(
		expectedTSS,
		shared.TrainingLoadCtlTauDays,
		shared.TrainingLoadAtlTauDays,
	)

	for i, day := range expectedDays {
		if series[i].Day != day {
			t.Fatalf("day %d: expected %s, got %s", i, day, series[i].Day)
		}
		if series[i].TSS != round2Value(expectedTSS[i]) {
			t.Fatalf("tss %d: expected %.2f, got %.2f", i, expectedTSS[i], series[i].TSS)
		}
		if series[i].CTL != round2Value(expectedLoad[i].CTL) {
			t.Fatalf("ctl %d: expected %.2f, got %.2f", i, round2Value(expectedLoad[i].CTL), series[i].CTL)
		}
		if series[i].ATL != round2Value(expectedLoad[i].ATL) {
			t.Fatalf("atl %d: expected %.2f, got %.2f", i, round2Value(expectedLoad[i].ATL), series[i].ATL)
		}
		if series[i].TSB != round2Value(expectedLoad[i].TSB) {
			t.Fatalf("tsb %d: expected %.2f, got %.2f", i, round2Value(expectedLoad[i].TSB), series[i].TSB)
		}
	}
}

func insertActivityWithTSS(t *testing.T, db *DB, activityID, athleteID int64, day string, tss float64) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO activities (id, athlete_id, name, sport_type, start_date, start_date_local)
		VALUES (?, ?, 'Test Activity', 'Ride', ?, ?)
	`, activityID, athleteID, day, day)
	if err != nil {
		t.Fatalf("failed to insert activity: %v", err)
	}

	_, err = db.ExecContext(context.Background(), `
		INSERT INTO activity_training_load (activity_id, athlete_id, tss)
		VALUES (?, ?, ?)
	`, activityID, athleteID, tss)
	if err != nil {
		t.Fatalf("failed to insert activity training load: %v", err)
	}
}

func round2Value(v float64) float64 {
	return math.Round(v*100) / 100
}

func TestTrainingLoadRepository_ComputeAndUpsertActivity_CyclingPower(t *testing.T) {
	db := testDB(t)
	athleteID := int64(102)
	activityID := int64(1001)
	createTestAthlete(t, db, athleteID)

	// Insert activity
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO activities (id, athlete_id, name, sport_type, start_date, start_date_local, moving_time)
		VALUES (?, ?, 'Test Ride', 'Ride', '2024-06-01', '2024-06-01', 3600)
	`, activityID, athleteID)
	if err != nil {
		t.Fatalf("failed to insert activity: %v", err)
	}

	// Insert power stream data (watts)
	wattsData := make([]float64, 100)
	for i := range wattsData {
		wattsData[i] = 200.0 // Constant 200W for simplicity
	}
	encoded, _ := encodeFloat64Array(wattsData)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO activity_streams (activity_id, stream_type, data, original_size, resolution, series_type)
		VALUES (?, 'watts', ?, ?, 'high', 'time')
	`, activityID, encoded, len(wattsData))
	if err != nil {
		t.Fatalf("failed to insert stream: %v", err)
	}

	// Insert FTP metric
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO athlete_metrics (athlete_id, metric, value, recorded_at)
		VALUES (?, 'ftp_cycling_watts', 250, '2024-01-01')
	`, athleteID)
	if err != nil {
		t.Fatalf("failed to insert FTP: %v", err)
	}

	streams := NewStreamRepository(db)
	metrics := NewAthleteMetricsRepository(db)
	repo := NewTrainingLoadRepository(db, streams, metrics, nil)

	// Run computation
	err = repo.EnsureComputedForRange(context.Background(), athleteID, nil, nil)
	if err != nil {
		t.Fatalf("EnsureComputedForRange error: %v", err)
	}

	// Verify TSS was computed
	var tss float64
	err = db.QueryRowContext(context.Background(), `
		SELECT tss FROM activity_training_load WHERE activity_id = ?
	`, activityID).Scan(&tss)
	if err != nil {
		t.Fatalf("failed to get computed TSS: %v", err)
	}
	if tss <= 0 {
		t.Errorf("expected positive TSS, got %.2f", tss)
	}

	// Verify method is cycling_power
	var method string
	err = db.QueryRowContext(context.Background(), `
		SELECT method FROM activity_training_load WHERE activity_id = ?
	`, activityID).Scan(&method)
	if err != nil {
		t.Fatalf("failed to get method: %v", err)
	}
	if method != "cycling_power" {
		t.Errorf("expected method cycling_power, got %s", method)
	}
}

func TestTrainingLoadRepository_ComputeAndUpsertActivity_SkipsWithoutFTP(t *testing.T) {
	db := testDB(t)
	athleteID := int64(103)
	activityID := int64(1002)
	createTestAthlete(t, db, athleteID)

	// Insert activity with power data but no FTP
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO activities (id, athlete_id, name, sport_type, start_date, start_date_local, moving_time)
		VALUES (?, ?, 'Test Ride', 'Ride', '2024-06-01', '2024-06-01', 3600)
	`, activityID, athleteID)
	if err != nil {
		t.Fatalf("failed to insert activity: %v", err)
	}

	// Insert power stream data
	wattsData := make([]float64, 100)
	for i := range wattsData {
		wattsData[i] = 200.0
	}
	encoded, _ := encodeFloat64Array(wattsData)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO activity_streams (activity_id, stream_type, data, original_size, resolution, series_type)
		VALUES (?, 'watts', ?, ?, 'high', 'time')
	`, activityID, encoded, len(wattsData))
	if err != nil {
		t.Fatalf("failed to insert stream: %v", err)
	}

	streams := NewStreamRepository(db)
	metrics := NewAthleteMetricsRepository(db)
	repo := NewTrainingLoadRepository(db, streams, metrics, nil)

	// Run computation - should not create a training load record without FTP
	err = repo.EnsureComputedForRange(context.Background(), athleteID, nil, nil)
	if err != nil {
		t.Fatalf("EnsureComputedForRange error: %v", err)
	}

	// Verify no TSS record was created
	var count int
	err = db.QueryRowContext(context.Background(), `
		SELECT COUNT(*) FROM activity_training_load WHERE activity_id = ?
	`, activityID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count training load records: %v", err)
	}
	if count != 0 {
		t.Errorf("expected no training load record without FTP, got %d", count)
	}
}

func TestTrainingLoadRepository_GetDiagnostics(t *testing.T) {
	db := testDB(t)
	athleteID := int64(104)
	createTestAthlete(t, db, athleteID)

	// Insert activities with different stream types
	activities := []struct {
		id        int64
		sportType string
		streams   []string
	}{
		{2001, "Ride", []string{"watts"}},
		{2002, "Ride", []string{"watts", "heartrate"}},
		{2003, "Run", []string{"velocity_smooth", "heartrate"}},
		{2004, "Run", []string{"heartrate"}},
	}

	for _, a := range activities {
		_, err := db.ExecContext(context.Background(), `
			INSERT INTO activities (id, athlete_id, name, sport_type, start_date, start_date_local, moving_time)
			VALUES (?, ?, 'Test', ?, '2024-06-01', '2024-06-01', 3600)
		`, a.id, athleteID, a.sportType)
		if err != nil {
			t.Fatalf("failed to insert activity %d: %v", a.id, err)
		}

		for _, streamType := range a.streams {
			streamData := []float64{100, 100, 100}
			data, _ := encodeFloat64Array(streamData)
			_, err = db.ExecContext(context.Background(), `
				INSERT INTO activity_streams (activity_id, stream_type, data, original_size, resolution, series_type)
				VALUES (?, ?, ?, ?, 'high', 'time')
			`, a.id, streamType, data, len(streamData))
			if err != nil {
				t.Fatalf("failed to insert stream %s for activity %d: %v", streamType, a.id, err)
			}
		}
	}

	// Insert one TSS record
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO activity_training_load (activity_id, athlete_id, tss, method)
		VALUES (?, ?, 50, 'cycling_power')
	`, activities[0].id, athleteID)
	if err != nil {
		t.Fatalf("failed to insert training load: %v", err)
	}

	streams := NewStreamRepository(db)
	metrics := NewAthleteMetricsRepository(db)
	repo := NewTrainingLoadRepository(db, streams, metrics, nil)

	diag, err := repo.GetDiagnostics(context.Background(), athleteID)
	if err != nil {
		t.Fatalf("GetDiagnostics error: %v", err)
	}

	if diag.TotalActivities != 4 {
		t.Errorf("expected 4 total activities, got %d", diag.TotalActivities)
	}
	if diag.ActivitiesWithPower != 2 {
		t.Errorf("expected 2 activities with power, got %d", diag.ActivitiesWithPower)
	}
	if diag.ActivitiesWithSpeed != 1 {
		t.Errorf("expected 1 activity with speed, got %d", diag.ActivitiesWithSpeed)
	}
	if diag.ActivitiesWithHR != 3 {
		t.Errorf("expected 3 activities with HR, got %d", diag.ActivitiesWithHR)
	}
	if diag.ActivitiesWithTSS != 1 {
		t.Errorf("expected 1 activity with TSS, got %d", diag.ActivitiesWithTSS)
	}
	if diag.HasCyclingFTP {
		t.Error("expected HasCyclingFTP to be false")
	}
	if diag.HasRunningFTP {
		t.Error("expected HasRunningFTP to be false")
	}

	// Should have warning about missing cycling FTP
	if len(diag.MissingConfigWarnings) == 0 {
		t.Error("expected missing config warnings")
	}
}

func TestTrainingLoadRepository_GetDiagnostics_WithFTP(t *testing.T) {
	db := testDB(t)
	athleteID := int64(105)
	createTestAthlete(t, db, athleteID)

	// Insert cycling FTP
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO athlete_metrics (athlete_id, metric, value, recorded_at)
		VALUES (?, 'ftp_cycling_watts', 280, '2024-01-01')
	`, athleteID)
	if err != nil {
		t.Fatalf("failed to insert FTP: %v", err)
	}

	// Insert running FTP (threshold pace in m/s, e.g., 4:00/km = 4.17 m/s)
	_, err = db.ExecContext(context.Background(), `
		INSERT INTO athlete_metrics (athlete_id, metric, value, recorded_at)
		VALUES (?, 'ftp_running_mps', 4.17, '2024-01-01')
	`, athleteID)
	if err != nil {
		t.Fatalf("failed to insert running FTP: %v", err)
	}

	streams := NewStreamRepository(db)
	metrics := NewAthleteMetricsRepository(db)
	repo := NewTrainingLoadRepository(db, streams, metrics, nil)

	diag, err := repo.GetDiagnostics(context.Background(), athleteID)
	if err != nil {
		t.Fatalf("GetDiagnostics error: %v", err)
	}

	if !diag.HasCyclingFTP {
		t.Error("expected HasCyclingFTP to be true")
	}
	if diag.CyclingFTPValue == nil || *diag.CyclingFTPValue != 280 {
		t.Errorf("expected cycling FTP 280, got %v", diag.CyclingFTPValue)
	}
	if !diag.HasRunningFTP {
		t.Error("expected HasRunningFTP to be true")
	}
	if diag.RunningFTPValue == nil || *diag.RunningFTPValue != 4.17 {
		t.Errorf("expected running FTP 4.17, got %v", diag.RunningFTPValue)
	}
	if len(diag.MissingConfigWarnings) != 0 {
		t.Errorf("expected no warnings with FTP configured, got %v", diag.MissingConfigWarnings)
	}
}
