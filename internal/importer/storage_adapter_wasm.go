//go:build js && wasm

// Package importer handles importing data from Strava.
//
// This file provides a WASM-mode storage adapter that wraps repositories
// behind the shared service registry, keeping the import logic platform-agnostic.
package importer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// ErrStorageUnavailable is returned when the WASM storage adapter is not initialized.
// This error indicates a configuration/initialization problem that should be addressed
// before attempting import operations.
var ErrStorageUnavailable = errors.New("WASM storage adapter not initialized")

// registryNilWarningOnce ensures we only log the nil registry warning once per session.
var registryNilWarningOnce sync.Once

// logRegistryNilWarning logs a warning that the registry is nil (only once per session).
func logRegistryNilWarning(operation string) {
	registryNilWarningOnce.Do(func() {
		slog.Warn("WASM storage adapter registry is nil - storage operations will be skipped",
			"first_operation", operation)
	})
}

// WasmStorageAdapter implements ImportStorage for WASM mode.
// It wraps the service registry to access repositories in a testable way.
type WasmStorageAdapter struct {
	registry *services.ServiceRegistry
}

// NewWasmStorageAdapter creates a new WASM storage adapter.
func NewWasmStorageAdapter(registry *services.ServiceRegistry) *WasmStorageAdapter {
	return &WasmStorageAdapter{registry: registry}
}

func (a *WasmStorageAdapter) registryOrError(operation string) (*services.ServiceRegistry, error) {
	if a == nil || a.registry == nil {
		logRegistryNilWarning(operation)
		return nil, fmt.Errorf("%w: service registry unavailable", ErrStorageUnavailable)
	}
	return a.registry, nil
}

// ============================================================================
// Athletes
// ============================================================================

// SaveAthlete stores an athlete profile.
// In WASM mode, athletes are typically stored via the Strava adapter during auth,
// so this is a no-op. The athlete data comes from the JS side.
func (a *WasmStorageAdapter) SaveAthlete(_ context.Context, _ *Athlete) error {
	// In WASM mode, athlete storage is handled by the JS layer during OAuth flow.
	// The importer receives athlete info from the Strava adapter.
	return nil
}

// ============================================================================
// Activities
// ============================================================================

// SaveActivity stores an activity.
// REQUIRED method - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveActivity(ctx context.Context, athleteIDParam int64, act *Activity) error {
	if act == nil {
		return fmt.Errorf("activity is nil")
	}

	registry, err := a.registryOrError("SaveActivity")
	if err != nil {
		return err
	}
	if registry.Activities() == nil {
		return fmt.Errorf("%w: activities repository unavailable", ErrStorageUnavailable)
	}

	storageAct := ConvertActivityToStorage(athleteIDParam, act)
	return registry.Activities().Upsert(ctx, storageAct)
}

// ============================================================================
// Streams
// ============================================================================

// SaveStream stores a stream.
// REQUIRED method when streams are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveStream(ctx context.Context, activityID int64, streamType string, s *Stream) error {
	stream, err := ConvertStreamToStorage(activityID, streamType, s)
	if err != nil {
		return err
	}
	if stream == nil {
		return nil
	}

	registry, err := a.registryOrError("SaveStream")
	if err != nil {
		return err
	}
	if registry.Streams() == nil {
		return fmt.Errorf("%w: streams repository unavailable", ErrStorageUnavailable)
	}

	return registry.Streams().Upsert(ctx, stream)
}

// ============================================================================
// Gear
// ============================================================================

// SaveGear stores gear.
// REQUIRED method when gear is imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveGear(ctx context.Context, athleteIDParam int64, g *Gear) error {
	if g == nil {
		return fmt.Errorf("gear is nil")
	}

	registry, err := a.registryOrError("SaveGear")
	if err != nil {
		return err
	}
	if registry.Gear() == nil {
		return fmt.Errorf("%w: gear repository unavailable", ErrStorageUnavailable)
	}

	storageGear := ConvertGearToStorage(athleteIDParam, g)
	return registry.Gear().Upsert(ctx, storageGear)
}

// ============================================================================
// Segments
// ============================================================================

// SaveSegment stores a segment.
// REQUIRED method when segments are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveSegment(ctx context.Context, s *Segment) error {
	if s == nil {
		return fmt.Errorf("segment is nil")
	}

	registry, err := a.registryOrError("SaveSegment")
	if err != nil {
		return err
	}
	if registry.Segments() == nil {
		return fmt.Errorf("%w: segments repository unavailable", ErrStorageUnavailable)
	}

	seg := ConvertSegmentToStorage(s)
	return registry.Segments().UpsertSegment(ctx, seg)
}

// SaveSegmentEffort stores a segment effort.
// REQUIRED method when segments are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveSegmentEffort(ctx context.Context, athleteIDParam int64, activityID int64, e *SegmentEffort, country string) error {
	if e == nil {
		return fmt.Errorf("segment effort is nil")
	}

	registry, err := a.registryOrError("SaveSegmentEffort")
	if err != nil {
		return err
	}
	if registry.Segments() == nil {
		return fmt.Errorf("%w: segments repository unavailable", ErrStorageUnavailable)
	}

	effort := ConvertSegmentEffortToStorage(athleteIDParam, activityID, e, country)
	return registry.Segments().UpsertEffort(ctx, effort)
}

// SaveSegmentWithEffort atomically saves both segment and effort in a single transaction.
// REQUIRED method when segments are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveSegmentWithEffort(ctx context.Context, athleteIDParam int64, activityID int64, e *SegmentEffort, country string) error {
	if e == nil {
		return fmt.Errorf("segment effort is nil")
	}

	registry, err := a.registryOrError("SaveSegmentWithEffort")
	if err != nil {
		return err
	}
	if registry.Segments() == nil {
		return fmt.Errorf("%w: segments repository unavailable", ErrStorageUnavailable)
	}

	seg := ConvertSegmentToStorage(&e.Segment)
	effort := ConvertSegmentEffortToStorage(athleteIDParam, activityID, e, country)
	return registry.Segments().UpsertSegmentWithEffort(ctx, seg, effort)
}

// ============================================================================
// Best Efforts
// ============================================================================

// SaveBestEfforts stores best efforts for an activity.
// REQUIRED method when best efforts are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SaveBestEfforts(ctx context.Context, athleteIDParam, activityID int64, sportType string, efforts []BestEffort) error {
	// Empty efforts slice is valid (activity may have no best efforts)
	if efforts == nil {
		return nil
	}

	registry, err := a.registryOrError("SaveBestEfforts")
	if err != nil {
		return err
	}
	if registry.BestEfforts() == nil {
		return fmt.Errorf("%w: best efforts repository unavailable", ErrStorageUnavailable)
	}

	storageEfforts := ConvertBestEffortsToStorage(athleteIDParam, activityID, sportType, efforts)
	return registry.BestEfforts().ReplaceForActivity(ctx, athleteIDParam, activityID, sportType, storageEfforts)
}

// ============================================================================
// Photos
// ============================================================================

// SavePhoto stores a photo.
// REQUIRED method when photos are imported - returns error if storage is unavailable.
func (a *WasmStorageAdapter) SavePhoto(ctx context.Context, athleteIDParam, activityID int64, p *Photo) error {
	if p == nil {
		return fmt.Errorf("photo is nil")
	}

	photo := ConvertPhotoToStorage(athleteIDParam, activityID, p)
	if photo == nil {
		// No valid URL found - this is a data issue, not a storage issue
		return nil
	}

	registry, err := a.registryOrError("SavePhoto")
	if err != nil {
		return err
	}
	if registry.Photos() == nil {
		return fmt.Errorf("%w: photos repository unavailable", ErrStorageUnavailable)
	}

	return registry.Photos().Upsert(ctx, photo)
}

// ============================================================================
// Gear Linking
// ============================================================================

// ResolveCustomGearID resolves a custom gear ID from hashtags in activity name.
// OPTIONAL method - returns empty string if gear repository is unavailable.
func (a *WasmStorageAdapter) ResolveCustomGearID(ctx context.Context, athleteIDParam int64, activityName string) (string, error) {
	registry, err := a.registryOrError("ResolveCustomGearID")
	if err != nil || registry.Gear() == nil {
		return "", nil
	}
	return registry.Gear().ResolveCustomGearIDFromActivityName(ctx, athleteIDParam, activityName)
}

// ============================================================================
// Maintenance
// ============================================================================

// LogMaintenanceFromHashtags logs maintenance events from activity hashtags.
// OPTIONAL method - returns nil if maintenance repository is unavailable.
func (a *WasmStorageAdapter) LogMaintenanceFromHashtags(ctx context.Context, athleteIDParam, activityID int64, startDate time.Time, name string) error {
	registry, err := a.registryOrError("LogMaintenanceFromHashtags")
	if err != nil || registry.Maintenance() == nil {
		return nil
	}
	_, err = registry.Maintenance().LogFromActivityHashtags(ctx, athleteIDParam, activityID, startDate, name)
	return err
}

// ============================================================================
// Sync History
// ============================================================================

// StartSyncRun starts a new sync run record.
// OPTIONAL method - returns empty SyncRun if sync history is unavailable.
func (a *WasmStorageAdapter) StartSyncRun(ctx context.Context, athleteIDParam int64, opts storage.SyncRunOptions) (*storage.SyncRun, error) {
	registry, err := a.registryOrError("StartSyncRun")
	if err != nil || registry.SyncHistory() == nil {
		return &storage.SyncRun{}, nil
	}
	return registry.SyncHistory().StartRun(ctx, athleteIDParam, opts)
}

// CompleteSyncRun completes a sync run with success.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) CompleteSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	registry, err := a.registryOrError("CompleteSyncRun")
	if err != nil || registry.SyncHistory() == nil {
		return nil
	}
	return registry.SyncHistory().CompleteRun(ctx, runID, counts)
}

// FailSyncRun marks a sync run as failed.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) FailSyncRun(ctx context.Context, runID int64, errMsg string, counts storage.SyncRunCounts) error {
	registry, err := a.registryOrError("FailSyncRun")
	if err != nil || registry.SyncHistory() == nil {
		return nil
	}
	return registry.SyncHistory().FailRun(ctx, runID, errMsg, counts)
}

// CancelSyncRun marks a sync run as canceled.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) CancelSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	registry, err := a.registryOrError("CancelSyncRun")
	if err != nil || registry.SyncHistory() == nil {
		return nil
	}
	return registry.SyncHistory().CancelRun(ctx, runID, counts)
}

// GetSyncWatermark returns the sync watermark for incremental syncs.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) GetSyncWatermark(ctx context.Context, athleteIDParam int64) (*storage.SyncWatermark, error) {
	registry, err := a.registryOrError("GetSyncWatermark")
	if err != nil || registry.SyncHistory() == nil {
		return nil, nil
	}
	return registry.SyncHistory().GetWatermark(ctx, athleteIDParam)
}

// SetSyncWatermark sets the sync watermark for incremental syncs.
// OPTIONAL method - returns nil if sync history is unavailable.
func (a *WasmStorageAdapter) SetSyncWatermark(ctx context.Context, athleteIDParam int64, wm *storage.SyncWatermark) error {
	registry, err := a.registryOrError("SetSyncWatermark")
	if err != nil || registry.SyncHistory() == nil {
		return nil
	}
	return registry.SyncHistory().SetWatermark(ctx, athleteIDParam, wm)
}

// ============================================================================
// State Persistence
// ============================================================================

// SaveImportState saves the current import state for resume capability.
// REQUIRED for resume capability - logs warning if unavailable but doesn't fail.
func (a *WasmStorageAdapter) SaveImportState(ctx context.Context, state *ImportState) error {
	if state == nil {
		return fmt.Errorf("import state is nil")
	}

	registry, err := a.registryOrError("SaveImportState")
	if err != nil || registry.AppState() == nil {
		// Return nil to allow import to continue (resume just won't work)
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return registry.AppState().Set(ctx, "import_state", string(data))
}

// LoadImportState loads the saved import state.
// REQUIRED for resume capability - returns nil if unavailable.
func (a *WasmStorageAdapter) LoadImportState(ctx context.Context) (*ImportState, error) {
	registry, err := a.registryOrError("LoadImportState")
	if err != nil || registry.AppState() == nil {
		return nil, nil
	}
	data, err := registry.AppState().Get(ctx, "import_state")
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}
	var state ImportState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// ClearImportState clears the saved import state.
// REQUIRED for resume capability - returns nil if unavailable.
func (a *WasmStorageAdapter) ClearImportState(ctx context.Context) error {
	registry, err := a.registryOrError("ClearImportState")
	if err != nil || registry.AppState() == nil {
		return nil
	}
	return registry.AppState().Delete(ctx, "import_state")
}
