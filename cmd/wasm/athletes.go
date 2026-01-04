//go:build js && wasm

package main

import (
	"fmt"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Athlete Write
// ============================================================================

//wasm:category Athlete - Write

// saveAthlete stores an athlete in the database
// Called from JS: goStorage.saveAthlete(athleteJSON)
//
//wasm:export
var saveAthlete = wrapWasm("saveAthlete", func(wc *WasmContext) interface{} {
	var req struct {
		ID            int64  `json:"id"`
		Username      string `json:"username"`
		FirstName     string `json:"firstname"`
		LastName      string `json:"lastname"`
		ProfileMedium string `json:"profile_medium"`
		Profile       string `json:"profile"`
		City          string `json:"city"`
		State         string `json:"state"`
		Country       string `json:"country"`
		Sex           string `json:"sex"`
		Premium       bool   `json:"premium"`
		Summit        bool   `json:"summit"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing athlete: %w", err))
	}

	athlete := &storage.Athlete{
		ID:            req.ID,
		Username:      req.Username,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		ProfileMedium: req.ProfileMedium,
		Profile:       req.Profile,
		City:          req.City,
		State:         req.State,
		Country:       req.Country,
		Sex:           req.Sex,
		Premium:       req.Premium,
		Summit:        req.Summit,
	}

	if err := wc.Registry.Athletes().Upsert(wc.Ctx, athlete); err != nil {
		return errorJSON(err)
	}

	// Set the athlete ID on the bridge
	// Note: This is technically a race, but acceptable since athlete ID is set once
	// during initialization and read-only thereafter.
	wc.Bridge.athleteID = req.ID

	return successJSON(fmt.Sprintf("Athlete %d saved", req.ID))
})

// ============================================================================
// Athlete Metrics (FTP/Weight)
// ============================================================================

//wasm:category Athlete - Metrics

// getFtpHistory returns FTP history for cycling
// Called from JS: goStorage.getFtpHistory()
//
//wasm:export
var getFtpHistory = wrapWasmAthlete("getFtpHistory", func(wc *WasmContext) interface{} {
	points, err := wc.Registry.AthleteMetrics().List(wc.Ctx, wc.AthleteID, "ftp_cycling_watts")
	if err != nil {
		return errorJSON(err)
	}

	data := make([]map[string]interface{}, len(points))
	for i, p := range points {
		data[i] = map[string]interface{}{
			"recorded_at": p.RecordedAt.Format(time.RFC3339),
			"value":       p.Value,
		}
	}

	return dataJSON(data)
})

// getFtpRunningHistory returns FTP history for running
// Called from JS: goStorage.getFtpRunningHistory()
//
//wasm:export
var getFtpRunningHistory = wrapWasmAthlete("getFtpRunningHistory", func(wc *WasmContext) interface{} {
	points, err := wc.Registry.AthleteMetrics().List(wc.Ctx, wc.AthleteID, "ftp_running_mps")
	if err != nil {
		return errorJSON(err)
	}

	data := make([]map[string]interface{}, len(points))
	for i, p := range points {
		data[i] = map[string]interface{}{
			"recorded_at": p.RecordedAt.Format(time.RFC3339),
			"value":       p.Value,
		}
	}

	return dataJSON(data)
})

// getWeightHistory returns weight history
// Called from JS: goStorage.getWeightHistory()
//
//wasm:export
var getWeightHistory = wrapWasmAthlete("getWeightHistory", func(wc *WasmContext) interface{} {
	points, err := wc.Registry.AthleteMetrics().List(wc.Ctx, wc.AthleteID, "weight_kg")
	if err != nil {
		return errorJSON(err)
	}

	data := make([]map[string]interface{}, len(points))
	for i, p := range points {
		data[i] = map[string]interface{}{
			"recorded_at": p.RecordedAt.Format(time.RFC3339),
			"value":       p.Value,
		}
	}

	return dataJSON(data)
})

// updateFtpHistory replaces FTP history
// Called from JS: goStorage.updateFtpHistory(entriesJSON)
//
//wasm:export
var updateFtpHistory = wrapWasmAthlete("updateFtpHistory", func(wc *WasmContext) interface{} {
	var entries []struct {
		RecordedAt string  `json:"recorded_at"`
		Value      float64 `json:"value"`
	}
	if err := wc.ArgJSON(0, &entries); err != nil {
		return errorJSON(fmt.Errorf("parsing entries: %w", err))
	}

	points := make([]storage.AthleteMetricPoint, len(entries))
	for i, e := range entries {
		t, err := time.Parse(time.RFC3339, e.RecordedAt)
		if err != nil {
			return errorJSON(fmt.Errorf("parsing date %s: %w", e.RecordedAt, err))
		}
		points[i] = storage.AthleteMetricPoint{
			RecordedAt: storage.SQLiteTime{Time: t},
			Value:      e.Value,
		}
	}

	if err := wc.Registry.AthleteMetrics().Replace(wc.Ctx, wc.AthleteID, "ftp_cycling_watts", points); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("FTP history updated (%d entries)", len(points)))
})

// updateFtpRunningHistory replaces running FTP history
// Called from JS: goStorage.updateFtpRunningHistory(entriesJSON)
//
//wasm:export
var updateFtpRunningHistory = wrapWasmAthlete("updateFtpRunningHistory", func(wc *WasmContext) interface{} {
	var entries []struct {
		RecordedAt string  `json:"recorded_at"`
		Value      float64 `json:"value"`
	}
	if err := wc.ArgJSON(0, &entries); err != nil {
		return errorJSON(fmt.Errorf("parsing entries: %w", err))
	}

	points := make([]storage.AthleteMetricPoint, len(entries))
	for i, e := range entries {
		t, err := time.Parse(time.RFC3339, e.RecordedAt)
		if err != nil {
			return errorJSON(fmt.Errorf("parsing date %s: %w", e.RecordedAt, err))
		}
		points[i] = storage.AthleteMetricPoint{
			RecordedAt: storage.SQLiteTime{Time: t},
			Value:      e.Value,
		}
	}

	if err := wc.Registry.AthleteMetrics().Replace(wc.Ctx, wc.AthleteID, "ftp_running_mps", points); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Running FTP history updated (%d entries)", len(points)))
})

// updateWeightHistory replaces weight history
// Called from JS: goStorage.updateWeightHistory(entriesJSON)
//
//wasm:export
var updateWeightHistory = wrapWasmAthlete("updateWeightHistory", func(wc *WasmContext) interface{} {
	var entries []struct {
		RecordedAt string  `json:"recorded_at"`
		Value      float64 `json:"value"`
	}
	if err := wc.ArgJSON(0, &entries); err != nil {
		return errorJSON(fmt.Errorf("parsing entries: %w", err))
	}

	points := make([]storage.AthleteMetricPoint, len(entries))
	for i, e := range entries {
		t, err := time.Parse(time.RFC3339, e.RecordedAt)
		if err != nil {
			return errorJSON(fmt.Errorf("parsing date %s: %w", e.RecordedAt, err))
		}
		points[i] = storage.AthleteMetricPoint{
			RecordedAt: storage.SQLiteTime{Time: t},
			Value:      e.Value,
		}
	}

	if err := wc.Registry.AthleteMetrics().Replace(wc.Ctx, wc.AthleteID, "weight_kg", points); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Weight history updated (%d entries)", len(points)))
})
