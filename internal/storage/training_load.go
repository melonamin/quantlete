package storage

import (
	"context"
	"database/sql"
	"math"
	"strings"
	"time"

	"github.com/sasha/stata/internal/analysis"
)

type DailyTrainingLoadPoint struct {
	Day string  `json:"day"` // YYYY-MM-DD
	TSS float64 `json:"tss"`
	CTL float64 `json:"ctl"`
	ATL float64 `json:"atl"`
	TSB float64 `json:"tsb"`
}

type TrainingLoadRepository struct {
	db      *DB
	streams *StreamRepository
	metrics *AthleteMetricsRepository
	zones   *ZonesRepository
}

func NewTrainingLoadRepository(db *DB, streams *StreamRepository, metrics *AthleteMetricsRepository, zones *ZonesRepository) *TrainingLoadRepository {
	return &TrainingLoadRepository{
		db:      db,
		streams: streams,
		metrics: metrics,
		zones:   zones,
	}
}

type ActivityForLoad struct {
	ID          int64
	SportType   string
	StartDate            SQLiteTime
	MovingTimeS int
}

func (r *TrainingLoadRepository) EnsureComputedForRange(ctx context.Context, athleteID int64, after, before *time.Time) error {
	query := `
		SELECT a.id, a.sport_type, a.start_date, a.moving_time
		FROM activities a
		LEFT JOIN activity_training_load tl ON tl.activity_id = a.id
		WHERE a.athlete_id = ? AND tl.activity_id IS NULL
			AND (
				EXISTS (SELECT 1 FROM activity_streams s WHERE s.activity_id = a.id AND s.stream_type = 'watts')
				OR (
					a.sport_type LIKE '%Run%'
					AND EXISTS (SELECT 1 FROM activity_streams s WHERE s.activity_id = a.id AND s.stream_type IN ('velocity_smooth', 'heartrate'))
				)
			)
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
	query += " ORDER BY a.start_date ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var a ActivityForLoad
		if err := rows.Scan(&a.ID, &a.SportType, &a.StartDate, &a.MovingTimeS); err != nil {
			return err
		}
		if err := r.computeAndUpsertActivity(ctx, athleteID, a); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (r *TrainingLoadRepository) computeAndUpsertActivity(ctx context.Context, athleteID int64, a ActivityForLoad) error {
	streams, err := r.streams.GetByActivityID(ctx, a.ID)
	if err != nil {
		return err
	}

	var wattsRaw []byte
	var speedRaw []byte
	var hrRaw []byte
	for _, s := range streams {
		if s.StreamType == "watts" {
			wattsRaw = s.Data
		}
		if s.StreamType == "velocity_smooth" {
			speedRaw = s.Data
		}
		if s.StreamType == "heartrate" {
			hrRaw = s.Data
		}
	}

	// 1) Cycling: power-based TSS.
	if len(wattsRaw) > 0 {
		ftpPoint, err := r.metrics.LatestBefore(ctx, athleteID, "ftp_cycling_watts", a.StartDate.Time)
		if err != nil {
			return err
		}
		if ftpPoint != nil && ftpPoint.Value > 0 {
			ftp := ftpPoint.Value
			watts, err := decodeFloat64Array(wattsRaw)
			if err != nil {
				return err
			}
			np := analysis.NormalizedPower(watts)
			ifactor := analysis.IntensityFactor(np, ftp)
			tss := analysis.TrainingStressScore(a.MovingTimeS, np, ftp)
			return r.upsertActivity(ctx, athleteID, a, "cycling_power", ftp, np, ifactor, tss)
		}
	}

	isRun := strings.Contains(a.SportType, "Run")

	// 2) Running: pace/speed-based TSS (threshold speed).
	if isRun && len(speedRaw) > 0 {
		ftpPoint, err := r.metrics.LatestBefore(ctx, athleteID, "ftp_running_mps", a.StartDate.Time)
		if err != nil {
			return err
		}
		if ftpPoint != nil && ftpPoint.Value > 0 {
			ftp := ftpPoint.Value
			speeds, err := decodeFloat64Array(speedRaw)
			if err != nil {
				return err
			}
			np := analysis.NormalizedPower(speeds)
			ifactor := analysis.IntensityFactor(np, ftp)
			tss := analysis.TrainingStressScore(a.MovingTimeS, np, ftp)
			return r.upsertActivity(ctx, athleteID, a, "running_pace", ftp, np, ifactor, tss)
		}
	}

	// 3) Running: HR-based TSS (approximate LTHR from HR zone definition).
	if isRun && len(hrRaw) > 0 && r.zones != nil {
		def, cfg, err := r.zones.GetApplicableHR(ctx, athleteID, a.SportType, a.StartDate.Time)
		if err != nil {
			return err
		}
		if cfg != nil && len(cfg.Bounds) >= 4 {
			threshold := cfg.Bounds[3]
			if def != nil && def.Method == "percent_hrmax" {
				if cfg.HRMax <= 0 {
					return nil
				}
				threshold = threshold * cfg.HRMax
			}
			if threshold > 0 {
				hrs, err := decodeFloat64Array(hrRaw)
				if err != nil {
					return err
				}
				nhr := analysis.NormalizedPower(hrs)
				ifactor := analysis.IntensityFactor(nhr, threshold)
				tss := analysis.TrainingStressScore(a.MovingTimeS, nhr, threshold)
				return r.upsertActivity(ctx, athleteID, a, "running_hr", threshold, nhr, ifactor, tss)
			}
		}
	}

	return nil
}

func (r *TrainingLoadRepository) upsertActivity(ctx context.Context, athleteID int64, a ActivityForLoad, method string, ftpUsed, normalized, ifactor, tss float64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO activity_training_load (activity_id, athlete_id, sport_type, method, ftp_used, normalized_power, intensity_factor, tss, computed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (activity_id) DO UPDATE SET
			method = EXCLUDED.method,
			ftp_used = EXCLUDED.ftp_used,
			normalized_power = EXCLUDED.normalized_power,
			intensity_factor = EXCLUDED.intensity_factor,
			tss = EXCLUDED.tss,
			computed_at = EXCLUDED.computed_at
	`, a.ID, athleteID, a.SportType, method, ftpUsed, normalized, ifactor, tss, SQLiteTime{Time: time.Now()})
	return err
}

func (r *TrainingLoadRepository) GetDailySeries(ctx context.Context, athleteID int64, after, before *time.Time) ([]DailyTrainingLoadPoint, error) {
	query := `
		SELECT
			date(a.start_date_local) AS day,
			COALESCE(SUM(tl.tss), 0) AS tss
		FROM activities a
		LEFT JOIN activity_training_load tl ON tl.activity_id = a.id
		WHERE a.athlete_id = ?
	`
	args := []any{athleteID}
	if after != nil {
		query += " AND a.start_date_local >= ?"
		args = append(args, *after)
	}
	if before != nil {
		query += " AND a.start_date_local <= ?"
		args = append(args, *before)
	}
	query += `
		GROUP BY day
		ORDER BY day ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	type dayRow struct {
		Day SQLiteTime
		TSS float64
	}
	var days []dayRow
	for rows.Next() {
		var d dayRow
		if err := rows.Scan(&d.Day, &d.TSS); err != nil {
			return nil, err
		}
		days = append(days, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// EWMA constants.
	const ctlTau = 42.0
	const atlTau = 7.0

	var ctl, atl float64
	out := make([]DailyTrainingLoadPoint, 0, len(days))
	for i, d := range days {
		if i == 0 {
			ctl = d.TSS
			atl = d.TSS
		} else {
			ctl += (d.TSS - ctl) * (1.0 / ctlTau)
			atl += (d.TSS - atl) * (1.0 / atlTau)
		}
		tsb := ctl - atl
		out = append(out, DailyTrainingLoadPoint{
			Day: d.Day.Format("2006-01-02"),
			TSS: round2(d.TSS),
			CTL: round2(ctl),
			ATL: round2(atl),
			TSB: round2(tsb),
		})
	}

	// Persist the computed daily series for reuse (best-effort).
	_ = r.upsertDaily(ctx, athleteID, out)

	return out, nil
}

func (r *TrainingLoadRepository) upsertDaily(ctx context.Context, athleteID int64, series []DailyTrainingLoadPoint) error {
	for _, p := range series {
		day, err := time.Parse("2006-01-02", p.Day)
		if err != nil {
			continue
		}
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO daily_training_load (athlete_id, day, tss, ctl, atl, tsb)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (athlete_id, day) DO UPDATE SET
				tss = EXCLUDED.tss,
				ctl = EXCLUDED.ctl,
				atl = EXCLUDED.atl,
				tsb = EXCLUDED.tsb
		`, athleteID, day, p.TSS, p.CTL, p.ATL, p.TSB)
		if err != nil {
			return err
		}
	}
	return nil
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

func (r *TrainingLoadRepository) GetSummary(ctx context.Context, athleteID int64) (*DailyTrainingLoadPoint, error) {
	var day SQLiteTime
	var tss, ctl, atl, tsb float64
	err := r.db.QueryRowContext(ctx, `
		SELECT day, tss, ctl, atl, tsb
		FROM daily_training_load
		WHERE athlete_id = ?
		ORDER BY day DESC
		LIMIT 1
	`, athleteID).Scan(&day, &tss, &ctl, &atl, &tsb)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &DailyTrainingLoadPoint{
		Day: day.Format("2006-01-02"),
		TSS: round2(tss),
		CTL: round2(ctl),
		ATL: round2(atl),
		TSB: round2(tsb),
	}, nil
}

func (r *TrainingLoadRepository) GetActivityTSS(ctx context.Context, athleteID int64, activityID int64) (float64, error) {
	var tss sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `
		SELECT tss
		FROM activity_training_load
		WHERE athlete_id = ? AND activity_id = ?
	`, athleteID, activityID).Scan(&tss)
	if err != nil {
		if isNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	if !tss.Valid {
		return 0, nil
	}
	return tss.Float64, nil
}
