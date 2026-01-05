package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/notifications"
	"github.com/melonamin/quantlete/internal/storage"
)

// Training load alert type constants.
const (
	AlertTypeFatigueWarning   = "fatigue_warning"
	AlertTypeRecoveryAlert    = "recovery_alert"
	AlertTypeOvertrainingRisk = "overtraining_risk"
)

// SettingsRepository defines the interface for accessing athlete settings.
type SettingsRepository interface {
	Get(ctx context.Context, athleteID int64) (*storage.AthleteSettings, error)
}

// NotificationService handles notification business logic.
type NotificationService struct {
	logger   *slog.Logger
	settings SettingsRepository
}

// NewNotificationService creates a new notification service.
func NewNotificationService(logger *slog.Logger, settings SettingsRepository) *NotificationService {
	return &NotificationService{
		logger:   logger,
		settings: settings,
	}
}

// ImportStats contains statistics about a completed import.
type ImportStats struct {
	ActivitiesImported int
	ActivitiesUpdated  int
	Duration           time.Duration
}

// MaintenanceItem represents a maintenance item that is due.
type MaintenanceItem struct {
	ComponentID    int64
	ComponentName  string
	GearName       string
	RuleType       string
	CurrentValue   float64
	ThresholdValue float64
}

// Achievement represents a single achievement from a sync.
type Achievement struct {
	Type          string // One of importer.AchievementType* constants
	SubType       string // Used for training load alert subtypes (importer.AlertSubType* constants)
	Title         string
	Value         string
	PreviousValue string
}

// TrainingLoadAlert represents a training load status alert.
type TrainingLoadAlert struct {
	Type string  // One of AlertType* constants
	TSB  float64 // Training Stress Balance
	ATL  float64 // Acute Training Load
	CTL  float64 // Chronic Training Load
}

// GearMilestone represents a gear distance milestone.
type GearMilestone struct {
	GearName  string
	Distance  float64 // in meters
	Milestone int     // milestone in km: 5000, 10000, etc
}

// NotifyImportComplete sends notification when import finishes (if enabled).
func (s *NotificationService) NotifyImportComplete(ctx context.Context, athleteID int64, stats ImportStats) error {
	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil || !config.Enabled || !config.Events.ImportComplete {
		s.logger.Debug("import complete notification disabled",
			slog.Int64("athlete_id", athleteID),
		)
		return nil
	}

	sender, err := notifications.NewSender(s.logger, config.Services)
	if err != nil {
		return fmt.Errorf("creating notification sender: %w", err)
	}

	title := "Import Complete"
	message := fmt.Sprintf(
		"Strava import finished successfully.\n\n"+
			"Activities imported: %d\n"+
			"Activities updated: %d\n"+
			"Duration: %s",
		stats.ActivitiesImported,
		stats.ActivitiesUpdated,
		stats.Duration.Round(time.Second),
	)

	if err := sender.Send(ctx, title, message); err != nil {
		s.logger.Error("failed to send import complete notification",
			slog.Int64("athlete_id", athleteID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("sending notification: %w", err)
	}

	s.logger.Info("import complete notification sent",
		slog.Int64("athlete_id", athleteID),
		slog.Int("activities_imported", stats.ActivitiesImported),
	)

	return nil
}

// NotifyMaintenanceDue sends notification about pending maintenance items (if enabled).
func (s *NotificationService) NotifyMaintenanceDue(ctx context.Context, athleteID int64, items []MaintenanceItem) error {
	if len(items) == 0 {
		return nil
	}

	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil || !config.Enabled || !config.Events.MaintenanceDue {
		s.logger.Debug("maintenance due notification disabled",
			slog.Int64("athlete_id", athleteID),
		)
		return nil
	}

	sender, err := notifications.NewSender(s.logger, config.Services)
	if err != nil {
		return fmt.Errorf("creating notification sender: %w", err)
	}

	title := "Maintenance Due"
	message := fmt.Sprintf("You have %d maintenance item(s) due:\n\n", len(items))

	for _, item := range items {
		message += fmt.Sprintf("- %s (%s): %s threshold reached (%.0f/%.0f)\n",
			item.ComponentName,
			item.GearName,
			item.RuleType,
			item.CurrentValue,
			item.ThresholdValue,
		)
	}

	if err := sender.Send(ctx, title, message); err != nil {
		s.logger.Error("failed to send maintenance due notification",
			slog.Int64("athlete_id", athleteID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("sending notification: %w", err)
	}

	s.logger.Info("maintenance due notification sent",
		slog.Int64("athlete_id", athleteID),
		slog.Int("items_count", len(items)),
	)

	return nil
}

// NotifyAchievements sends batched notification for achievements from a sync (if enabled).
func (s *NotificationService) NotifyAchievements(ctx context.Context, athleteID int64, achievements []Achievement) error {
	if len(achievements) == 0 {
		return nil
	}

	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil || !config.Enabled {
		s.logger.Debug("achievements notification disabled (notifications off)",
			slog.Int64("athlete_id", athleteID),
		)
		return nil
	}

	// Filter achievements based on which event types are enabled
	filtered := s.filterAchievements(achievements, config.Events)
	if len(filtered) == 0 {
		s.logger.Debug("no achievements match enabled event types",
			slog.Int64("athlete_id", athleteID),
		)
		return nil
	}

	sender, err := notifications.NewSender(s.logger, config.Services)
	if err != nil {
		return fmt.Errorf("creating notification sender: %w", err)
	}

	title := "Achievements"
	message := formatAchievementsMessage(filtered)

	if err := sender.Send(ctx, title, message); err != nil {
		s.logger.Error("failed to send achievements notification",
			slog.Int64("athlete_id", athleteID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("sending notification: %w", err)
	}

	s.logger.Info("achievements notification sent",
		slog.Int64("athlete_id", athleteID),
		slog.Int("achievements_count", len(filtered)),
	)

	return nil
}

// filterAchievements filters achievements based on enabled event types.
func (s *NotificationService) filterAchievements(achievements []Achievement, events notifications.EventConfig) []Achievement {
	var filtered []Achievement
	for _, a := range achievements {
		switch a.Type {
		case string(importer.AchievementPersonalRecord):
			if events.PersonalRecords {
				filtered = append(filtered, a)
			}
		case string(importer.AchievementSegmentPR):
			if events.SegmentPRs {
				filtered = append(filtered, a)
			}
		case string(importer.AchievementEddingtonIncrease):
			if events.EddingtonIncrease {
				filtered = append(filtered, a)
			}
		case string(importer.AchievementPowerRecord):
			if events.PowerRecords {
				filtered = append(filtered, a)
			}
		case string(importer.AchievementGoalComplete):
			if events.GoalComplete {
				filtered = append(filtered, a)
			}
		case string(importer.AchievementGearMilestone):
			if events.GearMilestones {
				filtered = append(filtered, a)
			}
		case string(importer.AchievementTrainingLoadAlert):
			// Training load alerts check specific sub-types
			if shouldIncludeTrainingLoadAlert(a.SubType, events) {
				filtered = append(filtered, a)
			}
		}
	}
	return filtered
}

// shouldIncludeTrainingLoadAlert checks if a training load alert should be included
// based on its subtype and the enabled event types.
func shouldIncludeTrainingLoadAlert(subType string, events notifications.EventConfig) bool {
	switch subType {
	case importer.AlertSubTypeFatigue:
		return events.FatigueWarning
	case importer.AlertSubTypePeakForm:
		return events.RecoveryAlert
	case importer.AlertSubTypeOvertraining:
		return events.OvertrainingRisk
	default:
		return false
	}
}

// formatAchievementsMessage formats the achievements notification message.
func formatAchievementsMessage(achievements []Achievement) string {
	message := "Achievements:\n"
	for _, a := range achievements {
		if a.PreviousValue != "" {
			message += fmt.Sprintf("- %s: %s (was %s)\n", a.Title, a.Value, a.PreviousValue)
		} else {
			message += fmt.Sprintf("- %s: %s\n", a.Title, a.Value)
		}
	}
	return message
}

// NotifyTrainingLoad sends a training load status alert (if enabled).
func (s *NotificationService) NotifyTrainingLoad(ctx context.Context, athleteID int64, alert TrainingLoadAlert) error {
	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil || !config.Enabled {
		s.logger.Debug("training load notification disabled (notifications off)",
			slog.Int64("athlete_id", athleteID),
		)
		return nil
	}

	// Check if this specific alert type is enabled
	enabled := false
	switch alert.Type {
	case AlertTypeFatigueWarning:
		enabled = config.Events.FatigueWarning
	case AlertTypeRecoveryAlert:
		enabled = config.Events.RecoveryAlert
	case AlertTypeOvertrainingRisk:
		enabled = config.Events.OvertrainingRisk
	}

	if !enabled {
		s.logger.Debug("training load notification disabled for type",
			slog.Int64("athlete_id", athleteID),
			slog.String("type", alert.Type),
		)
		return nil
	}

	sender, err := notifications.NewSender(s.logger, config.Services)
	if err != nil {
		return fmt.Errorf("creating notification sender: %w", err)
	}

	title := "Training Load Alert"
	message := formatTrainingLoadMessage(alert)

	if err := sender.Send(ctx, title, message); err != nil {
		s.logger.Error("failed to send training load notification",
			slog.Int64("athlete_id", athleteID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("sending notification: %w", err)
	}

	s.logger.Info("training load notification sent",
		slog.Int64("athlete_id", athleteID),
		slog.String("type", alert.Type),
	)

	return nil
}

// formatTrainingLoadMessage formats the training load alert message.
func formatTrainingLoadMessage(alert TrainingLoadAlert) string {
	switch alert.Type {
	case AlertTypeFatigueWarning:
		return fmt.Sprintf("High fatigue (TSB: %.0f). Consider rest.", alert.TSB)
	case AlertTypeRecoveryAlert:
		return fmt.Sprintf("Recovered! TSB: +%.0f. Ready to push.", alert.TSB)
	case AlertTypeOvertrainingRisk:
		return fmt.Sprintf("Overtraining risk: ATL (%.0f) >> CTL (%.0f)", alert.ATL, alert.CTL)
	default:
		return fmt.Sprintf("Training load alert: TSB=%.0f, ATL=%.0f, CTL=%.0f", alert.TSB, alert.ATL, alert.CTL)
	}
}

// NotifyGearMilestone sends a gear distance milestone notification (if enabled).
func (s *NotificationService) NotifyGearMilestone(ctx context.Context, athleteID int64, gear GearMilestone) error {
	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil || !config.Enabled || !config.Events.GearMilestones {
		s.logger.Debug("gear milestone notification disabled",
			slog.Int64("athlete_id", athleteID),
		)
		return nil
	}

	sender, err := notifications.NewSender(s.logger, config.Services)
	if err != nil {
		return fmt.Errorf("creating notification sender: %w", err)
	}

	title := "Gear Milestone"
	message := fmt.Sprintf("%s hit %d km!", gear.GearName, gear.Milestone)

	if err := sender.Send(ctx, title, message); err != nil {
		s.logger.Error("failed to send gear milestone notification",
			slog.Int64("athlete_id", athleteID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("sending notification: %w", err)
	}

	s.logger.Info("gear milestone notification sent",
		slog.Int64("athlete_id", athleteID),
		slog.String("gear_name", gear.GearName),
		slog.Int("milestone", gear.Milestone),
	)

	return nil
}

// NotifyWeeklyDigest sends a weekly activity digest notification (if enabled).
func (s *NotificationService) NotifyWeeklyDigest(ctx context.Context, athleteID int64, stats storage.DigestStats) error {
	return s.notifyDigest(ctx, athleteID, stats, "week", "Weekly Activity Digest", func(events notifications.EventConfig) bool {
		return events.WeeklyDigest
	})
}

// NotifyMonthlyDigest sends a monthly activity digest notification (if enabled).
func (s *NotificationService) NotifyMonthlyDigest(ctx context.Context, athleteID int64, stats storage.DigestStats) error {
	return s.notifyDigest(ctx, athleteID, stats, "month", "Monthly Activity Digest", func(events notifications.EventConfig) bool {
		return events.MonthlyDigest
	})
}

// notifyDigest is a helper that sends a digest notification if enabled.
func (s *NotificationService) notifyDigest(
	ctx context.Context,
	athleteID int64,
	stats storage.DigestStats,
	period string,
	title string,
	isEnabled func(notifications.EventConfig) bool,
) error {
	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil || !config.Enabled || !isEnabled(config.Events) {
		s.logger.Debug(period+" digest notification disabled",
			slog.Int64("athlete_id", athleteID),
		)
		return nil
	}

	sender, err := notifications.NewSender(s.logger, config.Services)
	if err != nil {
		return fmt.Errorf("creating notification sender: %w", err)
	}

	message := formatDigestMessage(stats, period)

	if err := sender.Send(ctx, title, message); err != nil {
		s.logger.Error("failed to send "+period+" digest notification",
			slog.Int64("athlete_id", athleteID),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("sending notification: %w", err)
	}

	s.logger.Info(period+" digest notification sent",
		slog.Int64("athlete_id", athleteID),
		slog.Int("activities", stats.ActivityCount),
	)

	return nil
}

// formatDigestMessage formats the digest notification message.
func formatDigestMessage(stats storage.DigestStats, period string) string {
	distanceKm := stats.TotalDistance / 1000.0
	hours := stats.TotalTime / 3600
	minutes := (stats.TotalTime % 3600) / 60

	var timeStr string
	if hours > 0 {
		timeStr = fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		timeStr = fmt.Sprintf("%dm", minutes)
	}

	message := fmt.Sprintf(
		"Your %s in review (%s - %s):\n\n"+
			"Activities: %d\n"+
			"Distance: %.1f km\n"+
			"Time: %s\n"+
			"Elevation: %.0f m",
		period,
		stats.StartDate.Format("Jan 2"),
		stats.EndDate.Format("Jan 2"),
		stats.ActivityCount,
		distanceKm,
		timeStr,
		stats.TotalElevation,
	)

	if stats.TotalCalories > 0 {
		message += fmt.Sprintf("\nCalories: %.0f", stats.TotalCalories)
	}

	return message
}

// TestNotification sends a test message to verify configuration.
func (s *NotificationService) TestNotification(ctx context.Context, athleteID int64, serviceID string) error {
	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil {
		return BadRequest("notifications not configured")
	}

	// Find the service by ID
	var targetService *notifications.ServiceConfig
	for _, svc := range config.Services {
		if svc.ID == serviceID {
			targetService = &svc
			break
		}
	}

	if targetService == nil {
		return NotFound("notification service")
	}

	if err := notifications.TestService(ctx, s.logger, *targetService); err != nil {
		return Wrapf(ErrBadRequest, "test notification failed: %v", err)
	}

	return nil
}

// TestAllResult represents the result of testing a single notification service.
type TestAllResult struct {
	ServiceID   string `json:"serviceId"`
	ServiceName string `json:"serviceName"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// TestAllNotifications tests all enabled notification services.
func (s *NotificationService) TestAllNotifications(ctx context.Context, athleteID int64) ([]TestAllResult, error) {
	config, err := s.getNotificationConfig(ctx, athleteID)
	if err != nil {
		return nil, fmt.Errorf("getting notification config: %w", err)
	}

	if config == nil {
		return []TestAllResult{}, nil
	}

	var results []TestAllResult
	for _, svc := range config.Services {
		if !svc.Enabled {
			continue
		}

		result := TestAllResult{
			ServiceID:   svc.ID,
			ServiceName: svc.Name,
			Success:     true,
		}

		if err := notifications.TestService(ctx, s.logger, svc); err != nil {
			result.Success = false
			result.Error = err.Error()
		}

		results = append(results, result)
	}

	return results, nil
}

// getNotificationConfig retrieves the notification configuration for an athlete.
func (s *NotificationService) getNotificationConfig(ctx context.Context, athleteID int64) (*notifications.NotificationConfig, error) {
	settings, err := s.settings.Get(ctx, athleteID)
	if err != nil {
		return nil, fmt.Errorf("getting athlete settings: %w", err)
	}

	return settings.Notifications, nil
}
