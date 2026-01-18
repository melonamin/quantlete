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

// ComputationStats provides diagnostic information about TSS computation.
type ComputationStats struct {
	TotalActivities      int  `json:"total_activities"`
	ActivitiesWithPower  int  `json:"activities_with_power"`
	ActivitiesWithTSS    int  `json:"activities_with_tss"`
	CyclingFTPConfigured bool `json:"cycling_ftp_configured"`
	RunningFTPConfigured bool `json:"running_ftp_configured"`
	HRZonesConfigured    bool `json:"hr_zones_configured"`
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
		args = append(args, SQLiteTime{Time: *after})
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, SQLiteTime{Time: *before})
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
	slog.Debug("computing TSS for activity",
		"activity_id", a.ID,
		"sport_type", a.SportType,
		"start_date", a.StartDate.Format("2006-01-02"),
		"moving_time_s", a.MovingTimeS)

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

	hasPower := len(wattsRaw) > 0
	hasSpeed := len(speedRaw) > 0
	hasHR := len(hrRaw) > 0
	isRun := strings.Contains(a.SportType, "Run")

	slog.Debug("activity stream availability",
		"activity_id", a.ID,
		"has_power", hasPower,
		"has_speed", hasSpeed,
		"has_hr", hasHR,
		"is_run", isRun)

	// 1) Cycling: power-based TSS.
	if hasPower {
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
			slog.Debug("computed TSS using cycling power",
				"activity_id", a.ID,
				"method", "cycling_power",
				"ftp", ftp,
				"np", np,
				"if", ifactor,
				"tss", tss)
			return r.upsertActivity(ctx, athleteID, a, "cycling_power", ftp, np, ifactor, tss)
		}
		slog.Debug("skipping activity: has power data but no cycling FTP configured",
			"activity_id", a.ID,
			"sport_type", a.SportType)
	}

	// 2) Running: pace/speed-based TSS (threshold speed).
	// NOTE: This applies the cycling Normalized Power algorithm (30s rolling avg, 4th power)
	// to running speed data. While not physiologically identical to running-specific metrics
	// like NGP (Normalized Graded Pace), it provides a reasonable TSS approximation for
	// comparing training load across activities. Values are not directly comparable to
	// cycling TSS or TrainingPeaks rTSS.
	if isRun && hasSpeed {
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
			slog.Debug("computed TSS using running pace",
				"activity_id", a.ID,
				"method", "running_pace",
				"threshold_speed", ftp,
				"np", np,
				"if", ifactor,
				"tss", tss)
			return r.upsertActivity(ctx, athleteID, a, "running_pace", ftp, np, ifactor, tss)
		}
		slog.Debug("skipping activity: has speed data but no running FTP configured",
			"activity_id", a.ID,
			"sport_type", a.SportType)
	}

	// 3) Running: HR-based TSS (approximate LTHR from HR zone definition).
	// NOTE: Similar to pace-based TSS, this applies the cycling NP algorithm to heart rate
	// data with threshold HR as the "FTP" equivalent. This is an approximation that enables
	// training load tracking when pace/power data isn't available.
	if isRun && hasHR && r.zones != nil {
		def, cfg, err := r.zones.GetApplicableHR(ctx, athleteID, a.SportType, a.StartDate.Time)
		if err != nil {
			return err
		}
		if cfg != nil && len(cfg.Bounds) >= 4 {
			threshold := cfg.Bounds[3]
			if def != nil && def.Method == "percent_hrmax" {
				if cfg.HRMax <= 0 {
					slog.Debug("skipping activity: HR zones use percent_hrmax but HRMax not configured",
						"activity_id", a.ID,
						"sport_type", a.SportType)
					return nil
				}
				threshold *= cfg.HRMax
			}
			if threshold > 0 {
				hrs, err := decodeFloat64Array(hrRaw)
				if err != nil {
					return err
				}
				nhr := analysis.NormalizedPower(hrs)
				ifactor := analysis.IntensityFactor(nhr, threshold)
				tss := analysis.TrainingStressScore(a.MovingTimeS, nhr, threshold)
				slog.Debug("computed TSS using running HR",
					"activity_id", a.ID,
					"method", "running_hr",
					"threshold_hr", threshold,
					"nhr", nhr,
					"if", ifactor,
					"tss", tss)
				return r.upsertActivity(ctx, athleteID, a, "running_hr", threshold, nhr, ifactor, tss)
			}
		}
		slog.Debug("skipping activity: has HR data but no valid HR zones configured",
			"activity_id", a.ID,
			"sport_type", a.SportType)
	}

	// Log why we couldn't compute TSS for this activity
	if !hasPower && !hasSpeed && !hasHR {
		slog.Debug("skipping activity: no power, speed, or HR stream data available",
			"activity_id", a.ID,
			"sport_type", a.SportType)
	} else if !isRun && !hasPower {
		slog.Debug("skipping activity: non-running activity without power data",
			"activity_id", a.ID,
			"sport_type", a.SportType)
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
		args = append(args, SQLiteTime{Time: *after})
	}
	if before != nil {
		query += " AND a.start_date_local <= ?"
		args = append(args, SQLiteTime{Time: *before})
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

// GetComputationStats returns diagnostic information about TSS computation status.
func (r *TrainingLoadRepository) GetComputationStats(ctx context.Context, athleteID int64, after, before *time.Time) (*ComputationStats, error) {
	stats := &ComputationStats{}

	// Build date filter clause
	dateClause := ""
	args := []any{athleteID}
	if after != nil {
		dateClause += " AND a.start_date >= ?"
		args = append(args, SQLiteTime{Time: *after})
	}
	if before != nil {
		dateClause += " AND a.start_date <= ?"
		args = append(args, SQLiteTime{Time: *before})
	}

	// Count total activities in range
	var totalActivities int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM activities a
		WHERE a.athlete_id = ?`+dateClause, args...).Scan(&totalActivities)
	if err != nil {
		return nil, err
	}
	stats.TotalActivities = totalActivities

	// Count activities with power streams
	var withPower int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT a.id)
		FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id
		WHERE a.athlete_id = ? AND s.stream_type = 'watts'`+dateClause, args...).Scan(&withPower)
	if err != nil {
		return nil, err
	}
	stats.ActivitiesWithPower = withPower

	// Count activities with computed TSS
	var withTSS int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM activities a
		JOIN activity_training_load tl ON tl.activity_id = a.id
		WHERE a.athlete_id = ? AND tl.tss > 0`+dateClause, args...).Scan(&withTSS)
	if err != nil {
		return nil, err
	}
	stats.ActivitiesWithTSS = withTSS

	// Check if cycling FTP is configured
	ftpPoint, err := r.metrics.LatestBefore(ctx, athleteID, "ftp_cycling_watts", time.Now())
	if err != nil {
		return nil, err
	}
	stats.CyclingFTPConfigured = ftpPoint != nil && ftpPoint.Value > 0

	// Check if running FTP is configured
	runFtpPoint, err := r.metrics.LatestBefore(ctx, athleteID, "ftp_running_mps", time.Now())
	if err != nil {
		return nil, err
	}
	stats.RunningFTPConfigured = runFtpPoint != nil && runFtpPoint.Value > 0

	// Check if HR zones are configured
	if r.zones != nil {
		// Check for any HR zone definition
		var zoneCount int
		err = r.db.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM hr_zone_definitions
			WHERE athlete_id = ?`, athleteID).Scan(&zoneCount)
		if err != nil {
			return nil, err
		}
		stats.HRZonesConfigured = zoneCount > 0
	}

	return stats, nil
}
