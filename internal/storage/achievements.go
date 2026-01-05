package storage

import (
	"context"
	"strings"
)

// AchievementRepository provides data for achievement detection during import.
// This repository aggregates queries from multiple sources (best efforts, segments,
// gear, power, training load) needed by the importer.AchievementStorage interface.
type AchievementRepository struct {
	db           *DB
	stats        *StatsRepository
	gear         *GearRepository
	power        *PowerRepository
	trainingLoad *TrainingLoadRepository
}

// NewAchievementRepository creates a new achievement repository.
func NewAchievementRepository(
	db *DB,
	stats *StatsRepository,
	gear *GearRepository,
	power *PowerRepository,
	trainingLoad *TrainingLoadRepository,
) *AchievementRepository {
	return &AchievementRepository{
		db:           db,
		stats:        stats,
		gear:         gear,
		power:        power,
		trainingLoad: trainingLoad,
	}
}

// GetEddingtonNumber returns the current Eddington number for the athlete.
func (r *AchievementRepository) GetEddingtonNumber(ctx context.Context, athleteID int64, sportTypes []string) (int, error) {
	result, err := r.stats.GetEddingtonData(ctx, athleteID, sportTypes)
	if err != nil {
		return 0, err
	}
	if result == nil {
		return 0, nil
	}
	return result.Number, nil
}

// GetGearDistances returns current distances for all gear owned by the athlete.
// Returns map from gear ID to distance in meters.
func (r *AchievementRepository) GetGearDistances(ctx context.Context, athleteID int64) (map[string]float64, error) {
	gear, err := r.gear.GetByAthleteID(ctx, athleteID, false)
	if err != nil {
		return nil, err
	}

	distances := make(map[string]float64, len(gear))
	for _, g := range gear {
		distances[g.ID] = g.Distance
	}
	return distances, nil
}

// BestEffortPRInfo contains PR information for achievement detection.
type BestEffortPRInfo struct {
	ActivityID   int64
	ActivityName string
	DistanceType string
	Name         string
	ElapsedTime  int
}

// GetBestEffortPRsForActivities returns best efforts with pr_rank=1 for the given activity IDs.
//
//nolint:dupl // Similar structure to GetSegmentPRsForActivities but different query and types
func (r *AchievementRepository) GetBestEffortPRsForActivities(ctx context.Context, athleteID int64, activityIDs []int64) ([]BestEffortPRInfo, error) {
	if len(activityIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(activityIDs))
	args := make([]any, 0, len(activityIDs)+1)
	args = append(args, athleteID)
	for i, id := range activityIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `
		SELECT be.activity_id, a.name, be.distance_type, COALESCE(be.name, be.distance_type), be.elapsed_time
		FROM best_efforts be
		JOIN activities a ON a.id = be.activity_id AND a.athlete_id = be.athlete_id
		WHERE be.athlete_id = ?
			AND be.activity_id IN (` + strings.Join(placeholders, ",") + `)
			AND be.pr_rank = 1
		ORDER BY be.activity_id, be.distance_m ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var prs []BestEffortPRInfo
	for rows.Next() {
		var pr BestEffortPRInfo
		if err := rows.Scan(&pr.ActivityID, &pr.ActivityName, &pr.DistanceType, &pr.Name, &pr.ElapsedTime); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

// SegmentPRInfo contains segment PR information for achievement detection.
type SegmentPRInfo struct {
	ActivityID   int64
	ActivityName string
	SegmentID    int64
	SegmentName  string
	ElapsedTime  int
}

// GetSegmentPRsForActivities returns segment efforts with pr_rank=1 for the given activity IDs.
//
//nolint:dupl // Similar structure to GetBestEffortPRsForActivities but different query and types
func (r *AchievementRepository) GetSegmentPRsForActivities(ctx context.Context, athleteID int64, activityIDs []int64) ([]SegmentPRInfo, error) {
	if len(activityIDs) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(activityIDs))
	args := make([]any, 0, len(activityIDs)+1)
	args = append(args, athleteID)
	for i, id := range activityIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := `
		SELECT se.activity_id, a.name, se.segment_id, s.name, se.elapsed_time
		FROM segment_efforts se
		JOIN activities a ON a.id = se.activity_id AND a.athlete_id = se.athlete_id
		JOIN segments s ON s.id = se.segment_id
		WHERE se.athlete_id = ?
			AND se.activity_id IN (` + strings.Join(placeholders, ",") + `)
			AND se.pr_rank = 1
		ORDER BY se.activity_id, se.segment_id
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var prs []SegmentPRInfo
	for rows.Next() {
		var pr SegmentPRInfo
		if err := rows.Scan(&pr.ActivityID, &pr.ActivityName, &pr.SegmentID, &pr.SegmentName, &pr.ElapsedTime); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	return prs, rows.Err()
}

// GetPowerRecords returns all-time best power values for standard durations.
// Returns map from duration (seconds) to best watts.
func (r *AchievementRepository) GetPowerRecords(ctx context.Context, athleteID int64) (map[int]float64, error) {
	query := `
		SELECT duration_s, MAX(best_avg_watts) AS watts
		FROM power_best_efforts
		WHERE athlete_id = ?
		GROUP BY duration_s
	`

	rows, err := r.db.QueryContext(ctx, query, athleteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	records := make(map[int]float64)
	for rows.Next() {
		var duration int
		var watts float64
		if err := rows.Scan(&duration, &watts); err != nil {
			return nil, err
		}
		records[duration] = watts
	}
	return records, rows.Err()
}

// PowerRecordInfo contains power record information for achievement detection.
type PowerRecordInfo struct {
	ActivityID int64
	DurationS  int
	Watts      float64
}

// GetPowerRecordsForActivities returns power records set in the given activity IDs.
func (r *AchievementRepository) GetPowerRecordsForActivities(ctx context.Context, athleteID int64, activityIDs []int64, durations []int) ([]PowerRecordInfo, error) {
	if len(activityIDs) == 0 || len(durations) == 0 {
		return nil, nil
	}

	actPlaceholders := make([]string, len(activityIDs))
	durPlaceholders := make([]string, len(durations))
	args := make([]any, 0, 1+len(activityIDs)+len(durations))
	args = append(args, athleteID)

	for i, id := range activityIDs {
		actPlaceholders[i] = "?"
		args = append(args, id)
	}
	for i, d := range durations {
		durPlaceholders[i] = "?"
		args = append(args, d)
	}

	// Use window function to find records that are the all-time best for each duration
	query := `
		WITH all_time_best AS (
			SELECT duration_s, MAX(best_avg_watts) AS max_watts
			FROM power_best_efforts
			WHERE athlete_id = ?
			GROUP BY duration_s
		)
		SELECT p.activity_id, p.duration_s, p.best_avg_watts
		FROM power_best_efforts p
		JOIN all_time_best ab ON ab.duration_s = p.duration_s AND ab.max_watts = p.best_avg_watts
		WHERE p.activity_id IN (` + strings.Join(actPlaceholders, ",") + `)
			AND p.duration_s IN (` + strings.Join(durPlaceholders, ",") + `)
		ORDER BY p.duration_s
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []PowerRecordInfo
	for rows.Next() {
		var rec PowerRecordInfo
		if err := rows.Scan(&rec.ActivityID, &rec.DurationS, &rec.Watts); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	return records, rows.Err()
}

// TrainingLoadInfo contains training load metrics for achievement detection.
type TrainingLoadInfo struct {
	Day string
	CTL float64
	ATL float64
	TSB float64
}

// GetLatestTrainingLoad returns the most recent training load metrics.
func (r *AchievementRepository) GetLatestTrainingLoad(ctx context.Context, athleteID int64) (*TrainingLoadInfo, error) {
	if r.trainingLoad == nil {
		return nil, nil
	}

	summary, err := r.trainingLoad.GetSummary(ctx, athleteID)
	if err != nil {
		return nil, err
	}
	if summary == nil {
		return nil, nil
	}

	return &TrainingLoadInfo{
		Day: summary.Day,
		CTL: summary.CTL,
		ATL: summary.ATL,
		TSB: summary.TSB,
	}, nil
}
