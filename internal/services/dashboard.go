package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
)

// DashboardService handles dashboard business logic.
type DashboardService struct {
	db     *storage.DB
	stats  *storage.StatsRepository
	config *storage.DashboardConfigRepository
}

// NewDashboardService creates a new dashboard service.
func NewDashboardService(db *storage.DB, stats *storage.StatsRepository, config *storage.DashboardConfigRepository) *DashboardService {
	return &DashboardService{
		db:     db,
		stats:  stats,
		config: config,
	}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// GetDashboardInput contains parameters for getting the combined dashboard.
type GetDashboardInput struct {
	AthleteID int64
}

// DashboardStatsOutput represents aggregated statistics for the dashboard.
type DashboardStatsOutput struct {
	TotalActivities    int     `json:"total_activities"`
	TotalDistance      float64 `json:"total_distance"`
	TotalMovingTime    int     `json:"total_moving_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	TotalCalories      float64 `json:"total_calories"`
	YearActivities     int     `json:"year_activities"`
	YearDistance       float64 `json:"year_distance"`
	YearMovingTime     int     `json:"year_moving_time"`
	YearElevationGain  float64 `json:"year_elevation_gain"`
	MonthActivities    int     `json:"month_activities"`
	MonthDistance      float64 `json:"month_distance"`
	MonthMovingTime    int     `json:"month_moving_time"`
	MonthElevationGain float64 `json:"month_elevation_gain"`
}

// WeeklyStatOutput represents statistics for a single sport type in the current week.
type WeeklyStatOutput struct {
	SportType      string  `json:"sport_type"`
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// RecentActivityOutput represents a simplified activity for the dashboard.
type RecentActivityOutput struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	SportType       string  `json:"sport_type"`
	StartDate       string  `json:"start_date"`
	Distance        float64 `json:"distance"`
	MovingTime      int     `json:"moving_time"`
	ElevationGain   float64 `json:"elevation_gain"`
	SummaryPolyline string  `json:"summary_polyline,omitempty"`
}

// SportTypeStatOutput represents statistics for a single sport type.
type SportTypeStatOutput struct {
	SportType      string  `json:"sport_type"`
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// DashboardOutput combines all dashboard data.
type DashboardOutput struct {
	Stats            *DashboardStatsOutput  `json:"stats"`
	WeeklyStats      []WeeklyStatOutput     `json:"weekly_stats"`
	RecentActivities []RecentActivityOutput `json:"recent_activities"`
	SportTypeStats   []SportTypeStatOutput  `json:"sport_type_stats"`
}

// GetRecentActivitiesInput contains parameters for getting recent activities.
type GetRecentActivitiesInput struct {
	AthleteID int64
	Limit     int
}

// GetMonthlyStatsInput contains parameters for getting monthly stats.
type GetMonthlyStatsInput struct {
	AthleteID int64
	Year      int // 0 for all years
}

// MonthlyStatOutput represents statistics for a single month.
type MonthlyStatOutput struct {
	Month          string  `json:"month"`
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// YearlyStatOutput represents statistics for a single year.
type YearlyStatOutput struct {
	Year           int     `json:"year"`
	ActivityCount  int     `json:"activity_count"`
	TotalDistance  float64 `json:"total_distance"`
	TotalTime      int     `json:"total_time"`
	TotalElevation float64 `json:"total_elevation"`
}

// GetCalendarDataInput contains parameters for getting calendar data.
type GetCalendarDataInput struct {
	AthleteID int64
	Year      int
}

// CalendarDayOutput represents activity data for a single day.
type CalendarDayOutput struct {
	Date          string  `json:"date"`
	ActivityCount int     `json:"activity_count"`
	TotalDistance float64 `json:"total_distance"`
	TotalTime     int     `json:"total_time"`
	TotalCalories float64 `json:"total_calories"`
}

// GetCalendarActivitiesInput contains parameters for getting calendar activities.
type GetCalendarActivitiesInput struct {
	AthleteID int64
	Year      int
	Month     int
}

// CalendarActivityOutput represents an activity summary for the calendar view.
type CalendarActivityOutput struct {
	ID                 int64   `json:"id"`
	Name               string  `json:"name"`
	SportType          string  `json:"sport_type"`
	StartDate          string  `json:"start_date"`
	Distance           float64 `json:"distance"`
	MovingTime         int     `json:"moving_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
}

// CalendarMonthSummaryOutput represents a monthly summary for the calendar.
type CalendarMonthSummaryOutput struct {
	Year                int     `json:"year"`
	Month               int     `json:"month"`
	ActivityCount       int     `json:"activity_count"`
	TotalDistance       float64 `json:"total_distance"`
	TotalElevationGain  float64 `json:"total_elevation_gain"`
	TotalMovingTime     int     `json:"total_moving_time"`
	TotalCalories       float64 `json:"total_calories"`
	WorkoutCount        int     `json:"workout_count"`
	ChallengesCompleted int     `json:"challenges_completed"`
}

// DistributionSliceOutput represents a single slice of a distribution.
type DistributionSliceOutput struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// GetDistributionInput contains parameters for distribution queries.
type GetDistributionInput struct {
	AthleteID  int64
	SportTypes []string
}

// ExportStatsOutput represents export statistics.
type ExportStatsOutput struct {
	TotalActivities int     `json:"total_activities"`
	FirstActivity   *string `json:"first_activity"`
	LastActivity    *string `json:"last_activity"`
}

// DashboardConfigOutput represents the dashboard configuration.
type DashboardConfigOutput struct {
	Version int                           `json:"version"`
	Widgets []DashboardWidgetConfigOutput `json:"widgets"`
}

// DashboardWidgetConfigOutput represents a widget configuration.
type DashboardWidgetConfigOutput struct {
	ID       string         `json:"id"`
	Width    int            `json:"width"`
	Height   int            `json:"height,omitempty"`
	Hidden   bool           `json:"hidden"`
	Settings map[string]any `json:"settings,omitempty"`
}

// UpdateDashboardConfigInput contains parameters for updating dashboard config.
type UpdateDashboardConfigInput struct {
	AthleteID int64
	Config    DashboardConfigOutput
}

// ============================================================================
// Service Methods
// ============================================================================

// GetDashboard returns the combined dashboard data.
func (s *DashboardService) GetDashboard(ctx context.Context, in GetDashboardInput) (*DashboardOutput, error) {
	stats, err := s.GetStats(ctx, in)
	if err != nil {
		return nil, err
	}

	weeklyStats, err := s.GetWeeklyStats(ctx, in)
	if err != nil {
		return nil, err
	}

	recentActivities, err := s.GetRecentActivities(ctx, GetRecentActivitiesInput{
		AthleteID: in.AthleteID,
		Limit:     5,
	})
	if err != nil {
		return nil, err
	}

	sportTypeStats, err := s.GetSportTypeStats(ctx, in)
	if err != nil {
		return nil, err
	}

	return &DashboardOutput{
		Stats:            stats,
		WeeklyStats:      weeklyStats,
		RecentActivities: recentActivities,
		SportTypeStats:   sportTypeStats,
	}, nil
}

// GetStats returns aggregated dashboard statistics.
func (s *DashboardService) GetStats(ctx context.Context, in GetDashboardInput) (*DashboardStatsOutput, error) {
	stats, err := s.stats.GetDashboardStats(ctx, in.AthleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get dashboard stats: %v", err)
	}

	return &DashboardStatsOutput{
		TotalActivities:    stats.TotalActivities,
		TotalDistance:      stats.TotalDistance,
		TotalMovingTime:    stats.TotalMovingTime,
		TotalElevationGain: stats.TotalElevationGain,
		TotalCalories:      stats.TotalCalories,
		YearActivities:     stats.YearActivities,
		YearDistance:       stats.YearDistance,
		YearMovingTime:     stats.YearMovingTime,
		YearElevationGain:  stats.YearElevationGain,
		MonthActivities:    stats.MonthActivities,
		MonthDistance:      stats.MonthDistance,
		MonthMovingTime:    stats.MonthMovingTime,
		MonthElevationGain: stats.MonthElevationGain,
	}, nil
}

// GetWeeklyStats returns statistics grouped by sport type for the current week.
func (s *DashboardService) GetWeeklyStats(ctx context.Context, in GetDashboardInput) ([]WeeklyStatOutput, error) {
	stats, err := s.stats.GetWeeklyStats(ctx, in.AthleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get weekly stats: %v", err)
	}

	out := make([]WeeklyStatOutput, len(stats))
	for i, stat := range stats {
		out[i] = WeeklyStatOutput{
			SportType:      stat.SportType,
			ActivityCount:  stat.ActivityCount,
			TotalDistance:  stat.TotalDistance,
			TotalTime:      stat.TotalTime,
			TotalElevation: stat.TotalElevation,
		}
	}

	return out, nil
}

// GetRecentActivities returns the most recent activities.
func (s *DashboardService) GetRecentActivities(ctx context.Context, in GetRecentActivitiesInput) ([]RecentActivityOutput, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	activities, err := s.stats.GetRecentActivities(ctx, in.AthleteID, limit)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get recent activities: %v", err)
	}

	out := make([]RecentActivityOutput, len(activities))
	for i, a := range activities {
		out[i] = RecentActivityOutput{
			ID:              a.ID,
			Name:            a.Name,
			SportType:       a.SportType,
			StartDate:       a.StartDate.Format(time.RFC3339),
			Distance:        a.Distance,
			MovingTime:      a.MovingTime,
			ElevationGain:   a.ElevationGain,
			SummaryPolyline: a.SummaryPolyline,
		}
	}

	return out, nil
}

// GetSportTypeStats returns statistics grouped by sport type.
func (s *DashboardService) GetSportTypeStats(ctx context.Context, in GetDashboardInput) ([]SportTypeStatOutput, error) {
	stats, err := s.stats.GetStatsBySportType(ctx, in.AthleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get sport type stats: %v", err)
	}

	out := make([]SportTypeStatOutput, len(stats))
	for i, stat := range stats {
		out[i] = SportTypeStatOutput{
			SportType:      stat.SportType,
			ActivityCount:  stat.ActivityCount,
			TotalDistance:  stat.TotalDistance,
			TotalTime:      stat.TotalTime,
			TotalElevation: stat.TotalElevation,
		}
	}

	return out, nil
}

// GetMonthlyStats returns statistics grouped by month.
func (s *DashboardService) GetMonthlyStats(ctx context.Context, in GetMonthlyStatsInput) ([]MonthlyStatOutput, error) {
	stats, err := s.stats.GetMonthlyStats(ctx, in.AthleteID, in.Year)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get monthly stats: %v", err)
	}

	out := make([]MonthlyStatOutput, len(stats))
	for i, stat := range stats {
		out[i] = MonthlyStatOutput{
			Month:          stat.Month,
			ActivityCount:  stat.ActivityCount,
			TotalDistance:  stat.TotalDistance,
			TotalTime:      stat.TotalTime,
			TotalElevation: stat.TotalElevation,
		}
	}

	return out, nil
}

// GetYearlyStats returns statistics grouped by year.
func (s *DashboardService) GetYearlyStats(ctx context.Context, in GetDashboardInput) ([]YearlyStatOutput, error) {
	stats, err := s.stats.GetYearlyStats(ctx, in.AthleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get yearly stats: %v", err)
	}

	out := make([]YearlyStatOutput, len(stats))
	for i, stat := range stats {
		out[i] = YearlyStatOutput{
			Year:           stat.Year,
			ActivityCount:  stat.ActivityCount,
			TotalDistance:  stat.TotalDistance,
			TotalTime:      stat.TotalTime,
			TotalElevation: stat.TotalElevation,
		}
	}

	return out, nil
}

// GetCalendarData returns daily activity counts for a given year.
func (s *DashboardService) GetCalendarData(ctx context.Context, in GetCalendarDataInput) ([]CalendarDayOutput, error) {
	year := in.Year
	if year == 0 {
		year = time.Now().Year()
	}

	data, err := s.stats.GetCalendarData(ctx, in.AthleteID, year)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get calendar data: %v", err)
	}

	out := make([]CalendarDayOutput, len(data))
	for i, d := range data {
		out[i] = CalendarDayOutput{
			Date:          d.Date,
			ActivityCount: d.ActivityCount,
			TotalDistance: d.TotalDistance,
			TotalTime:     d.TotalTime,
			TotalCalories: d.TotalCalories,
		}
	}

	return out, nil
}

// GetCalendarActivities returns activities for a specific month.
func (s *DashboardService) GetCalendarActivities(ctx context.Context, in GetCalendarActivitiesInput) ([]CalendarActivityOutput, error) {
	year := in.Year
	month := in.Month
	if year == 0 {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}

	activities, err := s.stats.GetCalendarActivities(ctx, in.AthleteID, year, month)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get calendar activities: %v", err)
	}

	out := make([]CalendarActivityOutput, len(activities))
	for i, a := range activities {
		out[i] = CalendarActivityOutput{
			ID:                 a.ID,
			Name:               a.Name,
			SportType:          a.SportType,
			StartDate:          a.StartDate.Format(time.RFC3339),
			Distance:           a.Distance,
			MovingTime:         a.MovingTime,
			TotalElevationGain: a.TotalElevationGain,
		}
	}

	return out, nil
}

// GetCalendarSummary returns the monthly summary for the calendar.
func (s *DashboardService) GetCalendarSummary(ctx context.Context, in GetCalendarActivitiesInput) (*CalendarMonthSummaryOutput, error) {
	year := in.Year
	month := in.Month
	if year == 0 {
		now := time.Now()
		year = now.Year()
		month = int(now.Month())
	}
	if month < 1 || month > 12 {
		month = int(time.Now().Month())
	}

	summary, err := s.stats.GetCalendarMonthSummary(ctx, in.AthleteID, year, month)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get calendar summary: %v", err)
	}

	return &CalendarMonthSummaryOutput{
		Year:                summary.Year,
		Month:               summary.Month,
		ActivityCount:       summary.ActivityCount,
		TotalDistance:       summary.TotalDistance,
		TotalElevationGain:  summary.TotalElevationGain,
		TotalMovingTime:     summary.TotalMovingTime,
		TotalCalories:       summary.TotalCalories,
		WorkoutCount:        summary.WorkoutCount,
		ChallengesCompleted: summary.ChallengesCompleted,
	}, nil
}

// GetDaytimeDistribution returns activity counts by time of day.
func (s *DashboardService) GetDaytimeDistribution(ctx context.Context, in GetDistributionInput) ([]DistributionSliceOutput, error) {
	// Build query with optional sport type filter
	query := `
		SELECT
			CASE
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 5 AND 11 THEN 'Morning'
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 12 AND 16 THEN 'Afternoon'
				WHEN CAST(strftime('%H', start_date_local) AS INTEGER) BETWEEN 17 AND 21 THEN 'Evening'
				ELSE 'Night'
			END AS bucket,
			COUNT(*) AS count
		FROM activities
		WHERE athlete_id = ?
	`
	args := []any{in.AthleteID}

	if len(in.SportTypes) > 0 {
		placeholders := make([]string, len(in.SportTypes))
		for i, st := range in.SportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + strings.Join(placeholders, ",") + ")"
	}
	query += " GROUP BY bucket"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get daytime distribution: %v", err)
	}
	defer func() { _ = rows.Close() }()

	counts := map[string]int{}
	for rows.Next() {
		var bucket string
		var count int
		if err := rows.Scan(&bucket, &count); err != nil {
			return nil, Wrapf(ErrInternal, "failed to scan daytime distribution: %v", err)
		}
		counts[bucket] = count
	}
	if err := rows.Err(); err != nil {
		return nil, Wrapf(ErrInternal, "failed to iterate daytime distribution: %v", err)
	}

	out := make([]DistributionSliceOutput, len(shared.DaytimeLabels))
	for i, k := range shared.DaytimeLabels {
		out[i] = DistributionSliceOutput{Label: k, Count: counts[k]}
	}

	return out, nil
}

// GetWeekdayDistribution returns activity counts by day of week.
func (s *DashboardService) GetWeekdayDistribution(ctx context.Context, in GetDistributionInput) ([]DistributionSliceOutput, error) {
	query := `
		SELECT CAST(strftime('%w', start_date_local) AS INTEGER) AS weekday, COUNT(*) AS count
		FROM activities
		WHERE athlete_id = ?
	`
	args := []any{in.AthleteID}

	if len(in.SportTypes) > 0 {
		placeholders := make([]string, len(in.SportTypes))
		for i, st := range in.SportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + strings.Join(placeholders, ",") + ")"
	}
	query += " GROUP BY weekday"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get weekday distribution: %v", err)
	}
	defer func() { _ = rows.Close() }()

	counts := map[int]int{}
	for rows.Next() {
		var weekday int
		var count int
		if err := rows.Scan(&weekday, &count); err != nil {
			return nil, Wrapf(ErrInternal, "failed to scan weekday distribution: %v", err)
		}
		counts[weekday] = count
	}
	if err := rows.Err(); err != nil {
		return nil, Wrapf(ErrInternal, "failed to iterate weekday distribution: %v", err)
	}

	out := make([]DistributionSliceOutput, len(shared.WeekdayLabels))
	for i, label := range shared.WeekdayLabels {
		out[i] = DistributionSliceOutput{Label: label, Count: counts[i]}
	}

	return out, nil
}

// GetExportStats returns export statistics (total count, date range).
func (s *DashboardService) GetExportStats(ctx context.Context, in GetDashboardInput) (*ExportStatsOutput, error) {
	query := `
		SELECT
			COUNT(*) as total,
			MIN(start_date_local) as first_activity,
			MAX(start_date_local) as last_activity
		FROM activities
		WHERE athlete_id = ?
	`

	var total int
	var firstActivity, lastActivity *string

	row := s.db.QueryRowContext(ctx, query, in.AthleteID)
	if err := row.Scan(&total, &firstActivity, &lastActivity); err != nil {
		return nil, Wrapf(ErrInternal, "failed to get export stats: %v", err)
	}

	return &ExportStatsOutput{
		TotalActivities: total,
		FirstActivity:   firstActivity,
		LastActivity:    lastActivity,
	}, nil
}

// GetConfig returns the dashboard configuration for an athlete.
func (s *DashboardService) GetConfig(ctx context.Context, in GetDashboardInput) (*DashboardConfigOutput, error) {
	if s.config == nil {
		return nil, Wrapf(ErrInternal, "dashboard config repository not available")
	}

	cfg, err := s.config.Get(ctx, in.AthleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get dashboard config: %v", err)
	}

	return configToOutput(cfg), nil
}

// UpdateConfig updates the dashboard configuration for an athlete.
func (s *DashboardService) UpdateConfig(ctx context.Context, in UpdateDashboardConfigInput) (*DashboardConfigOutput, error) {
	if s.config == nil {
		return nil, Wrapf(ErrInternal, "dashboard config repository not available")
	}

	// Validate config
	if err := validateDashboardConfig(in.Config); err != nil {
		return nil, err
	}

	// Convert to storage type
	cfg := outputToConfig(in.Config)

	if err := s.config.Upsert(ctx, in.AthleteID, cfg); err != nil {
		return nil, Wrapf(ErrInternal, "failed to save dashboard config: %v", err)
	}

	return &in.Config, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateDashboardConfig(cfg DashboardConfigOutput) error {
	seen := make(map[string]bool)
	for _, w := range cfg.Widgets {
		if strings.TrimSpace(w.ID) == "" {
			return BadRequest("widget id is required")
		}
		if seen[w.ID] {
			return BadRequest(fmt.Sprintf("duplicate widget id: %s", w.ID))
		}
		seen[w.ID] = true

		switch storage.WidgetWidth(w.Width) {
		case storage.WidgetWidthOneThird, storage.WidgetWidthHalf, storage.WidgetWidthTwoThird, storage.WidgetWidthFull:
		default:
			return BadRequest(fmt.Sprintf("invalid widget width for %s", w.ID))
		}
	}
	return nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

func configToOutput(cfg *storage.DashboardConfig) *DashboardConfigOutput {
	widgets := make([]DashboardWidgetConfigOutput, len(cfg.Widgets))
	for i, w := range cfg.Widgets {
		widgets[i] = DashboardWidgetConfigOutput{
			ID:       w.ID,
			Width:    int(w.Width),
			Height:   int(w.Height),
			Hidden:   w.Hidden,
			Settings: w.Settings,
		}
	}
	return &DashboardConfigOutput{
		Version: cfg.Version,
		Widgets: widgets,
	}
}

func outputToConfig(out DashboardConfigOutput) storage.DashboardConfig {
	widgets := make([]storage.DashboardWidgetConfig, len(out.Widgets))
	for i, w := range out.Widgets {
		widgets[i] = storage.DashboardWidgetConfig{
			ID:       w.ID,
			Width:    storage.WidgetWidth(w.Width),
			Height:   storage.WidgetHeight(w.Height),
			Hidden:   w.Hidden,
			Settings: w.Settings,
		}
	}
	return storage.DashboardConfig{
		Version: out.Version,
		Widgets: widgets,
	}
}
