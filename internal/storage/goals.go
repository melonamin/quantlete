package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type GoalMetric string
type GoalPeriod string

const (
	GoalMetricDistance   GoalMetric = "distance"
	GoalMetricElevation  GoalMetric = "elevation"
	GoalMetricMovingTime GoalMetric = "moving_time"

	GoalPeriodWeek     GoalPeriod = "week"
	GoalPeriodMonth    GoalPeriod = "month"
	GoalPeriodYear     GoalPeriod = "year"
	GoalPeriodLifetime GoalPeriod = "lifetime"
)

type GoalsMetricTargets struct {
	DistanceM   *float64 `json:"distance_m,omitempty"`
	ElevationM  *float64 `json:"elevation_m,omitempty"`
	MovingTimeS *int     `json:"moving_time_s,omitempty"`
}

type GoalsSportConfig struct {
	Name       string                            `json:"name"`
	SportTypes []string                          `json:"sport_types"`
	Targets    map[GoalPeriod]GoalsMetricTargets `json:"targets"`
}

type TrainingGoalsConfig struct {
	Version int                `json:"version"`
	Sports  []GoalsSportConfig `json:"sports"`
}

func DefaultTrainingGoalsConfig() TrainingGoalsConfig {
	return TrainingGoalsConfig{
		Version: 1,
		Sports: []GoalsSportConfig{
			{
				Name:       "Rides",
				SportTypes: []string{"Ride", "VirtualRide", "GravelRide", "MountainBikeRide"},
				Targets:    map[GoalPeriod]GoalsMetricTargets{},
			},
			{
				Name:       "Runs",
				SportTypes: []string{"Run", "VirtualRun", "TrailRun"},
				Targets:    map[GoalPeriod]GoalsMetricTargets{},
			},
		},
	}
}

type GoalsProgress struct {
	DistanceM     float64 `json:"distance_m"`
	ElevationM    float64 `json:"elevation_m"`
	MovingTimeS   int     `json:"moving_time_s"`
	ActivityCount int     `json:"activity_count"`
}

type TrainingGoalsResponse struct {
	Config   TrainingGoalsConfig                     `json:"config"`
	Progress map[string]map[GoalPeriod]GoalsProgress `json:"progress"` // sport name -> period -> progress
}

type GoalsRepository struct {
	db *DB
}

func NewGoalsRepository(db *DB) *GoalsRepository {
	return &GoalsRepository{db: db}
}

func (r *GoalsRepository) GetConfig(ctx context.Context, athleteID int64) (*TrainingGoalsConfig, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx, "SELECT config FROM training_goals WHERE athlete_id = ?", athleteID).Scan(&raw)
	if err != nil {
		if isNotFound(err) {
			cfg := DefaultTrainingGoalsConfig()
			return &cfg, nil
		}
		return nil, err
	}

	var cfg TrainingGoalsConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("decoding training goals: %w", err)
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	return &cfg, nil
}

func (r *GoalsRepository) UpsertConfig(ctx context.Context, athleteID int64, cfg TrainingGoalsConfig) error {
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding training goals: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO training_goals (athlete_id, config, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT (athlete_id) DO UPDATE SET
			config = EXCLUDED.config,
			updated_at = EXCLUDED.updated_at
	`, athleteID, b, time.Now())
	return err
}

func (r *GoalsRepository) GetProgress(ctx context.Context, athleteID int64, sportTypes []string, period GoalPeriod) (GoalsProgress, error) {
	query := `
		SELECT
			COUNT(*) as activity_count,
			COALESCE(SUM(distance), 0) as distance_m,
			COALESCE(SUM(total_elevation_gain), 0) as elevation_m,
			COALESCE(SUM(moving_time), 0) as moving_time_s
		FROM activities
		WHERE athlete_id = ?
	`
	args := []any{athleteID}

	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + joinStrings(placeholders, ",") + ")"
	}

	if start := periodStart(time.Now(), period); start != nil {
		query += " AND start_date >= ?"
		args = append(args, *start)
	}

	var p GoalsProgress
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&p.ActivityCount, &p.DistanceM, &p.ElevationM, &p.MovingTimeS); err != nil {
		return GoalsProgress{}, err
	}
	return p, nil
}

func periodStart(now time.Time, period GoalPeriod) *time.Time {
	switch period {
	case GoalPeriodWeek:
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		return &start
	case GoalPeriodMonth:
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return &start
	case GoalPeriodYear:
		start := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		return &start
	case GoalPeriodLifetime:
		return nil
	default:
		return nil
	}
}
