//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Dashboard Config
// ============================================================================

//wasm:category Dashboard - Config

// getDashboardConfig returns the dashboard widget configuration
// Called from JS: goStorage.getDashboardConfig()
//wasm:export
func getDashboardConfig(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getDashboardConfig")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	cfg, err := bridge.dashboardConfig.Get(ctx, bridge.athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(cfg)
}

// updateDashboardConfig updates the dashboard widget configuration
// Called from JS: goStorage.updateDashboardConfig(configJSON)
//wasm:export
func updateDashboardConfig(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateDashboardConfig")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing config"))
	}

	configJSON := args[0].String()
	var cfg storage.DashboardConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return errorJSON(fmt.Errorf("parsing config: %w", err))
	}

	ctx := context.Background()
	if err := bridge.dashboardConfig.Upsert(ctx, bridge.athleteID, cfg); err != nil {
		return errorJSON(err)
	}

	return dataJSON(cfg)
}

// ============================================================================
// Dashboard Stats
// ============================================================================

//wasm:category Dashboard - Stats

// getDashboardStats returns aggregated statistics
// Called from JS: goStorage.getDashboardStats()
//wasm:export
func getDashboardStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getDashboardStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	dashStats, err := bridge.dashboardService.GetStats(ctx, services.GetDashboardInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(dashStats)
}

// getWeeklyStats returns statistics for the current week
// Called from JS: goStorage.getWeeklyStats()
//wasm:export
func getWeeklyStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getWeeklyStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	weeklyStats, err := bridge.dashboardService.GetWeeklyStats(ctx, services.GetDashboardInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(weeklyStats)
}

// getRecentActivities returns recent activities for the dashboard
// Called from JS: goStorage.getRecentActivities(limit)
//wasm:export
func getRecentActivities(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getRecentActivities")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	limit := 5
	if len(args) > 0 {
		limit = args[0].Int()
	}

	ctx := context.Background()
	recentList, err := bridge.dashboardService.GetRecentActivities(ctx, services.GetRecentActivitiesInput{
		AthleteID: bridge.athleteID,
		Limit:     limit,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(recentList)
}

// getSportTypeStats returns statistics grouped by sport type
// Called from JS: goStorage.getSportTypeStats()
//wasm:export
func getSportTypeStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSportTypeStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	sportStats, err := bridge.dashboardService.GetSportTypeStats(ctx, services.GetDashboardInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(sportStats)
}

// getMonthlyStats returns monthly statistics
// Called from JS: goStorage.getMonthlyStats(year?)
//wasm:export
func getMonthlyStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getMonthlyStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	year := 0 // 0 means all years
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		y := args[0].Int()
		if y > 2000 && y < 2100 {
			year = y
		}
	}

	ctx := context.Background()
	monthlyStats, err := bridge.dashboardService.GetMonthlyStats(ctx, services.GetMonthlyStatsInput{
		AthleteID: bridge.athleteID,
		Year:      year,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(monthlyStats)
}

// getYearlyStats returns yearly statistics
// Called from JS: goStorage.getYearlyStats()
//wasm:export
func getYearlyStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getYearlyStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	yearlyStats, err := bridge.dashboardService.GetYearlyStats(ctx, services.GetDashboardInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(yearlyStats)
}

// ============================================================================
// Heatmap
// ============================================================================

//wasm:category Heatmap

// getHeatmapData returns heatmap data
// Called from JS: goStorage.getHeatmapData(filtersJSON)
//wasm:export
func getHeatmapData(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getHeatmapData")

	// Parse filters
	var filters storage.HeatmapFilters
	if len(args) > 0 {
		filtersJSON := args[0].String()
		var req struct {
			SportType   string `json:"sport_type"`
			Year        int    `json:"year"`
			Commute     *bool  `json:"commute"`
			WorkoutType *int   `json:"workout_type"`
			Limit       int    `json:"limit"`
			Offset      int    `json:"offset"`
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
		filters.WorkoutType = req.WorkoutType
		filters.Limit = req.Limit
		filters.Offset = req.Offset
	}

	ctx := context.Background()
	heatmapData, err := bridge.stats.GetHeatmapData(ctx, bridge.athleteID, filters)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(heatmapData)
}

// ============================================================================
// Distribution Stats
// ============================================================================

//wasm:category Distribution

// getDaytimeDistribution returns activity counts by time of day
// Called from JS: goStorage.getDaytimeDistribution()
//wasm:export
func getDaytimeDistribution(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getDaytimeDistribution")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	result, err := bridge.dashboardService.GetDaytimeDistribution(ctx, services.GetDistributionInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// getWeekdayDistribution returns activity counts by day of week
// Called from JS: goStorage.getWeekdayDistribution()
//wasm:export
func getWeekdayDistribution(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getWeekdayDistribution")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	result, err := bridge.dashboardService.GetWeekdayDistribution(ctx, services.GetDistributionInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// getExportStats returns export statistics (total count, date range)
// Called from JS: goStorage.getExportStats()
//wasm:export
func getExportStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getExportStats")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	result, err := bridge.dashboardService.GetExportStats(ctx, services.GetDashboardInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// ============================================================================
// Calendar
// ============================================================================

//wasm:category Calendar

// getCalendarData retrieves calendar data for a year
// Called from JS: goStorage.getCalendarData(year)
//wasm:export
func getCalendarData(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getCalendarData")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	year := time.Now().Year()
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		year = args[0].Int()
	}

	ctx := context.Background()
	data, err := bridge.dashboardService.GetCalendarData(ctx, services.GetCalendarDataInput{
		AthleteID: bridge.athleteID,
		Year:      year,
	})
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}
