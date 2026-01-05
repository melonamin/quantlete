package notifications

import (
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
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

	// Achievements (during import)
	PersonalRecords   bool `json:"personalRecords"`   // Best effort PRs
	SegmentPRs        bool `json:"segmentPRs"`        // Segment personal records
	EddingtonIncrease bool `json:"eddingtonIncrease"` // Eddington number goes up
	PowerRecords      bool `json:"powerRecords"`      // Peak power PRs
	GoalComplete      bool `json:"goalComplete"`      // 100% goal completion

	// Training Load (during import)
	FatigueWarning   bool `json:"fatigueWarning"`   // TSB < -20
	RecoveryAlert    bool `json:"recoveryAlert"`    // TSB > +10
	OvertrainingRisk bool `json:"overtrainingRisk"` // ATL > CTL × 1.5

	// Summaries (scheduled)
	WeeklyDigest  bool `json:"weeklyDigest"`
	MonthlyDigest bool `json:"monthlyDigest"`

	// Gear
	GearMilestones bool `json:"gearMilestones"` // Every 5000 km
}

// DefaultNotificationConfig returns the default opt-in config.
// Note: When notifications are enabled, importComplete defaults to true as the most common use case.
func DefaultNotificationConfig() NotificationConfig {
	return NotificationConfig{
		Enabled:  false,
		Services: []ServiceConfig{},
		Events: EventConfig{
			ImportComplete:      false,
			MaintenanceDue:      false,
			MaintenanceSchedule: MaintenanceScheduleWeekly,
			// Achievements (during import)
			PersonalRecords:   false,
			SegmentPRs:        false,
			EddingtonIncrease: false,
			PowerRecords:      false,
			GoalComplete:      false,
			// Training Load (during import)
			FatigueWarning:   false,
			RecoveryAlert:    false,
			OvertrainingRisk: false,
			// Summaries (scheduled)
			WeeklyDigest:  false,
			MonthlyDigest: false,
			// Gear
			GearMilestones: false,
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
	port := s.Config["port"]
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

	// Validate port if provided
	if port != "" {
		portNum, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("smtp port must be a number: %w", err)
		}
		if portNum < 1 || portNum > 65535 {
			return fmt.Errorf("smtp port must be between 1 and 65535")
		}
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
// Generic webhooks in shoutrrr use the "generic" scheme format:
// generic://webhook.example.com/path or generic+https://webhook.example.com/path
// However, we also accept standard HTTP/HTTPS URLs for convenience and convert them.
func (s *ServiceConfig) validateGeneric() error {
	urlStr := s.Config["url"]
	if urlStr == "" {
		return fmt.Errorf("generic url is required")
	}

	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("generic url is invalid: %w", err)
	}

	// Accept http, https, generic, generic+http, generic+https schemes
	validSchemes := map[string]bool{
		"http":          true,
		"https":         true,
		"generic":       true,
		"generic+http":  true,
		"generic+https": true,
	}
	if !validSchemes[parsed.Scheme] {
		return fmt.Errorf("generic url must use http, https, or generic scheme (got %q)", parsed.Scheme)
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
