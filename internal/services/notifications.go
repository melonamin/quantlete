package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/melonamin/quantlete/internal/notifications"
	"github.com/melonamin/quantlete/internal/storage"
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
