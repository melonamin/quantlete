//go:build js && wasm

package main

import (
	"fmt"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Challenges
// ============================================================================

//wasm:category Challenges

// getChallenges returns paginated challenges for the athlete
// Called from JS: goStorage.getChallenges(filtersJSON)
// NOTE: Replaced by generated adapter genGetChallenges
var getChallenges = wrapWasmAthlete("getChallenges", func(wc *WasmContext) interface{} {
	var req struct {
		Month    string `json:"month"`
		Page     int    `json:"page"`
		PerPage  int    `json:"per_page"`
		OrderBy  string `json:"order_by"`
		OrderDir string `json:"order_dir"`
	}
	if wc.HasArg(0) && wc.ArgString(0) != "" {
		if err := wc.ArgJSON(0, &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
	}

	result, err := wc.Registry.ChallengesService.List(wc.Ctx, services.ListChallengesInput{
		AthleteID: wc.AthleteID,
		Month:     req.Month,
		Page:      req.Page,
		PerPage:   req.PerPage,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
	})
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":          true,
		"data":        result.Data,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
})

// ============================================================================
// Training Goals
// ============================================================================

//wasm:category Training Goals

// getTrainingGoals returns training goals config and progress
// Called from JS: goStorage.getTrainingGoals(year?)
//
//wasm:export
var getTrainingGoals = wrapWasmAthlete("getTrainingGoals", func(wc *WasmContext) interface{} {
	// Get config
	cfg, err := wc.Registry.Goals().GetConfig(wc.Ctx, wc.AthleteID)
	if err != nil {
		return errorJSON(err)
	}

	// Build progress for each sport and period
	progress := make(map[string]map[string]interface{})
	for _, sport := range cfg.Sports {
		sportProgress := make(map[string]interface{})
		for period := range sport.Targets {
			p, err := wc.Registry.Goals().GetProgress(wc.Ctx, wc.AthleteID, sport.SportTypes, period)
			if err != nil {
				continue
			}
			sportProgress[string(period)] = map[string]interface{}{
				"distance_m":     p.DistanceM,
				"elevation_m":    p.ElevationM,
				"moving_time_s":  p.MovingTimeS,
				"activity_count": p.ActivityCount,
			}
		}
		progress[sport.Name] = sportProgress
	}

	return toJSON(map[string]interface{}{
		"ok":       true,
		"config":   cfg,
		"progress": progress,
	})
})

// updateTrainingGoals updates training goals config
// Called from JS: goStorage.updateTrainingGoals(configJSON)
//
//wasm:export
var updateTrainingGoals = wrapWasmAthlete("updateTrainingGoals", func(wc *WasmContext) interface{} {
	var cfg storage.TrainingGoalsConfig
	if err := wc.ArgJSON(0, &cfg); err != nil {
		return errorJSON(fmt.Errorf("parsing config: %w", err))
	}

	if err := wc.Registry.Goals().UpsertConfig(wc.Ctx, wc.AthleteID, cfg); err != nil {
		return errorJSON(err)
	}

	return successJSON("Training goals updated")
})
