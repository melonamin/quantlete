package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Segment struct {
	ID                   int64
	Name                 string
	ActivityType         string
	Distance             float64
	AverageGrade         float64
	MaximumGrade         float64
	ElevationHigh        float64
	ElevationLow         float64
	ClimbCategory        int
	StartLat             *float64
	StartLng             *float64
	EndLat               *float64
	EndLng               *float64
	Starred              bool
	Polyline             string
	AthleteKOMRank       *int
	AthleteEffortCount   *int
	AthletePRElapsedTime *int
	AthletePRDate        *SQLiteTime
	CreatedAt            SQLiteTime
	UpdatedAt            SQLiteTime
}

type SegmentEffort struct {
	ID               int64
	SegmentID        int64
	ActivityID       int64
	AthleteID        int64
	Name             string
	ElapsedTime      int
	MovingTime       int
	StartDate        *SQLiteTime
	StartDateLocal   *SQLiteTime
	Distance         float64
	AverageWatts     *float64
	AverageHeartrate *float64
	MaxHeartrate     *int
	PRRank           *int
	Country          string
	CreatedAt        SQLiteTime
}

type SegmentRepository struct {
	db *DB
}

func NewSegmentRepository(db *DB) *SegmentRepository {
	return &SegmentRepository{db: db}
}

func (r *SegmentRepository) UpsertSegment(ctx context.Context, s *Segment) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO segments (
			id, name, activity_type, distance, average_grade, maximum_grade,
			elevation_high, elevation_low, climb_category,
			start_lat, start_lng, end_lat, end_lng,
			starred, polyline,
			athlete_kom_rank, athlete_effort_count, athlete_pr_elapsed_time, athlete_pr_date,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			activity_type = EXCLUDED.activity_type,
			distance = EXCLUDED.distance,
			average_grade = EXCLUDED.average_grade,
			maximum_grade = EXCLUDED.maximum_grade,
			elevation_high = EXCLUDED.elevation_high,
			elevation_low = EXCLUDED.elevation_low,
			climb_category = EXCLUDED.climb_category,
			start_lat = EXCLUDED.start_lat,
			start_lng = EXCLUDED.start_lng,
			end_lat = EXCLUDED.end_lat,
			end_lng = EXCLUDED.end_lng,
			starred = EXCLUDED.starred,
			polyline = EXCLUDED.polyline,
			athlete_kom_rank = EXCLUDED.athlete_kom_rank,
			athlete_effort_count = EXCLUDED.athlete_effort_count,
			athlete_pr_elapsed_time = EXCLUDED.athlete_pr_elapsed_time,
			athlete_pr_date = EXCLUDED.athlete_pr_date,
			updated_at = EXCLUDED.updated_at
	`, s.ID, s.Name, s.ActivityType, s.Distance, s.AverageGrade, s.MaximumGrade,
		s.ElevationHigh, s.ElevationLow, s.ClimbCategory,
		s.StartLat, s.StartLng, s.EndLat, s.EndLng,
		s.Starred, s.Polyline,
		s.AthleteKOMRank, s.AthleteEffortCount, s.AthletePRElapsedTime, s.AthletePRDate,
		SQLiteTime{Time: time.Now()},
	)
	return err
}

func (r *SegmentRepository) UpsertEffort(ctx context.Context, e *SegmentEffort) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO segment_efforts (
			id, segment_id, activity_id, athlete_id,
			name, elapsed_time, moving_time,
			start_date, start_date_local,
			distance, average_watts, average_heartrate, max_heartrate,
			pr_rank, country
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			elapsed_time = EXCLUDED.elapsed_time,
			moving_time = EXCLUDED.moving_time,
			start_date = EXCLUDED.start_date,
			start_date_local = EXCLUDED.start_date_local,
			distance = EXCLUDED.distance,
			average_watts = EXCLUDED.average_watts,
			average_heartrate = EXCLUDED.average_heartrate,
			max_heartrate = EXCLUDED.max_heartrate,
			pr_rank = EXCLUDED.pr_rank,
			country = EXCLUDED.country
	`, e.ID, e.SegmentID, e.ActivityID, e.AthleteID,
		e.Name, e.ElapsedTime, e.MovingTime,
		e.StartDate, e.StartDateLocal,
		e.Distance, e.AverageWatts, e.AverageHeartrate, e.MaxHeartrate,
		e.PRRank, e.Country,
	)
	return err
}

type SegmentListItem struct {
	Segment
	TimesCompleted  int         `json:"times_completed"`
	LastEffortDate  *SQLiteTime `json:"last_effort_date,omitempty"`
	BestElapsedTime *int        `json:"best_elapsed_time,omitempty"`
}

type SegmentFilters struct {
	ActivityType string
	Country      string
	Starred      *bool
	KOMOnly      bool
	Search       string
}

func (r *SegmentRepository) List(ctx context.Context, athleteID int64, f SegmentFilters, limit int) ([]SegmentListItem, error) {
	if limit <= 0 || limit > 5000 {
		limit = 500
	}

	query := `
		SELECT
			s.id, s.name, s.activity_type, s.distance, s.average_grade, s.maximum_grade,
			s.elevation_high, s.elevation_low, s.climb_category,
			s.start_lat, s.start_lng, s.end_lat, s.end_lng,
			s.starred, COALESCE(s.polyline, ''),
			s.athlete_kom_rank, s.athlete_effort_count, s.athlete_pr_elapsed_time, s.athlete_pr_date,
			s.created_at, s.updated_at,
			COUNT(e.id) AS times_completed,
			MAX(e.start_date) AS last_effort_date,
			MIN(NULLIF(e.elapsed_time, 0)) AS best_elapsed_time
		FROM segments s
		LEFT JOIN segment_efforts e ON e.segment_id = s.id AND e.athlete_id = ?
		WHERE 1=1
	`
	args := []any{athleteID}

	if f.ActivityType != "" {
		query += " AND s.activity_type = ?"
		args = append(args, f.ActivityType)
	}
	if f.Country != "" {
		query += " AND e.country = ?"
		args = append(args, f.Country)
	}
	if f.Starred != nil {
		query += " AND s.starred = ?"
		args = append(args, *f.Starred)
	}
	if f.KOMOnly {
		query += " AND s.athlete_kom_rank = 1"
	}
	if f.Search != "" {
		query += " AND s.name LIKE ? COLLATE NOCASE"
		args = append(args, "%"+f.Search+"%")
	}

	query += `
		GROUP BY
			s.id, s.name, s.activity_type, s.distance, s.average_grade, s.maximum_grade,
			s.elevation_high, s.elevation_low, s.climb_category,
			s.start_lat, s.start_lng, s.end_lat, s.end_lng,
			s.starred, s.polyline,
			s.athlete_kom_rank, s.athlete_effort_count, s.athlete_pr_elapsed_time, s.athlete_pr_date,
			s.created_at, s.updated_at
		ORDER BY times_completed DESC, s.distance DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []SegmentListItem
	for rows.Next() {
		var it SegmentListItem
		err := rows.Scan(
			&it.ID, &it.Name, &it.ActivityType, &it.Distance, &it.AverageGrade, &it.MaximumGrade,
			&it.ElevationHigh, &it.ElevationLow, &it.ClimbCategory,
			&it.StartLat, &it.StartLng, &it.EndLat, &it.EndLng,
			&it.Starred, &it.Polyline,
			&it.AthleteKOMRank, &it.AthleteEffortCount, &it.AthletePRElapsedTime, &it.AthletePRDate,
			&it.CreatedAt, &it.UpdatedAt,
			&it.TimesCompleted, &it.LastEffortDate, &it.BestElapsedTime,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *SegmentRepository) GetByID(ctx context.Context, id int64) (*Segment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			id, name, activity_type, distance, average_grade, maximum_grade,
			elevation_high, elevation_low, climb_category,
			start_lat, start_lng, end_lat, end_lng,
			starred, COALESCE(polyline, ''),
			athlete_kom_rank, athlete_effort_count, athlete_pr_elapsed_time, athlete_pr_date,
			created_at, updated_at
		FROM segments
		WHERE id = ?
	`, id)

	var s Segment
	err := row.Scan(
		&s.ID, &s.Name, &s.ActivityType, &s.Distance, &s.AverageGrade, &s.MaximumGrade,
		&s.ElevationHigh, &s.ElevationLow, &s.ClimbCategory,
		&s.StartLat, &s.StartLng, &s.EndLat, &s.EndLng,
		&s.Starred, &s.Polyline,
		&s.AthleteKOMRank, &s.AthleteEffortCount, &s.AthletePRElapsedTime, &s.AthletePRDate,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SegmentRepository) ListEfforts(ctx context.Context, athleteID int64, segmentID int64, limit int) ([]SegmentEffort, error) {
	if limit <= 0 || limit > 5000 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id, segment_id, activity_id, athlete_id, name,
			elapsed_time, moving_time, start_date, start_date_local,
			distance, average_watts, average_heartrate, max_heartrate,
			pr_rank, COALESCE(country, ''), created_at
		FROM segment_efforts
		WHERE athlete_id = ? AND segment_id = ?
		ORDER BY start_date DESC
		LIMIT ?
	`, athleteID, segmentID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var efforts []SegmentEffort
	for rows.Next() {
		var e SegmentEffort
		err := rows.Scan(
			&e.ID, &e.SegmentID, &e.ActivityID, &e.AthleteID, &e.Name,
			&e.ElapsedTime, &e.MovingTime, &e.StartDate, &e.StartDateLocal,
			&e.Distance, &e.AverageWatts, &e.AverageHeartrate, &e.MaxHeartrate,
			&e.PRRank, &e.Country, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		efforts = append(efforts, e)
	}
	return efforts, rows.Err()
}

func (r *SegmentRepository) GetCountries(ctx context.Context, athleteID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT COALESCE(country, '') AS country
		FROM segment_efforts
		WHERE athlete_id = ? AND COALESCE(country, '') != ''
		ORDER BY country ASC
	`, athleteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

type CountryStat struct {
	Country string `json:"country"`
	Count   int    `json:"count"`
}

func (r *SegmentRepository) ListCountryStats(ctx context.Context, athleteID int64) ([]CountryStat, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT COALESCE(country, '') AS country, COUNT(DISTINCT segment_id) AS count
		FROM segment_efforts
		WHERE athlete_id = ? AND COALESCE(country, '') != ''
		GROUP BY country
		ORDER BY count DESC, country ASC
	`, athleteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []CountryStat
	for rows.Next() {
		var s CountryStat
		if err := rows.Scan(&s.Country, &s.Count); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *SegmentRepository) Count(ctx context.Context, athleteID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT s.id)
		FROM segments s
		JOIN segment_efforts e ON e.segment_id = s.id
		WHERE e.athlete_id = ?
	`, athleteID).Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("counting segments: %w", err)
	}
	return count, nil
}
