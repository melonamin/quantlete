//go:build js && wasm

package main

import (
	"fmt"
)

// ============================================================================
// Strava Credentials
// ============================================================================

//wasm:category App State - Credentials

// getStravaCredentials returns stored Strava API credentials
// Called from JS: goStorage.getStravaCredentials()
//
//wasm:export
var getStravaCredentials = wrapWasm("getStravaCredentials", func(wc *WasmContext) interface{} {
	clientID, err := wc.Registry.AppState().Get(wc.Ctx, "strava_client_id")
	if err != nil {
		return errorJSON(err)
	}

	clientSecret, err := wc.Registry.AppState().Get(wc.Ctx, "strava_client_secret")
	if err != nil {
		return errorJSON(err)
	}

	if clientID == "" || clientSecret == "" {
		return dataJSON(nil)
	}

	return dataJSON(map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
	})
})

// saveStravaCredentials stores Strava API credentials
// Called from JS: goStorage.saveStravaCredentials(credentialsJSON)
//
//wasm:export
var saveStravaCredentials = wrapWasm("saveStravaCredentials", func(wc *WasmContext) interface{} {
	var req struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := wc.ArgJSON(0, &req); err != nil {
		return errorJSON(fmt.Errorf("parsing credentials: %w", err))
	}

	if req.ClientID == "" || req.ClientSecret == "" {
		return errorJSON(fmt.Errorf("client_id and client_secret are required"))
	}

	if err := wc.Registry.AppState().Set(wc.Ctx, "strava_client_id", req.ClientID); err != nil {
		return errorJSON(err)
	}

	if err := wc.Registry.AppState().Set(wc.Ctx, "strava_client_secret", req.ClientSecret); err != nil {
		return errorJSON(err)
	}

	return successJSON("Credentials saved")
})

// deleteStravaCredentials removes stored Strava API credentials
// Called from JS: goStorage.deleteStravaCredentials()
//
//wasm:export
var deleteStravaCredentials = wrapWasm("deleteStravaCredentials", func(wc *WasmContext) interface{} {
	if err := wc.Registry.AppState().Delete(wc.Ctx, "strava_client_id"); err != nil {
		return errorJSON(err)
	}

	if err := wc.Registry.AppState().Delete(wc.Ctx, "strava_client_secret"); err != nil {
		return errorJSON(err)
	}

	return successJSON("Credentials deleted")
})

// hasStravaCredentials checks if credentials are configured
// Called from JS: goStorage.hasStravaCredentials()
//
//wasm:export
var hasStravaCredentials = wrapWasm("hasStravaCredentials", func(wc *WasmContext) interface{} {
	clientID, _ := wc.Registry.AppState().Get(wc.Ctx, "strava_client_id")
	clientSecret, _ := wc.Registry.AppState().Get(wc.Ctx, "strava_client_secret")

	return dataJSON(map[string]bool{
		"has_credentials": clientID != "" && clientSecret != "",
	})
})
