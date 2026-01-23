package storage

import (
	"context"
	"testing"

	algorithms "github.com/melonamin/quantlete/algorithms/go"
)

func TestStatsRepository_GetEddingtonHistory_UsesAlgorithm(t *testing.T) {
	db := testDB(t)
	athleteID := int64(202)
	createTestAthlete(t, db, athleteID)

	insertActivityWithDistance(t, db, 1, athleteID, "2024-01-01", 10000)
	insertActivityWithDistance(t, db, 2, athleteID, "2024-01-02", 20000)
	insertActivityWithDistance(t, db, 3, athleteID, "2024-01-03", 5000)

	repo := NewStatsRepository(db)
	history, err := repo.GetEddingtonHistory(context.Background(), athleteID, nil)
	if err != nil {
		t.Fatalf("GetEddingtonHistory error: %v", err)
	}

	distances := []float64{10, 20, 5}
	expected := algorithms.EddingtonHistory(distances)
	if len(history) != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), len(history))
	}

	expectedDates := []string{"2024-01-01", "2024-01-02", "2024-01-03"}
	for i, point := range history {
		if point.Date != expectedDates[i] {
			t.Fatalf("date %d: expected %s, got %s", i, expectedDates[i], point.Date)
		}
		if point.Number != expected[i] {
			t.Fatalf("number %d: expected %d, got %d", i, expected[i], point.Number)
		}
	}
}

func insertActivityWithDistance(t *testing.T, db *DB, activityID, athleteID int64, day string, distanceM float64) {
	t.Helper()

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO activities (id, athlete_id, name, sport_type, start_date, start_date_local, distance)
		VALUES (?, ?, 'Test Activity', 'Ride', ?, ?, ?)
	`, activityID, athleteID, day, day, distanceM)
	if err != nil {
		t.Fatalf("failed to insert activity: %v", err)
	}
}
