//go:build js && wasm

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// ============================================================================
// Segments Write
// ============================================================================

//wasm:category Segments - Write

// saveSegment stores a segment in the database
// Called from JS: goStorage.saveSegment(segmentJSON)
//wasm:export
func saveSegment(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("saveSegment")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing segment JSON"))
	}

	segmentJSON := args[0].String()
	var req struct {
		ID                   int64    `json:"id"`
		Name                 string   `json:"name"`
		ActivityType         string   `json:"activity_type"`
		Distance             float64  `json:"distance"`
		AverageGrade         float64  `json:"average_grade"`
		MaximumGrade         float64  `json:"maximum_grade"`
		ElevationHigh        float64  `json:"elevation_high"`
		ElevationLow         float64  `json:"elevation_low"`
		ClimbCategory        int      `json:"climb_category"`
		StartLat             *float64 `json:"start_lat"`
		StartLng             *float64 `json:"start_lng"`
		EndLat               *float64 `json:"end_lat"`
		EndLng               *float64 `json:"end_lng"`
		Starred              bool     `json:"starred"`
		Polyline             string   `json:"polyline"`
		AthleteKOMRank       *int     `json:"athlete_kom_rank"`
		AthleteEffortCount   *int     `json:"athlete_effort_count"`
		AthletePRElapsedTime *int     `json:"athlete_pr_elapsed_time"`
		AthletePRDate        string   `json:"athlete_pr_date"`
	}
	if err := json.Unmarshal([]byte(segmentJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing segment: %w", err))
	}

	seg := &storage.Segment{
		ID:                   req.ID,
		Name:                 req.Name,
		ActivityType:         req.ActivityType,
		Distance:             req.Distance,
		AverageGrade:         req.AverageGrade,
		MaximumGrade:         req.MaximumGrade,
		ElevationHigh:        req.ElevationHigh,
		ElevationLow:         req.ElevationLow,
		ClimbCategory:        req.ClimbCategory,
		StartLat:             req.StartLat,
		StartLng:             req.StartLng,
		EndLat:               req.EndLat,
		EndLng:               req.EndLng,
		Starred:              req.Starred,
		Polyline:             req.Polyline,
		AthleteKOMRank:       req.AthleteKOMRank,
		AthleteEffortCount:   req.AthleteEffortCount,
		AthletePRElapsedTime: req.AthletePRElapsedTime,
	}

	// Parse PR date if provided
	if req.AthletePRDate != "" {
		if t, err := time.Parse(time.RFC3339, req.AthletePRDate); err == nil {
			seg.AthletePRDate = &storage.SQLiteTime{Time: t}
		}
	}

	ctx := context.Background()
	if err := bridge.segments.UpsertSegment(ctx, seg); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Segment %d saved", req.ID))
}

// saveSegmentEffort stores a segment effort in the database
// Called from JS: goStorage.saveSegmentEffort(effortJSON)
//wasm:export
func saveSegmentEffort(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("saveSegmentEffort")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing segment effort JSON"))
	}

	effortJSON := args[0].String()
	var req struct {
		ID               int64    `json:"id"`
		SegmentID        int64    `json:"segment_id"`
		ActivityID       int64    `json:"activity_id"`
		AthleteID        int64    `json:"athlete_id"`
		Name             string   `json:"name"`
		ElapsedTime      int      `json:"elapsed_time"`
		MovingTime       int      `json:"moving_time"`
		StartDate        string   `json:"start_date"`
		StartDateLocal   string   `json:"start_date_local"`
		Distance         float64  `json:"distance"`
		AverageWatts     *float64 `json:"average_watts"`
		AverageHeartrate *float64 `json:"average_heartrate"`
		MaxHeartrate     *int     `json:"max_heartrate"`
		PRRank           *int     `json:"pr_rank"`
	}
	if err := json.Unmarshal([]byte(effortJSON), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing segment effort: %w", err))
	}

	effort := &storage.SegmentEffort{
		ID:               req.ID,
		SegmentID:        req.SegmentID,
		ActivityID:       req.ActivityID,
		AthleteID:        req.AthleteID,
		Name:             req.Name,
		ElapsedTime:      req.ElapsedTime,
		MovingTime:       req.MovingTime,
		Distance:         req.Distance,
		AverageWatts:     req.AverageWatts,
		AverageHeartrate: req.AverageHeartrate,
		MaxHeartrate:     req.MaxHeartrate,
		PRRank:           req.PRRank,
	}

	// Parse dates
	if req.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, req.StartDate); err == nil {
			effort.StartDate = &storage.SQLiteTime{Time: t}
		}
	}
	if req.StartDateLocal != "" {
		if t, err := time.Parse(time.RFC3339, req.StartDateLocal); err == nil {
			effort.StartDateLocal = &storage.SQLiteTime{Time: t}
		}
	}

	ctx := context.Background()
	if err := bridge.segments.UpsertEffort(ctx, effort); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Segment effort %d saved", req.ID))
}

// ============================================================================
// Segments Read
// ============================================================================

//wasm:category Segments - Read

// getSegments retrieves paginated segments list
// Called from JS: goStorage.getSegments(filtersJSON)
//wasm:export
func getSegments(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSegments")

	var req struct {
		Page         int    `json:"page"`
		PerPage      int    `json:"per_page"`
		Starred      *bool  `json:"starred"`
		Search       string `json:"search"`
		ActivityType string `json:"activity_type"`
		Country      string `json:"country"`
		KOMOnly      bool   `json:"kom_only"`
		OrderBy      string `json:"order_by"`
		OrderDir     string `json:"order_dir"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing segment filters: %w", err))
		}
	}

	ctx := context.Background()
	result, err := bridge.segmentsService.List(ctx, services.ListSegmentsInput{
		AthleteID:    bridge.athleteID,
		ActivityType: req.ActivityType,
		Country:      req.Country,
		Search:       req.Search,
		KOMOnly:      req.KOMOnly,
		Starred:      req.Starred,
		Page:         req.Page,
		PerPage:      req.PerPage,
		OrderBy:      req.OrderBy,
		OrderDir:     req.OrderDir,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// getSegmentDetail retrieves a single segment by ID with its effort history
// Called from JS: goStorage.getSegmentDetail(id)
//wasm:export
func getSegmentDetail(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSegmentDetail")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing segment ID"))
	}

	id := int64(args[0].Int())
	ctx := context.Background()

	result, err := bridge.segmentsService.GetByID(ctx, services.GetSegmentInput{
		AthleteID: bridge.athleteID,
		SegmentID: id,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// ============================================================================
// Segment Efforts
// ============================================================================

//wasm:category Segment Efforts

// getSegmentEfforts returns paginated segment efforts for a segment
// Called from JS: goStorage.getSegmentEfforts(filtersJSON)
//wasm:export
func getSegmentEfforts(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSegmentEfforts")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing filters"))
	}

	var req struct {
		SegmentID int64  `json:"segment_id"`
		Page      int    `json:"page"`
		PerPage   int    `json:"per_page"`
		OrderBy   string `json:"order_by"`
		OrderDir  string `json:"order_dir"`
	}
	if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
		return errorJSON(fmt.Errorf("parsing filters: %w", err))
	}

	ctx := context.Background()
	result, err := bridge.segmentsService.ListEfforts(ctx, services.ListEffortsInput{
		AthleteID: bridge.athleteID,
		SegmentID: req.SegmentID,
		Page:      req.Page,
		PerPage:   req.PerPage,
		OrderBy:   req.OrderBy,
		OrderDir:  req.OrderDir,
	})
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}

// getSegmentCountries returns country statistics for segments
// Called from JS: goStorage.getSegmentCountries()
//wasm:export
func getSegmentCountries(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSegmentCountries")

	ctx := context.Background()
	result, err := bridge.segmentsService.ListCountries(ctx, bridge.athleteID)
	if err != nil {
		return errorJSON(err)
	}

	return dataJSON(result)
}
