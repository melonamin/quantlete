package storage

import (
	"context"
	"database/sql"
	"math"
	"time"
)

type RewindTotals struct {
	Activities    int     `json:"activities"`
	DistanceM     float64 `json:"distance_m"`
	ElevationM    float64 `json:"elevation_m"`
	MovingTimeS   int     `json:"moving_time_s"`
	Kudos         int     `json:"kudos"`
	CommuteDistM  float64 `json:"commute_distance_m"`
	CarbonSavedKg float64 `json:"carbon_saved_kg"`
}

type RewindMonth struct {
	Month      string  `json:"month"` // YYYY-MM
	Activities int     `json:"activities"`
	DistanceM  float64 `json:"distance_m"`
	ElevationM float64 `json:"elevation_m"`
	PRs        int     `json:"prs"`
}

type RewindSportTime struct {
	SportType   string `json:"sport_type"`
	MovingTimeS int    `json:"moving_time_s"`
}

type RewindHourCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

type RewindLocationPoint struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Count int     `json:"count"`
}

type RewindBiggestActivity struct {
	ActivityID     int64   `json:"activity_id"`
	Name           string  `json:"name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
	Value          float64 `json:"value"`
}

type RewindBiggest struct {
	LongestDistance *RewindBiggestActivity `json:"longest_distance,omitempty"`
	MostElevation   *RewindBiggestActivity `json:"most_elevation,omitempty"`
	LongestDuration *RewindBiggestActivity `json:"longest_duration,omitempty"`
}

type RewindStreaks struct {
	LongestActiveDays int `json:"longest_active_days"`
	LongestRestDays   int `json:"longest_rest_days"`
}

type RewindPhoto struct {
	ID           string `json:"id"`
	ActivityID   int64  `json:"activity_id"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Caption      string `json:"caption,omitempty"`
}

type RewindReport struct {
	Year              int                   `json:"year"` // 0 = all-time (range from first to last activity)
	RangeStart        string                `json:"range_start"`
	RangeEnd          string                `json:"range_end"`
	TotalDays         int                   `json:"total_days"`
	ActiveDays        int                   `json:"active_days"`
	RestDays          int                   `json:"rest_days"`
	Totals            RewindTotals          `json:"totals"`
	Months            []RewindMonth         `json:"months,omitempty"` // year-only
	MovingTimeBySport []RewindSportTime     `json:"moving_time_by_sport"`
	StartTimesByHour  []RewindHourCount     `json:"start_times_by_hour"`
	Locations         []RewindLocationPoint `json:"locations"`
	Streaks           RewindStreaks         `json:"streaks"`
	RandomPhoto       *RewindPhoto          `json:"random_photo,omitempty"`
	Biggest           RewindBiggest         `json:"biggest"`
}

func (r *StatsRepository) ListRewindYears(ctx context.Context, athleteID int64) ([]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT CAST(strftime('%Y', start_date_local) AS INTEGER) AS y
		FROM activities
		WHERE athlete_id = ?
		ORDER BY y DESC
	`, athleteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var years []int
	for rows.Next() {
		var y int
		if err := rows.Scan(&y); err != nil {
			return nil, err
		}
		years = append(years, y)
	}
	return years, rows.Err()
}

func (r *StatsRepository) GetRewind(ctx context.Context, athleteID int64, year int) (*RewindReport, error) {
	var start time.Time
	var end time.Time
	if year > 0 {
		start = time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		// All-time: from first to last activity.
		var minS, maxS sql.NullString
		if err := r.db.QueryRowContext(ctx, `
			SELECT MIN(date(start_date_local)), MAX(date(start_date_local))
			FROM activities
			WHERE athlete_id = ?
		`, athleteID).Scan(&minS, &maxS); err != nil {
			return nil, err
		}
		if !minS.Valid || !maxS.Valid {
			now := time.Now().UTC()
			start, end = now, now
		} else {
			minT, _ := time.Parse("2006-01-02", minS.String)
			maxT, _ := time.Parse("2006-01-02", maxS.String)
			start = time.Date(minT.Year(), minT.Month(), minT.Day(), 0, 0, 0, 0, time.UTC)
			// end is exclusive
			end = time.Date(maxT.Year(), maxT.Month(), maxT.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, 1)
		}
	}

	// Totals.
	var totals RewindTotals
	var dist, elev, commute sql.NullFloat64
	var moving, kudos sql.NullInt64
	var activities int
	if err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS activities,
			SUM(distance) AS distance_m,
			SUM(total_elevation_gain) AS elevation_m,
			SUM(moving_time) AS moving_time_s,
			SUM(kudos_count) AS kudos,
			SUM(CASE WHEN commute AND sport_type LIKE '%Ride%' THEN distance ELSE 0 END) AS commute_distance_m
		FROM activities
		WHERE athlete_id = ? AND start_date_local >= ? AND start_date_local < ?
	`, athleteID, start, end).Scan(&activities, &dist, &elev, &moving, &kudos, &commute); err != nil {
		return nil, err
	}
	totals.Activities = activities
	if dist.Valid {
		totals.DistanceM = dist.Float64
	}
	if elev.Valid {
		totals.ElevationM = elev.Float64
	}
	if moving.Valid {
		totals.MovingTimeS = int(moving.Int64)
	}
	if kudos.Valid {
		totals.Kudos = int(kudos.Int64)
	}
	if commute.Valid {
		totals.CommuteDistM = commute.Float64
	}
	// Very rough estimate: average passenger vehicle emissions per km (kg CO2/km).
	const carKgPerKm = 0.192
	totals.CarbonSavedKg = round2((totals.CommuteDistM / 1000.0) * carKgPerKm)

	// Active days within range.
	var activeDays int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT date(start_date_local))
		FROM activities
		WHERE athlete_id = ? AND start_date_local >= ? AND start_date_local < ?
	`, athleteID, start, end).Scan(&activeDays); err != nil {
		return nil, err
	}

	totalDays := int(math.Round(end.Sub(start).Hours()/24.0 + 0.0))
	if totalDays < 0 {
		totalDays = 0
	}
	restDays := totalDays - activeDays
	if restDays < 0 {
		restDays = 0
	}

	report := &RewindReport{
		Year:       year,
		RangeStart: start.Format("2006-01-02"),
		RangeEnd:   end.AddDate(0, 0, -1).Format("2006-01-02"),
		TotalDays:  totalDays,
		ActiveDays: activeDays,
		RestDays:   restDays,
		Totals:     totals,
	}
	if totals.Activities == 0 {
		report.RangeEnd = report.RangeStart
		report.TotalDays = 0
		report.ActiveDays = 0
		report.RestDays = 0
	}

	// Months (year-only, with zero-fill).
	if year > 0 {
		months := make([]RewindMonth, 12)
		for m := 1; m <= 12; m++ {
			months[m-1] = RewindMonth{
				Month: time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC).Format("2006-01"),
			}
		}

		rows, err := r.db.QueryContext(ctx, `
			SELECT
				CAST(strftime('%m', start_date_local) AS INTEGER) AS m,
				COUNT(*) AS activities,
				SUM(distance) AS distance_m,
				SUM(total_elevation_gain) AS elevation_m
			FROM activities
			WHERE athlete_id = ? AND start_date_local >= ? AND start_date_local < ?
			GROUP BY m
			ORDER BY m ASC
		`, athleteID, start, end)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var m int
			var a int
			var d, e sql.NullFloat64
			if err := rows.Scan(&m, &a, &d, &e); err != nil {
				_ = rows.Close()
				return nil, err
			}
			if m >= 1 && m <= 12 {
				months[m-1].Activities = a
				if d.Valid {
					months[m-1].DistanceM = d.Float64
				}
				if e.Valid {
					months[m-1].ElevationM = e.Float64
				}
			}
		}
		_ = rows.Close()

		// PRs by month (requires best_efforts; best-effort if missing).
		prRows, err := r.db.QueryContext(ctx, `
			WITH w AS (
				SELECT
					be.distance_type,
					a.start_date_local AS dt,
					be.elapsed_time,
					MIN(be.elapsed_time) OVER (
						PARTITION BY be.distance_type
						ORDER BY a.start_date_local ASC
						ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
					) AS best_so_far
				FROM best_efforts be
				JOIN activities a
					ON a.id = be.activity_id AND a.athlete_id = be.athlete_id
				WHERE be.athlete_id = ? AND a.start_date_local < ?
			),
			m AS (
				SELECT
					dt,
					best_so_far,
					LAG(best_so_far) OVER (PARTITION BY distance_type ORDER BY dt ASC) AS prev_best
				FROM w
			)
			SELECT
				CAST(strftime('%m', dt) AS INTEGER) AS mon,
				COUNT(*) AS prs
			FROM m
			WHERE dt >= ? AND (prev_best IS NULL OR best_so_far < prev_best)
			GROUP BY mon
			ORDER BY mon ASC
		`, athleteID, end, start)
		if err == nil {
			for prRows.Next() {
				var mon int
				var prs int
				if err := prRows.Scan(&mon, &prs); err != nil {
					continue
				}
				if mon >= 1 && mon <= 12 {
					months[mon-1].PRs = prs
				}
			}
			_ = prRows.Close()
		}

		report.Months = months
	}

	// Moving time by sport type.
	mtRows, err := r.db.QueryContext(ctx, `
		SELECT sport_type, SUM(moving_time) AS seconds
		FROM activities
		WHERE athlete_id = ? AND start_date_local >= ? AND start_date_local < ?
		GROUP BY sport_type
		ORDER BY seconds DESC
	`, athleteID, start, end)
	if err != nil {
		return nil, err
	}
	for mtRows.Next() {
		var st string
		var secs sql.NullInt64
		if err := mtRows.Scan(&st, &secs); err != nil {
			_ = mtRows.Close()
			return nil, err
		}
		if secs.Valid {
			report.MovingTimeBySport = append(report.MovingTimeBySport, RewindSportTime{SportType: st, MovingTimeS: int(secs.Int64)})
		}
	}
	_ = mtRows.Close()

	// Start times by hour (0..23).
	hours := make([]RewindHourCount, 24)
	for h := 0; h < 24; h++ {
		hours[h] = RewindHourCount{Hour: h, Count: 0}
	}
	hrRows, err := r.db.QueryContext(ctx, `
		SELECT CAST(strftime('%H', start_date_local) AS INTEGER) AS h, COUNT(*) AS c
		FROM activities
		WHERE athlete_id = ? AND start_date_local >= ? AND start_date_local < ?
		GROUP BY h
	`, athleteID, start, end)
	if err != nil {
		return nil, err
	}
	for hrRows.Next() {
		var h, c int
		if err := hrRows.Scan(&h, &c); err != nil {
			_ = hrRows.Close()
			return nil, err
		}
		if h >= 0 && h <= 23 {
			hours[h].Count = c
		}
	}
	_ = hrRows.Close()
	report.StartTimesByHour = hours

	// Locations (rounded buckets to reduce volume).
	locRows, err := r.db.QueryContext(ctx, `
		SELECT
			ROUND(start_lat, 2) AS lat,
			ROUND(start_lng, 2) AS lng,
			COUNT(*) AS c
		FROM activities
		WHERE athlete_id = ?
			AND start_lat IS NOT NULL
			AND start_lng IS NOT NULL
			AND start_date_local >= ? AND start_date_local < ?
		GROUP BY lat, lng
		ORDER BY c DESC
		LIMIT 2000
	`, athleteID, start, end)
	if err != nil {
		return nil, err
	}
	for locRows.Next() {
		var lat, lng float64
		var c int
		if err := locRows.Scan(&lat, &lng, &c); err != nil {
			_ = locRows.Close()
			return nil, err
		}
		if c > 0 {
			report.Locations = append(report.Locations, RewindLocationPoint{Lat: lat, Lng: lng, Count: c})
		}
	}
	_ = locRows.Close()

	// Streaks.
	report.Streaks = computeRewindStreaks(ctx, r.db, athleteID, start, end)

	// Biggest activities.
	report.Biggest = RewindBiggest{
		LongestDistance: queryBiggest(ctx, r.db, athleteID, start, end, "distance"),
		MostElevation:   queryBiggest(ctx, r.db, athleteID, start, end, "total_elevation_gain"),
		LongestDuration: queryBiggest(ctx, r.db, athleteID, start, end, "moving_time"),
	}

	// Random photo (best-effort; photos may not be imported).
	var photo RewindPhoto
	err = r.db.QueryRowContext(ctx, `
		SELECT p.id, p.activity_id, p.url, p.thumbnail_url, p.caption
		FROM photos p
		JOIN activities a ON a.id = p.activity_id AND a.athlete_id = ?
		WHERE a.start_date_local >= ? AND a.start_date_local < ?
		ORDER BY random()
		LIMIT 1
	`, athleteID, start, end).Scan(&photo.ID, &photo.ActivityID, &photo.URL, &photo.ThumbnailURL, &photo.Caption)
	if err == nil && photo.ID != "" && photo.URL != "" {
		report.RandomPhoto = &photo
	}

	return report, nil
}

func computeRewindStreaks(ctx context.Context, db *DB, athleteID int64, start, end time.Time) RewindStreaks {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT date(start_date_local) AS day
		FROM activities
		WHERE athlete_id = ? AND start_date_local >= ? AND start_date_local < ?
		ORDER BY day ASC
	`, athleteID, start, end)
	if err != nil {
		return RewindStreaks{}
	}
	defer func() { _ = rows.Close() }()

	active := map[string]bool{}
	for rows.Next() {
		var day SQLiteTime
		if err := rows.Scan(&day); err != nil {
			continue
		}
		active[day.Format("2006-01-02")] = true
	}

	longestActive := 0
	longestRest := 0
	curActive := 0
	curRest := 0

	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		if active[key] {
			curActive++
			curRest = 0
			if curActive > longestActive {
				longestActive = curActive
			}
		} else {
			curRest++
			curActive = 0
			if curRest > longestRest {
				longestRest = curRest
			}
		}
	}

	return RewindStreaks{LongestActiveDays: longestActive, LongestRestDays: longestRest}
}

func queryBiggest(ctx context.Context, db *DB, athleteID int64, start, end time.Time, metric string) *RewindBiggestActivity {
	orderBy := metric
	valueField := metric
	if metric == "moving_time" {
		valueField = "moving_time"
	}
	row := db.QueryRowContext(ctx, `
		SELECT id, name, sport_type, start_date_local, `+valueField+`
		FROM activities
		WHERE athlete_id = ? AND start_date_local >= ? AND start_date_local < ?
		ORDER BY `+orderBy+` DESC
		LIMIT 1
	`, athleteID, start, end)

	var a RewindBiggestActivity
	var startLocal SQLiteTime
	var v sql.NullFloat64
	if metric == "moving_time" {
		var vv sql.NullInt64
		if err := row.Scan(&a.ActivityID, &a.Name, &a.SportType, &startLocal, &vv); err != nil {
			return nil
		}
		a.StartDateLocal = startLocal.Format(time.RFC3339)
		if vv.Valid {
			a.Value = float64(vv.Int64)
		}
		return &a
	}
	if err := row.Scan(&a.ActivityID, &a.Name, &a.SportType, &startLocal, &v); err != nil {
		return nil
	}
	a.StartDateLocal = startLocal.Format(time.RFC3339)
	if v.Valid {
		a.Value = v.Float64
	}
	return &a
}
