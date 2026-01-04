// Package importer handles importing data from Strava.
//
// This file defines interfaces for platform-agnostic import orchestration,
// enabling the same import logic to run in both server mode (native Go) and
// browser mode (WASM with JS Strava client).
package importer

import (
	"context"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Strava Client Interface
// ============================================================================

// StravaClient abstracts Strava API access for the importer.
// In server mode, this is implemented by *strava.Client.
// In WASM mode, this is implemented by a bridge that calls JS stravaFetch.
type StravaClient interface {
	// GetAthlete returns the currently authenticated athlete.
	GetAthlete() *Athlete

	// GetActivities fetches a page of activities.
	GetActivities(ctx context.Context, opts GetActivitiesOptions) ([]Activity, error)

	// GetActivity fetches a single activity with full details.
	GetActivity(ctx context.Context, id int64) (*Activity, error)

	// GetActivityStreams fetches stream data for an activity.
	GetActivityStreams(ctx context.Context, id int64, types []string) (*StreamSet, error)

	// GetGear fetches gear details by ID.
	GetGear(ctx context.Context, id string) (*Gear, error)

	// GetSegment fetches segment details by ID.
	GetSegment(ctx context.Context, id int64) (*Segment, error)

	// GetActivityPhotos fetches photos for an activity.
	GetActivityPhotos(ctx context.Context, id int64) ([]Photo, error)

	// RateLimitInfo returns current rate limit status.
	RateLimitInfo() RateLimitInfo
}

// GetActivitiesOptions configures activity list fetching.
type GetActivitiesOptions struct {
	Page    int
	PerPage int
	After   *time.Time // Only return activities after this date
	Before  *time.Time // Only return activities before this date
}

// Athlete represents an authenticated Strava athlete.
type Athlete struct {
	ID            int64
	Username      string
	FirstName     string
	LastName      string
	ProfileMedium string
}

// Activity represents a Strava activity (summary or detailed).
type Activity struct {
	ID                   int64
	Name                 string
	SportType            string
	StartDate            time.Time
	StartDateLocal       time.Time
	Timezone             string
	Distance             float64
	MovingTime           int
	ElapsedTime          int
	TotalElevationGain   float64
	AverageSpeed         float64
	MaxSpeed             float64
	AverageHeartrate     *float64
	MaxHeartrate         *float64
	AverageWatts         *float64
	MaxWatts             *float64
	WeightedAverageWatts *float64
	Kilojoules           *float64
	AverageCadence       *float64
	Calories             *float64
	GearID               string
	Commute              bool
	Trainer              bool
	Private              bool
	WorkoutType          *int
	LocationCity         string
	LocationState        string
	LocationCountry      string
	SummaryPolyline      string
	StartLat             *float64
	StartLng             *float64
	Description          string
	DeviceName           string
	KudosCount           int
	PhotoCount           int

	// Detail fields (only present when fetching single activity)
	SegmentEfforts []SegmentEffort
	BestEfforts    []BestEffort
}

// StreamSet contains activity stream data.
type StreamSet struct {
	Time           *Stream
	Distance       *Stream
	Altitude       *Stream
	Heartrate      *Stream
	Watts          *Stream
	Cadence        *Stream
	VelocitySmooth *Stream
	Latlng         *Stream
}

// Stream represents a single stream type.
type Stream struct {
	Type         string
	Data         interface{}
	OriginalSize int
	Resolution   string
	SeriesType   string
}

// Gear represents Strava gear.
type Gear struct {
	ID          string
	Name        string
	Primary     bool
	Retired     bool
	Distance    float64
	BrandName   string
	ModelName   string
	Description string
}

// Segment represents a Strava segment.
type Segment struct {
	ID                  int64
	Name                string
	ActivityType        string
	Distance            float64
	AverageGrade        float64
	MaximumGrade        float64
	ElevationHigh       float64
	ElevationLow        float64
	ClimbCategory       int
	StartLatlng         []float64
	EndLatlng           []float64
	Starred             bool
	Polyline            string
	AthleteSegmentStats SegmentStats
}

// SegmentStats contains athlete-specific segment statistics.
type SegmentStats struct {
	PRElapsedTime int
	PRDate        *time.Time
	EffortCount   int
	KOMRank       *int
}

// SegmentEffort represents an effort on a segment.
type SegmentEffort struct {
	ID               int64
	Segment          Segment
	Name             string
	ActivityID       int64
	AthleteID        int64
	ElapsedTime      int
	MovingTime       int
	StartDate        time.Time
	StartDateLocal   time.Time
	Distance         float64
	AverageWatts     float64
	AverageHeartrate float64
	MaxHeartrate     float64
	PRRank           *int
}

// BestEffort represents a best effort on a standard distance.
type BestEffort struct {
	ID          int64
	Name        string
	ElapsedTime int
	MovingTime  int
	StartDate   time.Time
	Distance    float64
	PRRank      *int
	StartIndex  *int
	EndIndex    *int
}

// Photo represents an activity photo.
type Photo struct {
	UniqueID   string
	ActivityID int64
	URLs       map[string]string
	Caption    string
	Location   []float64
	CreatedAt  time.Time
}

// RateLimitInfo contains Strava rate limit information.
type RateLimitInfo struct {
	Used15Min  int
	Limit15Min int
	UsedDaily  int
	LimitDaily int
	RetryAfter time.Duration
}

// ============================================================================
// Storage Interface
// ============================================================================

// ImportStorage abstracts storage operations for the importer.
// In server mode, this wraps the repository layer.
// In WASM mode, this calls goStorage bridge functions.
//
// # Method Categories
//
// Methods are categorized as REQUIRED or OPTIONAL:
//   - REQUIRED: Must be implemented. Returning nil without storing data will cause data loss.
//   - OPTIONAL: May return nil to no-op. Used for platform-specific features.
//
// # Error Handling
//
// All methods should return errors on failure. The importer will:
//   - Log the error for debugging
//   - Track failures in the progress/error aggregation
//   - Continue importing other items (non-fatal for individual items)
type ImportStorage interface {
	// ─────────────────────────────────────────────────────────────────────────
	// Athletes (OPTIONAL)
	// ─────────────────────────────────────────────────────────────────────────

	// SaveAthlete stores an athlete profile.
	// OPTIONAL: In WASM mode, athlete storage is handled by the JS layer during OAuth.
	// Implementations may return nil to no-op.
	SaveAthlete(ctx context.Context, a *Athlete) error

	// ─────────────────────────────────────────────────────────────────────────
	// Activities (REQUIRED)
	// ─────────────────────────────────────────────────────────────────────────

	// SaveActivity stores an activity.
	// REQUIRED: This is the core data type. Must be implemented.
	SaveActivity(ctx context.Context, athleteID int64, a *Activity) error

	// ─────────────────────────────────────────────────────────────────────────
	// Streams (REQUIRED when streams are imported)
	// ─────────────────────────────────────────────────────────────────────────

	// SaveStream stores a stream.
	// REQUIRED: Must be implemented when streams are not skipped.
	SaveStream(ctx context.Context, activityID int64, streamType string, s *Stream) error

	// ─────────────────────────────────────────────────────────────────────────
	// Gear (REQUIRED when gear is imported)
	// ─────────────────────────────────────────────────────────────────────────

	// SaveGear stores gear.
	// REQUIRED: Must be implemented to persist gear data.
	SaveGear(ctx context.Context, athleteID int64, g *Gear) error

	// ─────────────────────────────────────────────────────────────────────────
	// Segments (REQUIRED when segments are imported)
	// ─────────────────────────────────────────────────────────────────────────

	// SaveSegment stores a segment.
	// REQUIRED: Must be implemented when segments are not skipped.
	SaveSegment(ctx context.Context, s *Segment) error

	// SaveSegmentEffort stores a segment effort.
	// REQUIRED: Must be implemented when segments are not skipped.
	SaveSegmentEffort(ctx context.Context, athleteID int64, activityID int64, e *SegmentEffort, country string) error

	// SaveSegmentWithEffort atomically saves both segment and effort in a single transaction.
	// REQUIRED: Must be implemented when segments are not skipped.
	SaveSegmentWithEffort(ctx context.Context, athleteID int64, activityID int64, e *SegmentEffort, country string) error

	// ─────────────────────────────────────────────────────────────────────────
	// Best Efforts (REQUIRED when best efforts are imported)
	// ─────────────────────────────────────────────────────────────────────────

	// SaveBestEfforts stores best efforts for an activity.
	// REQUIRED: Must be implemented when best efforts are not skipped.
	SaveBestEfforts(ctx context.Context, athleteID, activityID int64, sportType string, efforts []BestEffort) error

	// ─────────────────────────────────────────────────────────────────────────
	// Photos (REQUIRED when photos are imported)
	// ─────────────────────────────────────────────────────────────────────────

	// SavePhoto stores a photo.
	// REQUIRED: Must be implemented when photos are not skipped.
	SavePhoto(ctx context.Context, athleteID, activityID int64, p *Photo) error

	// ─────────────────────────────────────────────────────────────────────────
	// Gear Linking (OPTIONAL)
	// ─────────────────────────────────────────────────────────────────────────

	// ResolveCustomGearID resolves a custom gear ID from hashtags in activity name.
	// OPTIONAL: Returns empty string if custom gear linking is not supported.
	ResolveCustomGearID(ctx context.Context, athleteID int64, activityName string) (string, error)

	// ─────────────────────────────────────────────────────────────────────────
	// Maintenance (OPTIONAL)
	// ─────────────────────────────────────────────────────────────────────────

	// LogMaintenanceFromHashtags logs maintenance events from activity hashtags.
	// OPTIONAL: Returns nil if maintenance tracking is not supported.
	LogMaintenanceFromHashtags(ctx context.Context, athleteID, activityID int64, startDate time.Time, name string) error

	// ─────────────────────────────────────────────────────────────────────────
	// Sync History (OPTIONAL)
	// ─────────────────────────────────────────────────────────────────────────
	// These methods track sync run history and watermarks for incremental sync.
	// Implementations may return empty results to disable sync history.

	// StartSyncRun starts a new sync run record.
	// OPTIONAL: Returns empty SyncRun if history tracking is not supported.
	StartSyncRun(ctx context.Context, athleteID int64, opts storage.SyncRunOptions) (*storage.SyncRun, error)

	// CompleteSyncRun completes a sync run with success.
	// OPTIONAL: Returns nil if history tracking is not supported.
	CompleteSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error

	// FailSyncRun marks a sync run as failed.
	// OPTIONAL: Returns nil if history tracking is not supported.
	FailSyncRun(ctx context.Context, runID int64, errMsg string, counts storage.SyncRunCounts) error

	// CancelSyncRun marks a sync run as canceled.
	// OPTIONAL: Returns nil if history tracking is not supported.
	CancelSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error

	// GetSyncWatermark returns the sync watermark for incremental syncs.
	// OPTIONAL: Returns nil if incremental sync is not supported.
	GetSyncWatermark(ctx context.Context, athleteID int64) (*storage.SyncWatermark, error)

	// SetSyncWatermark sets the sync watermark for incremental syncs.
	// OPTIONAL: Returns nil if incremental sync is not supported.
	SetSyncWatermark(ctx context.Context, athleteID int64, wm *storage.SyncWatermark) error

	// ─────────────────────────────────────────────────────────────────────────
	// State Persistence (REQUIRED for resume capability)
	// ─────────────────────────────────────────────────────────────────────────
	// These methods persist import state for resume capability.
	// If not implemented, imports cannot be resumed after interruption.

	// SaveImportState saves the current import state for resume capability.
	// REQUIRED for resume: Without this, interrupted imports cannot resume.
	SaveImportState(ctx context.Context, state *ImportState) error

	// LoadImportState loads the saved import state.
	// REQUIRED for resume: Returns nil if no saved state exists.
	LoadImportState(ctx context.Context) (*ImportState, error)

	// ClearImportState clears the saved import state.
	// REQUIRED for resume: Called after successful import completion.
	ClearImportState(ctx context.Context) error
}

// ============================================================================
// Progress Callback Interface
// ============================================================================

// ProgressCallback receives import progress updates.
// In server mode, this emits SSE events.
// In WASM mode, this calls JS callbacks for UI updates.
type ProgressCallback interface {
	// OnProgress is called when progress changes.
	OnProgress(progress Progress)

	// OnDataChanged is called when data types have been modified.
	OnDataChanged(changes DataChangeSet)

	// OnComplete is called when import finishes (success, failure, or cancel).
	OnComplete(status string, err error)
}
