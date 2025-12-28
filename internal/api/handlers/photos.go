package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/sasha/stata/internal/geo"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

type PhotosHandler struct {
	photos *storage.PhotoRepository
	strava *strava.Client
}

func NewPhotosHandler(photos *storage.PhotoRepository, stravaClient *strava.Client) *PhotosHandler {
	return &PhotosHandler{photos: photos, strava: stravaClient}
}

type photoResponse struct {
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

type facetResponse struct {
	Value string `json:"value"`
	ISO2  string `json:"iso2,omitempty"`
	Count int    `json:"count"`
}

type photosListResponse struct {
	Data       []photoResponse `json:"data"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
	Countries  []facetResponse `json:"countries"`
	SportTypes []facetResponse `json:"sport_types"`
}

// List handles GET /api/v1/photos
func (h *PhotosHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	q := r.URL.Query()
	var filters storage.PhotoListFilters
	if sport := strings.TrimSpace(q.Get("sport_type")); sport != "" {
		parts := strings.Split(sport, ",")
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				filters.SportTypes = append(filters.SportTypes, s)
			}
		}
	}
	filters.Country = strings.TrimSpace(q.Get("country"))

	page := 1
	if v := q.Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			page = n
		}
	}
	perPage := 60
	if v := q.Get("per_page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			perPage = n
		}
	}

	result, err := h.photos.List(r.Context(), athlete.ID, filters, page, perPage)
	if err != nil {
		slog.Error("failed to fetch photos", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch photos"})
		return
	}
	if result == nil {
		writeJSON(w, http.StatusOK, photosListResponse{
			Data:       []photoResponse{},
			Total:      0,
			Page:       page,
			PerPage:    perPage,
			TotalPages: 0,
			Countries:  []facetResponse{},
			SportTypes: []facetResponse{},
		})
		return
	}

	respItems := make([]photoResponse, 0, len(result.Items))
	for _, it := range result.Items {
		respItems = append(respItems, photoResponse{
			ID:              it.ID,
			ActivityID:      it.ActivityID,
			URL:             it.URL,
			ThumbnailURL:    it.ThumbnailURL,
			Caption:         it.Caption,
			CreatedAt:       it.CreatedAt.Format(time.RFC3339),
			ActivityName:    it.ActivityName,
			SportType:       it.SportType,
			StartDateLocal:  it.StartDateLocal.Format(time.RFC3339),
			LocationCountry: it.LocationCountry,
		})
	}

	countries := make([]facetResponse, 0, len(result.Countries))
	for _, c := range result.Countries {
		item := facetResponse{Value: c.Value, Count: c.Count}
		if iso2, ok := geo.ISO2FromCountry(c.Value); ok {
			item.ISO2 = iso2
		}
		countries = append(countries, item)
	}
	sportTypes := make([]facetResponse, 0, len(result.SportTypes))
	for _, s := range result.SportTypes {
		sportTypes = append(sportTypes, facetResponse{Value: s.Value, Count: s.Count})
	}

	totalPages := (result.Total + perPage - 1) / perPage
	writeJSON(w, http.StatusOK, photosListResponse{
		Data:       respItems,
		Total:      result.Total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		Countries:  countries,
		SportTypes: sportTypes,
	})
}

// ActivityPhotos handles GET /api/v1/activities/{id}/photos
func (h *PhotosHandler) ActivityPhotos(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid activity id"})
		return
	}

	photos, err := h.photos.ListByActivity(r.Context(), athlete.ID, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch photos"})
		return
	}
	if photos == nil {
		photos = []storage.Photo{}
	}
	writeJSON(w, http.StatusOK, photos)
}
