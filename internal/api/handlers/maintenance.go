package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

// MaintenanceHandler handles maintenance-related HTTP requests.
type MaintenanceHandler struct {
	svc    *services.MaintenanceService
	strava *strava.Client
}

// NewMaintenanceHandler creates a new maintenance handler.
func NewMaintenanceHandler(svc *services.MaintenanceService, stravaClient *strava.Client) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc, strava: stravaClient}
}

type createComponentRequest struct {
	Name               string               `json:"name"`
	ImageURL           string               `json:"image_url,omitempty"`
	MaintenanceHashtag string               `json:"maintenance_hashtag,omitempty"`
	Rules              []services.RuleInput `json:"rules,omitempty"`
}

// ListGearComponents handles GET /api/v1/gear/{id}/components
func (h *MaintenanceHandler) ListGearComponents(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	gearID := chi.URLParam(r, "id")
	q := r.URL.Query()
	params := pagination.ParseQueryParams(q)

	result, err := h.svc.ListComponents(r.Context(), services.ListComponentsInput{
		AthleteID: athlete.ID,
		GearID:    gearID,
		Page:      params.Page,
		PerPage:   params.PerPage,
		OrderBy:   params.OrderBy,
		OrderDir:  params.OrderDir,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// CreateGearComponent handles POST /api/v1/gear/{id}/components
func (h *MaintenanceHandler) CreateGearComponent(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	gearID := chi.URLParam(r, "id")

	var req createComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	created, err := h.svc.CreateComponent(r.Context(), services.CreateComponentInput{
		AthleteID:          athlete.ID,
		GearID:             gearID,
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              req.Rules,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteJSONResponse(w, http.StatusCreated, shared.SuccessResponse(created))
}

type updateComponentRequest struct {
	Name               *string               `json:"name,omitempty"`
	ImageURL           *string               `json:"image_url,omitempty"`
	MaintenanceHashtag *string               `json:"maintenance_hashtag,omitempty"`
	Rules              *[]services.RuleInput `json:"rules,omitempty"`
}

// UpdateComponent handles PUT /api/v1/components/{id}
func (h *MaintenanceHandler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid component id"))
		return
	}

	var req updateComponentRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	updated, err := h.svc.UpdateComponent(r.Context(), services.UpdateComponentInput{
		AthleteID:          athlete.ID,
		ComponentID:        id,
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              req.Rules,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, updated)
}

// DeleteComponent handles DELETE /api/v1/components/{id}
func (h *MaintenanceHandler) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid component id"))
		return
	}

	if err := h.svc.DeleteComponent(r.Context(), athlete.ID, id); err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, map[string]any{"deleted": true})
}

type logMaintenanceRequest struct {
	ActivityID  *int64  `json:"activity_id,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"` // RFC3339 or YYYY-MM-DD
}

// LogMaintenance handles POST /api/v1/components/{id}/maintenance
func (h *MaintenanceHandler) LogMaintenance(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid component id"))
		return
	}

	var req logMaintenanceRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
			return
		}
	}

	completed := time.Now()
	if req.CompletedAt != nil && *req.CompletedAt != "" {
		if t, err := time.Parse(time.RFC3339, *req.CompletedAt); err == nil {
			completed = t
		} else if t, err := time.Parse("2006-01-02", *req.CompletedAt); err == nil {
			completed = t
		} else {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid completed_at"))
			return
		}
	}

	if err := h.svc.LogMaintenance(r.Context(), services.LogMaintenanceInput{
		AthleteID:   athlete.ID,
		ComponentID: id,
		ActivityID:  req.ActivityID,
		CompletedAt: completed,
	}); err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, map[string]any{"logged": true})
}

// Due handles GET /api/v1/maintenance/due
func (h *MaintenanceHandler) Due(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	items, err := h.svc.ListDue(r.Context(), athlete.ID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, items)
}
