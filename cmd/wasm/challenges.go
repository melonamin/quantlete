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
// Challenges
// ============================================================================

// getChallenges returns paginated challenges for the athlete
// Called from JS: goStorage.getChallenges(filtersJSON)
func getChallenges(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getChallenges")

	var req struct {
		Month   string `json:"month"`
		Page    int    `json:"page"`
		PerPage int    `json:"per_page"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
	}

	ctx := context.Background()
	filters := storage.ChallengeFilters{
		Month: req.Month,
	}
	filters.Page = req.Page
	filters.PerPage = req.PerPage

	result, err := challenges.ListPaginated(ctx, athleteID, filters)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	items := make([]map[string]interface{}, len(result.Items))
	for i, c := range result.Items {
		item := map[string]interface{}{
			"id":         c.ID,
			"athlete_id": c.AthleteID,
			"name":       c.Name,
			"created_at": c.CreatedAt.Format(time.RFC3339),
		}
		if c.Slug != "" {
			item["slug"] = c.Slug
		}
		if c.BadgeURL != "" {
			item["badge_url"] = c.BadgeURL
		}
		if c.LocalBadgeURL != "" {
			item["local_badge_url"] = c.LocalBadgeURL
		}
		if c.CompletionDate != nil {
			item["completion_date"] = c.CompletionDate.Format(time.RFC3339)
		}
		if c.Month != "" {
			item["month"] = c.Month
		}
		items[i] = item
	}

	return toJSON(map[string]interface{}{
		"ok":          true,
		"data":        items,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// ============================================================================
// Training Goals
// ============================================================================

// getTrainingGoals returns training goals config and progress
// Called from JS: goStorage.getTrainingGoals(year?)
func getTrainingGoals(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getTrainingGoals")

	ctx := context.Background()

	// Get config
	cfg, err := goals.GetConfig(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	// Build progress for each sport and period
	progress := make(map[string]map[string]interface{})
	for _, sport := range cfg.Sports {
		sportProgress := make(map[string]interface{})
		for period := range sport.Targets {
			p, err := goals.GetProgress(ctx, athleteID, sport.SportTypes, period)
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
	if err := goals.UpsertConfig(ctx, athleteID, cfg); err != nil {
		return errorJSON(err)
	}

	return successJSON("Training goals updated")
}
