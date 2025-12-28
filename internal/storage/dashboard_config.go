package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type WidgetWidth int

const (
	WidgetWidthOneThird WidgetWidth = 4
	WidgetWidthHalf     WidgetWidth = 6
	WidgetWidthTwoThird WidgetWidth = 8
	WidgetWidthFull     WidgetWidth = 12
)

type DashboardWidgetConfig struct {
	ID       string         `json:"id"`
	Width    WidgetWidth    `json:"width"`
	Hidden   bool           `json:"hidden"`
	Settings map[string]any `json:"settings,omitempty"`
}

type DashboardConfig struct {
	Version int                     `json:"version"`
	Widgets []DashboardWidgetConfig `json:"widgets"`
}

func DefaultDashboardConfig() DashboardConfig {
	return DashboardConfig{
		Version: 1,
		Widgets: []DashboardWidgetConfig{
			{ID: "recent_activities", Width: WidgetWidthTwoThird},
			{ID: "weekly_stats", Width: WidgetWidthOneThird},
			{ID: "monthly_chart", Width: WidgetWidthTwoThird},
			{ID: "sport_breakdown", Width: WidgetWidthOneThird},
			{ID: "sport_chart", Width: WidgetWidthOneThird},
			{ID: "training_goals", Width: WidgetWidthOneThird},
			{ID: "activity_calendar", Width: WidgetWidthFull},
			// Phase 8 widgets (default hidden)
			{ID: "recent_challenges", Width: WidgetWidthOneThird, Hidden: true},
			{ID: "challenge_consistency", Width: WidgetWidthTwoThird, Hidden: true},
			// Phase 9 widgets (default hidden)
			{ID: "eddington", Width: WidgetWidthOneThird, Hidden: true},
			// Phase 6 widgets (default hidden)
			{ID: "peak_power_outputs", Width: WidgetWidthOneThird, Hidden: true},
			{ID: "heart_rate_zones", Width: WidgetWidthOneThird, Hidden: true},
			{ID: "training_load", Width: WidgetWidthTwoThird, Hidden: true},
		},
	}
}

type DashboardConfigRepository struct {
	db *DB
}

func NewDashboardConfigRepository(db *DB) *DashboardConfigRepository {
	return &DashboardConfigRepository{db: db}
}

func (r *DashboardConfigRepository) Get(ctx context.Context, athleteID int64) (*DashboardConfig, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx, "SELECT config FROM dashboard_config WHERE athlete_id = ?", athleteID).Scan(&raw)
	if err != nil {
		if isNotFound(err) {
			cfg := DefaultDashboardConfig()
			return &cfg, nil
		}
		return nil, err
	}

	var cfg DashboardConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("decoding dashboard config: %w", err)
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	return &cfg, nil
}

func (r *DashboardConfigRepository) Upsert(ctx context.Context, athleteID int64, cfg DashboardConfig) error {
	if cfg.Version == 0 {
		cfg.Version = 1
	}
	b, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encoding dashboard config: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO dashboard_config (athlete_id, config, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT (athlete_id) DO UPDATE SET
			config = EXCLUDED.config,
			updated_at = EXCLUDED.updated_at
	`, athleteID, b, time.Now())
	return err
}
