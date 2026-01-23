//go:build js && wasm

package main

import (
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// Standard durations for power curve (matching Go backend and TS importer)
var powerDurations = []int{5, 10, 30, 60, 300, 480, 1200, 3600}

// Power zone bounds based on Coggan's standard 5-zone training model.
// Values represent upper bound as percentage of FTP:
//   - Zone 1 (Active Recovery): <= 55% FTP
//   - Zone 2 (Endurance): 55-75% FTP
//   - Zone 3 (Tempo): 75-90% FTP
//   - Zone 4 (Lactate Threshold): 90-105% FTP
//   - Zone 5 (VO2max): > 105% FTP
//
// The final value (MaxFloat64) is a sentinel for the unbounded upper zone.
var powerZoneBounds = []float64{0.55, 0.75, 0.9, 1.05, math.MaxFloat64}

// ============================================================================
// Power Functions
// ============================================================================

//wasm:category Power

// computePowerBestEfforts computes and stores power best efforts for an activity
// Called from JS: goStorage.computePowerBestEfforts(dataJSON)
//
//wasm:export
var computePowerBestEfforts = wrapWasm("computePowerBestEfforts", func(wc *WasmContext) interface{} {
	var req struct {
		ActivityID int64 `json:"activity_id"`
		AthleteID  int64 `json:"athlete_id"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing power request: %w", err))
	}

	if err := wc.Registry.Power().EnsureActivityComputed(wc.Ctx, req.AthleteID, req.ActivityID, powerDurations); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Power best efforts computed for activity %d", req.ActivityID))
})

// PowerZonesResponse matches the server API response format
type PowerZonesResponse struct {
	FTPWatts      float64   `json:"ftp_watts,omitempty"`
	SecondsByZone []int     `json:"seconds_by_zone"`
	TotalSeconds  int       `json:"total_seconds"`
	Bounds        []float64 `json:"bounds"`
}

// activityInfo holds activity ID and start time for power zone computation.
type activityInfo struct {
	id    int64
	start storage.SQLiteTime
}

// parseDateFilter parses a YYYY-MM-DD date string into a time.Time.
// Uses local timezone to match user expectations for date filters.
func parseDateFilter(dateStr string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", dateStr, time.Local)
}

// binWattsIntoZones categorizes watt samples into power zones based on FTP.
// Returns the count per zone and total valid samples processed.
func binWattsIntoZones(watts []float64, ftp float64, secondsByZone []int) int {
	totalSeconds := 0
	for _, wv := range watts {
		if wv <= 0 {
			continue
		}
		p := wv / ftp
		for i, b := range powerZoneBounds {
			if p <= b {
				secondsByZone[i]++
				totalSeconds++
				break
			}
		}
	}
	return totalSeconds
}

// getPowerZones computes time spent in each power zone across all activities
// Called from JS: goStorage.getPowerZones(optionalFiltersJSON)
//
// Note: FTP lookup must be per-activity because FTP values change over time.
// Stream fetching is batched to avoid N+1 queries.
//
//wasm:export
//wasm:category Stats
var getPowerZones = wrapWasmAthlete("getPowerZones", func(wc *WasmContext) interface{} {
	var req struct {
		After  *string `json:"after,omitempty"`
		Before *string `json:"before,omitempty"`
	}
	if wc.HasArg(0) && wc.ArgString(0) != "" {
		if err := wc.ArgJSON(0, &req); err != nil {
			return errorJSON(fmt.Errorf("parsing power zones request: %w", err))
		}
	}

	// Parse and validate date filters - return error for invalid dates
	var after, before *time.Time
	if req.After != nil {
		t, err := parseDateFilter(*req.After)
		if err != nil {
			return errorJSON(fmt.Errorf("invalid 'after' date %q: expected YYYY-MM-DD format", *req.After))
		}
		after = &t
	}
	if req.Before != nil {
		t, err := parseDateFilter(*req.Before)
		if err != nil {
			return errorJSON(fmt.Errorf("invalid 'before' date %q: expected YYYY-MM-DD format", *req.Before))
		}
		before = &t
	}

	// Validate date range if both are provided
	if after != nil && before != nil && after.After(*before) {
		return errorJSON(fmt.Errorf("'after' date must be before 'before' date"))
	}

	metricsRepo := wc.Registry.AthleteMetrics()
	streamsRepo := wc.Registry.Streams()

	// Get the athlete's current (most recent) FTP for the response
	currentFTP, err := metricsRepo.LatestBefore(wc.Ctx, wc.AthleteID, "ftp_cycling_watts", time.Now())
	var ftpForResponse float64
	if err == nil && currentFTP != nil && currentFTP.Value > 0 {
		ftpForResponse = currentFTP.Value
	}

	// Query activities that have watts streams
	query := `
		SELECT DISTINCT a.id, a.start_date
		FROM activities a
		JOIN activity_streams s ON s.activity_id = a.id AND s.stream_type = 'watts'
		WHERE a.athlete_id = ?
	`
	args := []any{wc.AthleteID}
	if after != nil {
		query += " AND a.start_date >= ?"
		args = append(args, storage.TimeToSQL(*after))
	}
	if before != nil {
		query += " AND a.start_date <= ?"
		args = append(args, storage.TimeToSQL(*before))
	}

	rows, err := wc.Registry.DB().QueryContext(wc.Ctx, query, args...)
	if err != nil {
		return errorJSON(fmt.Errorf("querying activities: %w", err))
	}

	// Collect activity info
	var activities []activityInfo
	for rows.Next() {
		var a activityInfo
		if err := rows.Scan(&a.id, &a.start); err != nil {
			slog.Warn("skipping activity with scan error in power zones",
				"athlete_id", wc.AthleteID,
				"error", err)
			continue
		}
		activities = append(activities, a)
	}
	if err := rows.Err(); err != nil {
		slog.Warn("row iteration error in power zones",
			"athlete_id", wc.AthleteID,
			"error", err)
	}
	_ = rows.Close()

	if len(activities) == 0 {
		return dataJSON(PowerZonesResponse{
			FTPWatts:      ftpForResponse,
			SecondsByZone: make([]int, 5),
			TotalSeconds:  0,
			Bounds:        powerZoneBounds[:4],
		})
	}

	// Batch fetch all watts streams in a single query (avoids N+1)
	activityIDs := make([]int64, len(activities))
	for i, a := range activities {
		activityIDs[i] = a.id
	}

	wattsStreams, err := streamsRepo.GetStreamTypeByActivityIDs(wc.Ctx, activityIDs, "watts")
	if err != nil {
		slog.Warn("failed to batch fetch watts streams",
			"athlete_id", wc.AthleteID,
			"error", err)
		// Continue with empty map - will skip all activities gracefully
		wattsStreams = make(map[int64][]byte)
	}

	secondsByZone := make([]int, 5)
	totalSeconds := 0

	for _, a := range activities {
		// Get FTP valid at activity start time (FTP changes over time - must be per-activity)
		ftpPoint, err := metricsRepo.LatestBefore(wc.Ctx, wc.AthleteID, "ftp_cycling_watts", a.start.Time)
		if err != nil {
			slog.Debug("FTP lookup failed for power zones",
				"activity_id", a.id,
				"error", err)
			continue
		}
		if ftpPoint == nil || ftpPoint.Value <= 0 {
			// No FTP defined for this activity's date - skip silently (expected for early activities)
			continue
		}
		ftp := ftpPoint.Value

		// Get pre-fetched watts stream
		wattsRaw, ok := wattsStreams[a.id]
		if !ok || len(wattsRaw) == 0 {
			continue
		}

		watts, err := storage.DecodeFloat64Array(wattsRaw)
		if err != nil {
			slog.Warn("failed to decode watts stream for power zones",
				"activity_id", a.id,
				"error", err)
			continue
		}

		// Bin watts into zones
		totalSeconds += binWattsIntoZones(watts, ftp, secondsByZone)
	}

	return dataJSON(PowerZonesResponse{
		FTPWatts:      ftpForResponse,
		SecondsByZone: secondsByZone,
		TotalSeconds:  totalSeconds,
		Bounds:        powerZoneBounds[:4], // Return first 4 bounds (not the sentinel)
	})
})
