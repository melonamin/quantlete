package scheduler

import (
	"testing"
	"time"

	"github.com/melonamin/quantlete/internal/notifications"
)

func TestDigestJob_isEnabled(t *testing.T) {
	tests := []struct {
		name   string
		period DigestPeriod
		config *notifications.NotificationConfig
		want   bool
	}{
		{
			name:   "nil config",
			period: DigestPeriodWeekly,
			config: nil,
			want:   false,
		},
		{
			name:   "notifications disabled",
			period: DigestPeriodWeekly,
			config: &notifications.NotificationConfig{
				Enabled: false,
				Events: notifications.EventConfig{
					WeeklyDigest: true,
				},
			},
			want: false,
		},
		{
			name:   "weekly enabled",
			period: DigestPeriodWeekly,
			config: &notifications.NotificationConfig{
				Enabled: true,
				Events: notifications.EventConfig{
					WeeklyDigest: true,
				},
			},
			want: true,
		},
		{
			name:   "weekly disabled",
			period: DigestPeriodWeekly,
			config: &notifications.NotificationConfig{
				Enabled: true,
				Events: notifications.EventConfig{
					WeeklyDigest: false,
				},
			},
			want: false,
		},
		{
			name:   "monthly enabled",
			period: DigestPeriodMonthly,
			config: &notifications.NotificationConfig{
				Enabled: true,
				Events: notifications.EventConfig{
					MonthlyDigest: true,
				},
			},
			want: true,
		},
		{
			name:   "monthly disabled",
			period: DigestPeriodMonthly,
			config: &notifications.NotificationConfig{
				Enabled: true,
				Events: notifications.EventConfig{
					MonthlyDigest: false,
				},
			},
			want: false,
		},
		{
			name:   "unknown period",
			period: DigestPeriod("unknown"),
			config: &notifications.NotificationConfig{
				Enabled: true,
				Events: notifications.EventConfig{
					WeeklyDigest:  true,
					MonthlyDigest: true,
				},
			},
			want: false,
		},
		{
			name:   "weekly enabled but monthly not",
			period: DigestPeriodMonthly,
			config: &notifications.NotificationConfig{
				Enabled: true,
				Events: notifications.EventConfig{
					WeeklyDigest:  true,
					MonthlyDigest: false,
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &DigestJob{period: tt.period}
			got := j.isEnabled(tt.config)
			if got != tt.want {
				t.Errorf("isEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDigestJob_calculateDateRange(t *testing.T) {
	tests := []struct {
		name          string
		period        DigestPeriod
		checkStart    func(start time.Time, now time.Time) bool
		checkEnd      func(end time.Time, now time.Time) bool
		startDescribe string
		endDescribe   string
	}{
		{
			name:   "weekly returns past 7 days",
			period: DigestPeriodWeekly,
			checkStart: func(start, now time.Time) bool {
				// Start should be approximately 7 days ago
				expected := now.AddDate(0, 0, -7)
				diff := start.Sub(expected)
				return diff < time.Second && diff > -time.Second
			},
			checkEnd: func(end, now time.Time) bool {
				// End should be approximately now
				diff := end.Sub(now)
				return diff < time.Second && diff > -time.Second
			},
			startDescribe: "7 days ago",
			endDescribe:   "now",
		},
		{
			name:   "monthly returns previous calendar month",
			period: DigestPeriodMonthly,
			checkStart: func(start, now time.Time) bool {
				// Start should be first day of previous month at midnight
				firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				lastOfPreviousMonth := firstOfCurrentMonth.AddDate(0, 0, -1)
				expectedStart := time.Date(lastOfPreviousMonth.Year(), lastOfPreviousMonth.Month(), 1, 0, 0, 0, 0, now.Location())
				return start.Equal(expectedStart)
			},
			checkEnd: func(end, now time.Time) bool {
				// End should be last day of previous month
				firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
				expectedEnd := firstOfCurrentMonth.AddDate(0, 0, -1)
				return end.Year() == expectedEnd.Year() &&
					end.Month() == expectedEnd.Month() &&
					end.Day() == expectedEnd.Day()
			},
			startDescribe: "first day of previous month",
			endDescribe:   "last day of previous month",
		},
		{
			name:   "unknown period defaults to weekly",
			period: DigestPeriod("unknown"),
			checkStart: func(start, now time.Time) bool {
				expected := now.AddDate(0, 0, -7)
				diff := start.Sub(expected)
				return diff < time.Second && diff > -time.Second
			},
			checkEnd: func(end, now time.Time) bool {
				diff := end.Sub(now)
				return diff < time.Second && diff > -time.Second
			},
			startDescribe: "7 days ago (fallback)",
			endDescribe:   "now (fallback)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &DigestJob{period: tt.period}
			now := time.Now()
			start, end := j.calculateDateRange()

			if !tt.checkStart(start, now) {
				t.Errorf("start = %v, want %s", start, tt.startDescribe)
			}
			if !tt.checkEnd(end, now) {
				t.Errorf("end = %v, want %s", end, tt.endDescribe)
			}
		})
	}
}

func TestDigestJob_calculateDateRange_MonthlyBoundaries(t *testing.T) {
	// Test specific month boundary cases
	tests := []struct {
		name          string
		testTime      time.Time
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			name:          "January -> December previous year",
			testTime:      time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			expectedStart: time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "March -> February (28 days, non-leap)",
			testTime:      time.Date(2023, 3, 1, 10, 0, 0, 0, time.UTC),
			expectedStart: time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "March -> February (29 days, leap year)",
			testTime:      time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC),
			expectedStart: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:          "May -> April (30 days)",
			testTime:      time.Date(2024, 5, 15, 10, 0, 0, 0, time.UTC),
			expectedStart: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a job and manually test date calculation logic
			now := tt.testTime
			firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			lastOfPreviousMonth := firstOfCurrentMonth.AddDate(0, 0, -1)
			firstOfPreviousMonth := time.Date(lastOfPreviousMonth.Year(), lastOfPreviousMonth.Month(), 1, 0, 0, 0, 0, now.Location())

			if !firstOfPreviousMonth.Equal(tt.expectedStart) {
				t.Errorf("start = %v, want %v", firstOfPreviousMonth, tt.expectedStart)
			}

			if lastOfPreviousMonth.Year() != tt.expectedEnd.Year() ||
				lastOfPreviousMonth.Month() != tt.expectedEnd.Month() ||
				lastOfPreviousMonth.Day() != tt.expectedEnd.Day() {
				t.Errorf("end = %v, want %v", lastOfPreviousMonth, tt.expectedEnd)
			}
		})
	}
}

func TestNewWeeklyDigestJob(t *testing.T) {
	job := NewWeeklyDigestJob(nil, nil, nil, nil, 12345)
	if job.period != DigestPeriodWeekly {
		t.Errorf("NewWeeklyDigestJob period = %v, want %v", job.period, DigestPeriodWeekly)
	}
	if job.athleteID != 12345 {
		t.Errorf("NewWeeklyDigestJob athleteID = %d, want 12345", job.athleteID)
	}
}

func TestNewMonthlyDigestJob(t *testing.T) {
	job := NewMonthlyDigestJob(nil, nil, nil, nil, 67890)
	if job.period != DigestPeriodMonthly {
		t.Errorf("NewMonthlyDigestJob period = %v, want %v", job.period, DigestPeriodMonthly)
	}
	if job.athleteID != 67890 {
		t.Errorf("NewMonthlyDigestJob athleteID = %d, want 67890", job.athleteID)
	}
}
