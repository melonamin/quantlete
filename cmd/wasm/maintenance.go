//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
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
//
//wasm:export
var getMaintenanceDue = wrapWasmAthlete("getMaintenanceDue", func(wc *WasmContext) interface{} {
	items, err := wc.Registry.MaintenanceService.ListDue(wc.Ctx, wc.AthleteID)
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(items)
})

// getGearComponents returns components for a specific gear item
// Called from JS: goStorage.getGearComponents(filtersJSON)
// NOTE: Replaced by generated adapter genGetGearComponents
var getGearComponents = wrapWasmAthlete("getGearComponents", func(wc *WasmContext) interface{} {
	var req struct {
		GearID  string `json:"gear_id"`
		Page    int    `json:"page"`
		PerPage int    `json:"per_page"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing filters: %w", err))
	}

	result, err := wc.Registry.MaintenanceService.ListComponents(wc.Ctx, services.ListComponentsInput{
		AthleteID: wc.AthleteID,
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
})

// createComponent creates a new component for a gear item
// Called from JS: goStorage.createComponent(componentJSON)
// NOTE: Replaced by generated adapter genCreateComponent
var createComponent = wrapWasmAthlete("createComponent", func(wc *WasmContext) interface{} {
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
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing component: %w", err))
	}

	rules := make([]services.RuleInput, 0, len(req.Rules))
	for _, r := range req.Rules {
		rules = append(rules, services.RuleInput{
			Type:           r.Type,
			ThresholdValue: r.ThresholdValue,
		})
	}

	comp, err := wc.Registry.MaintenanceService.CreateComponent(wc.Ctx, services.CreateComponentInput{
		AthleteID:          wc.AthleteID,
		GearID:             req.GearID,
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(comp)
})

// updateComponent updates an existing component
// Called from JS: goStorage.updateComponent(componentJSON)
// NOTE: Replaced by generated adapter genUpdateComponent
var updateComponent = wrapWasmAthlete("updateComponent", func(wc *WasmContext) interface{} {
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
	if err := wc.ArgJSON(0, &req); err != nil {
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

	comp, err := wc.Registry.MaintenanceService.UpdateComponent(wc.Ctx, services.UpdateComponentInput{
		AthleteID:          wc.AthleteID,
		ComponentID:        req.ID,
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(comp)
})

// deleteComponent deletes a component
// Called from JS: goStorage.deleteComponent(id)
//
//wasm:export
var deleteComponent = wrapWasmAthlete("deleteComponent", func(wc *WasmContext) interface{} {
	id := wc.ArgInt64(0)
	if err := wc.Registry.MaintenanceService.DeleteComponent(wc.Ctx, wc.AthleteID, id); err != nil {
		return errorJSON(err)
	}
	return successJSON("Component deleted")
})

// logMaintenance logs a maintenance event for a component
// Called from JS: goStorage.logMaintenance(logJSON)
//
//wasm:export
var logMaintenance = wrapWasmAthlete("logMaintenance", func(wc *WasmContext) interface{} {
	var req struct {
		ComponentID int64  `json:"component_id"`
		ActivityID  *int64 `json:"activity_id"`
		CompletedAt string `json:"completed_at"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
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

	if err := wc.Registry.MaintenanceService.LogMaintenance(wc.Ctx, services.LogMaintenanceInput{
		AthleteID:   wc.AthleteID,
		ComponentID: req.ComponentID,
		ActivityID:  req.ActivityID,
		CompletedAt: completedAt,
	}); err != nil {
		return errorJSON(err)
	}

	return successJSON("Maintenance logged")
})

// ============================================================================
// Settings
// ============================================================================

//wasm:category Settings

// getAppSettings returns the athlete's app settings
// Called from JS: goStorage.getAppSettings()
//
//wasm:export
var getAppSettings = wrapWasmAthlete("getAppSettings", func(wc *WasmContext) interface{} {
	s, err := wc.Registry.Settings().Get(wc.Ctx, wc.AthleteID)
	if err != nil {
		return errorJSON(err)
	}
	return dataJSON(s)
})

// updateAppSettings updates the athlete's app settings
// Called from JS: goStorage.updateAppSettings(settingsJSON)
//
//wasm:export
var updateAppSettings = wrapWasmAthlete("updateAppSettings", func(wc *WasmContext) interface{} {
	var s storage.AthleteSettings
	if err := wc.ArgJSON(0, &s); err != nil {
		return errorJSON(fmt.Errorf("parsing settings: %w", err))
	}

	if err := wc.Registry.Settings().Upsert(wc.Ctx, wc.AthleteID, s); err != nil {
		return errorJSON(err)
	}

	return successJSON("Settings updated")
})

// ============================================================================
// Custom Gear
// ============================================================================

//wasm:category Custom Gear

// getCustomGear returns paginated custom gear for the athlete
// Called from JS: goStorage.getCustomGear(filtersJSON)
//
//wasm:export
var getCustomGear = wrapWasmAthlete("getCustomGear", func(wc *WasmContext) interface{} {
	var req struct {
		IncludeRetired bool   `json:"include_retired"`
		Page           int    `json:"page"`
		PerPage        int    `json:"per_page"`
		OrderBy        string `json:"order_by"`
		OrderDir       string `json:"order_dir"`
	}
	if wc.HasArg(0) && wc.ArgString(0) != "" {
		if err := wc.ArgJSON(0, &req); err != nil {
			return errorJSON(fmt.Errorf("parsing filters: %w", err))
		}
	}

	result, err := wc.Registry.GearService.ListCustom(wc.Ctx, services.ListGearInput{
		AthleteID:      wc.AthleteID,
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
})

// createCustomGear creates a new custom gear item
// Called from JS: goStorage.createCustomGear(gearJSON)
//
//wasm:export
var createCustomGear = wrapWasmAthlete("createCustomGear", func(wc *WasmContext) interface{} {
	var req struct {
		Name             string   `json:"name"`
		Hashtag          string   `json:"hashtag"`
		Retired          bool     `json:"retired"`
		PurchasePrice    *float64 `json:"purchase_price"`
		PurchaseCurrency string   `json:"purchase_currency"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing gear: %w", err))
	}

	result, err := wc.Registry.GearService.CreateCustom(wc.Ctx, services.CreateCustomGearInput{
		AthleteID:        wc.AthleteID,
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
})

// updateCustomGear updates an existing custom gear item
// Called from JS: goStorage.updateCustomGear(gearJSON)
//
//wasm:export
var updateCustomGear = wrapWasmAthlete("updateCustomGear", func(wc *WasmContext) interface{} {
	var req struct {
		ID               string    `json:"id"`
		Name             *string   `json:"name"`
		Hashtag          *string   `json:"hashtag"`
		Retired          *bool     `json:"retired"`
		PurchasePrice    **float64 `json:"purchase_price"`
		PurchaseCurrency *string   `json:"purchase_currency"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing gear: %w", err))
	}

	result, err := wc.Registry.GearService.UpdateCustom(wc.Ctx, services.UpdateCustomGearInput{
		AthleteID:        wc.AthleteID,
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
})

// deleteCustomGear deletes a custom gear item
// Called from JS: goStorage.deleteCustomGear(deleteJSON)
//
//wasm:export
var deleteCustomGear = wrapWasmAthlete("deleteCustomGear", func(wc *WasmContext) interface{} {
	var req struct {
		ID    string `json:"id"`
		Force bool   `json:"force"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing request: %w", err))
	}

	result, err := wc.Registry.GearService.DeleteCustom(wc.Ctx, services.DeleteCustomGearInput{
		AthleteID: wc.AthleteID,
		GearID:    req.ID,
		Force:     req.Force,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
})

// ============================================================================
// HR Zones
// ============================================================================

//wasm:category HR Zones

// getHrZoneDefinitions returns HR zone definitions for the athlete
// Called from JS: goStorage.getHrZoneDefinitions()
//
//wasm:export
var getHrZoneDefinitions = wrapWasmAthlete("getHrZoneDefinitions", func(wc *WasmContext) interface{} {
	defs, err := wc.Registry.Zones().ListHR(wc.Ctx, wc.AthleteID)
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

	return dataJSON(items)
})

// upsertHrZoneDefinition creates or updates an HR zone definition
// Called from JS: goStorage.upsertHrZoneDefinition(zoneJSON)
//
//wasm:export
var upsertHrZoneDefinition = wrapWasmAthlete("upsertHrZoneDefinition", func(wc *WasmContext) interface{} {
	var hrZoneReq struct {
		SportType     string          `json:"sport_type"`
		EffectiveFrom string          `json:"effective_from"`
		Method        string          `json:"method"`
		Zones         json.RawMessage `json:"zones"`
	}
	if err := wc.ArgJSON(0, &hrZoneReq); err != nil {
		return errorJSON(fmt.Errorf("parsing zone: %w", err))
	}

	def := storage.HRZoneDefinition{
		SportType:     hrZoneReq.SportType,
		EffectiveFrom: hrZoneReq.EffectiveFrom,
		Method:        hrZoneReq.Method,
		Zones:         hrZoneReq.Zones,
	}

	if err := wc.Registry.Zones().UpsertHR(wc.Ctx, wc.AthleteID, def); err != nil {
		return errorJSON(err)
	}

	return successJSON("HR zone definition saved")
})

// deleteHrZoneDefinition deletes an HR zone definition
// Called from JS: goStorage.deleteHrZoneDefinition(deleteJSON)
//
//wasm:export
var deleteHrZoneDefinition = wrapWasmAthlete("deleteHrZoneDefinition", func(wc *WasmContext) interface{} {
	var hrZoneDeleteReq struct {
		SportType     string `json:"sport_type"`
		EffectiveFrom string `json:"effective_from"`
	}
	if err := wc.ArgJSON(0, &hrZoneDeleteReq); err != nil {
		return errorJSON(fmt.Errorf("parsing request: %w", err))
	}

	if err := wc.Registry.Zones().DeleteHR(wc.Ctx, wc.AthleteID, hrZoneDeleteReq.SportType, hrZoneDeleteReq.EffectiveFrom); err != nil {
		return errorJSON(err)
	}

	return successJSON("HR zone definition deleted")
})
