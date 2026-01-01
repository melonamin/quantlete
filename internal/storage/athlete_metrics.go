package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type AthleteMetricPoint struct {
	RecordedAt SQLiteTime `json:"recorded_at"`
	Value      float64    `json:"value"`
}

type AthleteMetricsRepository struct {
	db *DB
}

func NewAthleteMetricsRepository(db *DB) *AthleteMetricsRepository {
	return &AthleteMetricsRepository{db: db}
}

func (r *AthleteMetricsRepository) List(ctx context.Context, athleteID int64, metric string) ([]AthleteMetricPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT recorded_at, value
		FROM athlete_metrics
		WHERE athlete_id = ? AND metric = ?
		ORDER BY recorded_at ASC
	`, athleteID, metric)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var points []AthleteMetricPoint
	for rows.Next() {
		var p AthleteMetricPoint
		if err := rows.Scan(&p.RecordedAt, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (r *AthleteMetricsRepository) Replace(ctx context.Context, athleteID int64, metric string, points []AthleteMetricPoint) error {
	tx, err := r.db.Conn().BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err = tx.ExecContext(ctx, "DELETE FROM athlete_metrics WHERE athlete_id = ? AND metric = ?", athleteID, metric); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO athlete_metrics (athlete_id, metric, value, recorded_at)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()

	for _, p := range points {
		if _, err := stmt.ExecContext(ctx, athleteID, metric, p.Value, p.RecordedAt); err != nil {
			return fmt.Errorf("inserting point: %w", err)
		}
	}

	return tx.Commit()
}

func (r *AthleteMetricsRepository) LatestBefore(ctx context.Context, athleteID int64, metric string, t time.Time) (*AthleteMetricPoint, error) {
	var p AthleteMetricPoint
	err := r.db.QueryRowContext(ctx, `
		SELECT recorded_at, value
		FROM athlete_metrics
		WHERE athlete_id = ? AND metric = ? AND recorded_at <= ?
		ORDER BY recorded_at DESC
		LIMIT 1
	`, athleteID, metric, t).Scan(&p.RecordedAt, &p.Value)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}
