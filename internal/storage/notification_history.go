package storage

import (
	"context"
	"time"
)

// NotificationHistoryItem represents an achievement notification record.
type NotificationHistoryItem struct {
	AthleteID       int64
	AchievementType string
	AchievementKey  string
}

// NotificationHistoryRepository provides access to notification history storage.
type NotificationHistoryRepository struct {
	db      *DB
	queries *Queries
}

// NewNotificationHistoryRepository creates a new notification history repository.
func NewNotificationHistoryRepository(db *DB) *NotificationHistoryRepository {
	return &NotificationHistoryRepository{
		db:      db,
		queries: NewQueries(db),
	}
}

// HasBeenNotified checks if an achievement has already been notified.
func (r *NotificationHistoryRepository) HasBeenNotified(ctx context.Context, athleteID int64, achievementType, key string) (bool, error) {
	row, err := r.queries.HasBeenNotified(ctx, athleteID, achievementType, key)
	if err != nil {
		return false, err
	}
	if row == nil {
		return false, nil
	}
	return row.Count > 0, nil
}

// MarkNotified records that an achievement has been notified.
func (r *NotificationHistoryRepository) MarkNotified(ctx context.Context, athleteID int64, achievementType, key string) error {
	return r.queries.MarkNotified(ctx, athleteID, achievementType, key, SQLiteTime{Time: time.Now()})
}

// MarkNotifiedBatch records multiple achievements as notified in a single transaction.
func (r *NotificationHistoryRepository) MarkNotifiedBatch(ctx context.Context, athleteID int64, items []NotificationHistoryItem) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	txQueries := r.queries.WithTx(tx)
	now := SQLiteTime{Time: time.Now()}
	for _, item := range items {
		if err := txQueries.MarkNotified(ctx, athleteID, item.AchievementType, item.AchievementKey, now); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// FilterNotified filters out achievements that have already been notified.
// Returns only the items that have NOT been notified yet.
// Uses batch lookup to avoid N+1 queries.
func (r *NotificationHistoryRepository) FilterNotified(ctx context.Context, athleteID int64, items []NotificationHistoryItem) ([]NotificationHistoryItem, error) {
	if len(items) == 0 {
		return nil, nil
	}

	// Batch fetch all notified keys for this athlete
	notifiedRows, err := r.queries.GetNotifiedKeys(ctx, athleteID)
	if err != nil {
		return nil, err
	}

	// Build a set of notified (type, key) pairs for O(1) lookup
	notifiedSet := make(map[string]struct{}, len(notifiedRows))
	for _, row := range notifiedRows {
		notifiedSet[row.AchievementType+":"+row.AchievementKey] = struct{}{}
	}

	// Filter out already-notified items
	var result []NotificationHistoryItem
	for _, item := range items {
		key := item.AchievementType + ":" + item.AchievementKey
		if _, notified := notifiedSet[key]; !notified {
			result = append(result, item)
		}
	}
	return result, nil
}
