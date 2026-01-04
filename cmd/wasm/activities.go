//go:build js && wasm

package main

// ============================================================================
// Auth
// ============================================================================

//wasm:category Auth

// getAuthStatus returns the current authentication status
// Called from JS: goStorage.getAuthStatus()
//
//wasm:export
var getAuthStatus = wrapWasm("getAuthStatus", func(wc *WasmContext) interface{} {
	// Check if we have an athlete
	if wc.Bridge.athleteID == 0 {
		return toJSON(map[string]interface{}{
			"authenticated": false,
		})
	}

	// Get athlete info
	athlete, err := wc.Registry.Athletes().GetByID(wc.Ctx, wc.Bridge.athleteID)
	if err != nil {
		return errorJSON(err)
	}

	if athlete == nil {
		return toJSON(map[string]interface{}{
			"authenticated": false,
		})
	}

	return toJSON(map[string]interface{}{
		"authenticated": true,
		"athlete": map[string]interface{}{
			"id":        athlete.ID,
			"username":  athlete.Username,
			"firstname": athlete.FirstName,
			"lastname":  athlete.LastName,
			"profile":   athlete.ProfileMedium,
		},
	})
})
