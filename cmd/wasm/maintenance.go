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
// Maintenance
// ============================================================================

// getMaintenanceDue returns components with maintenance status
// Called from JS: goStorage.getMaintenanceDue()
func getMaintenanceDue(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getMaintenanceDue")

	ctx := context.Background()
	due, err := maintenance.Due(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	items := make([]map[string]interface{}, len(due))
	for i, d := range due {
		item := map[string]interface{}{
			"id":                d.ID,
			"gear_id":           d.GearID,
			"name":              d.Name,
			"created_at":        d.CreatedAt.Format(time.RFC3339),
			"updated_at":        d.UpdatedAt.Format(time.RFC3339),
			"distance_since":    d.DistanceSince,
			"moving_time_since": d.MovingTimeSince,
			"days_since":        d.DaysSince,
			"is_due":            d.IsDue,
		}
		if d.ImageURL != "" {
			item["image_url"] = d.ImageURL
		}
		if d.MaintenanceHashtag != "" {
			item["maintenance_hashtag"] = d.MaintenanceHashtag
		}
		if d.LastCompletedAt != nil {
			item["last_completed_at"] = d.LastCompletedAt.Format(time.RFC3339)
		}

		// Add rules
		rules := make([]map[string]interface{}, len(d.Rules))
		for j, r := range d.Rules {
			rules[j] = map[string]interface{}{
				"id":              r.ID,
				"component_id":    r.ComponentID,
				"type":            r.Type,
				"threshold_value": r.ThresholdValue,
			}
		}
		item["rules"] = rules

		// Add progress
		progress := make([]map[string]interface{}, len(d.Progress))
		for j, p := range d.Progress {
			progress[j] = map[string]interface{}{
				"type":            p.Type,
				"threshold_value": p.ThresholdValue,
				"current_value":   p.CurrentValue,
				"percent":         p.Percent,
				"due":             p.Due,
			}
		}
		item["progress"] = progress

		items[i] = item
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": items,
	})
}

// getGearComponents returns components for a specific gear item
// Called from JS: goStorage.getGearComponents(filtersJSON)
func getGearComponents(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getGearComponents")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing filters"))
	}

	var req struct {
		GearID  string `json:"gear_id"`
		Page    int    `json:"page"`
		PerPage int    `json:"per_page"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing filters: %w", err))
	}

	if req.GearID == "" {
		return errorJSON(fmt.Errorf("gear_id is required"))
	}

	ctx := context.Background()
	filters := storage.ComponentFilters{}
	filters.Page = req.Page
	filters.PerPage = req.PerPage

	result, err := maintenance.ListComponentsPaginated(ctx, athleteID, req.GearID, filters)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	items := make([]map[string]interface{}, len(result.Items))
	for i, c := range result.Items {
		item := map[string]interface{}{
			"id":         c.ID,
			"gear_id":    c.GearID,
			"name":       c.Name,
			"created_at": c.CreatedAt.Format(time.RFC3339),
			"updated_at": c.UpdatedAt.Format(time.RFC3339),
		}
		if c.ImageURL != "" {
			item["image_url"] = c.ImageURL
		}
		if c.MaintenanceHashtag != "" {
			item["maintenance_hashtag"] = c.MaintenanceHashtag
		}
		if c.LastCompletedAt != nil {
			item["last_completed_at"] = c.LastCompletedAt.Format(time.RFC3339)
		}

		// Add rules
		rules := make([]map[string]interface{}, len(c.Rules))
		for j, r := range c.Rules {
			rules[j] = map[string]interface{}{
				"id":              r.ID,
				"component_id":    r.ComponentID,
				"type":            r.Type,
				"threshold_value": r.ThresholdValue,
			}
		}
		item["rules"] = rules

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

// createComponent creates a new component for a gear item
// Called from JS: goStorage.createComponent(componentJSON)
func createComponent(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("createComponent")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing component data"))
	}

	var req struct {
		GearID             string `json:"gear_id"`
		Name               string `json:"name"`
		ImageURL           string `json:"image_url"`
		MaintenanceHashtag string `json:"maintenance_hashtag"`
		Rules              []struct {
			Type           string  `json:"type"`
			ThresholdValue float64 `json:"threshold_value"`
		} `json:"rules"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing component: %w", err))
	}

	if req.GearID == "" {
		return errorJSON(fmt.Errorf("gear_id is required"))
	}

	input := storage.CreateComponentInput{
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
	}
	for _, r := range req.Rules {
		input.Rules = append(input.Rules, storage.CreateRuleInput{
			Type:           r.Type,
			ThresholdValue: r.ThresholdValue,
		})
	}

	ctx := context.Background()
	comp, err := maintenance.CreateComponent(ctx, athleteID, req.GearID, input)
	if err != nil {
		return errorJSON(err)
	}
	if comp == nil {
		return errorJSON(fmt.Errorf("gear not found or not owned"))
	}

	return toJSON(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"id":      comp.ID,
			"gear_id": comp.GearID,
			"name":    comp.Name,
		},
	})
}

// updateComponent updates an existing component
// Called from JS: goStorage.updateComponent(componentJSON)
func updateComponent(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateComponent")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing component data"))
	}

	var req struct {
		ID                 int64   `json:"id"`
		Name               *string `json:"name"`
		ImageURL           *string `json:"image_url"`
		MaintenanceHashtag *string `json:"maintenance_hashtag"`
		Rules              *[]struct {
			Type           string  `json:"type"`
			ThresholdValue float64 `json:"threshold_value"`
		} `json:"rules"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing component: %w", err))
	}

	if req.ID == 0 {
		return errorJSON(fmt.Errorf("id is required"))
	}

	input := storage.UpdateComponentInput{
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
	}
	if req.Rules != nil {
		rules := make([]storage.CreateRuleInput, len(*req.Rules))
		for i, r := range *req.Rules {
			rules[i] = storage.CreateRuleInput{
				Type:           r.Type,
				ThresholdValue: r.ThresholdValue,
			}
		}
		input.Rules = &rules
	}

	ctx := context.Background()
	comp, err := maintenance.UpdateComponent(ctx, athleteID, req.ID, input)
	if err != nil {
		return errorJSON(err)
	}
	if comp == nil {
		return errorJSON(fmt.Errorf("component not found or not owned"))
	}

	return toJSON(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"id":      comp.ID,
			"gear_id": comp.GearID,
			"name":    comp.Name,
		},
	})
}

// deleteComponent deletes a component
// Called from JS: goStorage.deleteComponent(id)
func deleteComponent(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("deleteComponent")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing component ID"))
	}

	id := int64(args[0].Int())
	ctx := context.Background()

	if err := maintenance.DeleteComponent(ctx, athleteID, id); err != nil {
		return errorJSON(err)
	}

	return successJSON("Component deleted")
}

// logMaintenance logs a maintenance event for a component
// Called from JS: goStorage.logMaintenance(logJSON)
func logMaintenance(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("logMaintenance")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing log data"))
	}

	var req struct {
		ComponentID int64  `json:"component_id"`
		ActivityID  *int64 `json:"activity_id"`
		CompletedAt string `json:"completed_at"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing log: %w", err))
	}

	if req.ComponentID == 0 {
		return errorJSON(fmt.Errorf("component_id is required"))
	}

	completedAt := time.Now()
	if req.CompletedAt != "" {
		t, err := time.Parse(time.RFC3339, req.CompletedAt)
		if err != nil {
			return errorJSON(fmt.Errorf("invalid completed_at: %w", err))
		}
		completedAt = t
	}

	ctx := context.Background()
	if err := maintenance.LogMaintenance(ctx, athleteID, req.ComponentID, req.ActivityID, completedAt); err != nil {
		return errorJSON(err)
	}

	return successJSON("Maintenance logged")
}

// ============================================================================
// Settings
// ============================================================================

// getAppSettings returns the athlete's app settings
// Called from JS: goStorage.getAppSettings()
func getAppSettings(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getAppSettings")

	ctx := context.Background()
	s, err := settings.Get(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": s,
	})
}

// updateAppSettings updates the athlete's app settings
// Called from JS: goStorage.updateAppSettings(settingsJSON)
func updateAppSettings(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateAppSettings")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing settings"))
	}

	var s storage.AthleteSettings
	if err := json.Unmarshal([]byte(args[0].String()), &s); err != nil {
		return errorJSON(fmt.Errorf("parsing settings: %w", err))
	}

	ctx := context.Background()
	if err := settings.Upsert(ctx, athleteID, s); err != nil {
		return errorJSON(err)
	}

	return successJSON("Settings updated")
}

// ============================================================================
// Custom Gear
// ============================================================================

// getCustomGear returns paginated custom gear for the athlete
// Called from JS: goStorage.getCustomGear(filtersJSON)
func getCustomGear(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getCustomGear")

	var req struct {
		IncludeRetired bool   `json:"include_retired"`
		Page           int    `json:"page"`
		PerPage        int    `json:"per_page"`
		OrderBy        string `json:"order_by"`
		OrderDir       string `json:"order_dir"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
	}

	ctx := context.Background()
	filters := storage.GearFilters{
		IncludeRetired: req.IncludeRetired,
	}
	filters.Page = req.Page
	filters.PerPage = req.PerPage
	filters.OrderBy = req.OrderBy
	filters.OrderDir = req.OrderDir

	result, err := gear.ListCustomPaginated(ctx, athleteID, filters)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	items := make([]map[string]interface{}, len(result.Items))
	for i, g := range result.Items {
		item := map[string]interface{}{
			"id":         g.ID,
			"athlete_id": g.AthleteID,
			"name":       g.Name,
			"primary":    g.Primary,
			"retired":    g.Retired,
			"distance":   g.Distance,
			"source":     g.Source,
		}
		if g.Hashtag != "" {
			item["hashtag"] = g.Hashtag
		}
		if g.PurchasePrice != nil {
			item["purchase_price"] = *g.PurchasePrice
		}
		if g.PurchaseCurrency != "" {
			item["purchase_currency"] = g.PurchaseCurrency
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

// createCustomGear creates a new custom gear item
// Called from JS: goStorage.createCustomGear(gearJSON)
func createCustomGear(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("createCustomGear")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing gear data"))
	}

	var req struct {
		Name             string   `json:"name"`
		Hashtag          string   `json:"hashtag"`
		Retired          bool     `json:"retired"`
		PurchasePrice    *float64 `json:"purchase_price"`
		PurchaseCurrency string   `json:"purchase_currency"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing gear: %w", err))
	}

	input := storage.CustomGearCreate{
		Name:             req.Name,
		Hashtag:          req.Hashtag,
		Retired:          req.Retired,
		PurchasePrice:    req.PurchasePrice,
		PurchaseCurrency: req.PurchaseCurrency,
	}

	ctx := context.Background()
	g, err := gear.CreateCustom(ctx, athleteID, input)
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"id":       g.ID,
			"name":     g.Name,
			"hashtag":  g.Hashtag,
			"retired":  g.Retired,
			"distance": g.Distance,
		},
	})
}

// updateCustomGear updates an existing custom gear item
// Called from JS: goStorage.updateCustomGear(gearJSON)
func updateCustomGear(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateCustomGear")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing gear data"))
	}

	var req struct {
		ID               string    `json:"id"`
		Name             *string   `json:"name"`
		Hashtag          *string   `json:"hashtag"`
		Retired          *bool     `json:"retired"`
		PurchasePrice    **float64 `json:"purchase_price"`
		PurchaseCurrency *string   `json:"purchase_currency"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing gear: %w", err))
	}

	if req.ID == "" {
		return errorJSON(fmt.Errorf("id is required"))
	}

	input := storage.CustomGearUpdate{
		Name:             req.Name,
		Hashtag:          req.Hashtag,
		Retired:          req.Retired,
		PurchasePrice:    req.PurchasePrice,
		PurchaseCurrency: req.PurchaseCurrency,
	}

	ctx := context.Background()
	g, err := gear.UpdateCustom(ctx, athleteID, req.ID, input)
	if err != nil {
		return errorJSON(err)
	}
	if g == nil {
		return errorJSON(fmt.Errorf("gear not found or not custom"))
	}

	return toJSON(map[string]interface{}{
		"ok": true,
		"data": map[string]interface{}{
			"id":       g.ID,
			"name":     g.Name,
			"hashtag":  g.Hashtag,
			"retired":  g.Retired,
			"distance": g.Distance,
		},
	})
}

// deleteCustomGear deletes a custom gear item
// Called from JS: goStorage.deleteCustomGear(deleteJSON)
func deleteCustomGear(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("deleteCustomGear")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing delete data"))
	}

	var req struct {
		ID    string `json:"id"`
		Force bool   `json:"force"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing request: %w", err))
	}

	if req.ID == "" {
		return errorJSON(fmt.Errorf("id is required"))
	}

	ctx := context.Background()
	hadActivities, err := gear.DeleteCustom(ctx, athleteID, req.ID, req.Force)
	if err != nil {
		return errorJSON(err)
	}

	if hadActivities && !req.Force {
		return toJSON(map[string]interface{}{
			"ok":             false,
			"has_activities": true,
			"message":        "Gear has activities. Set force=true to delete anyway.",
		})
	}

	return successJSON("Custom gear deleted")
}

// ============================================================================
// HR Zones
// ============================================================================

// getHrZoneDefinitions returns HR zone definitions for the athlete
// Called from JS: goStorage.getHrZoneDefinitions()
func getHrZoneDefinitions(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getHrZoneDefinitions")

	ctx := context.Background()
	defs, err := zones.ListHR(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	items := make([]map[string]interface{}, len(defs))
	for i, d := range defs {
		items[i] = map[string]interface{}{
			"sport_type":     d.SportType,
			"effective_from": d.EffectiveFrom,
			"method":         d.Method,
			"zones":          json.RawMessage(d.Zones),
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": items,
	})
}

// upsertHrZoneDefinition creates or updates an HR zone definition
// Called from JS: goStorage.upsertHrZoneDefinition(zoneJSON)
func upsertHrZoneDefinition(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("upsertHrZoneDefinition")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing zone data"))
	}

	var req struct {
		SportType     string          `json:"sport_type"`
		EffectiveFrom string          `json:"effective_from"`
		Method        string          `json:"method"`
		Zones         json.RawMessage `json:"zones"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing zone: %w", err))
	}

	def := storage.HRZoneDefinition{
		SportType:     req.SportType,
		EffectiveFrom: req.EffectiveFrom,
		Method:        req.Method,
		Zones:         req.Zones,
	}

	ctx := context.Background()
	if err := zones.UpsertHR(ctx, athleteID, def); err != nil {
		return errorJSON(err)
	}

	return successJSON("HR zone definition saved")
}

// deleteHrZoneDefinition deletes an HR zone definition
// Called from JS: goStorage.deleteHrZoneDefinition(deleteJSON)
func deleteHrZoneDefinition(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("deleteHrZoneDefinition")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing delete data"))
	}

	var req struct {
		SportType     string `json:"sport_type"`
		EffectiveFrom string `json:"effective_from"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing request: %w", err))
	}

	ctx := context.Background()
	if err := zones.DeleteHR(ctx, athleteID, req.SportType, req.EffectiveFrom); err != nil {
		return errorJSON(err)
	}

	return successJSON("HR zone definition deleted")
}
