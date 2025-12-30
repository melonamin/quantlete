package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/sasha/stata/internal/weather"
)

// WeatherRepository handles activity weather data storage.
type WeatherRepository struct {
	db *sql.DB
}

// NewWeatherRepository creates a new weather repository.
func NewWeatherRepository(db *sql.DB) *WeatherRepository {
	return &WeatherRepository{db: db}
}

// GetByActivityID retrieves cached weather data for an activity.
func (r *WeatherRepository) GetByActivityID(ctx context.Context, activityID int64) (*weather.ActivityWeather, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			activity_id,
			source,
			temperature_c,
			feels_like_c,
			humidity_percent,
			wind_speed_mps,
			wind_direction_deg,
			precipitation_mm,
			weather_code,
			temp_min_c,
			temp_max_c,
			temp_avg_c,
			temp_stream,
			fetched_at
		FROM activity_weather
		WHERE activity_id = ?
	`, activityID)

	var w weather.ActivityWeather
	var tempStream sql.NullString
	var fetchedAt string

	err := row.Scan(
		&w.ActivityID,
		&w.Source,
		&w.TemperatureC,
		&w.FeelsLikeC,
		&w.HumidityPct,
		&w.WindSpeedMps,
		&w.WindDirDeg,
		&w.PrecipMM,
		&w.WeatherCode,
		&w.TempMinC,
		&w.TempMaxC,
		&w.TempAvgC,
		&tempStream,
		&fetchedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Parse temp stream JSON
	if tempStream.Valid && tempStream.String != "" {
		if err := json.Unmarshal([]byte(tempStream.String), &w.TempStream); err != nil {
			// Ignore parse errors, just leave TempStream empty
			w.TempStream = nil
		}
	}

	// Parse fetched_at
	if t, err := time.Parse(time.RFC3339, fetchedAt); err == nil {
		w.FetchedAt = t
	} else if t, err := time.Parse("2006-01-02 15:04:05", fetchedAt); err == nil {
		w.FetchedAt = t
	}

	return &w, nil
}

// Upsert inserts or updates weather data for an activity.
func (r *WeatherRepository) Upsert(ctx context.Context, w *weather.ActivityWeather) error {
	var tempStreamJSON sql.NullString
	if len(w.TempStream) > 0 {
		data, err := json.Marshal(w.TempStream)
		if err != nil {
			return err
		}
		tempStreamJSON = sql.NullString{String: string(data), Valid: true}
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO activity_weather (
			activity_id,
			source,
			temperature_c,
			feels_like_c,
			humidity_percent,
			wind_speed_mps,
			wind_direction_deg,
			precipitation_mm,
			weather_code,
			temp_min_c,
			temp_max_c,
			temp_avg_c,
			temp_stream,
			fetched_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(activity_id) DO UPDATE SET
			source = excluded.source,
			temperature_c = excluded.temperature_c,
			feels_like_c = excluded.feels_like_c,
			humidity_percent = excluded.humidity_percent,
			wind_speed_mps = excluded.wind_speed_mps,
			wind_direction_deg = excluded.wind_direction_deg,
			precipitation_mm = excluded.precipitation_mm,
			weather_code = excluded.weather_code,
			temp_min_c = excluded.temp_min_c,
			temp_max_c = excluded.temp_max_c,
			temp_avg_c = excluded.temp_avg_c,
			temp_stream = excluded.temp_stream,
			fetched_at = excluded.fetched_at
	`,
		w.ActivityID,
		w.Source,
		w.TemperatureC,
		w.FeelsLikeC,
		w.HumidityPct,
		w.WindSpeedMps,
		w.WindDirDeg,
		w.PrecipMM,
		w.WeatherCode,
		w.TempMinC,
		w.TempMaxC,
		w.TempAvgC,
		tempStreamJSON,
		w.FetchedAt.Format(time.RFC3339),
	)
	return err
}

// Delete removes weather data for an activity.
func (r *WeatherRepository) Delete(ctx context.Context, activityID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM activity_weather WHERE activity_id = ?`, activityID)
	return err
}
