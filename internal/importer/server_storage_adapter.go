// Package importer handles importing data from Strava.
//
// This file provides a server-mode adapter that wraps storage repositories
// to implement the ImportStorage interface for native Go execution.
package importer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// ServerStorageAdapter wraps storage repositories to implement ImportStorage.
type ServerStorageAdapter struct {
	athletes    *storage.AthleteRepository
	activities  *storage.ActivityRepository
	streams     *storage.StreamRepository
	gear        *storage.GearRepository
	segments    *storage.SegmentRepository
	bestEfforts *storage.BestEffortsRepository
	photos      *storage.PhotoRepository
	maintenance *storage.MaintenanceRepository
	syncHistory *storage.SyncHistoryRepository
	appState    *storage.AppStateRepository
}

// NewServerStorageAdapter creates a server-mode storage adapter.
func NewServerStorageAdapter(
	athletes *storage.AthleteRepository,
	activities *storage.ActivityRepository,
	streams *storage.StreamRepository,
	gear *storage.GearRepository,
	segments *storage.SegmentRepository,
	bestEfforts *storage.BestEffortsRepository,
	photos *storage.PhotoRepository,
	maintenance *storage.MaintenanceRepository,
	syncHistory *storage.SyncHistoryRepository,
	appState *storage.AppStateRepository,
) *ServerStorageAdapter {
	return &ServerStorageAdapter{
		athletes:    athletes,
		activities:  activities,
		streams:     streams,
		gear:        gear,
		segments:    segments,
		bestEfforts: bestEfforts,
		photos:      photos,
		maintenance: maintenance,
		syncHistory: syncHistory,
		appState:    appState,
	}
}

// ============================================================================
// Athletes
// ============================================================================

// SaveAthlete stores an athlete profile.
func (a *ServerStorageAdapter) SaveAthlete(ctx context.Context, athlete *Athlete) error {
	if a.athletes == nil {
		return nil // Optional in some configurations
	}
	storageAthlete := &storage.Athlete{
		ID:            athlete.ID,
		Username:      athlete.Username,
		FirstName:     athlete.FirstName,
		LastName:      athlete.LastName,
		ProfileMedium: athlete.ProfileMedium,
	}
	return a.athletes.Upsert(ctx, storageAthlete)
}

// ============================================================================
// Activities
// ============================================================================

// SaveActivity stores an activity.
func (a *ServerStorageAdapter) SaveActivity(ctx context.Context, athleteID int64, act *Activity) error {
	storageAct := ConvertActivityToStorage(athleteID, act)
	return a.activities.Upsert(ctx, storageAct)
}

// ============================================================================
// Streams
// ============================================================================

// SaveStream stores a stream.
func (a *ServerStorageAdapter) SaveStream(ctx context.Context, activityID int64, streamType string, s *Stream) error {
	stream, err := ConvertStreamToStorage(activityID, streamType, s)
	if err != nil {
		return err
	}
	if stream == nil {
		return nil
	}
	return a.streams.Upsert(ctx, stream)
}

// ============================================================================
// Gear
// ============================================================================

// SaveGear stores gear.
func (a *ServerStorageAdapter) SaveGear(ctx context.Context, athleteID int64, g *Gear) error {
	gear := ConvertGearToStorage(athleteID, g)
	return a.gear.Upsert(ctx, gear)
}

// ============================================================================
// Segments
// ============================================================================

// SaveSegment stores a segment.
func (a *ServerStorageAdapter) SaveSegment(ctx context.Context, s *Segment) error {
	seg := ConvertSegmentToStorage(s)
	return a.segments.UpsertSegment(ctx, seg)
}

// SaveSegmentEffort stores a segment effort.
func (a *ServerStorageAdapter) SaveSegmentEffort(ctx context.Context, athleteID, activityID int64, e *SegmentEffort, country string) error {
	effort := ConvertSegmentEffortToStorage(athleteID, activityID, e, country)
	return a.segments.UpsertEffort(ctx, effort)
}

// SaveSegmentWithEffort atomically saves both segment and effort in a single transaction.
func (a *ServerStorageAdapter) SaveSegmentWithEffort(ctx context.Context, athleteID, activityID int64, e *SegmentEffort, country string) error {
	seg := ConvertSegmentToStorage(&e.Segment)
	effort := ConvertSegmentEffortToStorage(athleteID, activityID, e, country)
	return a.segments.UpsertSegmentWithEffort(ctx, seg, effort)
}

// ============================================================================
// Best Efforts
// ============================================================================

// SaveBestEfforts stores best efforts for an activity.
func (a *ServerStorageAdapter) SaveBestEfforts(ctx context.Context, athleteID, activityID int64, sportType string, efforts []BestEffort) error {
	if a.bestEfforts == nil {
		return nil
	}
	storageEfforts := ConvertBestEffortsToStorage(athleteID, activityID, sportType, efforts)
	return a.bestEfforts.ReplaceForActivity(ctx, athleteID, activityID, sportType, storageEfforts)
}

// ============================================================================
// Photos
// ============================================================================

// SavePhoto stores a photo.
func (a *ServerStorageAdapter) SavePhoto(ctx context.Context, athleteID, activityID int64, p *Photo) error {
	if a.photos == nil {
		return nil
	}
	photo := ConvertPhotoToStorage(athleteID, activityID, p)
	if photo == nil {
		return nil
	}
	return a.photos.Upsert(ctx, photo)
}

// ============================================================================
// Gear Linking
// ============================================================================

// ResolveCustomGearID resolves a custom gear ID from hashtags in activity name.
func (a *ServerStorageAdapter) ResolveCustomGearID(ctx context.Context, athleteID int64, activityName string) (string, error) {
	if a.gear == nil {
		return "", nil
	}
	return a.gear.ResolveCustomGearIDFromActivityName(ctx, athleteID, activityName)
}

// ============================================================================
// Maintenance
// ============================================================================

// LogMaintenanceFromHashtags logs maintenance events from activity hashtags.
func (a *ServerStorageAdapter) LogMaintenanceFromHashtags(ctx context.Context, athleteID, activityID int64, startDate time.Time, name string) error {
	if a.maintenance == nil {
		return nil
	}
	_, err := a.maintenance.LogFromActivityHashtags(ctx, athleteID, activityID, startDate, name)
	return err
}

// ============================================================================
// Sync History
// ============================================================================

// StartSyncRun starts a new sync run record.
func (a *ServerStorageAdapter) StartSyncRun(ctx context.Context, athleteID int64, opts storage.SyncRunOptions) (*storage.SyncRun, error) {
	if a.syncHistory == nil {
		return &storage.SyncRun{}, nil
	}
	return a.syncHistory.StartRun(ctx, athleteID, opts)
}

// CompleteSyncRun completes a sync run with success.
func (a *ServerStorageAdapter) CompleteSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	if a.syncHistory == nil {
		return nil
	}
	return a.syncHistory.CompleteRun(ctx, runID, counts)
}

// FailSyncRun marks a sync run as failed.
func (a *ServerStorageAdapter) FailSyncRun(ctx context.Context, runID int64, errMsg string, counts storage.SyncRunCounts) error {
	if a.syncHistory == nil {
		return nil
	}
	return a.syncHistory.FailRun(ctx, runID, errMsg, counts)
}

// CancelSyncRun marks a sync run as canceled.
func (a *ServerStorageAdapter) CancelSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	if a.syncHistory == nil {
		return nil
	}
	return a.syncHistory.CancelRun(ctx, runID, counts)
}

// GetSyncWatermark returns the sync watermark for incremental syncs.
func (a *ServerStorageAdapter) GetSyncWatermark(ctx context.Context, athleteID int64) (*storage.SyncWatermark, error) {
	if a.syncHistory == nil {
		return nil, nil
	}
	return a.syncHistory.GetWatermark(ctx, athleteID)
}

// SetSyncWatermark sets the sync watermark for incremental syncs.
func (a *ServerStorageAdapter) SetSyncWatermark(ctx context.Context, athleteID int64, wm *storage.SyncWatermark) error {
	if a.syncHistory == nil {
		return nil
	}
	return a.syncHistory.SetWatermark(ctx, athleteID, wm)
}

// ============================================================================
// State Persistence
// ============================================================================

// SaveImportState saves the current import state for resume capability.
func (a *ServerStorageAdapter) SaveImportState(ctx context.Context, state *ImportState) error {
	if a.appState == nil {
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return a.appState.Set(ctx, "import_state", string(data))
}

// LoadImportState loads the saved import state.
func (a *ServerStorageAdapter) LoadImportState(ctx context.Context) (*ImportState, error) {
	if a.appState == nil {
		return nil, nil
	}
	data, err := a.appState.Get(ctx, "import_state")
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
func (a *ServerStorageAdapter) ClearImportState(ctx context.Context) error {
	if a.appState == nil {
		return nil
	}
	return a.appState.Delete(ctx, "import_state")
}
