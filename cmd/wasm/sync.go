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
// Sync History Write
// ============================================================================

//wasm:category Sync History - Write

// createSyncRun creates a new sync history record
// Called from JS: goStorage.createSyncRun(dataJSON)
//wasm:export
func createSyncRun(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("createSyncRun")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing data JSON"))
	}

	dataJSON := args[0].String()
	var req struct {
		AthleteID       int64 `json:"athlete_id"`
		FullSync        bool  `json:"full_sync"`
		SkipStreams     bool  `json:"skip_streams"`
		SkipSegments    bool  `json:"skip_segments"`
		SkipBestEfforts bool  `json:"skip_best_efforts"`
		SkipPhotos      bool  `json:"skip_photos"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing sync run: %w", err))
	}

	ctx := context.Background()
	run, err := bridge.syncHistory.StartRun(ctx, req.AthleteID, storage.SyncRunOptions{
		FullSync:        req.FullSync,
		SkipStreams:     req.SkipStreams,
		SkipSegments:    req.SkipSegments,
		SkipBestEfforts: req.SkipBestEfforts,
		SkipPhotos:      req.SkipPhotos,
	})
	if err != nil {
		return errorJSON(err)
	}

	return toJSON(map[string]interface{}{
		"ok": true,
		"id": run.ID,
	})
}

// updateSyncRun updates a sync history record status (for pause/resume)
// Called from JS: goStorage.updateSyncRun(dataJSON)
//wasm:export
func updateSyncRun(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("updateSyncRun")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing data JSON"))
	}

	dataJSON := args[0].String()
	var req struct {
		ID                 int64  `json:"id"`
		Status             string `json:"status"`
		ActivitiesTotal    int    `json:"activities_total"`
		ActivitiesImported int    `json:"activities_imported"`
		StreamsImported    int    `json:"streams_imported"`
		FailedCount        int    `json:"failed_count"`
		NewestActivityDate string `json:"newest_activity_date"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing sync run update: %w", err))
	}

	// Parse newest activity date
	var newestDate *time.Time
	if req.NewestActivityDate != "" {
		if t, err := time.Parse(time.RFC3339, req.NewestActivityDate); err == nil {
			newestDate = &t
		}
	}

	counts := storage.SyncRunCounts{
		ActivitiesTotal:    req.ActivitiesTotal,
		ActivitiesImported: req.ActivitiesImported,
		StreamsImported:    req.StreamsImported,
		FailedCount:        req.FailedCount,
		NewestActivityDate: newestDate,
	}

	// Use the appropriate method based on status
	ctx := context.Background()
	var err error
	switch req.Status {
	case "canceled":
		err = bridge.syncHistory.CancelRun(ctx, req.ID, counts)
	case "paused":
		// For pause, just update the counts - no dedicated method, keep status as running
		// The TS side handles pause state
		return successJSON(fmt.Sprintf("Sync run %d paused", req.ID))
	default:
		return errorJSON(fmt.Errorf("unknown status for update: %s", req.Status))
	}

	if err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Sync run %d updated", req.ID))
}

// completeSyncRun marks a sync run as complete
// Called from JS: goStorage.completeSyncRun(dataJSON)
//wasm:export
func completeSyncRun(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("completeSyncRun")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing data JSON"))
	}

	dataJSON := args[0].String()
	var req struct {
		ID                 int64  `json:"id"`
		Status             string `json:"status"`
		Error              string `json:"error"`
		ActivitiesTotal    int    `json:"activities_total"`
		ActivitiesImported int    `json:"activities_imported"`
		ActivitiesSkipped  int    `json:"activities_skipped"`
		GearImported       int    `json:"gear_imported"`
		StreamsImported    int    `json:"streams_imported"`
		SegmentsImported   int    `json:"segments_imported"`
		PhotosImported     int    `json:"photos_imported"`
		FailedCount        int    `json:"failed_count"`
		NewestActivityDate string `json:"newest_activity_date"`
	}
	if err := json.Unmarshal([]byte(dataJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing sync run completion: %w", err))
	}

	// Parse newest activity date
	var newestDate *time.Time
	if req.NewestActivityDate != "" {
		if t, err := time.Parse(time.RFC3339, req.NewestActivityDate); err == nil {
			newestDate = &t
		}
	}

	counts := storage.SyncRunCounts{
		ActivitiesTotal:    req.ActivitiesTotal,
		ActivitiesImported: req.ActivitiesImported,
		ActivitiesSkipped:  req.ActivitiesSkipped,
		GearImported:       req.GearImported,
		StreamsImported:    req.StreamsImported,
		SegmentsImported:   req.SegmentsImported,
		PhotosImported:     req.PhotosImported,
		FailedCount:        req.FailedCount,
		NewestActivityDate: newestDate,
	}

	ctx := context.Background()
	var err error
	switch req.Status {
	case "completed":
		err = bridge.syncHistory.CompleteRun(ctx, req.ID, counts)
	case "failed":
		err = bridge.syncHistory.FailRun(ctx, req.ID, req.Error, counts)
	case "canceled":
		err = bridge.syncHistory.CancelRun(ctx, req.ID, counts)
	default:
		return errorJSON(fmt.Errorf("unknown status for completion: %s", req.Status))
	}

	if err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Sync run %d completed with status %s", req.ID, req.Status))
}

// ============================================================================
// Sync History Read
// ============================================================================

//wasm:category Sync History - Read

// getSyncHistory retrieves sync history records
// Called from JS: goStorage.getSyncHistory(limit?)
//wasm:export
func getSyncHistory(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSyncHistory")

	limit := 10
	if len(args) > 0 && args[0].Type() == js.TypeNumber {
		limit = args[0].Int()
	}

	ctx := context.Background()
	runs, err := bridge.syncHistory.GetLatest(ctx, bridge.athleteID, limit)
	if err != nil {
		return errorJSON(err)
	}

	items := make([]map[string]interface{}, len(runs))
	for i, r := range runs {
		item := map[string]interface{}{
			"id":                  r.ID,
			"athlete_id":          r.AthleteID,
			"started_at":          r.StartedAt.Format(time.RFC3339),
			"status":              r.Status,
			"activities_total":    r.ActivitiesTotal,
			"activities_imported": r.ActivitiesImported,
			"activities_skipped":  r.ActivitiesSkipped,
			"streams_imported":    r.StreamsImported,
			"failed_count":        r.FailedCount,
			"full_sync":           r.FullSync,
			"skip_streams":        r.SkipStreams,
		}
		if r.CompletedAt != nil {
			item["completed_at"] = r.CompletedAt.Format(time.RFC3339)
		}
		if r.DurationSeconds != nil {
			item["duration_seconds"] = *r.DurationSeconds
		}
		if r.Error != "" {
			item["error"] = r.Error
		}
		if r.NewestActivityDate != nil {
			item["newest_activity_date"] = r.NewestActivityDate.Format(time.RFC3339)
		}
		items[i] = item
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": items,
	})
}
