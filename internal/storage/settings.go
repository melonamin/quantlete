package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/melonamin/quantlete/internal/notifications"
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

type PullSchedule string

const (
	PullScheduleMidnight    PullSchedule = "midnight"
	PullScheduleHourly      PullSchedule = "hourly"
	PullScheduleEvery6Hours PullSchedule = "every_6_hours"
)

// Settings version constants.
const (
	// AthleteSettingsVersion is the current schema version for AthleteSettings.
	// Increment when adding new fields or changing structure.
	AthleteSettingsVersion = 5

	// SchedulerSettingsVersion is the current schema version for SchedulerSettings.
	SchedulerSettingsVersion = 2
)

type PullSettings struct {
	Enabled  bool         `json:"enabled"`
	Schedule PullSchedule `json:"schedule"`
}

type PushSettings struct {
	Enabled bool `json:"enabled"`
}

type SchedulerSettings struct {
	Version int          `json:"version"`
	Pull    PullSettings `json:"pull"`
	Push    PushSettings `json:"push"`
}

type AthleteSettings struct {
	Version                int                               `json:"version"`
	VirtualWorldTileLayers map[string]VirtualWorldTileLayer  `json:"virtual_world_tile_layers"`
	EddingtonDefinitions   []EddingtonDefinition             `json:"eddington_definitions,omitempty"`
	Scheduler              SchedulerSettings                 `json:"scheduler"`
	EnablePublicBadges     bool                              `json:"enable_public_badges"`
	Notifications          *notifications.NotificationConfig `json:"notifications,omitempty"`
}

func DefaultAthleteSettings() AthleteSettings {
	defaultNotifications := notifications.DefaultNotificationConfig()
	return AthleteSettings{
		Version:                AthleteSettingsVersion,
		VirtualWorldTileLayers: map[string]VirtualWorldTileLayer{},
		EddingtonDefinitions:   defaultEddingtonDefinitions(),
		Scheduler:              defaultSchedulerSettings(),
		Notifications:          &defaultNotifications,
	}
}

func defaultEddingtonDefinitions() []EddingtonDefinition {
	return []EddingtonDefinition{
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
	}
}

func defaultSchedulerSettings() SchedulerSettings {
	return SchedulerSettings{
		Version: SchedulerSettingsVersion,
		Pull: PullSettings{
			Enabled:  false,
			Schedule: PullScheduleMidnight,
		},
		Push: PushSettings{
			Enabled: false,
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
	migrateAthleteSettings(&s)
	return &s, nil
}

func (r *SettingsRepository) Upsert(ctx context.Context, athleteID int64, s AthleteSettings) error {
	migrateAthleteSettings(&s)
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

// migrateAthleteSettings applies incremental migrations to bring settings to
// the current version. Each migration step handles one version bump.
func migrateAthleteSettings(s *AthleteSettings) {
	// Fast path: skip migration if already at current version and map is initialized.
	if s.Version >= AthleteSettingsVersion && s.Scheduler.Version >= SchedulerSettingsVersion && s.VirtualWorldTileLayers != nil {
		return
	}

	// Ensure map is initialized (required for all versions).
	if s.VirtualWorldTileLayers == nil {
		s.VirtualWorldTileLayers = map[string]VirtualWorldTileLayer{}
	}

	// v0 -> v1: Initial settings with eddington definitions.
	if s.Version < 1 {
		if len(s.EddingtonDefinitions) == 0 {
			s.EddingtonDefinitions = defaultEddingtonDefinitions()
		}
		s.Version = 1
	}

	// v1 -> v2: Added scheduler settings.
	if s.Version < 2 {
		if s.Scheduler.Version == 0 {
			s.Scheduler = defaultSchedulerSettings()
		}
		s.Version = 2
	}

	// v2 -> v3: Added public badges setting.
	if s.Version < 3 {
		// EnablePublicBadges defaults to false (zero value), no action needed.
		s.Version = 3
	}

	// v3 -> v4: Added notification settings.
	if s.Version < 4 {
		if s.Notifications == nil {
			defaultNotifications := notifications.DefaultNotificationConfig()
			s.Notifications = &defaultNotifications
		}
		s.Version = 4
	}

	// v4 -> v5: Added new notification event types.
	// New fields: PersonalRecords, SegmentPRs, EddingtonIncrease, PowerRecords,
	// GoalComplete, FatigueWarning, RecoveryAlert, OvertrainingRisk,
	// WeeklyDigest, MonthlyDigest, GearMilestones.
	// All default to false (zero value), no action needed beyond version bump.
	if s.Version < 5 {
		s.Version = 5
	}

	// Always normalize scheduler settings.
	migrateSchedulerSettings(&s.Scheduler)
}

// migrateSchedulerSettings applies migrations to scheduler settings.
func migrateSchedulerSettings(s *SchedulerSettings) {
	// v0 -> v1: Initial scheduler with pull settings.
	if s.Version < 1 {
		if s.Pull.Schedule == "" {
			s.Pull.Schedule = PullScheduleMidnight
		}
		s.Version = 1
	}

	// v1 -> v2: Added push settings.
	if s.Version < 2 {
		// Push.Enabled defaults to false (zero value), no action needed.
		s.Version = 2
	}

	// Normalize pull schedule to valid values.
	s.Pull.Schedule = normalizePullSchedule(s.Pull.Schedule, PullScheduleMidnight)
}

func normalizePullSchedule(v, def PullSchedule) PullSchedule {
	switch v {
	case PullScheduleMidnight, PullScheduleHourly, PullScheduleEvery6Hours:
		return v
	default:
		return def
	}
}

// GetPublicBadgesAthleteID returns the athlete ID if public badges are enabled,
// or 0 if no athlete has public badges enabled.
// In single-user mode, this is the only authenticated athlete.
func (r *SettingsRepository) GetPublicBadgesAthleteID(ctx context.Context) (int64, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT athlete_id, settings FROM athlete_settings")
	if err != nil {
		return 0, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var athleteID int64
		var raw []byte
		if err := rows.Scan(&athleteID, &raw); err != nil {
			return 0, err
		}
		var s AthleteSettings
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		if s.EnablePublicBadges {
			return athleteID, nil
		}
	}
	return 0, nil
}
