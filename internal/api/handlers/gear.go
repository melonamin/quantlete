package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// GearHandler handles gear-related endpoints.
type GearHandler struct {
	repo   *storage.GearRepository
	strava *strava.Client
}

// NewGearHandler creates a new gear handler.
func NewGearHandler(repo *storage.GearRepository, stravaClient *strava.Client) *GearHandler {
	return &GearHandler{
		repo:   repo,
		strava: stravaClient,
	}
}

// GearResponse represents gear in API responses.
type GearResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Primary       bool    `json:"primary"`
	Retired       bool    `json:"retired"`
	Distance      float64 `json:"distance"`
	BrandName     string  `json:"brand_name,omitempty"`
	ModelName     string  `json:"model_name,omitempty"`
	Description   string  `json:"description,omitempty"`
	ActivityCount int     `json:"activity_count"`
}

// List handles GET /api/v1/gear
func (h *GearHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	includeRetired := r.URL.Query().Get("include_retired") == "true"

	gear, err := h.repo.GetByAthleteID(r.Context(), athlete.ID, includeRetired)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch gear"})
		return
	}

	responses := make([]GearResponse, len(gear))
	for i, g := range gear {
		count, _ := h.repo.GetActivityCount(r.Context(), g.ID)
		responses[i] = GearResponse{
			ID:            g.ID,
			Name:          g.Name,
			Primary:       g.Primary,
			Retired:       g.Retired,
			Distance:      g.Distance,
			BrandName:     g.BrandName,
			ModelName:     g.ModelName,
			Description:   g.Description,
			ActivityCount: count,
		}
	}

	writeJSON(w, http.StatusOK, responses)
}

// GetByID handles GET /api/v1/gear/{id}
func (h *GearHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	gearID := chi.URLParam(r, "id")
	if gearID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gear ID required"})
		return
	}

	gear, err := h.repo.GetByID(r.Context(), gearID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch gear"})
		return
	}

	if gear == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "gear not found"})
		return
	}

	count, _ := h.repo.GetActivityCount(r.Context(), gearID)

	writeJSON(w, http.StatusOK, GearResponse{
		ID:            gear.ID,
		Name:          gear.Name,
		Primary:       gear.Primary,
		Retired:       gear.Retired,
		Distance:      gear.Distance,
		BrandName:     gear.BrandName,
		ModelName:     gear.ModelName,
		Description:   gear.Description,
		ActivityCount: count,
	})
}
