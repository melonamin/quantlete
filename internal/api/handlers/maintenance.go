package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type componentsListResponse struct {
	Data       []storage.ComponentWithRules `json:"data"`
	Total      int                          `json:"total"`
	Page       int                          `json:"page"`
	PerPage    int                          `json:"per_page"`
	TotalPages int                          `json:"total_pages"`
}

type MaintenanceHandler struct {
	repo   *storage.MaintenanceRepository
	strava *strava.Client
}

func NewMaintenanceHandler(repo *storage.MaintenanceRepository, stravaClient *strava.Client) *MaintenanceHandler {
	return &MaintenanceHandler{repo: repo, strava: stravaClient}
}

type maintenanceRuleInput struct {
	Type           string  `json:"type"`
	ThresholdValue float64 `json:"threshold_value"`
}

type createComponentRequest struct {
	Name               string                 `json:"name"`
	ImageURL           string                 `json:"image_url,omitempty"`
	MaintenanceHashtag string                 `json:"maintenance_hashtag,omitempty"`
	Rules              []maintenanceRuleInput `json:"rules,omitempty"`
}

// ListGearComponents handles GET /api/v1/gear/{id}/components
func (h *MaintenanceHandler) ListGearComponents(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	gearID := chi.URLParam(r, "id")
	if gearID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gear id required"})
		return
	}

	q := r.URL.Query()
	f := storage.ComponentFilters{
		QueryParams: pagination.ParseQueryParams(q),
	}

	result, err := h.repo.ListComponentsPaginated(r.Context(), athlete.ID, gearID, f)
	if err != nil {
		slog.Error("failed to list gear components", "error", err, "athlete_id", athlete.ID, "gear_id", gearID, "filters", f)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch components"})
		return
	}

	items := result.Items
	if items == nil {
		items = []storage.ComponentWithRules{}
	}

	writeJSON(w, http.StatusOK, componentsListResponse{
		Data:       items,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
}

// CreateGearComponent handles POST /api/v1/gear/{id}/components
func (h *MaintenanceHandler) CreateGearComponent(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	gearID := chi.URLParam(r, "id")
	if gearID == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gear id required"})
		return
	}

	var req createComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	rules := make([]storage.CreateRuleInput, 0, len(req.Rules))
	for _, rule := range req.Rules {
		rules = append(rules, storage.CreateRuleInput{
			Type:           rule.Type,
			ThresholdValue: rule.ThresholdValue,
		})
	}

	created, err := h.repo.CreateComponent(r.Context(), athlete.ID, gearID, storage.CreateComponentInput{
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if created == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "gear not found"})
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

type updateComponentRequest struct {
	Name               *string                 `json:"name,omitempty"`
	ImageURL           *string                 `json:"image_url,omitempty"`
	MaintenanceHashtag *string                 `json:"maintenance_hashtag,omitempty"`
	Rules              *[]maintenanceRuleInput `json:"rules,omitempty"`
}

// UpdateComponent handles PUT /api/v1/components/{id}
func (h *MaintenanceHandler) UpdateComponent(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid component id"})
		return
	}

	var req updateComponentRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	var rules *[]storage.CreateRuleInput
	if req.Rules != nil {
		out := make([]storage.CreateRuleInput, 0, len(*req.Rules))
		for _, rule := range *req.Rules {
			out = append(out, storage.CreateRuleInput{
				Type:           rule.Type,
				ThresholdValue: rule.ThresholdValue,
			})
		}
		rules = &out
	}

	updated, err := h.repo.UpdateComponent(r.Context(), athlete.ID, id, storage.UpdateComponentInput{
		Name:               req.Name,
		ImageURL:           req.ImageURL,
		MaintenanceHashtag: req.MaintenanceHashtag,
		Rules:              rules,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if updated == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "component not found"})
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// DeleteComponent handles DELETE /api/v1/components/{id}
func (h *MaintenanceHandler) DeleteComponent(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid component id"})
		return
	}

	if err := h.repo.DeleteComponent(r.Context(), athlete.ID, id); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to delete component"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

type logMaintenanceRequest struct {
	ActivityID  *int64  `json:"activity_id,omitempty"`
	CompletedAt *string `json:"completed_at,omitempty"` // RFC3339 or YYYY-MM-DD
}

// LogMaintenance handles POST /api/v1/components/{id}/maintenance
func (h *MaintenanceHandler) LogMaintenance(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid component id"})
		return
	}

	var req logMaintenanceRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
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
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid completed_at"})
			return
		}
	}

	if err := h.repo.LogMaintenance(r.Context(), athlete.ID, id, req.ActivityID, completed); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to log maintenance"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"logged": true})
}

// Due handles GET /api/v1/maintenance/due
func (h *MaintenanceHandler) Due(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	items, err := h.repo.Due(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch due maintenance"})
		return
	}
	if items == nil {
		items = []storage.DueComponent{}
	}
	writeJSON(w, http.StatusOK, items)
}
