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

	// ATL/CTL ratio for overtraining risk.
	overtrainingRiskRatio = 1.5 // ATL > CTL * this ratio indicates risk

	// Analysis configuration.
	minDataPointsForCTL = 7 // Minimum data points for CTL trend analysis
	maxInsightsToReturn = 4 // Maximum insights to return
)

// InsightsService analyzes training load metrics and generates coaching insights.
type InsightsService struct {
	trainingLoad *storage.TrainingLoadRepository
	logger       *slog.Logger
}

// NewInsightsService creates a new insights service.
func NewInsightsService(
	trainingLoad *storage.TrainingLoadRepository,
) *InsightsService {
	return &InsightsService{
		trainingLoad: trainingLoad,
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

	// Get historical training load for CTL trend analysis.
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	series, err := s.trainingLoad.GetDailySeries(ctx, in.AthleteID, &sevenDaysAgo, nil)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to get training load series: %v", err)
	}

	// Sort series by date to ensure oldest-first ordering for trend analysis.
	sort.Slice(series, func(i, j int) bool {
		return series[i].Day < series[j].Day
	})

	// Apply training load rules to generate insights.
	if summary != nil {
		insights = append(insights, s.applyTrainingLoadRules(summary, series)...)
	}

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
