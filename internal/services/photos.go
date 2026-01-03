package services

import (
	"context"
	"time"

	"github.com/melonamin/quantlete/internal/geo"
	"github.com/melonamin/quantlete/internal/storage"
)

// PhotosService handles photo business logic.
type PhotosService struct {
	repo *storage.PhotoRepository
}

// NewPhotosService creates a new photos service.
func NewPhotosService(repo *storage.PhotoRepository) *PhotosService {
	return &PhotosService{repo: repo}
}

// ============================================================================
// Input/Output Types
// ============================================================================

// ListPhotosInput contains parameters for listing photos.
type ListPhotosInput struct {
	AthleteID  int64
	SportTypes []string
	Country    string
	Page       int
	PerPage    int
}

// PhotoItem represents a photo in responses.
type PhotoItem struct {
	ID              string `json:"id"`
	ActivityID      int64  `json:"activity_id"`
	URL             string `json:"url"`
	ThumbnailURL    string `json:"thumbnail_url,omitempty"`
	Caption         string `json:"caption,omitempty"`
	CreatedAt       string `json:"created_at"`
	ActivityName    string `json:"activity_name"`
	SportType       string `json:"sport_type"`
	StartDateLocal  string `json:"start_date_local"`
	LocationCountry string `json:"location_country,omitempty"`
}

// FacetItem represents a facet value with count for filtering.
type FacetItem struct {
	Value string `json:"value"`
	ISO2  string `json:"iso2,omitempty"`
	Count int    `json:"count"`
}

// ListPhotosOutput contains a paginated list of photos with facets.
type ListPhotosOutput struct {
	Data       []PhotoItem `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
	Countries  []FacetItem `json:"countries"`
	SportTypes []FacetItem `json:"sport_types"`
}

// GetActivityPhotosInput contains parameters for getting photos of an activity.
type GetActivityPhotosInput struct {
	AthleteID  int64
	ActivityID int64
}

// ActivityPhotoItem represents a photo attached to an activity.
type ActivityPhotoItem struct {
	ID           string `json:"id"`
	ActivityID   int64  `json:"activity_id"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
	Caption      string `json:"caption,omitempty"`
	CreatedAt    string `json:"created_at"`
}

// ============================================================================
// Service Methods
// ============================================================================

// List returns a paginated list of photos for an athlete with facets.
func (s *PhotosService) List(ctx context.Context, in ListPhotosInput) (*ListPhotosOutput, error) {
	// Apply defaults
	page := in.Page
	if page <= 0 {
		page = 1
	}
	perPage := in.PerPage
	if perPage <= 0 {
		perPage = 60
	}
	if perPage > 200 {
		perPage = 200
	}

	filters := storage.PhotoListFilters{
		SportTypes: in.SportTypes,
		Country:    in.Country,
	}

	result, err := s.repo.List(ctx, in.AthleteID, filters, page, perPage)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch photos: %v", err)
	}

	// Handle nil result (empty state)
	if result == nil {
		return &ListPhotosOutput{
			Data:       []PhotoItem{},
			Total:      0,
			Page:       page,
			PerPage:    perPage,
			TotalPages: 0,
			Countries:  []FacetItem{},
			SportTypes: []FacetItem{},
		}, nil
	}

	items := make([]PhotoItem, 0, len(result.Items))
	for _, it := range result.Items {
		items = append(items, photoListItemToItem(it))
	}

	// Convert country facets with ISO2 lookup
	countries := make([]FacetItem, 0, len(result.Countries))
	for _, c := range result.Countries {
		item := FacetItem{Value: c.Value, Count: c.Count}
		if iso2, ok := geo.ISO2FromCountry(c.Value); ok {
			item.ISO2 = iso2
		}
		countries = append(countries, item)
	}

	// Convert sport type facets
	sportTypes := make([]FacetItem, 0, len(result.SportTypes))
	for _, st := range result.SportTypes {
		sportTypes = append(sportTypes, FacetItem{Value: st.Value, Count: st.Count})
	}

	totalPages := (result.Total + perPage - 1) / perPage

	return &ListPhotosOutput{
		Data:       items,
		Total:      result.Total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		Countries:  countries,
		SportTypes: sportTypes,
	}, nil
}

// ListByActivity returns photos for a specific activity.
func (s *PhotosService) ListByActivity(ctx context.Context, in GetActivityPhotosInput) ([]ActivityPhotoItem, error) {
	if in.ActivityID == 0 {
		return nil, BadRequest("activity ID required")
	}

	photos, err := s.repo.ListByActivity(ctx, in.AthleteID, in.ActivityID)
	if err != nil {
		return nil, Wrapf(ErrInternal, "failed to fetch activity photos: %v", err)
	}

	if photos == nil {
		return []ActivityPhotoItem{}, nil
	}

	items := make([]ActivityPhotoItem, 0, len(photos))
	for _, p := range photos {
		items = append(items, photoToActivityItem(p))
	}

	return items, nil
}

// ============================================================================
// Mapping Helpers
// ============================================================================

func photoListItemToItem(p storage.PhotoListItem) PhotoItem {
	return PhotoItem{
		ID:              p.ID,
		ActivityID:      p.ActivityID,
		URL:             p.URL,
		ThumbnailURL:    p.ThumbnailURL,
		Caption:         p.Caption,
		CreatedAt:       p.CreatedAt.Format(time.RFC3339),
		ActivityName:    p.ActivityName,
		SportType:       p.SportType,
		StartDateLocal:  p.StartDateLocal.Format(time.RFC3339),
		LocationCountry: p.LocationCountry,
	}
}

func photoToActivityItem(p storage.Photo) ActivityPhotoItem {
	return ActivityPhotoItem{
		ID:           p.ID,
		ActivityID:   p.ActivityID,
		URL:          p.URL,
		ThumbnailURL: p.ThumbnailURL,
		Caption:      p.Caption,
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
	}
}
