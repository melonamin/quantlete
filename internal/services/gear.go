package services

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/storage"
)

// Validation constants for gear inputs.
const (
	MaxGearNameLength    = 100
	MaxHashtagLength     = 50
	MaxDescriptionLength = 500
)

// hashtagRe matches valid hashtag format: alphanumeric with optional underscores/hyphens.
var hashtagRe = regexp.MustCompile(`^#?[A-Za-z0-9][A-Za-z0-9_-]{0,48}$`)

// GearService handles gear business logic.
type GearService struct {
	repo *storage.GearRepository
}

// NewGearService creates a new gear service.
func NewGearService(repo *storage.GearRepository) *GearService {
	return &GearService{repo: repo}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// ListGearInput contains parameters for listing gear.
type ListGearInput struct {
	AthleteID      int64  `json:"-" adapter:"context"`
	IncludeRetired bool   `json:"include_retired" adapter:"query"`
	Page           int    `json:"page" adapter:"query"`
	PerPage        int    `json:"per_page" adapter:"query"`
	OrderBy        string `json:"order_by" adapter:"query"`
	OrderDir       string `json:"order_dir" adapter:"query"`
}

// GearItem represents a gear item in responses.
type GearItem struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Primary          bool     `json:"primary"`
	Retired          bool     `json:"retired"`
	Distance         float64  `json:"distance"`
	BrandName        string   `json:"brand_name,omitempty"`
	ModelName        string   `json:"model_name,omitempty"`
	Description      string   `json:"description,omitempty"`
	Source           string   `json:"source"`
	Hashtag          string   `json:"hashtag,omitempty"`
	PurchasePrice    *float64 `json:"purchase_price,omitempty"`
	PurchaseCurrency string   `json:"purchase_currency,omitempty"`
	ActivityCount    int      `json:"activity_count"`
}

// ListGearOutput contains a paginated list of gear.
type ListGearOutput struct {
	Data       []GearItem `json:"data"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PerPage    int        `json:"per_page"`
	TotalPages int        `json:"total_pages"`
}

// GetGearInput contains parameters for getting a single gear item.
type GetGearInput struct {
	AthleteID int64  `json:"-" adapter:"context"`
	GearID    string `json:"gear_id" adapter:"path,param=id"`
}

// CreateCustomGearInput contains parameters for creating custom gear.
type CreateCustomGearInput struct {
	AthleteID        int64
	Name             string   `json:"name"`
	Hashtag          string   `json:"hashtag"`
	Retired          bool     `json:"retired"`
	PurchasePrice    *float64 `json:"purchase_price,omitempty"`
	PurchaseCurrency string   `json:"purchase_currency,omitempty"`
}

// UpdateCustomGearInput contains parameters for updating custom gear.
type UpdateCustomGearInput struct {
	AthleteID        int64
	GearID           string
	Name             *string   `json:"name,omitempty"`
	Hashtag          *string   `json:"hashtag,omitempty"`
	Retired          *bool     `json:"retired,omitempty"`
	PurchasePrice    **float64 `json:"purchase_price,omitempty"`
	PurchaseCurrency *string   `json:"purchase_currency,omitempty"`
}

// DeleteCustomGearInput contains parameters for deleting custom gear.
type DeleteCustomGearInput struct {
	AthleteID int64
	GearID    string
	Force     bool
}

// DeleteGearOutput contains the result of a delete operation.
type DeleteGearOutput struct {
	Deleted bool `json:"deleted"`
}

// MonthlyUsageInput contains parameters for getting monthly gear usage.
type MonthlyUsageInput struct {
	AthleteID      int64 `json:"-" adapter:"context"`
	IncludeRetired bool  `json:"include_retired" adapter:"query"`
}

// ============================================================================
// Service Methods
// ============================================================================

// List returns a paginated list of gear for an athlete.
//adapter:wasm getGearList category=Gear
//adapter:http GET /api/v1/gear
func (s *GearService) List(ctx context.Context, in ListGearInput) (*ListGearOutput, error) {
	return s.listCommon(ctx, in, s.repo.ListPaginated)
}

// ListCustom returns a paginated list of custom gear for an athlete.
func (s *GearService) ListCustom(ctx context.Context, in ListGearInput) (*ListGearOutput, error) {
	return s.listCommon(ctx, in, s.repo.ListCustomPaginated)
}

// listCommon handles the common logic for listing gear.
type gearListFetcher func(ctx context.Context, athleteID int64, f storage.GearFilters) (storage.GearListResult, error)

func (s *GearService) listCommon(ctx context.Context, in ListGearInput, fetcher gearListFetcher) (*ListGearOutput, error) {
	f := storage.GearFilters{
		IncludeRetired: in.IncludeRetired,
		QueryParams: pagination.QueryParams{
			Page:     in.Page,
			PerPage:  in.PerPage,
			OrderBy:  in.OrderBy,
			OrderDir: in.OrderDir,
		},
	}

	result, err := fetcher(ctx, in.AthleteID, f)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to list gear: %v", err)
	}

	// Batch fetch activity counts to avoid N+1 queries
	gearIDs := make([]string, len(result.Items))
	for i, g := range result.Items {
		gearIDs[i] = g.ID
	}
	activityCounts, err := s.repo.GetActivityCountsBatch(ctx, gearIDs)
	if err != nil {
		slog.Error("failed to get activity counts", "error", err)
		activityCounts = make(map[string]int)
	}

	items := make([]GearItem, 0, len(result.Items))
	for _, g := range result.Items {
		items = append(items, gearToItem(g, activityCounts[g.ID]))
	}

	return &ListGearOutput{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}, nil
}

// GetByID retrieves a single gear item by ID.
//adapter:wasm getGearById category=Gear
//adapter:http GET /api/v1/gear/{id}
func (s *GearService) GetByID(ctx context.Context, in GetGearInput) (*GearItem, error) {
	if in.GearID == "" {
		return nil, BadRequest("gear ID required")
	}

	gear, err := s.repo.GetByID(ctx, in.GearID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch gear: %v", err)
	}
	if gear == nil {
		return nil, NotFound("gear")
	}

	// Authorization: verify gear belongs to the athlete
	if gear.AthleteID != in.AthleteID {
		return nil, Wrap(ErrForbidden, "access denied")
	}

	count, err := s.repo.GetActivityCount(ctx, in.AthleteID, in.GearID)
	if err != nil {
		slog.Error("failed to get activity count for gear", "gear_id", in.GearID, "error", err)
	}

	item := gearToItem(*gear, count)
	return &item, nil
}

// CreateCustom creates a new custom gear item.
func (s *GearService) CreateCustom(ctx context.Context, in CreateCustomGearInput) (*GearItem, error) {
	// Validate input
	if err := validateCustomGearRequest(in.Name, in.Hashtag, in.PurchaseCurrency, true); err != nil {
		return nil, err
	}

	// Check hashtag uniqueness
	existing, err := s.repo.GetCustomByHashtag(ctx, in.AthleteID, in.Hashtag)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to validate hashtag: %v", err)
	}
	if existing != nil {
		return nil, Conflict("hashtag already in use")
	}

	created, err := s.repo.CreateCustom(ctx, in.AthleteID, storage.CustomGearCreate{
		Name:             in.Name,
		Hashtag:          in.Hashtag,
		Retired:          in.Retired,
		PurchasePrice:    in.PurchasePrice,
		PurchaseCurrency: in.PurchaseCurrency,
	})
	if err != nil {
		return nil, Wrapf(ErrBadRequest, "%v", err)
	}

	count, err := s.repo.GetActivityCount(ctx, in.AthleteID, created.ID)
	if err != nil {
		slog.Error("failed to get activity count for gear", "gear_id", created.ID, "error", err)
	}

	item := gearToItem(*created, count)
	return &item, nil
}

// UpdateCustom updates an existing custom gear item.
func (s *GearService) UpdateCustom(ctx context.Context, in UpdateCustomGearInput) (*GearItem, error) {
	if in.GearID == "" {
		return nil, BadRequest("gear ID required")
	}

	// Validate updated fields
	if in.Name != nil {
		if err := validateGearName(*in.Name); err != nil {
			return nil, err
		}
	}
	if in.Hashtag != nil {
		if err := validateHashtag(*in.Hashtag); err != nil {
			return nil, err
		}
	}
	if in.PurchaseCurrency != nil {
		if err := validatePurchaseCurrency(*in.PurchaseCurrency); err != nil {
			return nil, err
		}
	}

	// Check hashtag uniqueness if being updated
	if in.Hashtag != nil {
		existing, err := s.repo.GetCustomByHashtag(ctx, in.AthleteID, *in.Hashtag)
		if err != nil {
			return nil, Wrapf(ErrInternal, "failed to validate hashtag: %v", err)
		}
		if existing != nil && existing.ID != in.GearID {
			return nil, Conflict("hashtag already in use")
		}
	}

	update := storage.CustomGearUpdate{
		Name:             in.Name,
		Hashtag:          in.Hashtag,
		Retired:          in.Retired,
		PurchaseCurrency: in.PurchaseCurrency,
		PurchasePrice:    in.PurchasePrice,
	}

	updated, err := s.repo.UpdateCustom(ctx, in.AthleteID, in.GearID, update)
	if err != nil {
		return nil, Wrapf(ErrBadRequest, "%v", err)
	}
	if updated == nil {
		return nil, NotFound("custom gear")
	}

	count, err := s.repo.GetActivityCount(ctx, in.AthleteID, updated.ID)
	if err != nil {
		slog.Error("failed to get activity count for gear", "gear_id", updated.ID, "error", err)
	}

	item := gearToItem(*updated, count)
	return &item, nil
}

// DeleteCustom deletes a custom gear item.
func (s *GearService) DeleteCustom(ctx context.Context, in DeleteCustomGearInput) (*DeleteGearOutput, error) {
	if in.GearID == "" {
		return nil, BadRequest("gear ID required")
	}

	hadActivities, err := s.repo.DeleteCustom(ctx, in.AthleteID, in.GearID, in.Force)
	if err != nil {
		return nil, Wrapf(ErrBadRequest, "%v", err)
	}
	if hadActivities && !in.Force {
		return nil, Conflict("gear is referenced by activities; use force=true to unlink and delete")
	}

	return &DeleteGearOutput{Deleted: true}, nil
}

// MonthlyUsage returns monthly usage statistics for gear.
//adapter:wasm getGearMonthlyStats category=Gear
//adapter:http GET /api/v1/gear/stats/monthly
func (s *GearService) MonthlyUsage(ctx context.Context, in MonthlyUsageInput) ([]storage.GearMonthlyUsage, error) {
	stats, err := s.repo.GetMonthlyUsage(ctx, in.AthleteID, in.IncludeRetired)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch gear stats: %v", err)
	}
	return stats, nil
}

// ============================================================================
// Validation Helpers
// ============================================================================

func validateGearName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return BadRequest("name is required")
	}
	if len(name) > MaxGearNameLength {
		return BadRequest(fmt.Sprintf("name must be %d characters or less", MaxGearNameLength))
	}
	return nil
}

func validateHashtag(hashtag string) error {
	hashtag = strings.TrimSpace(hashtag)
	if hashtag == "" {
		return BadRequest("hashtag is required")
	}
	if len(hashtag) > MaxHashtagLength {
		return BadRequest(fmt.Sprintf("hashtag must be %d characters or less", MaxHashtagLength))
	}
	if !hashtagRe.MatchString(hashtag) {
		return BadRequest("hashtag must be alphanumeric (may include underscores and hyphens)")
	}
	return nil
}

func validatePurchaseCurrency(currency string) error {
	currency = strings.TrimSpace(currency)
	if currency == "" {
		return nil // optional field
	}
	if len(currency) != 3 {
		return BadRequest("currency must be a 3-letter ISO code (e.g., USD, EUR)")
	}
	return nil
}

func validateCustomGearRequest(name, hashtag, currency string, isCreate bool) error {
	if isCreate || name != "" {
		if err := validateGearName(name); err != nil {
			return err
		}
	}
	if isCreate || hashtag != "" {
		if err := validateHashtag(hashtag); err != nil {
			return err
		}
	}
	if err := validatePurchaseCurrency(currency); err != nil {
		return err
	}
	return nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

func gearToItem(g storage.Gear, activityCount int) GearItem {
	hashtag := strings.TrimSpace(g.Hashtag)
	if hashtag != "" && !strings.HasPrefix(hashtag, "#") {
		hashtag = "#" + hashtag
	}
	return GearItem{
		ID:               g.ID,
		Name:             g.Name,
		Primary:          g.Primary,
		Retired:          g.Retired,
		Distance:         g.Distance,
		BrandName:        g.BrandName,
		ModelName:        g.ModelName,
		Description:      g.Description,
		Source:           g.Source,
		Hashtag:          hashtag,
		PurchasePrice:    g.PurchasePrice,
		PurchaseCurrency: g.PurchaseCurrency,
		ActivityCount:    activityCount,
	}
}
