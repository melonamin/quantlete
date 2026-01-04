package services

import (
	"context"

	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/storage"
)

// ChallengesService handles challenges business logic.
type ChallengesService struct {
	repo *storage.ChallengeRepository
}

// NewChallengesService creates a new challenges service.
func NewChallengesService(repo *storage.ChallengeRepository) *ChallengesService {
	return &ChallengesService{repo: repo}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// ListChallengesInput contains parameters for listing challenges.
type ListChallengesInput struct {
	AthleteID int64  `json:"-" adapter:"context"`
	Month     string `json:"month" adapter:"query"`
	Page      int    `json:"page" adapter:"query,default=1"`
	PerPage   int    `json:"per_page" adapter:"query,default=50"`
	OrderBy   string `json:"order_by" adapter:"query,default=completion_date"`
	OrderDir  string `json:"order_dir" adapter:"query,default=desc"`
}

// ChallengeItem represents a challenge item in responses.
type ChallengeItem struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug,omitempty"`
	BadgeURL       string  `json:"badge_url,omitempty"`
	LocalBadgeURL  string  `json:"local_badge_url,omitempty"`
	CompletionDate *string `json:"completion_date,omitempty"` // YYYY-MM-DD format
	Month          string  `json:"month,omitempty"`
}

// ListChallengesOutput contains a paginated list of challenges.
type ListChallengesOutput struct {
	Data       []ChallengeItem `json:"data"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

// ============================================================================
// Service Methods
// ============================================================================

// Import stores a challenge in the database.
func (s *ChallengesService) Import(ctx context.Context, c *storage.Challenge) error {
	if err := s.repo.Upsert(ctx, c); err != nil {
		return Wrapf(ErrInternal, "failed to import challenge: %v", err)
	}
	return nil
}

// List returns a paginated list of challenges for an athlete.
//
//adapter:wasm getChallenges category=Challenges
//adapter:http GET /api/v1/challenges
func (s *ChallengesService) List(ctx context.Context, in ListChallengesInput) (*ListChallengesOutput, error) {
	f := storage.ChallengeFilters{
		Month: in.Month,
		QueryParams: pagination.QueryParams{
			Page:     in.Page,
			PerPage:  in.PerPage,
			OrderBy:  in.OrderBy,
			OrderDir: in.OrderDir,
		},
	}

	result, err := s.repo.ListPaginated(ctx, in.AthleteID, f)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to list challenges: %v", err)
	}

	items := make([]ChallengeItem, 0, len(result.Items))
	for _, c := range result.Items {
		items = append(items, challengeToItem(c))
	}

	return &ListChallengesOutput{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	}, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

// challengeToItem converts a storage.Challenge to a ChallengeItem.
func challengeToItem(c storage.Challenge) ChallengeItem {
	var completionDate *string
	if c.CompletionDate != nil && !c.CompletionDate.IsZero() {
		v := c.CompletionDate.Format("2006-01-02")
		completionDate = &v
	}

	return ChallengeItem{
		ID:             c.ID,
		Name:           c.Name,
		Slug:           c.Slug,
		BadgeURL:       c.BadgeURL,
		LocalBadgeURL:  c.LocalBadgeURL,
		CompletionDate: completionDate,
		Month:          c.Month,
	}
}
