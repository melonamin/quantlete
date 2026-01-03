// Package importer handles importing data from Strava.
//
// This file provides a server-mode adapter that wraps storage repositories
// to implement the ImportStorage interface for native Go execution.
package importer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
)

// ServerStorageAdapter wraps storage repositories to implement ImportStorage.
type ServerStorageAdapter struct {
	activities   *storage.ActivityRepository
	streams      *storage.StreamRepository
	gear         *storage.GearRepository
	segments     *storage.SegmentRepository
	bestEfforts  *storage.BestEffortsRepository
	photos       *storage.PhotoRepository
	maintenance  *storage.MaintenanceRepository
	syncHistory  *storage.SyncHistoryRepository
	appState     *storage.AppStateRepository
}

// NewServerStorageAdapter creates a server-mode storage adapter.
func NewServerStorageAdapter(
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
// Activities
// ============================================================================

// SaveActivity stores an activity.
func (a *ServerStorageAdapter) SaveActivity(ctx context.Context, athleteID int64, act *Activity) error {
	storageAct := convertImporterActivityToStorage(athleteID, act)
	return a.activities.Upsert(ctx, storageAct)
}

// ============================================================================
// Streams
// ============================================================================

// SaveStream stores a stream.
func (a *ServerStorageAdapter) SaveStream(ctx context.Context, activityID int64, streamType string, s *Stream) error {
	if s == nil || s.Data == nil {
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

	return a.streams.Upsert(ctx, stream)
}

// ============================================================================
// Gear
// ============================================================================

// SaveGear stores gear.
func (a *ServerStorageAdapter) SaveGear(ctx context.Context, athleteID int64, g *Gear) error {
	gear := &storage.Gear{
		ID:          g.ID,
		AthleteID:   athleteID,
		Name:        g.Name,
		Primary:     g.Primary,
		Retired:     g.Retired,
		Distance:    g.Distance,
		BrandName:   g.BrandName,
		ModelName:   g.ModelName,
		Description: g.Description,
	}
	return a.gear.Upsert(ctx, gear)
}

// ============================================================================
// Segments
// ============================================================================

// SaveSegment stores a segment.
func (a *ServerStorageAdapter) SaveSegment(ctx context.Context, s *Segment) error {
	seg := convertImporterSegmentToStorage(s)
	return a.segments.UpsertSegment(ctx, seg)
}

// SaveSegmentEffort stores a segment effort.
func (a *ServerStorageAdapter) SaveSegmentEffort(ctx context.Context, athleteID int64, activityID int64, e *SegmentEffort, country string) error {
	effort := convertImporterSegmentEffortToStorage(athleteID, activityID, e, country)
	return a.segments.UpsertEffort(ctx, effort)
}

// ============================================================================
// Best Efforts
// ============================================================================

// SaveBestEfforts stores best efforts for an activity.
func (a *ServerStorageAdapter) SaveBestEfforts(ctx context.Context, athleteID, activityID int64, sportType string, efforts []BestEffort) error {
	if a.bestEfforts == nil {
		return nil
	}

	storageEfforts := make([]storage.BestEffort, len(efforts))
	for i, e := range efforts {
		dt, canonM := shared.CanonicalBestEffortDistanceType(e.Distance, e.Name)
		storageEfforts[i] = storage.BestEffort{
			AthleteID:    athleteID,
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

	url, thumb := bestPhotoURLsFromMap(p.URLs)
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
		AthleteID:    athleteID,
		ActivityID:   activityID,
		URL:          url,
		ThumbnailURL: thumb,
		Caption:      p.Caption,
		Location:     loc,
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

// ============================================================================
// Conversion Helpers
// ============================================================================

// convertImporterActivityToStorage converts importer.Activity to storage.Activity.
func convertImporterActivityToStorage(athleteID int64, a *Activity) *storage.Activity {
	act := &storage.Activity{
		ID:                   a.ID,
		AthleteID:            athleteID,
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

// convertImporterSegmentToStorage converts importer.Segment to storage.Segment.
func convertImporterSegmentToStorage(s *Segment) *storage.Segment {
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

// convertImporterSegmentEffortToStorage converts importer.SegmentEffort to storage.SegmentEffort.
func convertImporterSegmentEffortToStorage(athleteID, activityID int64, e *SegmentEffort, country string) *storage.SegmentEffort {
	effort := &storage.SegmentEffort{
		ID:             e.ID,
		SegmentID:      e.Segment.ID,
		ActivityID:     activityID,
		AthleteID:      athleteID,
		Name:           e.Name,
		ElapsedTime:    e.ElapsedTime,
		MovingTime:     e.MovingTime,
		Distance:       e.Distance,
		PRRank:         e.PRRank,
		Country:        country,
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

// bestPhotoURLsFromMap extracts best and thumbnail URLs from a URL map.
func bestPhotoURLsFromMap(urls map[string]string) (best, thumb string) {
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
