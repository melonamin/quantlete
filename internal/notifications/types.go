package notifications

import (
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

// ServiceType constants for notification service types.
const (
	ServiceTypeTelegram = "telegram"
	ServiceTypeSMTP     = "smtp"
	ServiceTypeGeneric  = "generic"
)

// MaintenanceSchedule constants for maintenance notification schedules.
const (
	MaintenanceScheduleWeekly  = "weekly"
	MaintenanceScheduleMonthly = "monthly"
)

// NotificationConfig holds the notification configuration for an athlete.
type NotificationConfig struct {
	Enabled  bool            `json:"enabled"`
	Services []ServiceConfig `json:"services"`
	Events   EventConfig     `json:"events"`
}

// ServiceConfig defines a notification service configuration.
type ServiceConfig struct {
	ID      string            `json:"id"`   // UUID
	Type    string            `json:"type"` // telegram|smtp|generic
	Name    string            `json:"name"` // User label
	Enabled bool              `json:"enabled"`
	Config  map[string]string `json:"config"` // Service-specific
}

// EventConfig defines which events trigger notifications.
type EventConfig struct {
	ImportComplete      bool   `json:"importComplete"`
	MaintenanceDue      bool   `json:"maintenanceDue"`
	MaintenanceSchedule string `json:"maintenanceSchedule"` // "weekly"|"monthly"
}

// DefaultNotificationConfig returns the default opt-in config.
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		Enabled:  false,
		Services: []ServiceConfig{},
		Events: EventConfig{
			ImportComplete:      false,
			MaintenanceDue:      false,
			MaintenanceSchedule: MaintenanceScheduleWeekly,
		},
	}
}

// Validate checks if the NotificationConfig is valid.
func (c *NotificationConfig) Validate() error {
	for i, svc := range c.Services {
		if err := svc.Validate(); err != nil {
			return fmt.Errorf("service[%d] %q: %w", i, svc.Name, err)
		}
	}
	if err := c.Events.Validate(); err != nil {
		return fmt.Errorf("events: %w", err)
	}
	return nil
}

// Validate checks if the ServiceConfig is valid.
func (s *ServiceConfig) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("id is required")
	}
	if s.Name == "" {
		return fmt.Errorf("name is required")
	}

	switch s.Type {
	case ServiceTypeTelegram:
		return s.validateTelegram()
	case ServiceTypeSMTP:
		return s.validateSMTP()
	case ServiceTypeGeneric:
		return s.validateGeneric()
	default:
		return fmt.Errorf("unknown service type: %q", s.Type)
	}
}

// validateTelegram validates Telegram-specific configuration.
func (s *ServiceConfig) validateTelegram() error {
	token := s.Config["token"]
	chatID := s.Config["chat_id"]

	if token == "" {
		return fmt.Errorf("telegram token is required")
	}
	if chatID == "" {
		return fmt.Errorf("telegram chat_id is required")
	}

	// Basic token format validation (should contain a colon)
	if !strings.Contains(token, ":") {
		return fmt.Errorf("telegram token format is invalid (expected format: BOT_ID:SECRET)")
	}

	return nil
}

// validateSMTP validates SMTP-specific configuration.
func (s *ServiceConfig) validateSMTP() error {
	host := s.Config["host"]
	from := s.Config["from"]
	to := s.Config["to"]

	if host == "" {
		return fmt.Errorf("smtp host is required")
	}
	if from == "" {
		return fmt.Errorf("smtp from address is required")
	}
	if to == "" {
		return fmt.Errorf("smtp to address is required")
	}

	// Validate email addresses
	if _, err := mail.ParseAddress(from); err != nil {
		return fmt.Errorf("smtp from address is invalid: %w", err)
	}
	if _, err := mail.ParseAddress(to); err != nil {
		return fmt.Errorf("smtp to address is invalid: %w", err)
	}

	return nil
}

// validateGeneric validates generic webhook configuration.
func (s *ServiceConfig) validateGeneric() error {
	urlStr := s.Config["url"]
	if urlStr == "" {
		return fmt.Errorf("generic url is required")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("generic url is invalid: %w", err)
	}

	// Only allow http/https schemes
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("generic url must use http or https scheme")
	}

	if parsed.Host == "" {
		return fmt.Errorf("generic url must have a host")
	}

	return nil
}

// Validate checks if the EventConfig is valid.
func (e *EventConfig) Validate() error {
	switch e.MaintenanceSchedule {
	case MaintenanceScheduleWeekly, MaintenanceScheduleMonthly, "":
		return nil
	default:
		return fmt.Errorf("invalid maintenance schedule: %q (must be %q or %q)",
			e.MaintenanceSchedule, MaintenanceScheduleWeekly, MaintenanceScheduleMonthly)
	}
}
