package strava

import (
	"context"
	"fmt"
	"net/http"
)

// GetActivities fetches a page of activities for the authenticated athlete.
func (c *Client) GetActivities(ctx context.Context, page, perPage int) ([]Activity, error) {
	path := fmt.Sprintf("/athlete/activities?page=%d&per_page=%d", page, perPage)

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
