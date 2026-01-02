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
