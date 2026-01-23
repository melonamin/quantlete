package storage

import (
	"context"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/melonamin/quantlete/internal/analysis"
)

type DailyTrainingLoadPoint struct {
	Day string  `json:"day"` // YYYY-MM-DD
	TSS float64 `json:"tss"`
	CTL float64 `json:"ctl"`
	ATL float64 `json:"atl"`
	TSB float64 `json:"tsb"`
}

type TrainingLoadDiagnostics struct {
	TotalActivities       int      `json:"total_activities"`
	ActivitiesWithPower   int      `json:"activities_with_power"`
	ActivitiesWithSpeed   int      `json:"activities_with_speed"`
	ActivitiesWithHR      int      `json:"activities_with_hr"`
	ActivitiesWithTSS     int      `json:"activities_with_tss"`
	HasCyclingFTP         bool     `json:"has_cycling_ftp"`
	HasRunningFTP         bool     `json:"has_running_ftp"`
	CyclingFTPValue       *float64 `json:"cycling_ftp_value,omitempty"`
	RunningFTPValue       *float64 `json:"running_ftp_value,omitempty"`
	MissingConfigWarnings []string `json:"missing_config_warnings,omitempty"`
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
	StartDate   SQLiteTime
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
		args = append(args, TimeToSQL(*after))
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, TimeToSQL(*before))
	}
	query += " ORDER BY a.start_date ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	// Collect all activities first to avoid nested queries with open rows cursor.
	// SQLite with single connection can deadlock if we query while rows are open.
	var activities []ActivityForLoad
	for rows.Next() {
		var a ActivityForLoad
		if err := rows.Scan(&a.ID, &a.SportType, &a.StartDate, &a.MovingTimeS); err != nil {
			_ = rows.Close()
			return err
		}
		activities = append(activities, a)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return err
	}
	_ = rows.Close()

	// Now process each activity with the cursor closed
	for _, a := range activities {
		if err := r.computeAndUpsertActivity(ctx, athleteID, a); err != nil {
			return err
		}
	}
	return nil
}

func (r *TrainingLoadRepository) computeAndUpsertActivity(ctx context.Context, athleteID int64, a ActivityForLoad) error {
	streams, err := r.streams.GetByActivityID(ctx, a.ID)
	if err != nil {
		slog.Debug("training load: failed to get streams", "activity_id", a.ID, "error", err)
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

	slog.Debug("training load: processing activity",
		"activity_id", a.ID,
		"sport_type", a.SportType,
		"has_watts", len(wattsRaw) > 0,
		"has_speed", len(speedRaw) > 0,
		"has_hr", len(hrRaw) > 0,
	)

	// 1) Cycling: power-based TSS.
	if len(wattsRaw) > 0 {
		ftpPoint, err := r.metrics.LatestBefore(ctx, athleteID, "ftp_cycling_watts", a.StartDate.Time)
		if err != nil {
			slog.Debug("training load: failed to get cycling FTP", "activity_id", a.ID, "error", err)
			return err
		}
		if ftpPoint != nil && ftpPoint.Value > 0 {
			ftp := ftpPoint.Value
			watts, err := decodeFloat64Array(wattsRaw)
			if err != nil {
				slog.Debug("training load: failed to decode watts", "activity_id", a.ID, "error", err)
				return err
			}
			np := analysis.NormalizedPower(watts)
			ifactor := analysis.IntensityFactor(np, ftp)
			tss := analysis.TrainingStressScore(a.MovingTimeS, np, ftp)
			slog.Debug("training load: computed cycling TSS",
				"activity_id", a.ID,
				"ftp", ftp,
				"np", np,
				"if", ifactor,
				"tss", tss,
			)
			return r.upsertActivity(ctx, athleteID, a, "cycling_power", ftp, np, ifactor, tss)
		}
		slog.Debug("training load: skipping cycling activity, no FTP configured",
			"activity_id", a.ID,
			"ftp_found", ftpPoint != nil,
		)
	}

	isRun := strings.Contains(a.SportType, "Run")

	// 2) Running: pace/speed-based TSS (threshold speed).
	// NOTE: This applies the cycling Normalized Power algorithm (30s rolling avg, 4th power)
	// to running speed data. While not physiologically identical to running-specific metrics
	// like NGP (Normalized Graded Pace), it provides a reasonable TSS approximation for
	// comparing training load across activities. Values are not directly comparable to
	// cycling TSS or TrainingPeaks rTSS.
	if isRun && len(speedRaw) > 0 {
		ftpPoint, err := r.metrics.LatestBefore(ctx, athleteID, "ftp_running_mps", a.StartDate.Time)
		if err != nil {
			slog.Debug("training load: failed to get running FTP", "activity_id", a.ID, "error", err)
			return err
		}
		if ftpPoint != nil && ftpPoint.Value > 0 {
			ftp := ftpPoint.Value
			speeds, err := decodeFloat64Array(speedRaw)
			if err != nil {
				slog.Debug("training load: failed to decode speeds", "activity_id", a.ID, "error", err)
				return err
			}
			np := analysis.NormalizedPower(speeds)
			ifactor := analysis.IntensityFactor(np, ftp)
			tss := analysis.TrainingStressScore(a.MovingTimeS, np, ftp)
			slog.Debug("training load: computed running pace TSS",
				"activity_id", a.ID,
				"ftp", ftp,
				"np", np,
				"if", ifactor,
				"tss", tss,
			)
			return r.upsertActivity(ctx, athleteID, a, "running_pace", ftp, np, ifactor, tss)
		}
		slog.Debug("training load: skipping running activity, no pace FTP configured",
			"activity_id", a.ID,
			"ftp_found", ftpPoint != nil,
		)
	}

	// 3) Running: HR-based TSS (approximate LTHR from HR zone definition).
	// NOTE: Similar to pace-based TSS, this applies the cycling NP algorithm to heart rate
	// data with threshold HR as the "FTP" equivalent. This is an approximation that enables
	// training load tracking when pace/power data isn't available.
	if isRun && len(hrRaw) > 0 && r.zones != nil {
		def, cfg, err := r.zones.GetApplicableHR(ctx, athleteID, a.SportType, a.StartDate.Time)
		if err != nil {
			slog.Debug("training load: failed to get HR zones", "activity_id", a.ID, "error", err)
			return err
		}
		if cfg != nil && len(cfg.Bounds) >= 4 {
			threshold := cfg.Bounds[3]
			if def != nil && def.Method == "percent_hrmax" {
				if cfg.HRMax <= 0 {
					slog.Debug("training load: skipping HR-based TSS, no max HR configured",
						"activity_id", a.ID,
					)
					return nil
				}
				threshold *= cfg.HRMax
			}
			if threshold > 0 {
				hrs, err := decodeFloat64Array(hrRaw)
				if err != nil {
					slog.Debug("training load: failed to decode heart rates", "activity_id", a.ID, "error", err)
					return err
				}
				nhr := analysis.NormalizedPower(hrs)
				ifactor := analysis.IntensityFactor(nhr, threshold)
				tss := analysis.TrainingStressScore(a.MovingTimeS, nhr, threshold)
				slog.Debug("training load: computed running HR TSS",
					"activity_id", a.ID,
					"threshold", threshold,
					"nhr", nhr,
					"if", ifactor,
					"tss", tss,
				)
				return r.upsertActivity(ctx, athleteID, a, "running_hr", threshold, nhr, ifactor, tss)
			}
		}
		slog.Debug("training load: skipping running HR activity, no valid HR zone config",
			"activity_id", a.ID,
			"has_config", cfg != nil,
		)
	}

	slog.Debug("training load: no computation method matched",
		"activity_id", a.ID,
		"sport_type", a.SportType,
		"is_run", isRun,
		"has_watts", len(wattsRaw) > 0,
		"has_speed", len(speedRaw) > 0,
		"has_hr", len(hrRaw) > 0,
	)
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
	`, a.ID, athleteID, a.SportType, method, ftpUsed, normalized, ifactor, tss, TimeToSQL(time.Now()))
	return err
}

func (r *TrainingLoadRepository) GetDailySeries(ctx context.Context, athleteID int64, after, before *time.Time) ([]DailyTrainingLoadPoint, error) {
	// Query active days with TSS values.
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
		args = append(args, TimeToSQL(*after))
	}
	if before != nil {
		query += " AND a.start_date_local <= ?"
		args = append(args, TimeToSQL(*before))
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

	// Build a map of day -> TSS for quick lookup.
	tssByDay := make(map[string]float64)
	var firstDay, lastDay time.Time
	for rows.Next() {
		var day SQLiteTime
		var tss float64
		if err := rows.Scan(&day, &tss); err != nil {
			return nil, err
		}
		dayStr := day.Format("2006-01-02")
		tssByDay[dayStr] = tss
		dayTime := day.Time
		if firstDay.IsZero() || dayTime.Before(firstDay) {
			firstDay = dayTime
		}
		if lastDay.IsZero() || dayTime.After(lastDay) {
			lastDay = dayTime
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(tssByDay) == 0 {
		return nil, nil
	}

	// Apply date range filters if provided.
	if after != nil && after.After(firstDay) {
		firstDay = *after
	}
	if before != nil && before.Before(lastDay) {
		lastDay = *before
	}

	// Normalize to start of day.
	firstDay = time.Date(firstDay.Year(), firstDay.Month(), firstDay.Day(), 0, 0, 0, 0, time.UTC)
	lastDay = time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), 0, 0, 0, 0, time.UTC)

	// EWMA constants.
	const ctlTau = 42.0
	const atlTau = 7.0

	// Initialize CTL/ATL to 0, letting the EWMA build naturally.
	var ctl, atl float64
	totalDays := int(lastDay.Sub(firstDay).Hours()/24) + 1
	out := make([]DailyTrainingLoadPoint, 0, totalDays)

	// Iterate through every day in the range, including rest days (TSS=0).
	for d := firstDay; !d.After(lastDay); d = d.AddDate(0, 0, 1) {
		dayStr := d.Format("2006-01-02")
		tss := tssByDay[dayStr] // 0 if not present (rest day)

		// Apply EWMA formula for each day.
		ctl += (tss - ctl) * (1.0 / ctlTau)
		atl += (tss - atl) * (1.0 / atlTau)
		tsb := ctl - atl

		out = append(out, DailyTrainingLoadPoint{
			Day: dayStr,
			TSS: round2(tss),
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
		// Validate date format but use the string directly (go-sqlite3-js can't serialize time.Time)
		if _, err := time.Parse("2006-01-02", p.Day); err != nil {
			continue
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO daily_training_load (athlete_id, day, tss, ctl, atl, tsb)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT (athlete_id, day) DO UPDATE SET
				tss = EXCLUDED.tss,
				ctl = EXCLUDED.ctl,
				atl = EXCLUDED.atl,
				tsb = EXCLUDED.tsb
		`, athleteID, p.Day, p.TSS, p.CTL, p.ATL, p.TSB)
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
	q := NewQueries(r.db.Conn())
	row, err := q.GetTrainingLoadSummary(ctx, athleteID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return &DailyTrainingLoadPoint{
		Day: row.Day,
		TSS: round2(row.TSS),
		CTL: round2(row.CTL),
		ATL: round2(row.ATL),
		TSB: round2(row.TSB),
	}, nil
}

func (r *TrainingLoadRepository) GetActivityTSS(ctx context.Context, athleteID, activityID int64) (float64, error) {
	q := NewQueries(r.db.Conn())
	row, err := q.GetActivityTSSByAthlete(ctx, athleteID, activityID)
	if err != nil {
		return 0, err
	}
	if row == nil {
		return 0, nil
	}
	return row.TSS, nil
}

func (r *TrainingLoadRepository) GetDiagnostics(ctx context.Context, athleteID int64) (*TrainingLoadDiagnostics, error) {
	diag := &TrainingLoadDiagnostics{}

	// Get total activity count
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM activities WHERE athlete_id = ?
	`, athleteID).Scan(&diag.TotalActivities)
	if err != nil {
		return nil, err
	}

	// Get count of activities with power data
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT a.id) FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id
		WHERE a.athlete_id = ? AND s.stream_type = 'watts'
	`, athleteID).Scan(&diag.ActivitiesWithPower)
	if err != nil {
		return nil, err
	}

	// Get count of activities with speed data
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT a.id) FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id
		WHERE a.athlete_id = ? AND s.stream_type = 'velocity_smooth'
	`, athleteID).Scan(&diag.ActivitiesWithSpeed)
	if err != nil {
		return nil, err
	}

	// Get count of activities with HR data
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT a.id) FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id
		WHERE a.athlete_id = ? AND s.stream_type = 'heartrate'
	`, athleteID).Scan(&diag.ActivitiesWithHR)
	if err != nil {
		return nil, err
	}

	// Get count of activities with computed TSS
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM activity_training_load WHERE athlete_id = ?
	`, athleteID).Scan(&diag.ActivitiesWithTSS)
	if err != nil {
		return nil, err
	}

	// Check for cycling FTP
	cyclingFTP, err := r.metrics.GetCurrentFTP(ctx, athleteID, "ftp_cycling_watts")
	if err != nil {
		return nil, err
	}
	if cyclingFTP != nil && cyclingFTP.Value > 0 {
		diag.HasCyclingFTP = true
		diag.CyclingFTPValue = &cyclingFTP.Value
	}

	// Check for running FTP
	runningFTP, err := r.metrics.GetCurrentFTP(ctx, athleteID, "ftp_running_mps")
	if err != nil {
		return nil, err
	}
	if runningFTP != nil && runningFTP.Value > 0 {
		diag.HasRunningFTP = true
		diag.RunningFTPValue = &runningFTP.Value
	}

	// Generate warnings
	if diag.ActivitiesWithPower > 0 && !diag.HasCyclingFTP {
		diag.MissingConfigWarnings = append(diag.MissingConfigWarnings,
			"You have activities with power data but no cycling FTP configured. Set your FTP in Settings to enable TSS calculation.")
	}
	if diag.ActivitiesWithSpeed > 0 && !diag.HasRunningFTP {
		// Count running activities specifically
		var runningWithSpeed int
		_ = r.db.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT a.id) FROM activities a
			JOIN activity_streams s ON s.activity_id = a.id
			WHERE a.athlete_id = ? AND s.stream_type = 'velocity_smooth' AND a.sport_type LIKE '%Run%'
		`, athleteID).Scan(&runningWithSpeed)
		if runningWithSpeed > 0 {
			diag.MissingConfigWarnings = append(diag.MissingConfigWarnings,
				"You have running activities with pace data but no running threshold pace configured. Set your threshold pace in Settings to enable TSS calculation.")
		}
	}

	return diag, nil
}
