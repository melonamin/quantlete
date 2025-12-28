package strava

import "time"

// Athlete represents a Strava athlete.
type Athlete struct {
	ID            int64   `json:"id"`
	Username      string  `json:"username"`
	FirstName     string  `json:"firstname"`
	LastName      string  `json:"lastname"`
	City          string  `json:"city"`
	State         string  `json:"state"`
	Country       string  `json:"country"`
	Sex           string  `json:"sex"`
	Premium       bool    `json:"premium"`
	Summit        bool    `json:"summit"`
	ProfileMedium string  `json:"profile_medium"`
	Profile       string  `json:"profile"`
	Weight        float64 `json:"weight"` // kg
}

// Activity represents a Strava activity.
type Activity struct {
	ID                   int64           `json:"id"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	SportType            string          `json:"sport_type"`
	StartDate            time.Time       `json:"start_date"`
	StartDateLocal       time.Time       `json:"start_date_local"`
	Timezone             string          `json:"timezone"`
	Distance             float64         `json:"distance"`             // meters
	MovingTime           int             `json:"moving_time"`          // seconds
	ElapsedTime          int             `json:"elapsed_time"`         // seconds
	TotalElevationGain   float64         `json:"total_elevation_gain"` // meters
	ElevHigh             float64         `json:"elev_high"`
	ElevLow              float64         `json:"elev_low"`
	AverageSpeed         float64         `json:"average_speed"` // m/s
	MaxSpeed             float64         `json:"max_speed"`     // m/s
	AverageHeartrate     float64         `json:"average_heartrate"`
	MaxHeartrate         int             `json:"max_heartrate"`
	AverageWatts         float64         `json:"average_watts"`
	MaxWatts             int             `json:"max_watts"`
	WeightedAverageWatts int             `json:"weighted_average_watts"`
	Kilojoules           float64         `json:"kilojoules"`
	AverageCadence       float64         `json:"average_cadence"`
	Calories             float64         `json:"calories"`
	KudosCount           int             `json:"kudos_count"`
	CommentCount         int             `json:"comment_count"`
	PhotoCount           int             `json:"total_photo_count"`
	Commute              bool            `json:"commute"`
	Private              bool            `json:"private"`
	Trainer              bool            `json:"trainer"`
	WorkoutType          int             `json:"workout_type"`
	DeviceName           string          `json:"device_name"`
	GearID               string          `json:"gear_id"`
	StartLatlng          []float64       `json:"start_latlng"`
	EndLatlng            []float64       `json:"end_latlng"`
	Map                  ActivityMap     `json:"map"`
	SegmentEfforts       []SegmentEffort `json:"segment_efforts,omitempty"`
	SplitsMetric         []Split         `json:"splits_metric,omitempty"`
	Laps                 []Lap           `json:"laps,omitempty"`
}

// ActivityMap contains map data for an activity.
type ActivityMap struct {
	ID              string `json:"id"`
	Polyline        string `json:"polyline"`
	SummaryPolyline string `json:"summary_polyline"`
}

// SegmentEffort represents an effort on a segment.
type SegmentEffort struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	ElapsedTime      int       `json:"elapsed_time"`
	MovingTime       int       `json:"moving_time"`
	StartDate        time.Time `json:"start_date"`
	StartDateLocal   time.Time `json:"start_date_local"`
	Distance         float64   `json:"distance"`
	AverageWatts     float64   `json:"average_watts"`
	AverageHeartrate float64   `json:"average_heartrate"`
	MaxHeartrate     int       `json:"max_heartrate"`
	PRRank           *int      `json:"pr_rank"`
	Segment          Segment   `json:"segment"`
}

// Segment represents a Strava segment.
type Segment struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	ActivityType  string    `json:"activity_type"`
	Distance      float64   `json:"distance"`
	AverageGrade  float64   `json:"average_grade"`
	MaximumGrade  float64   `json:"maximum_grade"`
	ElevationHigh float64   `json:"elevation_high"`
	ElevationLow  float64   `json:"elevation_low"`
	ClimbCategory int       `json:"climb_category"`
	StartLatlng   []float64 `json:"start_latlng"`
	EndLatlng     []float64 `json:"end_latlng"`
	Starred       bool      `json:"starred"`
}

// Split represents a split in an activity.
type Split struct {
	Distance         float64 `json:"distance"`
	ElapsedTime      int     `json:"elapsed_time"`
	MovingTime       int     `json:"moving_time"`
	AverageSpeed     float64 `json:"average_speed"`
	AverageHeartrate float64 `json:"average_heartrate"`
	PaceZone         int     `json:"pace_zone"`
	Split            int     `json:"split"`
}

// Lap represents a lap in an activity.
type Lap struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	ElapsedTime      int     `json:"elapsed_time"`
	MovingTime       int     `json:"moving_time"`
	Distance         float64 `json:"distance"`
	AverageSpeed     float64 `json:"average_speed"`
	MaxSpeed         float64 `json:"max_speed"`
	AverageWatts     float64 `json:"average_watts"`
	AverageHeartrate float64 `json:"average_heartrate"`
	MaxHeartrate     int     `json:"max_heartrate"`
	AverageCadence   float64 `json:"average_cadence"`
	LapIndex         int     `json:"lap_index"`
}

// Gear represents a piece of equipment.
type Gear struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Primary     bool    `json:"primary"`
	Retired     bool    `json:"retired"`
	Distance    float64 `json:"distance"` // meters
	BrandName   string  `json:"brand_name"`
	ModelName   string  `json:"model_name"`
	Description string  `json:"description"`
}

// StreamSet represents a collection of activity streams.
type StreamSet struct {
	Time           *Stream `json:"time,omitempty"`
	Distance       *Stream `json:"distance,omitempty"`
	Latlng         *Stream `json:"latlng,omitempty"`
	Altitude       *Stream `json:"altitude,omitempty"`
	VelocitySmooth *Stream `json:"velocity_smooth,omitempty"`
	Heartrate      *Stream `json:"heartrate,omitempty"`
	Cadence        *Stream `json:"cadence,omitempty"`
	Watts          *Stream `json:"watts,omitempty"`
	Temp           *Stream `json:"temp,omitempty"`
	Moving         *Stream `json:"moving,omitempty"`
	GradeSmooth    *Stream `json:"grade_smooth,omitempty"`
}

// Stream represents a single data stream.
type Stream struct {
	Type         string `json:"type"`
	SeriesType   string `json:"series_type"`
	OriginalSize int    `json:"original_size"`
	Resolution   string `json:"resolution"`
	Data         []any  `json:"data"`
}
