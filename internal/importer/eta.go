// Package importer handles importing data from Strava.
package importer

import (
	"fmt"
	"time"
)

// ETAEstimator estimates completion time based on rate limits.
type ETAEstimator struct {
	// Rate limits (from Strava)
	FifteenMinLimit int
	FifteenMinUsed  int
	FifteenMinReset time.Time
	DailyLimit      int
	DailyUsed       int
	DailyReset      time.Time

	// Average API call duration
	AvgCallDuration time.Duration
}

// NewETAEstimator creates a new ETA estimator with default values.
func NewETAEstimator() *ETAEstimator {
	return &ETAEstimator{
		FifteenMinLimit: 100,
		DailyLimit:      1000,
		AvgCallDuration: 500 * time.Millisecond,
	}
}

// UpdateFromRateLimits updates the estimator from rate limit info.
func (e *ETAEstimator) UpdateFromRateLimits(fifteenMinUsed, fifteenMinLimit int, fifteenMinReset time.Time, dailyUsed, dailyLimit int, dailyReset time.Time) {
	e.FifteenMinUsed = fifteenMinUsed
	e.FifteenMinLimit = fifteenMinLimit
	e.FifteenMinReset = fifteenMinReset
	e.DailyUsed = dailyUsed
	e.DailyLimit = dailyLimit
	e.DailyReset = dailyReset
}

// FifteenMinRemaining returns remaining calls in current 15-min window.
func (e *ETAEstimator) FifteenMinRemaining() int {
	return e.FifteenMinLimit - e.FifteenMinUsed
}

// DailyRemaining returns remaining calls in current day.
func (e *ETAEstimator) DailyRemaining() int {
	return e.DailyLimit - e.DailyUsed
}

// EstimateCompletion estimates time to complete given number of API calls.
func (e *ETAEstimator) EstimateCompletion(remainingCalls int) time.Duration {
	if remainingCalls <= 0 {
		return 0
	}

	now := time.Now()
	fifteenMinRemain := e.FifteenMinRemaining()
	dailyRemain := e.DailyRemaining()

	// If we can complete in current 15-min window
	if remainingCalls <= fifteenMinRemain && remainingCalls <= dailyRemain {
		return time.Duration(remainingCalls) * e.AvgCallDuration
	}

	// Calculate time accounting for rate limit windows
	totalDuration := time.Duration(0)
	callsLeft := remainingCalls

	// Use remaining calls in current 15-min window
	if fifteenMinRemain > 0 && dailyRemain > 0 {
		usable := min(fifteenMinRemain, dailyRemain, callsLeft)
		totalDuration += time.Duration(usable) * e.AvgCallDuration
		callsLeft -= usable
		dailyRemain -= usable
	}

	if callsLeft <= 0 {
		return totalDuration
	}

	// Wait for current 15-min window to reset
	if !e.FifteenMinReset.IsZero() && e.FifteenMinReset.After(now) {
		totalDuration += e.FifteenMinReset.Sub(now)
	} else {
		totalDuration += 15 * time.Minute
	}

	// Calculate full 15-min windows needed
	callsPerWindow := min(e.FifteenMinLimit, dailyRemain)
	if callsPerWindow <= 0 {
		// Daily limit exhausted, need to wait for daily reset
		if !e.DailyReset.IsZero() && e.DailyReset.After(now) {
			return totalDuration + e.DailyReset.Sub(now)
		}
		return totalDuration + 24*time.Hour
	}

	fullWindows := callsLeft / callsPerWindow
	totalDuration += time.Duration(fullWindows) * 15 * time.Minute
	callsLeft -= fullWindows * callsPerWindow

	// Partial final window
	if callsLeft > 0 {
		totalDuration += time.Duration(callsLeft) * e.AvgCallDuration
	}

	return totalDuration
}

// FormatETA formats the ETA duration as human-readable string.
func FormatETA(d time.Duration) string {
	if d <= 0 {
		return "< 1 min"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 24 {
		days := hours / 24
		hours %= 24
		if hours > 0 {
			return formatPlural(days, "day") + " " + formatPlural(hours, "hour")
		}
		return formatPlural(days, "day")
	}

	if hours > 0 {
		if minutes > 0 {
			return formatPlural(hours, "hour") + " " + formatPlural(minutes, "min")
		}
		return formatPlural(hours, "hour")
	}

	if minutes > 0 {
		return formatPlural(minutes, "min")
	}

	return "< 1 min"
}

func formatPlural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return fmt.Sprintf("%d %ss", n, unit)
}
