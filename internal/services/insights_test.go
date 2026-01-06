package services

import (
	"testing"
	"time"

	"github.com/melonamin/quantlete/internal/storage"
)

func TestInsightsService_ApplyTrainingLoadRules(t *testing.T) {
	// Create a minimal service for testing (no repos needed for rule testing).
	svc := &InsightsService{}

	tests := []struct {
		name           string
		summary        *storage.DailyTrainingLoadPoint
		series         []storage.DailyTrainingLoadPoint
		wantInsightIDs []string
	}{
		{
			name: "high fatigue (TSB < -20)",
			summary: &storage.DailyTrainingLoadPoint{
				TSB: -25,
				CTL: 50,
				ATL: 75,
			},
			series:         nil,
			wantInsightIDs: []string{"high_fatigue"},
		},
		{
			name: "overtraining risk (ATL > CTL * 1.5)",
			summary: &storage.DailyTrainingLoadPoint{
				TSB: -5,
				CTL: 40,
				ATL: 65, // 65 > 40 * 1.5 = 60
			},
			series:         nil,
			wantInsightIDs: []string{"overtraining_risk", "building_fitness"},
		},
		{
			name: "well recovered (TSB > 10)",
			summary: &storage.DailyTrainingLoadPoint{
				TSB: 15,
				CTL: 50,
				ATL: 35,
			},
			series:         nil,
			wantInsightIDs: []string{"well_recovered"},
		},
		{
			name: "building fitness (TSB between -10 and 10, CTL > 0)",
			summary: &storage.DailyTrainingLoadPoint{
				TSB: 0,
				CTL: 45,
				ATL: 45,
			},
			series:         nil,
			wantInsightIDs: []string{"building_fitness"},
		},
		{
			name: "CTL drop detection (>10% in 7 days)",
			summary: &storage.DailyTrainingLoadPoint{
				TSB: 5,
				CTL: 40,
				ATL: 35,
			},
			series: []storage.DailyTrainingLoadPoint{
				{Day: "2024-01-01", CTL: 50}, // Start (higher)
				{Day: "2024-01-02", CTL: 48},
				{Day: "2024-01-03", CTL: 46},
				{Day: "2024-01-04", CTL: 44},
				{Day: "2024-01-05", CTL: 42},
				{Day: "2024-01-06", CTL: 41},
				{Day: "2024-01-07", CTL: 40}, // End (dropped 20%)
			},
			wantInsightIDs: []string{"ctl_drop", "building_fitness"},
		},
		{
			name: "multiple warnings (high fatigue + overtraining)",
			summary: &storage.DailyTrainingLoadPoint{
				TSB: -30, // High fatigue
				CTL: 30,
				ATL: 60, // ATL > CTL * 1.5
			},
			series:         nil,
			wantInsightIDs: []string{"high_fatigue", "overtraining_risk"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insights := svc.applyTrainingLoadRules(tt.summary, tt.series)

			// Build a set of returned insight IDs.
			gotIDs := make(map[string]bool)
			for _, insight := range insights {
				gotIDs[insight.ID] = true
			}

			// Check that all expected insights are present.
			for _, wantID := range tt.wantInsightIDs {
				if !gotIDs[wantID] {
					t.Errorf("expected insight %q not found in results: %v", wantID, insights)
				}
			}

			// Check that no unexpected insights are present.
			for _, insight := range insights {
				found := false
				for _, wantID := range tt.wantInsightIDs {
					if insight.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("unexpected insight %q in results", insight.ID)
				}
			}
		})
	}
}

func TestInsightsService_SeverityPriority(t *testing.T) {
	// Verify severity priority ordering.
	if severityPriority["warning"] >= severityPriority["success"] {
		t.Error("warning should have higher priority (lower number) than success")
	}
	if severityPriority["success"] >= severityPriority["info"] {
		t.Error("success should have higher priority (lower number) than info")
	}
}

func TestInsightsService_InsightStructure(t *testing.T) {
	svc := &InsightsService{}

	// Test that insights have all required fields populated.
	summary := &storage.DailyTrainingLoadPoint{
		TSB: -25,
		CTL: 50,
		ATL: 75,
	}

	insights := svc.applyTrainingLoadRules(summary, nil)

	for _, insight := range insights {
		if insight.ID == "" {
			t.Error("Insight ID should not be empty")
		}
		if insight.Type == "" {
			t.Error("Insight Type should not be empty")
		}
		if insight.Severity == "" {
			t.Error("Insight Severity should not be empty")
		}
		if insight.Title == "" {
			t.Error("Insight Title should not be empty")
		}
		if insight.Description == "" {
			t.Error("Insight Description should not be empty")
		}

		// Verify severity is valid.
		if _, ok := severityPriority[insight.Severity]; !ok {
			t.Errorf("Invalid severity %q", insight.Severity)
		}
	}
}

func TestInsightsService_ApplyCTLInsights(t *testing.T) {
	svc := &InsightsService{}

	tests := []struct {
		name           string
		series         []storage.DailyTrainingLoadPoint
		wantInsightIDs []string
	}{
		{
			name:           "insufficient data",
			series:         []storage.DailyTrainingLoadPoint{{Day: "2024-01-01", CTL: 50}},
			wantInsightIDs: nil,
		},
		{
			name: "fitness peak - CTL at 90-day high",
			series: func() []storage.DailyTrainingLoadPoint {
				// Slow increase: 40 → 44 over 10 days = 10% over whole period.
				// Last 7 days: 41 → 44 = ~7% (below 15% threshold).
				s := make([]storage.DailyTrainingLoadPoint, 10)
				for i := range s {
					s[i] = storage.DailyTrainingLoadPoint{Day: "2024-01-0" + string(rune('1'+i)), CTL: 40 + float64(i)*0.4}
				}
				return s
			}(),
			wantInsightIDs: []string{"fitness_peak"},
		},
		{
			name: "ramp too fast - CTL increased >15% in 7 days",
			series: func() []storage.DailyTrainingLoadPoint {
				// Fast increase: 40 → 50 over 10 days.
				// Last 7 days: series[3]=42 → series[9]=50 = 19% increase (> 15%).
				s := make([]storage.DailyTrainingLoadPoint, 10)
				for i := range s {
					s[i] = storage.DailyTrainingLoadPoint{Day: "2024-01-0" + string(rune('1'+i)), CTL: 40 + float64(i)*1.1}
				}
				return s
			}(),
			wantInsightIDs: []string{"fitness_peak", "ramp_too_fast"},
		},
		{
			name: "no fitness peak when CTL is decreasing",
			series: func() []storage.DailyTrainingLoadPoint {
				s := make([]storage.DailyTrainingLoadPoint, 10)
				for i := range s {
					s[i] = storage.DailyTrainingLoadPoint{Day: "2024-01-0" + string(rune('1'+i)), CTL: float64(50 - i)}
				}
				return s
			}(),
			wantInsightIDs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insights := svc.applyCTLInsights(tt.series)

			gotIDs := make(map[string]bool)
			for _, insight := range insights {
				gotIDs[insight.ID] = true
			}

			for _, wantID := range tt.wantInsightIDs {
				if !gotIDs[wantID] {
					t.Errorf("expected insight %q not found in results: %v", wantID, insights)
				}
			}

			for _, insight := range insights {
				found := false
				for _, wantID := range tt.wantInsightIDs {
					if insight.ID == wantID {
						found = true
						break
					}
				}
				if !found && len(tt.wantInsightIDs) > 0 {
					t.Errorf("unexpected insight %q in results", insight.ID)
				}
			}
		})
	}
}

func TestInsightsService_ApplyPowerPRInsight(t *testing.T) {
	svc := &InsightsService{}

	tests := []struct {
		name           string
		prs            []storage.PeakPowerBest
		wantInsightIDs []string
	}{
		{
			name:           "no PRs",
			prs:            nil,
			wantInsightIDs: nil,
		},
		{
			name: "single PR",
			prs: []storage.PeakPowerBest{
				{DurationS: 300, Watts: 350},
			},
			wantInsightIDs: []string{"power_pr"},
		},
		{
			name: "multiple PRs - prefers longer duration",
			prs: []storage.PeakPowerBest{
				{DurationS: 5, Watts: 800},
				{DurationS: 60, Watts: 450},
				{DurationS: 1200, Watts: 280},
			},
			wantInsightIDs: []string{"power_pr"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insights := svc.applyPowerPRInsight(tt.prs)

			if len(tt.wantInsightIDs) == 0 && len(insights) > 0 {
				t.Errorf("expected no insights, got %v", insights)
			}

			for _, wantID := range tt.wantInsightIDs {
				found := false
				for _, insight := range insights {
					if insight.ID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected insight %q not found", wantID)
				}
			}
		})
	}
}

func TestInsightsService_ApplyActivityPatternInsights(t *testing.T) {
	svc := &InsightsService{}

	// Helper to create activities with dates.
	makeActivities := func(daysAgo []int, sport string) []storage.Activity {
		activities := make([]storage.Activity, len(daysAgo))
		now := timeNow()
		for i, d := range daysAgo {
			activities[i] = storage.Activity{
				StartDateLocal: storage.SQLiteTime{Time: now.AddDate(0, 0, -d)},
				SportType:      sport,
			}
		}
		return activities
	}

	tests := []struct {
		name           string
		activities     []storage.Activity
		wantInsightIDs []string
	}{
		{
			name:           "no activities",
			activities:     nil,
			wantInsightIDs: nil,
		},
		{
			name:           "rest day needed - 7+ consecutive days",
			activities:     makeActivities([]int{0, 1, 2, 3, 4, 5, 6, 7}, "Ride"),
			wantInsightIDs: []string{"rest_day_needed"},
		},
		{
			name: "variety check - 10+ same sport",
			activities: func() []storage.Activity {
				acts := make([]storage.Activity, 12)
				now := timeNow()
				for i := range acts {
					acts[i] = storage.Activity{
						StartDateLocal: storage.SQLiteTime{Time: now.AddDate(0, 0, -i*2)},
						SportType:      "Ride",
					}
				}
				return acts
			}(),
			wantInsightIDs: []string{"variety_check"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			insights := svc.applyActivityPatternInsights(tt.activities)

			gotIDs := make(map[string]bool)
			for _, insight := range insights {
				gotIDs[insight.ID] = true
			}

			for _, wantID := range tt.wantInsightIDs {
				if !gotIDs[wantID] {
					t.Errorf("expected insight %q not found in results: %v", wantID, insights)
				}
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{5, "5s"},
		{30, "30s"},
		{60, "1m"},
		{300, "5m"},
		{3600, "1h"},
		{7200, "2h"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.seconds)
		if got != tt.want {
			t.Errorf("formatDuration(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestCountConsecutiveActivityDays(t *testing.T) {
	// Note: This test uses dates relative to the function's time.Now().
	// In production, the function checks consecutive days ending at today.

	// Create dates map with consecutive days from today.
	dates := make(map[string]bool)
	today := time.Now()
	for i := 0; i < 5; i++ {
		dates[today.AddDate(0, 0, -i).Format("2006-01-02")] = true
	}

	count := countConsecutiveActivityDays(dates)
	if count != 5 {
		t.Errorf("countConsecutiveActivityDays = %d, want 5", count)
	}
}

func TestCountWeeksWithActivity(t *testing.T) {
	dates := map[string]bool{
		"2024-01-01": true, // Week 1
		"2024-01-08": true, // Week 2
		"2024-01-09": true, // Week 2 (same week)
		"2024-01-15": true, // Week 3
		"2024-01-22": true, // Week 4
	}

	count := countWeeksWithActivity(dates)
	if count != 4 {
		t.Errorf("countWeeksWithActivity = %d, want 4", count)
	}
}

// timeNow returns current time wrapped in SQLiteTime for test helpers.
func timeNow() time.Time {
	return time.Now()
}
