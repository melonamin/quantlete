// Package importer handles importing data from Strava.
//
// This file defines shared JSON types that match the Strava API response format.
// These types are used by both the server adapter (which converts from strava.* types)
// and the WASM adapter (which unmarshals directly from JSON).
//
// This consolidation eliminates duplicate type definitions and conversion logic
// between the two adapters.
package importer

import (
	"time"

	"github.com/melonamin/quantlete/internal/strava"
)

// ============================================================================
// Activity
// ============================================================================

// ActivityJSON represents a Strava activity in JSON-serializable form.
// Dates are strings to handle direct JSON unmarshaling in WASM mode.
type ActivityJSON struct {
	ID                   int64               `json:"id"`
	Name                 string              `json:"name"`
	SportType            string              `json:"sport_type"`
	Type                 string              `json:"type"` // Fallback if sport_type is empty
	StartDate            string              `json:"start_date"`
	StartDateLocal       string              `json:"start_date_local"`
	Timezone             string              `json:"timezone"`
	Distance             float64             `json:"distance"`
	MovingTime           int                 `json:"moving_time"`
	ElapsedTime          int                 `json:"elapsed_time"`
	TotalElevationGain   float64             `json:"total_elevation_gain"`
	AverageSpeed         float64             `json:"average_speed"`
	MaxSpeed             float64             `json:"max_speed"`
	AverageHeartrate     *float64            `json:"average_heartrate,omitempty"`
	MaxHeartrate         *float64            `json:"max_heartrate,omitempty"`
	AverageWatts         *float64            `json:"average_watts,omitempty"`
	MaxWatts             *float64            `json:"max_watts,omitempty"`
	WeightedAverageWatts *float64            `json:"weighted_average_watts,omitempty"`
	Kilojoules           *float64            `json:"kilojoules,omitempty"`
	AverageCadence       *float64            `json:"average_cadence,omitempty"`
	Calories             *float64            `json:"calories,omitempty"`
	GearID               string              `json:"gear_id,omitempty"`
	Commute              bool                `json:"commute"`
	Trainer              bool                `json:"trainer"`
	Private              bool                `json:"private"`
	WorkoutType          *int                `json:"workout_type,omitempty"`
	LocationCity         string              `json:"location_city,omitempty"`
	LocationState        string              `json:"location_state,omitempty"`
	LocationCountry      string              `json:"location_country,omitempty"`
	Map                  *MapJSON            `json:"map,omitempty"`
	StartLatlng          []float64           `json:"start_latlng,omitempty"`
	Description          string              `json:"description,omitempty"`
	DeviceName           string              `json:"device_name,omitempty"`
	KudosCount           int                 `json:"kudos_count"`
	PhotoCount           int                 `json:"total_photo_count"`
	SegmentEfforts       []SegmentEffortJSON `json:"segment_efforts,omitempty"`
	BestEfforts          []BestEffortJSON    `json:"best_efforts,omitempty"`
}

// MapJSON represents activity map data.
type MapJSON struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	SummaryPolyline string `json:"summary_polyline"`
}

// ToImporter converts ActivityJSON to importer.Activity.
func (a *ActivityJSON) ToImporter() Activity {
	sportType := a.SportType
	if sportType == "" {
		sportType = a.Type
	}

	result := Activity{
		ID:                   a.ID,
		Name:                 a.Name,
		SportType:            sportType,
		Timezone:             a.Timezone,
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
	result.SegmentEfforts = make([]SegmentEffort, len(a.SegmentEfforts))
	for i := range a.SegmentEfforts {
		result.SegmentEfforts[i] = a.SegmentEfforts[i].ToImporter()
	}

	// Convert best efforts
	result.BestEfforts = make([]BestEffort, len(a.BestEfforts))
	for i := range a.BestEfforts {
		result.BestEfforts[i] = a.BestEfforts[i].ToImporter()
	}

	return result
}

// ActivityJSONFromStrava converts a native strava.Activity to ActivityJSON.
func ActivityJSONFromStrava(a *strava.Activity) *ActivityJSON {
	result := &ActivityJSON{
		ID:                 a.ID,
		Name:               a.Name,
		SportType:          a.SportType,
		StartDate:          a.StartDate.Format(time.RFC3339),
		StartDateLocal:     a.StartDateLocal.Format(time.RFC3339),
		Timezone:           a.Timezone,
		Distance:           a.Distance,
		MovingTime:         a.MovingTime,
		ElapsedTime:        a.ElapsedTime,
		TotalElevationGain: a.TotalElevationGain,
		AverageSpeed:       a.AverageSpeed,
		MaxSpeed:           a.MaxSpeed,
		GearID:             a.GearID,
		Commute:            a.Commute,
		Trainer:            a.Trainer,
		Private:            a.Private,
		LocationCity:       a.LocationCity,
		LocationState:      a.LocationState,
		LocationCountry:    a.LocationCountry,
		Description:        a.Description,
		DeviceName:         a.DeviceName,
		KudosCount:         a.KudosCount,
		PhotoCount:         a.PhotoCount,
		StartLatlng:        a.StartLatlng,
	}

	// Handle map
	result.Map = &MapJSON{
		ID:              a.Map.ID,
		Polyline:        a.Map.Polyline,
		SummaryPolyline: a.Map.SummaryPolyline,
	}

	// Convert optional floats (0 means not set in strava types)
	if a.AverageHeartrate > 0 {
		v := a.AverageHeartrate
		result.AverageHeartrate = &v
	}
	if a.MaxHeartrate > 0 {
		v := a.MaxHeartrate
		result.MaxHeartrate = &v
	}
	if a.AverageWatts > 0 {
		v := a.AverageWatts
		result.AverageWatts = &v
	}
	if a.MaxWatts > 0 {
		v := a.MaxWatts
		result.MaxWatts = &v
	}
	if a.WeightedAverageWatts > 0 {
		v := a.WeightedAverageWatts
		result.WeightedAverageWatts = &v
	}
	if a.Kilojoules > 0 {
		v := a.Kilojoules
		result.Kilojoules = &v
	}
	if a.AverageCadence > 0 {
		v := a.AverageCadence
		result.AverageCadence = &v
	}
	if a.Calories > 0 {
		v := a.Calories
		result.Calories = &v
	}
	if a.WorkoutType > 0 {
		v := a.WorkoutType
		result.WorkoutType = &v
	}

	// Convert segment efforts
	result.SegmentEfforts = make([]SegmentEffortJSON, len(a.SegmentEfforts))
	for i := range a.SegmentEfforts {
		result.SegmentEfforts[i] = *SegmentEffortJSONFromStrava(&a.SegmentEfforts[i])
	}

	// Convert best efforts
	result.BestEfforts = make([]BestEffortJSON, len(a.BestEfforts))
	for i := range a.BestEfforts {
		result.BestEfforts[i] = *BestEffortJSONFromStrava(&a.BestEfforts[i])
	}

	return result
}

// ============================================================================
// Segment
// ============================================================================

// SegmentJSON represents a Strava segment in JSON-serializable form.
type SegmentJSON struct {
	ID            int64             `json:"id"`
	Name          string            `json:"name"`
	ActivityType  string            `json:"activity_type"`
	Distance      float64           `json:"distance"`
	AverageGrade  float64           `json:"average_grade"`
	MaximumGrade  float64           `json:"maximum_grade"`
	ElevationHigh float64           `json:"elevation_high"`
	ElevationLow  float64           `json:"elevation_low"`
	ClimbCategory int               `json:"climb_category"`
	StartLatlng   []float64         `json:"start_latlng,omitempty"`
	EndLatlng     []float64         `json:"end_latlng,omitempty"`
	Starred       bool              `json:"starred"`
	Map           *MapJSON          `json:"map,omitempty"`
	AthleteStats  *SegmentStatsJSON `json:"athlete_segment_stats,omitempty"`
}

// SegmentStatsJSON represents athlete-specific segment statistics.
type SegmentStatsJSON struct {
	PRElapsedTime int    `json:"pr_elapsed_time"`
	PRDate        string `json:"pr_date,omitempty"`
	EffortCount   int    `json:"effort_count"`
	KOMRank       *int   `json:"kom_rank,omitempty"`
}

// ToImporter converts SegmentJSON to importer.Segment.
func (s *SegmentJSON) ToImporter() *Segment {
	result := &Segment{
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

// SegmentJSONFromStrava converts a native strava.Segment to SegmentJSON.
func SegmentJSONFromStrava(s *strava.Segment) *SegmentJSON {
	result := &SegmentJSON{
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
		Map: &MapJSON{
			Polyline: s.Map.Polyline,
		},
		AthleteStats: &SegmentStatsJSON{
			PRElapsedTime: s.AthleteSegmentStats.PRElapsedTime,
			EffortCount:   s.AthleteSegmentStats.EffortCount,
			KOMRank:       s.AthleteSegmentStats.KOMRank,
		},
	}

	// Handle PR date
	if s.AthleteSegmentStats.PRDate != nil && !s.AthleteSegmentStats.PRDate.IsZero() {
		result.AthleteStats.PRDate = s.AthleteSegmentStats.PRDate.Format("2006-01-02")
	}

	return result
}

// ============================================================================
// Segment Effort
// ============================================================================

// SegmentEffortJSON represents a segment effort in JSON-serializable form.
type SegmentEffortJSON struct {
	ID               int64       `json:"id"`
	Segment          SegmentJSON `json:"segment"`
	Name             string      `json:"name"`
	ElapsedTime      int         `json:"elapsed_time"`
	MovingTime       int         `json:"moving_time"`
	StartDate        string      `json:"start_date"`
	StartDateLocal   string      `json:"start_date_local"`
	Distance         float64     `json:"distance"`
	AverageWatts     float64     `json:"average_watts"`
	AverageHeartrate float64     `json:"average_heartrate"`
	MaxHeartrate     float64     `json:"max_heartrate"`
	PRRank           *int        `json:"pr_rank,omitempty"`
}

// ToImporter converts SegmentEffortJSON to importer.SegmentEffort.
func (e *SegmentEffortJSON) ToImporter() SegmentEffort {
	result := SegmentEffort{
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

	result.Segment = *e.Segment.ToImporter()

	if t, err := time.Parse(time.RFC3339, e.StartDate); err == nil {
		result.StartDate = t
	}
	if t, err := time.Parse(time.RFC3339, e.StartDateLocal); err == nil {
		result.StartDateLocal = t
	}

	return result
}

// SegmentEffortJSONFromStrava converts a native strava.SegmentEffort to SegmentEffortJSON.
func SegmentEffortJSONFromStrava(e *strava.SegmentEffort) *SegmentEffortJSON {
	return &SegmentEffortJSON{
		ID:               e.ID,
		Segment:          *SegmentJSONFromStrava(&e.Segment),
		Name:             e.Name,
		ElapsedTime:      e.ElapsedTime,
		MovingTime:       e.MovingTime,
		StartDate:        e.StartDate.Format(time.RFC3339),
		StartDateLocal:   e.StartDateLocal.Format(time.RFC3339),
		Distance:         e.Distance,
		AverageWatts:     e.AverageWatts,
		AverageHeartrate: e.AverageHeartrate,
		MaxHeartrate:     e.MaxHeartrate,
		PRRank:           e.PRRank,
	}
}

// ============================================================================
// Best Effort
// ============================================================================

// BestEffortJSON represents a best effort in JSON-serializable form.
type BestEffortJSON struct {
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

// ToImporter converts BestEffortJSON to importer.BestEffort.
func (e *BestEffortJSON) ToImporter() BestEffort {
	result := BestEffort{
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

// BestEffortJSONFromStrava converts a native strava.BestEffort to BestEffortJSON.
func BestEffortJSONFromStrava(e *strava.BestEffort) *BestEffortJSON {
	return &BestEffortJSON{
		ID:          int64(e.ElapsedTime), // No separate ID in strava.BestEffort, use elapsed time
		Name:        e.Name,
		ElapsedTime: e.ElapsedTime,
		MovingTime:  e.MovingTime,
		StartDate:   e.StartDate.Format(time.RFC3339),
		Distance:    e.Distance,
		PRRank:      e.PRRank,
		StartIndex:  e.StartIndex,
		EndIndex:    e.EndIndex,
	}
}

// ============================================================================
// Gear
// ============================================================================

// GearJSON represents gear in JSON-serializable form.
type GearJSON struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Primary     bool    `json:"primary"`
	Retired     bool    `json:"retired"`
	Distance    float64 `json:"distance"`
	BrandName   string  `json:"brand_name"`
	ModelName   string  `json:"model_name"`
	Description string  `json:"description"`
}

// ToImporter converts GearJSON to importer.Gear.
func (g *GearJSON) ToImporter() *Gear {
	return &Gear{
		ID:          g.ID,
		Name:        g.Name,
		Primary:     g.Primary,
		Retired:     g.Retired,
		Distance:    g.Distance,
		BrandName:   g.BrandName,
		ModelName:   g.ModelName,
		Description: g.Description,
	}
}

// GearJSONFromStrava converts a native strava.Gear to GearJSON.
func GearJSONFromStrava(g *strava.Gear) *GearJSON {
	return &GearJSON{
		ID:          g.ID,
		Name:        g.Name,
		Primary:     g.Primary,
		Retired:     g.Retired,
		Distance:    g.Distance,
		BrandName:   g.BrandName,
		ModelName:   g.ModelName,
		Description: g.Description,
	}
}

// ============================================================================
// Photo
// ============================================================================

// PhotoJSON represents a photo in JSON-serializable form.
type PhotoJSON struct {
	UniqueID   string            `json:"unique_id"`
	ActivityID int64             `json:"activity_id"`
	URLs       map[string]string `json:"urls"`
	Caption    string            `json:"caption"`
	Location   []float64         `json:"location"`
	CreatedAt  string            `json:"created_at"`
}

// ToImporter converts PhotoJSON to importer.Photo.
func (p *PhotoJSON) ToImporter(activityID int64) Photo {
	result := Photo{
		UniqueID:   p.UniqueID,
		ActivityID: activityID,
		URLs:       p.URLs,
		Caption:    p.Caption,
		Location:   p.Location,
	}

	if p.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, p.CreatedAt); err == nil {
			result.CreatedAt = t
		}
	}

	return result
}

// ============================================================================
// Streams
// ============================================================================

// StreamSetJSON represents a collection of activity streams.
type StreamSetJSON struct {
	Time           *StreamJSON `json:"time,omitempty"`
	Distance       *StreamJSON `json:"distance,omitempty"`
	Altitude       *StreamJSON `json:"altitude,omitempty"`
	Heartrate      *StreamJSON `json:"heartrate,omitempty"`
	Watts          *StreamJSON `json:"watts,omitempty"`
	Cadence        *StreamJSON `json:"cadence,omitempty"`
	VelocitySmooth *StreamJSON `json:"velocity_smooth,omitempty"`
	Latlng         *StreamJSON `json:"latlng,omitempty"`
}

// StreamJSON represents a single stream type.
type StreamJSON struct {
	Type         string      `json:"type"`
	Data         interface{} `json:"data"`
	OriginalSize int         `json:"original_size"`
	Resolution   string      `json:"resolution"`
	SeriesType   string      `json:"series_type"`
}

// ToImporter converts StreamJSON to importer.Stream.
func (s *StreamJSON) ToImporter() *Stream {
	return &Stream{
		Type:         s.Type,
		Data:         s.Data,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
	}
}

// ToImporter converts StreamSetJSON to importer.StreamSet.
func (ss *StreamSetJSON) ToImporter() *StreamSet {
	result := &StreamSet{}

	if ss.Time != nil {
		result.Time = ss.Time.ToImporter()
	}
	if ss.Distance != nil {
		result.Distance = ss.Distance.ToImporter()
	}
	if ss.Altitude != nil {
		result.Altitude = ss.Altitude.ToImporter()
	}
	if ss.Heartrate != nil {
		result.Heartrate = ss.Heartrate.ToImporter()
	}
	if ss.Watts != nil {
		result.Watts = ss.Watts.ToImporter()
	}
	if ss.Cadence != nil {
		result.Cadence = ss.Cadence.ToImporter()
	}
	if ss.VelocitySmooth != nil {
		result.VelocitySmooth = ss.VelocitySmooth.ToImporter()
	}
	if ss.Latlng != nil {
		result.Latlng = ss.Latlng.ToImporter()
	}

	return result
}

// StreamSetJSONFromStrava converts a native strava.StreamSet to StreamSetJSON.
func StreamSetJSONFromStrava(s *strava.StreamSet) *StreamSetJSON {
	result := &StreamSetJSON{}

	if s.Time != nil {
		result.Time = StreamJSONFromStrava(s.Time)
	}
	if s.Distance != nil {
		result.Distance = StreamJSONFromStrava(s.Distance)
	}
	if s.Altitude != nil {
		result.Altitude = StreamJSONFromStrava(s.Altitude)
	}
	if s.Heartrate != nil {
		result.Heartrate = StreamJSONFromStrava(s.Heartrate)
	}
	if s.Watts != nil {
		result.Watts = StreamJSONFromStrava(s.Watts)
	}
	if s.Cadence != nil {
		result.Cadence = StreamJSONFromStrava(s.Cadence)
	}
	if s.VelocitySmooth != nil {
		result.VelocitySmooth = StreamJSONFromStrava(s.VelocitySmooth)
	}
	if s.Latlng != nil {
		result.Latlng = StreamJSONFromStrava(s.Latlng)
	}

	return result
}

// StreamJSONFromStrava converts a native strava.Stream to StreamJSON.
func StreamJSONFromStrava(s *strava.Stream) *StreamJSON {
	return &StreamJSON{
		Type:         s.Type,
		Data:         s.Data,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
	}
}
