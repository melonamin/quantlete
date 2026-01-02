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
// Segments Write
// ============================================================================

// saveSegment stores a segment in the database
// Called from JS: goStorage.saveSegment(segmentJSON)
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
	if err := segments.UpsertSegment(ctx, seg); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Segment %d saved", req.ID))
}

// saveSegmentEffort stores a segment effort in the database
// Called from JS: goStorage.saveSegmentEffort(effortJSON)
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
	if err := segments.UpsertEffort(ctx, effort); err != nil {
		return errorJSON(err)
	}

	return successJSON(fmt.Sprintf("Segment effort %d saved", req.ID))
}

// ============================================================================
// Segments Read
// ============================================================================

// getSegments retrieves paginated segments list
// Called from JS: goStorage.getSegments(filtersJSON)
func getSegments(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSegments")

	var req struct {
		Page    int    `json:"page"`
		PerPage int    `json:"per_page"`
		Starred *bool  `json:"starred"`
		Search  string `json:"search"`
	}
	if len(args) > 0 && args[0].String() != "" {
		if err := json.Unmarshal([]byte(args[0].String()), &req); err != nil {
			return errorJSON(fmt.Errorf("parsing segment filters: %w", err))
		}
	}

	ctx := context.Background()
	result, err := segments.List(ctx, athleteID, storage.SegmentFilters{
		Starred: req.Starred,
		Search:  req.Search,
	})
	if err != nil {
		return errorJSON(err)
	}

	items := make([]map[string]interface{}, len(result.Items))
	for i, s := range result.Items {
		items[i] = segmentListItemToMap(s)
	}

	return toJSON(map[string]interface{}{
		"ok":          true,
		"data":        items,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// getSegmentDetail retrieves a single segment by ID
// Called from JS: goStorage.getSegmentDetail(id)
func getSegmentDetail(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSegmentDetail")

	if len(args) < 1 {
		return errorJSON(fmt.Errorf("missing segment ID"))
	}

	id := int64(args[0].Int())
	ctx := context.Background()

	seg, err := segments.GetByID(ctx, id)
	if err != nil {
		return errorJSON(err)
	}
	if seg == nil {
		return errorJSON(fmt.Errorf("segment not found"))
	}

	return dataJSON(segmentToMap(*seg))
}

// ============================================================================
// Segment Efforts
// ============================================================================

// getSegmentEfforts returns paginated segment efforts for a segment
// Called from JS: goStorage.getSegmentEfforts(filtersJSON)
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

	if req.SegmentID == 0 {
		return errorJSON(fmt.Errorf("segment_id is required"))
	}

	ctx := context.Background()
	filters := storage.SegmentEffortFilters{}
	filters.Page = req.Page
	filters.PerPage = req.PerPage
	filters.OrderBy = req.OrderBy
	filters.OrderDir = req.OrderDir
	result, err := segments.ListEffortsPaginated(ctx, athleteID, req.SegmentID, filters)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	items := make([]map[string]interface{}, len(result.Items))
	for i, e := range result.Items {
		item := map[string]interface{}{
			"id":           e.ID,
			"segment_id":   e.SegmentID,
			"activity_id":  e.ActivityID,
			"athlete_id":   e.AthleteID,
			"name":         e.Name,
			"elapsed_time": e.ElapsedTime,
			"moving_time":  e.MovingTime,
			"distance":     e.Distance,
			"country":      e.Country,
		}
		if e.StartDate != nil {
			item["start_date"] = e.StartDate.Format(time.RFC3339)
		}
		if e.StartDateLocal != nil {
			item["start_date_local"] = e.StartDateLocal.Format(time.RFC3339)
		}
		if e.AverageWatts != nil {
			item["average_watts"] = *e.AverageWatts
		}
		if e.AverageHeartrate != nil {
			item["average_heartrate"] = *e.AverageHeartrate
		}
		if e.MaxHeartrate != nil {
			item["max_heartrate"] = *e.MaxHeartrate
		}
		if e.PRRank != nil {
			item["pr_rank"] = *e.PRRank
		}
		items[i] = item
	}

	return toJSON(map[string]interface{}{
		"ok":          true,
		"data":        items,
		"total":       result.Total,
		"page":        result.Page,
		"per_page":    result.PerPage,
		"total_pages": result.TotalPages,
	})
}

// getSegmentCountries returns country statistics for segments
// Called from JS: goStorage.getSegmentCountries()
func getSegmentCountries(this js.Value, args []js.Value) interface{} {
	defer recoverPanic("getSegmentCountries")

	ctx := context.Background()
	countryStats, err := segments.ListCountryStats(ctx, athleteID)
	if err != nil {
		return errorJSON(err)
	}

	// Convert to response format
	data := make([]map[string]interface{}, len(countryStats))
	for i, cs := range countryStats {
		data[i] = map[string]interface{}{
			"country": cs.Country,
			"count":   cs.Count,
		}
	}

	return toJSON(map[string]interface{}{
		"ok":   true,
		"data": data,
	})
}

// ============================================================================
// Segment Map Helpers
// ============================================================================

func segmentListItemToMap(s storage.SegmentListItem) map[string]interface{} {
	m := map[string]interface{}{
		"id":              s.ID,
		"name":            s.Name,
		"activity_type":   s.ActivityType,
		"distance":        s.Distance,
		"average_grade":   s.AverageGrade,
		"maximum_grade":   s.MaximumGrade,
		"elevation_high":  s.ElevationHigh,
		"elevation_low":   s.ElevationLow,
		"climb_category":  s.ClimbCategory,
		"starred":         s.Starred,
		"times_completed": s.TimesCompleted,
	}
	if s.StartLat != nil {
		m["start_lat"] = *s.StartLat
	}
	if s.StartLng != nil {
		m["start_lng"] = *s.StartLng
	}
	if s.EndLat != nil {
		m["end_lat"] = *s.EndLat
	}
	if s.EndLng != nil {
		m["end_lng"] = *s.EndLng
	}
	if s.Polyline != "" {
		m["polyline"] = s.Polyline
	}
	if s.AthleteEffortCount != nil {
		m["athlete_effort_count"] = *s.AthleteEffortCount
	}
	if s.AthletePRElapsedTime != nil {
		m["athlete_pr_elapsed_time"] = *s.AthletePRElapsedTime
	}
	if s.AthletePRDate != nil {
		m["athlete_pr_date"] = s.AthletePRDate.Format(time.RFC3339)
	}
	if s.LastEffortDate != nil {
		m["last_effort_date"] = s.LastEffortDate.Format(time.RFC3339)
	}
	if s.BestElapsedTime != nil {
		m["best_elapsed_time"] = *s.BestElapsedTime
	}
	return m
}

func segmentToMap(s storage.Segment) map[string]interface{} {
	m := map[string]interface{}{
		"id":             s.ID,
		"name":           s.Name,
		"activity_type":  s.ActivityType,
		"distance":       s.Distance,
		"average_grade":  s.AverageGrade,
		"maximum_grade":  s.MaximumGrade,
		"elevation_high": s.ElevationHigh,
		"elevation_low":  s.ElevationLow,
		"climb_category": s.ClimbCategory,
		"starred":        s.Starred,
	}
	if s.StartLat != nil {
		m["start_lat"] = *s.StartLat
	}
	if s.StartLng != nil {
		m["start_lng"] = *s.StartLng
	}
	if s.EndLat != nil {
		m["end_lat"] = *s.EndLat
	}
	if s.EndLng != nil {
		m["end_lng"] = *s.EndLng
	}
	if s.Polyline != "" {
		m["polyline"] = s.Polyline
	}
	if s.AthleteEffortCount != nil {
		m["athlete_effort_count"] = *s.AthleteEffortCount
	}
	if s.AthletePRElapsedTime != nil {
		m["athlete_pr_elapsed_time"] = *s.AthletePRElapsedTime
	}
	if s.AthletePRDate != nil {
		m["athlete_pr_date"] = s.AthletePRDate.Format(time.RFC3339)
	}
	return m
}
