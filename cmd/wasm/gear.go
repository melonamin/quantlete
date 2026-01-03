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
// Gear Write (for importer - uses repository directly)
// ============================================================================

//wasm:category Gear - Write

// saveGear stores gear in the database
// Called from JS: goStorage.saveGear(gearJSON)
//wasm:export
func saveGear(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("saveGear")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing gear JSON"))
	}

	gearJSON := args[0].String()
	var req struct {
		ID          string  `json:"id"`
		AthleteID   int64   `json:"athlete_id"`
		Name        string  `json:"name"`
		Primary     bool    `json:"primary"`
		Retired     bool    `json:"retired"`
		Distance    float64 `json:"distance"`
		BrandName   string  `json:"brand_name"`
		ModelName   string  `json:"model_name"`
		Description string  `json:"description"`
	}
	if err := json.Unmarshal([]byte(gearJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing gear: %w", err))
	}

	g := &storage.Gear{
		ID:          req.ID,
		AthleteID:   req.AthleteID,
		Name:        req.Name,
		Primary:     req.Primary,
		Retired:     req.Retired,
		Distance:    req.Distance,
		BrandName:   req.BrandName,
		ModelName:   req.ModelName,
		Description: req.Description,
		Source:      "strava",
	}

	ctx := context.Background()
	if err := bridge.gear.Upsert(ctx, g); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Gear %s saved", req.ID))
}

// ============================================================================
// Gear Read (uses GearService)
// ============================================================================

//wasm:category Gear - Read

// getGear retrieves paginated gear list
// Called from JS: goStorage.getGear(filtersJSON)
//wasm:export
func getGear(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getGear")

	var input services.ListGearInput
	input.AthleteID = bridge.athleteID

	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			IncludeRetired bool `json:"include_retired"`
			Page           int  `json:"page"`
			PerPage        int  `json:"per_page"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing gear filters: %w", err))
		}
		input.IncludeRetired = req.IncludeRetired
		input.Page = req.Page
		input.PerPage = req.PerPage
	}

	ctx := context.Background()
	result, err := bridge.gearService.List(ctx, input)
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

// getGearDetail retrieves a single gear by ID
// Called from JS: goStorage.getGearDetail(gearId)
//wasm:export
func getGearDetail(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getGearDetail")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing gear ID"))
	}

	ctx := context.Background()
	result, err := bridge.gearService.GetByID(ctx, services.GetGearInput{
		AthleteID: bridge.athleteID,
		GearID:    args[0].String(),
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// ============================================================================
// Gear Stats (uses GearService)
// ============================================================================

//wasm:category Gear - Stats

// getGearMonthlyUsage returns monthly usage statistics for gear
// Called from JS: goStorage.getGearMonthlyUsage(filtersJSON)
//wasm:export
func getGearMonthlyUsage(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getGearMonthlyUsage")

	var input services.MonthlyUsageInput
	input.AthleteID = bridge.athleteID

	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			IncludeRetired bool `json:"include_retired"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		input.IncludeRetired = req.IncludeRetired
	}

	ctx := context.Background()
	result, err := bridge.gearService.MonthlyUsage(ctx, input)
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": result,
	})
}
