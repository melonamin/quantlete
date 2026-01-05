// Package importer handles importing data from Strava.
package importer

import (
	"context"

	"github.com/melonamin/quantlete/internal/storage"
)

// AchievementStorageAdapter wraps a storage.AchievementRepository to implement
// the importer.AchievementStorage interface. This adapter handles type conversions
// between storage layer types and importer types.
type AchievementStorageAdapter struct {
	repo *storage.AchievementRepository
}

// NewAchievementStorageAdapter creates a new adapter wrapping the storage repository.
func NewAchievementStorageAdapter(repo *storage.AchievementRepository) *AchievementStorageAdapter {
	return &AchievementStorageAdapter{repo: repo}
}

// GetEddingtonNumber returns the current Eddington number for the athlete.
func (a *AchievementStorageAdapter) GetEddingtonNumber(ctx context.Context, athleteID int64, sportTypes []string) (int, error) {
	return a.repo.GetEddingtonNumber(ctx, athleteID, sportTypes)
}

// GetGearDistances returns current distances for all gear owned by the athlete.
func (a *AchievementStorageAdapter) GetGearDistances(ctx context.Context, athleteID int64) (map[string]float64, error) {
	return a.repo.GetGearDistances(ctx, athleteID)
}

// GetBestEffortPRsForActivities returns best efforts with pr_rank=1 for the given activity IDs.
func (a *AchievementStorageAdapter) GetBestEffortPRsForActivities(ctx context.Context, athleteID int64, activityIDs []int64) ([]BestEffortPRInfo, error) {
	prs, err := a.repo.GetBestEffortPRsForActivities(ctx, athleteID, activityIDs)
	if err != nil {
		return nil, err
	}

	result := make([]BestEffortPRInfo, len(prs))
	for i, pr := range prs {
		result[i] = BestEffortPRInfo{
			ActivityID:   pr.ActivityID,
			ActivityName: pr.ActivityName,
			DistanceType: pr.DistanceType,
			Name:         pr.Name,
			ElapsedTime:  pr.ElapsedTime,
		}
	}
	return result, nil
}

// GetSegmentPRsForActivities returns segment efforts with pr_rank=1 for the given activity IDs.
func (a *AchievementStorageAdapter) GetSegmentPRsForActivities(ctx context.Context, athleteID int64, activityIDs []int64) ([]SegmentPRInfo, error) {
	prs, err := a.repo.GetSegmentPRsForActivities(ctx, athleteID, activityIDs)
	if err != nil {
		return nil, err
	}

	result := make([]SegmentPRInfo, len(prs))
	for i, pr := range prs {
		result[i] = SegmentPRInfo{
			ActivityID:   pr.ActivityID,
			ActivityName: pr.ActivityName,
			SegmentID:    pr.SegmentID,
			SegmentName:  pr.SegmentName,
			ElapsedTime:  pr.ElapsedTime,
		}
	}
	return result, nil
}

// GetPowerRecords returns all-time best power values for standard durations.
func (a *AchievementStorageAdapter) GetPowerRecords(ctx context.Context, athleteID int64) (map[int]float64, error) {
	return a.repo.GetPowerRecords(ctx, athleteID)
}

// GetPowerRecordsForActivities returns power records set in the given activity IDs.
func (a *AchievementStorageAdapter) GetPowerRecordsForActivities(ctx context.Context, athleteID int64, activityIDs []int64, durations []int) ([]PowerRecordInfo, error) {
	records, err := a.repo.GetPowerRecordsForActivities(ctx, athleteID, activityIDs, durations)
	if err != nil {
		return nil, err
	}

	result := make([]PowerRecordInfo, len(records))
	for i, rec := range records {
		result[i] = PowerRecordInfo{
			ActivityID: rec.ActivityID,
			DurationS:  rec.DurationS,
			Watts:      rec.Watts,
		}
	}
	return result, nil
}

// GetLatestTrainingLoad returns the most recent training load metrics.
func (a *AchievementStorageAdapter) GetLatestTrainingLoad(ctx context.Context, athleteID int64) (*TrainingLoadInfo, error) {
	load, err := a.repo.GetLatestTrainingLoad(ctx, athleteID)
	if err != nil {
		return nil, err
	}
	if load == nil {
		return nil, nil
	}

	return &TrainingLoadInfo{
		Day: load.Day,
		CTL: load.CTL,
		ATL: load.ATL,
		TSB: load.TSB,
	}, nil
}

// NotificationHistoryAdapter wraps a storage.NotificationHistoryRepository to implement
// the importer.NotificationHistoryChecker interface.
type NotificationHistoryAdapter struct {
	repo *storage.NotificationHistoryRepository
}

// NewNotificationHistoryAdapter creates a new adapter wrapping the storage repository.
func NewNotificationHistoryAdapter(repo *storage.NotificationHistoryRepository) *NotificationHistoryAdapter {
	return &NotificationHistoryAdapter{repo: repo}
}

// HasBeenNotified checks if an achievement has already been notified.
func (a *NotificationHistoryAdapter) HasBeenNotified(ctx context.Context, athleteID int64, achievementType, key string) (bool, error) {
	return a.repo.HasBeenNotified(ctx, athleteID, achievementType, key)
}

// FilterNotified filters out achievements that have already been notified.
// Returns only achievements that have NOT been notified yet.
// Uses batch lookup to avoid N+1 queries.
func (a *NotificationHistoryAdapter) FilterNotified(ctx context.Context, athleteID int64, achievements []Achievement) ([]Achievement, error) {
	if len(achievements) == 0 {
		return nil, nil
	}

	// Convert to storage items, keeping track of original achievements
	items := make([]storage.NotificationHistoryItem, 0, len(achievements))
	// Track achievements without keys - these will always be included
	var keylessAchievements []Achievement

	for _, ach := range achievements {
		if ach.Key == "" {
			keylessAchievements = append(keylessAchievements, ach)
			continue
		}
		items = append(items, storage.NotificationHistoryItem{
			AthleteID:       athleteID,
			AchievementType: string(ach.Type),
			AchievementKey:  ach.Key,
		})
	}

	// If all achievements lack keys, return them all
	if len(items) == 0 {
		return keylessAchievements, nil
	}

	// Use batch lookup in storage layer
	notNotified, err := a.repo.FilterNotified(ctx, athleteID, items)
	if err != nil {
		return nil, err
	}

	// Build set of not-notified (type, key) pairs for O(1) lookup
	notNotifiedSet := make(map[string]struct{}, len(notNotified))
	for _, item := range notNotified {
		notNotifiedSet[item.AchievementType+":"+item.AchievementKey] = struct{}{}
	}

	// Filter original achievements, keeping keyless ones and not-notified ones
	result := make([]Achievement, 0, len(keylessAchievements)+len(notNotified))
	result = append(result, keylessAchievements...)

	for _, ach := range achievements {
		if ach.Key == "" {
			continue // Already added keyless achievements
		}
		key := string(ach.Type) + ":" + ach.Key
		if _, ok := notNotifiedSet[key]; ok {
			result = append(result, ach)
		}
	}

	return result, nil
}

// MarkNotified records that an achievement has been notified.
func (a *NotificationHistoryAdapter) MarkNotified(ctx context.Context, athleteID int64, achievementType, key string) error {
	return a.repo.MarkNotified(ctx, athleteID, achievementType, key)
}

// MarkNotifiedBatch records multiple achievements as notified.
func (a *NotificationHistoryAdapter) MarkNotifiedBatch(ctx context.Context, athleteID int64, achievements []Achievement) error {
	if len(achievements) == 0 {
		return nil
	}

	items := make([]storage.NotificationHistoryItem, 0, len(achievements))
	for _, ach := range achievements {
		if ach.Key == "" {
			continue // Skip achievements without keys
		}
		items = append(items, storage.NotificationHistoryItem{
			AthleteID:       athleteID,
			AchievementType: string(ach.Type),
			AchievementKey:  ach.Key,
		})
	}

	// Early return if all achievements had empty keys
	if len(items) == 0 {
		return nil
	}

	return a.repo.MarkNotifiedBatch(ctx, athleteID, items)
}
