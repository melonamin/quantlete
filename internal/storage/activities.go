package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/melonamin/quantlete/internal/pagination"
)

// Activity represents a stored activity record.
type Activity struct {
	ID                   int64
	AthleteID            int64
	Name                 string
	Description          string
	SportType            string
	StartDate            SQLiteTime
	StartDateLocal       SQLiteTime
	Timezone             string
	LocationCity         string
	LocationState        string
	LocationCountry      string
	Distance             float64
	MovingTime           int
	ElapsedTime          int
	TotalElevationGain   float64
	ElevHigh             *float64
	ElevLow              *float64
	AverageSpeed         float64
	MaxSpeed             float64
	AverageHeartrate     *float64
	MaxHeartrate         *float64
	AverageWatts         *float64
	MaxWatts             *float64
	WeightedAverageWatts *float64
	Kilojoules           *float64
	AverageCadence       *float64
	Calories             *float64
	KudosCount           int
	CommentCount         int
	PhotoCount           int
	Commute              bool
	Private              bool
	Trainer              bool
	WorkoutType          *int
	DeviceName           string
	GearID               string
	StartLat             *float64
	StartLng             *float64
	EndLat               *float64
	EndLng               *float64
	Polyline             string
	SummaryPolyline      string
	CreatedAt            SQLiteTime
	UpdatedAt            SQLiteTime
}

// ActivityFilters defines query filters for activities.
type ActivityFilters struct {
	AthleteID    int64
	SportTypes   []string
	StartAfter   *time.Time
	StartBefore  *time.Time
	GearID       string
	Commute      *bool
	Trainer      *bool
	Search       string
	MinDistanceM *float64
	MaxDistanceM *float64
	MinDurationS *int
	MaxDurationS *int
}

// Pagination defines pagination parameters.
type Pagination struct {
	Page     int
	PerPage  int
	OrderBy  string
	OrderDir string
}

// Normalize applies default values and enforces constraints on pagination parameters.
// Uses constants from the pagination package for consistency.
func (p *Pagination) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Page > 10000 {
		p.Page = 10000
	}
	if p.PerPage < 1 {
		p.PerPage = 50
	}
	if p.PerPage > 200 {
		p.PerPage = 200
	}
}

// ActivityRepository handles activity persistence.
type ActivityRepository struct {
	db *DB
}

// NewActivityRepository creates a new activity repository.
func NewActivityRepository(db *DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// Upsert inserts or updates an activity.
func (r *ActivityRepository) Upsert(ctx context.Context, a *Activity) error {
	_, err := r.db.Exec(`
		INSERT INTO activities (
			id, athlete_id, name, description, sport_type,
			start_date, start_date_local, timezone,
			location_city, location_state, location_country,
			distance, moving_time, elapsed_time, total_elevation_gain,
			elev_high, elev_low, average_speed, max_speed,
			average_heartrate, max_heartrate,
			average_watts, max_watts, weighted_average_watts, kilojoules,
			average_cadence, calories,
			kudos_count, comment_count, photo_count,
			commute, private, trainer, workout_type,
			device_name, gear_id,
			start_lat, start_lng, end_lat, end_lng,
			polyline, summary_polyline,
			created_at, updated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?, ?, ?,
			?, ?,
			?, ?
		)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			sport_type = EXCLUDED.sport_type,
			start_date = EXCLUDED.start_date,
			start_date_local = EXCLUDED.start_date_local,
			timezone = EXCLUDED.timezone,
			location_city = EXCLUDED.location_city,
			location_state = EXCLUDED.location_state,
			location_country = EXCLUDED.location_country,
			distance = EXCLUDED.distance,
			moving_time = EXCLUDED.moving_time,
			elapsed_time = EXCLUDED.elapsed_time,
			total_elevation_gain = EXCLUDED.total_elevation_gain,
			elev_high = EXCLUDED.elev_high,
			elev_low = EXCLUDED.elev_low,
			average_speed = EXCLUDED.average_speed,
			max_speed = EXCLUDED.max_speed,
			average_heartrate = EXCLUDED.average_heartrate,
			max_heartrate = EXCLUDED.max_heartrate,
			average_watts = EXCLUDED.average_watts,
			max_watts = EXCLUDED.max_watts,
			weighted_average_watts = EXCLUDED.weighted_average_watts,
			kilojoules = EXCLUDED.kilojoules,
			average_cadence = EXCLUDED.average_cadence,
			calories = EXCLUDED.calories,
			kudos_count = EXCLUDED.kudos_count,
			comment_count = EXCLUDED.comment_count,
			photo_count = EXCLUDED.photo_count,
			commute = EXCLUDED.commute,
			private = EXCLUDED.private,
			trainer = EXCLUDED.trainer,
			workout_type = EXCLUDED.workout_type,
			device_name = EXCLUDED.device_name,
			gear_id = EXCLUDED.gear_id,
			start_lat = EXCLUDED.start_lat,
			start_lng = EXCLUDED.start_lng,
			end_lat = EXCLUDED.end_lat,
			end_lng = EXCLUDED.end_lng,
			polyline = EXCLUDED.polyline,
			summary_polyline = EXCLUDED.summary_polyline,
			updated_at = EXCLUDED.updated_at
	`,
		a.ID, a.AthleteID, a.Name, a.Description, a.SportType,
		a.StartDate, a.StartDateLocal, a.Timezone,
		a.LocationCity, a.LocationState, a.LocationCountry,
		a.Distance, a.MovingTime, a.ElapsedTime, a.TotalElevationGain,
		a.ElevHigh, a.ElevLow, a.AverageSpeed, a.MaxSpeed,
		a.AverageHeartrate, a.MaxHeartrate,
		a.AverageWatts, a.MaxWatts, a.WeightedAverageWatts, a.Kilojoules,
		a.AverageCadence, a.Calories,
		a.KudosCount, a.CommentCount, a.PhotoCount,
		a.Commute, a.Private, a.Trainer, a.WorkoutType,
		a.DeviceName, a.GearID,
		a.StartLat, a.StartLng, a.EndLat, a.EndLng,
		a.Polyline, a.SummaryPolyline,
		SQLiteTime{Time: time.Now()}, SQLiteTime{Time: time.Now()},
	)
	return err
}

// GetByID retrieves an activity by ID.
func (r *ActivityRepository) GetByID(ctx context.Context, id int64) (*Activity, error) {
	row := r.db.QueryRow(`
		SELECT
			id, athlete_id, name, description, sport_type,
			start_date, start_date_local, timezone,
			COALESCE(location_city, ''), COALESCE(location_state, ''), COALESCE(location_country, ''),
			distance, moving_time, elapsed_time, total_elevation_gain,
			elev_high, elev_low, average_speed, max_speed,
			average_heartrate, max_heartrate,
			average_watts, max_watts, weighted_average_watts, kilojoules,
			average_cadence, calories,
			kudos_count, comment_count, photo_count,
			commute, private, trainer, workout_type,
			device_name, gear_id,
			start_lat, start_lng, end_lat, end_lng,
			polyline, summary_polyline,
			created_at, updated_at
		FROM activities WHERE id = ?
	`, id)

	a, err := scanActivity(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return a, err
}

// List retrieves activities with filters and pagination.
func (r *ActivityRepository) List(ctx context.Context, filters ActivityFilters, page Pagination) ([]Activity, int, error) {
	// Build WHERE clause
	var conditions []string
	var args []any

	if filters.AthleteID > 0 {
		conditions = append(conditions, "athlete_id = ?")
		args = append(args, filters.AthleteID)
	}

	if len(filters.SportTypes) > 0 {
		placeholders := make([]string, len(filters.SportTypes))
		for i, st := range filters.SportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		conditions = append(conditions, fmt.Sprintf("sport_type IN (%s)", strings.Join(placeholders, ",")))
	}

	conditions, args = AddTimeRangeFilter(conditions, args, "start_date", filters.StartAfter, filters.StartBefore)

	if filters.GearID != "" {
		conditions = append(conditions, "gear_id = ?")
		args = append(args, filters.GearID)
	}

	if filters.Commute != nil {
		conditions = append(conditions, "commute = ?")
		args = append(args, *filters.Commute)
	}

	if filters.Trainer != nil {
		conditions = append(conditions, "trainer = ?")
		args = append(args, *filters.Trainer)
	}

	if filters.Search != "" {
		conditions = append(conditions, "name LIKE ? COLLATE NOCASE")
		args = append(args, "%"+filters.Search+"%")
	}

	// Distance filters (distance is stored in meters)
	if filters.MinDistanceM != nil {
		conditions = append(conditions, "distance >= ?")
		args = append(args, *filters.MinDistanceM)
	}
	if filters.MaxDistanceM != nil {
		conditions = append(conditions, "distance <= ?")
		args = append(args, *filters.MaxDistanceM)
	}

	// Duration filters (moving_time is stored in seconds)
	if filters.MinDurationS != nil {
		conditions = append(conditions, "moving_time >= ?")
		args = append(args, *filters.MinDurationS)
	}
	if filters.MaxDurationS != nil {
		conditions = append(conditions, "moving_time <= ?")
		args = append(args, *filters.MaxDurationS)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM activities %s", where)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting activities: %w", err)
	}

	// Build ORDER BY with validation
	orderBy := pagination.BuildOrderClause(page.OrderBy, page.OrderDir, allowedActivityOrderColumns, "start_date DESC")

	// Normalize pagination params
	p := pagination.NewParams(page.Page, page.PerPage)

	// Query activities
	query := fmt.Sprintf(`
		SELECT
			id, athlete_id, name, description, sport_type,
			start_date, start_date_local, timezone,
			COALESCE(location_city, ''), COALESCE(location_state, ''), COALESCE(location_country, ''),
			distance, moving_time, elapsed_time, total_elevation_gain,
			elev_high, elev_low, average_speed, max_speed,
			average_heartrate, max_heartrate,
			average_watts, max_watts, weighted_average_watts, kilojoules,
			average_cadence, calories,
			kudos_count, comment_count, photo_count,
			commute, private, trainer, workout_type,
			device_name, gear_id,
			start_lat, start_lng, end_lat, end_lng,
			polyline, summary_polyline,
			created_at, updated_at
		FROM activities %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, where, orderBy)

	args = append(args, p.PerPage, p.Offset())

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying activities: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var activities []Activity
	for rows.Next() {
		a, err := scanActivityRow(rows)
		if err != nil {
			return nil, 0, err
		}
		activities = append(activities, *a)
	}

	return activities, total, rows.Err()
}

// DeleteByID deletes an activity and its dependent computed rows.
//
// Most child tables use ON DELETE CASCADE, but a few computed tables do not.
func (r *ActivityRepository) DeleteByID(ctx context.Context, athleteID, activityID int64) error {
	tx, err := r.db.Conn().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// These tables reference activities(id) without ON DELETE CASCADE.
	if _, err := tx.ExecContext(ctx, "DELETE FROM power_best_efforts WHERE activity_id = ? AND athlete_id = ?", activityID, athleteID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM activity_training_load WHERE activity_id = ? AND athlete_id = ?", activityID, athleteID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM activities WHERE id = ? AND athlete_id = ?", activityID, athleteID); err != nil {
		return err
	}

	return tx.Commit()
}

// GetTotals returns aggregate statistics for activities.
func (r *ActivityRepository) GetTotals(ctx context.Context, filters ActivityFilters) (*ActivityTotals, error) {
	// Build WHERE clause (same as List)
	var conditions []string
	var args []any

	if filters.AthleteID > 0 {
		conditions = append(conditions, "athlete_id = ?")
		args = append(args, filters.AthleteID)
	}

	if len(filters.SportTypes) > 0 {
		placeholders := make([]string, len(filters.SportTypes))
		for i, st := range filters.SportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		conditions = append(conditions, fmt.Sprintf("sport_type IN (%s)", strings.Join(placeholders, ",")))
	}

	conditions, args = AddTimeRangeFilter(conditions, args, "start_date", filters.StartAfter, filters.StartBefore)

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT
			COUNT(*) as count,
			COALESCE(SUM(distance), 0) as distance,
			COALESCE(SUM(moving_time), 0) as moving_time,
			COALESCE(SUM(elapsed_time), 0) as elapsed_time,
			COALESCE(SUM(total_elevation_gain), 0) as elevation,
			COALESCE(SUM(calories), 0) as calories,
			COALESCE(SUM(kilojoules), 0) as kilojoules
		FROM activities %s
	`, where)

	var t ActivityTotals
	err := r.db.QueryRow(query, args...).Scan(
		&t.Count, &t.Distance, &t.MovingTime, &t.ElapsedTime,
		&t.Elevation, &t.Calories, &t.Kilojoules,
	)
	if err != nil {
		return nil, fmt.Errorf("getting totals: %w", err)
	}

	return &t, nil
}

// ActivityTotals represents aggregate statistics.
type ActivityTotals struct {
	Count       int     `json:"count"`
	Distance    float64 `json:"distance"`     // meters
	MovingTime  int     `json:"moving_time"`  // seconds
	ElapsedTime int     `json:"elapsed_time"` // seconds
	Elevation   float64 `json:"elevation"`    // meters
	Calories    float64 `json:"calories"`
	Kilojoules  float64 `json:"kilojoules"`
}

// GetMostRecent returns the most recent activities.
func (r *ActivityRepository) GetMostRecent(ctx context.Context, athleteID int64, limit int) ([]Activity, error) {
	rows, err := r.db.Query(`
		SELECT
			id, athlete_id, name, description, sport_type,
			start_date, start_date_local, timezone,
			COALESCE(location_city, ''), COALESCE(location_state, ''), COALESCE(location_country, ''),
			distance, moving_time, elapsed_time, total_elevation_gain,
			elev_high, elev_low, average_speed, max_speed,
			average_heartrate, max_heartrate,
			average_watts, max_watts, weighted_average_watts, kilojoules,
			average_cadence, calories,
			kudos_count, comment_count, photo_count,
			commute, private, trainer, workout_type,
			device_name, gear_id,
			start_lat, start_lng, end_lat, end_lng,
			polyline, summary_polyline,
			created_at, updated_at
		FROM activities
		WHERE athlete_id = ?
		ORDER BY start_date DESC
		LIMIT ?
	`, athleteID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var activities []Activity
	for rows.Next() {
		a, err := scanActivityRow(rows)
		if err != nil {
			return nil, err
		}
		activities = append(activities, *a)
	}

	return activities, rows.Err()
}

// GetLatestActivityDate returns the start date of the most recent activity.
func (r *ActivityRepository) GetLatestActivityDate(ctx context.Context, athleteID int64) (*time.Time, error) {
	var startDate SQLiteTime
	err := r.db.QueryRow(`
		SELECT start_date FROM activities
		WHERE athlete_id = ?
		ORDER BY start_date DESC
		LIMIT 1
	`, athleteID).Scan(&startDate)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if startDate.IsZero() {
		return nil, nil
	}
	t := startDate.Time
	return &t, nil
}

// scanner interface for sql.Row and sql.Rows
type scanner interface {
	Scan(dest ...any) error
}

func scanActivity(s scanner) (*Activity, error) {
	var a Activity
	err := s.Scan(
		&a.ID, &a.AthleteID, &a.Name, &a.Description, &a.SportType,
		&a.StartDate, &a.StartDateLocal, &a.Timezone,
		&a.LocationCity, &a.LocationState, &a.LocationCountry,
		&a.Distance, &a.MovingTime, &a.ElapsedTime, &a.TotalElevationGain,
		&a.ElevHigh, &a.ElevLow, &a.AverageSpeed, &a.MaxSpeed,
		&a.AverageHeartrate, &a.MaxHeartrate,
		&a.AverageWatts, &a.MaxWatts, &a.WeightedAverageWatts, &a.Kilojoules,
		&a.AverageCadence, &a.Calories,
		&a.KudosCount, &a.CommentCount, &a.PhotoCount,
		&a.Commute, &a.Private, &a.Trainer, &a.WorkoutType,
		&a.DeviceName, &a.GearID,
		&a.StartLat, &a.StartLng, &a.EndLat, &a.EndLng,
		&a.Polyline, &a.SummaryPolyline,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scanning activity: %w", err)
	}
	return &a, nil
}

func scanActivityRow(rows *sql.Rows) (*Activity, error) {
	return scanActivity(rows)
}

var allowedActivityOrderColumns = map[string]string{
	"start_date":           "start_date",
	"distance":             "distance",
	"moving_time":          "moving_time",
	"elapsed_time":         "elapsed_time",
	"total_elevation_gain": "total_elevation_gain",
	"sport_type":           "sport_type",
	"name":                 "name",
	"average_speed":        "average_speed",
	"average_heartrate":    "average_heartrate",
	"average_watts":        "average_watts",
	"average_cadence":      "average_cadence",
	"calories":             "calories",
	"created_at":           "created_at",
	"updated_at":           "updated_at",
}

// StreamRepository handles activity stream persistence.
type StreamRepository struct {
	db *DB
}

// NewStreamRepository creates a new stream repository.
func NewStreamRepository(db *DB) *StreamRepository {
	return &StreamRepository{db: db}
}

// ActivityStream represents a stored stream.
type ActivityStream struct {
	ActivityID   int64
	StreamType   string
	OriginalSize int
	Resolution   string
	SeriesType   string
	Data         json.RawMessage
	CreatedAt    SQLiteTime
}

// Upsert inserts or updates a stream.
func (r *StreamRepository) Upsert(ctx context.Context, s *ActivityStream) error {
	_, err := r.db.Exec(`
		INSERT INTO activity_streams (
			activity_id, stream_type, original_size, resolution, series_type, data, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (activity_id, stream_type) DO UPDATE SET
			original_size = EXCLUDED.original_size,
			resolution = EXCLUDED.resolution,
			series_type = EXCLUDED.series_type,
			data = EXCLUDED.data
	`,
		s.ActivityID, s.StreamType, s.OriginalSize, s.Resolution, s.SeriesType, s.Data, SQLiteTime{Time: time.Now()},
	)
	return err
}

// GetByActivityID retrieves all streams for an activity.
func (r *StreamRepository) GetByActivityID(ctx context.Context, activityID int64) ([]ActivityStream, error) {
	rows, err := r.db.Query(`
		SELECT activity_id, stream_type, original_size, resolution, series_type, data, created_at
		FROM activity_streams
		WHERE activity_id = ?
	`, activityID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var streams []ActivityStream
	for rows.Next() {
		var s ActivityStream
		if err := rows.Scan(&s.ActivityID, &s.StreamType, &s.OriginalSize, &s.Resolution, &s.SeriesType, &s.Data, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning stream: %w", err)
		}
		streams = append(streams, s)
	}

	return streams, rows.Err()
}

// DeleteByActivityID deletes all streams for an activity.
func (r *StreamRepository) DeleteByActivityID(ctx context.Context, activityID int64) error {
	_, err := r.db.Exec("DELETE FROM activity_streams WHERE activity_id = ?", activityID)
	return err
}
