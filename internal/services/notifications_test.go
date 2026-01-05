package services

import (
	"testing"

	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/notifications"
)

func TestFilterAchievements(t *testing.T) {
	s := &NotificationService{}

	tests := []struct {
		name         string
		achievements []Achievement
		events       notifications.EventConfig
		wantCount    int
		wantTypes    []string
	}{
		{
			name: "filters personal records when enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementPersonalRecord), Title: "5K PR"},
				{Type: string(importer.AchievementSegmentPR), Title: "Hawk Hill"},
			},
			events:    notifications.EventConfig{PersonalRecords: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementPersonalRecord)},
		},
		{
			name: "filters eddington increase",
			achievements: []Achievement{
				{Type: string(importer.AchievementEddingtonIncrease), Title: "Eddington 51"},
			},
			events:    notifications.EventConfig{EddingtonIncrease: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementEddingtonIncrease)},
		},
		{
			name: "filters gear milestones when enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementGearMilestone), Title: "Canyon Aeroad hit 10,000 km"},
			},
			events:    notifications.EventConfig{GearMilestones: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementGearMilestone)},
		},
		{
			name: "filters training load alerts when fatigue warning enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypeFatigue, Title: "High Fatigue"},
			},
			events:    notifications.EventConfig{FatigueWarning: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementTrainingLoadAlert)},
		},
		{
			name: "filters training load alerts when recovery enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypePeakForm, Title: "Peak Form"},
			},
			events:    notifications.EventConfig{RecoveryAlert: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementTrainingLoadAlert)},
		},
		{
			name: "filters training load alerts when overtraining risk enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypeOvertraining, Title: "Overtraining Warning"},
			},
			events:    notifications.EventConfig{OvertrainingRisk: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementTrainingLoadAlert)},
		},
		{
			name: "excludes fatigue alert when only recovery enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypeFatigue, Title: "High Fatigue"},
			},
			events:    notifications.EventConfig{RecoveryAlert: true},
			wantCount: 0,
			wantTypes: nil,
		},
		{
			name: "excludes recovery alert when only fatigue enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypePeakForm, Title: "Peak Form"},
			},
			events:    notifications.EventConfig{FatigueWarning: true},
			wantCount: 0,
			wantTypes: nil,
		},
		{
			name: "excludes overtraining alert when only fatigue enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypeOvertraining, Title: "Overtraining Warning"},
			},
			events:    notifications.EventConfig{FatigueWarning: true},
			wantCount: 0,
			wantTypes: nil,
		},
		{
			name: "includes only matching training load alerts",
			achievements: []Achievement{
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypeFatigue, Title: "High Fatigue"},
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypePeakForm, Title: "Peak Form"},
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypeOvertraining, Title: "Overtraining Warning"},
			},
			events:    notifications.EventConfig{FatigueWarning: true, OvertrainingRisk: true},
			wantCount: 2,
			wantTypes: []string{string(importer.AchievementTrainingLoadAlert), string(importer.AchievementTrainingLoadAlert)},
		},
		{
			name: "drops all when nothing enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementPersonalRecord), Title: "5K PR"},
				{Type: string(importer.AchievementEddingtonIncrease), Title: "Eddington 51"},
				{Type: string(importer.AchievementGearMilestone), Title: "10,000 km"},
				{Type: string(importer.AchievementTrainingLoadAlert), SubType: importer.AlertSubTypeFatigue, Title: "High Fatigue"},
			},
			events:    notifications.EventConfig{},
			wantCount: 0,
			wantTypes: nil,
		},
		{
			name: "includes multiple types when enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementPersonalRecord), Title: "5K PR"},
				{Type: string(importer.AchievementSegmentPR), Title: "Hawk Hill"},
				{Type: string(importer.AchievementPowerRecord), Title: "20min power"},
			},
			events:    notifications.EventConfig{PersonalRecords: true, SegmentPRs: true, PowerRecords: true},
			wantCount: 3,
			wantTypes: []string{string(importer.AchievementPersonalRecord), string(importer.AchievementSegmentPR), string(importer.AchievementPowerRecord)},
		},
		{
			name: "filters goal complete when enabled",
			achievements: []Achievement{
				{Type: string(importer.AchievementGoalComplete), Title: "Monthly goal reached"},
			},
			events:    notifications.EventConfig{GoalComplete: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementGoalComplete)},
		},
		{
			name:         "empty achievements returns empty slice",
			achievements: []Achievement{},
			events:       notifications.EventConfig{PersonalRecords: true},
			wantCount:    0,
			wantTypes:    nil,
		},
		{
			name:         "nil achievements returns empty slice",
			achievements: nil,
			events:       notifications.EventConfig{PersonalRecords: true},
			wantCount:    0,
			wantTypes:    nil,
		},
		{
			name: "unknown achievement type is ignored",
			achievements: []Achievement{
				{Type: "unknown_type", Title: "Unknown"},
				{Type: string(importer.AchievementPersonalRecord), Title: "5K PR"},
			},
			events:    notifications.EventConfig{PersonalRecords: true},
			wantCount: 1,
			wantTypes: []string{string(importer.AchievementPersonalRecord)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.filterAchievements(tt.achievements, tt.events)

			if len(got) != tt.wantCount {
				t.Errorf("filterAchievements() returned %d achievements, want %d", len(got), tt.wantCount)
			}

			for i, a := range got {
				if i < len(tt.wantTypes) && a.Type != tt.wantTypes[i] {
					t.Errorf("filterAchievements()[%d].Type = %q, want %q", i, a.Type, tt.wantTypes[i])
				}
			}
		})
	}
}

func TestFormatAchievementsMessage(t *testing.T) {
	tests := []struct {
		name         string
		achievements []Achievement
		want         string
	}{
		{
			name: "single achievement without previous value",
			achievements: []Achievement{
				{Type: string(importer.AchievementPersonalRecord), Title: "5K PR", Value: "18:30"},
			},
			want: "Achievements:\n- 5K PR: 18:30\n",
		},
		{
			name: "single achievement with previous value",
			achievements: []Achievement{
				{Type: string(importer.AchievementPersonalRecord), Title: "5K PR", Value: "18:30", PreviousValue: "19:00"},
			},
			want: "Achievements:\n- 5K PR: 18:30 (was 19:00)\n",
		},
		{
			name: "multiple achievements mixed",
			achievements: []Achievement{
				{Type: string(importer.AchievementPersonalRecord), Title: "5K PR", Value: "18:30", PreviousValue: "19:00"},
				{Type: string(importer.AchievementEddingtonIncrease), Title: "Eddington Number", Value: "52"},
				{Type: string(importer.AchievementGearMilestone), Title: "Canyon Aeroad", Value: "10,000 km"},
			},
			want: "Achievements:\n- 5K PR: 18:30 (was 19:00)\n- Eddington Number: 52\n- Canyon Aeroad: 10,000 km\n",
		},
		{
			name:         "empty achievements",
			achievements: []Achievement{},
			want:         "Achievements:\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAchievementsMessage(tt.achievements)
			if got != tt.want {
				t.Errorf("formatAchievementsMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTrainingLoadMessage(t *testing.T) {
	tests := []struct {
		name  string
		alert TrainingLoadAlert
		want  string
	}{
		{
			name:  "fatigue warning",
			alert: TrainingLoadAlert{Type: AlertTypeFatigueWarning, TSB: -25, ATL: 80, CTL: 55},
			want:  "High fatigue (TSB: -25). Consider rest.",
		},
		{
			name:  "recovery alert",
			alert: TrainingLoadAlert{Type: AlertTypeRecoveryAlert, TSB: 15, ATL: 40, CTL: 55},
			want:  "Recovered! TSB: +15. Ready to push.",
		},
		{
			name:  "overtraining risk",
			alert: TrainingLoadAlert{Type: AlertTypeOvertrainingRisk, TSB: -30, ATL: 120, CTL: 60},
			want:  "Overtraining risk: ATL (120) >> CTL (60)",
		},
		{
			name:  "unknown type falls back to default",
			alert: TrainingLoadAlert{Type: "unknown", TSB: 10, ATL: 50, CTL: 40},
			want:  "Training load alert: TSB=10, ATL=50, CTL=40",
		},
		{
			name:  "empty type falls back to default",
			alert: TrainingLoadAlert{Type: "", TSB: 5, ATL: 45, CTL: 40},
			want:  "Training load alert: TSB=5, ATL=45, CTL=40",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTrainingLoadMessage(tt.alert)
			if got != tt.want {
				t.Errorf("formatTrainingLoadMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}
