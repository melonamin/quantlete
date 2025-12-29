package strava

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// API pagination limits (per Strava documentation)
const (
	// MaxActivitiesPerPage is the maximum number of activities returned per API call.
	// Strava API limit: 200. Using this value minimizes API calls during sync.
	MaxActivitiesPerPage = 200
)

// GetActivitiesOptions configures activity list fetching.
type GetActivitiesOptions struct {
	Page    int
	PerPage int
	After   *time.Time // Only return activities after this time (for incremental sync)
	Before  *time.Time // Only return activities before this time
}

// GetActivities fetches a page of activities for the authenticated athlete.
func (c *Client) GetActivities(ctx context.Context, page, perPage int) ([]Activity, error) {
	return c.GetActivitiesWithOptions(ctx, GetActivitiesOptions{
		Page:    page,
		PerPage: perPage,
	})
}

// GetActivitiesWithOptions fetches activities with additional filter options.
func (c *Client) GetActivitiesWithOptions(ctx context.Context, opts GetActivitiesOptions) ([]Activity, error) {
	if opts.Page <= 0 {
		opts.Page = 1
	}
	if opts.PerPage <= 0 {
		opts.PerPage = MaxActivitiesPerPage
	}

	path := fmt.Sprintf("/athlete/activities?page=%d&per_page=%d", opts.Page, opts.PerPage)

	if opts.After != nil {
		path += fmt.Sprintf("&after=%d", opts.After.Unix())
	}
	if opts.Before != nil {
		path += fmt.Sprintf("&before=%d", opts.Before.Unix())
	}

	var activities []Activity
	if err := c.do(ctx, http.MethodGet, path, &activities); err != nil {
		return nil, fmt.Errorf("fetching activities: %w", err)
	}

	return activities, nil
}

// GetActivity fetches a single activity by ID.
func (c *Client) GetActivity(ctx context.Context, id int64) (*Activity, error) {
	path := fmt.Sprintf("/activities/%d?include_all_efforts=true", id)

	var activity Activity
	if err := c.do(ctx, http.MethodGet, path, &activity); err != nil {
		return nil, fmt.Errorf("fetching activity %d: %w", id, err)
	}

	return &activity, nil
}

// GetActivityStreams fetches stream data for an activity.
func (c *Client) GetActivityStreams(ctx context.Context, id int64, types []string) (*StreamSet, error) {
	// Default stream types if none specified
	if len(types) == 0 {
		types = []string{
			"time", "distance", "latlng", "altitude",
			"velocity_smooth", "heartrate", "cadence",
			"watts", "temp", "moving", "grade_smooth",
		}
	}

	// Build comma-separated list
	typeList := ""
	for i, t := range types {
		if i > 0 {
			typeList += ","
		}
		typeList += t
	}

	path := fmt.Sprintf("/activities/%d/streams?keys=%s&key_by_type=true", id, typeList)

	var streams StreamSet
	if err := c.do(ctx, http.MethodGet, path, &streams); err != nil {
		return nil, fmt.Errorf("fetching streams for activity %d: %w", id, err)
	}

	return &streams, nil
}

// GetGear fetches gear details by ID.
func (c *Client) GetGear(ctx context.Context, id string) (*Gear, error) {
	path := fmt.Sprintf("/gear/%s", id)

	var gear Gear
	if err := c.do(ctx, http.MethodGet, path, &gear); err != nil {
		return nil, fmt.Errorf("fetching gear %s: %w", id, err)
	}

	return &gear, nil
}

// GetSegment fetches segment details by ID.
func (c *Client) GetSegment(ctx context.Context, id int64) (*Segment, error) {
	path := fmt.Sprintf("/segments/%d", id)

	var segment Segment
	if err := c.do(ctx, http.MethodGet, path, &segment); err != nil {
		return nil, fmt.Errorf("fetching segment %d: %w", id, err)
	}

	return &segment, nil
}
