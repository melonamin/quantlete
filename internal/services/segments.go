package services

import (
	"context"
	"fmt"
	"time"

	"github.com/melonamin/quantlete/internal/geo"
	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/storage"
)

// SegmentsService handles segment business logic.
type SegmentsService struct {
	repo *storage.SegmentRepository
}

// NewSegmentsService creates a new segments service.
func NewSegmentsService(repo *storage.SegmentRepository) *SegmentsService {
	return &SegmentsService{repo: repo}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// ListSegmentsInput contains parameters for listing segments.
type ListSegmentsInput struct {
	AthleteID    int64  `json:"-" adapter:"context"`
	ActivityType string `json:"activity_type" adapter:"query"`
	Country      string `json:"country" adapter:"query"`
	Search       string `json:"search" adapter:"query"`
	KOMOnly      bool   `json:"kom_only" adapter:"query"`
	Starred      *bool  `json:"starred" adapter:"query"`
	Page         int    `json:"page" adapter:"query,default=1"`
	PerPage      int    `json:"per_page" adapter:"query,default=50"`
	OrderBy      string `json:"order_by" adapter:"query,default=name"`
	OrderDir     string `json:"order_dir" adapter:"query,default=asc"`
}

// SegmentItem represents a segment in responses.
type SegmentItem struct {
	ID                   int64    `json:"id"`
	Name                 string   `json:"name"`
	ActivityType         string   `json:"activity_type"`
	Distance             float64  `json:"distance"`
	AverageGrade         float64  `json:"average_grade"`
	MaximumGrade         float64  `json:"maximum_grade"`
	ElevationHigh        float64  `json:"elevation_high"`
	ElevationLow         float64  `json:"elevation_low"`
	ClimbCategory        int      `json:"climb_category"`
	StartLat             *float64 `json:"start_lat,omitempty"`
	StartLng             *float64 `json:"start_lng,omitempty"`
	EndLat               *float64 `json:"end_lat,omitempty"`
	EndLng               *float64 `json:"end_lng,omitempty"`
	Starred              bool     `json:"starred"`
	Polyline             string   `json:"polyline,omitempty"`
	AthleteKOMRank       *int     `json:"athlete_kom_rank,omitempty"`
	AthleteEffortCount   *int     `json:"athlete_effort_count,omitempty"`
	AthletePRElapsedTime *int     `json:"athlete_pr_elapsed_time,omitempty"`
	AthletePRDate        *string  `json:"athlete_pr_date,omitempty"`
}

// SegmentListItem represents a segment in list responses with effort statistics.
type SegmentListItem struct {
	SegmentItem
	TimesCompleted  int     `json:"times_completed"`
	LastEffortDate  *string `json:"last_effort_date,omitempty"`
	BestElapsedTime *int    `json:"best_elapsed_time,omitempty"`
}

// ListSegmentsOutput contains a paginated list of segments.
type ListSegmentsOutput struct {
	Data       []SegmentListItem `json:"data"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PerPage    int               `json:"per_page"`
	TotalPages int               `json:"total_pages"`
}

// GetSegmentInput contains parameters for getting a single segment.
type GetSegmentInput struct {
	AthleteID int64 `json:"-" adapter:"context"`
	SegmentID int64 `json:"segment_id" adapter:"path,param=id"`
}

// SegmentDetailOutput contains a segment with its effort history.
type SegmentDetailOutput struct {
	Segment SegmentItem         `json:"segment"`
	Efforts []SegmentEffortItem `json:"efforts"`
}

// ListEffortsInput contains parameters for listing segment efforts.
type ListEffortsInput struct {
	AthleteID int64  `json:"-" adapter:"context"`
	SegmentID int64  `json:"segment_id" adapter:"path,param=id"`
	Page      int    `json:"page" adapter:"query,default=1"`
	PerPage   int    `json:"per_page" adapter:"query,default=50"`
	OrderBy   string `json:"order_by" adapter:"query,default=start_date"`
	OrderDir  string `json:"order_dir" adapter:"query,default=desc"`
}

// SegmentEffortItem represents a segment effort in responses.
type SegmentEffortItem struct {
	ID               int64    `json:"id"`
	SegmentID        int64    `json:"segment_id"`
	ActivityID       int64    `json:"activity_id"`
	AthleteID        int64    `json:"athlete_id"`
	Name             string   `json:"name,omitempty"`
	ElapsedTime      int      `json:"elapsed_time"`
	MovingTime       int      `json:"moving_time"`
	StartDate        *string  `json:"start_date,omitempty"`
	StartDateLocal   *string  `json:"start_date_local,omitempty"`
	Distance         float64  `json:"distance"`
	AverageWatts     *float64 `json:"average_watts,omitempty"`
	AverageHeartrate *float64 `json:"average_heartrate,omitempty"`
	MaxHeartrate     *int     `json:"max_heartrate,omitempty"`
	PRRank           *int     `json:"pr_rank,omitempty"`
	Country          string   `json:"country,omitempty"`
}

// ListEffortsOutput contains a paginated list of segment efforts.
type ListEffortsOutput struct {
	Data       []SegmentEffortItem `json:"data"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

// CountryStats represents segment statistics for a country.
type CountryStats struct {
	Country string `json:"country"`
	ISO2    string `json:"iso2,omitempty"`
	Count   int    `json:"count"`
}

// SaveSegmentInput contains parameters for saving a segment.
type SaveSegmentInput struct {
	ID                   int64    `json:"id" adapter:"body"`
	Name                 string   `json:"name" adapter:"body"`
	ActivityType         string   `json:"activity_type" adapter:"body"`
	Distance             float64  `json:"distance" adapter:"body"`
	AverageGrade         float64  `json:"average_grade" adapter:"body"`
	MaximumGrade         float64  `json:"maximum_grade" adapter:"body"`
	ElevationHigh        float64  `json:"elevation_high" adapter:"body"`
	ElevationLow         float64  `json:"elevation_low" adapter:"body"`
	ClimbCategory        int      `json:"climb_category" adapter:"body"`
	StartLat             *float64 `json:"start_lat" adapter:"body"`
	StartLng             *float64 `json:"start_lng" adapter:"body"`
	EndLat               *float64 `json:"end_lat" adapter:"body"`
	EndLng               *float64 `json:"end_lng" adapter:"body"`
	Starred              bool     `json:"starred" adapter:"body"`
	Polyline             string   `json:"polyline" adapter:"body"`
	AthleteKOMRank       *int     `json:"athlete_kom_rank" adapter:"body"`
	AthleteEffortCount   *int     `json:"athlete_effort_count" adapter:"body"`
	AthletePRElapsedTime *int     `json:"athlete_pr_elapsed_time" adapter:"body"`
	AthletePRDate        string   `json:"athlete_pr_date" adapter:"body"`
}

// SaveSegmentOutput contains the result of saving a segment.
type SaveSegmentOutput struct {
	Message string `json:"message"`
}

// SaveSegmentEffortInput contains parameters for saving a segment effort.
type SaveSegmentEffortInput struct {
	ID               int64    `json:"id" adapter:"body"`
	SegmentID        int64    `json:"segment_id" adapter:"body"`
	ActivityID       int64    `json:"activity_id" adapter:"body"`
	AthleteID        int64    `json:"athlete_id" adapter:"body"`
	Name             string   `json:"name" adapter:"body"`
	ElapsedTime      int      `json:"elapsed_time" adapter:"body"`
	MovingTime       int      `json:"moving_time" adapter:"body"`
	StartDate        string   `json:"start_date" adapter:"body"`
	StartDateLocal   string   `json:"start_date_local" adapter:"body"`
	Distance         float64  `json:"distance" adapter:"body"`
	AverageWatts     *float64 `json:"average_watts" adapter:"body"`
	AverageHeartrate *float64 `json:"average_heartrate" adapter:"body"`
	MaxHeartrate     *int     `json:"max_heartrate" adapter:"body"`
	PRRank           *int     `json:"pr_rank" adapter:"body"`
}

// SaveSegmentEffortOutput contains the result of saving a segment effort.
type SaveSegmentEffortOutput struct {
	Message string `json:"message"`
}

// ============================================================================
// Service Methods
// ============================================================================

// List returns a paginated list of segments for an athlete.
//
//adapter:wasm getSegments category=Segments
//adapter:http GET /api/v1/segments
//nolint:dupl // Similar pagination pattern to ListEfforts but different filter/result types
func (s *SegmentsService) List(ctx context.Context, in ListSegmentsInput) (*ListSegmentsOutput, error) {
	f := storage.SegmentFilters{
		ActivityType: in.ActivityType,
		Country:      in.Country,
		Search:       in.Search,
		KOMOnly:      in.KOMOnly,
		Starred:      in.Starred,
		QueryParams: pagination.QueryParams{
			Page:     in.Page,
			PerPage:  in.PerPage,
			OrderBy:  in.OrderBy,
			OrderDir: in.OrderDir,
		},
	}

	result, err := s.repo.List(ctx, in.AthleteID, f)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to list segments: %v", err)
	}

	items := make([]SegmentListItem, 0, len(result.Items))
	for _, it := range result.Items {
		items = append(items, segmentListItemToResponse(it))
	}

	return &ListSegmentsOutput{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}, nil
}

// GetByID retrieves a single segment by ID with its effort history.
//
//adapter:wasm getSegmentDetail category=Segments
//adapter:http GET /api/v1/segments/{id}
func (s *SegmentsService) GetByID(ctx context.Context, in GetSegmentInput) (*SegmentDetailOutput, error) {
	if in.SegmentID == 0 {
		return nil, BadRequest("segment ID required")
	}

	seg, err := s.repo.GetByID(ctx, in.SegmentID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch segment: %v", err)
	}
	if seg == nil {
		return nil, NotFound("segment")
	}

	// Fetch effort history (default limit of 200)
	efforts, err := s.repo.ListEfforts(ctx, in.AthleteID, in.SegmentID, 200)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch efforts: %v", err)
	}

	effortItems := make([]SegmentEffortItem, 0, len(efforts))
	for _, e := range efforts {
		effortItems = append(effortItems, segmentEffortToResponse(e))
	}

	return &SegmentDetailOutput{
		Segment: segmentToResponse(seg),
		Efforts: effortItems,
	}, nil
}

// ListEfforts returns a paginated list of efforts for a segment.
//
//adapter:wasm getSegmentEfforts category=Segments
//adapter:http GET /api/v1/segments/{id}/efforts
//nolint:dupl // Similar pagination pattern to List but different filter/result types
func (s *SegmentsService) ListEfforts(ctx context.Context, in ListEffortsInput) (*ListEffortsOutput, error) {
	if in.SegmentID == 0 {
		return nil, BadRequest("segment ID required")
	}

	f := storage.SegmentEffortFilters{
		QueryParams: pagination.QueryParams{
			Page:     in.Page,
			PerPage:  in.PerPage,
			OrderBy:  in.OrderBy,
			OrderDir: in.OrderDir,
		},
	}

	result, err := s.repo.ListEffortsPaginated(ctx, in.AthleteID, in.SegmentID, f)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to list segment efforts: %v", err)
	}

	items := make([]SegmentEffortItem, 0, len(result.Items))
	for _, e := range result.Items {
		items = append(items, segmentEffortToResponse(e))
	}

	return &ListEffortsOutput{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}, nil
}

// ListCountries returns country statistics for segments with ISO2 codes.
func (s *SegmentsService) ListCountries(ctx context.Context, athleteID int64) ([]CountryStats, error) {
	stats, err := s.repo.ListCountryStats(ctx, athleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to list countries: %v", err)
	}

	out := make([]CountryStats, 0, len(stats))
	for _, cs := range stats {
		item := CountryStats{
			Country: cs.Country,
			Count:   cs.Count,
		}
		if iso2, ok := geo.ISO2FromCountry(cs.Country); ok {
			item.ISO2 = iso2
		}
		out = append(out, item)
	}

	return out, nil
}

// SaveSegment stores a segment in the database.
//
//adapter:wasm saveSegment category=Segments-Write
func (s *SegmentsService) SaveSegment(ctx context.Context, in SaveSegmentInput) (*SaveSegmentOutput, error) {
	seg := &storage.Segment{
		ID:                   in.ID,
		Name:                 in.Name,
		ActivityType:         in.ActivityType,
		Distance:             in.Distance,
		AverageGrade:         in.AverageGrade,
		MaximumGrade:         in.MaximumGrade,
		ElevationHigh:        in.ElevationHigh,
		ElevationLow:         in.ElevationLow,
		ClimbCategory:        in.ClimbCategory,
		StartLat:             in.StartLat,
		StartLng:             in.StartLng,
		EndLat:               in.EndLat,
		EndLng:               in.EndLng,
		Starred:              in.Starred,
		Polyline:             in.Polyline,
		AthleteKOMRank:       in.AthleteKOMRank,
		AthleteEffortCount:   in.AthleteEffortCount,
		AthletePRElapsedTime: in.AthletePRElapsedTime,
	}

	// Parse PR date if provided
	if in.AthletePRDate != "" {
		if t, err := time.Parse(time.RFC3339, in.AthletePRDate); err == nil {
			seg.AthletePRDate = &storage.SQLiteTime{Time: t}
		}
	}

	if err := s.repo.UpsertSegment(ctx, seg); err != nil {
		return nil, Wrapf(ErrInternal, "saving segment: %v", err)
	}

	return &SaveSegmentOutput{
		Message: fmt.Sprintf("Segment %d saved", in.ID),
	}, nil
}

// SaveSegmentEffort stores a segment effort in the database.
//
//adapter:wasm saveSegmentEffort category=Segments-Write
func (s *SegmentsService) SaveSegmentEffort(ctx context.Context, in SaveSegmentEffortInput) (*SaveSegmentEffortOutput, error) {
	effort := &storage.SegmentEffort{
		ID:               in.ID,
		SegmentID:        in.SegmentID,
		ActivityID:       in.ActivityID,
		AthleteID:        in.AthleteID,
		Name:             in.Name,
		ElapsedTime:      in.ElapsedTime,
		MovingTime:       in.MovingTime,
		Distance:         in.Distance,
		AverageWatts:     in.AverageWatts,
		AverageHeartrate: in.AverageHeartrate,
		MaxHeartrate:     in.MaxHeartrate,
		PRRank:           in.PRRank,
	}

	// Parse dates
	if in.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, in.StartDate); err == nil {
			effort.StartDate = &storage.SQLiteTime{Time: t}
		}
	}
	if in.StartDateLocal != "" {
		if t, err := time.Parse(time.RFC3339, in.StartDateLocal); err == nil {
			effort.StartDateLocal = &storage.SQLiteTime{Time: t}
		}
	}

	if err := s.repo.UpsertEffort(ctx, effort); err != nil {
		return nil, Wrapf(ErrInternal, "saving segment effort: %v", err)
	}

	return &SaveSegmentEffortOutput{
		Message: fmt.Sprintf("Segment effort %d saved", in.ID),
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

func segmentToResponse(seg *storage.Segment) SegmentItem {
	var prDate *string
	if seg.AthletePRDate != nil && !seg.AthletePRDate.IsZero() {
		v := seg.AthletePRDate.Format(time.RFC3339)
		prDate = &v
	}

	return SegmentItem{
		ID:                   seg.ID,
		Name:                 seg.Name,
		ActivityType:         seg.ActivityType,
		Distance:             seg.Distance,
		AverageGrade:         seg.AverageGrade,
		MaximumGrade:         seg.MaximumGrade,
		ElevationHigh:        seg.ElevationHigh,
		ElevationLow:         seg.ElevationLow,
		ClimbCategory:        seg.ClimbCategory,
		StartLat:             seg.StartLat,
		StartLng:             seg.StartLng,
		EndLat:               seg.EndLat,
		EndLng:               seg.EndLng,
		Starred:              seg.Starred,
		Polyline:             seg.Polyline,
		AthleteKOMRank:       seg.AthleteKOMRank,
		AthleteEffortCount:   seg.AthleteEffortCount,
		AthletePRElapsedTime: seg.AthletePRElapsedTime,
		AthletePRDate:        prDate,
	}
}

func segmentListItemToResponse(it storage.SegmentListItem) SegmentListItem {
	seg := segmentToResponse(&it.Segment)
	// Clear polyline for list responses to reduce payload size
	seg.Polyline = ""

	var lastEffortDate *string
	if it.LastEffortDate != nil && !it.LastEffortDate.IsZero() {
		v := it.LastEffortDate.Format(time.RFC3339)
		lastEffortDate = &v
	}

	return SegmentListItem{
		SegmentItem:     seg,
		TimesCompleted:  it.TimesCompleted,
		LastEffortDate:  lastEffortDate,
		BestElapsedTime: it.BestElapsedTime,
	}
}

func segmentEffortToResponse(e storage.SegmentEffort) SegmentEffortItem {
	var startDate *string
	if e.StartDate != nil && !e.StartDate.IsZero() {
		v := e.StartDate.Format(time.RFC3339)
		startDate = &v
	}

	var startLocal *string
	if e.StartDateLocal != nil && !e.StartDateLocal.IsZero() {
		v := e.StartDateLocal.Format(time.RFC3339)
		startLocal = &v
	}

	return SegmentEffortItem{
		ID:               e.ID,
		SegmentID:        e.SegmentID,
		ActivityID:       e.ActivityID,
		AthleteID:        e.AthleteID,
		Name:             e.Name,
		ElapsedTime:      e.ElapsedTime,
		MovingTime:       e.MovingTime,
		StartDate:        startDate,
		StartDateLocal:   startLocal,
		Distance:         e.Distance,
		AverageWatts:     e.AverageWatts,
		AverageHeartrate: e.AverageHeartrate,
		MaxHeartrate:     e.MaxHeartrate,
		PRRank:           e.PRRank,
		Country:          e.Country,
	}
}
