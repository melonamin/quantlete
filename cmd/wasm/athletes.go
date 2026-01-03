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
// Athlete Write
// ============================================================================

//wasm:category Athlete - Write

// saveAthlete stores an athlete in the database
// Called from JS: goStorage.saveAthlete(athleteJSON)
//wasm:export
func saveAthlete(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("saveAthlete")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing athlete JSON"))
	}

	athleteJSON := args[0].String()
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
	if err := json.Unmarshal([]byte(athleteJSON), &req); err != nil {
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

	ctx := context.Background()
	if err := bridge.athletes.Upsert(ctx, athlete); err != nil {
		return errorJSON(err)
	}

	// Set the athlete ID on the bridge
	bridge.athleteID = req.ID

	return successJSON(fmt.Sprintf("Athlete %d saved", req.ID))
}

// ============================================================================
// Athlete Metrics (FTP/Weight)
// ============================================================================

//wasm:category Athlete - Metrics

// getFtpHistory returns FTP history for cycling
// Called from JS: goStorage.getFtpHistory()
//wasm:export
func getFtpHistory(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getFtpHistory")

	ctx := context.Background()
	points, err := bridge.athleteMetrics.List(ctx, bridge.athleteID, "ftp_cycling_watts")
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

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}

// getWeightHistory returns weight history
// Called from JS: goStorage.getWeightHistory()
//wasm:export
func getWeightHistory(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getWeightHistory")

	ctx := context.Background()
	points, err := bridge.athleteMetrics.List(ctx, bridge.athleteID, "weight")
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

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}

// updateFtpHistory replaces FTP history
// Called from JS: goStorage.updateFtpHistory(entriesJSON)
//wasm:export
func updateFtpHistory(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateFtpHistory")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing entries"))
	}

	var entries []struct {
		RecordedAt string  `json:"recorded_at"`
		Value      float64 `json:"value"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &entries); err != nil {
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

	ctx := context.Background()
	if err := bridge.athleteMetrics.Replace(ctx, bridge.athleteID, "ftp_cycling_watts", points); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("FTP history updated (%d entries)", len(points)))
}

// updateWeightHistory replaces weight history
// Called from JS: goStorage.updateWeightHistory(entriesJSON)
//wasm:export
func updateWeightHistory(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateWeightHistory")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing entries"))
	}

	var entries []struct {
		RecordedAt string  `json:"recorded_at"`
		Value      float64 `json:"value"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &entries); err != nil {
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

	ctx := context.Background()
	if err := bridge.athleteMetrics.Replace(ctx, bridge.athleteID, "weight", points); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Weight history updated (%d entries)", len(points)))
}
