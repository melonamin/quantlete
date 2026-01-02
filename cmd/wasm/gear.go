//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Gear Write
// ============================================================================

// saveGear stores gear in the database
// Called from JS: goStorage.saveGear(gearJSON)
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
	if err := gear.Upsert(ctx, g); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Gear %s saved", req.ID))
}

// ============================================================================
// Gear Read
// ============================================================================

// getGear retrieves paginated gear list
// Called from JS: goStorage.getGear(filtersJSON)
func getGear(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getGear")

	// Parse optional filters
	var req struct {
		IncludeRetired bool `json:"include_retired"`
		Page           int  `json:"page"`
		PerPage        int  `json:"per_page"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing gear filters: %w", err))
		}
	}

	ctx := context.Background()
	result, err := gear.ListPaginated(ctx, athleteID, storage.GearFilters{
		IncludeRetired: req.IncludeRetired,
	})
	if err != nil {
		return errorJSON(err)
	}

	// Get activity counts for all gear
	gearIDs := make([]string, len(result.Items))
	for i, g := range result.Items {
		gearIDs[i] = g.ID
	}
	counts, _ := gear.GetActivityCountsBatch(ctx, gearIDs)

	// Convert to response format
	items := make([]map[string]interface{}, len(result.Items))
	for i, g := range result.Items {
		items[i] = gearToMap(g, counts[g.ID])
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

// getGearDetail retrieves a single gear by ID
// Called from JS: goStorage.getGearDetail(id)
func getGearDetail(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getGearDetail")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing gear ID"))
	}

	id := args[0].String()
	ctx := context.Background()

	g, err := gear.GetByID(ctx, id)
	if err != nil {
		return errorJSON(err)
	}
	if g == nil {
		return errorJSON(fmt.Errorf("gear not found"))
	}

	count, _ := gear.GetActivityCount(ctx, id)
	return dataJSON(gearToMap(*g, count))
}

// ============================================================================
// Gear Stats
// ============================================================================

// getGearMonthlyUsage returns monthly usage statistics for gear
// Called from JS: goStorage.getGearMonthlyUsage(filtersJSON)
func getGearMonthlyUsage(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getGearMonthlyUsage")

	includeRetired := false
	if len(args) > 0 && args[0].String() != "" {
		var req struct {
			IncludeRetired bool `json:"include_retired"`
		}
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
		includeRetired = req.IncludeRetired
	}

	ctx := context.Background()
	usage, err := gear.GetMonthlyUsage(ctx, athleteID, includeRetired)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	data := make([]map[string]interface{}, len(usage))
	for i, u := range usage {
		item := map[string]interface{}{
			"month":          u.Month,
			"gear_id":        u.GearID,
			"gear_name":      u.GearName,
			"source":         u.Source,
			"retired":        u.Retired,
			"activity_count": u.ActivityCount,
			"distance":       u.Distance,
			"moving_time":    u.MovingTime,
		}
		if u.Hashtag != "" {
			item["hashtag"] = u.Hashtag
		}
		if u.PurchasePrice != nil {
			item["purchase_price"] = *u.PurchasePrice
		}
		if u.PurchaseCurrency != "" {
			item["purchase_currency"] = u.PurchaseCurrency
		}
		data[i] = item
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}

// ============================================================================
// Gear Map Helper
// ============================================================================

func gearToMap(g storage.Gear, activityCount int) map[string]interface{} {
	m := map[string]interface{}{
		"id":             g.ID,
		"athlete_id":     g.AthleteID,
		"name":           g.Name,
		"primary":        g.Primary,
		"retired":        g.Retired,
		"distance":       g.Distance,
		"brand_name":     g.BrandName,
		"model_name":     g.ModelName,
		"description":    g.Description,
		"source":         g.Source,
		"activity_count": activityCount,
	}
	if g.Hashtag != "" {
		m["hashtag"] = g.Hashtag
	}
	if g.PurchasePrice != nil {
		m["purchase_price"] = *g.PurchasePrice
	}
	if g.PurchaseCurrency != "" {
		m["purchase_currency"] = g.PurchaseCurrency
	}
	return m
}
