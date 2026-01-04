//go:build js && wasm

// Package main provides WASM bridge functionality.
//
// This file provides a WASM-mode Strava client adapter that bridges to JS stravaFetch.
// The adapter uses JavaScript callbacks to handle async API calls.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/importer"
)

// WasmStravaAdapter implements importer.StravaClient by calling JavaScript's stravaFetch.
// It uses Promise-based callbacks to handle async JS calls from synchronous Go code.
type WasmStravaAdapter struct {
	mu sync.Mutex

	// athleteID and athleteData are cached from JS
	athleteID   int64
	athleteData *importer.Athlete
}

// NewWasmStravaAdapter creates a new WASM Strava client adapter.
func NewWasmStravaAdapter() *WasmStravaAdapter {
	return &WasmStravaAdapter{}
}

// SetAthlete sets the athlete data from JS side.
func (a *WasmStravaAdapter) SetAthlete(athlete *importer.Athlete) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.athleteData = athlete
	if athlete != nil {
		a.athleteID = athlete.ID
	}
}

// GetAthlete returns the currently authenticated athlete.
func (a *WasmStravaAdapter) GetAthlete() *importer.Athlete {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.athleteData
}

// callStravaFetch calls JavaScript's stravaFetch and waits for the result.
// Uses a channel to synchronize the async JS call with Go.
func (a *WasmStravaAdapter) callStravaFetch(path string) (json.RawMessage, error) {
	// Get the stravaFetch function from JS
	stravaFetchJS := js.Global().Get("stravaFetch")
	if !stravaFetchJS.Truthy() {
		return nil, fmt.Errorf("stravaFetch not available in JS global scope")
	}

	// Create a channel to receive the result
	resultCh := make(chan struct {
		data json.RawMessage
		err  error
	}, 1)

	// Create success callback
	onSuccess := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) == 0 {
			resultCh <- struct {
				data json.RawMessage
				err  error
			}{nil, nil}
			return nil
		}

		// Convert JS object to JSON string
		jsonStr := js.Global().Get("JSON").Call("stringify", args[0]).String()
		resultCh <- struct {
			data json.RawMessage
			err  error
		}{json.RawMessage(jsonStr), nil}
		return nil
	})
	defer onSuccess.Release()

	// Create error callback
	onError := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errMsg := "unknown error"
		if len(args) > 0 {
			errMsg = args[0].String()
		}
		resultCh <- struct {
			data json.RawMessage
			err  error
		}{nil, fmt.Errorf("strava API error: %s", errMsg)}
		return nil
	})
	defer onError.Release()

	// Call stravaFetch with path, then attach .then() and .catch()
	promise := stravaFetchJS.Invoke(path)
	promise.Call("then", onSuccess).Call("catch", onError)

	// Wait for result
	result := <-resultCh
	return result.data, result.err
}

// GetActivities fetches a page of activities.
func (a *WasmStravaAdapter) GetActivities(ctx context.Context, opts importer.GetActivitiesOptions) ([]importer.Activity, error) {
	path := fmt.Sprintf("/athlete/activities?page=%d&per_page=%d", opts.Page, opts.PerPage)
	if opts.After != nil {
		path += fmt.Sprintf("&after=%d", opts.After.Unix())
	}
	if opts.Before != nil {
		path += fmt.Sprintf("&before=%d", opts.Before.Unix())
	}

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var activities []stravaActivityJSON
	if err := json.Unmarshal(data, &activities); err != nil {
		return nil, fmt.Errorf("parsing activities: %w", err)
	}

	result := make([]importer.Activity, len(activities))
	for i, act := range activities {
		result[i] = convertJSONActivity(&act)
	}
	return result, nil
}

// GetActivity fetches a single activity with full details.
func (a *WasmStravaAdapter) GetActivity(ctx context.Context, id int64) (*importer.Activity, error) {
	path := fmt.Sprintf("/activities/%d?include_all_efforts=true", id)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var act stravaActivityJSON
	if err := json.Unmarshal(data, &act); err != nil {
		return nil, fmt.Errorf("parsing activity: %w", err)
	}

	result := convertJSONActivity(&act)
	return &result, nil
}

// GetActivityStreams fetches stream data for an activity.
func (a *WasmStravaAdapter) GetActivityStreams(ctx context.Context, id int64, types []string) (*importer.StreamSet, error) {
	streamTypes := "time,distance,latlng,altitude,heartrate,cadence,watts,temp,velocity_smooth"
	path := fmt.Sprintf("/activities/%d/streams?keys=%s&key_by_type=true", id, streamTypes)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var streams stravaStreamSetJSON
	if err := json.Unmarshal(data, &streams); err != nil {
		return nil, fmt.Errorf("parsing streams: %w", err)
	}

	return convertJSONStreamSet(&streams), nil
}

// GetGear fetches gear details by ID.
func (a *WasmStravaAdapter) GetGear(ctx context.Context, id string) (*importer.Gear, error) {
	path := fmt.Sprintf("/gear/%s", id)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var gear stravaGearJSON
	if err := json.Unmarshal(data, &gear); err != nil {
		return nil, fmt.Errorf("parsing gear: %w", err)
	}

	return &importer.Gear{
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
func (a *WasmStravaAdapter) GetSegment(ctx context.Context, id int64) (*importer.Segment, error) {
	path := fmt.Sprintf("/segments/%d", id)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var seg stravaSegmentJSON
	if err := json.Unmarshal(data, &seg); err != nil {
		return nil, fmt.Errorf("parsing segment: %w", err)
	}

	return convertJSONSegment(&seg), nil
}

// GetActivityPhotos fetches photos for an activity.
func (a *WasmStravaAdapter) GetActivityPhotos(ctx context.Context, id int64) ([]importer.Photo, error) {
	path := fmt.Sprintf("/activities/%d/photos?size=600", id)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var photos []stravaPhotoJSON
	if err := json.Unmarshal(data, &photos); err != nil {
		return nil, fmt.Errorf("parsing photos: %w", err)
	}

	result := make([]importer.Photo, len(photos))
	for i, p := range photos {
		result[i] = importer.Photo{
			UniqueID:   p.UniqueID,
			ActivityID: id,
			URLs:       p.URLs,
			Caption:    p.Caption,
			Location:   p.Location,
			CreatedAt:  time.Time{},
		}
	}
	return result, nil
}

// RateLimitInfo returns current rate limit status from JS.
func (a *WasmStravaAdapter) RateLimitInfo() importer.RateLimitInfo {
	// Get rate limit info from JS
	getRateLimitInfo := js.Global().Get("getRateLimitInfo")
	if !getRateLimitInfo.Truthy() {
		return importer.RateLimitInfo{}
	}

	info := getRateLimitInfo.Invoke()
	return importer.RateLimitInfo{
		Used15Min:  info.Get("usage15Min").Int(),
		Limit15Min: info.Get("limit15Min").Int(),
		UsedDaily:  info.Get("usageDaily").Int(),
		LimitDaily: info.Get("limitDaily").Int(),
	}
}

// ============================================================================
// JSON Types for Strava API responses
// ============================================================================

type stravaActivityJSON struct {
	ID                   int64                     `json:"id"`
	Name                 string                    `json:"name"`
	SportType            string                    `json:"sport_type"`
	Type                 string                    `json:"type"`
	StartDate            string                    `json:"start_date"`
	StartDateLocal       string                    `json:"start_date_local"`
	Timezone             string                    `json:"timezone"`
	Distance             float64                   `json:"distance"`
	MovingTime           int                       `json:"moving_time"`
	ElapsedTime          int                       `json:"elapsed_time"`
	TotalElevationGain   float64                   `json:"total_elevation_gain"`
	AverageSpeed         float64                   `json:"average_speed"`
	MaxSpeed             float64                   `json:"max_speed"`
	AverageHeartrate     *float64                  `json:"average_heartrate,omitempty"`
	MaxHeartrate         *float64                  `json:"max_heartrate,omitempty"`
	AverageWatts         *float64                  `json:"average_watts,omitempty"`
	MaxWatts             *float64                  `json:"max_watts,omitempty"`
	WeightedAverageWatts *float64                  `json:"weighted_average_watts,omitempty"`
	Kilojoules           *float64                  `json:"kilojoules,omitempty"`
	AverageCadence       *float64                  `json:"average_cadence,omitempty"`
	Calories             *float64                  `json:"calories,omitempty"`
	GearID               string                    `json:"gear_id,omitempty"`
	Commute              bool                      `json:"commute"`
	Trainer              bool                      `json:"trainer"`
	Private              bool                      `json:"private"`
	WorkoutType          *int                      `json:"workout_type,omitempty"`
	LocationCity         string                    `json:"location_city,omitempty"`
	LocationState        string                    `json:"location_state,omitempty"`
	LocationCountry      string                    `json:"location_country,omitempty"`
	Map                  *stravaMapJSON            `json:"map,omitempty"`
	StartLatlng          []float64                 `json:"start_latlng,omitempty"`
	Description          string                    `json:"description,omitempty"`
	DeviceName           string                    `json:"device_name,omitempty"`
	KudosCount           int                       `json:"kudos_count"`
	PhotoCount           int                       `json:"total_photo_count"`
	SegmentEfforts       []stravaSegmentEffortJSON `json:"segment_efforts,omitempty"`
	BestEfforts          []stravaBestEffortJSON    `json:"best_efforts,omitempty"`
}

type stravaMapJSON struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	SummaryPolyline string `json:"summary_polyline"`
}

type stravaSegmentEffortJSON struct {
	ID               int64             `json:"id"`
	Segment          stravaSegmentJSON `json:"segment"`
	Name             string            `json:"name"`
	ElapsedTime      int               `json:"elapsed_time"`
	MovingTime       int               `json:"moving_time"`
	StartDate        string            `json:"start_date"`
	StartDateLocal   string            `json:"start_date_local"`
	Distance         float64           `json:"distance"`
	AverageWatts     float64           `json:"average_watts"`
	AverageHeartrate float64           `json:"average_heartrate"`
	MaxHeartrate     float64           `json:"max_heartrate"`
	PRRank           *int              `json:"pr_rank,omitempty"`
}

type stravaBestEffortJSON struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	ElapsedTime int     `json:"elapsed_time"`
	MovingTime  int     `json:"moving_time"`
	StartDate   string  `json:"start_date"`
	Distance    float64 `json:"distance"`
	PRRank      *int    `json:"pr_rank,omitempty"`
	StartIndex  *int    `json:"start_index,omitempty"`
	EndIndex    *int    `json:"end_index,omitempty"`
}

type stravaSegmentJSON struct {
	ID            int64                   `json:"id"`
	Name          string                  `json:"name"`
	ActivityType  string                  `json:"activity_type"`
	Distance      float64                 `json:"distance"`
	AverageGrade  float64                 `json:"average_grade"`
	MaximumGrade  float64                 `json:"maximum_grade"`
	ElevationHigh float64                 `json:"elevation_high"`
	ElevationLow  float64                 `json:"elevation_low"`
	ClimbCategory int                     `json:"climb_category"`
	StartLatlng   []float64               `json:"start_latlng,omitempty"`
	EndLatlng     []float64               `json:"end_latlng,omitempty"`
	Starred       bool                    `json:"starred"`
	Map           *stravaMapJSON          `json:"map,omitempty"`
	AthleteStats  *stravaSegmentStatsJSON `json:"athlete_segment_stats,omitempty"`
}

type stravaSegmentStatsJSON struct {
	PRElapsedTime int    `json:"pr_elapsed_time"`
	PRDate        string `json:"pr_date,omitempty"`
	EffortCount   int    `json:"effort_count"`
	KOMRank       *int   `json:"kom_rank,omitempty"`
}

type stravaGearJSON struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Primary     bool    `json:"primary"`
	Retired     bool    `json:"retired"`
	Distance    float64 `json:"distance"`
	BrandName   string  `json:"brand_name"`
	ModelName   string  `json:"model_name"`
	Description string  `json:"description"`
}

type stravaStreamSetJSON struct {
	Time           *stravaStreamJSON `json:"time,omitempty"`
	Distance       *stravaStreamJSON `json:"distance,omitempty"`
	Altitude       *stravaStreamJSON `json:"altitude,omitempty"`
	Heartrate      *stravaStreamJSON `json:"heartrate,omitempty"`
	Watts          *stravaStreamJSON `json:"watts,omitempty"`
	Cadence        *stravaStreamJSON `json:"cadence,omitempty"`
	VelocitySmooth *stravaStreamJSON `json:"velocity_smooth,omitempty"`
	Latlng         *stravaStreamJSON `json:"latlng,omitempty"`
}

type stravaStreamJSON struct {
	Type         string      `json:"type"`
	Data         interface{} `json:"data"`
	OriginalSize int         `json:"original_size"`
	Resolution   string      `json:"resolution"`
	SeriesType   string      `json:"series_type"`
}

type stravaPhotoJSON struct {
	UniqueID   string            `json:"unique_id"`
	ActivityID int64             `json:"activity_id"`
	URLs       map[string]string `json:"urls"`
	Caption    string            `json:"caption"`
	Location   []float64         `json:"location"`
	CreatedAt  string            `json:"created_at"`
}

// ============================================================================
// Conversion Helpers
// ============================================================================

func convertJSONActivity(a *stravaActivityJSON) importer.Activity {
	sportType := a.SportType
	if sportType == "" {
		sportType = a.Type
	}

	result := importer.Activity{
		ID:                   a.ID,
		Name:                 a.Name,
		SportType:            sportType,
		Distance:             a.Distance,
		MovingTime:           a.MovingTime,
		ElapsedTime:          a.ElapsedTime,
		TotalElevationGain:   a.TotalElevationGain,
		AverageSpeed:         a.AverageSpeed,
		MaxSpeed:             a.MaxSpeed,
		AverageHeartrate:     a.AverageHeartrate,
		MaxHeartrate:         a.MaxHeartrate,
		AverageWatts:         a.AverageWatts,
		MaxWatts:             a.MaxWatts,
		WeightedAverageWatts: a.WeightedAverageWatts,
		Kilojoules:           a.Kilojoules,
		AverageCadence:       a.AverageCadence,
		Calories:             a.Calories,
		GearID:               a.GearID,
		Commute:              a.Commute,
		Trainer:              a.Trainer,
		Private:              a.Private,
		WorkoutType:          a.WorkoutType,
		LocationCity:         a.LocationCity,
		LocationState:        a.LocationState,
		LocationCountry:      a.LocationCountry,
		Description:          a.Description,
		DeviceName:           a.DeviceName,
		KudosCount:           a.KudosCount,
		PhotoCount:           a.PhotoCount,
		Timezone:             a.Timezone,
	}

	// Parse dates
	if t, err := time.Parse(time.RFC3339, a.StartDate); err == nil {
		result.StartDate = t
	}
	if t, err := time.Parse(time.RFC3339, a.StartDateLocal); err == nil {
		result.StartDateLocal = t
	}

	// Handle map
	if a.Map != nil {
		result.SummaryPolyline = a.Map.SummaryPolyline
	}

	// Handle location
	if len(a.StartLatlng) >= 2 {
		result.StartLat = &a.StartLatlng[0]
		result.StartLng = &a.StartLatlng[1]
	}

	// Convert segment efforts
	result.SegmentEfforts = make([]importer.SegmentEffort, len(a.SegmentEfforts))
	for i, e := range a.SegmentEfforts {
		result.SegmentEfforts[i] = convertJSONSegmentEffort(&e)
	}

	// Convert best efforts
	result.BestEfforts = make([]importer.BestEffort, len(a.BestEfforts))
	for i, e := range a.BestEfforts {
		result.BestEfforts[i] = convertJSONBestEffort(&e)
	}

	return result
}

func convertJSONSegment(s *stravaSegmentJSON) *importer.Segment {
	result := &importer.Segment{
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
	}

	if s.Map != nil {
		result.Polyline = s.Map.Polyline
	}

	if s.AthleteStats != nil {
		result.AthleteSegmentStats.PRElapsedTime = s.AthleteStats.PRElapsedTime
		result.AthleteSegmentStats.EffortCount = s.AthleteStats.EffortCount
		result.AthleteSegmentStats.KOMRank = s.AthleteStats.KOMRank

		if s.AthleteStats.PRDate != "" {
			if t, err := time.Parse("2006-01-02", s.AthleteStats.PRDate); err == nil {
				result.AthleteSegmentStats.PRDate = &t
			}
		}
	}

	return result
}

func convertJSONSegmentEffort(e *stravaSegmentEffortJSON) importer.SegmentEffort {
	result := importer.SegmentEffort{
		ID:               e.ID,
		Name:             e.Name,
		ElapsedTime:      e.ElapsedTime,
		MovingTime:       e.MovingTime,
		Distance:         e.Distance,
		AverageWatts:     e.AverageWatts,
		AverageHeartrate: e.AverageHeartrate,
		MaxHeartrate:     e.MaxHeartrate,
		PRRank:           e.PRRank,
	}

	result.Segment = *convertJSONSegment(&e.Segment)

	if t, err := time.Parse(time.RFC3339, e.StartDate); err == nil {
		result.StartDate = t
	}
	if t, err := time.Parse(time.RFC3339, e.StartDateLocal); err == nil {
		result.StartDateLocal = t
	}

	return result
}

func convertJSONBestEffort(e *stravaBestEffortJSON) importer.BestEffort {
	result := importer.BestEffort{
		ID:          e.ID,
		Name:        e.Name,
		ElapsedTime: e.ElapsedTime,
		MovingTime:  e.MovingTime,
		Distance:    e.Distance,
		PRRank:      e.PRRank,
		StartIndex:  e.StartIndex,
		EndIndex:    e.EndIndex,
	}

	if t, err := time.Parse(time.RFC3339, e.StartDate); err == nil {
		result.StartDate = t
	}

	return result
}

func convertJSONStreamSet(s *stravaStreamSetJSON) *importer.StreamSet {
	result := &importer.StreamSet{}

	if s.Time != nil {
		result.Time = convertJSONStream(s.Time)
	}
	if s.Distance != nil {
		result.Distance = convertJSONStream(s.Distance)
	}
	if s.Altitude != nil {
		result.Altitude = convertJSONStream(s.Altitude)
	}
	if s.Heartrate != nil {
		result.Heartrate = convertJSONStream(s.Heartrate)
	}
	if s.Watts != nil {
		result.Watts = convertJSONStream(s.Watts)
	}
	if s.Cadence != nil {
		result.Cadence = convertJSONStream(s.Cadence)
	}
	if s.VelocitySmooth != nil {
		result.VelocitySmooth = convertJSONStream(s.VelocitySmooth)
	}
	if s.Latlng != nil {
		result.Latlng = convertJSONStream(s.Latlng)
	}

	return result
}

func convertJSONStream(s *stravaStreamJSON) *importer.Stream {
	return &importer.Stream{
		Type:         s.Type,
		Data:         s.Data,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
	}
}
