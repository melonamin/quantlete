//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Maintenance
// ============================================================================

//wasm:category Maintenance

// getMaintenanceDue returns components with maintenance status
// Called from JS: goStorage.getMaintenanceDue()
//wasm:export
func getMaintenanceDue(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getMaintenanceDue")

	ctx := context.Background()
	items, err := bridge.maintenanceService.ListDue(ctx, bridge.athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": items,
	})
}

// getGearComponents returns components for a specific gear item
// Called from JS: goStorage.getGearComponents(filtersJSON)
//wasm:export
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

	ctx := context.Background()
	result, err := bridge.maintenanceService.ListComponents(ctx, services.ListComponentsInput{
		AthleteID: bridge.athleteID,
		GearID:    req.GearID,
		Page:      req.Page,
		PerPage:   req.PerPage,
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

// createComponent creates a new component for a gear item
// Called from JS: goStorage.createComponent(componentJSON)
//wasm:export
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

	rules := make([]services.RuleInput, 0, len(req.Rules))
	for _, r := range req.Rules {
		rules = append(rules, services.RuleInput{
			Type:           r.Type,
			ThresholdValue: r.ThresholdValue,
		})
	}

	ctx := context.Background()
	comp, err := bridge.maintenanceService.CreateComponent(ctx, services.CreateComponentInput{
		AthleteID:          bridge.athleteID,
		GearID:             req.GearID,
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": comp,
	})
}

// updateComponent updates an existing component
// Called from JS: goStorage.updateComponent(componentJSON)
//wasm:export
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

	var rules *[]services.RuleInput
	if req.Rules != nil {
		r := make([]services.RuleInput, len(*req.Rules))
		for i, rule := range *req.Rules {
			r[i] = services.RuleInput{
				Type:           rule.Type,
				ThresholdValue: rule.ThresholdValue,
			}
		}
		rules = &r
	}

	ctx := context.Background()
	comp, err := bridge.maintenanceService.UpdateComponent(ctx, services.UpdateComponentInput{
		AthleteID:          bridge.athleteID,
		ComponentID:        req.ID,
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": comp,
	})
}

// deleteComponent deletes a component
// Called from JS: goStorage.deleteComponent(id)
//wasm:export
func deleteComponent(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("deleteComponent")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing component ID"))
	}

	id := int64(args[0].Int())
	ctx := context.Background()

	if err := bridge.maintenanceService.DeleteComponent(ctx, bridge.athleteID, id); err != nil {
		return errorJSON(err)
	}

	return successJSON("Component deleted")
}

// logMaintenance logs a maintenance event for a component
// Called from JS: goStorage.logMaintenance(logJSON)
//wasm:export
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

	completedAt := time.Now()
	if req.CompletedAt != "" {
		t, err := time.Parse(time.RFC3339, req.CompletedAt)
		if err != nil {
			return errorJSON(fmt.Errorf("invalid completed_at: %w", err))
		}
		completedAt = t
	}

	ctx := context.Background()
	if err := bridge.maintenanceService.LogMaintenance(ctx, services.LogMaintenanceInput{
		AthleteID:   bridge.athleteID,
		ComponentID: req.ComponentID,
		ActivityID:  req.ActivityID,
		CompletedAt: completedAt,
	}); err != nil {
		return errorJSON(err)
	}

	return successJSON("Maintenance logged")
}

// ============================================================================
// Settings
// ============================================================================

//wasm:category Settings

// getAppSettings returns the athlete's app settings
// Called from JS: goStorage.getAppSettings()
//wasm:export
func getAppSettings(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getAppSettings")

	ctx := context.Background()
	s, err := bridge.settings.Get(ctx, bridge.athleteID)
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
//wasm:export
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
	if err := bridge.settings.Upsert(ctx, bridge.athleteID, s); err != nil {
		return errorJSON(err)
	}

	return successJSON("Settings updated")
}

// ============================================================================
// Custom Gear
// ============================================================================

//wasm:category Custom Gear

// getCustomGear returns paginated custom gear for the athlete
// Called from JS: goStorage.getCustomGear(filtersJSON)
//wasm:export
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
	result, err := bridge.gearService.ListCustom(ctx, services.ListGearInput{
		AthleteID:      bridge.athleteID,
		IncludeRetired: req.IncludeRetired,
		Page:           req.Page,
		PerPage:        req.PerPage,
		OrderBy:        req.OrderBy,
		OrderDir:       req.OrderDir,
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

// createCustomGear creates a new custom gear item
// Called from JS: goStorage.createCustomGear(gearJSON)
//wasm:export
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

	ctx := context.Background()
	result, err := bridge.gearService.CreateCustom(ctx, services.CreateCustomGearInput{
		AthleteID:        bridge.athleteID,
		Name:             req.Name,
		Hashtag:          req.Hashtag,
		Retired:          req.Retired,
		PurchasePrice:    req.PurchasePrice,
		PurchaseCurrency: req.PurchaseCurrency,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// updateCustomGear updates an existing custom gear item
// Called from JS: goStorage.updateCustomGear(gearJSON)
//wasm:export
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

	ctx := context.Background()
	result, err := bridge.gearService.UpdateCustom(ctx, services.UpdateCustomGearInput{
		AthleteID:        bridge.athleteID,
		GearID:           req.ID,
		Name:             req.Name,
		Hashtag:          req.Hashtag,
		Retired:          req.Retired,
		PurchasePrice:    req.PurchasePrice,
		PurchaseCurrency: req.PurchaseCurrency,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// deleteCustomGear deletes a custom gear item
// Called from JS: goStorage.deleteCustomGear(deleteJSON)
//wasm:export
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

	ctx := context.Background()
	result, err := bridge.gearService.DeleteCustom(ctx, services.DeleteCustomGearInput{
		AthleteID: bridge.athleteID,
		GearID:    req.ID,
		Force:     req.Force,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// ============================================================================
// HR Zones
// ============================================================================

//wasm:category HR Zones

// getHrZoneDefinitions returns HR zone definitions for the athlete
// Called from JS: goStorage.getHrZoneDefinitions()
//wasm:export
func getHrZoneDefinitions(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getHrZoneDefinitions")

	ctx := context.Background()
	defs, err := bridge.zones.ListHR(ctx, bridge.athleteID)
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
//wasm:export
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
	if err := bridge.zones.UpsertHR(ctx, bridge.athleteID, def); err != nil {
		return errorJSON(err)
	}

	return successJSON("HR zone definition saved")
}

// deleteHrZoneDefinition deletes an HR zone definition
// Called from JS: goStorage.deleteHrZoneDefinition(deleteJSON)
//wasm:export
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
	if err := bridge.zones.DeleteHR(ctx, bridge.athleteID, req.SportType, req.EffectiveFrom); err != nil {
		return errorJSON(err)
	}

	return successJSON("HR zone definition deleted")
}
