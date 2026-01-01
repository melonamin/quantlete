package storage

import (
	"context"
	"database/sql"
	"fmt"
)

type BestEffort struct {
	AthleteID    int64
	ActivityID   int64
	SportType    string
	DistanceType string
	Name         string
	DistanceM    float64
	ElapsedTimeS int
	MovingTimeS  *int
	StartIndex   *int
	EndIndex     *int
	PRRank       *int
	StartDate    *SQLiteTime
}

type BestEffortPR struct {
	DistanceType   string
	Name           string
	DistanceM      float64
	ElapsedTimeS   int
	MovingTimeS    *int
	PRRank         *int
	ActivityID     int64
	ActivityName   string
	SportType      string
	StartDateLocal SQLiteTime
}

type BestEffortListItem struct {
	DistanceType   string
	Name           string
	DistanceM      float64
	ElapsedTimeS   int
	MovingTimeS    *int
	PRRank         *int
	StartIndex     *int
	EndIndex       *int
	ActivityID     int64
	ActivityName   string
	SportType      string
	StartDateLocal SQLiteTime
}

type BestEffortsRepository struct {
	db *DB
}

func NewBestEffortsRepository(db *DB) *BestEffortsRepository {
	return &BestEffortsRepository{db: db}
}

func (r *BestEffortsRepository) ReplaceForActivity(ctx context.Context, athleteID int64, activityID int64, sportType string, efforts []BestEffort) error {
	tx, err := r.db.Conn().BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM best_efforts WHERE athlete_id = ? AND activity_id = ?`, athleteID, activityID); err != nil {
		return err
	}

	if len(efforts) == 0 {
		return tx.Commit()
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO best_efforts (
			athlete_id, activity_id, sport_type,
			distance_type, name, distance_m,
			elapsed_time, moving_time,
			start_index, end_index,
			pr_rank, start_date
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for _, e := range efforts {
		if e.DistanceType == "" || e.DistanceM <= 0 || e.ElapsedTimeS <= 0 {
			continue
		}
		name := e.Name
		if name == "" {
			name = e.DistanceType
		}
		if _, err := stmt.ExecContext(
			ctx,
			athleteID,
			activityID,
			sportType,
			e.DistanceType,
			name,
			e.DistanceM,
			e.ElapsedTimeS,
			nullIntPtr(e.MovingTimeS),
			nullIntPtr(e.StartIndex),
			nullIntPtr(e.EndIndex),
			nullIntPtr(e.PRRank),
			nullTimePtr(e.StartDate),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *BestEffortsRepository) ListPRs(ctx context.Context, athleteID int64, sportTypes []string) ([]BestEffortPR, error) {
	query := `
		WITH ranked AS (
			SELECT
				be.distance_type,
				COALESCE(be.name, be.distance_type) AS name,
				be.distance_m,
				be.elapsed_time,
				be.moving_time,
				be.pr_rank,
				a.id AS activity_id,
				a.name AS activity_name,
				a.sport_type,
				a.start_date_local,
				ROW_NUMBER() OVER (
					PARTITION BY be.distance_type
					ORDER BY be.elapsed_time ASC, a.start_date_local ASC
				) AS rn
			FROM best_efforts be
			JOIN activities a
				ON a.id = be.activity_id AND a.athlete_id = be.athlete_id
			WHERE be.athlete_id = ?
	`
	args := []any{athleteID}
	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND a.sport_type IN (" + JoinStrings(placeholders, ",") + ")"
	}
	query += `
		)
		SELECT
			distance_type, name, distance_m, elapsed_time, moving_time, pr_rank,
			activity_id, activity_name, sport_type, start_date_local
		FROM ranked
		WHERE rn = 1
		ORDER BY distance_m ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []BestEffortPR
	for rows.Next() {
		var b BestEffortPR
		var moving sql.NullInt64
		var prRank sql.NullInt64
		if err := rows.Scan(
			&b.DistanceType,
			&b.Name,
			&b.DistanceM,
			&b.ElapsedTimeS,
			&moving,
			&prRank,
			&b.ActivityID,
			&b.ActivityName,
			&b.SportType,
			&b.StartDateLocal,
		); err != nil {
			return nil, err
		}
		if moving.Valid {
			v := int(moving.Int64)
			b.MovingTimeS = &v
		}
		if prRank.Valid {
			v := int(prRank.Int64)
			b.PRRank = &v
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (r *BestEffortsRepository) ListByDistanceType(ctx context.Context, athleteID int64, distanceType string, sportTypes []string) ([]BestEffortListItem, error) {
	if distanceType == "" {
		return nil, fmt.Errorf("distance_type is required")
	}
	query := `
		SELECT
			be.distance_type,
			COALESCE(be.name, be.distance_type) AS name,
			be.distance_m,
			be.elapsed_time,
			be.moving_time,
			be.pr_rank,
			be.start_index,
			be.end_index,
			a.id AS activity_id,
			a.name AS activity_name,
			a.sport_type,
			a.start_date_local
		FROM best_efforts be
		JOIN activities a
			ON a.id = be.activity_id AND a.athlete_id = be.athlete_id
		WHERE be.athlete_id = ? AND be.distance_type = ?
	`
	args := []any{athleteID, distanceType}
	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND a.sport_type IN (" + JoinStrings(placeholders, ",") + ")"
	}
	query += " ORDER BY a.start_date_local ASC, be.elapsed_time ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []BestEffortListItem
	for rows.Next() {
		var it BestEffortListItem
		var moving sql.NullInt64
		var prRank sql.NullInt64
		var startIdx sql.NullInt64
		var endIdx sql.NullInt64
		if err := rows.Scan(
			&it.DistanceType,
			&it.Name,
			&it.DistanceM,
			&it.ElapsedTimeS,
			&moving,
			&prRank,
			&startIdx,
			&endIdx,
			&it.ActivityID,
			&it.ActivityName,
			&it.SportType,
			&it.StartDateLocal,
		); err != nil {
			return nil, err
		}
		if moving.Valid {
			v := int(moving.Int64)
			it.MovingTimeS = &v
		}
		if prRank.Valid {
			v := int(prRank.Int64)
			it.PRRank = &v
		}
		if startIdx.Valid {
			v := int(startIdx.Int64)
			it.StartIndex = &v
		}
		if endIdx.Valid {
			v := int(endIdx.Int64)
			it.EndIndex = &v
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func nullIntPtr(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullTimePtr(t *SQLiteTime) any {
	if t == nil {
		return nil
	}
	return t.Time
}
