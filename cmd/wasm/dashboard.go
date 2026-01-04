//go:build js && wasm

package main

import (
	"fmt"
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
//
//wasm:export
var getDashboardConfig = wrapWasmAthlete("getDashboardConfig", func(wc *WasmContext) interface{} {
	cfg, err := wc.Registry.DashboardConfig().Get(wc.Ctx, wc.AthleteID)
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(cfg)
})

// updateDashboardConfig updates the dashboard widget configuration
// Called from JS: goStorage.updateDashboardConfig(configJSON)
//
//wasm:export
var updateDashboardConfig = wrapWasmAthlete("updateDashboardConfig", func(wc *WasmContext) interface{} {
	var cfg storage.DashboardConfig
	if err := wc.ArgJSON(0, &cfg); err != nil {
		return errorJSON(fmt.Errorf("parsing config: %w", err))
	}

	if err := wc.Registry.DashboardConfig().Upsert(wc.Ctx, wc.AthleteID, cfg); err != nil {
		return errorJSON(err)
	}
	return dataJSON(cfg)
})

// ============================================================================
// Dashboard Stats
// ============================================================================

//wasm:category Dashboard - Stats

// getDashboardStats returns aggregated statistics
// Called from JS: goStorage.getDashboardStats()
//
//wasm:export
var getDashboardStats = wrapWasmAthlete("getDashboardStats", func(wc *WasmContext) interface{} {
	dashStats, err := wc.Registry.DashboardService.GetStats(wc.Ctx, services.GetDashboardInput{
		AthleteID: wc.AthleteID,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(dashStats)
})

// getWeeklyStats returns statistics for the current week
// Called from JS: goStorage.getWeeklyStats()
//
//wasm:export
var getWeeklyStats = wrapWasmAthlete("getWeeklyStats", func(wc *WasmContext) interface{} {
	weeklyStats, err := wc.Registry.DashboardService.GetWeeklyStats(wc.Ctx, services.GetDashboardInput{
		AthleteID: wc.AthleteID,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(weeklyStats)
})

// getRecentActivities returns recent activities for the dashboard
// Called from JS: goStorage.getRecentActivities(limit)
//
//wasm:export
var getRecentActivities = wrapWasmAthlete("getRecentActivities", func(wc *WasmContext) interface{} {
	limit := 5
	if wc.HasArg(0) {
		limit = wc.ArgInt(0)
	}

	recentList, err := wc.Registry.DashboardService.GetRecentActivities(wc.Ctx, services.GetRecentActivitiesInput{
		AthleteID: wc.AthleteID,
		Limit:     limit,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(recentList)
})

// getSportTypeStats returns statistics grouped by sport type
// Called from JS: goStorage.getSportTypeStats()
//
//wasm:export
var getSportTypeStats = wrapWasmAthlete("getSportTypeStats", func(wc *WasmContext) interface{} {
	sportStats, err := wc.Registry.DashboardService.GetSportTypeStats(wc.Ctx, services.GetDashboardInput{
		AthleteID: wc.AthleteID,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(sportStats)
})

// getMonthlyStats returns monthly statistics
// Called from JS: goStorage.getMonthlyStats(year?)
//
//wasm:export
var getMonthlyStats = wrapWasmAthlete("getMonthlyStats", func(wc *WasmContext) interface{} {
	year := 0 // 0 means all years
	if wc.HasArg(0) {
		y := wc.ArgInt(0)
		if y > 2000 && y < 2100 {
			year = y
		}
	}

	monthlyStats, err := wc.Registry.DashboardService.GetMonthlyStats(wc.Ctx, services.GetMonthlyStatsInput{
		AthleteID: wc.AthleteID,
		Year:      year,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(monthlyStats)
})

// getYearlyStats returns yearly statistics
// Called from JS: goStorage.getYearlyStats()
//
//wasm:export
var getYearlyStats = wrapWasmAthlete("getYearlyStats", func(wc *WasmContext) interface{} {
	yearlyStats, err := wc.Registry.DashboardService.GetYearlyStats(wc.Ctx, services.GetDashboardInput{
		AthleteID: wc.AthleteID,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(yearlyStats)
})

// ============================================================================
// Heatmap
// ============================================================================

//wasm:category Heatmap

// getHeatmapData returns heatmap data
// Called from JS: goStorage.getHeatmapData(filtersJSON)
//
//wasm:export
var getHeatmapData = wrapWasmAthlete("getHeatmapData", func(wc *WasmContext) interface{} {
	// Parse filters
	var filters storage.HeatmapFilters
	if wc.HasArg(0) {
		var req struct {
			SportType   string `json:"sport_type"`
			Year        int    `json:"year"`
			Commute     *bool  `json:"commute"`
			WorkoutType *int   `json:"workout_type"`
			Limit       int    `json:"limit"`
			Offset      int    `json:"offset"`
		}
		if err := wc.ArgJSON(0, &req); err != nil {
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

	heatmapData, err := wc.Registry.Stats().GetHeatmapData(wc.Ctx, wc.AthleteID, filters)
	if err != nil {
		return errorJSON(err)
	}

	// Also get countries for the filter
	countries, err := wc.Registry.Stats().GetHeatmapCountries(wc.Ctx, wc.AthleteID, filters)
	if err != nil {
		// Non-fatal - return data without countries
		countries = nil
	}

	return dataJSON(map[string]interface{}{
		"activities": heatmapData,
		"countries":  countries,
	})
})

// ============================================================================
// Distribution Stats
// ============================================================================

//wasm:category Distribution

// getDaytimeDistribution returns activity counts by time of day
// Called from JS: goStorage.getDaytimeDistribution()
//
//wasm:export
var getDaytimeDistribution = wrapWasmAthlete("getDaytimeDistribution", func(wc *WasmContext) interface{} {
	result, err := wc.Registry.DashboardService.GetDaytimeDistribution(wc.Ctx, services.GetDistributionInput{
		AthleteID: wc.AthleteID,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(result)
})

// getWeekdayDistribution returns activity counts by day of week
// Called from JS: goStorage.getWeekdayDistribution()
//
//wasm:export
var getWeekdayDistribution = wrapWasmAthlete("getWeekdayDistribution", func(wc *WasmContext) interface{} {
	result, err := wc.Registry.DashboardService.GetWeekdayDistribution(wc.Ctx, services.GetDistributionInput{
		AthleteID: wc.AthleteID,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(result)
})

// getExportStats returns export statistics (total count, date range)
// Called from JS: goStorage.getExportStats()
//
//wasm:export
var getExportStats = wrapWasmAthlete("getExportStats", func(wc *WasmContext) interface{} {
	result, err := wc.Registry.DashboardService.GetExportStats(wc.Ctx, services.GetDashboardInput{
		AthleteID: wc.AthleteID,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(result)
})

// ============================================================================
// Calendar
// ============================================================================

//wasm:category Calendar

// getCalendarData retrieves calendar data for a year
// Called from JS: goStorage.getCalendarData(year)
//
//wasm:export
var getCalendarData = wrapWasmAthlete("getCalendarData", func(wc *WasmContext) interface{} {
	year := time.Now().Year()
	if wc.HasArg(0) {
		year = wc.ArgInt(0)
	}

	data, err := wc.Registry.DashboardService.GetCalendarData(wc.Ctx, services.GetCalendarDataInput{
		AthleteID: wc.AthleteID,
		Year:      year,
	})
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(data)
})
