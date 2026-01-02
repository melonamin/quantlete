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

// Standard durations for power curve (matching Go backend and TS importer)
var powerDurations = []int{5, 10, 30, 60, 300, 480, 1200, 3600}

// ============================================================================
// Best Efforts Write
// ============================================================================

// saveBestEfforts stores best efforts for an activity (replaces existing)
// Called from JS: goStorage.saveBestEfforts(dataJSON)
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

	// Convert to storage format
	efforts := make([]storage.BestEffort, 0, len(req.Efforts))
	for _, e := range req.Efforts {
		be := storage.BestEffort{
			AthleteID:    req.AthleteID,
			ActivityID:   req.ActivityID,
			SportType:    req.SportType,
			DistanceType: e.DistanceType,
			Name:         e.Name,
			DistanceM:    e.DistanceM,
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
	if err := bestEfforts.ReplaceForActivity(ctx, req.AthleteID, req.ActivityID, req.SportType, efforts); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Best efforts for activity %d saved (%d efforts)", req.ActivityID, len(efforts)))
}

// ============================================================================
// Best Efforts Read
// ============================================================================

// getBestEffortPRs retrieves personal records by distance type
// Called from JS: goStorage.getBestEffortPRs(sportType?)
func getBestEffortPRs(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getBestEffortPRs")

	var sportTypes []string
	if len(args) > 0 && args[0].String() != "" {
		sportTypes = []string{args[0].String()}
	}

	ctx := context.Background()
	prs, err := bestEfforts.ListPRs(ctx, athleteID, sportTypes)
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
			"start_date":    pr.StartDateLocal.Format(time.RFC3339),
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": items,
	})
}

// getBestEffortsForType retrieves all efforts for a specific distance type
// Called from JS: goStorage.getBestEffortsForType(distanceType, sportType?)
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
	efforts, err := bestEfforts.ListByDistanceType(ctx, athleteID, distanceType, sportTypes)
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
			"start_date":    e.StartDateLocal.Format(time.RFC3339),
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

// getEddingtonData returns Eddington history and computes the current number
// Called from JS: goStorage.getEddingtonData(filtersJSON)
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
	history, err := stats.GetEddingtonHistory(ctx, athleteID, sportTypes)
	if err != nil {
		return errorJSON(err)
	}

	// The current Eddington number is the last point in history
	currentE := 0
	if len(history) > 0 {
		currentE = history[len(history)-1].Number
	}

	// Convert history to response format
	historyData := make([]map[string]interface{}, len(history))
	for i, h := range history {
		historyData[i] = map[string]interface{}{
			"date":   h.Date,
			"number": h.Number,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":      true,
		"number":  currentE,
		"history": historyData,
	})
}

// ============================================================================
// Power Functions
// ============================================================================

// computePowerBestEfforts computes and stores power best efforts for an activity
// Called from JS: goStorage.computePowerBestEfforts(dataJSON)
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
	if err := power.EnsureActivityComputed(ctx, req.AthleteID, req.ActivityID, powerDurations); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Power best efforts computed for activity %d", req.ActivityID))
}

// ============================================================================
// Training Load
// ============================================================================

// getTrainingLoad returns training load data (daily series + summary)
// Called from JS: goStorage.getTrainingLoad(filtersJSON)
func getTrainingLoad(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getTrainingLoad")

	var after, before *time.Time
	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			After  string `json:"after"`
			Before string `json:"before"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		if req.After != "" {
			t, err := time.Parse(time.RFC3339, req.After)
			if err == nil {
				after = &t
			}
		}
		if req.Before != "" {
			t, err := time.Parse(time.RFC3339, req.Before)
			if err == nil {
				before = &t
			}
		}
	}

	ctx := context.Background()

	// Get daily series
	series, err := trainingLoad.GetDailySeries(ctx, athleteID, after, before)
	if err != nil {
		return errorJSON(err)
	}

	// Get summary (latest CTL/ATL/TSB)
	summary, err := trainingLoad.GetSummary(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	// Convert series to response format
	seriesData := make([]map[string]interface{}, len(series))
	for i, p := range series {
		seriesData[i] = map[string]interface{}{
			"day": p.Day,
			"tss": p.TSS,
			"ctl": p.CTL,
			"atl": p.ATL,
			"tsb": p.TSB,
		}
	}

	result := map[string]interface{}{
		"ok":     true,
		"series": seriesData,
	}

	if summary != nil {
		result["summary"] = map[string]interface{}{
			"day": summary.Day,
			"tss": summary.TSS,
			"ctl": summary.CTL,
			"atl": summary.ATL,
			"tsb": summary.TSB,
		}
	}

	return dataJSON(result)
}

// ============================================================================
// Power Stats
// ============================================================================

// getPowerStats returns power best efforts and history
// Called from JS: goStorage.getPowerStats(filtersJSON)
func getPowerStats(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getPowerStats")

	// Default durations matching the Go backend
	durations := []int{5, 10, 30, 60, 300, 480, 1200, 3600}
	var after, before *time.Time
	var sportTypes []string
	historyDuration := 300 // Default to 5 minutes for history

	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			Durations       []int    `json:"durations"`
			After           string   `json:"after"`
			Before          string   `json:"before"`
			SportTypes      []string `json:"sport_types"`
			HistoryDuration int      `json:"history_duration"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		if len(req.Durations) > 0 {
			durations = req.Durations
		}
		if req.After != "" {
			t, err := time.Parse(time.RFC3339, req.After)
			if err == nil {
				after = &t
			}
		}
		if req.Before != "" {
			t, err := time.Parse(time.RFC3339, req.Before)
			if err == nil {
				before = &t
			}
		}
		sportTypes = req.SportTypes
		if req.HistoryDuration > 0 {
			historyDuration = req.HistoryDuration
		}
	}

	ctx := context.Background()

	// Get best efforts for each duration
	best, err := power.GetBest(ctx, athleteID, durations, after, before, sportTypes)
	if err != nil {
		return errorJSON(err)
	}

	// Get history for the specified duration
	history, err := power.GetHistory(ctx, athleteID, historyDuration, after, before, sportTypes)
	if err != nil {
		return errorJSON(err)
	}

	// Convert best efforts to response format
	bestData := make([]map[string]interface{}, len(best))
	for i, b := range best {
		bestData[i] = map[string]interface{}{
			"duration_s":  b.DurationS,
			"watts":       b.Watts,
			"activity_id": b.ActivityID,
			"start_date":  b.StartDate.Format(time.RFC3339),
		}
	}

	// Convert history to response format
	historyData := make([]map[string]interface{}, len(history))
	for i, h := range history {
		historyData[i] = map[string]interface{}{
			"date":  h.Date,
			"watts": h.Watts,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":      true,
		"best":    bestData,
		"history": historyData,
	})
}
