package services

import (
	"testing"

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
