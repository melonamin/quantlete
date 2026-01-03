//go:build js && wasm

// Package main provides WASM bridge functionality.
//
// This file provides a WASM-mode storage adapter that wraps the existing
// WASM repositories to implement the ImportStorage interface.
package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
)

// WasmStorageAdapter implements importer.ImportStorage for WASM mode.
// It wraps the existing repository globals from main.go.
type WasmStorageAdapter struct{}

// NewWasmStorageAdapter creates a new WASM storage adapter.
func NewWasmStorageAdapter() *WasmStorageAdapter {
	return &WasmStorageAdapter{}
}

// ============================================================================
// Activities
// ============================================================================

// SaveActivity stores an activity.
func (a *WasmStorageAdapter) SaveActivity(ctx context.Context, athleteIDParam int64, act *importer.Activity) error {
	if bridge == nil || bridge.activities == nil {
		return nil
	}

	storageAct := convertImporterActivityToStorageWasm(athleteIDParam, act)
	return bridge.activities.Upsert(ctx, storageAct)
}

// ============================================================================
// Streams
// ============================================================================

// SaveStream stores a stream.
func (a *WasmStorageAdapter) SaveStream(ctx context.Context, activityID int64, streamType string, s *importer.Stream) error {
	if bridge == nil || bridge.streams == nil || s == nil || s.Data == nil {
		return nil
	}

	data, err := json.Marshal(s.Data)
	if err != nil {
		return err
	}

	stream := &storage.ActivityStream{
		ActivityID:   activityID,
		StreamType:   streamType,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
		Data:         data,
	}

	return bridge.streams.Upsert(ctx, stream)
}

// ============================================================================
// Gear
// ============================================================================

// SaveGear stores gear.
func (a *WasmStorageAdapter) SaveGear(ctx context.Context, athleteIDParam int64, g *importer.Gear) error {
	if bridge == nil || bridge.gear == nil {
		return nil
	}

	storageGear := &storage.Gear{
		ID:          g.ID,
		AthleteID:   athleteIDParam,
		Name:        g.Name,
		Primary:     g.Primary,
		Retired:     g.Retired,
		Distance:    g.Distance,
		BrandName:   g.BrandName,
		ModelName:   g.ModelName,
		Description: g.Description,
	}
	return bridge.gear.Upsert(ctx, storageGear)
}

// ============================================================================
// Segments
// ============================================================================

// SaveSegment stores a segment.
func (a *WasmStorageAdapter) SaveSegment(ctx context.Context, s *importer.Segment) error {
	if bridge == nil || bridge.segments == nil {
		return nil
	}

	seg := convertImporterSegmentToStorageWasm(s)
	return bridge.segments.UpsertSegment(ctx, seg)
}

// SaveSegmentEffort stores a segment effort.
func (a *WasmStorageAdapter) SaveSegmentEffort(ctx context.Context, athleteIDParam int64, activityID int64, e *importer.SegmentEffort, country string) error {
	if bridge == nil || bridge.segments == nil {
		return nil
	}

	effort := convertImporterSegmentEffortToStorageWasm(athleteIDParam, activityID, e, country)
	return bridge.segments.UpsertEffort(ctx, effort)
}

// ============================================================================
// Best Efforts
// ============================================================================

// SaveBestEfforts stores best efforts for an activity.
func (a *WasmStorageAdapter) SaveBestEfforts(ctx context.Context, athleteIDParam, activityID int64, sportType string, efforts []importer.BestEffort) error {
	if bridge == nil || bridge.bestEfforts == nil {
		return nil
	}

	storageEfforts := make([]storage.BestEffort, len(efforts))
	for i, e := range efforts {
		dt, canonM := shared.CanonicalBestEffortDistanceType(e.Distance, e.Name)
		storageEfforts[i] = storage.BestEffort{
			AthleteID:    athleteIDParam,
			ActivityID:   activityID,
			SportType:    sportType,
			DistanceType: dt,
			Name:         e.Name,
			DistanceM:    canonM,
			ElapsedTimeS: e.ElapsedTime,
			MovingTimeS:  &e.MovingTime,
			StartIndex:   e.StartIndex,
			EndIndex:     e.EndIndex,
			PRRank:       e.PRRank,
		}
		if !e.StartDate.IsZero() {
			t := storage.SQLiteTime{Time: e.StartDate}
			storageEfforts[i].StartDate = &t
		}
	}

	return bridge.bestEfforts.ReplaceForActivity(ctx, athleteIDParam, activityID, sportType, storageEfforts)
}

// ============================================================================
// Photos
// ============================================================================

// SavePhoto stores a photo.
func (a *WasmStorageAdapter) SavePhoto(ctx context.Context, athleteIDParam, activityID int64, p *importer.Photo) error {
	if bridge == nil || bridge.photos == nil {
		return nil
	}

	url, thumb := bestPhotoURLsFromMapWasm(p.URLs)
	if url == "" {
		return nil
	}

	var loc json.RawMessage
	if len(p.Location) > 0 {
		if b, err := json.Marshal(p.Location); err == nil {
			loc = b
		}
	}

	photo := &storage.Photo{
		ID:           p.UniqueID,
		AthleteID:    athleteIDParam,
		ActivityID:   activityID,
		URL:          url,
		ThumbnailURL: thumb,
		Caption:      p.Caption,
		Location:     loc,
	}

	return bridge.photos.Upsert(ctx, photo)
}

// ============================================================================
// Gear Linking
// ============================================================================

// ResolveCustomGearID resolves a custom gear ID from hashtags in activity name.
func (a *WasmStorageAdapter) ResolveCustomGearID(ctx context.Context, athleteIDParam int64, activityName string) (string, error) {
	if bridge == nil || bridge.gear == nil {
		return "", nil
	}
	return bridge.gear.ResolveCustomGearIDFromActivityName(ctx, athleteIDParam, activityName)
}

// ============================================================================
// Maintenance
// ============================================================================

// LogMaintenanceFromHashtags logs maintenance events from activity hashtags.
func (a *WasmStorageAdapter) LogMaintenanceFromHashtags(ctx context.Context, athleteIDParam, activityID int64, startDate time.Time, name string) error {
	if bridge == nil || bridge.maintenance == nil {
		return nil
	}
	_, err := bridge.maintenance.LogFromActivityHashtags(ctx, athleteIDParam, activityID, startDate, name)
	return err
}

// ============================================================================
// Sync History
// ============================================================================

// StartSyncRun starts a new sync run record.
func (a *WasmStorageAdapter) StartSyncRun(ctx context.Context, athleteIDParam int64, opts storage.SyncRunOptions) (*storage.SyncRun, error) {
	if bridge == nil || bridge.syncHistory == nil {
		return &storage.SyncRun{}, nil
	}
	return bridge.syncHistory.StartRun(ctx, athleteIDParam, opts)
}

// CompleteSyncRun completes a sync run with success.
func (a *WasmStorageAdapter) CompleteSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	if bridge == nil || bridge.syncHistory == nil {
		return nil
	}
	return bridge.syncHistory.CompleteRun(ctx, runID, counts)
}

// FailSyncRun marks a sync run as failed.
func (a *WasmStorageAdapter) FailSyncRun(ctx context.Context, runID int64, errMsg string, counts storage.SyncRunCounts) error {
	if bridge == nil || bridge.syncHistory == nil {
		return nil
	}
	return bridge.syncHistory.FailRun(ctx, runID, errMsg, counts)
}

// CancelSyncRun marks a sync run as canceled.
func (a *WasmStorageAdapter) CancelSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error {
	if bridge == nil || bridge.syncHistory == nil {
		return nil
	}
	return bridge.syncHistory.CancelRun(ctx, runID, counts)
}

// GetSyncWatermark returns the sync watermark for incremental syncs.
func (a *WasmStorageAdapter) GetSyncWatermark(ctx context.Context, athleteIDParam int64) (*storage.SyncWatermark, error) {
	if bridge == nil || bridge.syncHistory == nil {
		return nil, nil
	}
	return bridge.syncHistory.GetWatermark(ctx, athleteIDParam)
}

// SetSyncWatermark sets the sync watermark for incremental syncs.
func (a *WasmStorageAdapter) SetSyncWatermark(ctx context.Context, athleteIDParam int64, wm *storage.SyncWatermark) error {
	if bridge == nil || bridge.syncHistory == nil {
		return nil
	}
	return bridge.syncHistory.SetWatermark(ctx, athleteIDParam, wm)
}

// ============================================================================
// State Persistence
// ============================================================================

// SaveImportState saves the current import state for resume capability.
func (a *WasmStorageAdapter) SaveImportState(ctx context.Context, state *importer.ImportState) error {
	if bridge == nil || bridge.appState == nil {
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return bridge.appState.Set(ctx, "import_state", string(data))
}

// LoadImportState loads the saved import state.
func (a *WasmStorageAdapter) LoadImportState(ctx context.Context) (*importer.ImportState, error) {
	if bridge == nil || bridge.appState == nil {
		return nil, nil
	}
	data, err := bridge.appState.Get(ctx, "import_state")
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
func (a *WasmStorageAdapter) ClearImportState(ctx context.Context) error {
	if bridge == nil || bridge.appState == nil {
		return nil
	}
	return bridge.appState.Delete(ctx, "import_state")
}

// ============================================================================
// Conversion Helpers
// ============================================================================

// convertImporterActivityToStorageWasm converts importer.Activity to storage.Activity.
func convertImporterActivityToStorageWasm(athleteIDParam int64, a *importer.Activity) *storage.Activity {
	act := &storage.Activity{
		ID:                   a.ID,
		AthleteID:            athleteIDParam,
		Name:                 a.Name,
		Description:          a.Description,
		SportType:            a.SportType,
		StartDate:            storage.SQLiteTime{Time: a.StartDate},
		StartDateLocal:       storage.SQLiteTime{Time: a.StartDateLocal},
		Timezone:             a.Timezone,
		LocationCity:         a.LocationCity,
		LocationState:        a.LocationState,
		LocationCountry:      a.LocationCountry,
		Distance:             a.Distance,
		MovingTime:           a.MovingTime,
		ElapsedTime:          a.ElapsedTime,
		TotalElevationGain:   a.TotalElevationGain,
		AverageSpeed:         a.AverageSpeed,
		MaxSpeed:             a.MaxSpeed,
		AverageHeartrate:     a.AverageHeartrate,
		MaxHeartrate:         a.MaxHeartrate,
		AverageWatts:         a.AverageWatts,
		MaxWatts:             a.MaxWatts,
		WeightedAverageWatts: a.WeightedAverageWatts,
		Kilojoules:           a.Kilojoules,
		AverageCadence:       a.AverageCadence,
		Calories:             a.Calories,
		KudosCount:           a.KudosCount,
		PhotoCount:           a.PhotoCount,
		Commute:              a.Commute,
		Private:              a.Private,
		Trainer:              a.Trainer,
		WorkoutType:          a.WorkoutType,
		DeviceName:           a.DeviceName,
		GearID:               a.GearID,
		StartLat:             a.StartLat,
		StartLng:             a.StartLng,
		SummaryPolyline:      a.SummaryPolyline,
	}

	return act
}

// convertImporterSegmentToStorageWasm converts importer.Segment to storage.Segment.
func convertImporterSegmentToStorageWasm(s *importer.Segment) *storage.Segment {
	seg := &storage.Segment{
		ID:            s.ID,
		Name:          s.Name,
		ActivityType:  s.ActivityType,
		Distance:      s.Distance,
		AverageGrade:  s.AverageGrade,
		MaximumGrade:  s.MaximumGrade,
		ElevationHigh: s.ElevationHigh,
		ElevationLow:  s.ElevationLow,
		ClimbCategory: s.ClimbCategory,
		Starred:       s.Starred,
		Polyline:      s.Polyline,
	}

	// Handle latlng arrays
	if len(s.StartLatlng) >= 2 {
		seg.StartLat = &s.StartLatlng[0]
		seg.StartLng = &s.StartLatlng[1]
	}
	if len(s.EndLatlng) >= 2 {
		seg.EndLat = &s.EndLatlng[0]
		seg.EndLng = &s.EndLatlng[1]
	}

	// Handle athlete segment stats
	if s.AthleteSegmentStats.EffortCount > 0 {
		seg.AthleteEffortCount = &s.AthleteSegmentStats.EffortCount
	}
	if s.AthleteSegmentStats.PRElapsedTime > 0 {
		seg.AthletePRElapsedTime = &s.AthleteSegmentStats.PRElapsedTime
	}
	if s.AthleteSegmentStats.PRDate != nil {
		seg.AthletePRDate = &storage.SQLiteTime{Time: *s.AthleteSegmentStats.PRDate}
	}
	seg.AthleteKOMRank = s.AthleteSegmentStats.KOMRank

	return seg
}

// convertImporterSegmentEffortToStorageWasm converts importer.SegmentEffort to storage.SegmentEffort.
func convertImporterSegmentEffortToStorageWasm(athleteIDParam, activityID int64, e *importer.SegmentEffort, country string) *storage.SegmentEffort {
	effort := &storage.SegmentEffort{
		ID:          e.ID,
		SegmentID:   e.Segment.ID,
		ActivityID:  activityID,
		AthleteID:   athleteIDParam,
		Name:        e.Name,
		ElapsedTime: e.ElapsedTime,
		MovingTime:  e.MovingTime,
		Distance:    e.Distance,
		PRRank:      e.PRRank,
		Country:     country,
	}

	if !e.StartDate.IsZero() {
		effort.StartDate = &storage.SQLiteTime{Time: e.StartDate}
	}
	if !e.StartDateLocal.IsZero() {
		effort.StartDateLocal = &storage.SQLiteTime{Time: e.StartDateLocal}
	}
	if e.AverageWatts > 0 {
		effort.AverageWatts = &e.AverageWatts
	}
	if e.AverageHeartrate > 0 {
		effort.AverageHeartrate = &e.AverageHeartrate
	}
	if e.MaxHeartrate > 0 {
		v := int(e.MaxHeartrate)
		effort.MaxHeartrate = &v
	}

	return effort
}

// bestPhotoURLsFromMapWasm extracts best and thumbnail URLs from a URL map.
func bestPhotoURLsFromMapWasm(urls map[string]string) (best, thumb string) {
	if len(urls) == 0 {
		return "", ""
	}

	// Prefer larger sizes for best
	for _, size := range []string{"1000", "600", "200", "100"} {
		if url, ok := urls[size]; ok && url != "" {
			if best == "" {
				best = url
			}
			thumb = url // Keep updating thumb to get smallest
		}
	}

	// Fallback to any URL
	if best == "" {
		for _, url := range urls {
			if url != "" {
				return url, url
			}
		}
	}

	return best, thumb
}
