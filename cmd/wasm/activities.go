//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Auth
// ============================================================================

// getAuthStatus returns the current authentication status
// Called from JS: goStorage.getAuthStatus()
func getAuthStatus(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getAuthStatus")

	ctx := context.Background()

	// Check if we have an athlete
	if athleteID == 0 {
		return toJSON(map[string]interface{}{
			"authenticated": false,
		})
	}

	// Get athlete info
	athlete, err := athletes.GetByID(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	if athlete == nil {
		return toJSON(map[string]interface{}{
			"authenticated": false,
		})
	}

	return toJSON(map[string]interface{}{
		"authenticated": true,
		"athlete": map[string]interface{}{
			"id":        athlete.ID,
			"username":  athlete.Username,
			"firstname": athlete.FirstName,
			"lastname":  athlete.LastName,
			"profile":   athlete.ProfileMedium,
		},
	})
}

// ============================================================================
// Activities Read
// ============================================================================

// getActivities retrieves activities with filters and pagination
// Called from JS: goStorage.getActivities(filtersJSON)
func getActivities(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getActivities")

	if err := ensureInitialized(); err != nil {
		return errorJSON(err)
	}
	if err := ensureAthleteID(); err != nil {
		return errorJSON(err)
	}

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing filters"))
	}

	// Parse filters from JSON
	filtersJSON := args[0].String()
	var req struct {
		Page      int    `json:"page"`
		PerPage   int    `json:"per_page"`
		SportType string `json:"sport_type"`
		After     string `json:"after"`
		Before    string `json:"before"`
		GearID    string `json:"gear_id"`
		Commute   *bool  `json:"commute"`
		Trainer   *bool  `json:"trainer"`
		Search    string `json:"search"`
	}
	if err := json.Unmarshal([]byte(filtersJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing filters: %w", err))
	}

	// Build storage filters
	filters := storage.ActivityFilters{
		AthleteID: athleteID,
		GearID:    req.GearID,
		Commute:   req.Commute,
		Trainer:   req.Trainer,
		Search:    req.Search,
	}

	if req.SportType != "" {
		filters.SportTypes = []string{req.SportType}
	}

	if req.After != "" {
		t, err := time.Parse(time.RFC3339, req.After)
		if err == nil {
			filters.StartAfter = &t
		}
	}

	if req.Before != "" {
		t, err := time.Parse(time.RFC3339, req.Before)
		if err == nil {
			filters.StartBefore = &t
		}
	}

	// Pagination
	page := storage.Pagination{
		Page:    req.Page,
		PerPage: req.PerPage,
	}
	if page.Page == 0 {
		page.Page = 1
	}
	if page.PerPage == 0 {
		page.PerPage = 50
	}

	ctx := context.Background()
	activityList, total, err := activities.List(ctx, filters, page)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	data := make([]map[string]interface{}, len(activityList))
	for i, a := range activityList {
		data[i] = activityToMap(a)
	}

	totalPages := (total + page.PerPage - 1) / page.PerPage

	return dataJSON(map[string]interface{}{
		"data":        data,
		"total":       total,
		"page":        page.Page,
		"per_page":    page.PerPage,
		"total_pages": totalPages,
	})
}

// getActivity retrieves a single activity by ID
// Called from JS: goStorage.getActivity(id)
func getActivity(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getActivity")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing activity ID"))
	}

	id := int64(args[0].Int())
	ctx := context.Background()

	activity, err := activities.GetByID(ctx, id)
	if err != nil {
		return errorJSON(err)
	}

	if activity == nil {
		return errorJSON(fmt.Errorf("activity not found"))
	}

	return dataJSON(activityToMap(*activity))
}

// getActivityStreams retrieves streams for an activity
// Called from JS: goStorage.getActivityStreams(activityId)
func getActivityStreams(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getActivityStreams")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing activity ID"))
	}

	activityID := int64(args[0].Int())
	ctx := context.Background()

	streamList, err := streams.GetByActivityID(ctx, activityID)
	if err != nil {
		return errorJSON(err)
	}

	// Convert streams to response format
	data := make([]map[string]interface{}, len(streamList))
	for i, s := range streamList {
		data[i] = map[string]interface{}{
			"type":       s.StreamType,
			"data":       s.Data,
			"resolution": s.Resolution,
		}
	}

	return dataJSON(data)
}

// ============================================================================
// Activities Write
// ============================================================================

// saveActivity stores an activity in the database
// Called from JS: goStorage.saveActivity(activityJSON)
func saveActivity(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("saveActivity")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing activity JSON"))
	}

	// Parse activity from JSON
	activityJSON := args[0].String()
	var req struct {
		ID                 int64    `json:"id"`
		AthleteID          int64    `json:"athlete_id"`
		Name               string   `json:"name"`
		SportType          string   `json:"sport_type"`
		StartDate          string   `json:"start_date"`
		StartDateLocal     string   `json:"start_date_local"`
		Timezone           string   `json:"timezone"`
		Distance           float64  `json:"distance"`
		MovingTime         int      `json:"moving_time"`
		ElapsedTime        int      `json:"elapsed_time"`
		TotalElevationGain float64  `json:"total_elevation_gain"`
		AverageSpeed       float64  `json:"average_speed"`
		MaxSpeed           float64  `json:"max_speed"`
		AverageHeartrate   *float64 `json:"average_heartrate"`
		MaxHeartrate       *float64 `json:"max_heartrate"`
		AverageWatts       *float64 `json:"average_watts"`
		MaxWatts           *float64 `json:"max_watts"`
		WeightedAvgWatts   *float64 `json:"weighted_average_watts"`
		Kilojoules         *float64 `json:"kilojoules"`
		AverageCadence     *float64 `json:"average_cadence"`
		Calories           *float64 `json:"calories"`
		GearID             string   `json:"gear_id"`
		Commute            bool     `json:"commute"`
		WorkoutType        *int     `json:"workout_type"`
		LocationCity       string   `json:"location_city"`
		LocationState      string   `json:"location_state"`
		LocationCountry    string   `json:"location_country"`
		SummaryPolyline    string   `json:"summary_polyline"`
		StartLat           *float64 `json:"start_lat"`
		StartLng           *float64 `json:"start_lng"`
		Description        string   `json:"description"`
		DeviceName         string   `json:"device_name"`
		Trainer            bool     `json:"trainer"`
		Private            bool     `json:"private"`
		KudosCount         int      `json:"kudos_count"`
		PhotoCount         int      `json:"photo_count"`
	}
	if err := json.Unmarshal([]byte(activityJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing activity: %w", err))
	}

	// Parse dates - these are required fields
	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return errorJSON(fmt.Errorf("parsing start_date %q: %w", req.StartDate, err))
	}
	startDateLocal, err := time.Parse(time.RFC3339, req.StartDateLocal)
	if err != nil {
		return errorJSON(fmt.Errorf("parsing start_date_local %q: %w", req.StartDateLocal, err))
	}

	activity := &storage.Activity{
		ID:                   req.ID,
		AthleteID:            req.AthleteID,
		Name:                 req.Name,
		SportType:            req.SportType,
		StartDate:            storage.SQLiteTime{Time: startDate},
		StartDateLocal:       storage.SQLiteTime{Time: startDateLocal},
		Timezone:             req.Timezone,
		Distance:             req.Distance,
		MovingTime:           req.MovingTime,
		ElapsedTime:          req.ElapsedTime,
		TotalElevationGain:   req.TotalElevationGain,
		AverageSpeed:         req.AverageSpeed,
		MaxSpeed:             req.MaxSpeed,
		AverageHeartrate:     req.AverageHeartrate,
		MaxHeartrate:         req.MaxHeartrate,
		AverageWatts:         req.AverageWatts,
		MaxWatts:             req.MaxWatts,
		WeightedAverageWatts: req.WeightedAvgWatts,
		Kilojoules:           req.Kilojoules,
		AverageCadence:       req.AverageCadence,
		Calories:             req.Calories,
		GearID:               req.GearID,
		Commute:              req.Commute,
		WorkoutType:          req.WorkoutType,
		LocationCity:         req.LocationCity,
		LocationState:        req.LocationState,
		LocationCountry:      req.LocationCountry,
		SummaryPolyline:      req.SummaryPolyline,
		StartLat:             req.StartLat,
		StartLng:             req.StartLng,
		Description:          req.Description,
		DeviceName:           req.DeviceName,
		Trainer:              req.Trainer,
		Private:              req.Private,
		KudosCount:           req.KudosCount,
		PhotoCount:           req.PhotoCount,
	}

	ctx := context.Background()
	if err := activities.Upsert(ctx, activity); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Activity %d saved", req.ID))
}

// saveStream stores an activity stream in the database
// Called from JS: goStorage.saveStream(streamJSON)
func saveStream(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("saveStream")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing stream JSON"))
	}

	streamJSON := args[0].String()
	var req struct {
		ActivityID   int64       `json:"activity_id"`
		StreamType   string      `json:"stream_type"`
		Data         interface{} `json:"data"`
		SeriesType   string      `json:"series_type"`
		OriginalSize int         `json:"original_size"`
		Resolution   string      `json:"resolution"`
	}
	if err := json.Unmarshal([]byte(streamJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing stream: %w", err))
	}

	// Validate original size to prevent memory exhaustion
	if req.OriginalSize > maxStreamDataSize {
		return errorJSON(fmt.Errorf("stream original_size exceeds maximum: %d > %d", req.OriginalSize, maxStreamDataSize))
	}

	// Convert data to JSON for storage
	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		return errorJSON(fmt.Errorf("marshaling data: %w", err))
	}

	stream := &storage.ActivityStream{
		ActivityID:   req.ActivityID,
		StreamType:   req.StreamType,
		Data:         json.RawMessage(dataJSON),
		SeriesType:   req.SeriesType,
		OriginalSize: req.OriginalSize,
		Resolution:   req.Resolution,
	}

	ctx := context.Background()
	if err := streams.Upsert(ctx, stream); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Stream %s for activity %d saved", req.StreamType, req.ActivityID))
}

// ============================================================================
// Activity Map Helper
// ============================================================================

// activityToMap converts an Activity to a map for JSON serialization
func activityToMap(a storage.Activity) map[string]interface{} {
	m := map[string]interface{}{
		"id":                   a.ID,
		"athlete_id":           a.AthleteID,
		"name":                 a.Name,
		"sport_type":           a.SportType,
		"start_date":           a.StartDate.Format(time.RFC3339),
		"start_date_local":     a.StartDateLocal.Format(time.RFC3339),
		"timezone":             a.Timezone,
		"distance":             a.Distance,
		"moving_time":          a.MovingTime,
		"elapsed_time":         a.ElapsedTime,
		"total_elevation_gain": a.TotalElevationGain,
		"average_speed":        a.AverageSpeed,
		"max_speed":            a.MaxSpeed,
		"kudos_count":          a.KudosCount,
		"comment_count":        a.CommentCount,
		"photo_count":          a.PhotoCount,
		"commute":              a.Commute,
		"private":              a.Private,
		"trainer":              a.Trainer,
		"gear_id":              a.GearID,
		"summary_polyline":     a.SummaryPolyline,
	}

	// Add optional fields
	if a.Description != "" {
		m["description"] = a.Description
	}
	if a.LocationCity != "" {
		m["location_city"] = a.LocationCity
	}
	if a.LocationState != "" {
		m["location_state"] = a.LocationState
	}
	if a.LocationCountry != "" {
		m["location_country"] = a.LocationCountry
	}
	if a.AverageHeartrate != nil {
		m["average_heartrate"] = *a.AverageHeartrate
	}
	if a.MaxHeartrate != nil {
		m["max_heartrate"] = *a.MaxHeartrate
	}
	if a.AverageWatts != nil {
		m["average_watts"] = *a.AverageWatts
	}
	if a.MaxWatts != nil {
		m["max_watts"] = *a.MaxWatts
	}
	if a.WeightedAverageWatts != nil {
		m["weighted_average_watts"] = *a.WeightedAverageWatts
	}
	if a.Kilojoules != nil {
		m["kilojoules"] = *a.Kilojoules
	}
	if a.AverageCadence != nil {
		m["average_cadence"] = *a.AverageCadence
	}
	if a.Calories != nil {
		m["calories"] = *a.Calories
	}
	if a.DeviceName != "" {
		m["device_name"] = a.DeviceName
	}
	if a.WorkoutType != nil {
		m["workout_type"] = *a.WorkoutType
	}
	if a.StartLat != nil {
		m["start_lat"] = *a.StartLat
	}
	if a.StartLng != nil {
		m["start_lng"] = *a.StartLng
	}

	return m
}
