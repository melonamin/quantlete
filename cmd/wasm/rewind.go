//go:build js && wasm

package main

import (
	"context"
	"syscall/js"

	"github.com/melonamin/quantlete/internal/services"
)

// ============================================================================
// Rewind Functions
// ============================================================================

//wasm:category Rewind

// getRewindYears returns years that have activity data for rewind
// Called from JS: goStorage.getRewindYears()
//wasm:export
func getRewindYears(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getRewindYears")

	ctx := context.Background()
	years, err := bridge.statsService.GetRewindYears(ctx, services.GetRewindYearsInput{
		AthleteID: bridge.athleteID,
	})
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": years,
	})
}

// getRewind returns the rewind report for a specific year
// Called from JS: goStorage.getRewind(year)
//wasm:export
func getRewind(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getRewind")

	year := 0 // 0 = all-time
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		year = args[0].Int()
	}

	ctx := context.Background()
	report, err := bridge.statsService.GetRewind(ctx, services.GetRewindInput{
		AthleteID: bridge.athleteID,
		Year:      year,
	})
	if err != nil {
		return errorJSON(err)
	}

	if report == nil {
		return toJSON(map[string]interface{}{
			"ok":   true,
			"data": nil,
		})
	}

	// Convert to response format
	data := map[string]interface{}{
		"year":        report.Year,
		"range_start": report.RangeStart,
		"range_end":   report.RangeEnd,
		"total_days":  report.TotalDays,
		"active_days": report.ActiveDays,
		"rest_days":   report.RestDays,
		"totals": map[string]interface{}{
			"activities":      report.Totals.Activities,
			"distance_m":      report.Totals.DistanceM,
			"elevation_m":     report.Totals.ElevationM,
			"moving_time_s":   report.Totals.MovingTimeS,
			"kudos":           report.Totals.Kudos,
			"commute_dist_m":  report.Totals.CommuteDistM,
			"carbon_saved_kg": report.Totals.CarbonSavedKg,
		},
		"streaks": map[string]interface{}{
			"longest_active_days": report.Streaks.LongestActiveDays,
			"longest_rest_days":   report.Streaks.LongestRestDays,
		},
	}

	// Months
	if len(report.Months) > 0 {
		months := make([]map[string]interface{}, len(report.Months))
		for i, m := range report.Months {
			months[i] = map[string]interface{}{
				"month":       m.Month,
				"activities":  m.Activities,
				"distance_m":  m.DistanceM,
				"elevation_m": m.ElevationM,
				"prs":         m.PRs,
			}
		}
		data["months"] = months
	}

	// Moving time by sport
	if len(report.MovingTimeBySport) > 0 {
		sports := make([]map[string]interface{}, len(report.MovingTimeBySport))
		for i, s := range report.MovingTimeBySport {
			sports[i] = map[string]interface{}{
				"sport_type":    s.SportType,
				"moving_time_s": s.MovingTimeS,
			}
		}
		data["moving_time_by_sport"] = sports
	}

	// Start times by hour
	if len(report.StartTimesByHour) > 0 {
		hours := make([]map[string]interface{}, len(report.StartTimesByHour))
		for i, h := range report.StartTimesByHour {
			hours[i] = map[string]interface{}{
				"hour":  h.Hour,
				"count": h.Count,
			}
		}
		data["start_times_by_hour"] = hours
	}

	// Locations
	if len(report.Locations) > 0 {
		locations := make([]map[string]interface{}, len(report.Locations))
		for i, l := range report.Locations {
			locations[i] = map[string]interface{}{
				"lat":   l.Lat,
				"lng":   l.Lng,
				"count": l.Count,
			}
		}
		data["locations"] = locations
	}

	// Biggest activities
	biggest := map[string]interface{}{}
	if report.Biggest.LongestDistance != nil {
		biggest["longest_distance"] = map[string]interface{}{
			"activity_id":      report.Biggest.LongestDistance.ActivityID,
			"name":             report.Biggest.LongestDistance.Name,
			"sport_type":       report.Biggest.LongestDistance.SportType,
			"start_date_local": report.Biggest.LongestDistance.StartDateLocal,
			"value":            report.Biggest.LongestDistance.Value,
		}
	}
	if report.Biggest.MostElevation != nil {
		biggest["most_elevation"] = map[string]interface{}{
			"activity_id":      report.Biggest.MostElevation.ActivityID,
			"name":             report.Biggest.MostElevation.Name,
			"sport_type":       report.Biggest.MostElevation.SportType,
			"start_date_local": report.Biggest.MostElevation.StartDateLocal,
			"value":            report.Biggest.MostElevation.Value,
		}
	}
	if report.Biggest.LongestDuration != nil {
		biggest["longest_duration"] = map[string]interface{}{
			"activity_id":      report.Biggest.LongestDuration.ActivityID,
			"name":             report.Biggest.LongestDuration.Name,
			"sport_type":       report.Biggest.LongestDuration.SportType,
			"start_date_local": report.Biggest.LongestDuration.StartDateLocal,
			"value":            report.Biggest.LongestDuration.Value,
		}
	}
	data["biggest"] = biggest

	// Random photo
	if report.RandomPhoto != nil {
		data["random_photo"] = map[string]interface{}{
			"id":            report.RandomPhoto.ID,
			"activity_id":   report.RandomPhoto.ActivityID,
			"url":           report.RandomPhoto.URL,
			"thumbnail_url": report.RandomPhoto.ThumbnailURL,
			"caption":       report.RandomPhoto.Caption,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}
