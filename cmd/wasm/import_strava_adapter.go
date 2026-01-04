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

// goWasmNamespace is the namespaced global object for Go WASM interop.
// This must match GO_WASM_NAMESPACE in go-storage.ts.
const goWasmNamespace = "__quantlete_go_wasm__"

// callStravaFetch calls JavaScript's stravaFetch and waits for the result.
// Uses a channel to synchronize the async JS call with Go.
func (a *WasmStravaAdapter) callStravaFetch(path string) (json.RawMessage, error) {
	// Get the stravaFetch function from namespaced JS object
	nsObj := js.Global().Get(goWasmNamespace)
	if !nsObj.Truthy() {
		return nil, fmt.Errorf("Go WASM namespace %q not available in JS global scope", goWasmNamespace)
	}
	stravaFetchJS := nsObj.Get("stravaFetch")
	if !stravaFetchJS.Truthy() {
		return nil, fmt.Errorf("stravaFetch not available in %s namespace", goWasmNamespace)
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

	var activities []importer.ActivityJSON
	if err := json.Unmarshal(data, &activities); err != nil {
		return nil, fmt.Errorf("parsing activities: %w", err)
	}

	result := make([]importer.Activity, len(activities))
	for i := range activities {
		result[i] = activities[i].ToImporter()
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

	var act importer.ActivityJSON
	if err := json.Unmarshal(data, &act); err != nil {
		return nil, fmt.Errorf("parsing activity: %w", err)
	}

	result := act.ToImporter()
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

	var streams importer.StreamSetJSON
	if err := json.Unmarshal(data, &streams); err != nil {
		return nil, fmt.Errorf("parsing streams: %w", err)
	}

	return streams.ToImporter(), nil
}

// GetGear fetches gear details by ID.
func (a *WasmStravaAdapter) GetGear(ctx context.Context, id string) (*importer.Gear, error) {
	path := fmt.Sprintf("/gear/%s", id)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var gear importer.GearJSON
	if err := json.Unmarshal(data, &gear); err != nil {
		return nil, fmt.Errorf("parsing gear: %w", err)
	}

	return gear.ToImporter(), nil
}

// GetSegment fetches segment details by ID.
func (a *WasmStravaAdapter) GetSegment(ctx context.Context, id int64) (*importer.Segment, error) {
	path := fmt.Sprintf("/segments/%d", id)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var seg importer.SegmentJSON
	if err := json.Unmarshal(data, &seg); err != nil {
		return nil, fmt.Errorf("parsing segment: %w", err)
	}

	return seg.ToImporter(), nil
}

// GetActivityPhotos fetches photos for an activity.
func (a *WasmStravaAdapter) GetActivityPhotos(ctx context.Context, id int64) ([]importer.Photo, error) {
	path := fmt.Sprintf("/activities/%d/photos?size=600", id)

	data, err := a.callStravaFetch(path)
	if err != nil {
		return nil, err
	}

	var photos []importer.PhotoJSON
	if err := json.Unmarshal(data, &photos); err != nil {
		return nil, fmt.Errorf("parsing photos: %w", err)
	}

	result := make([]importer.Photo, len(photos))
	for i := range photos {
		result[i] = photos[i].ToImporter(id)
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
