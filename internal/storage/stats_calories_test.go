package storage

import (
	"context"
	"testing"
)

func TestGeneratedStatsQueriesTruncateFractionalCalories(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)

	_, err := db.ExecContext(ctx, `
		INSERT INTO activities (
			id, athlete_id, name, sport_type, start_date, start_date_local,
			distance, moving_time, elapsed_time, calories
		) VALUES (?, ?, 'Fractional calories', 'Ride', ?, ?, 1000, 300, 320, ?)
	`, int64(67890), athleteID, "2026-07-19T12:00:00Z", "2026-07-19T08:00:00", 611.66)
	if err != nil {
		t.Fatalf("insert activity: %v", err)
	}

	queries := NewQueries(db)

	yearDays, err := queries.GetCalendarData(ctx, athleteID, "2026")
	if err != nil {
		t.Fatalf("GetCalendarData() error = %v", err)
	}
	if len(yearDays) != 1 {
		t.Fatalf("GetCalendarData() returned %d days, want 1", len(yearDays))
	}
	if yearDays[0].TotalCalories != 611 {
		t.Errorf("GetCalendarData() total calories = %d, want 611", yearDays[0].TotalCalories)
	}

	days, err := queries.GetCalendarDataRange(ctx, athleteID, "2026-07-01", "2026-07-31")
	if err != nil {
		t.Fatalf("GetCalendarDataRange() error = %v", err)
	}
	if len(days) != 1 {
		t.Fatalf("GetCalendarDataRange() returned %d days, want 1", len(days))
	}
	if days[0].TotalCalories != 611 {
		t.Errorf("GetCalendarDataRange() total calories = %d, want 611", days[0].TotalCalories)
	}

	calendarSummary, err := queries.GetCalendarSummary(ctx, athleteID, "2026", "07")
	if err != nil {
		t.Fatalf("GetCalendarSummary() error = %v", err)
	}
	if calendarSummary == nil {
		t.Fatal("GetCalendarSummary() returned nil")
	}
	if calendarSummary.TotalCalories != 611 {
		t.Errorf("GetCalendarSummary() total calories = %d, want 611", calendarSummary.TotalCalories)
	}

	dashboardSummary, err := queries.GetDashboardStats(ctx, athleteID)
	if err != nil {
		t.Fatalf("GetDashboardStats() error = %v", err)
	}
	if dashboardSummary == nil {
		t.Fatal("GetDashboardStats() returned nil")
	}
	if dashboardSummary.TotalCalories != 611 {
		t.Errorf("GetDashboardStats() total calories = %d, want 611", dashboardSummary.TotalCalories)
	}
}
