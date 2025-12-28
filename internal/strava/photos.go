package strava

import (
	"context"
	"fmt"
	"net/http"
)

type ActivityPhoto struct {
	ID       int64             `json:"id"`
	UniqueID string            `json:"unique_id"`
	Caption  string            `json:"caption"`
	Location []float64         `json:"location"`
	URLs     map[string]string `json:"urls"`
}

// GetActivityPhotos fetches photos for an activity.
// Strava returns a list of photos with multiple URL sizes in the `urls` map.
func (c *Client) GetActivityPhotos(ctx context.Context, id int64) ([]ActivityPhoto, error) {
	path := fmt.Sprintf("/activities/%d/photos?size=1000", id)

	var photos []ActivityPhoto
	if err := c.do(ctx, http.MethodGet, path, &photos); err != nil {
		return nil, fmt.Errorf("fetching photos for activity %d: %w", id, err)
	}
	return photos, nil
}
