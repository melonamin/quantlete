// Package importer handles importing data from Strava.
//
// This file provides a server-mode adapter that wraps strava.Client
// to implement the StravaClient interface for native Go execution.
package importer

import (
	"context"
	"time"

	"github.com/melonamin/quantlete/internal/strava"
)

// ServerStravaAdapter wraps strava.Client to implement StravaClient interface.
type ServerStravaAdapter struct {
	client *strava.Client
}

// NewServerStravaAdapter creates a server-mode Strava client adapter.
func NewServerStravaAdapter(client *strava.Client) *ServerStravaAdapter {
	return &ServerStravaAdapter{client: client}
}

// GetAthlete returns the currently authenticated athlete.
func (a *ServerStravaAdapter) GetAthlete() *Athlete {
	athlete := a.client.GetAthlete()
	if athlete == nil {
		return nil
	}
	return &Athlete{
		ID:            athlete.ID,
		Username:      athlete.Username,
		FirstName:     athlete.FirstName,
		LastName:      athlete.LastName,
		ProfileMedium: athlete.ProfileMedium,
	}
}

// GetActivities fetches a page of activities.
func (a *ServerStravaAdapter) GetActivities(ctx context.Context, opts GetActivitiesOptions) ([]Activity, error) {
	stravaOpts := strava.GetActivitiesOptions{
		Page:    opts.Page,
		PerPage: opts.PerPage,
		After:   opts.After,
		Before:  opts.Before,
	}

	activities, err := a.client.GetActivitiesWithOptions(ctx, stravaOpts)
	if err != nil {
		return nil, err
	}

	result := make([]Activity, len(activities))
	for i, act := range activities {
		result[i] = convertStravaActivity(&act)
	}
	return result, nil
}

// GetActivity fetches a single activity with full details.
func (a *ServerStravaAdapter) GetActivity(ctx context.Context, id int64) (*Activity, error) {
	act, err := a.client.GetActivity(ctx, id)
	if err != nil {
		return nil, err
	}
	result := convertStravaActivity(act)
	return &result, nil
}

// GetActivityStreams fetches stream data for an activity.
func (a *ServerStravaAdapter) GetActivityStreams(ctx context.Context, id int64, types []string) (*StreamSet, error) {
	streams, err := a.client.GetActivityStreams(ctx, id, types)
	if err != nil {
		return nil, err
	}
	return convertStravaStreamSet(streams), nil
}

// GetGear fetches gear details by ID.
func (a *ServerStravaAdapter) GetGear(ctx context.Context, id string) (*Gear, error) {
	gear, err := a.client.GetGear(ctx, id)
	if err != nil {
		return nil, err
	}
	return &Gear{
		ID:          gear.ID,
		Name:        gear.Name,
		Primary:     gear.Primary,
		Retired:     gear.Retired,
		Distance:    gear.Distance,
		BrandName:   gear.BrandName,
		ModelName:   gear.ModelName,
		Description: gear.Description,
	}, nil
}

// GetSegment fetches segment details by ID.
func (a *ServerStravaAdapter) GetSegment(ctx context.Context, id int64) (*Segment, error) {
	seg, err := a.client.GetSegment(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertStravaSegment(seg), nil
}

// GetActivityPhotos fetches photos for an activity.
func (a *ServerStravaAdapter) GetActivityPhotos(ctx context.Context, id int64) ([]Photo, error) {
	photos, err := a.client.GetActivityPhotos(ctx, id)
	if err != nil {
		return nil, err
	}

	result := make([]Photo, len(photos))
	for i, p := range photos {
		result[i] = Photo{
			UniqueID:   p.UniqueID,
			ActivityID: id,
			URLs:       p.URLs,
			Caption:    p.Caption,
			Location:   p.Location,
			CreatedAt:  time.Time{}, // Not available from strava.ActivityPhoto
		}
	}
	return result, nil
}

// RateLimitInfo returns current rate limit status.
func (a *ServerStravaAdapter) RateLimitInfo() RateLimitInfo {
	limiter := a.client.RateLimiter()
	if limiter == nil {
		return RateLimitInfo{}
	}

	status := limiter.Status()
	return RateLimitInfo{
		Used15Min:  status.Usage15Min,
		Limit15Min: status.Limit15Min,
		UsedDaily:  status.UsageDaily,
		LimitDaily: status.LimitDaily,
	}
}

// ============================================================================
// Conversion Helpers
// ============================================================================

// convertStravaActivity converts strava.Activity to importer.Activity.
func convertStravaActivity(a *strava.Activity) Activity {
	result := Activity{
		ID:                 a.ID,
		Name:               a.Name,
		SportType:          a.SportType,
		StartDate:          a.StartDate,
		StartDateLocal:     a.StartDateLocal,
		Timezone:           a.Timezone,
		Distance:           a.Distance,
		MovingTime:         a.MovingTime,
		ElapsedTime:        a.ElapsedTime,
		TotalElevationGain: a.TotalElevationGain,
		AverageSpeed:       a.AverageSpeed,
		MaxSpeed:           a.MaxSpeed,
		GearID:             a.GearID,
		Commute:            a.Commute,
		Trainer:            a.Trainer,
		Private:            a.Private,
		LocationCity:       a.LocationCity,
		LocationState:      a.LocationState,
		LocationCountry:    a.LocationCountry,
		SummaryPolyline:    a.Map.SummaryPolyline,
		Description:        a.Description,
		DeviceName:         a.DeviceName,
		KudosCount:         a.KudosCount,
		PhotoCount:         a.PhotoCount,
	}

	// Convert optional floats
	if a.AverageHeartrate > 0 {
		v := a.AverageHeartrate
		result.AverageHeartrate = &v
	}
	if a.MaxHeartrate > 0 {
		v := a.MaxHeartrate
		result.MaxHeartrate = &v
	}
	if a.AverageWatts > 0 {
		v := a.AverageWatts
		result.AverageWatts = &v
	}
	if a.MaxWatts > 0 {
		v := a.MaxWatts
		result.MaxWatts = &v
	}
	if a.WeightedAverageWatts > 0 {
		v := a.WeightedAverageWatts
		result.WeightedAverageWatts = &v
	}
	if a.Kilojoules > 0 {
		v := a.Kilojoules
		result.Kilojoules = &v
	}
	if a.AverageCadence > 0 {
		v := a.AverageCadence
		result.AverageCadence = &v
	}
	if a.Calories > 0 {
		v := a.Calories
		result.Calories = &v
	}
	if a.WorkoutType > 0 {
		v := a.WorkoutType
		result.WorkoutType = &v
	}

	// Convert start location
	if len(a.StartLatlng) >= 2 {
		result.StartLat = &a.StartLatlng[0]
		result.StartLng = &a.StartLatlng[1]
	}

	// Convert segment efforts
	result.SegmentEfforts = make([]SegmentEffort, len(a.SegmentEfforts))
	for i, e := range a.SegmentEfforts {
		result.SegmentEfforts[i] = convertStravaSegmentEffort(&e)
	}

	// Convert best efforts
	result.BestEfforts = make([]BestEffort, len(a.BestEfforts))
	for i, e := range a.BestEfforts {
		result.BestEfforts[i] = BestEffort{
			ID:          int64(e.ElapsedTime), // No separate ID in strava.BestEffort, use elapsed time
			Name:        e.Name,
			ElapsedTime: e.ElapsedTime,
			MovingTime:  e.MovingTime,
			StartDate:   e.StartDate,
			Distance:    e.Distance,
			PRRank:      e.PRRank,
			StartIndex:  e.StartIndex,
			EndIndex:    e.EndIndex,
		}
	}

	return result
}

// convertStravaSegment converts strava.Segment to importer.Segment.
func convertStravaSegment(s *strava.Segment) *Segment {
	result := &Segment{
		ID:            s.ID,
		Name:          s.Name,
		ActivityType:  s.ActivityType,
		Distance:      s.Distance,
		AverageGrade:  s.AverageGrade,
		MaximumGrade:  s.MaximumGrade,
		ElevationHigh: s.ElevationHigh,
		ElevationLow:  s.ElevationLow,
		ClimbCategory: s.ClimbCategory,
		StartLatlng:   s.StartLatlng,
		EndLatlng:     s.EndLatlng,
		Starred:       s.Starred,
		Polyline:      s.Map.Polyline,
		AthleteSegmentStats: SegmentStats{
			PRElapsedTime: s.AthleteSegmentStats.PRElapsedTime,
			EffortCount:   s.AthleteSegmentStats.EffortCount,
			KOMRank:       s.AthleteSegmentStats.KOMRank,
		},
	}

	// Convert PR date
	if s.AthleteSegmentStats.PRDate != nil && !s.AthleteSegmentStats.PRDate.IsZero() {
		t := s.AthleteSegmentStats.PRDate.Time
		result.AthleteSegmentStats.PRDate = &t
	}

	return result
}

// convertStravaSegmentEffort converts strava.SegmentEffort to importer.SegmentEffort.
func convertStravaSegmentEffort(e *strava.SegmentEffort) SegmentEffort {
	return SegmentEffort{
		ID:               e.ID,
		Segment:          *convertStravaSegment(&e.Segment),
		Name:             e.Name,
		ActivityID:       0, // Not available in strava.SegmentEffort
		AthleteID:        0, // Not available in strava.SegmentEffort
		ElapsedTime:      e.ElapsedTime,
		MovingTime:       e.MovingTime,
		StartDate:        e.StartDate,
		StartDateLocal:   e.StartDateLocal,
		Distance:         e.Distance,
		AverageWatts:     e.AverageWatts,
		AverageHeartrate: e.AverageHeartrate,
		MaxHeartrate:     e.MaxHeartrate,
		PRRank:           e.PRRank,
	}
}

// convertStravaStreamSet converts strava.StreamSet to importer.StreamSet.
func convertStravaStreamSet(s *strava.StreamSet) *StreamSet {
	result := &StreamSet{}

	if s.Time != nil {
		result.Time = convertStravaStream(s.Time)
	}
	if s.Distance != nil {
		result.Distance = convertStravaStream(s.Distance)
	}
	if s.Altitude != nil {
		result.Altitude = convertStravaStream(s.Altitude)
	}
	if s.Heartrate != nil {
		result.Heartrate = convertStravaStream(s.Heartrate)
	}
	if s.Watts != nil {
		result.Watts = convertStravaStream(s.Watts)
	}
	if s.Cadence != nil {
		result.Cadence = convertStravaStream(s.Cadence)
	}
	if s.VelocitySmooth != nil {
		result.VelocitySmooth = convertStravaStream(s.VelocitySmooth)
	}
	if s.Latlng != nil {
		result.Latlng = convertStravaStream(s.Latlng)
	}

	return result
}

// convertStravaStream converts strava.Stream to importer.Stream.
func convertStravaStream(s *strava.Stream) *Stream {
	return &Stream{
		Type:         s.Type,
		Data:         s.Data,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
	}
}
