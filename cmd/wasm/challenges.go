//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Challenges
// ============================================================================

//wasm:category Challenges

// getChallenges returns paginated challenges for the athlete
// Called from JS: goStorage.getChallenges(filtersJSON)
//wasm:export
func getChallenges(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getChallenges")

	var req struct {
		Month    string `json:"month"`
		Page     int    `json:"page"`
		PerPage  int    `json:"per_page"`
		OrderBy  string `json:"order_by"`
		OrderDir string `json:"order_dir"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
	}

	ctx := context.Background()
	result, err := bridge.challengesService.List(ctx, services.ListChallengesInput{
		AthleteID: bridge.athleteID,
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
}

// ============================================================================
// Training Goals
// ============================================================================

//wasm:category Training Goals

// getTrainingGoals returns training goals config and progress
// Called from JS: goStorage.getTrainingGoals(year?)
//wasm:export
func getTrainingGoals(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getTrainingGoals")

	ctx := context.Background()

	// Get config
	cfg, err := bridge.goals.GetConfig(ctx, bridge.athleteID)
	if err != nil {
		return errorJSON(err)
	}

	// Build progress for each sport and period
	progress := make(map[string]map[string]interface{})
	for _, sport := range cfg.Sports {
		sportProgress := make(map[string]interface{})
		for period := range sport.Targets {
			p, err := bridge.goals.GetProgress(ctx, bridge.athleteID, sport.SportTypes, period)
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
}

// updateTrainingGoals updates training goals config
// Called from JS: goStorage.updateTrainingGoals(configJSON)
//wasm:export
func updateTrainingGoals(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateTrainingGoals")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing config"))
	}

	var cfg storage.TrainingGoalsConfig
	if err := json.Unmarshal([]byte(args[0].String()), &cfg); err != nil {
		return errorJSON(fmt.Errorf("parsing config: %w", err))
	}

	ctx := context.Background()
	if err := bridge.goals.UpsertConfig(ctx, bridge.athleteID, cfg); err != nil {
		return errorJSON(err)
	}

	return successJSON("Training goals updated")
}
