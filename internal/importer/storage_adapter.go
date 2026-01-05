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
