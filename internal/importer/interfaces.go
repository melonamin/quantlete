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
	ID           int64
	Name         string
	ElapsedTime  int
	MovingTime   int
	StartDate    time.Time
	Distance     float64
	PRRank       *int
	StartIndex   *int
	EndIndex     *int
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
type ImportStorage interface {
	// Activities
	SaveActivity(ctx context.Context, athleteID int64, a *Activity) error

	// Streams
	SaveStream(ctx context.Context, activityID int64, streamType string, s *Stream) error

	// Gear
	SaveGear(ctx context.Context, athleteID int64, g *Gear) error

	// Segments
	SaveSegment(ctx context.Context, s *Segment) error
	SaveSegmentEffort(ctx context.Context, athleteID int64, activityID int64, e *SegmentEffort, country string) error

	// Best Efforts
	SaveBestEfforts(ctx context.Context, athleteID, activityID int64, sportType string, efforts []BestEffort) error

	// Photos
	SavePhoto(ctx context.Context, athleteID, activityID int64, p *Photo) error

	// Gear linking
	ResolveCustomGearID(ctx context.Context, athleteID int64, activityName string) (string, error)

	// Maintenance
	LogMaintenanceFromHashtags(ctx context.Context, athleteID, activityID int64, startDate time.Time, name string) error

	// Sync History
	StartSyncRun(ctx context.Context, athleteID int64, opts storage.SyncRunOptions) (*storage.SyncRun, error)
	CompleteSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error
	FailSyncRun(ctx context.Context, runID int64, errMsg string, counts storage.SyncRunCounts) error
	CancelSyncRun(ctx context.Context, runID int64, counts storage.SyncRunCounts) error
	GetSyncWatermark(ctx context.Context, athleteID int64) (*storage.SyncWatermark, error)
	SetSyncWatermark(ctx context.Context, athleteID int64, wm *storage.SyncWatermark) error

	// State Persistence
	SaveImportState(ctx context.Context, state *ImportState) error
	LoadImportState(ctx context.Context) (*ImportState, error)
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
