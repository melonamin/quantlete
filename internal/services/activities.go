package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// ActivityService handles activity business logic.
type ActivityService struct {
	repo    *storage.ActivityRepository
	streams *storage.StreamRepository
}

// NewActivityService creates a new activity service.
func NewActivityService(repo *storage.ActivityRepository, streams *storage.StreamRepository) *ActivityService {
	return &ActivityService{
		repo:    repo,
		streams: streams,
	}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// ListActivitiesInput contains parameters for listing activities.
type ListActivitiesInput struct {
	AthleteID   int64      `json:"-" adapter:"context"`
	SportTypes  []string   `json:"sport_types" adapter:"query,name=sport_type,split=,"`
	StartAfter  *time.Time `json:"after" adapter:"query"`
	StartBefore *time.Time `json:"before" adapter:"query"`
	GearID      string     `json:"gear_id" adapter:"query"`
	Commute     *bool      `json:"commute" adapter:"query"`
	Trainer     *bool      `json:"trainer" adapter:"query"`
	Search      string     `json:"search" adapter:"query"`
	Page        int        `json:"page" adapter:"query,default=1"`
	PerPage     int        `json:"per_page" adapter:"query,default=50"`
	OrderBy     string     `json:"order_by" adapter:"query,default=start_date"`
	OrderDir    string     `json:"order_dir" adapter:"query,default=desc"`
}

// ActivityItem represents an activity in responses.
type ActivityItem struct {
	ID                   int64    `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description,omitempty"`
	SportType            string   `json:"sport_type"`
	StartDate            string   `json:"start_date"`
	StartDateLocal       string   `json:"start_date_local"`
	Timezone             string   `json:"timezone,omitempty"`
	Distance             float64  `json:"distance"`
	MovingTime           int      `json:"moving_time"`
	ElapsedTime          int      `json:"elapsed_time"`
	TotalElevationGain   float64  `json:"total_elevation_gain"`
	ElevHigh             *float64 `json:"elev_high,omitempty"`
	ElevLow              *float64 `json:"elev_low,omitempty"`
	AverageSpeed         float64  `json:"average_speed"`
	MaxSpeed             float64  `json:"max_speed"`
	AverageHeartrate     *float64 `json:"average_heartrate,omitempty"`
	MaxHeartrate         *float64 `json:"max_heartrate,omitempty"`
	AverageWatts         *float64 `json:"average_watts,omitempty"`
	MaxWatts             *float64 `json:"max_watts,omitempty"`
	WeightedAverageWatts *float64 `json:"weighted_average_watts,omitempty"`
	Kilojoules           *float64 `json:"kilojoules,omitempty"`
	AverageCadence       *float64 `json:"average_cadence,omitempty"`
	Calories             *float64 `json:"calories,omitempty"`
	KudosCount           int      `json:"kudos_count"`
	CommentCount         int      `json:"comment_count"`
	PhotoCount           int      `json:"photo_count"`
	Commute              bool     `json:"commute"`
	Private              bool     `json:"private"`
	Trainer              bool     `json:"trainer"`
	WorkoutType          *int     `json:"workout_type,omitempty"`
	DeviceName           string   `json:"device_name,omitempty"`
	GearID               string   `json:"gear_id,omitempty"`
	StartLat             *float64 `json:"start_lat,omitempty"`
	StartLng             *float64 `json:"start_lng,omitempty"`
	EndLat               *float64 `json:"end_lat,omitempty"`
	EndLng               *float64 `json:"end_lng,omitempty"`
	SummaryPolyline      string   `json:"summary_polyline,omitempty"`
	LocationCity         string   `json:"location_city,omitempty"`
	LocationCountry      string   `json:"location_country,omitempty"`
}

// ListActivitiesOutput contains a paginated list of activities.
type ListActivitiesOutput struct {
	Data       []ActivityItem `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

// GetActivityInput contains parameters for getting a single activity.
type GetActivityInput struct {
	AthleteID  int64 `json:"-" adapter:"context"`
	ActivityID int64 `json:"activity_id" adapter:"path,param=id"`
}

// GetActivityStreamsInput contains parameters for getting activity streams.
type GetActivityStreamsInput struct {
	AthleteID  int64 `json:"-" adapter:"context"`
	ActivityID int64 `json:"activity_id" adapter:"path,param=id"`
}

// ActivityStreamItem represents a stream in responses.
type ActivityStreamItem struct {
	ActivityID   int64           `json:"activity_id"`
	StreamType   string          `json:"stream_type"`
	OriginalSize int             `json:"original_size"`
	Resolution   string          `json:"resolution"`
	SeriesType   string          `json:"series_type"`
	Data         json.RawMessage `json:"data"`
}

// SaveActivityInput contains parameters for saving an activity.
type SaveActivityInput struct {
	ID                   int64    `json:"id" adapter:"body"`
	AthleteID            int64    `json:"athlete_id" adapter:"body"`
	Name                 string   `json:"name" adapter:"body"`
	SportType            string   `json:"sport_type" adapter:"body"`
	StartDate            string   `json:"start_date" adapter:"body"`
	StartDateLocal       string   `json:"start_date_local" adapter:"body"`
	Timezone             string   `json:"timezone" adapter:"body"`
	Distance             float64  `json:"distance" adapter:"body"`
	MovingTime           int      `json:"moving_time" adapter:"body"`
	ElapsedTime          int      `json:"elapsed_time" adapter:"body"`
	TotalElevationGain   float64  `json:"total_elevation_gain" adapter:"body"`
	AverageSpeed         float64  `json:"average_speed" adapter:"body"`
	MaxSpeed             float64  `json:"max_speed" adapter:"body"`
	AverageHeartrate     *float64 `json:"average_heartrate" adapter:"body"`
	MaxHeartrate         *float64 `json:"max_heartrate" adapter:"body"`
	AverageWatts         *float64 `json:"average_watts" adapter:"body"`
	MaxWatts             *float64 `json:"max_watts" adapter:"body"`
	WeightedAverageWatts *float64 `json:"weighted_average_watts" adapter:"body"`
	Kilojoules           *float64 `json:"kilojoules" adapter:"body"`
	AverageCadence       *float64 `json:"average_cadence" adapter:"body"`
	Calories             *float64 `json:"calories" adapter:"body"`
	GearID               string   `json:"gear_id" adapter:"body"`
	Commute              bool     `json:"commute" adapter:"body"`
	WorkoutType          *int     `json:"workout_type" adapter:"body"`
	LocationCity         string   `json:"location_city" adapter:"body"`
	LocationState        string   `json:"location_state" adapter:"body"`
	LocationCountry      string   `json:"location_country" adapter:"body"`
	SummaryPolyline      string   `json:"summary_polyline" adapter:"body"`
	StartLat             *float64 `json:"start_lat" adapter:"body"`
	StartLng             *float64 `json:"start_lng" adapter:"body"`
	Description          string   `json:"description" adapter:"body"`
	DeviceName           string   `json:"device_name" adapter:"body"`
	Trainer              bool     `json:"trainer" adapter:"body"`
	Private              bool     `json:"private" adapter:"body"`
	KudosCount           int      `json:"kudos_count" adapter:"body"`
	PhotoCount           int      `json:"photo_count" adapter:"body"`
}

// SaveActivityOutput contains the result of saving an activity.
type SaveActivityOutput struct {
	Message string `json:"message"`
}

// SaveStreamInput contains parameters for saving an activity stream.
type SaveStreamInput struct {
	ActivityID   int64       `json:"activity_id" adapter:"body"`
	StreamType   string      `json:"stream_type" adapter:"body"`
	Data         interface{} `json:"data" adapter:"body"`
	SeriesType   string      `json:"series_type" adapter:"body"`
	OriginalSize int         `json:"original_size" adapter:"body"`
	Resolution   string      `json:"resolution" adapter:"body"`
}

// SaveStreamOutput contains the result of saving a stream.
type SaveStreamOutput struct {
	Message string `json:"message"`
}

// Maximum allowed stream data size to prevent memory exhaustion.
const maxStreamDataSize = 100000

// ============================================================================
// Service Methods
// ============================================================================

// List returns a paginated list of activities for an athlete.
//
//adapter:wasm getActivities category=Activities
//adapter:http GET /api/v1/activities
func (s *ActivityService) List(ctx context.Context, in ListActivitiesInput) (*ListActivitiesOutput, error) {
	filters := storage.ActivityFilters{
		AthleteID:   in.AthleteID,
		SportTypes:  in.SportTypes,
		StartAfter:  in.StartAfter,
		StartBefore: in.StartBefore,
		GearID:      in.GearID,
		Commute:     in.Commute,
		Trainer:     in.Trainer,
		Search:      in.Search,
	}

	page := storage.Pagination{
		Page:     in.Page,
		PerPage:  in.PerPage,
		OrderBy:  in.OrderBy,
		OrderDir: in.OrderDir,
	}
	page.Normalize()

	activities, total, err := s.repo.List(ctx, filters, page)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to list activities: %v", err)
	}

	items := make([]ActivityItem, len(activities))
	for i, a := range activities {
		items[i] = activityToItem(&a)
	}

	totalPages := (total + page.PerPage - 1) / page.PerPage

	return &ListActivitiesOutput{
		Data:       items,
		Total:      total,
		Page:       page.Page,
		PerPage:    page.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetByID retrieves a single activity by ID.
//
//adapter:wasm getActivity category=Activities
//adapter:http GET /api/v1/activities/{id}
func (s *ActivityService) GetByID(ctx context.Context, in GetActivityInput) (*ActivityItem, error) {
	if in.ActivityID == 0 {
		return nil, BadRequest("activity ID required")
	}

	activity, err := s.repo.GetByID(ctx, in.ActivityID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch activity: %v", err)
	}
	if activity == nil {
		return nil, NotFound("activity")
	}

	// Authorization: verify activity belongs to the athlete
	if activity.AthleteID != in.AthleteID {
		return nil, Wrap(ErrForbidden, "access denied")
	}

	item := activityToItem(activity)
	return &item, nil
}

// GetStreams retrieves streams for an activity.
//
//adapter:wasm getActivityStreams category=Activities
//adapter:http GET /api/v1/activities/{id}/streams
func (s *ActivityService) GetStreams(ctx context.Context, in GetActivityStreamsInput) ([]ActivityStreamItem, error) {
	if in.ActivityID == 0 {
		return nil, BadRequest("activity ID required")
	}

	// First verify the activity exists and belongs to the athlete
	activity, err := s.repo.GetByID(ctx, in.ActivityID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch activity: %v", err)
	}
	if activity == nil {
		return nil, NotFound("activity")
	}

	// Authorization: verify activity belongs to the athlete
	if activity.AthleteID != in.AthleteID {
		return nil, Wrap(ErrForbidden, "access denied")
	}

	streams, err := s.streams.GetByActivityID(ctx, in.ActivityID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch streams: %v", err)
	}

	items := make([]ActivityStreamItem, 0, len(streams))
	for _, stream := range streams {
		items = append(items, ActivityStreamItem{
			ActivityID:   stream.ActivityID,
			StreamType:   stream.StreamType,
			OriginalSize: stream.OriginalSize,
			Resolution:   stream.Resolution,
			SeriesType:   stream.SeriesType,
			Data:         stream.Data,
		})
	}

	return items, nil
}

// SaveActivity stores an activity in the database.
//
//adapter:wasm saveActivity category=Activities-Write
func (s *ActivityService) SaveActivity(ctx context.Context, in SaveActivityInput) (*SaveActivityOutput, error) {
	// Parse dates - these are required fields
	startDate, err := time.Parse(time.RFC3339, in.StartDate)
	if err != nil {
		return nil, BadRequestf("parsing start_date %q: %v", in.StartDate, err)
	}
	startDateLocal, err := time.Parse(time.RFC3339, in.StartDateLocal)
	if err != nil {
		return nil, BadRequestf("parsing start_date_local %q: %v", in.StartDateLocal, err)
	}

	activity := &storage.Activity{
		ID:                   in.ID,
		AthleteID:            in.AthleteID,
		Name:                 in.Name,
		SportType:            in.SportType,
		StartDate:            storage.SQLiteTime{Time: startDate},
		StartDateLocal:       storage.SQLiteTime{Time: startDateLocal},
		Timezone:             in.Timezone,
		Distance:             in.Distance,
		MovingTime:           in.MovingTime,
		ElapsedTime:          in.ElapsedTime,
		TotalElevationGain:   in.TotalElevationGain,
		AverageSpeed:         in.AverageSpeed,
		MaxSpeed:             in.MaxSpeed,
		AverageHeartrate:     in.AverageHeartrate,
		MaxHeartrate:         in.MaxHeartrate,
		AverageWatts:         in.AverageWatts,
		MaxWatts:             in.MaxWatts,
		WeightedAverageWatts: in.WeightedAverageWatts,
		Kilojoules:           in.Kilojoules,
		AverageCadence:       in.AverageCadence,
		Calories:             in.Calories,
		GearID:               in.GearID,
		Commute:              in.Commute,
		WorkoutType:          in.WorkoutType,
		LocationCity:         in.LocationCity,
		LocationState:        in.LocationState,
		LocationCountry:      in.LocationCountry,
		SummaryPolyline:      in.SummaryPolyline,
		StartLat:             in.StartLat,
		StartLng:             in.StartLng,
		Description:          in.Description,
		DeviceName:           in.DeviceName,
		Trainer:              in.Trainer,
		Private:              in.Private,
		KudosCount:           in.KudosCount,
		PhotoCount:           in.PhotoCount,
	}

	if err := s.repo.Upsert(ctx, activity); err != nil {
		return nil, Wrapf(ErrInternal, "saving activity: %v", err)
	}

	return &SaveActivityOutput{
		Message: fmt.Sprintf("Activity %d saved", in.ID),
	}, nil
}

// SaveStream stores an activity stream in the database.
//
//adapter:wasm saveStream category=Activities-Write
func (s *ActivityService) SaveStream(ctx context.Context, in SaveStreamInput) (*SaveStreamOutput, error) {
	// Validate original size to prevent memory exhaustion
	if in.OriginalSize > maxStreamDataSize {
		return nil, BadRequestf("stream original_size exceeds maximum: %d > %d", in.OriginalSize, maxStreamDataSize)
	}

	// Convert data to JSON for storage
	dataJSON, err := json.Marshal(in.Data)
	if err != nil {
		return nil, BadRequestf("marshaling data: %v", err)
	}

	stream := &storage.ActivityStream{
		ActivityID:   in.ActivityID,
		StreamType:   in.StreamType,
		Data:         json.RawMessage(dataJSON),
		SeriesType:   in.SeriesType,
		OriginalSize: in.OriginalSize,
		Resolution:   in.Resolution,
	}

	if err := s.streams.Upsert(ctx, stream); err != nil {
		return nil, Wrapf(ErrInternal, "saving stream: %v", err)
	}

	return &SaveStreamOutput{
		Message: fmt.Sprintf("Stream %s for activity %d saved", in.StreamType, in.ActivityID),
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

func activityToItem(a *storage.Activity) ActivityItem {
	return ActivityItem{
		ID:                   a.ID,
		Name:                 a.Name,
		Description:          a.Description,
		SportType:            a.SportType,
		StartDate:            a.StartDate.Format(time.RFC3339),
		StartDateLocal:       a.StartDateLocal.Format(time.RFC3339),
		Timezone:             a.Timezone,
		Distance:             a.Distance,
		MovingTime:           a.MovingTime,
		ElapsedTime:          a.ElapsedTime,
		TotalElevationGain:   a.TotalElevationGain,
		ElevHigh:             a.ElevHigh,
		ElevLow:              a.ElevLow,
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
		CommentCount:         a.CommentCount,
		PhotoCount:           a.PhotoCount,
		Commute:              a.Commute,
		Private:              a.Private,
		Trainer:              a.Trainer,
		WorkoutType:          a.WorkoutType,
		DeviceName:           a.DeviceName,
		GearID:               a.GearID,
		StartLat:             a.StartLat,
		StartLng:             a.StartLng,
		EndLat:               a.EndLat,
		EndLng:               a.EndLng,
		SummaryPolyline:      a.SummaryPolyline,
		LocationCity:         a.LocationCity,
		LocationCountry:      a.LocationCountry,
	}
}
