package strava

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter tracks Strava API rate limits.
type RateLimiter struct {
	mu sync.RWMutex

	// 15-minute limits
	limit15Min int
	usage15Min int

	// Daily limits
	limitDaily int
	usageDaily int

	// Last update time
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limit15Min: 100, // Default Strava limits
		limitDaily: 1000,
	}
}

// UpdateFromHeaders updates rate limit info from response headers.
func (r *RateLimiter) UpdateFromHeaders(headers http.Header) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// X-RateLimit-Limit: 100,1000 (15min, daily)
	if limit := headers.Get("X-RateLimit-Limit"); limit != "" {
		parts := strings.Split(limit, ",")
		if len(parts) >= 2 {
			if v, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				r.limit15Min = v
			}
			if v, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				r.limitDaily = v
			}
		}
	}

	// X-RateLimit-Usage: 50,500 (15min, daily)
	if usage := headers.Get("X-RateLimit-Usage"); usage != "" {
		parts := strings.Split(usage, ",")
		if len(parts) >= 2 {
			if v, err := strconv.Atoi(strings.TrimSpace(parts[0])); err == nil {
				r.usage15Min = v
			}
			if v, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
				r.usageDaily = v
			}
		}
	}

	r.lastUpdate = time.Now()
}

// Wait blocks if we're at or near the rate limit.
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.RLock()
	usage15 := r.usage15Min
	limit15 := r.limit15Min
	usageDaily := r.usageDaily
	limitDaily := r.limitDaily
	r.mu.RUnlock()

	// If we're at 90% of either limit, slow down
	threshold15 := int(float64(limit15) * 0.9)
	thresholdDaily := int(float64(limitDaily) * 0.9)

	if usage15 >= threshold15 || usageDaily >= thresholdDaily {
		// Calculate wait time based on how close we are to the limit
		var waitTime time.Duration

		if usage15 >= limit15 {
			// At 15-min limit, wait until the 15-min window resets
			waitTime = 15 * time.Minute
		} else if usageDaily >= limitDaily {
			// At daily limit, we need to wait until tomorrow
			waitTime = 24 * time.Hour
		} else {
			// Close to limit, add a small delay
			waitTime = 1 * time.Second
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			return nil
		}
	}

	return nil
}

// Status returns the current rate limit status.
type RateLimitStatus struct {
	Limit15Min int       `json:"limit_15min"`
	Usage15Min int       `json:"usage_15min"`
	LimitDaily int       `json:"limit_daily"`
	UsageDaily int       `json:"usage_daily"`
	LastUpdate time.Time `json:"last_update"`
}

// Status returns the current rate limit status.
func (r *RateLimiter) Status() RateLimitStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return RateLimitStatus{
		Limit15Min: r.limit15Min,
		Usage15Min: r.usage15Min,
		LimitDaily: r.limitDaily,
		UsageDaily: r.usageDaily,
		LastUpdate: r.lastUpdate,
	}
}

// CanMakeRequest returns true if we're under the rate limits.
func (r *RateLimiter) CanMakeRequest() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.usage15Min < r.limit15Min && r.usageDaily < r.limitDaily
}

// Remaining15Min returns the number of requests remaining in the 15-min window.
func (r *RateLimiter) Remaining15Min() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	remaining := r.limit15Min - r.usage15Min
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RemainingDaily returns the number of requests remaining today.
func (r *RateLimiter) RemainingDaily() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	remaining := r.limitDaily - r.usageDaily
	if remaining < 0 {
		return 0
	}
	return remaining
}

// persistedState is the JSON structure for persisted rate limit state.
type persistedState struct {
	Limit15Min int       `json:"limit_15min"`
	Usage15Min int       `json:"usage_15min"`
	LimitDaily int       `json:"limit_daily"`
	UsageDaily int       `json:"usage_daily"`
	LastUpdate time.Time `json:"last_update"`
}

// ToJSON serializes the rate limit state to JSON for persistence.
func (r *RateLimiter) ToJSON() (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state := persistedState{
		Limit15Min: r.limit15Min,
		Usage15Min: r.usage15Min,
		LimitDaily: r.limitDaily,
		UsageDaily: r.usageDaily,
		LastUpdate: r.lastUpdate,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// LoadFromJSON restores rate limit state from JSON.
// It applies decay based on time elapsed since last update.
func (r *RateLimiter) LoadFromJSON(data string) error {
	if data == "" {
		return nil
	}

	var state persistedState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Apply time-based decay to usage counters
	now := time.Now()
	elapsed := now.Sub(state.LastUpdate)

	// 15-minute limit resets every 15 minutes
	// If more than 15 minutes have passed, reset 15-min usage
	if elapsed >= 15*time.Minute {
		state.Usage15Min = 0
	}

	// Daily limit resets at midnight UTC
	// Check if we've crossed a day boundary
	lastDay := state.LastUpdate.UTC().Truncate(24 * time.Hour)
	today := now.UTC().Truncate(24 * time.Hour)
	if today.After(lastDay) {
		state.UsageDaily = 0
	}

	r.limit15Min = state.Limit15Min
	r.usage15Min = state.Usage15Min
	r.limitDaily = state.LimitDaily
	r.usageDaily = state.UsageDaily
	r.lastUpdate = state.LastUpdate

	// Apply defaults if limits are 0 (corrupted state)
	if r.limit15Min == 0 {
		r.limit15Min = 100
	}
	if r.limitDaily == 0 {
		r.limitDaily = 1000
	}

	return nil
}
