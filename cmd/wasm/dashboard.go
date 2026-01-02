//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Dashboard
// ============================================================================

// getDashboardStats returns aggregated statistics
// Called from JS: goStorage.getDashboardStats()
func getDashboardStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getDashboardStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	dashStats, err := stats.GetDashboardStats(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(dashStats)
}

// getWeeklyStats returns statistics for the current week
// Called from JS: goStorage.getWeeklyStats()
func getWeeklyStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getWeeklyStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	weeklyStats, err := stats.GetWeeklyStats(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(weeklyStats)
}

// getRecentActivities returns recent activities for the dashboard
// Called from JS: goStorage.getRecentActivities(limit)
func getRecentActivities(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getRecentActivities")

	limit := 5
	if len(args) > 0 {
		limit = args[0].Int()
	}

	ctx := context.Background()
	recentList, err := stats.GetRecentActivities(ctx, athleteID, limit)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(recentList)
}

// getSportTypeStats returns statistics grouped by sport type
// Called from JS: goStorage.getSportTypeStats()
func getSportTypeStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSportTypeStats")

	ctx := context.Background()
	sportStats, err := stats.GetStatsBySportType(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(sportStats)
}

// getMonthlyStats returns monthly statistics
// Called from JS: goStorage.getMonthlyStats(year?)
func getMonthlyStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getMonthlyStats")

	year := time.Now().Year()
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		year = args[0].Int()
	}

	ctx := context.Background()
	monthlyStats, err := stats.GetMonthlyStats(ctx, athleteID, year)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(monthlyStats)
}

// getYearlyStats returns yearly statistics
// Called from JS: goStorage.getYearlyStats()
func getYearlyStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getYearlyStats")

	ctx := context.Background()
	yearlyStats, err := stats.GetYearlyStats(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(yearlyStats)
}

// ============================================================================
// Heatmap
// ============================================================================

// getHeatmapData returns heatmap data
// Called from JS: goStorage.getHeatmapData(filtersJSON)
func getHeatmapData(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getHeatmapData")

	// Parse filters
	var filters storage.HeatmapFilters
	if len(args) > 0 {
		filtersJSON := args[0].String()
		var req struct {
			SportType string `json:"sport_type"`
			Year      int    `json:"year"`
			Commute   *bool  `json:"commute"`
		}
		if err := json.Unmarshal([]byte(filtersJSON), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		if req.SportType != "" {
			filters.SportTypes = []string{req.SportType}
		}
		// Convert year to date range
		if req.Year > 0 {
			startOfYear := time.Date(req.Year, 1, 1, 0, 0, 0, 0, time.UTC)
			endOfYear := time.Date(req.Year+1, 1, 1, 0, 0, 0, 0, time.UTC)
			filters.StartAfter = &startOfYear
			filters.StartBefore = &endOfYear
		}
		filters.Commute = req.Commute
	}

	ctx := context.Background()
	heatmapData, err := stats.GetHeatmapData(ctx, athleteID, filters)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(heatmapData)
}

// ============================================================================
// Distribution Stats
// ============================================================================

// getDaytimeDistribution returns activity counts by time of day
// Called from JS: goStorage.getDaytimeDistribution()
func getDaytimeDistribution(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getDaytimeDistribution")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()

	// Query activities grouped by hour, then bucket into time-of-day categories
	query := `
		SELECT
			CASE
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 5 AND 11 THEN 'Morning'
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 12 AND 16 THEN 'Afternoon'
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 17 AND 21 THEN 'Evening'
				ELSE 'Night'
			END AS label,
			COUNT(*) AS count
		FROM activities
		WHERE athlete_id = ?
		GROUP BY label
		ORDER BY
			CASE label
				WHEN 'Morning' THEN 1
				WHEN 'Afternoon' THEN 2
				WHEN 'Evening' THEN 3
				WHEN 'Night' THEN 4
			END
	`

	rows, err := db.Conn().QueryContext(ctx, query, athleteID)
	if err != nil {
		return errorJSON(err)
	}
	defer func() { _ = rows.Close() }()

	var result []map[string]interface{}
	for rows.Next() {
		var label string
		var count int
		if err := rows.Scan(&label, &count); err != nil {
			return errorJSON(err)
		}
		result = append(result, map[string]interface{}{
			"label": label,
			"count": count,
		})
	}
	if err := rows.Err(); err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// getWeekdayDistribution returns activity counts by day of week
// Called from JS: goStorage.getWeekdayDistribution()
func getWeekdayDistribution(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getWeekdayDistribution")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()

	// Query activities grouped by weekday (0=Sunday, 6=Saturday)
	query := `
		SELECT CAST(strftime('%w', start_date_local) AS INTEGER) AS weekday, COUNT(*) AS count
		FROM activities
		WHERE athlete_id = ?
		GROUP BY weekday
		ORDER BY weekday
	`

	rows, err := db.Conn().QueryContext(ctx, query, athleteID)
	if err != nil {
		return errorJSON(err)
	}
	defer func() { _ = rows.Close() }()

	// Map weekday numbers to names
	weekdayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}

	// Initialize all days with 0 count
	countByDay := make(map[int]int)
	for rows.Next() {
		var weekday, count int
		if err := rows.Scan(&weekday, &count); err != nil {
			return errorJSON(err)
		}
		countByDay[weekday] = count
	}
	if err := rows.Err(); err != nil {
		return errorJSON(err)
	}

	// Build result with all days (including 0 counts)
	var result []map[string]interface{}
	for i := 0; i < 7; i++ {
		result = append(result, map[string]interface{}{
			"label": weekdayNames[i],
			"count": countByDay[i],
		})
	}

	return dataJSON(result)
}

// getExportStats returns export statistics (total count, date range)
// Called from JS: goStorage.getExportStats()
func getExportStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getExportStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()

	// Get total count and date range
	query := `
		SELECT
			COUNT(*) as total,
			MIN(start_date_local) as first_activity,
			MAX(start_date_local) as last_activity
		FROM activities
		WHERE athlete_id = ?
	`

	var total int
	var firstActivity, lastActivity *string

	row := db.Conn().QueryRowContext(ctx, query, athleteID)
	if err := row.Scan(&total, &firstActivity, &lastActivity); err != nil {
		return errorJSON(err)
	}

	result := map[string]interface{}{
		"total_activities": total,
		"first_activity":   firstActivity,
		"last_activity":    lastActivity,
	}

	return dataJSON(result)
}

// ============================================================================
// Calendar
// ============================================================================

// getCalendarData retrieves calendar data for a year
// Called from JS: goStorage.getCalendarData(year)
func getCalendarData(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getCalendarData")

	year := time.Now().Year()
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		year = args[0].Int()
	}

	ctx := context.Background()
	data, err := stats.GetCalendarData(ctx, athleteID, year)
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}
