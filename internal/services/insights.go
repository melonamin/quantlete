package services

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

// Insight rule thresholds - these define when insights are triggered.
const (
	// TSB thresholds for fatigue/recovery state.
	tsbHighFatigueThreshold   = -20.0 // TSB below this indicates high fatigue
	tsbWellRecoveredThreshold = 10.0  // TSB above this indicates well recovered
	tsbBalancedLow            = -10.0 // Lower bound for balanced training
	tsbBalancedHigh           = 10.0  // Upper bound for balanced training

	// CTL thresholds for fitness trends.
	ctlDropPercentThreshold = -10.0 // CTL drop percentage that triggers warning
	ctlMinForAnalysis       = 5.0   // Minimum CTL for meaningful percentage calculations
	ctlPeakLookbackDays     = 90    // Days to check for CTL peak
	ctlRampThreshold        = 15.0  // % CTL increase that's "too fast"
	ctlRampWindowDays       = 7     // Days to measure ramp rate

	// ATL/CTL ratio for overtraining risk.
	overtrainingRiskRatio = 1.5 // ATL > CTL * this ratio indicates risk

	// Power PR detection.
	powerPRWindowDays = 7 // Days to look for recent power PRs

	// Activity pattern thresholds.
	restDayThreshold    = 7  // Consecutive days without rest triggers warning
	consistencyWeeks    = 4  // Weeks of consistent training for positive insight
	varietyThreshold    = 10 // Same sport activities before suggesting variety
	patternLookbackDays = 30 // Days to analyze for patterns

	// Analysis configuration.
	minDataPointsForCTL = 7 // Minimum data points for CTL trend analysis
	maxInsightsToReturn = 4 // Maximum insights to return
)

// InsightsService analyzes training load metrics and generates coaching insights.
type InsightsService struct {
	trainingLoad *storage.TrainingLoadRepository
	power        *storage.PowerRepository
	activities   *storage.ActivityRepository
	logger       *slog.Logger
}

// NewInsightsService creates a new insights service.
func NewInsightsService(
	trainingLoad *storage.TrainingLoadRepository,
	power *storage.PowerRepository,
	activities *storage.ActivityRepository,
) *InsightsService {
	return &InsightsService{
		trainingLoad: trainingLoad,
		power:        power,
		activities:   activities,
		logger:       slog.Default().With("service", "insights"),
	}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// GetInsightsInput contains parameters for getting training insights.
type GetInsightsInput struct {
	AthleteID int64 `json:"-" adapter:"context"`
}

// Insight represents a single coaching insight.
type Insight struct {
	ID          string `json:"id"`
	Type        string `json:"type"`     // fatigue, recovery, streak, fitness
	Severity    string `json:"severity"` // info, warning, success
	Title       string `json:"title"`
	Description string `json:"description"`
}

// InsightsOutput contains the list of coaching insights.
type InsightsOutput struct {
	Insights []Insight `json:"insights"`
}

// Severity priority for sorting (lower = higher priority).
var severityPriority = map[string]int{
	"warning": 0,
	"success": 1,
	"info":    2,
}

// severityRank returns the priority rank for a severity level.
// Unknown severities get the lowest priority (highest rank number).
func severityRank(s string) int {
	if rank, ok := severityPriority[s]; ok {
		return rank
	}
	return len(severityPriority) // Unknown severities get lowest priority
}

// ============================================================================
// Service Methods
// ============================================================================

// GetInsights returns coaching insights based on training load metrics.
//
//adapter:wasm getInsights category=Stats
//adapter:http GET /api/v1/stats/insights
func (s *InsightsService) GetInsights(ctx context.Context, in GetInsightsInput) (*InsightsOutput, error) {
	start := time.Now()
	s.logger.Debug("generating insights", "athlete_id", in.AthleteID)

	var insights []Insight

	// Get current training load summary (latest TSB, CTL, ATL).
	summary, err := s.trainingLoad.GetSummary(ctx, in.AthleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get training load summary: %v", err)
	}

	// Get 90-day CTL history for fitness peak and ramp detection.
	ninetyDaysAgo := time.Now().AddDate(0, 0, -ctlPeakLookbackDays)
	ctlSeries, err := s.trainingLoad.GetDailySeries(ctx, in.AthleteID, &ninetyDaysAgo, nil)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get training load series: %v", err)
	}

	// Sort series by date to ensure oldest-first ordering for trend analysis.
	sort.Slice(ctlSeries, func(i, j int) bool {
		return ctlSeries[i].Day < ctlSeries[j].Day
	})

	// Get recent activities for pattern analysis.
	patternStart := time.Now().AddDate(0, 0, -patternLookbackDays)
	activities, _, err := s.activities.List(ctx, storage.ActivityFilters{
		AthleteID:  in.AthleteID,
		StartAfter: &patternStart,
	}, storage.Pagination{PerPage: 500})
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get activities: %v", err)
	}

	// Get recent power PRs.
	prWindow := time.Now().AddDate(0, 0, -powerPRWindowDays)
	powerPRs, err := s.power.GetBest(ctx, in.AthleteID, []int{5, 60, 300, 1200}, &prWindow, nil, nil)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get power PRs: %v", err)
	}

	// Apply training load rules to generate insights.
	if summary != nil {
		insights = append(insights, s.applyTrainingLoadRules(summary, ctlSeries)...)
	}

	// Apply CTL-based insights (fitness peak, ramp rate).
	if len(ctlSeries) > 0 {
		insights = append(insights, s.applyCTLInsights(ctlSeries)...)
	}

	// Apply power PR insight.
	insights = append(insights, s.applyPowerPRInsight(powerPRs)...)

	// Apply activity pattern insights.
	insights = append(insights, s.applyActivityPatternInsights(activities)...)

	// Sort by severity priority (stable sort preserves order within same priority).
	sort.SliceStable(insights, func(i, j int) bool {
		return severityRank(insights[i].Severity) < severityRank(insights[j].Severity)
	})

	if len(insights) > maxInsightsToReturn {
		insights = insights[:maxInsightsToReturn]
	}

	s.logger.Debug("insights generated",
		"athlete_id", in.AthleteID,
		"insight_count", len(insights),
		"duration_ms", time.Since(start).Milliseconds(),
	)

	return &InsightsOutput{Insights: insights}, nil
}

// applyTrainingLoadRules generates insights from training load metrics.
func (s *InsightsService) applyTrainingLoadRules(summary *storage.DailyTrainingLoadPoint, series []storage.DailyTrainingLoadPoint) []Insight {
	var insights []Insight

	tsb := summary.TSB
	ctl := summary.CTL
	atl := summary.ATL

	// Rule 1: High Fatigue.
	if tsb < tsbHighFatigueThreshold {
		insights = append(insights, Insight{
			ID:          "high_fatigue",
			Type:        "fatigue",
			Severity:    "warning",
			Title:       "High fatigue detected",
			Description: fmt.Sprintf("Your TSB is %.0f. Consider a recovery day.", tsb),
		})
	}

	// Rule 2: Overtraining Risk.
	if ctl > 0 && atl > ctl*overtrainingRiskRatio {
		insights = append(insights, Insight{
			ID:          "overtraining_risk",
			Type:        "fitness",
			Severity:    "warning",
			Title:       "Overtraining risk",
			Description: "Your acute load is significantly higher than chronic fitness.",
		})
	}

	// Rule 3: CTL Drop - only analyze with sufficient data and meaningful CTL values.
	if len(series) >= minDataPointsForCTL {
		oldCTL := series[0].CTL
		newCTL := series[len(series)-1].CTL
		// Only calculate percentage if CTL is above minimum threshold.
		if oldCTL > ctlMinForAnalysis {
			ctlChange := ((newCTL - oldCTL) / oldCTL) * 100
			if ctlChange < ctlDropPercentThreshold {
				insights = append(insights, Insight{
					ID:          "ctl_drop",
					Type:        "fitness",
					Severity:    "warning",
					Title:       "Fitness declining",
					Description: fmt.Sprintf("Your CTL has dropped %.0f%% this week.", -ctlChange),
				})
			}
		}
	}

	// Rule 4: Well Recovered.
	if tsb > tsbWellRecoveredThreshold {
		insights = append(insights, Insight{
			ID:          "well_recovered",
			Type:        "recovery",
			Severity:    "success",
			Title:       "Well recovered",
			Description: "You're fresh and ready for a big effort!",
		})
	}

	// Rule 5: Building Fitness (balanced TSB with positive CTL).
	if tsb >= tsbBalancedLow && tsb <= tsbBalancedHigh && ctl > 0 {
		insights = append(insights, Insight{
			ID:          "building_fitness",
			Type:        "fitness",
			Severity:    "info",
			Title:       "Building fitness",
			Description: "You're in a good training balance.",
		})
	}

	return insights
}

// applyCTLInsights generates insights from CTL trends (fitness peak, ramp rate).
func (s *InsightsService) applyCTLInsights(series []storage.DailyTrainingLoadPoint) []Insight {
	var insights []Insight

	if len(series) < minDataPointsForCTL {
		return insights
	}

	currentCTL := series[len(series)-1].CTL

	// Rule: Fitness Peak - CTL at 90-day high.
	maxCTL := 0.0
	for _, p := range series[:len(series)-1] { // Exclude current day.
		if p.CTL > maxCTL {
			maxCTL = p.CTL
		}
	}
	if currentCTL > ctlMinForAnalysis && currentCTL >= maxCTL {
		insights = append(insights, Insight{
			ID:          "fitness_peak",
			Type:        "fitness",
			Severity:    "success",
			Title:       "Peak fitness!",
			Description: fmt.Sprintf("Your CTL (%.0f) is at its highest in %d days. Great time for a race!", currentCTL, ctlPeakLookbackDays),
		})
	}

	// Rule: Ramp Too Fast - CTL increased >15% in 7 days.
	if len(series) >= ctlRampWindowDays {
		weekAgoCTL := series[len(series)-ctlRampWindowDays].CTL
		if weekAgoCTL > ctlMinForAnalysis {
			rampPercent := ((currentCTL - weekAgoCTL) / weekAgoCTL) * 100
			if rampPercent > ctlRampThreshold {
				insights = append(insights, Insight{
					ID:          "ramp_too_fast",
					Type:        "fitness",
					Severity:    "warning",
					Title:       "Ramp rate warning",
					Description: fmt.Sprintf("Your CTL increased %.0f%% in %d days. Watch for injury signs.", rampPercent, ctlRampWindowDays),
				})
			}
		}
	}

	return insights
}

// applyPowerPRInsight generates insight for recent power PRs.
func (s *InsightsService) applyPowerPRInsight(prs []storage.PeakPowerBest) []Insight {
	var insights []Insight

	if len(prs) == 0 {
		return insights
	}

	// Find the most impressive PR (longest duration with good power).
	// Prefer longer durations as they're harder to set PRs on.
	var bestPR *storage.PeakPowerBest
	for i := range prs {
		if prs[i].Watts > 0 {
			if bestPR == nil || prs[i].DurationS > bestPR.DurationS {
				bestPR = &prs[i]
			}
		}
	}

	if bestPR != nil {
		durationStr := formatDuration(bestPR.DurationS)
		insights = append(insights, Insight{
			ID:          "power_pr",
			Type:        "fitness",
			Severity:    "success",
			Title:       "New power PR!",
			Description: fmt.Sprintf("You set a new %s power record (%.0fW)!", durationStr, bestPR.Watts),
		})
	}

	return insights
}

// applyActivityPatternInsights generates insights from activity patterns.
func (s *InsightsService) applyActivityPatternInsights(activities []storage.Activity) []Insight {
	var insights []Insight

	if len(activities) == 0 {
		return insights
	}

	// Build a set of activity dates and sport type counts.
	activityDates := make(map[string]bool)
	sportCounts := make(map[string]int)
	for _, a := range activities {
		dayStr := a.StartDateLocal.Format("2006-01-02")
		activityDates[dayStr] = true
		sportCounts[a.SportType]++
	}

	// Rule: Rest Day Needed - consecutive days without rest.
	consecutiveDays := countConsecutiveActivityDays(activityDates)
	if consecutiveDays >= restDayThreshold {
		insights = append(insights, Insight{
			ID:          "rest_day_needed",
			Type:        "recovery",
			Severity:    "warning",
			Title:       "Take a rest day",
			Description: fmt.Sprintf("You've trained %d days straight. Rest is when you get stronger.", consecutiveDays),
		})
	}

	// Rule: Consistency King - activities every week for 4+ weeks.
	weeksWithActivity := countWeeksWithActivity(activityDates)
	if weeksWithActivity >= consistencyWeeks {
		insights = append(insights, Insight{
			ID:          "consistency_king",
			Type:        "fitness",
			Severity:    "success",
			Title:       "Consistent training",
			Description: fmt.Sprintf("%d weeks of regular training. Consistency beats intensity!", weeksWithActivity),
		})
	}

	// Rule: Variety Check - same sport for 10+ consecutive activities.
	if len(activities) >= varietyThreshold {
		// Check last N activities for same sport.
		lastSport := activities[0].SportType // Most recent.
		sameCount := 0
		for _, a := range activities {
			if a.SportType == lastSport {
				sameCount++
			} else {
				break
			}
		}
		if sameCount >= varietyThreshold {
			insights = append(insights, Insight{
				ID:          "variety_check",
				Type:        "fitness",
				Severity:    "info",
				Title:       "Mix it up?",
				Description: fmt.Sprintf("%d %s in a row. Cross-training can prevent overuse injuries.", sameCount, lastSport),
			})
		}
	}

	return insights
}

// countConsecutiveActivityDays counts consecutive days with activities ending today.
func countConsecutiveActivityDays(dates map[string]bool) int {
	today := time.Now()
	count := 0
	for i := 0; i < 30; i++ { // Check up to 30 days back.
		dayStr := today.AddDate(0, 0, -i).Format("2006-01-02")
		if dates[dayStr] {
			count++
		} else {
			break
		}
	}
	return count
}

// countWeeksWithActivity counts distinct weeks (Mon-Sun) with at least one activity.
func countWeeksWithActivity(dates map[string]bool) int {
	weeksWithActivity := make(map[string]bool)
	for dateStr := range dates {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		// Get ISO week (year, week number).
		year, week := t.ISOWeek()
		weekKey := fmt.Sprintf("%d-%02d", year, week)
		weeksWithActivity[weekKey] = true
	}
	return len(weeksWithActivity)
}

// formatDuration formats seconds into human-readable duration.
func formatDuration(seconds int) string {
	switch {
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%dm", seconds/60)
	default:
		return fmt.Sprintf("%dh", seconds/3600)
	}
}
