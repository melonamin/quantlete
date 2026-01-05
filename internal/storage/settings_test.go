package storage

import (
	"encoding/json"
	"testing"

	"github.com/melonamin/quantlete/internal/notifications"
)

func TestAthleteSettings_JSONRoundTrip(t *testing.T) {
	// Create settings with notifications configured
	original := AthleteSettings{
		Version: AthleteSettingsVersion,
		VirtualWorldTileLayers: map[string]VirtualWorldTileLayer{
			"custom": {
				Name:        "Custom Layer",
				URL:         "https://tiles.example.com/{z}/{x}/{y}.png",
				Attribution: "Custom",
				MaxZoom:     18,
			},
		},
		EddingtonDefinitions: []EddingtonDefinition{
			{
				ID:         "all",
				Name:       "All activities",
				SportTypes: nil,
				ShowInNav:  true,
			},
		},
		Scheduler: SchedulerSettings{
			Version: SchedulerSettingsVersion,
			Pull: PullSettings{
				Enabled:  true,
				Schedule: PullScheduleHourly,
			},
			Push: PushSettings{
				Enabled: false,
			},
		},
		EnablePublicBadges: true,
		Notifications: &notifications.NotificationConfig{
			Enabled: true,
			Services: []notifications.ServiceConfig{
				{
					ID:      "telegram-1",
					Type:    notifications.ServiceTypeTelegram,
					Name:    "My Telegram",
					Enabled: true,
					Config: map[string]string{
						"token":   "123456:ABC-DEF",
						"chat_id": "-1001234567890",
					},
				},
			},
			Events: notifications.EventConfig{
				ImportComplete:      true,
				MaintenanceDue:      false,
				MaintenanceSchedule: notifications.MaintenanceScheduleWeekly,
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal back
	var restored AthleteSettings
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify key fields
	if restored.Version != original.Version {
		t.Errorf("Version = %d, want %d", restored.Version, original.Version)
	}

	if restored.EnablePublicBadges != original.EnablePublicBadges {
		t.Errorf("EnablePublicBadges = %v, want %v", restored.EnablePublicBadges, original.EnablePublicBadges)
	}

	// Verify scheduler settings
	if restored.Scheduler.Pull.Enabled != original.Scheduler.Pull.Enabled {
		t.Errorf("Scheduler.Pull.Enabled = %v, want %v", restored.Scheduler.Pull.Enabled, original.Scheduler.Pull.Enabled)
	}
	if restored.Scheduler.Pull.Schedule != original.Scheduler.Pull.Schedule {
		t.Errorf("Scheduler.Pull.Schedule = %q, want %q", restored.Scheduler.Pull.Schedule, original.Scheduler.Pull.Schedule)
	}

	// Verify notification settings
	if restored.Notifications == nil {
		t.Fatal("Notifications = nil, want non-nil")
	}

	if restored.Notifications.Enabled != original.Notifications.Enabled {
		t.Errorf("Notifications.Enabled = %v, want %v", restored.Notifications.Enabled, original.Notifications.Enabled)
	}

	if len(restored.Notifications.Services) != len(original.Notifications.Services) {
		t.Fatalf("Notifications.Services len = %d, want %d", len(restored.Notifications.Services), len(original.Notifications.Services))
	}

	// Verify service config
	restoredSvc := restored.Notifications.Services[0]
	originalSvc := original.Notifications.Services[0]

	if restoredSvc.ID != originalSvc.ID {
		t.Errorf("Service.ID = %q, want %q", restoredSvc.ID, originalSvc.ID)
	}
	if restoredSvc.Type != originalSvc.Type {
		t.Errorf("Service.Type = %q, want %q", restoredSvc.Type, originalSvc.Type)
	}
	if restoredSvc.Name != originalSvc.Name {
		t.Errorf("Service.Name = %q, want %q", restoredSvc.Name, originalSvc.Name)
	}
	if restoredSvc.Config["token"] != originalSvc.Config["token"] {
		t.Errorf("Service.Config[token] = %q, want %q", restoredSvc.Config["token"], originalSvc.Config["token"])
	}

	// Verify events config
	if restored.Notifications.Events.ImportComplete != original.Notifications.Events.ImportComplete {
		t.Errorf("Events.ImportComplete = %v, want %v", restored.Notifications.Events.ImportComplete, original.Notifications.Events.ImportComplete)
	}
	if restored.Notifications.Events.MaintenanceSchedule != original.Notifications.Events.MaintenanceSchedule {
		t.Errorf("Events.MaintenanceSchedule = %q, want %q", restored.Notifications.Events.MaintenanceSchedule, original.Notifications.Events.MaintenanceSchedule)
	}
}

func TestAthleteSettings_MigrationV3ToV4(t *testing.T) {
	// Simulate old settings without notifications
	oldSettings := AthleteSettings{
		Version:                3,
		VirtualWorldTileLayers: map[string]VirtualWorldTileLayer{},
		Scheduler: SchedulerSettings{
			Version: SchedulerSettingsVersion,
		},
	}

	// Run migration
	migrateAthleteSettings(&oldSettings)

	// Verify version was bumped
	if oldSettings.Version != AthleteSettingsVersion {
		t.Errorf("Version after migration = %d, want %d", oldSettings.Version, AthleteSettingsVersion)
	}

	// Verify notifications were added with defaults
	if oldSettings.Notifications == nil {
		t.Fatal("Notifications = nil after migration, want non-nil")
	}

	if oldSettings.Notifications.Enabled {
		t.Error("Notifications.Enabled = true after migration, want false (opt-in)")
	}

	if len(oldSettings.Notifications.Services) != 0 {
		t.Errorf("Notifications.Services len = %d after migration, want 0", len(oldSettings.Notifications.Services))
	}
}

func TestAthleteSettings_DefaultHasNotifications(t *testing.T) {
	settings := DefaultAthleteSettings()

	if settings.Notifications == nil {
		t.Fatal("DefaultAthleteSettings().Notifications = nil, want non-nil")
	}

	if settings.Notifications.Enabled {
		t.Error("Default notifications should be disabled")
	}
}
