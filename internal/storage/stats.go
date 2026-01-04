package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/melonamin/quantlete/internal/geo"
)

// DashboardStats represents aggregated statistics for the dashboard.
type DashboardStats struct {
	TotalActivities    int     `json:"total_activities"`
	TotalDistance      float64 `json:"total_distance"`       // meters
	TotalMovingTime    int     `json:"total_moving_time"`    // seconds
	TotalElevationGain float64 `json:"total_elevation_gain"` // meters
	TotalCalories      float64 `json:"total_calories"`

	// This year
	YearActivities    int     `json:"year_activities"`
	YearDistance      float64 `json:"year_distance"`
	YearMovingTime    int     `json:"year_moving_time"`
	YearElevationGain float64 `json:"year_elevation_gain"`

	// This month
	MonthActivities    int     `json:"month_activities"`
	MonthDistance      float64 `json:"month_distance"`
	MonthMovingTime    int     `json:"month_moving_time"`
	MonthElevationGain float64 `json:"month_elevation_gain"`
}

// WeeklyStat represents statistics for a single sport type in the current week.
type WeeklyStat struct {
	SportType      string  `json:"sport_type"`
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// RecentActivity represents a simplified activity for the dashboard.
type RecentActivity struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	SportType       string     `json:"sport_type"`
	StartDate       SQLiteTime `json:"start_date"`
	Distance        float64    `json:"distance"`
	MovingTime      int        `json:"moving_time"`
	ElevationGain   float64    `json:"elevation_gain"`
	SummaryPolyline string     `json:"summary_polyline,omitempty"`
}

// StatsRepository handles statistics queries.
type StatsRepository struct {
	db *DB
}

// NewStatsRepository creates a new stats repository.
func NewStatsRepository(db *DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// GetDashboardStats returns aggregated statistics for the dashboard.
func (r *StatsRepository) GetDashboardStats(ctx context.Context, athleteID int64) (*DashboardStats, error) {
	now := time.Now()
	yearStart := fmt.Sprintf("%d-01-01", now.Year())
	monthStart := fmt.Sprintf("%d-%02d-01", now.Year(), int(now.Month()))

	stats := &DashboardStats{}

	// Total stats
	err := r.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(distance), 0),
			COALESCE(SUM(moving_time), 0),
			COALESCE(SUM(total_elevation_gain), 0),
			COALESCE(SUM(calories), 0)
		FROM activities
		WHERE athlete_id = ?
	`, athleteID).Scan(
		&stats.TotalActivities,
		&stats.TotalDistance,
		&stats.TotalMovingTime,
		&stats.TotalElevationGain,
		&stats.TotalCalories,
	)
	if err != nil {
		return nil, err
	}

	// Year stats (use start_date_local for correct timezone grouping)
	err = r.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(distance), 0),
			COALESCE(SUM(moving_time), 0),
			COALESCE(SUM(total_elevation_gain), 0)
		FROM activities
		WHERE athlete_id = ? AND DATE(start_date_local) >= ?
	`, athleteID, yearStart).Scan(
		&stats.YearActivities,
		&stats.YearDistance,
		&stats.YearMovingTime,
		&stats.YearElevationGain,
	)
	if err != nil {
		return nil, err
	}

	// Month stats (use start_date_local for correct timezone grouping)
	err = r.db.QueryRow(`
		SELECT
			COUNT(*),
			COALESCE(SUM(distance), 0),
			COALESCE(SUM(moving_time), 0),
			COALESCE(SUM(total_elevation_gain), 0)
		FROM activities
		WHERE athlete_id = ? AND DATE(start_date_local) >= ?
	`, athleteID, monthStart).Scan(
		&stats.MonthActivities,
		&stats.MonthDistance,
		&stats.MonthMovingTime,
		&stats.MonthElevationGain,
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetWeeklyStats returns statistics grouped by sport type for the current week.
func (r *StatsRepository) GetWeeklyStats(ctx context.Context, athleteID int64) ([]WeeklyStat, error) {
	now := time.Now()
	// Get start of week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}
	weekStart := fmt.Sprintf("%d-%02d-%02d", now.Year(), int(now.Month()), now.Day()-weekday+1)

	rows, err := r.db.Query(`
		SELECT
			sport_type,
			COUNT(*) as activity_count,
			COALESCE(SUM(distance), 0) as total_distance,
			COALESCE(SUM(moving_time), 0) as total_time,
			COALESCE(SUM(total_elevation_gain), 0) as total_elevation
		FROM activities
		WHERE athlete_id = ? AND DATE(start_date_local) >= ?
		GROUP BY sport_type
		ORDER BY total_distance DESC
	`, athleteID, weekStart)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []WeeklyStat
	for rows.Next() {
		var s WeeklyStat
		if err := rows.Scan(&s.SportType, &s.ActivityCount, &s.TotalDistance, &s.TotalTime, &s.TotalElevation); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, rows.Err()
}

// GetRecentActivities returns the most recent activities.
func (r *StatsRepository) GetRecentActivities(ctx context.Context, athleteID int64, limit int) ([]RecentActivity, error) {
	rows, err := r.db.Query(`
		SELECT
			id, name, sport_type, start_date,
			distance, moving_time, total_elevation_gain,
			COALESCE(summary_polyline, '')
		FROM activities
		WHERE athlete_id = ?
		ORDER BY start_date DESC
		LIMIT ?
	`, athleteID, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var activities []RecentActivity
	for rows.Next() {
		var a RecentActivity
		if err := rows.Scan(
			&a.ID, &a.Name, &a.SportType, &a.StartDate,
			&a.Distance, &a.MovingTime, &a.ElevationGain,
			&a.SummaryPolyline,
		); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}

	return activities, rows.Err()
}

// SportTypeStat represents statistics for a single sport type.
type SportTypeStat struct {
	SportType      string  `json:"sport_type"`
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// GetStatsBySportType returns statistics grouped by sport type.
func (r *StatsRepository) GetStatsBySportType(ctx context.Context, athleteID int64) ([]SportTypeStat, error) {
	q := NewQueries(r.db.Conn())
	rows, err := q.GetStatsBySportType(ctx, athleteID)
	if err != nil {
		return nil, err
	}
	stats := make([]SportTypeStat, len(rows))
	for i, row := range rows {
		stats[i] = SportTypeStat(row)
	}
	return stats, nil
}

// MonthlyStat represents statistics for a single month.
type MonthlyStat struct {
	Month          string  `json:"month"` // YYYY-MM format
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// GetMonthlyStats returns statistics grouped by month for a given year (or all years if year is 0).
func (r *StatsRepository) GetMonthlyStats(ctx context.Context, athleteID int64, year int) ([]MonthlyStat, error) {
	query := `
		SELECT
			strftime('%Y-%m', start_date_local) as month,
			COUNT(*) as activity_count,
			COALESCE(SUM(distance), 0) as total_distance,
			COALESCE(SUM(moving_time), 0) as total_time,
			COALESCE(SUM(total_elevation_gain), 0) as total_elevation
		FROM activities
		WHERE athlete_id = ?
	`
	args := []interface{}{athleteID}

	if year > 0 {
		query += ` AND strftime('%Y', start_date_local) = ?`
		args = append(args, fmt.Sprintf("%d", year))
	}

	query += `
		GROUP BY month
		ORDER BY month ASC
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var monthly []MonthlyStat
	for rows.Next() {
		var s MonthlyStat
		if err := rows.Scan(&s.Month, &s.ActivityCount, &s.TotalDistance, &s.TotalTime, &s.TotalElevation); err != nil {
			return nil, err
		}
		monthly = append(monthly, s)
	}

	return monthly, rows.Err()
}

// CalendarDay represents activity data for a single day.
type CalendarDay struct {
	Date          string  `json:"date"` // YYYY-MM-DD format
	ActivityCount int     `json:"activity_count"`
	TotalDistance float64 `json:"total_distance"`
	TotalTime     int     `json:"total_time"` // moving_time in seconds
	TotalCalories float64 `json:"total_calories"`
}

// CalendarActivity represents an activity summary for the calendar view.
type CalendarActivity struct {
	ID                 int64      `json:"id"`
	Name               string     `json:"name"`
	SportType          string     `json:"sport_type"`
	StartDate          SQLiteTime `json:"start_date"`
	Distance           float64    `json:"distance"`
	MovingTime         int        `json:"moving_time"`
	TotalElevationGain float64    `json:"total_elevation_gain"`
}

// GetCalendarData returns daily activity counts for a given year.
func (r *StatsRepository) GetCalendarData(ctx context.Context, athleteID int64, year int) ([]CalendarDay, error) {
	rows, err := r.db.Query(`
		SELECT
			strftime('%Y-%m-%d', start_date_local) as date,
			COUNT(*) as activity_count,
			COALESCE(SUM(distance), 0) as total_distance,
			COALESCE(SUM(moving_time), 0) as total_time,
			COALESCE(SUM(calories), 0) as total_calories
		FROM activities
		WHERE athlete_id = ? AND strftime('%Y', start_date_local) = ?
		GROUP BY date
		ORDER BY date ASC
	`, athleteID, fmt.Sprintf("%d", year))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var days []CalendarDay
	for rows.Next() {
		var d CalendarDay
		if err := rows.Scan(&d.Date, &d.ActivityCount, &d.TotalDistance, &d.TotalTime, &d.TotalCalories); err != nil {
			return nil, err
		}
		days = append(days, d)
	}

	return days, rows.Err()
}

// GetCalendarActivities returns activities for a specific month.
func (r *StatsRepository) GetCalendarActivities(ctx context.Context, athleteID int64, year, month int) ([]CalendarActivity, error) {
	rows, err := r.db.Query(`
		SELECT id, name, sport_type, start_date_local, distance, moving_time, COALESCE(total_elevation_gain, 0)
		FROM activities
		WHERE athlete_id = ?
		  AND CAST(strftime('%Y', start_date_local) AS INTEGER) = ?
		  AND CAST(strftime('%m', start_date_local) AS INTEGER) = ?
		ORDER BY start_date_local ASC
	`, athleteID, year, month)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var activities []CalendarActivity
	for rows.Next() {
		var a CalendarActivity
		if err := rows.Scan(&a.ID, &a.Name, &a.SportType, &a.StartDate, &a.Distance, &a.MovingTime, &a.TotalElevationGain); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}

	return activities, rows.Err()
}

type CalendarMonthSummary struct {
	Year                int     `json:"year"`
	Month               int     `json:"month"`
	ActivityCount       int     `json:"activity_count"`
	TotalDistance       float64 `json:"total_distance"`
	TotalElevationGain  float64 `json:"total_elevation_gain"`
	TotalMovingTime     int     `json:"total_moving_time"`
	TotalCalories       float64 `json:"total_calories"`
	WorkoutCount        int     `json:"workout_count"`
	ChallengesCompleted int     `json:"challenges_completed"`
}

func (r *StatsRepository) GetCalendarMonthSummary(ctx context.Context, athleteID int64, year, month int) (*CalendarMonthSummary, error) {
	var s CalendarMonthSummary
	s.Year = year
	s.Month = month

	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) AS activity_count,
			COALESCE(SUM(distance), 0) AS total_distance,
			COALESCE(SUM(total_elevation_gain), 0) AS total_elevation_gain,
			COALESCE(SUM(moving_time), 0) AS total_moving_time,
			COALESCE(SUM(calories), 0) AS total_calories,
			COALESCE(SUM(CASE WHEN workout_type IS NOT NULL AND workout_type != 0 THEN 1 ELSE 0 END), 0) AS workout_count
		FROM activities
		WHERE athlete_id = ?
		  AND CAST(strftime('%Y', start_date_local) AS INTEGER) = ?
		  AND CAST(strftime('%m', start_date_local) AS INTEGER) = ?
	`, athleteID, year, month).Scan(
		&s.ActivityCount,
		&s.TotalDistance,
		&s.TotalElevationGain,
		&s.TotalMovingTime,
		&s.TotalCalories,
		&s.WorkoutCount,
	)
	if err != nil {
		return nil, err
	}

	monthKey := fmt.Sprintf("%04d-%02d", year, month)
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM challenges
		WHERE athlete_id = ? AND month = ?
	`, athleteID, monthKey).Scan(&s.ChallengesCompleted)

	return &s, nil
}

// YearStat represents statistics for a single year.
type YearStat struct {
	Year           int     `json:"year"`
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// HeatmapActivity represents an activity for the heatmap visualization.
// IMPORTANT: Field order must match the SELECT column order in GetHeatmapData.
// When modifying fields, update both the struct and the SQL query together.
type HeatmapActivity struct {
	ID              int64   `json:"id"`               // Column 1
	Name            string  `json:"name"`             // Column 2
	SportType       string  `json:"sport_type"`       // Column 3
	StartDate       string  `json:"start_date"`       // Column 4
	Distance        float64 `json:"distance"`         // Column 5
	SummaryPolyline string  `json:"summary_polyline"` // Column 6
	StartLat        float64 `json:"start_lat"`        // Column 7
	StartLng        float64 `json:"start_lng"`        // Column 8
}

// HeatmapFilters contains filters for heatmap queries.
type HeatmapFilters struct {
	SportTypes  []string
	StartAfter  *time.Time
	StartBefore *time.Time
	Commute     *bool
	WorkoutType *int
	Limit       int // 0 means no limit
	Offset      int
}

// GetYearlyStats returns statistics grouped by year.
func (r *StatsRepository) GetYearlyStats(ctx context.Context, athleteID int64) ([]YearStat, error) {
	q := NewQueries(r.db.Conn())
	rows, err := q.GetYearlyStats(ctx, athleteID)
	if err != nil {
		return nil, err
	}
	yearly := make([]YearStat, len(rows))
	for i, row := range rows {
		year, err := strconv.Atoi(row.Year)
		if err != nil {
			return nil, fmt.Errorf("invalid year %q: %w", row.Year, err)
		}
		yearly[i] = YearStat{
			Year:           year,
			ActivityCount:  row.ActivityCount,
			TotalDistance:  row.TotalDistance,
			TotalTime:      row.TotalTime,
			TotalElevation: row.TotalElevation,
		}
	}
	return yearly, nil
}

// GetHeatmapData returns activities with polylines for heatmap visualization.
// NOTE: The SELECT column order must match HeatmapActivity field order for rows.Scan.
func (r *StatsRepository) GetHeatmapData(ctx context.Context, athleteID int64, filters HeatmapFilters) ([]HeatmapActivity, error) {
	// Column order: id, name, sport_type, start_date, distance, summary_polyline, start_lat, start_lng
	// Must match HeatmapActivity struct field order for rows.Scan to work correctly.
	query := `
		SELECT id, name, sport_type, start_date, COALESCE(distance, 0), summary_polyline, COALESCE(start_lat, 0), COALESCE(start_lng, 0)
		FROM activities
		WHERE athlete_id = ? AND summary_polyline IS NOT NULL AND summary_polyline != ''
	`
	args := []interface{}{athleteID}

	if len(filters.SportTypes) > 0 {
		placeholders := make([]string, len(filters.SportTypes))
		for i, st := range filters.SportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + joinStrings(placeholders, ",") + ")"
	}

	if filters.StartAfter != nil {
		query += " AND start_date >= ?"
		args = append(args, *filters.StartAfter)
	}

	if filters.StartBefore != nil {
		query += " AND start_date <= ?"
		args = append(args, *filters.StartBefore)
	}

	if filters.Commute != nil {
		query += " AND commute = ?"
		args = append(args, *filters.Commute)
	}

	if filters.WorkoutType != nil {
		query += " AND workout_type = ?"
		args = append(args, *filters.WorkoutType)
	}

	query += " ORDER BY start_date DESC"

	if filters.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filters.Limit)
		if filters.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filters.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var activities []HeatmapActivity
	for rows.Next() {
		var a HeatmapActivity
		if err := rows.Scan(&a.ID, &a.Name, &a.SportType, &a.StartDate, &a.Distance, &a.SummaryPolyline, &a.StartLat, &a.StartLng); err != nil {
			return nil, err
		}
		activities = append(activities, a)
	}

	return activities, rows.Err()
}

// CountHeatmapActivities returns the total count of activities matching the filters.
func (r *StatsRepository) CountHeatmapActivities(ctx context.Context, athleteID int64, filters HeatmapFilters) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM activities
		WHERE athlete_id = ? AND summary_polyline IS NOT NULL AND summary_polyline != ''
	`
	args := []interface{}{athleteID}

	if len(filters.SportTypes) > 0 {
		placeholders := make([]string, len(filters.SportTypes))
		for i, st := range filters.SportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + joinStrings(placeholders, ",") + ")"
	}

	if filters.StartAfter != nil {
		query += " AND start_date >= ?"
		args = append(args, *filters.StartAfter)
	}

	if filters.StartBefore != nil {
		query += " AND start_date <= ?"
		args = append(args, *filters.StartBefore)
	}

	if filters.Commute != nil {
		query += " AND commute = ?"
		args = append(args, *filters.Commute)
	}

	if filters.WorkoutType != nil {
		query += " AND workout_type = ?"
		args = append(args, *filters.WorkoutType)
	}

	var count int
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

type HeatmapCountryStat struct {
	Country string `json:"country"`
	ISO2    string `json:"iso2,omitempty"`
	Count   int    `json:"count"`
}

func (r *StatsRepository) GetHeatmapCountries(ctx context.Context, athleteID int64, filters HeatmapFilters) ([]HeatmapCountryStat, error) {
	query := `
		SELECT COALESCE(location_country, '') AS country, COUNT(*) AS count
		FROM activities
		WHERE athlete_id = ? AND COALESCE(location_country, '') != ''
	`
	args := []any{athleteID}

	if len(filters.SportTypes) > 0 {
		placeholders := make([]string, len(filters.SportTypes))
		for i, st := range filters.SportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + joinStrings(placeholders, ",") + ")"
	}
	if filters.StartAfter != nil {
		query += " AND start_date >= ?"
		args = append(args, *filters.StartAfter)
	}
	if filters.StartBefore != nil {
		query += " AND start_date <= ?"
		args = append(args, *filters.StartBefore)
	}
	if filters.Commute != nil {
		query += " AND commute = ?"
		args = append(args, *filters.Commute)
	}
	if filters.WorkoutType != nil {
		query += " AND workout_type = ?"
		args = append(args, *filters.WorkoutType)
	}

	query += " GROUP BY country ORDER BY count DESC, country ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []HeatmapCountryStat
	for rows.Next() {
		var c HeatmapCountryStat
		if err := rows.Scan(&c.Country, &c.Count); err != nil {
			return nil, err
		}
		if iso2, ok := geo.ISO2FromCountry(c.Country); ok {
			c.ISO2 = iso2
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

// EddingtonResult contains the Eddington number calculation result.
type EddingtonResult struct {
	Number       int             `json:"number"`
	Distribution []EddingtonDay  `json:"distribution"`
	NextSteps    []EddingtonStep `json:"next_steps"`
}

// EddingtonDay represents a day's distance for Eddington calculation.
type EddingtonDay struct {
	Date     string  `json:"date"`
	Distance float64 `json:"distance"` // in km
}

// EddingtonStep shows how many rides needed to reach the next Eddington number.
type EddingtonStep struct {
	Target      int `json:"target"`
	RidesNeeded int `json:"rides_needed"`
}

// GetEddingtonData returns data for Eddington number calculation.
func (r *StatsRepository) GetEddingtonData(ctx context.Context, athleteID int64, sportTypes []string) (*EddingtonResult, error) {
	query := `
		SELECT
			strftime('%Y-%m-%d', start_date_local) as date,
			SUM(distance) / 1000.0 as distance_km
		FROM activities
		WHERE athlete_id = ?
	`
	args := []interface{}{athleteID}

	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + joinStrings(placeholders, ",") + ")"
	}

	query += `
		GROUP BY date
		HAVING distance_km > 0
		ORDER BY distance_km DESC
	`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var days []EddingtonDay
	for rows.Next() {
		var d EddingtonDay
		if err := rows.Scan(&d.Date, &d.Distance); err != nil {
			return nil, err
		}
		days = append(days, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Calculate Eddington number
	eddington := 0
	for i, d := range days {
		// i+1 is the count of days with distance >= d.Distance
		if d.Distance >= float64(i+1) {
			eddington = i + 1
		} else {
			break
		}
	}

	// Calculate next steps (how many rides needed for next 5 numbers)
	var nextSteps []EddingtonStep
	for target := eddington + 1; target <= eddington+5; target++ {
		count := 0
		for _, d := range days {
			if d.Distance >= float64(target) {
				count++
			}
		}
		ridesNeeded := target - count
		if ridesNeeded > 0 {
			nextSteps = append(nextSteps, EddingtonStep{
				Target:      target,
				RidesNeeded: ridesNeeded,
			})
		}
	}

	return &EddingtonResult{
		Number:       eddington,
		Distribution: days,
		NextSteps:    nextSteps,
	}, nil
}
