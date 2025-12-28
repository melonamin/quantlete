package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type VirtualWorldTileLayer struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Attribution string `json:"attribution,omitempty"`
	MaxZoom     int    `json:"max_zoom,omitempty"`
}

type EddingtonDefinition struct {
	ID                    string   `json:"id"`
	Name                  string   `json:"name"`
	SportTypes            []string `json:"sport_types,omitempty"`
	ShowInNav             bool     `json:"show_in_nav,omitempty"`
	ShowInDashboardWidget bool     `json:"show_in_dashboard_widget,omitempty"`
}

type AthleteSettings struct {
	Version                int                              `json:"version"`
	VirtualWorldTileLayers map[string]VirtualWorldTileLayer `json:"virtual_world_tile_layers"`
	EddingtonDefinitions   []EddingtonDefinition            `json:"eddington_definitions,omitempty"`
}

func DefaultAthleteSettings() AthleteSettings {
	return AthleteSettings{
		Version:                2,
		VirtualWorldTileLayers: map[string]VirtualWorldTileLayer{},
		EddingtonDefinitions: []EddingtonDefinition{
			{
				ID:                    "all",
				Name:                  "All activities",
				SportTypes:            nil,
				ShowInNav:             true,
				ShowInDashboardWidget: true,
			},
			{
				ID:                    "rides",
				Name:                  "Rides",
				SportTypes:            []string{"Ride", "MountainBikeRide", "GravelRide", "EBikeRide", "VirtualRide"},
				ShowInNav:             true,
				ShowInDashboardWidget: true,
			},
			{
				ID:                    "runs",
				Name:                  "Runs",
				SportTypes:            []string{"Run", "TrailRun", "VirtualRun"},
				ShowInNav:             true,
				ShowInDashboardWidget: true,
			},
		},
	}
}

type SettingsRepository struct {
	db *DB
}

func NewSettingsRepository(db *DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(ctx context.Context, athleteID int64) (*AthleteSettings, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx, "SELECT settings FROM athlete_settings WHERE athlete_id = ?", athleteID).Scan(&raw)
	if err != nil {
		if isNotFound(err) {
			s := DefaultAthleteSettings()
			return &s, nil
		}
		return nil, err
	}
	var s AthleteSettings
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("decoding athlete settings: %w", err)
	}
	applyAthleteSettingsDefaults(&s)
	return &s, nil
}

func (r *SettingsRepository) Upsert(ctx context.Context, athleteID int64, s AthleteSettings) error {
	applyAthleteSettingsDefaults(&s)
	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("encoding athlete settings: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO athlete_settings (athlete_id, settings, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT (athlete_id) DO UPDATE SET
			settings = EXCLUDED.settings,
			updated_at = EXCLUDED.updated_at
	`, athleteID, b, SQLiteTime{Time: time.Now()})
	return err
}

func applyAthleteSettingsDefaults(s *AthleteSettings) {
	if s.Version == 0 {
		s.Version = 2
	}
	if s.VirtualWorldTileLayers == nil {
		s.VirtualWorldTileLayers = map[string]VirtualWorldTileLayer{}
	}
	if len(s.EddingtonDefinitions) == 0 {
		def := DefaultAthleteSettings()
		s.EddingtonDefinitions = def.EddingtonDefinitions
	}
}
