package storage

import (
	"context"
	"testing"
)

func TestNotificationHistoryRepository_HasBeenNotified(t *testing.T) {
	db := testDB(t)
	repo := NewNotificationHistoryRepository(db)
	ctx := context.Background()

	// Create required foreign key reference
	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)

	achievementType := "personal_record"
	key := "pr_5k_activity_123"

	// Should return false for non-existent record
	notified, err := repo.HasBeenNotified(ctx, athleteID, achievementType, key)
	if err != nil {
		t.Fatalf("HasBeenNotified() error = %v", err)
	}
	if notified {
		t.Error("HasBeenNotified() = true for non-existent record, want false")
	}

	// Mark as notified
	err = repo.MarkNotified(ctx, athleteID, achievementType, key)
	if err != nil {
		t.Fatalf("MarkNotified() error = %v", err)
	}

	// Should return true after marking
	notified, err = repo.HasBeenNotified(ctx, athleteID, achievementType, key)
	if err != nil {
		t.Fatalf("HasBeenNotified() after mark error = %v", err)
	}
	if !notified {
		t.Error("HasBeenNotified() = false after marking, want true")
	}
}

func TestNotificationHistoryRepository_MarkNotified_Idempotent(t *testing.T) {
	db := testDB(t)
	repo := NewNotificationHistoryRepository(db)
	ctx := context.Background()

	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)

	achievementType := "segment_pr"
	key := "segment_456_activity_789"

	// Mark multiple times - should not error
	for i := 0; i < 3; i++ {
		if err := repo.MarkNotified(ctx, athleteID, achievementType, key); err != nil {
			t.Fatalf("MarkNotified() iteration %d error = %v", i, err)
		}
	}

	// Verify only one record exists
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM notification_history WHERE athlete_id = ? AND achievement_type = ? AND achievement_key = ?",
		athleteID, achievementType, key).Scan(&count)
	if err != nil {
		t.Fatalf("count query error = %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}
}

func TestNotificationHistoryRepository_MarkNotifiedBatch(t *testing.T) {
	db := testDB(t)
	repo := NewNotificationHistoryRepository(db)
	ctx := context.Background()

	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)

	items := []NotificationHistoryItem{
		{AthleteID: athleteID, AchievementType: "personal_record", AchievementKey: "pr_5k_1"},
		{AthleteID: athleteID, AchievementType: "personal_record", AchievementKey: "pr_10k_1"},
		{AthleteID: athleteID, AchievementType: "eddington_increase", AchievementKey: "eddington_51"},
	}

	// Batch mark
	if err := repo.MarkNotifiedBatch(ctx, athleteID, items); err != nil {
		t.Fatalf("MarkNotifiedBatch() error = %v", err)
	}

	// Verify all were marked
	for _, item := range items {
		notified, err := repo.HasBeenNotified(ctx, athleteID, item.AchievementType, item.AchievementKey)
		if err != nil {
			t.Fatalf("HasBeenNotified() for %s error = %v", item.AchievementKey, err)
		}
		if !notified {
			t.Errorf("HasBeenNotified() for %s = false, want true", item.AchievementKey)
		}
	}
}

func TestNotificationHistoryRepository_MarkNotifiedBatch_Empty(t *testing.T) {
	db := testDB(t)
	repo := NewNotificationHistoryRepository(db)
	ctx := context.Background()

	athleteID := int64(12345)

	// Empty batch should not error
	if err := repo.MarkNotifiedBatch(ctx, athleteID, nil); err != nil {
		t.Fatalf("MarkNotifiedBatch(nil) error = %v", err)
	}

	if err := repo.MarkNotifiedBatch(ctx, athleteID, []NotificationHistoryItem{}); err != nil {
		t.Fatalf("MarkNotifiedBatch([]) error = %v", err)
	}
}

func TestNotificationHistoryRepository_FilterNotified(t *testing.T) {
	db := testDB(t)
	repo := NewNotificationHistoryRepository(db)
	ctx := context.Background()

	athleteID := int64(12345)
	createTestAthlete(t, db, athleteID)

	// Pre-populate some already-notified items
	existingItems := []NotificationHistoryItem{
		{AthleteID: athleteID, AchievementType: "personal_record", AchievementKey: "pr_5k_old"},
		{AthleteID: athleteID, AchievementType: "eddington_increase", AchievementKey: "eddington_50"},
	}
	if err := repo.MarkNotifiedBatch(ctx, athleteID, existingItems); err != nil {
		t.Fatalf("setup MarkNotifiedBatch() error = %v", err)
	}

	// Create a mix of new and already-notified items
	toFilter := []NotificationHistoryItem{
		{AthleteID: athleteID, AchievementType: "personal_record", AchievementKey: "pr_5k_old"},       // Already notified
		{AthleteID: athleteID, AchievementType: "personal_record", AchievementKey: "pr_5k_new"},       // New
		{AthleteID: athleteID, AchievementType: "eddington_increase", AchievementKey: "eddington_50"}, // Already notified
		{AthleteID: athleteID, AchievementType: "eddington_increase", AchievementKey: "eddington_51"}, // New
		{AthleteID: athleteID, AchievementType: "gear_milestone", AchievementKey: "gear_5000km"},      // New
	}

	result, err := repo.FilterNotified(ctx, athleteID, toFilter)
	if err != nil {
		t.Fatalf("FilterNotified() error = %v", err)
	}

	// Should have 3 new items
	if len(result) != 3 {
		t.Fatalf("FilterNotified() returned %d items, want 3", len(result))
	}

	// Verify the correct items are returned
	expectedKeys := map[string]bool{
		"personal_record:pr_5k_new":       true,
		"eddington_increase:eddington_51": true,
		"gear_milestone:gear_5000km":      true,
	}
	for _, item := range result {
		key := item.AchievementType + ":" + item.AchievementKey
		if !expectedKeys[key] {
			t.Errorf("unexpected item in result: %s", key)
		}
		delete(expectedKeys, key)
	}
	if len(expectedKeys) > 0 {
		t.Errorf("missing items from result: %v", expectedKeys)
	}
}

func TestNotificationHistoryRepository_FilterNotified_Empty(t *testing.T) {
	db := testDB(t)
	repo := NewNotificationHistoryRepository(db)
	ctx := context.Background()

	athleteID := int64(12345)

	// Empty input should return nil
	result, err := repo.FilterNotified(ctx, athleteID, nil)
	if err != nil {
		t.Fatalf("FilterNotified(nil) error = %v", err)
	}
	if result != nil {
		t.Errorf("FilterNotified(nil) = %v, want nil", result)
	}

	result, err = repo.FilterNotified(ctx, athleteID, []NotificationHistoryItem{})
	if err != nil {
		t.Fatalf("FilterNotified([]) error = %v", err)
	}
	if result != nil {
		t.Errorf("FilterNotified([]) = %v, want nil", result)
	}
}

func TestNotificationHistoryRepository_AthleteIsolation(t *testing.T) {
	db := testDB(t)
	repo := NewNotificationHistoryRepository(db)
	ctx := context.Background()

	athlete1 := int64(11111)
	athlete2 := int64(22222)
	createTestAthlete(t, db, athlete1)
	createTestAthlete(t, db, athlete2)

	achievementType := "personal_record"
	key := "pr_5k_shared"

	// Mark for athlete1
	if err := repo.MarkNotified(ctx, athlete1, achievementType, key); err != nil {
		t.Fatalf("MarkNotified() athlete1 error = %v", err)
	}

	// Should be notified for athlete1
	notified, err := repo.HasBeenNotified(ctx, athlete1, achievementType, key)
	if err != nil {
		t.Fatalf("HasBeenNotified() athlete1 error = %v", err)
	}
	if !notified {
		t.Error("HasBeenNotified() athlete1 = false, want true")
	}

	// Should NOT be notified for athlete2
	notified, err = repo.HasBeenNotified(ctx, athlete2, achievementType, key)
	if err != nil {
		t.Fatalf("HasBeenNotified() athlete2 error = %v", err)
	}
	if notified {
		t.Error("HasBeenNotified() athlete2 = true, want false (different athlete)")
	}
}
