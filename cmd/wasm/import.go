//go:build js && wasm

// Package main provides WASM bridge functionality.
//
// This file exposes the Go importer to JavaScript via WASM bridge functions.
// The importer uses WasmStravaAdapter and WasmStorageAdapter for platform-agnostic operation.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/importer"
)

// importerState holds the global importer instance and associated state.
var importerState struct {
	mu          sync.Mutex
	importer    *importer.Importer
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
	unsubscribe func()
}

// ============================================================================
// Import Control
// ============================================================================

//wasm:category Import

// startImport starts a new import process.
// Called from JS: goStorage.startImport(optionsJSON, athleteJSON)
//
//wasm:export
var startImport = wrapWasmRaw("startImport", func(this js.Value, args []js.Value) interface{} {
	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}

	if len(args) < 2 {
		return errorJSON(fmt.Errorf("missing options or athlete JSON"))
	}

	// Parse options
	optionsJSON := args[0].String()
	var opts struct {
		FullSync        bool `json:"full_sync"`
		Resume          bool `json:"resume"`
		SkipStreams     bool `json:"skip_streams"`
		SkipSegments    bool `json:"skip_segments"`
		SkipBestEfforts bool `json:"skip_best_efforts"`
		SkipPhotos      bool `json:"skip_photos"`
	}
	if err := json.Unmarshal([]byte(optionsJSON), &opts); err != nil {
		return errorJSON(fmt.Errorf("parsing options: %w", err))
	}

	// Parse athlete
	athleteJSON := args[1].String()
	var athlete importer.Athlete
	if err := json.Unmarshal([]byte(athleteJSON), &athlete); err != nil {
		return errorJSON(fmt.Errorf("parsing athlete: %w", err))
	}

	importerState.mu.Lock()
	defer importerState.mu.Unlock()

	if importerState.running {
		return errorJSON(fmt.Errorf("import already running"))
	}

	// Create adapters
	stravaAdapter := NewWasmStravaAdapter()
	stravaAdapter.SetAthlete(&athlete)
	storageAdapter := NewWasmStorageAdapter()

	// Create importer
	imp, err := importer.New(stravaAdapter, storageAdapter)
	if err != nil {
		return errorJSON(fmt.Errorf("creating importer: %w", err))
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	importerState.ctx = ctx
	importerState.cancel = cancel
	importerState.importer = imp
	importerState.running = true

	// Subscribe to events and forward progress to JS
	eventCh := make(chan importer.Event, 10)
	importerState.unsubscribe = imp.Subscribe(eventCh)

	// Event listener goroutine
	go func() {
		for event := range eventCh {
			switch event.Type {
			case importer.EventSyncProgress:
				if progress, ok := event.Data.(*importer.Progress); ok {
					sendProgressToJS(progress)
				}
			case importer.EventSyncComplete:
				// Completion is handled in the import goroutine
			}
		}
	}()

	// Start import in a goroutine
	go func() {
		importOpts := importer.ImportOptions{
			FullSync:        opts.FullSync,
			Resume:          opts.Resume,
			SkipStreams:     opts.SkipStreams,
			SkipSegments:    opts.SkipSegments,
			SkipBestEfforts: opts.SkipBestEfforts,
			SkipPhotos:      opts.SkipPhotos,
		}

		err := imp.Start(ctx, importOpts)

		importerState.mu.Lock()
		importerState.running = false
		if importerState.unsubscribe != nil {
			importerState.unsubscribe()
			importerState.unsubscribe = nil
		}
		importerState.mu.Unlock()

		// Send completion event to JS
		if err != nil {
			sendImportCompleteToJS(false, err.Error())
		} else {
			sendImportCompleteToJS(true, "")
		}
	}()

	return successJSON("import started")
})

// cancelImport cancels the currently running import.
// Called from JS: goStorage.cancelImport()
//
//wasm:export
var cancelImport = wrapWasmRaw("cancelImport", func(this js.Value, args []js.Value) interface{} {
	importerState.mu.Lock()
	defer importerState.mu.Unlock()

	if !importerState.running {
		return errorJSON(fmt.Errorf("no import running"))
	}

	if importerState.importer != nil {
		importerState.importer.Cancel()
	}
	if importerState.cancel != nil {
		importerState.cancel()
	}

	return successJSON("import canceled")
})

// getImportProgress returns the current import progress.
// Called from JS: goStorage.getImportProgress()
//
//wasm:export
var getImportProgress = wrapWasmRaw("getImportProgress", func(this js.Value, args []js.Value) interface{} {
	importerState.mu.Lock()
	defer importerState.mu.Unlock()

	if importerState.importer == nil {
		return dataJSON(nil)
	}

	progress := importerState.importer.Progress()
	return dataJSON(formatProgress(&progress))
})

// isImportRunning returns whether an import is currently running.
// Called from JS: goStorage.isImportRunning()
//
//wasm:export
var isImportRunning = wrapWasmRaw("isImportRunning", func(this js.Value, args []js.Value) interface{} {
	importerState.mu.Lock()
	defer importerState.mu.Unlock()

	return toJSON(map[string]interface{}{
		"ok":      true,
		"running": importerState.running,
	})
})

// getImportState loads the saved import state for resume capability.
// Called from JS: goStorage.getImportState()
//
//wasm:export
var getImportState = wrapWasmRaw("getImportState", func(this js.Value, args []js.Value) interface{} {
	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	storageAdapter := NewWasmStorageAdapter()
	state, err := storageAdapter.LoadImportState(ctx)
	if err != nil {
		return errorJSON(err)
	}

	if state == nil {
		return dataJSON(nil)
	}

	return dataJSON(formatImportState(state))
})

// clearImportState clears the saved import state.
// Called from JS: goStorage.clearImportState()
//
//wasm:export
var clearImportState = wrapWasmRaw("clearImportState", func(this js.Value, args []js.Value) interface{} {
	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}

	ctx := context.Background()
	storageAdapter := NewWasmStorageAdapter()
	if err := storageAdapter.ClearImportState(ctx); err != nil {
		return errorJSON(err)
	}

	return successJSON("import state cleared")
})

// ============================================================================
// Helper Functions
// ============================================================================

// sendProgressToJS sends a progress update to JavaScript.
func sendProgressToJS(p *importer.Progress) {
	onImportProgress := js.Global().Get("onImportProgress")
	if !onImportProgress.Truthy() {
		return
	}

	progressData := formatProgress(p)
	jsonStr, err := json.Marshal(progressData)
	if err != nil {
		return
	}

	onImportProgress.Invoke(string(jsonStr))
}

// sendImportCompleteToJS sends a completion event to JavaScript.
func sendImportCompleteToJS(success bool, errMsg string) {
	onImportComplete := js.Global().Get("onImportComplete")
	if !onImportComplete.Truthy() {
		return
	}

	data := map[string]interface{}{
		"success": success,
	}
	if errMsg != "" {
		data["error"] = errMsg
	}

	jsonStr, err := json.Marshal(data)
	if err != nil {
		return
	}

	onImportComplete.Invoke(string(jsonStr))
}

// formatProgress converts importer.Progress to a map for JSON serialization.
func formatProgress(p *importer.Progress) map[string]interface{} {
	if p == nil {
		return nil
	}

	result := map[string]interface{}{
		"status":                 string(p.Status),
		"phase":                  string(p.Phase),
		"activities_total":       p.ActivitiesTotal,
		"activities_done":        p.ActivitiesDone,
		"gear_total":             p.GearTotal,
		"gear_done":              p.GearDone,
		"streams_total":          p.StreamsTotal,
		"streams_done":           p.StreamsDone,
		"details_total":          p.DetailsTotal,
		"details_done":           p.DetailsDone,
		"segments_total":         p.SegmentsTotal,
		"segments_done":          p.SegmentsDone,
		"photos_total":           p.PhotosTotal,
		"photos_done":            p.PhotosDone,
		"failed_count":           p.FailedCount,
		"remaining_api_calls":    p.RemainingAPICalls,
		"estimated_eta":          p.EstimatedETA,
		"rate_limit_used_15min":  p.RateLimitUsed15Min,
		"rate_limit_limit_15min": p.RateLimitLimit15Min,
		"rate_limit_used_daily":  p.RateLimitUsedDaily,
		"rate_limit_limit_daily": p.RateLimitLimitDaily,
		"waiting_for_rate_limit": p.WaitingForRateLimit,
	}

	if !p.StartedAt.IsZero() {
		result["started_at"] = p.StartedAt.Format(time.RFC3339)
	}
	if !p.WaitingUntil.IsZero() {
		result["waiting_until"] = p.WaitingUntil.Format(time.RFC3339)
	}
	if p.WaitingReason != "" {
		result["waiting_reason"] = p.WaitingReason
	}
	if len(p.Errors) > 0 {
		result["errors"] = p.Errors
	}

	return result
}

// formatImportState converts importer.ImportState to a map for JSON serialization.
func formatImportState(s *importer.ImportState) map[string]interface{} {
	if s == nil {
		return nil
	}

	result := map[string]interface{}{
		"phase":                string(s.Phase),
		"activities_total":     s.ActivitiesTotal,
		"activities_done":      s.ActivitiesDone,
		"gear_total":           s.GearTotal,
		"gear_done":            s.GearDone,
		"streams_total":        s.StreamsTotal,
		"streams_done":         s.StreamsDone,
		"details_total":        s.DetailsTotal,
		"details_done":         s.DetailsDone,
		"segments_total":       s.SegmentsTotal,
		"segments_done":        s.SegmentsDone,
		"photos_total":         s.PhotosTotal,
		"photos_done":          s.PhotosDone,
		"failed_count":         s.FailedCount,
		"skip_streams":         s.SkipStreams,
		"skip_segments":        s.SkipSegments,
		"skip_best_efforts":    s.SkipBestEfforts,
		"skip_photos":          s.SkipPhotos,
		"activity_ids_count":   len(s.ActivityIDs),
		"gear_ids_count":       len(s.GearIDs),
		"segment_ids_count":    len(s.SegmentIDsToFetch),
		"activities_last_page": s.ActivitiesLastPage,
	}

	if !s.StartedAt.IsZero() {
		result["started_at"] = s.StartedAt.Format(time.RFC3339)
	}
	if len(s.Errors) > 0 {
		result["errors"] = s.Errors
	}
	if s.AfterDate != nil {
		result["after_date"] = s.AfterDate.Format(time.RFC3339)
	}
	if s.NewestActivityDate != nil {
		result["newest_activity_date"] = s.NewestActivityDate.Format(time.RFC3339)
	}

	return result
}
