//go:build js && wasm

package main

import (
	"fmt"
)

// Standard durations for power curve (matching Go backend and TS importer)
var powerDurations = []int{5, 10, 30, 60, 300, 480, 1200, 3600}

// ============================================================================
// Power Functions
// ============================================================================

//wasm:category Power

// computePowerBestEfforts computes and stores power best efforts for an activity
// Called from JS: goStorage.computePowerBestEfforts(dataJSON)
//
//wasm:export
var computePowerBestEfforts = wrapWasm("computePowerBestEfforts", func(wc *WasmContext) interface{} {
	var req struct {
		ActivityID int64 `json:"activity_id"`
		AthleteID  int64 `json:"athlete_id"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing power request: %w", err))
	}

	if err := wc.Registry.Power().EnsureActivityComputed(wc.Ctx, req.AthleteID, req.ActivityID, powerDurations); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Power best efforts computed for activity %d", req.ActivityID))
})
