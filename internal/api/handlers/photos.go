package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

// PhotosHandler handles photo-related endpoints.
type PhotosHandler struct {
	svc    *services.PhotosService
	strava *strava.Client
}

// NewPhotosHandler creates a new photos handler.
func NewPhotosHandler(svc *services.PhotosService, stravaClient *strava.Client) *PhotosHandler {
	return &PhotosHandler{svc: svc, strava: stravaClient}
}

// List handles GET /api/v1/photos
func (h *PhotosHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	q := r.URL.Query()

	// Parse sport types from comma-separated query param
	var sportTypes []string
	if sport := strings.TrimSpace(q.Get("sport_type")); sport != "" {
		parts := strings.Split(sport, ",")
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				sportTypes = append(sportTypes, s)
			}
		}
	}

	// Parse pagination params
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

	result, err := h.svc.List(r.Context(), services.ListPhotosInput{
		AthleteID:  athlete.ID,
		SportTypes: sportTypes,
		Country:    strings.TrimSpace(q.Get("country")),
		Page:       page,
		PerPage:    perPage,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// ActivityPhotos handles GET /api/v1/activities/{id}/photos
func (h *PhotosHandler) ActivityPhotos(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid activity id"))
		return
	}

	photos, err := h.svc.ListByActivity(r.Context(), services.GetActivityPhotosInput{
		AthleteID:  athlete.ID,
		ActivityID: id,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, photos)
}
