package notifications

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
			MaintenanceSchedule: "weekly",
		},
	}
}
