package services

import (
	"context"
	"fmt"
	"time"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
)

// StatsService handles stats business logic.
type StatsService struct {
	stats            *storage.StatsRepository
	power            *storage.PowerRepository
	bestEfforts      *storage.BestEffortsRepository
	trainingLoad     *storage.TrainingLoadRepository
	zoneDistribution *storage.ZoneDistributionRepository
}

// NewStatsService creates a new stats service.
func NewStatsService(
	stats *storage.StatsRepository,
	power *storage.PowerRepository,
	bestEfforts *storage.BestEffortsRepository,
	trainingLoad *storage.TrainingLoadRepository,
	zoneDistribution *storage.ZoneDistributionRepository,
) *StatsService {
	return &StatsService{
		stats:            stats,
		power:            power,
		bestEfforts:      bestEfforts,
		trainingLoad:     trainingLoad,
		zoneDistribution: zoneDistribution,
	}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// --- Heatmap ---

// GetHeatmapDataInput contains parameters for getting heatmap data.
type GetHeatmapDataInput struct {
	AthleteID   int64      `json:"-" adapter:"context"`
	SportTypes  []string   `json:"sport_types" adapter:"query,name=sport_type,split=,"`
	StartAfter  *time.Time `json:"after" adapter:"query"`
	StartBefore *time.Time `json:"before" adapter:"query"`
	Commute     *bool      `json:"commute" adapter:"query"`
	WorkoutType *int       `json:"workout_type" adapter:"query"`
	Limit       int        `json:"limit" adapter:"query"`
	Offset      int        `json:"offset" adapter:"query"`
}

// HeatmapActivity represents an activity for the heatmap visualization.
type HeatmapActivity struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	SportType       string  `json:"sport_type"`
	StartDate       string  `json:"start_date"`
	Distance        float64 `json:"distance"`
	SummaryPolyline string  `json:"summary_polyline"`
	StartLat        float64 `json:"start_lat"`
	StartLng        float64 `json:"start_lng"`
}

// HeatmapCountryStat represents country statistics for heatmap.
type HeatmapCountryStat struct {
	Country string `json:"country"`
	ISO2    string `json:"iso2,omitempty"`
	Count   int    `json:"count"`
}

// HeatmapOutput contains the heatmap data and summary statistics.
type HeatmapOutput struct {
	Activities []HeatmapActivity    `json:"activities"`
	Total      int                  `json:"total"`
	Limit      int                  `json:"limit,omitempty"`
	Offset     int                  `json:"offset,omitempty"`
	Countries  []HeatmapCountryStat `json:"countries,omitempty"`
}

// --- Best Efforts ---

// GetBestEffortPRsInput contains parameters for getting best effort PRs.
type GetBestEffortPRsInput struct {
	AthleteID  int64    `json:"-" adapter:"context"`
	SportTypes []string `json:"sport_types" adapter:"query,name=sport_type,split=,"`
}

// BestEffortPR represents a personal record for a distance type.
type BestEffortPR struct {
	DistanceType   string  `json:"distance_type"`
	Name           string  `json:"name"`
	DistanceM      float64 `json:"distance_m"`
	ElapsedTimeS   int     `json:"elapsed_time_s"`
	MovingTimeS    *int    `json:"moving_time_s,omitempty"`
	PRRank         *int    `json:"pr_rank,omitempty"`
	ActivityID     int64   `json:"activity_id"`
	ActivityName   string  `json:"activity_name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
}

// GetBestEffortsForTypeInput contains parameters for getting best efforts by distance type.
type GetBestEffortsForTypeInput struct {
	AthleteID    int64    `json:"-" adapter:"context"`
	DistanceType string   `json:"distance_type" adapter:"path,param=distanceType"`
	SportTypes   []string `json:"sport_types" adapter:"query,name=sport_type,split=,"`
}

// BestEffortListItem represents a best effort item in a list.
type BestEffortListItem struct {
	DistanceType   string  `json:"distance_type"`
	Name           string  `json:"name"`
	DistanceM      float64 `json:"distance_m"`
	ElapsedTimeS   int     `json:"elapsed_time_s"`
	MovingTimeS    *int    `json:"moving_time_s,omitempty"`
	PRRank         *int    `json:"pr_rank,omitempty"`
	StartIndex     *int    `json:"start_index,omitempty"`
	EndIndex       *int    `json:"end_index,omitempty"`
	ActivityID     int64   `json:"activity_id"`
	ActivityName   string  `json:"activity_name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
}

// --- Eddington ---

// GetEddingtonDataInput contains parameters for getting Eddington data.
type GetEddingtonDataInput struct {
	AthleteID  int64    `json:"-" adapter:"context"`
	SportTypes []string `json:"sport_types" adapter:"query,name=sport_type,split=,"`
}

// EddingtonDay represents a day's distance for Eddington calculation.
type EddingtonDay struct {
	Date     string  `json:"date"`
	Distance float64 `json:"distance"`
}

// EddingtonStep shows how many rides needed to reach the next Eddington number.
type EddingtonStep struct {
	Target      int `json:"target"`
	RidesNeeded int `json:"rides_needed"`
}

// EddingtonOutput contains the Eddington number calculation result.
type EddingtonOutput struct {
	Number       int             `json:"number"`
	Distribution []EddingtonDay  `json:"distribution"`
	NextSteps    []EddingtonStep `json:"next_steps"`
}

// GetEddingtonHistoryInput contains parameters for getting Eddington history.
type GetEddingtonHistoryInput struct {
	AthleteID  int64    `json:"-" adapter:"context"`
	SportTypes []string `json:"sport_types" adapter:"query,name=sport_type,split=,"`
}

// EddingtonHistoryPoint represents a milestone point where the Eddington number increases.
type EddingtonHistoryPoint struct {
	Date   string `json:"date"`
	Number int    `json:"number"`
}

// --- Power Stats ---

// GetPowerStatsInput contains parameters for getting power stats.
type GetPowerStatsInput struct {
	AthleteID  int64      `json:"-" adapter:"context"`
	After      *time.Time `json:"after" adapter:"query"`
	Before     *time.Time `json:"before" adapter:"query"`
	SportTypes []string   `json:"sport_types" adapter:"query,name=sport_type,split=,"`
}

// PeakPowerBest represents the best power for a duration.
type PeakPowerBest struct {
	DurationS  int     `json:"duration_s"`
	Watts      float64 `json:"watts"`
	ActivityID int64   `json:"activity_id"`
	StartDate  string  `json:"start_date"`
}

// PeakPowerHistoryPoint represents a point in power history.
type PeakPowerHistoryPoint struct {
	Date  string  `json:"date"`
	Watts float64 `json:"watts"`
}

// PowerStatsOutput contains power statistics.
type PowerStatsOutput struct {
	DurationsS []int                           `json:"durations_s"`
	Best       []PeakPowerBest                 `json:"best"`
	History    map[int][]PeakPowerHistoryPoint `json:"history"`
}

// --- Training Load ---

// GetTrainingLoadInput contains parameters for getting training load data.
type GetTrainingLoadInput struct {
	AthleteID int64      `json:"-" adapter:"context"`
	After     *time.Time `json:"after" adapter:"query"`
	Before    *time.Time `json:"before" adapter:"query"`
}

// DailyTrainingLoadPoint represents training load for a single day.
type DailyTrainingLoadPoint struct {
	Day string  `json:"day"`
	TSS float64 `json:"tss"`
	CTL float64 `json:"ctl"`
	ATL float64 `json:"atl"`
	TSB float64 `json:"tsb"`
}

// ComputationStats provides diagnostic information about TSS computation.
type ComputationStats struct {
	TotalActivities      int  `json:"total_activities"`
	ActivitiesWithPower  int  `json:"activities_with_power"`
	ActivitiesWithTSS    int  `json:"activities_with_tss"`
	CyclingFTPConfigured bool `json:"cycling_ftp_configured"`
	RunningFTPConfigured bool `json:"running_ftp_configured"`
	HRZonesConfigured    bool `json:"hr_zones_configured"`
}

// WeeklyMetrics contains weekly training metrics.
type WeeklyMetrics struct {
	RestDays     int     `json:"rest_days"`      // Days with no activity (0-7)
	Monotony     float64 `json:"monotony"`       // Mean TSS / StdDev TSS
	WeeklyStrain float64 `json:"weekly_strain"`  // Total TSS last 7 days
	WeeklyTRIMP  float64 `json:"weekly_trimp"`   // Total TRIMP last 7 days
}

// PolarizedBreakdown shows training intensity distribution.
type PolarizedBreakdown struct {
	LowPercent      float64 `json:"low_percent"`      // Z1+Z2 %
	ModeratePercent float64 `json:"moderate_percent"` // Z3 %
	HighPercent     float64 `json:"high_percent"`     // Z4+Z5 %
	TotalSeconds    int     `json:"total_seconds"`
	PeriodDays      int     `json:"period_days"` // 30
}

// TrainingLoadOutput contains training load data.
type TrainingLoadOutput struct {
	Series        []DailyTrainingLoadPoint `json:"series"`
	Summary       *DailyTrainingLoadPoint  `json:"summary,omitempty"`
	Stats         *ComputationStats        `json:"stats,omitempty"`
	WeeklyMetrics *WeeklyMetrics           `json:"weekly_metrics,omitempty"`
	Polarized     *PolarizedBreakdown      `json:"polarized,omitempty"`
}

// --- Zone Trend ---

// GetZoneTrendInput contains parameters for getting zone trend data.
type GetZoneTrendInput struct {
	AthleteID int64 `json:"-" adapter:"context"`
	Weeks     int   `json:"weeks" adapter:"query"` // Default: 52, max: 104
}

const (
	defaultZoneTrendWeeks = 52
	maxZoneTrendWeeks     = 104 // ~2 years
)

// WeeklyZoneDistribution represents zone data for a single week.
type WeeklyZoneDistribution struct {
	Week      string  `json:"week"` // "YYYY-WNN" format
	SecondsZ1 int     `json:"seconds_z1"`
	SecondsZ2 int     `json:"seconds_z2"`
	SecondsZ3 int     `json:"seconds_z3"`
	SecondsZ4 int     `json:"seconds_z4"`
	SecondsZ5 int     `json:"seconds_z5"`
	Total     int     `json:"total"`
	PercentZ1 float64 `json:"percent_z1"`
	PercentZ2 float64 `json:"percent_z2"`
	PercentZ3 float64 `json:"percent_z3"`
	PercentZ4 float64 `json:"percent_z4"`
	PercentZ5 float64 `json:"percent_z5"`
}

// ZoneTrendOutput contains zone trend data.
type ZoneTrendOutput struct {
	Weeks []WeeklyZoneDistribution `json:"weeks"`
}

// --- Wrapped ---

// GetWrappedYearsInput contains parameters for getting available wrapped years.
type GetWrappedYearsInput struct {
	AthleteID int64 `json:"-" adapter:"context"`
}

// GetWrappedInput contains parameters for getting wrapped data.
type GetWrappedInput struct {
	AthleteID int64 `json:"-" adapter:"context"`
	Year      int   `json:"year" adapter:"query"` // 0 = all-time
}

// WrappedTotals represents totals for the wrapped report.
type WrappedTotals struct {
	Activities    int     `json:"activities"`
	DistanceM     float64 `json:"distance_m"`
	ElevationM    float64 `json:"elevation_m"`
	MovingTimeS   int     `json:"moving_time_s"`
	Kudos         int     `json:"kudos"`
	CommuteDistM  float64 `json:"commute_distance_m"`
	CarbonSavedKg float64 `json:"carbon_saved_kg"`
}

// WrappedMonth represents monthly statistics for wrapped.
type WrappedMonth struct {
	Month      string  `json:"month"`
	Activities int     `json:"activities"`
	DistanceM  float64 `json:"distance_m"`
	ElevationM float64 `json:"elevation_m"`
	PRs        int     `json:"prs"`
}

// WrappedSportTime represents moving time by sport type.
type WrappedSportTime struct {
	SportType   string `json:"sport_type"`
	MovingTimeS int    `json:"moving_time_s"`
}

// WrappedHourCount represents activity count by hour.
type WrappedHourCount struct {
	Hour  int `json:"hour"`
	Count int `json:"count"`
}

// WrappedLocationPoint represents a location bucket.
type WrappedLocationPoint struct {
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Count int     `json:"count"`
}

// WrappedBiggestActivity represents the biggest activity for a metric.
type WrappedBiggestActivity struct {
	ActivityID     int64   `json:"activity_id"`
	Name           string  `json:"name"`
	SportType      string  `json:"sport_type"`
	StartDateLocal string  `json:"start_date_local"`
	Value          float64 `json:"value"`
}

// WrappedBiggest contains the biggest activities.
type WrappedBiggest struct {
	LongestDistance *WrappedBiggestActivity `json:"longest_distance,omitempty"`
	MostElevation   *WrappedBiggestActivity `json:"most_elevation,omitempty"`
	LongestDuration *WrappedBiggestActivity `json:"longest_duration,omitempty"`
}

// WrappedStreaks represents streak data.
type WrappedStreaks struct {
	LongestActiveDays int `json:"longest_active_days"`
	LongestRestDays   int `json:"longest_rest_days"`
}

// WrappedPhoto represents a photo in the wrapped.
type WrappedPhoto struct {
	ID           string `json:"id"`
	ActivityID   int64  `json:"activity_id"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Caption      string `json:"caption,omitempty"`
}

// WrappedOutput contains the wrapped report.
type WrappedOutput struct {
	Year              int                    `json:"year"`
	RangeStart        string                 `json:"range_start"`
	RangeEnd          string                 `json:"range_end"`
	TotalDays         int                    `json:"total_days"`
	ActiveDays        int                    `json:"active_days"`
	RestDays          int                    `json:"rest_days"`
	Totals            WrappedTotals          `json:"totals"`
	Months            []WrappedMonth         `json:"months,omitempty"`
	MovingTimeBySport []WrappedSportTime     `json:"moving_time_by_sport"`
	StartTimesByHour  []WrappedHourCount     `json:"start_times_by_hour"`
	Locations         []WrappedLocationPoint `json:"locations"`
	Streaks           WrappedStreaks         `json:"streaks"`
	RandomPhoto       *WrappedPhoto          `json:"random_photo,omitempty"`
	Biggest           WrappedBiggest         `json:"biggest"`
}

// --- Best Efforts Write ---

// SaveBestEffortsInput contains parameters for saving best efforts for an activity.
type SaveBestEffortsInput struct {
	AthleteID  int64                      `json:"athlete_id" adapter:"body"`
	ActivityID int64                      `json:"activity_id" adapter:"body"`
	SportType  string                     `json:"sport_type" adapter:"body"`
	Efforts    []SaveBestEffortsInputItem `json:"efforts" adapter:"body"`
}

// SaveBestEffortsInputItem represents a single best effort to save.
type SaveBestEffortsInputItem struct {
	DistanceType string  `json:"distance_type" adapter:"body"`
	Name         string  `json:"name" adapter:"body"`
	DistanceM    float64 `json:"distance_m" adapter:"body"`
	ElapsedTime  int     `json:"elapsed_time" adapter:"body"`
	MovingTime   *int    `json:"moving_time" adapter:"body"`
	StartIndex   *int    `json:"start_index" adapter:"body"`
	EndIndex     *int    `json:"end_index" adapter:"body"`
	PRRank       *int    `json:"pr_rank" adapter:"body"`
	StartDate    string  `json:"start_date" adapter:"body"`
}

// SaveBestEffortsOutput contains the result of saving best efforts.
type SaveBestEffortsOutput struct {
	Message string `json:"message"`
}

// Maximum number of best efforts that can be saved per activity.
const maxBestEffortsPerSave = 1000

// ============================================================================
// Service Methods
// ============================================================================

// Standard durations for power curve analysis.
var powerDurations = []int{5, 10, 30, 60, 300, 480, 1200, 3600}

// GetHeatmapData returns activities with polylines for heatmap visualization.
//
//adapter:wasm getHeatmapData category=Stats
//adapter:http GET /api/v1/stats/heatmap
func (s *StatsService) GetHeatmapData(ctx context.Context, in GetHeatmapDataInput) (*HeatmapOutput, error) {
	filters := storage.HeatmapFilters{
		SportTypes:  in.SportTypes,
		StartAfter:  in.StartAfter,
		StartBefore: in.StartBefore,
		Commute:     in.Commute,
		WorkoutType: in.WorkoutType,
		Limit:       in.Limit,
		Offset:      in.Offset,
	}

	total, err := s.stats.CountHeatmapActivities(ctx, in.AthleteID, filters)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to count heatmap data: %v", err)
	}

	activities, err := s.stats.GetHeatmapData(ctx, in.AthleteID, filters)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get heatmap data: %v", err)
	}

	result := make([]HeatmapActivity, len(activities))
	for i, a := range activities {
		result[i] = HeatmapActivity{
			ID:              a.ID,
			Name:            a.Name,
			SportType:       a.SportType,
			StartDate:       a.StartDate,
			Distance:        a.Distance,
			SummaryPolyline: a.SummaryPolyline,
			StartLat:        a.StartLat,
			StartLng:        a.StartLng,
		}
	}

	countries, _ := s.stats.GetHeatmapCountries(ctx, in.AthleteID, filters)
	countriesOut := make([]HeatmapCountryStat, len(countries))
	for i, c := range countries {
		countriesOut[i] = HeatmapCountryStat{
			Country: c.Country,
			ISO2:    c.ISO2,
			Count:   c.Count,
		}
	}

	return &HeatmapOutput{
		Activities: result,
		Total:      total,
		Limit:      in.Limit,
		Offset:     in.Offset,
		Countries:  countriesOut,
	}, nil
}

// GetBestEffortPRs returns one all-time PR (fastest elapsed time) per distance_type.
//
//adapter:wasm getBestEffortPRs category=Stats
//adapter:http GET /api/v1/stats/best-efforts
func (s *StatsService) GetBestEffortPRs(ctx context.Context, in GetBestEffortPRsInput) ([]BestEffortPR, error) {
	if s.bestEfforts == nil {
		return []BestEffortPR{}, nil
	}

	prs, err := s.bestEfforts.ListPRs(ctx, in.AthleteID, in.SportTypes)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to load best efforts: %v", err)
	}

	result := make([]BestEffortPR, len(prs))
	for i, p := range prs {
		result[i] = BestEffortPR{
			DistanceType:   p.DistanceType,
			Name:           p.Name,
			DistanceM:      p.DistanceM,
			ElapsedTimeS:   p.ElapsedTimeS,
			MovingTimeS:    p.MovingTimeS,
			PRRank:         p.PRRank,
			ActivityID:     p.ActivityID,
			ActivityName:   p.ActivityName,
			SportType:      p.SportType,
			StartDateLocal: p.StartDateLocal.Format(time.RFC3339),
		}
	}
	return result, nil
}

// GetBestEffortsForType returns all best efforts for a specific distance type.
//
//adapter:wasm getBestEffortsForType category=Stats
//adapter:http GET /api/v1/stats/best-efforts/{distanceType}
func (s *StatsService) GetBestEffortsForType(ctx context.Context, in GetBestEffortsForTypeInput) ([]BestEffortListItem, error) {
	if s.bestEfforts == nil {
		return []BestEffortListItem{}, nil
	}

	if in.DistanceType == "" {
		return nil, BadRequest("distanceType is required")
	}

	items, err := s.bestEfforts.ListByDistanceType(ctx, in.AthleteID, in.DistanceType, in.SportTypes)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to load best efforts: %v", err)
	}

	result := make([]BestEffortListItem, len(items))
	for i, it := range items {
		result[i] = BestEffortListItem{
			DistanceType:   it.DistanceType,
			Name:           it.Name,
			DistanceM:      it.DistanceM,
			ElapsedTimeS:   it.ElapsedTimeS,
			MovingTimeS:    it.MovingTimeS,
			PRRank:         it.PRRank,
			StartIndex:     it.StartIndex,
			EndIndex:       it.EndIndex,
			ActivityID:     it.ActivityID,
			ActivityName:   it.ActivityName,
			SportType:      it.SportType,
			StartDateLocal: it.StartDateLocal.Format(time.RFC3339),
		}
	}
	return result, nil
}

// GetEddingtonData returns data for Eddington number calculation.
//
//adapter:wasm getEddingtonData category=Stats
//adapter:http GET /api/v1/stats/eddington
func (s *StatsService) GetEddingtonData(ctx context.Context, in GetEddingtonDataInput) (*EddingtonOutput, error) {
	result, err := s.stats.GetEddingtonData(ctx, in.AthleteID, in.SportTypes)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get eddington data: %v", err)
	}

	distribution := make([]EddingtonDay, len(result.Distribution))
	for i, d := range result.Distribution {
		distribution[i] = EddingtonDay{
			Date:     d.Date,
			Distance: d.Distance,
		}
	}

	nextSteps := make([]EddingtonStep, len(result.NextSteps))
	for i, s := range result.NextSteps {
		nextSteps[i] = EddingtonStep{
			Target:      s.Target,
			RidesNeeded: s.RidesNeeded,
		}
	}

	return &EddingtonOutput{
		Number:       result.Number,
		Distribution: distribution,
		NextSteps:    nextSteps,
	}, nil
}

// GetEddingtonHistory returns milestone points where the Eddington number increases.
//
//adapter:wasm getEddingtonHistory category=Stats
//adapter:http GET /api/v1/stats/eddington/history
func (s *StatsService) GetEddingtonHistory(ctx context.Context, in GetEddingtonHistoryInput) ([]EddingtonHistoryPoint, error) {
	points, err := s.stats.GetEddingtonHistory(ctx, in.AthleteID, in.SportTypes)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get eddington history: %v", err)
	}

	if points == nil {
		return []EddingtonHistoryPoint{}, nil
	}

	result := make([]EddingtonHistoryPoint, len(points))
	for i, p := range points {
		result[i] = EddingtonHistoryPoint{
			Date:   p.Date,
			Number: p.Number,
		}
	}
	return result, nil
}

// GetPowerStats returns power best efforts and history.
//
//adapter:wasm getPowerStats category=Stats
//adapter:http GET /api/v1/stats/power
func (s *StatsService) GetPowerStats(ctx context.Context, in GetPowerStatsInput) (*PowerStatsOutput, error) {
	if s.power == nil {
		return &PowerStatsOutput{
			DurationsS: powerDurations,
			Best:       []PeakPowerBest{},
			History:    make(map[int][]PeakPowerHistoryPoint),
		}, nil
	}

	if err := s.power.EnsureComputedForRange(ctx, in.AthleteID, in.After, in.Before, in.SportTypes, powerDurations); err != nil {
		return nil, Wrapf(ErrInternal, "failed to compute power stats: %v", err)
	}

	best, err := s.power.GetBest(ctx, in.AthleteID, powerDurations, in.After, in.Before, in.SportTypes)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to load power stats: %v", err)
	}

	bestOut := make([]PeakPowerBest, len(best))
	for i, b := range best {
		bestOut[i] = PeakPowerBest{
			DurationS:  b.DurationS,
			Watts:      b.Watts,
			ActivityID: b.ActivityID,
			StartDate:  b.StartDate.Format(time.RFC3339),
		}
	}

	history := make(map[int][]PeakPowerHistoryPoint)
	for _, d := range powerDurations {
		points, err := s.power.GetHistory(ctx, in.AthleteID, d, in.After, in.Before, in.SportTypes)
		if err != nil {
			return nil, Wrapf(ErrInternal, "failed to load power history: %v", err)
		}
		historyPoints := make([]PeakPowerHistoryPoint, len(points))
		for i, p := range points {
			historyPoints[i] = PeakPowerHistoryPoint{
				Date:  p.Date,
				Watts: p.Watts,
			}
		}
		history[d] = historyPoints
	}

	return &PowerStatsOutput{
		DurationsS: powerDurations,
		Best:       bestOut,
		History:    history,
	}, nil
}

// GetTrainingLoad returns training load data (daily series + summary).
//
//adapter:wasm getTrainingLoad category=Stats
//adapter:http GET /api/v1/stats/training-load
func (s *StatsService) GetTrainingLoad(ctx context.Context, in GetTrainingLoadInput) (*TrainingLoadOutput, error) {
	if s.trainingLoad == nil {
		return &TrainingLoadOutput{
			Series: []DailyTrainingLoadPoint{},
		}, nil
	}

	// Apply default date range if not specified.
	after := in.After
	before := in.Before
	if after == nil && before == nil {
		t := time.Now().AddDate(0, 0, -180)
		after = &t
	}

	if err := s.trainingLoad.EnsureComputedForRange(ctx, in.AthleteID, after, before); err != nil {
		return nil, Wrapf(ErrInternal, "failed to compute training load: %v", err)
	}

	series, err := s.trainingLoad.GetDailySeries(ctx, in.AthleteID, after, before)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to load training load: %v", err)
	}

	seriesOut := make([]DailyTrainingLoadPoint, len(series))
	for i, p := range series {
		seriesOut[i] = DailyTrainingLoadPoint{
			Day: p.Day,
			TSS: p.TSS,
			CTL: p.CTL,
			ATL: p.ATL,
			TSB: p.TSB,
		}
	}

	summary, _ := s.trainingLoad.GetSummary(ctx, in.AthleteID)
	var summaryOut *DailyTrainingLoadPoint
	if summary != nil {
		summaryOut = &DailyTrainingLoadPoint{
			Day: summary.Day,
			TSS: summary.TSS,
			CTL: summary.CTL,
			ATL: summary.ATL,
			TSB: summary.TSB,
		}
	}

	// Get computation stats for diagnostics
	repoStats, err := s.trainingLoad.GetComputationStats(ctx, in.AthleteID, after, before)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get computation stats: %v", err)
	}

	var statsOut *ComputationStats
	if repoStats != nil {
		statsOut = &ComputationStats{
			TotalActivities:      repoStats.TotalActivities,
			ActivitiesWithPower:  repoStats.ActivitiesWithPower,
			ActivitiesWithTSS:    repoStats.ActivitiesWithTSS,
			CyclingFTPConfigured: repoStats.CyclingFTPConfigured,
			RunningFTPConfigured: repoStats.RunningFTPConfigured,
			HRZonesConfigured:    repoStats.HRZonesConfigured,
		}
	}

	output := &TrainingLoadOutput{
		Series:  seriesOut,
		Summary: summaryOut,
		Stats:   statsOut,
	}

	// Calculate weekly metrics if we have enough data
	if len(seriesOut) > 0 {
		output.WeeklyMetrics = s.calculateWeeklyMetrics(seriesOut)
	}

	// Get polarized breakdown for last 30 days
	if s.zoneDistribution != nil {
		output.Polarized = s.getPolarizedBreakdown(ctx, in.AthleteID)
	}

	return output, nil
}

// calculateWeeklyMetrics computes rest days, monotony, and strain from the last 7 days.
func (s *StatsService) calculateWeeklyMetrics(series []DailyTrainingLoadPoint) *WeeklyMetrics {
	// Get last 7 days of data
	n := len(series)
	startIdx := n - 7
	if startIdx < 0 {
		startIdx = 0
	}

	lastWeek := series[startIdx:]
	if len(lastWeek) == 0 {
		return nil
	}

	// Calculate rest days and collect TSS values
	restDays := 0
	var dailyTSS []float64
	totalStrain := 0.0

	for _, p := range lastWeek {
		if p.TSS == 0 {
			restDays++
		}
		dailyTSS = append(dailyTSS, p.TSS)
		totalStrain += p.TSS
	}

	// Calculate monotony (mean / stddev)
	monotony := 0.0
	if len(dailyTSS) > 1 {
		mean := totalStrain / float64(len(dailyTSS))
		var sumSqDiff float64
		for _, tss := range dailyTSS {
			diff := tss - mean
			sumSqDiff += diff * diff
		}
		stdDev := 0.0
		if len(dailyTSS) > 0 {
			stdDev = sqrtFloat(sumSqDiff / float64(len(dailyTSS)))
		}
		if stdDev > 0 {
			monotony = mean / stdDev
		}
	}

	return &WeeklyMetrics{
		RestDays:     restDays,
		Monotony:     roundTo2(monotony),
		WeeklyStrain: roundTo2(totalStrain),
		WeeklyTRIMP:  0, // Will be computed when TRIMP data is available
	}
}

// getPolarizedBreakdown returns training intensity distribution for the last 30 days.
func (s *StatsService) getPolarizedBreakdown(ctx context.Context, athleteID int64) *PolarizedBreakdown {
	// Ensure zone distributions are computed
	if err := s.zoneDistribution.EnsureComputed(ctx, athleteID); err != nil {
		return nil
	}

	// Get aggregated zone distribution for last 30 days
	data, err := s.zoneDistribution.GetWeeklyDistribution(ctx, athleteID, 5) // ~5 weeks ≈ 30+ days
	if err != nil || len(data) == 0 {
		return nil
	}

	// Sum across all weeks
	var z1, z2, z3, z4, z5, total int
	for _, w := range data {
		z1 += w.SecondsZ1
		z2 += w.SecondsZ2
		z3 += w.SecondsZ3
		z4 += w.SecondsZ4
		z5 += w.SecondsZ5
		total += w.Total
	}

	if total == 0 {
		return nil
	}

	low := float64(z1+z2) / float64(total) * 100
	moderate := float64(z3) / float64(total) * 100
	high := float64(z4+z5) / float64(total) * 100

	return &PolarizedBreakdown{
		LowPercent:      roundTo2(low),
		ModeratePercent: roundTo2(moderate),
		HighPercent:     roundTo2(high),
		TotalSeconds:    total,
		PeriodDays:      30,
	}
}

func sqrtFloat(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Newton's method for square root
	z := x / 2
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func roundTo2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

// GetZoneTrend returns weekly HR zone distribution data.
//
//adapter:wasm getZoneTrend category=Stats
//adapter:http GET /api/v1/stats/zone-trend
func (s *StatsService) GetZoneTrend(ctx context.Context, in GetZoneTrendInput) (*ZoneTrendOutput, error) {
	if s.zoneDistribution == nil {
		return nil, Wrapf(ErrInternal, "zone distribution repository not configured")
	}

	// Validate and normalize weeks parameter
	weeks := in.Weeks
	if weeks <= 0 {
		weeks = defaultZoneTrendWeeks
	} else if weeks > maxZoneTrendWeeks {
		weeks = maxZoneTrendWeeks
	}

	// Ensure zone distributions are computed for activities with HR streams
	if err := s.zoneDistribution.EnsureComputed(ctx, in.AthleteID); err != nil {
		return nil, Wrapf(ErrInternal, "failed to compute zone distributions: %v", err)
	}

	data, err := s.zoneDistribution.GetWeeklyDistribution(ctx, in.AthleteID, weeks)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get zone trend: %v", err)
	}

	result := make([]WeeklyZoneDistribution, len(data))
	for i, d := range data {
		result[i] = WeeklyZoneDistribution{
			Week:      d.Week,
			SecondsZ1: d.SecondsZ1,
			SecondsZ2: d.SecondsZ2,
			SecondsZ3: d.SecondsZ3,
			SecondsZ4: d.SecondsZ4,
			SecondsZ5: d.SecondsZ5,
			Total:     d.Total,
			PercentZ1: d.PercentZ1,
			PercentZ2: d.PercentZ2,
			PercentZ3: d.PercentZ3,
			PercentZ4: d.PercentZ4,
			PercentZ5: d.PercentZ5,
		}
	}

	return &ZoneTrendOutput{
		Weeks: result,
	}, nil
}

// GetWrappedYears returns the list of years with activity data.
//
//adapter:wasm getWrappedYears category=Stats
//adapter:http GET /api/v1/stats/wrapped/years
func (s *StatsService) GetWrappedYears(ctx context.Context, in GetWrappedYearsInput) ([]int, error) {
	years, err := s.stats.ListWrappedYears(ctx, in.AthleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to load wrapped years: %v", err)
	}
	if years == nil {
		return []int{}, nil
	}
	return years, nil
}

// GetWrapped returns the wrapped report for a given year (or all-time if year is 0).
//
//adapter:wasm getWrapped category=Stats
//adapter:http GET /api/v1/stats/wrapped
func (s *StatsService) GetWrapped(ctx context.Context, in GetWrappedInput) (*WrappedOutput, error) {
	report, err := s.stats.GetWrapped(ctx, in.AthleteID, in.Year)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to build wrapped: %v", err)
	}

	// Convert storage types to service output types
	months := make([]WrappedMonth, len(report.Months))
	for i, m := range report.Months {
		months[i] = WrappedMonth{
			Month:      m.Month,
			Activities: m.Activities,
			DistanceM:  m.DistanceM,
			ElevationM: m.ElevationM,
			PRs:        m.PRs,
		}
	}

	movingTimeBySport := make([]WrappedSportTime, len(report.MovingTimeBySport))
	for i, m := range report.MovingTimeBySport {
		movingTimeBySport[i] = WrappedSportTime{
			SportType:   m.SportType,
			MovingTimeS: m.MovingTimeS,
		}
	}

	startTimesByHour := make([]WrappedHourCount, len(report.StartTimesByHour))
	for i, h := range report.StartTimesByHour {
		startTimesByHour[i] = WrappedHourCount{
			Hour:  h.Hour,
			Count: h.Count,
		}
	}

	locations := make([]WrappedLocationPoint, len(report.Locations))
	for i, l := range report.Locations {
		locations[i] = WrappedLocationPoint{
			Lat:   l.Lat,
			Lng:   l.Lng,
			Count: l.Count,
		}
	}

	var randomPhoto *WrappedPhoto
	if report.RandomPhoto != nil {
		randomPhoto = &WrappedPhoto{
			ID:           report.RandomPhoto.ID,
			ActivityID:   report.RandomPhoto.ActivityID,
			URL:          report.RandomPhoto.URL,
			ThumbnailURL: report.RandomPhoto.ThumbnailURL,
			Caption:      report.RandomPhoto.Caption,
		}
	}

	var longestDistance, mostElevation, longestDuration *WrappedBiggestActivity
	if report.Biggest.LongestDistance != nil {
		longestDistance = &WrappedBiggestActivity{
			ActivityID:     report.Biggest.LongestDistance.ActivityID,
			Name:           report.Biggest.LongestDistance.Name,
			SportType:      report.Biggest.LongestDistance.SportType,
			StartDateLocal: report.Biggest.LongestDistance.StartDateLocal,
			Value:          report.Biggest.LongestDistance.Value,
		}
	}
	if report.Biggest.MostElevation != nil {
		mostElevation = &WrappedBiggestActivity{
			ActivityID:     report.Biggest.MostElevation.ActivityID,
			Name:           report.Biggest.MostElevation.Name,
			SportType:      report.Biggest.MostElevation.SportType,
			StartDateLocal: report.Biggest.MostElevation.StartDateLocal,
			Value:          report.Biggest.MostElevation.Value,
		}
	}
	if report.Biggest.LongestDuration != nil {
		longestDuration = &WrappedBiggestActivity{
			ActivityID:     report.Biggest.LongestDuration.ActivityID,
			Name:           report.Biggest.LongestDuration.Name,
			SportType:      report.Biggest.LongestDuration.SportType,
			StartDateLocal: report.Biggest.LongestDuration.StartDateLocal,
			Value:          report.Biggest.LongestDuration.Value,
		}
	}

	return &WrappedOutput{
		Year:       report.Year,
		RangeStart: report.RangeStart,
		RangeEnd:   report.RangeEnd,
		TotalDays:  report.TotalDays,
		ActiveDays: report.ActiveDays,
		RestDays:   report.RestDays,
		Totals: WrappedTotals{
			Activities:    report.Totals.Activities,
			DistanceM:     report.Totals.DistanceM,
			ElevationM:    report.Totals.ElevationM,
			MovingTimeS:   report.Totals.MovingTimeS,
			Kudos:         report.Totals.Kudos,
			CommuteDistM:  report.Totals.CommuteDistM,
			CarbonSavedKg: report.Totals.CarbonSavedKg,
		},
		Months:            months,
		MovingTimeBySport: movingTimeBySport,
		StartTimesByHour:  startTimesByHour,
		Locations:         locations,
		Streaks: WrappedStreaks{
			LongestActiveDays: report.Streaks.LongestActiveDays,
			LongestRestDays:   report.Streaks.LongestRestDays,
		},
		RandomPhoto: randomPhoto,
		Biggest: WrappedBiggest{
			LongestDistance: longestDistance,
			MostElevation:   mostElevation,
			LongestDuration: longestDuration,
		},
	}, nil
}

// SaveBestEfforts stores best efforts for an activity (replaces existing).
//
//adapter:wasm saveBestEfforts category=BestEfforts-Write
func (s *StatsService) SaveBestEfforts(ctx context.Context, in SaveBestEffortsInput) (*SaveBestEffortsOutput, error) {
	// Validate efforts array size
	if len(in.Efforts) > maxBestEffortsPerSave {
		return nil, BadRequestf("too many best efforts: %d > %d", len(in.Efforts), maxBestEffortsPerSave)
	}

	// Convert to storage format with canonicalization
	efforts := make([]storage.BestEffort, 0, len(in.Efforts))
	for _, e := range in.Efforts {
		// Apply canonical distance type mapping
		distanceType, canonicalM := shared.CanonicalBestEffortDistanceType(e.DistanceM, e.Name)
		be := storage.BestEffort{
			AthleteID:    in.AthleteID,
			ActivityID:   in.ActivityID,
			SportType:    in.SportType,
			DistanceType: distanceType,
			Name:         e.Name,
			DistanceM:    canonicalM,
			ElapsedTimeS: e.ElapsedTime,
			MovingTimeS:  e.MovingTime,
			StartIndex:   e.StartIndex,
			EndIndex:     e.EndIndex,
			PRRank:       e.PRRank,
		}
		if e.StartDate != "" {
			if t, err := time.Parse(time.RFC3339, e.StartDate); err == nil {
				be.StartDate = &storage.SQLiteTime{Time: t}
			}
		}
		efforts = append(efforts, be)
	}

	if err := s.bestEfforts.ReplaceForActivity(ctx, in.AthleteID, in.ActivityID, in.SportType, efforts); err != nil {
		return nil, Wrapf(ErrInternal, "saving best efforts: %v", err)
	}

	return &SaveBestEffortsOutput{
		Message: fmt.Sprintf("Best efforts for activity %d saved (%d efforts)", in.ActivityID, len(efforts)),
	}, nil
}
