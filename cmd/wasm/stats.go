//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
)

// Standard durations for power curve (matching Go backend and TS importer)
var powerDurations = []int{5, 10, 30, 60, 300, 480, 1200, 3600}

// ============================================================================
// Best Efforts Write
// ============================================================================

//wasm:category Best Efforts - Write

// saveBestEfforts stores best efforts for an activity (replaces existing)
// Called from JS: goStorage.saveBestEfforts(dataJSON)
//wasm:export
func saveBestEfforts(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("saveBestEfforts")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing data JSON"))
	}

	dataJSON := args[0].String()
	var req struct {
		AthleteID  int64  `json:"athlete_id"`
		ActivityID int64  `json:"activity_id"`
		SportType  string `json:"sport_type"`
		Efforts    []struct {
			DistanceType string  `json:"distance_type"`
			Name         string  `json:"name"`
			DistanceM    float64 `json:"distance_m"`
			ElapsedTime  int     `json:"elapsed_time"`
			MovingTime   *int    `json:"moving_time"`
			StartIndex   *int    `json:"start_index"`
			EndIndex     *int    `json:"end_index"`
			PRRank       *int    `json:"pr_rank"`
			StartDate    string  `json:"start_date"`
		} `json:"efforts"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing best efforts: %w", err))
	}

	// Validate efforts array size
	if len(req.Efforts) > maxBestEffortsPerSave {
		return errorJSON(fmt.Errorf("too many best efforts: %d > %d", len(req.Efforts), maxBestEffortsPerSave))
	}

	// Convert to storage format with canonicalization
	efforts := make([]storage.BestEffort, 0, len(req.Efforts))
	for _, e := range req.Efforts {
		// Apply canonical distance type mapping
		distanceType, canonicalM := shared.CanonicalBestEffortDistanceType(e.DistanceM, e.Name)
		be := storage.BestEffort{
			AthleteID:    req.AthleteID,
			ActivityID:   req.ActivityID,
			SportType:    req.SportType,
			DistanceType: distanceType,
			Name:         e.Name,
			DistanceM:    canonicalM,
			ElapsedTimeS: e.ElapsedTime,
			MovingTimeS:  e.MovingTime,
			StartIndex:   e.StartIndex,
			EndIndex:     e.EndIndex,
			PRRank:       e.PRRank,
		}
		if e.StartDate != "" {
			if t, err := time.Parse(time.RFC3339, e.StartDate); err == nil {
				be.StartDate = &storage.SQLiteTime{Time: t}
			}
		}
		efforts = append(efforts, be)
	}

	ctx := context.Background()
	if err := bridge.bestEfforts.ReplaceForActivity(ctx, req.AthleteID, req.ActivityID, req.SportType, efforts); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Best efforts for activity %d saved (%d efforts)", req.ActivityID, len(efforts)))
}

// ============================================================================
// Best Efforts Read
// ============================================================================

//wasm:category Best Efforts - Read

// getBestEffortPRs retrieves personal records by distance type
// Called from JS: goStorage.getBestEffortPRs(sportType?)
//wasm:export
func getBestEffortPRs(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getBestEffortPRs")

	var sportTypes []string
	if len(args) > 0 && args[0].String() != "" {
		sportTypes = []string{args[0].String()}
	}

	ctx := context.Background()
	prs, err := bridge.statsService.GetBestEffortPRs(ctx, services.GetBestEffortPRsInput{
		AthleteID:  bridge.athleteID,
		SportTypes: sportTypes,
	})
	if err != nil {
		return errorJSON(err)
	}

	items := make([]map[string]interface{}, len(prs))
	for i, pr := range prs {
		items[i] = map[string]interface{}{
			"distance_type": pr.DistanceType,
			"distance_m":    pr.DistanceM,
			"elapsed_time":  pr.ElapsedTimeS,
			"activity_id":   pr.ActivityID,
			"activity_name": pr.ActivityName,
			"start_date":    pr.StartDateLocal,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": items,
	})
}

// getBestEffortsForType retrieves all efforts for a specific distance type
// Called from JS: goStorage.getBestEffortsForType(distanceType, sportType?)
//wasm:export
func getBestEffortsForType(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getBestEffortsForType")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing distance type"))
	}

	distanceType := args[0].String()
	var sportTypes []string
	if len(args) > 1 && args[1].String() != "" {
		sportTypes = []string{args[1].String()}
	}

	ctx := context.Background()
	efforts, err := bridge.statsService.GetBestEffortsForType(ctx, services.GetBestEffortsForTypeInput{
		AthleteID:    bridge.athleteID,
		DistanceType: distanceType,
		SportTypes:   sportTypes,
	})
	if err != nil {
		return errorJSON(err)
	}

	items := make([]map[string]interface{}, len(efforts))
	for i, e := range efforts {
		items[i] = map[string]interface{}{
			"distance_type": e.DistanceType,
			"distance_m":    e.DistanceM,
			"elapsed_time":  e.ElapsedTimeS,
			"activity_id":   e.ActivityID,
			"start_date":    e.StartDateLocal,
		}
		if e.PRRank != nil {
			items[i]["pr_rank"] = *e.PRRank
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": items,
	})
}

// ============================================================================
// Eddington Data
// ============================================================================

//wasm:category Eddington

// getEddingtonData returns Eddington data including current number, distribution, and next steps
// Called from JS: goStorage.getEddingtonData(filtersJSON)
//wasm:export
func getEddingtonData(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getEddingtonData")

	var sportTypes []string
	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			SportTypes []string `json:"sport_types"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		sportTypes = req.SportTypes
	}

	ctx := context.Background()
	result, err := bridge.statsService.GetEddingtonData(ctx, services.GetEddingtonDataInput{
		AthleteID:  bridge.athleteID,
		SportTypes: sportTypes,
	})
	if err != nil {
		return errorJSON(err)
	}

	// Also get history for backward compatibility
	history, err := bridge.statsService.GetEddingtonHistory(ctx, services.GetEddingtonHistoryInput{
		AthleteID:  bridge.athleteID,
		SportTypes: sportTypes,
	})
	if err != nil {
		return errorJSON(err)
	}

	// Convert history to response format
	historyData := make([]map[string]interface{}, len(history))
	for i, h := range history {
		historyData[i] = map[string]interface{}{
			"date":   h.Date,
			"number": h.Number,
		}
	}

	// Convert distribution to response format
	distributionData := make([]map[string]interface{}, len(result.Distribution))
	for i, d := range result.Distribution {
		distributionData[i] = map[string]interface{}{
			"date":     d.Date,
			"distance": d.Distance,
		}
	}

	// Convert next steps to response format
	nextStepsData := make([]map[string]interface{}, len(result.NextSteps))
	for i, s := range result.NextSteps {
		nextStepsData[i] = map[string]interface{}{
			"target":       s.Target,
			"rides_needed": s.RidesNeeded,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":           true,
		"number":       result.Number,
		"history":      historyData,
		"distribution": distributionData,
		"next_steps":   nextStepsData,
	})
}

// ============================================================================
// Power Functions
// ============================================================================

//wasm:category Power

// computePowerBestEfforts computes and stores power best efforts for an activity
// Called from JS: goStorage.computePowerBestEfforts(dataJSON)
//wasm:export
func computePowerBestEfforts(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("computePowerBestEfforts")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing data JSON"))
	}

	dataJSON := args[0].String()
	var req struct {
		ActivityID int64 `json:"activity_id"`
		AthleteID  int64 `json:"athlete_id"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing power request: %w", err))
	}

	ctx := context.Background()
	if err := bridge.power.EnsureActivityComputed(ctx, req.AthleteID, req.ActivityID, powerDurations); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Power best efforts computed for activity %d", req.ActivityID))
}

// ============================================================================
// Training Load
// ============================================================================

//wasm:category Training Load

// getTrainingLoad returns training load data (daily series + summary)
// Called from JS: goStorage.getTrainingLoad(filtersJSON)
//wasm:export
func getTrainingLoad(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getTrainingLoad")

	input := services.GetTrainingLoadInput{
		AthleteID: bridge.athleteID,
	}

	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			After  string `json:"after"`
			Before string `json:"before"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		if t, ok := shared.ParseDateParam(req.After); ok {
			input.After = &t
		}
		if t, ok := shared.ParseDateParam(req.Before); ok {
			input.Before = &t
		}
	}

	ctx := context.Background()
	result, err := bridge.statsService.GetTrainingLoad(ctx, input)
	if err != nil {
		return errorJSON(err)
	}

	// Convert series to response format
	seriesData := make([]map[string]interface{}, len(result.Series))
	for i, p := range result.Series {
		seriesData[i] = map[string]interface{}{
			"day": p.Day,
			"tss": p.TSS,
			"ctl": p.CTL,
			"atl": p.ATL,
			"tsb": p.TSB,
		}
	}

	output := map[string]interface{}{
		"ok":     true,
		"series": seriesData,
	}

	if result.Summary != nil {
		output["summary"] = map[string]interface{}{
			"day": result.Summary.Day,
			"tss": result.Summary.TSS,
			"ctl": result.Summary.CTL,
			"atl": result.Summary.ATL,
			"tsb": result.Summary.TSB,
		}
	}

	return dataJSON(output)
}

// ============================================================================
// Power Stats
// ============================================================================

//wasm:category Power Stats

// getPowerStats returns power best efforts and history
// Called from JS: goStorage.getPowerStats(filtersJSON)
//wasm:export
func getPowerStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getPowerStats")

	input := services.GetPowerStatsInput{
		AthleteID: bridge.athleteID,
	}

	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			After      string   `json:"after"`
			Before     string   `json:"before"`
			SportTypes []string `json:"sport_types"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		if t, ok := shared.ParseDateParam(req.After); ok {
			input.After = &t
		}
		if t, ok := shared.ParseDateParam(req.Before); ok {
			input.Before = &t
		}
		input.SportTypes = req.SportTypes
	}

	ctx := context.Background()
	result, err := bridge.statsService.GetPowerStats(ctx, input)
	if err != nil {
		return errorJSON(err)
	}

	// Convert best efforts to response format
	bestData := make([]map[string]interface{}, len(result.Best))
	for i, b := range result.Best {
		bestData[i] = map[string]interface{}{
			"duration_s":  b.DurationS,
			"watts":       b.Watts,
			"activity_id": b.ActivityID,
			"start_date":  b.StartDate,
		}
	}

	// Convert history to response format (use first available duration's history)
	historyData := make(map[string]interface{})
	for dur, points := range result.History {
		durPoints := make([]map[string]interface{}, len(points))
		for i, h := range points {
			durPoints[i] = map[string]interface{}{
				"date":  h.Date,
				"watts": h.Watts,
			}
		}
		historyData[fmt.Sprintf("%d", dur)] = durPoints
	}

	return toJSON(map[string]interface{}{
		"ok":        true,
		"durations": result.DurationsS,
		"best":      bestData,
		"history":   historyData,
	})
}
