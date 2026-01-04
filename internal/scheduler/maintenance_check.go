package scheduler

import (
	"context"
	"log/slog"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
)

// MaintenanceCheckJob checks for due maintenance and sends notifications.
type MaintenanceCheckJob struct {
	logger        *slog.Logger
	maintenance   *services.MaintenanceService
	notifications *services.NotificationService
	settings      *storage.SettingsRepository
	athleteID     int64
}

// NewMaintenanceCheckJob creates a new maintenance check job.
func NewMaintenanceCheckJob(
	logger *slog.Logger,
	maintenance *services.MaintenanceService,
	notifications *services.NotificationService,
	settings *storage.SettingsRepository,
	athleteID int64,
) *MaintenanceCheckJob {
	return &MaintenanceCheckJob{
		logger:        logger,
		maintenance:   maintenance,
		notifications: notifications,
		settings:      settings,
		athleteID:     athleteID,
	}
}

// Run executes the maintenance check (called by cron).
func (j *MaintenanceCheckJob) Run() {
	ctx := context.Background()

	// Get due maintenance items
	dueItems, err := j.maintenance.ListDue(ctx, j.athleteID)
	if err != nil {
		j.logger.Error("maintenance check: failed to get due items",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	if len(dueItems) == 0 {
		j.logger.Debug("maintenance check: no items due",
			slog.Int64("athlete_id", j.athleteID),
		)
		return
	}

	// Filter to only items that are actually due
	var notifyItems []services.MaintenanceItem
	for _, item := range dueItems {
		if !item.IsDue {
			continue
		}
		// Find the rule that triggered the due status
		for _, progress := range item.Progress {
			if progress.Due {
				notifyItems = append(notifyItems, services.MaintenanceItem{
					ComponentID:    item.ID,
					ComponentName:  item.Name,
					GearName:       item.GearID, // Use GearID as fallback
					RuleType:       progress.Type,
					CurrentValue:   progress.CurrentValue,
					ThresholdValue: progress.ThresholdValue,
				})
				break // Only report first due rule per component
			}
		}
	}

	if len(notifyItems) == 0 {
		return
	}

	// Send notification
	if err := j.notifications.NotifyMaintenanceDue(ctx, j.athleteID, notifyItems); err != nil {
		j.logger.Error("maintenance check: failed to send notification",
			slog.Int64("athlete_id", j.athleteID),
			slog.String("error", err.Error()),
		)
		return
	}

	j.logger.Info("maintenance check: notification sent",
		slog.Int64("athlete_id", j.athleteID),
		slog.Int("items_count", len(notifyItems)),
	)
}
