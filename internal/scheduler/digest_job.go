package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// WeeklyDigestJob sends weekly activity digest notifications.
type WeeklyDigestJob struct {
	logger        *slog.Logger
	stats         *storage.StatsRepository
	notifications *services.NotificationService
	settings      *storage.SettingsRepository
	athleteID     int64
}

// NewWeeklyDigestJob creates a new weekly digest job.
func NewWeeklyDigestJob(
	logger *slog.Logger,
	stats *storage.StatsRepository,
	notifications *services.NotificationService,
	settings *storage.SettingsRepository,
	athleteID int64,
) *WeeklyDigestJob {
	return &WeeklyDigestJob{
		logger:        logger,
		stats:         stats,
		notifications: notifications,
		settings:      settings,
		athleteID:     athleteID,
	}
}

// Run executes the weekly digest job (called by cron).
func (j *WeeklyDigestJob) Run() {
	ctx := context.Background()

	// Check if weekly digest is enabled in settings
	athleteSettings, err := j.settings.Get(ctx, j.athleteID)
	if err != nil {
		j.logger.Error("weekly digest: failed to get settings",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	if athleteSettings.Notifications == nil ||
		!athleteSettings.Notifications.Enabled ||
		!athleteSettings.Notifications.Events.WeeklyDigest {
		j.logger.Debug("weekly digest: notification disabled",
			slog.Int64("athlete_id", j.athleteID),
		)
		return
	}

	// Get stats for the past 7 days
	now := time.Now()
	startDate := now.AddDate(0, 0, -7)
	endDate := now

	storageStats, err := j.stats.GetDigestStats(ctx, j.athleteID, startDate, endDate)
	if err != nil {
		j.logger.Error("weekly digest: failed to get stats",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	// Convert storage stats to service stats
	stats := services.DigestStats{
		StartDate:      storageStats.StartDate,
		EndDate:        storageStats.EndDate,
		ActivityCount:  storageStats.ActivityCount,
		TotalDistance:  storageStats.TotalDistance,
		TotalTime:      storageStats.TotalTime,
		TotalElevation: storageStats.TotalElevation,
		TotalCalories:  storageStats.TotalCalories,
	}

	// Send notification
	if err := j.notifications.NotifyWeeklyDigest(ctx, j.athleteID, stats); err != nil {
		j.logger.Error("weekly digest: failed to send notification",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	j.logger.Info("weekly digest: notification sent",
		slog.Int64("athlete_id", j.athleteID),
		slog.Int("activities", stats.ActivityCount),
	)
}

// MonthlyDigestJob sends monthly activity digest notifications.
type MonthlyDigestJob struct {
	logger        *slog.Logger
	stats         *storage.StatsRepository
	notifications *services.NotificationService
	settings      *storage.SettingsRepository
	athleteID     int64
}

// NewMonthlyDigestJob creates a new monthly digest job.
func NewMonthlyDigestJob(
	logger *slog.Logger,
	stats *storage.StatsRepository,
	notifications *services.NotificationService,
	settings *storage.SettingsRepository,
	athleteID int64,
) *MonthlyDigestJob {
	return &MonthlyDigestJob{
		logger:        logger,
		stats:         stats,
		notifications: notifications,
		settings:      settings,
		athleteID:     athleteID,
	}
}

// Run executes the monthly digest job (called by cron).
func (j *MonthlyDigestJob) Run() {
	ctx := context.Background()

	// Check if monthly digest is enabled in settings
	athleteSettings, err := j.settings.Get(ctx, j.athleteID)
	if err != nil {
		j.logger.Error("monthly digest: failed to get settings",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	if athleteSettings.Notifications == nil ||
		!athleteSettings.Notifications.Enabled ||
		!athleteSettings.Notifications.Events.MonthlyDigest {
		j.logger.Debug("monthly digest: notification disabled",
			slog.Int64("athlete_id", j.athleteID),
		)
		return
	}

	// Get stats for the previous month
	now := time.Now()
	// Go back to the first day of the current month, then back one more day to get to previous month
	firstOfCurrentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	lastOfPreviousMonth := firstOfCurrentMonth.AddDate(0, 0, -1)
	firstOfPreviousMonth := time.Date(lastOfPreviousMonth.Year(), lastOfPreviousMonth.Month(), 1, 0, 0, 0, 0, now.Location())

	storageStats, err := j.stats.GetDigestStats(ctx, j.athleteID, firstOfPreviousMonth, lastOfPreviousMonth)
	if err != nil {
		j.logger.Error("monthly digest: failed to get stats",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	// Convert storage stats to service stats
	stats := services.DigestStats{
		StartDate:      storageStats.StartDate,
		EndDate:        storageStats.EndDate,
		ActivityCount:  storageStats.ActivityCount,
		TotalDistance:  storageStats.TotalDistance,
		TotalTime:      storageStats.TotalTime,
		TotalElevation: storageStats.TotalElevation,
		TotalCalories:  storageStats.TotalCalories,
	}

	// Send notification
	if err := j.notifications.NotifyMonthlyDigest(ctx, j.athleteID, stats); err != nil {
		j.logger.Error("monthly digest: failed to send notification",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	j.logger.Info("monthly digest: notification sent",
		slog.Int64("athlete_id", j.athleteID),
		slog.Int("activities", stats.ActivityCount),
	)
}
