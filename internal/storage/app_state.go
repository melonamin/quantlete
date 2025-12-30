package storage

import (
	"context"
	"time"
)

// App state keys for well-known configuration values.
const (
	// AppStateDemoMode indicates whether the app is running in demo mode.
	AppStateDemoMode = "demo_mode"
	// AppStateDemoAthleteID stores the athlete ID used in demo mode.
	AppStateDemoAthleteID = "demo_athlete_id"
	// AppStateStravaRateLimit stores the serialized Strava rate limit state.
	AppStateStravaRateLimit = "strava_rate_limit"
)

// AppStateRepository provides access to app-level state storage.
type AppStateRepository struct {
	db *DB
}

// NewAppStateRepository creates a new app state repository.
func NewAppStateRepository(db *DB) *AppStateRepository {
	return &AppStateRepository{db: db}
}

// Get retrieves a value by key. Returns empty string if not found.
func (r *AppStateRepository) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx,
		"SELECT value FROM app_state WHERE key = ?", key).Scan(&value)
	if err != nil {
		if isNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// Set stores a value by key (upsert).
func (r *AppStateRepository) Set(ctx context.Context, key, value string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO app_state (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT (key) DO UPDATE SET
			value = EXCLUDED.value,
			updated_at = EXCLUDED.updated_at
	`, key, value, SQLiteTime{Time: time.Now()})
	return err
}

// Delete removes a key.
func (r *AppStateRepository) Delete(ctx context.Context, key string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM app_state WHERE key = ?", key)
	return err
}
