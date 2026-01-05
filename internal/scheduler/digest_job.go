package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/melonamin/quantlete/internal/notifications"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// DigestPeriod represents the time period for a digest.
type DigestPeriod string

const (
	DigestPeriodWeekly  DigestPeriod = "weekly"
	DigestPeriodMonthly DigestPeriod = "monthly"
)

// DigestJob sends activity digest notifications for a configurable period.
type DigestJob struct {
	logger        *slog.Logger
	stats         *storage.StatsRepository
	notifications *services.NotificationService
	settings      *storage.SettingsRepository
	athleteID     int64
	period        DigestPeriod
}

// NewDigestJob creates a new digest job for the specified period.
func NewDigestJob(
	logger *slog.Logger,
	stats *storage.StatsRepository,
	notificationSvc *services.NotificationService,
	settings *storage.SettingsRepository,
	athleteID int64,
	period DigestPeriod,
) *DigestJob {
	return &DigestJob{
		logger:        logger,
		stats:         stats,
		notifications: notificationSvc,
		settings:      settings,
		athleteID:     athleteID,
		period:        period,
	}
}

// Run executes the digest job (called by cron).
func (j *DigestJob) Run() {
	ctx := context.Background()

	// Check if this digest type is enabled in settings
	athleteSettings, err := j.settings.Get(ctx, j.athleteID)
	if err != nil {
		j.logger.Error(string(j.period)+" digest: failed to get settings",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	if !j.isEnabled(athleteSettings.Notifications) {
		j.logger.Debug(string(j.period)+" digest: notification disabled",
			slog.Int64("athlete_id", j.athleteID),
		)
		return
	}

	// Calculate date range based on period
	startDate, endDate := j.calculateDateRange()

	stats, err := j.stats.GetDigestStats(ctx, j.athleteID, startDate, endDate)
	if err != nil {
		j.logger.Error(string(j.period)+" digest: failed to get stats",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	// Send notification
	if err := j.sendNotification(ctx, stats); err != nil {
		j.logger.Error(string(j.period)+" digest: failed to send notification",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	j.logger.Info(string(j.period)+" digest: notification sent",
		slog.Int64("athlete_id", j.athleteID),
		slog.Int("activities", stats.ActivityCount),
	)
}

// isEnabled checks if this digest type is enabled in the notification config.
func (j *DigestJob) isEnabled(config *notifications.NotificationConfig) bool {
	if config == nil || !config.Enabled {
		return false
	}

	switch j.period {
	case DigestPeriodWeekly:
		return config.Events.WeeklyDigest
	case DigestPeriodMonthly:
		return config.Events.MonthlyDigest
	default:
		return false
	}
}

// calculateDateRange returns the start and end dates for this digest period.
// Dates are truncated to midnight to ensure clean day boundaries.
func (j *DigestJob) calculateDateRange() (start, end time.Time) {
	now := time.Now()
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	switch j.period {
	case DigestPeriodWeekly:
		// Past 7 full days: from 7 days ago at midnight through end of yesterday.
		// Example: if today is Monday, covers Mon-Sun (previous week).
		start = today.AddDate(0, 0, -7)
		end = today // Exclusive end - activities with start_date < today
		return start, end

	case DigestPeriodMonthly:
		// Previous calendar month (full month)
		firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		firstOfPreviousMonth := firstOfCurrentMonth.AddDate(0, -1, 0)
		return firstOfPreviousMonth, firstOfCurrentMonth // Exclusive end

	default:
		// Fallback to weekly
		start = today.AddDate(0, 0, -7)
		return start, today
	}
}

// sendNotification sends the digest notification using the appropriate service method.
func (j *DigestJob) sendNotification(ctx context.Context, stats *storage.DigestStats) error {
	switch j.period {
	case DigestPeriodWeekly:
		return j.notifications.NotifyWeeklyDigest(ctx, j.athleteID, *stats)
	case DigestPeriodMonthly:
		return j.notifications.NotifyMonthlyDigest(ctx, j.athleteID, *stats)
	default:
		return j.notifications.NotifyWeeklyDigest(ctx, j.athleteID, *stats)
	}
}

// WeeklyDigestJob is a convenience wrapper for weekly digests.
//
// Deprecated: Use NewDigestJob with DigestPeriodWeekly instead.
type WeeklyDigestJob = DigestJob

// NewWeeklyDigestJob creates a new weekly digest job.
func NewWeeklyDigestJob(
	logger *slog.Logger,
	stats *storage.StatsRepository,
	notificationSvc *services.NotificationService,
	settings *storage.SettingsRepository,
	athleteID int64,
) *DigestJob {
	return NewDigestJob(logger, stats, notificationSvc, settings, athleteID, DigestPeriodWeekly)
}

// MonthlyDigestJob is a convenience wrapper for monthly digests.
//
// Deprecated: Use NewDigestJob with DigestPeriodMonthly instead.
type MonthlyDigestJob = DigestJob

// NewMonthlyDigestJob creates a new monthly digest job.
func NewMonthlyDigestJob(
	logger *slog.Logger,
	stats *storage.StatsRepository,
	notificationSvc *services.NotificationService,
	settings *storage.SettingsRepository,
	athleteID int64,
) *DigestJob {
	return NewDigestJob(logger, stats, notificationSvc, settings, athleteID, DigestPeriodMonthly)
}
