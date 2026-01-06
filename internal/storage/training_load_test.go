package storage

import (
	"context"
	"math"
	"testing"

	algorithms "github.com/melonamin/quantlete/algorithms/go"
	"github.com/melonamin/quantlete/internal/shared"
)

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
