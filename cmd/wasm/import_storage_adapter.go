//go:build js && wasm

// Package main provides WASM bridge functionality.
//
// This file provides a WASM-mode storage adapter that wraps the existing
// WASM repositories to implement the ImportStorage interface.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/storage"
)

// ErrStorageUnavailable is returned when the WASM storage bridge is not initialized.
// This error indicates a configuration/initialization problem that should be addressed
// before attempting import operations.
var ErrStorageUnavailable = errors.New("WASM storage bridge not initialized")

// bridgeNilWarningOnce ensures we only log the nil bridge warning once per session.
var bridgeNilWarningOnce sync.Once

// logBridgeNilWarning logs a warning that the bridge is nil (only once per session).
func logBridgeNilWarning(operation string) {
	bridgeNilWarningOnce.Do(func() {
		slog.Warn("WASM storage bridge is nil - storage operations will be skipped",
			"first_operation", operation)
	})
}

// WasmStorageAdapter implements importer.ImportStorage for WASM mode.
// It wraps the existing repository globals from main.go.
type WasmStorageAdapter struct{}

// NewWasmStorageAdapter creates a new WASM storage adapter.
func NewWasmStorageAdapter() *WasmStorageAdapter {
	return &WasmStorageAdapter{}
}

// ============================================================================
// Athletes
// ============================================================================

// SaveAthlete stores an athlete profile.
// In WASM mode, athletes are typically stored via the Strava adapter during auth,
// so this is a no-op. The athlete data comes from the JS side.
func (a *WasmStorageAdapter) SaveAthlete(_ context.Context, _ *importer.Athlete) error {
	// In WASM mode, athlete storage is handled by the JS layer during OAuth flow.
	// The importer receives athlete info from the Strava adapter.
	return nil
}

// ============================================================================
// Activities
// ============================================================================

// SaveActivity stores an activity.
// REQUIRED method - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveActivity(ctx context.Context, athleteIDParam int64, act *importer.Activity) error {
	if act == nil {
		return fmt.Errorf("activity is nil")
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Activities() == nil {
		logBridgeNilWarning("SaveActivity")
		return fmt.Errorf("%w: activities repository unavailable", ErrStorageUnavailable)
	}

	storageAct := importer.ConvertActivityToStorage(athleteIDParam, act)
	return b.registry.Activities().Upsert(ctx, storageAct)
}

// ============================================================================
// Streams
// ============================================================================

// SaveStream stores a stream.
// REQUIRED method when streams are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveStream(ctx context.Context, activityID int64, streamType string, s *importer.Stream) error {
	stream, err := importer.ConvertStreamToStorage(activityID, streamType, s)
	if err != nil {
		return err
	}
	if stream == nil {
		return nil
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Streams() == nil {
		logBridgeNilWarning("SaveStream")
		return fmt.Errorf("%w: streams repository unavailable", ErrStorageUnavailable)
	}

	return b.registry.Streams().Upsert(ctx, stream)
}

// ============================================================================
// Gear
// ============================================================================

// SaveGear stores gear.
// REQUIRED method when gear is imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveGear(ctx context.Context, athleteIDParam int64, g *importer.Gear) error {
	if g == nil {
		return fmt.Errorf("gear is nil")
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Gear() == nil {
		logBridgeNilWarning("SaveGear")
		return fmt.Errorf("%w: gear repository unavailable", ErrStorageUnavailable)
	}

	storageGear := importer.ConvertGearToStorage(athleteIDParam, g)
	return b.registry.Gear().Upsert(ctx, storageGear)
}

// ============================================================================
// Segments
// ============================================================================

// SaveSegment stores a segment.
// REQUIRED method when segments are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveSegment(ctx context.Context, s *importer.Segment) error {
	if s == nil {
		return fmt.Errorf("segment is nil")
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Segments() == nil {
		logBridgeNilWarning("SaveSegment")
		return fmt.Errorf("%w: segments repository unavailable", ErrStorageUnavailable)
	}

	seg := importer.ConvertSegmentToStorage(s)
	return b.registry.Segments().UpsertSegment(ctx, seg)
}

// SaveSegmentEffort stores a segment effort.
// REQUIRED method when segments are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveSegmentEffort(ctx context.Context, athleteIDParam int64, activityID int64, e *importer.SegmentEffort, country string) error {
	if e == nil {
		return fmt.Errorf("segment effort is nil")
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Segments() == nil {
		logBridgeNilWarning("SaveSegmentEffort")
		return fmt.Errorf("%w: segments repository unavailable", ErrStorageUnavailable)
	}

	effort := importer.ConvertSegmentEffortToStorage(athleteIDParam, activityID, e, country)
	return b.registry.Segments().UpsertEffort(ctx, effort)
}

// SaveSegmentWithEffort atomically saves both segment and effort in a single transaction.
// REQUIRED method when segments are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveSegmentWithEffort(ctx context.Context, athleteIDParam int64, activityID int64, e *importer.SegmentEffort, country string) error {
	if e == nil {
		return fmt.Errorf("segment effort is nil")
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Segments() == nil {
		logBridgeNilWarning("SaveSegmentWithEffort")
		return fmt.Errorf("%w: segments repository unavailable", ErrStorageUnavailable)
	}

	seg := importer.ConvertSegmentToStorage(&e.Segment)
	effort := importer.ConvertSegmentEffortToStorage(athleteIDParam, activityID, e, country)
	return b.registry.Segments().UpsertSegmentWithEffort(ctx, seg, effort)
}

// ============================================================================
// Best Efforts
// ============================================================================

// SaveBestEfforts stores best efforts for an activity.
// REQUIRED method when best efforts are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveBestEfforts(ctx context.Context, athleteIDParam, activityID int64, sportType string, efforts []importer.BestEffort) error {
	// Empty efforts slice is valid (activity may have no best efforts)
	if efforts == nil {
		return nil
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.BestEfforts() == nil {
		logBridgeNilWarning("SaveBestEfforts")
		return fmt.Errorf("%w: best efforts repository unavailable", ErrStorageUnavailable)
	}

	storageEfforts := importer.ConvertBestEffortsToStorage(athleteIDParam, activityID, sportType, efforts)
	return b.registry.BestEfforts().ReplaceForActivity(ctx, athleteIDParam, activityID, sportType, storageEfforts)
}

// ============================================================================
// Photos
// ============================================================================

// SavePhoto stores a photo.
// REQUIRED method when photos are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SavePhoto(ctx context.Context, athleteIDParam, activityID int64, p *importer.Photo) error {
	if p == nil {
		return fmt.Errorf("photo is nil")
	}

	photo := importer.ConvertPhotoToStorage(athleteIDParam, activityID, p)
	if photo == nil {
		// No valid URL found - this is a data issue, not a storage issue
		return nil
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Photos() == nil {
		logBridgeNilWarning("SavePhoto")
		return fmt.Errorf("%w: photos repository unavailable", ErrStorageUnavailable)
	}

	return b.registry.Photos().Upsert(ctx, photo)
}

// ============================================================================
// Gear Linking
// ============================================================================

// ResolveCustomGearID resolves a custom gear ID from hashtags in activity name.
// OPTIONAL method - returns empty string if gear repository is unavailable.
func (a *WasmStorageAdapter) ResolveCustomGearID(ctx context.Context, athleteIDParam int64, activityName string) (string, error) {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Gear() == nil {
		return "", nil
	}
	return b.registry.Gear().ResolveCustomGearIDFromActivityName(ctx, athleteIDParam, activityName)
}

// ============================================================================
// Maintenance
// ============================================================================

// LogMaintenanceFromHashtags logs maintenance events from activity hashtags.
// OPTIONAL method - returns nil if maintenance repository is unavailable.
func (a *WasmStorageAdapter) LogMaintenanceFromHashtags(ctx context.Context, athleteIDParam, activityID int64, startDate time.Time, name string) error {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.Maintenance() == nil {
		return nil
	}
	_, err := b.registry.Maintenance().LogFromActivityHashtags(ctx, athleteIDParam, activityID, startDate, name)
	return err
}

// ============================================================================
// Sync History
// ============================================================================

// StartSyncRun starts a new sync run record.
// OPTIONAL method - returns empty SyncRun if sync history is unavailable.
func (a *WasmStorageAdapter) StartSyncRun(ctx context.Context, athleteIDParam int64, opts storage.SyncRunOptions) (*storage.SyncRun, error) {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.SyncHistory() == nil {
		return &storage.SyncRun{}, nil
	}
	return b.registry.SyncHistory().StartRun(ctx, athleteIDParam, opts)
}

// CompleteSyncRun completes a sync run with success.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) CompleteSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.SyncHistory() == nil {
		return nil
	}
	return b.registry.SyncHistory().CompleteRun(ctx, runID, counts)
}

// FailSyncRun marks a sync run as failed.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) FailSyncRun(ctx context.Context, runID int64, errMsg string, counts storage.SyncRunCounts) error {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.SyncHistory() == nil {
		return nil
	}
	return b.registry.SyncHistory().FailRun(ctx, runID, errMsg, counts)
}

// CancelSyncRun marks a sync run as canceled.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) CancelSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.SyncHistory() == nil {
		return nil
	}
	return b.registry.SyncHistory().CancelRun(ctx, runID, counts)
}

// GetSyncWatermark returns the sync watermark for incremental syncs.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) GetSyncWatermark(ctx context.Context, athleteIDParam int64) (*storage.SyncWatermark, error) {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.SyncHistory() == nil {
		return nil, nil
	}
	return b.registry.SyncHistory().GetWatermark(ctx, athleteIDParam)
}

// SetSyncWatermark sets the sync watermark for incremental syncs.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) SetSyncWatermark(ctx context.Context, athleteIDParam int64, wm *storage.SyncWatermark) error {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.SyncHistory() == nil {
		return nil
	}
	return b.registry.SyncHistory().SetWatermark(ctx, athleteIDParam, wm)
}

// ============================================================================
// State Persistence
// ============================================================================

// SaveImportState saves the current import state for resume capability.
// REQUIRED for resume capability - logs warning if unavailable but doesn't fail.
func (a *WasmStorageAdapter) SaveImportState(ctx context.Context, state *importer.ImportState) error {
	if state == nil {
		return fmt.Errorf("import state is nil")
	}

	b := getBridge()
	if b == nil || b.registry == nil || b.registry.AppState() == nil {
		logBridgeNilWarning("SaveImportState")
		// Return nil to allow import to continue (resume just won't work)
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return b.registry.AppState().Set(ctx, "import_state", string(data))
}

// LoadImportState loads the saved import state.
// REQUIRED for resume capability - returns nil if unavailable.
func (a *WasmStorageAdapter) LoadImportState(ctx context.Context) (*importer.ImportState, error) {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.AppState() == nil {
		return nil, nil
	}
	data, err := b.registry.AppState().Get(ctx, "import_state")
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}
	var state importer.ImportState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// ClearImportState clears the saved import state.
// REQUIRED for resume capability - returns nil if unavailable.
func (a *WasmStorageAdapter) ClearImportState(ctx context.Context) error {
	b := getBridge()
	if b == nil || b.registry == nil || b.registry.AppState() == nil {
		return nil
	}
	return b.registry.AppState().Delete(ctx, "import_state")
}
