package strava

import "time"

// BestEffort represents a Strava "best_efforts" item on an activity detail response.
type BestEffort struct {
	Name           string    `json:"name"`
	ElapsedTime    int       `json:"elapsed_time"` // seconds
	MovingTime     int       `json:"moving_time"`  // seconds
	StartDate      time.Time `json:"start_date"`
	StartDateLocal time.Time `json:"start_date_local"`
	Distance       float64   `json:"distance"` // meters
	StartIndex     *int      `json:"start_index,omitempty"`
	EndIndex       *int      `json:"end_index,omitempty"`
	PRRank         *int      `json:"pr_rank,omitempty"`
}
