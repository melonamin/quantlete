package notifications

import (
	"strings"
	"testing"
)

func TestServiceConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ServiceConfig
		wantErr string
	}{
		{
			name: "valid telegram",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeTelegram,
				Name:    "My Telegram",
				Enabled: true,
				Config: map[string]string{
					"token":   "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
					"chat_id": "-1001234567890",
				},
			},
			wantErr: "",
		},
		{
			name: "telegram missing token",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeTelegram,
				Name:    "My Telegram",
				Enabled: true,
				Config: map[string]string{
					"chat_id": "-1001234567890",
				},
			},
			wantErr: "telegram token is required",
		},
		{
			name: "telegram missing chat_id",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeTelegram,
				Name:    "My Telegram",
				Enabled: true,
				Config: map[string]string{
					"token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
				},
			},
			wantErr: "telegram chat_id is required",
		},
		{
			name: "telegram invalid token format",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeTelegram,
				Name:    "My Telegram",
				Enabled: true,
				Config: map[string]string{
					"token":   "invalid-token-without-colon",
					"chat_id": "-1001234567890",
				},
			},
			wantErr: "telegram token format is invalid",
		},
		{
			name: "valid smtp",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeSMTP,
				Name:    "My SMTP",
				Enabled: true,
				Config: map[string]string{
					"host": "smtp.example.com",
					"port": "587",
					"from": "sender@example.com",
					"to":   "receiver@example.com",
				},
			},
			wantErr: "",
		},
		{
			name: "smtp missing host",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeSMTP,
				Name:    "My SMTP",
				Enabled: true,
				Config: map[string]string{
					"from": "sender@example.com",
					"to":   "receiver@example.com",
				},
			},
			wantErr: "smtp host is required",
		},
		{
			name: "smtp invalid from address",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeSMTP,
				Name:    "My SMTP",
				Enabled: true,
				Config: map[string]string{
					"host": "smtp.example.com",
					"from": "not-an-email",
					"to":   "receiver@example.com",
				},
			},
			wantErr: "smtp from address is invalid",
		},
		{
			name: "smtp invalid to address",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeSMTP,
				Name:    "My SMTP",
				Enabled: true,
				Config: map[string]string{
					"host": "smtp.example.com",
					"from": "sender@example.com",
					"to":   "also-not-an-email",
				},
			},
			wantErr: "smtp to address is invalid",
		},
		{
			name: "valid generic webhook",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeGeneric,
				Name:    "My Webhook",
				Enabled: true,
				Config: map[string]string{
					"url": "https://api.example.com/webhook",
				},
			},
			wantErr: "",
		},
		{
			name: "generic missing url",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeGeneric,
				Name:    "My Webhook",
				Enabled: true,
				Config:  map[string]string{},
			},
			wantErr: "generic url is required",
		},
		{
			name: "generic invalid scheme",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeGeneric,
				Name:    "My Webhook",
				Enabled: true,
				Config: map[string]string{
					"url": "ftp://example.com/webhook",
				},
			},
			wantErr: "generic url must use http or https scheme",
		},
		{
			name: "generic missing host",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeGeneric,
				Name:    "My Webhook",
				Enabled: true,
				Config: map[string]string{
					"url": "https:///path-only",
				},
			},
			wantErr: "generic url must have a host",
		},
		{
			name: "missing id",
			config: ServiceConfig{
				Type:    ServiceTypeTelegram,
				Name:    "My Telegram",
				Enabled: true,
				Config: map[string]string{
					"token":   "123456:ABC",
					"chat_id": "-100123",
				},
			},
			wantErr: "id is required",
		},
		{
			name: "missing name",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    ServiceTypeTelegram,
				Enabled: true,
				Config: map[string]string{
					"token":   "123456:ABC",
					"chat_id": "-100123",
				},
			},
			wantErr: "name is required",
		},
		{
			name: "unknown service type",
			config: ServiceConfig{
				ID:      "test-id",
				Type:    "unknown",
				Name:    "Unknown",
				Enabled: true,
				Config:  map[string]string{},
			},
			wantErr: "unknown service type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestEventConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  EventConfig
		wantErr string
	}{
		{
			name: "valid weekly",
			config: EventConfig{
				ImportComplete:      true,
				MaintenanceDue:      true,
				MaintenanceSchedule: MaintenanceScheduleWeekly,
			},
			wantErr: "",
		},
		{
			name: "valid monthly",
			config: EventConfig{
				ImportComplete:      false,
				MaintenanceDue:      true,
				MaintenanceSchedule: MaintenanceScheduleMonthly,
			},
			wantErr: "",
		},
		{
			name: "valid empty schedule",
			config: EventConfig{
				ImportComplete:      true,
				MaintenanceDue:      false,
				MaintenanceSchedule: "",
			},
			wantErr: "",
		},
		{
			name: "invalid schedule",
			config: EventConfig{
				ImportComplete:      true,
				MaintenanceDue:      true,
				MaintenanceSchedule: "daily",
			},
			wantErr: "invalid maintenance schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestNotificationConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  NotificationConfig
		wantErr string
	}{
		{
			name:    "valid default config",
			config:  DefaultNotificationConfig(),
			wantErr: "",
		},
		{
			name: "valid config with services",
			config: NotificationConfig{
				Enabled: true,
				Services: []ServiceConfig{
					{
						ID:      "svc-1",
						Type:    ServiceTypeTelegram,
						Name:    "Telegram",
						Enabled: true,
						Config: map[string]string{
							"token":   "123456:ABC",
							"chat_id": "-100123",
						},
					},
				},
				Events: EventConfig{
					ImportComplete:      true,
					MaintenanceSchedule: MaintenanceScheduleWeekly,
				},
			},
			wantErr: "",
		},
		{
			name: "invalid service in config",
			config: NotificationConfig{
				Enabled: true,
				Services: []ServiceConfig{
					{
						ID:      "svc-1",
						Type:    ServiceTypeTelegram,
						Name:    "Telegram",
						Enabled: true,
						Config:  map[string]string{}, // Missing required fields
					},
				},
				Events: EventConfig{
					MaintenanceSchedule: MaintenanceScheduleWeekly,
				},
			},
			wantErr: "telegram token is required",
		},
		{
			name: "invalid events config",
			config: NotificationConfig{
				Enabled:  true,
				Services: []ServiceConfig{},
				Events: EventConfig{
					MaintenanceSchedule: "invalid",
				},
			},
			wantErr: "invalid maintenance schedule",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, want error containing %q", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDefaultNotificationConfig(t *testing.T) {
	cfg := DefaultNotificationConfig()

	if cfg.Enabled {
		t.Error("default config should be disabled")
	}

	if len(cfg.Services) != 0 {
		t.Errorf("default config should have no services, got %d", len(cfg.Services))
	}

	if cfg.Events.ImportComplete {
		t.Error("default config should have ImportComplete disabled")
	}

	if cfg.Events.MaintenanceDue {
		t.Error("default config should have MaintenanceDue disabled")
	}

	if cfg.Events.MaintenanceSchedule != MaintenanceScheduleWeekly {
		t.Errorf("default config MaintenanceSchedule = %q, want %q",
			cfg.Events.MaintenanceSchedule, MaintenanceScheduleWeekly)
	}
}
