package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/melonamin/quantlete/internal/analysis"
)

type PeakPowerBest struct {
	DurationS  int        `json:"duration_s"`
	Watts      float64    `json:"watts"`
	ActivityID int64      `json:"activity_id"`
	StartDate  SQLiteTime `json:"start_date"`
}

type PeakPowerHistoryPoint struct {
	Date  string  `json:"date"`
	Watts float64 `json:"watts"`
}

type PowerRepository struct {
	db      *DB
	streams *StreamRepository
}

func NewPowerRepository(db *DB, streams *StreamRepository) *PowerRepository {
	return &PowerRepository{db: db, streams: streams}
}

func (r *PowerRepository) EnsureActivityComputed(ctx context.Context, athleteID int64, activityID int64, durations []int) error {
	if len(durations) == 0 {
		return nil
	}

	// Quick existence check for this activity.
	rows, err := r.db.QueryContext(ctx, `
		SELECT duration_s
		FROM power_best_efforts
		WHERE athlete_id = ? AND activity_id = ?
	`, athleteID, activityID)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	existing := map[int]bool{}
	for rows.Next() {
		var d int
		if err := rows.Scan(&d); err != nil {
			return err
		}
		existing[d] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}

	missing := make([]int, 0, len(durations))
	for _, d := range durations {
		if !existing[d] {
			missing = append(missing, d)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	streams, err := r.streams.GetByActivityID(ctx, activityID)
	if err != nil {
		return err
	}

	var wattsRaw []byte
	for _, s := range streams {
		if s.StreamType == "watts" {
			wattsRaw = s.Data
			break
		}
	}
	if len(wattsRaw) == 0 {
		return nil
	}

	watts, err := decodeFloat64Array(wattsRaw)
	if err != nil {
		return err
	}

	now := SQLiteTime{Time: time.Now()}
	for _, d := range missing {
		best := analysis.RollingMaxAverage(watts, d)
		if best <= 0 {
			continue
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO power_best_efforts (activity_id, athlete_id, duration_s, best_avg_watts, computed_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (activity_id, duration_s) DO UPDATE SET
				best_avg_watts = EXCLUDED.best_avg_watts,
				computed_at = EXCLUDED.computed_at
		`, activityID, athleteID, d, best, now)
		if err != nil {
			return fmt.Errorf("upserting power_best_efforts: %w", err)
		}
	}

	return nil
}

func (r *PowerRepository) EnsureComputedForRange(ctx context.Context, athleteID int64, after, before *time.Time, sportTypes []string, durations []int) error {
	query := `
		SELECT DISTINCT a.id
		FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id AND s.stream_type = 'watts'
		WHERE a.athlete_id = ?
	`
	args := []any{athleteID}

	if after != nil {
		query += " AND a.start_date >= ?"
		args = append(args, *after)
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, *before)
	}
	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND a.sport_type IN (" + joinStrings(placeholders, ",") + ")"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Collect all activity IDs first to avoid nested queries with open rows cursor.
	// SQLite with single connection can deadlock if we query while rows are open.
	var activityIDs []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			_ = rows.Close()
			return err
		}
		activityIDs = append(activityIDs, id)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	// Now process each activity with the cursor closed
	for _, id := range activityIDs {
		if err := r.EnsureActivityComputed(ctx, athleteID, id, durations); err != nil {
			return err
		}
	}
	return nil
}

func (r *PowerRepository) GetBest(ctx context.Context, athleteID int64, durations []int, after, before *time.Time, sportTypes []string) ([]PeakPowerBest, error) {
	if len(durations) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(durations))
	args := []any{athleteID}
	for i, d := range durations {
		placeholders[i] = "?"
		args = append(args, d)
	}

	// Use window function to get activity_id and start_date for the max watts per duration
	query := `
		WITH ranked AS (
			SELECT
				p.duration_s,
				p.best_avg_watts,
				p.activity_id,
				a.start_date,
				ROW_NUMBER() OVER (PARTITION BY p.duration_s ORDER BY p.best_avg_watts DESC) AS rn
			FROM power_best_efforts p
			JOIN activities a ON a.id = p.activity_id
			WHERE p.athlete_id = ? AND p.duration_s IN (` + joinStrings(placeholders, ",") + `)
	`

	if after != nil {
		query += " AND a.start_date >= ?"
		args = append(args, *after)
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, *before)
	}
	if len(sportTypes) > 0 {
		stPlace := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			stPlace[i] = "?"
			args = append(args, st)
		}
		query += " AND a.sport_type IN (" + joinStrings(stPlace, ",") + ")"
	}

	query += `
		)
		SELECT duration_s, best_avg_watts AS watts, activity_id, start_date
		FROM ranked
		WHERE rn = 1
		ORDER BY duration_s ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var best []PeakPowerBest
	for rows.Next() {
		var b PeakPowerBest
		if err := rows.Scan(&b.DurationS, &b.Watts, &b.ActivityID, &b.StartDate); err != nil {
			return nil, err
		}
		best = append(best, b)
	}
	return best, rows.Err()
}

func (r *PowerRepository) GetHistory(ctx context.Context, athleteID int64, duration int, after, before *time.Time, sportTypes []string) ([]PeakPowerHistoryPoint, error) {
	query := `
		SELECT a.start_date, p.best_avg_watts
		FROM power_best_efforts p
		JOIN activities a ON a.id = p.activity_id
		WHERE p.athlete_id = ? AND p.duration_s = ?
	`
	args := []any{athleteID, duration}
	if after != nil {
		query += " AND a.start_date >= ?"
		args = append(args, *after)
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, *before)
	}
	if len(sportTypes) > 0 {
		stPlace := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			stPlace[i] = "?"
			args = append(args, st)
		}
		query += " AND a.sport_type IN (" + joinStrings(stPlace, ",") + ")"
	}
	query += " ORDER BY a.start_date ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var points []PeakPowerHistoryPoint
	best := 0.0
	for rows.Next() {
		var t SQLiteTime
		var w float64
		if err := rows.Scan(&t, &w); err != nil {
			return nil, err
		}
		if w > best {
			best = w
		}
		points = append(points, PeakPowerHistoryPoint{
			Date:  t.Format("2006-01-02"),
			Watts: best,
		})
	}
	return points, rows.Err()
}
