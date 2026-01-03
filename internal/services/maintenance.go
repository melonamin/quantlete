package services

import (
	"context"
	"time"

	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/storage"
)

// MaintenanceService handles maintenance business logic.
type MaintenanceService struct {
	repo *storage.MaintenanceRepository
}

// NewMaintenanceService creates a new maintenance service.
func NewMaintenanceService(repo *storage.MaintenanceRepository) *MaintenanceService {
	return &MaintenanceService{repo: repo}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// ListComponentsInput contains parameters for listing components.
type ListComponentsInput struct {
	AthleteID int64
	GearID    string
	Page      int
	PerPage   int
	OrderBy   string
	OrderDir  string
}

// RuleItem represents a maintenance rule in responses.
type RuleItem struct {
	ID             int64   `json:"id"`
	ComponentID    int64   `json:"component_id"`
	Type           string  `json:"type"`
	ThresholdValue float64 `json:"threshold_value"`
}

// ComponentItem represents a component in responses.
type ComponentItem struct {
	ID                 int64      `json:"id"`
	GearID             string     `json:"gear_id"`
	Name               string     `json:"name"`
	ImageURL           string     `json:"image_url,omitempty"`
	MaintenanceHashtag string     `json:"maintenance_hashtag,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastCompletedAt    *time.Time `json:"last_completed_at,omitempty"`
	Rules              []RuleItem `json:"rules"`
}

// ListComponentsOutput contains a paginated list of components.
type ListComponentsOutput struct {
	Data       []ComponentItem `json:"data"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

// RuleInput represents a rule in create/update requests.
type RuleInput struct {
	Type           string  `json:"type"`
	ThresholdValue float64 `json:"threshold_value"`
}

// CreateComponentInput contains parameters for creating a component.
type CreateComponentInput struct {
	AthleteID          int64
	GearID             string
	Name               string      `json:"name"`
	ImageURL           string      `json:"image_url,omitempty"`
	MaintenanceHashtag string      `json:"maintenance_hashtag,omitempty"`
	Rules              []RuleInput `json:"rules,omitempty"`
}

// UpdateComponentInput contains parameters for updating a component.
type UpdateComponentInput struct {
	AthleteID          int64
	ComponentID        int64
	Name               *string      `json:"name,omitempty"`
	ImageURL           *string      `json:"image_url,omitempty"`
	MaintenanceHashtag *string      `json:"maintenance_hashtag,omitempty"`
	Rules              *[]RuleInput `json:"rules,omitempty"`
}

// LogMaintenanceInput contains parameters for logging maintenance.
type LogMaintenanceInput struct {
	AthleteID   int64
	ComponentID int64
	ActivityID  *int64
	CompletedAt time.Time
}

// RuleProgressItem represents progress for a single rule.
type RuleProgressItem struct {
	Type           string  `json:"type"`
	ThresholdValue float64 `json:"threshold_value"`
	CurrentValue   float64 `json:"current_value"`
	Percent        float64 `json:"percent"`
	Due            bool    `json:"due"`
}

// DueComponentItem represents a component with maintenance status.
type DueComponentItem struct {
	ID                 int64              `json:"id"`
	GearID             string             `json:"gear_id"`
	Name               string             `json:"name"`
	ImageURL           string             `json:"image_url,omitempty"`
	MaintenanceHashtag string             `json:"maintenance_hashtag,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	LastCompletedAt    *time.Time         `json:"last_completed_at,omitempty"`
	Rules              []RuleItem         `json:"rules"`
	DistanceSince      float64            `json:"distance_since"`
	MovingTimeSince    int                `json:"moving_time_since"`
	DaysSince          int                `json:"days_since"`
	Progress           []RuleProgressItem `json:"progress"`
	IsDue              bool               `json:"is_due"`
}

// ============================================================================
// Service Methods
// ============================================================================

// ListComponents returns a paginated list of components for a gear item.
func (s *MaintenanceService) ListComponents(ctx context.Context, in ListComponentsInput) (*ListComponentsOutput, error) {
	if in.GearID == "" {
		return nil, BadRequest("gear ID is required")
	}

	f := storage.ComponentFilters{
		QueryParams: pagination.QueryParams{
			Page:     in.Page,
			PerPage:  in.PerPage,
			OrderBy:  in.OrderBy,
			OrderDir: in.OrderDir,
		},
	}

	result, err := s.repo.ListComponentsPaginated(ctx, in.AthleteID, in.GearID, f)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to list components: %v", err)
	}

	items := make([]ComponentItem, 0, len(result.Items))
	for _, c := range result.Items {
		items = append(items, componentToItem(c))
	}

	return &ListComponentsOutput{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}, nil
}

// CreateComponent creates a new component for a gear item.
func (s *MaintenanceService) CreateComponent(ctx context.Context, in CreateComponentInput) (*ComponentItem, error) {
	if in.GearID == "" {
		return nil, BadRequest("gear ID is required")
	}

	rules := make([]storage.CreateRuleInput, 0, len(in.Rules))
	for _, r := range in.Rules {
		rules = append(rules, storage.CreateRuleInput{
			Type:           r.Type,
			ThresholdValue: r.ThresholdValue,
		})
	}

	created, err := s.repo.CreateComponent(ctx, in.AthleteID, in.GearID, storage.CreateComponentInput{
		Name:               in.Name,
		ImageURL:           in.ImageURL,
		MaintenanceHashtag: in.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		return nil, Wrapf(ErrBadRequest, "%v", err)
	}
	if created == nil {
		return nil, NotFound("gear")
	}

	item := componentToItem(*created)
	return &item, nil
}

// UpdateComponent updates an existing component.
func (s *MaintenanceService) UpdateComponent(ctx context.Context, in UpdateComponentInput) (*ComponentItem, error) {
	if in.ComponentID <= 0 {
		return nil, BadRequest("component ID is required")
	}

	var rules *[]storage.CreateRuleInput
	if in.Rules != nil {
		out := make([]storage.CreateRuleInput, 0, len(*in.Rules))
		for _, r := range *in.Rules {
			out = append(out, storage.CreateRuleInput{
				Type:           r.Type,
				ThresholdValue: r.ThresholdValue,
			})
		}
		rules = &out
	}

	updated, err := s.repo.UpdateComponent(ctx, in.AthleteID, in.ComponentID, storage.UpdateComponentInput{
		Name:               in.Name,
		ImageURL:           in.ImageURL,
		MaintenanceHashtag: in.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		return nil, Wrapf(ErrBadRequest, "%v", err)
	}
	if updated == nil {
		return nil, NotFound("component")
	}

	item := componentToItem(*updated)
	return &item, nil
}

// DeleteComponent deletes a component.
func (s *MaintenanceService) DeleteComponent(ctx context.Context, athleteID, componentID int64) error {
	if componentID <= 0 {
		return BadRequest("component ID is required")
	}

	if err := s.repo.DeleteComponent(ctx, athleteID, componentID); err != nil {
		return Wrapf(ErrInternal, "failed to delete component: %v", err)
	}

	return nil
}

// LogMaintenance logs a maintenance event for a component.
func (s *MaintenanceService) LogMaintenance(ctx context.Context, in LogMaintenanceInput) error {
	if in.ComponentID <= 0 {
		return BadRequest("component ID is required")
	}

	if err := s.repo.LogMaintenance(ctx, in.AthleteID, in.ComponentID, in.ActivityID, in.CompletedAt); err != nil {
		return Wrapf(ErrInternal, "failed to log maintenance: %v", err)
	}

	return nil
}

// ListDue returns components with their maintenance status for an athlete.
func (s *MaintenanceService) ListDue(ctx context.Context, athleteID int64) ([]DueComponentItem, error) {
	due, err := s.repo.Due(ctx, athleteID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch due maintenance: %v", err)
	}

	items := make([]DueComponentItem, 0, len(due))
	for _, d := range due {
		items = append(items, dueComponentToItem(d))
	}

	return items, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

func componentToItem(c storage.ComponentWithRules) ComponentItem {
	rules := make([]RuleItem, 0, len(c.Rules))
	for _, r := range c.Rules {
		rules = append(rules, RuleItem{
			ID:             r.ID,
			ComponentID:    r.ComponentID,
			Type:           r.Type,
			ThresholdValue: r.ThresholdValue,
		})
	}

	var lastCompleted *time.Time
	if c.LastCompletedAt != nil && !c.LastCompletedAt.IsZero() {
		t := c.LastCompletedAt.Time
		lastCompleted = &t
	}

	return ComponentItem{
		ID:                 c.ID,
		GearID:             c.GearID,
		Name:               c.Name,
		ImageURL:           c.ImageURL,
		MaintenanceHashtag: c.MaintenanceHashtag,
		CreatedAt:          c.CreatedAt.Time,
		UpdatedAt:          c.UpdatedAt.Time,
		LastCompletedAt:    lastCompleted,
		Rules:              rules,
	}
}

func dueComponentToItem(d storage.DueComponent) DueComponentItem {
	rules := make([]RuleItem, 0, len(d.Rules))
	for _, r := range d.Rules {
		rules = append(rules, RuleItem{
			ID:             r.ID,
			ComponentID:    r.ComponentID,
			Type:           r.Type,
			ThresholdValue: r.ThresholdValue,
		})
	}

	progress := make([]RuleProgressItem, 0, len(d.Progress))
	for _, p := range d.Progress {
		progress = append(progress, RuleProgressItem{
			Type:           p.Type,
			ThresholdValue: p.ThresholdValue,
			CurrentValue:   p.CurrentValue,
			Percent:        p.Percent,
			Due:            p.Due,
		})
	}

	var lastCompleted *time.Time
	if d.LastCompletedAt != nil && !d.LastCompletedAt.IsZero() {
		t := d.LastCompletedAt.Time
		lastCompleted = &t
	}

	return DueComponentItem{
		ID:                 d.ID,
		GearID:             d.GearID,
		Name:               d.Name,
		ImageURL:           d.ImageURL,
		MaintenanceHashtag: d.MaintenanceHashtag,
		CreatedAt:          d.CreatedAt.Time,
		UpdatedAt:          d.UpdatedAt.Time,
		LastCompletedAt:    lastCompleted,
		Rules:              rules,
		DistanceSince:      d.DistanceSince,
		MovingTimeSince:    d.MovingTimeSince,
		DaysSince:          d.DaysSince,
		Progress:           progress,
		IsDue:              d.IsDue,
	}
}
