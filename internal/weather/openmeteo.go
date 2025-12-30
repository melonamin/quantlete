package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	openMeteoArchiveURL = "https://archive-api.open-meteo.com/v1/archive"
	requestTimeout      = 10 * time.Second
	// Open-Meteo free tier: 10,000 requests/day. We use a conservative rate limit
	// to avoid hitting limits when fetching weather for many activities.
	minRequestInterval = 100 * time.Millisecond
)

// OpenMeteoClient fetches historical weather data from Open-Meteo.
type OpenMeteoClient struct {
	httpClient   *http.Client
	rateLimitMu  sync.Mutex
	lastRequest  time.Time
}

// NewOpenMeteoClient creates a new Open-Meteo API client.
func NewOpenMeteoClient() *OpenMeteoClient {
	return &OpenMeteoClient{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// FetchHistorical retrieves historical weather data for a location and time range.
func (c *OpenMeteoClient) FetchHistorical(ctx context.Context, req OpenMeteoRequest) (*OpenMeteoResponse, error) {
	// Apply rate limiting
	c.rateLimitMu.Lock()
	elapsed := time.Since(c.lastRequest)
	if elapsed < minRequestInterval {
		waitTime := minRequestInterval - elapsed
		c.rateLimitMu.Unlock()
		select {
		case <-time.After(waitTime):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		c.rateLimitMu.Lock()
	}
	c.lastRequest = time.Now()
	c.rateLimitMu.Unlock()

	u, err := url.Parse(openMeteoArchiveURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Open-Meteo URL: %w", err)
	}

	q := u.Query()
	q.Set("latitude", fmt.Sprintf("%.6f", req.Latitude))
	q.Set("longitude", fmt.Sprintf("%.6f", req.Longitude))
	q.Set("start_date", req.StartDate.Format("2006-01-02"))
	q.Set("end_date", req.EndDate.Format("2006-01-02"))
	q.Set("hourly", "temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,wind_direction_10m,precipitation,weather_code")
	q.Set("timezone", "auto")
	u.RawQuery = q.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("Open-Meteo returned status %d: %s", resp.StatusCode, string(body))
	}

	var result OpenMeteoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// FindClosestHour finds the index of the closest hour to the target time.
func FindClosestHour(times []string, target time.Time) int {
	if len(times) == 0 {
		return -1
	}

	targetHour := target.Hour()
	targetDate := target.Format("2006-01-02")

	bestIdx := 0
	bestDiff := 24 // Maximum possible hour difference

	for i, t := range times {
		// Parse "2006-01-02T15:00" format
		parsed, err := time.Parse("2006-01-02T15:04", t)
		if err != nil {
			continue
		}

		// Check if same date
		if parsed.Format("2006-01-02") != targetDate {
			continue
		}

		diff := abs(parsed.Hour() - targetHour)
		if diff < bestDiff {
			bestDiff = diff
			bestIdx = i
		}
	}

	return bestIdx
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
