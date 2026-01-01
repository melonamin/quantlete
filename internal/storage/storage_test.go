package storage

import (
	"context"
	"testing"
	"time"
)

// testDB creates an in-memory database with schema for testing.
func testDB(t *testing.T) *DB {
	t.Helper()

	db, err := OpenInMemory()
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}

	// Run migrations to create schema
	if err := db.Migrate(); err != nil {
		_ = db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

// createTestAthlete creates a test athlete record for foreign key constraints.
func createTestAthlete(t *testing.T, db *DB, athleteID int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO athletes (id, username, firstname, lastname)
		VALUES (?, 'testuser', 'Test', 'User')
	`, athleteID)
	if err != nil {
		t.Fatalf("failed to create test athlete: %v", err)
	}
}

// createTestActivity creates a test activity record for foreign key constraints.
func createTestActivity(t *testing.T, db *DB, activityID, athleteID int64) {
	t.Helper()
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO activities (id, athlete_id, name, sport_type, start_date, start_date_local, distance, moving_time, elapsed_time)
		VALUES (?, ?, 'Test Activity', 'Ride', datetime('now'), datetime('now'), 1000, 100, 120)
	`, activityID, athleteID)
	if err != nil {
		t.Fatalf("failed to create test activity: %v", err)
	}
}

func TestSegmentRepository_UpsertSegmentWithEffort(t *testing.T) {
	db := testDB(t)
	repo := NewSegmentRepository(db)
	ctx := context.Background()

	// Create required foreign key references
	athleteID := int64(99999)
	activityID := int64(11111)
	createTestAthlete(t, db, athleteID)
	createTestActivity(t, db, activityID, athleteID)

	segment := &Segment{
		ID:           12345,
		Name:         "Test Segment",
		ActivityType: "Ride",
		Distance:     1000.0,
		AverageGrade: 5.0,
		MaximumGrade: 10.0,
	}

	now := time.Now()
	startDate := SQLiteTime{Time: now}
	effort := &SegmentEffort{
		ID:             67890,
		SegmentID:      segment.ID,
		ActivityID:     activityID,
		AthleteID:      athleteID,
		Name:           "Test Segment",
		ElapsedTime:    120,
		MovingTime:     115,
		StartDate:      &startDate,
		StartDateLocal: &startDate,
		Distance:       1000.0,
	}

	// Test successful upsert
	err := repo.UpsertSegmentWithEffort(ctx, segment, effort)
	if err != nil {
		t.Fatalf("UpsertSegmentWithEffort() error = %v", err)
	}

	// Verify segment was stored
	stored, err := repo.GetByID(ctx, segment.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if stored == nil {
		t.Fatal("expected segment to be stored")
	}
	if stored.Name != segment.Name {
		t.Errorf("stored segment name = %q, want %q", stored.Name, segment.Name)
	}

	// Verify effort was stored
	efforts, err := repo.ListEfforts(ctx, effort.AthleteID, segment.ID, 10)
	if err != nil {
		t.Fatalf("ListEfforts() error = %v", err)
	}
	if len(efforts) != 1 {
		t.Fatalf("expected 1 effort, got %d", len(efforts))
	}
	if efforts[0].ID != effort.ID {
		t.Errorf("stored effort ID = %d, want %d", efforts[0].ID, effort.ID)
	}
}

func TestSegmentRepository_UpsertSegmentWithEffort_Idempotent(t *testing.T) {
	db := testDB(t)
	repo := NewSegmentRepository(db)
	ctx := context.Background()

	// Create required foreign key references
	athleteID := int64(99999)
	activityID := int64(11111)
	createTestAthlete(t, db, athleteID)
	createTestActivity(t, db, activityID, athleteID)

	segment := &Segment{
		ID:           12345,
		Name:         "Original Name",
		ActivityType: "Ride",
		Distance:     1000.0,
	}

	now := time.Now()
	startDate := SQLiteTime{Time: now}
	effort := &SegmentEffort{
		ID:         67890,
		SegmentID:  segment.ID,
		ActivityID: activityID,
		AthleteID:  athleteID,
		Name:       "Original Effort",
		StartDate:  &startDate,
	}

	// First upsert
	if err := repo.UpsertSegmentWithEffort(ctx, segment, effort); err != nil {
		t.Fatalf("first UpsertSegmentWithEffort() error = %v", err)
	}

	// Update segment and effort
	segment.Name = "Updated Name"
	effort.Name = "Updated Effort"

	// Second upsert (should update, not duplicate)
	if err := repo.UpsertSegmentWithEffort(ctx, segment, effort); err != nil {
		t.Fatalf("second UpsertSegmentWithEffort() error = %v", err)
	}

	// Verify segment was updated
	stored, err := repo.GetByID(ctx, segment.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if stored.Name != "Updated Name" {
		t.Errorf("segment name = %q, want %q", stored.Name, "Updated Name")
	}

	// Verify still only one effort (not duplicated)
	efforts, err := repo.ListEfforts(ctx, effort.AthleteID, segment.ID, 10)
	if err != nil {
		t.Fatalf("ListEfforts() error = %v", err)
	}
	if len(efforts) != 1 {
		t.Errorf("expected 1 effort, got %d", len(efforts))
	}
	if efforts[0].Name != "Updated Effort" {
		t.Errorf("effort name = %q, want %q", efforts[0].Name, "Updated Effort")
	}
}

func TestTokenRepository_Upsert_Transaction(t *testing.T) {
	db := testDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// Create required foreign key reference
	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)

	token := &AuthToken{
		AthleteID:    athleteID,
		AccessToken:  "access_token_1",
		RefreshToken: "refresh_token_1",
		TokenType:    "Bearer",
		ExpiresAt:    SQLiteTime{Time: time.Now().Add(time.Hour)},
	}

	// First insert
	if err := repo.Upsert(ctx, token); err != nil {
		t.Fatalf("first Upsert() error = %v", err)
	}

	// Verify token was stored
	stored, err := repo.GetByAthleteID(ctx, token.AthleteID)
	if err != nil {
		t.Fatalf("GetByAthleteID() error = %v", err)
	}
	if stored == nil {
		t.Fatal("expected token to be stored")
	}
	if stored.AccessToken != token.AccessToken {
		t.Errorf("access token = %q, want %q", stored.AccessToken, token.AccessToken)
	}

	// Update token
	token.AccessToken = "access_token_2"
	token.RefreshToken = "refresh_token_2"

	// Second upsert (should update, not leave orphaned)
	if upsertErr := repo.Upsert(ctx, token); upsertErr != nil {
		t.Fatalf("second Upsert() error = %v", upsertErr)
	}

	// Verify token was updated
	stored, err = repo.GetByAthleteID(ctx, token.AthleteID)
	if err != nil {
		t.Fatalf("GetByAthleteID() after update error = %v", err)
	}
	if stored.AccessToken != "access_token_2" {
		t.Errorf("access token = %q, want %q", stored.AccessToken, "access_token_2")
	}

	// Verify only one token exists (count all tokens for this athlete)
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM auth_tokens WHERE athlete_id = ?", token.AthleteID).Scan(&count)
	if err != nil {
		t.Fatalf("count query error = %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 token, got %d", count)
	}
}

func TestAppStateRepository_SaveLoad(t *testing.T) {
	db := testDB(t)
	repo := NewAppStateRepository(db)
	ctx := context.Background()

	key := "test_key"
	value := `{"phase":"streams","activityIds":[1,2,3]}`

	// Save state
	if err := repo.Set(ctx, key, value); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Load state
	loaded, err := repo.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if loaded != value {
		t.Errorf("loaded value = %q, want %q", loaded, value)
	}

	// Update state
	newValue := `{"phase":"details","activityIds":[1,2,3,4]}`
	if setErr := repo.Set(ctx, key, newValue); setErr != nil {
		t.Fatalf("Set() update error = %v", setErr)
	}

	// Verify update
	loaded, err = repo.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() after update error = %v", err)
	}
	if loaded != newValue {
		t.Errorf("loaded value = %q, want %q", loaded, newValue)
	}

	// Delete state
	if delErr := repo.Delete(ctx, key); delErr != nil {
		t.Fatalf("Delete() error = %v", delErr)
	}

	// Verify deletion
	loaded, err = repo.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() after delete error = %v", err)
	}
	if loaded != "" {
		t.Errorf("expected empty value after delete, got %q", loaded)
	}
}

func TestSQLiteTime_Scan(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantErr bool
	}{
		{"nil", nil, false},
		{"string ISO8601", "2024-01-15T10:30:00Z", false},
		{"string date only", "2024-01-15", false},
		{"string datetime", "2024-01-15 10:30:00", false},
		{"time.Time", time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC), false},
		{"invalid type", 12345, true},
		{"invalid string", "not a date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var st SQLiteTime
			err := st.Scan(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
