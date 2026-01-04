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
	for i := range activities {
		json := ActivityJSONFromStrava(&activities[i])
		result[i] = json.ToImporter()
	}
	return result, nil
}

// GetActivity fetches a single activity with full details.
func (a *ServerStravaAdapter) GetActivity(ctx context.Context, id int64) (*Activity, error) {
	act, err := a.client.GetActivity(ctx, id)
	if err != nil {
		return nil, err
	}
	json := ActivityJSONFromStrava(act)
	result := json.ToImporter()
	return &result, nil
}

// GetActivityStreams fetches stream data for an activity.
func (a *ServerStravaAdapter) GetActivityStreams(ctx context.Context, id int64, types []string) (*StreamSet, error) {
	streams, err := a.client.GetActivityStreams(ctx, id, types)
	if err != nil {
		return nil, err
	}
	json := StreamSetJSONFromStrava(streams)
	return json.ToImporter(), nil
}

// GetGear fetches gear details by ID.
func (a *ServerStravaAdapter) GetGear(ctx context.Context, id string) (*Gear, error) {
	gear, err := a.client.GetGear(ctx, id)
	if err != nil {
		return nil, err
	}
	json := GearJSONFromStrava(gear)
	return json.ToImporter(), nil
}

// GetSegment fetches segment details by ID.
func (a *ServerStravaAdapter) GetSegment(ctx context.Context, id int64) (*Segment, error) {
	seg, err := a.client.GetSegment(ctx, id)
	if err != nil {
		return nil, err
	}
	json := SegmentJSONFromStrava(seg)
	return json.ToImporter(), nil
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
