package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/sasha/stata/internal/pagination"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

type gearListResponse struct {
	Data       []GearResponse `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PerPage    int            `json:"per_page"`
	TotalPages int            `json:"total_pages"`
}

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

func gearToResponse(g storage.Gear, activityCount int) GearResponse {
	hashtag := strings.TrimSpace(g.Hashtag)
	if hashtag != "" && !strings.HasPrefix(hashtag, "#") {
		hashtag = "#" + hashtag
	}
	return GearResponse{
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

// List handles GET /api/v1/gear
func (h *GearHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	q := r.URL.Query()
	f := storage.GearFilters{
		IncludeRetired: q.Get("include_retired") == "true",
		QueryParams:    pagination.ParseQueryParams(q),
	}

	result, err := h.repo.ListPaginated(r.Context(), athlete.ID, f)
	if err != nil {
		slog.Error("failed to list gear", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch gear"})
		return
	}

	// Batch fetch activity counts to avoid N+1 queries
	gearIDs := make([]string, len(result.Items))
	for i, g := range result.Items {
		gearIDs[i] = g.ID
	}
	activityCounts, err := h.repo.GetActivityCountsBatch(r.Context(), gearIDs)
	if err != nil {
		slog.Error("failed to get activity counts", "error", err, "athlete_id", athlete.ID)
		activityCounts = make(map[string]int)
	}

	responses := make([]GearResponse, 0, len(result.Items))
	for _, g := range result.Items {
		responses = append(responses, gearToResponse(g, activityCounts[g.ID]))
	}

	writeJSON(w, http.StatusOK, gearListResponse{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
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

	// Authorization: verify gear belongs to the authenticated athlete
	if gear.AthleteID != athlete.ID {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "access denied"})
		return
	}

	count, err := h.repo.GetActivityCount(r.Context(), gearID)
	if err != nil {
		slog.Error("failed to get activity count for gear", "gear_id", gearID, "error", err)
	}

	writeJSON(w, http.StatusOK, gearToResponse(*gear, count))
}

type CustomGearCreateRequest struct {
	Name             string   `json:"name"`
	Hashtag          string   `json:"hashtag"`
	Retired          bool     `json:"retired"`
	PurchasePrice    *float64 `json:"purchase_price,omitempty"`
	PurchaseCurrency string   `json:"purchase_currency,omitempty"`
}

// ListCustom handles GET /api/v1/gear/custom
func (h *GearHandler) ListCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	q := r.URL.Query()
	f := storage.GearFilters{
		IncludeRetired: q.Get("include_retired") == "true",
		QueryParams:    pagination.ParseQueryParams(q),
	}

	result, err := h.repo.ListCustomPaginated(r.Context(), athlete.ID, f)
	if err != nil {
		slog.Error("failed to list custom gear", "error", err, "athlete_id", athlete.ID)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch custom gear"})
		return
	}

	// Batch fetch activity counts to avoid N+1 queries
	gearIDs := make([]string, len(result.Items))
	for i, g := range result.Items {
		gearIDs[i] = g.ID
	}
	activityCounts, err := h.repo.GetActivityCountsBatch(r.Context(), gearIDs)
	if err != nil {
		slog.Error("failed to get activity counts", "error", err, "athlete_id", athlete.ID)
		activityCounts = make(map[string]int)
	}

	responses := make([]GearResponse, 0, len(result.Items))
	for _, g := range result.Items {
		responses = append(responses, gearToResponse(g, activityCounts[g.ID]))
	}

	writeJSON(w, http.StatusOK, gearListResponse{
		Data:       responses,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
}

// CreateCustom handles POST /api/v1/gear/custom
func (h *GearHandler) CreateCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var req CustomGearCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	// Validate input
	if err := ValidateCustomGearRequest(req.Name, req.Hashtag, req.PurchaseCurrency, true); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	if existing, err := h.repo.GetCustomByHashtag(r.Context(), athlete.ID, req.Hashtag); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to validate hashtag"})
		return
	} else if existing != nil {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "hashtag already in use"})
		return
	}

	created, err := h.repo.CreateCustom(r.Context(), athlete.ID, storage.CustomGearCreate{
		Name:             req.Name,
		Hashtag:          req.Hashtag,
		Retired:          req.Retired,
		PurchasePrice:    req.PurchasePrice,
		PurchaseCurrency: req.PurchaseCurrency,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	count, err := h.repo.GetActivityCount(r.Context(), created.ID)
	if err != nil {
		slog.Error("failed to get activity count for gear", "gear_id", created.ID, "error", err)
	}
	writeJSON(w, http.StatusCreated, gearToResponse(*created, count))
}

// UpdateCustom handles PUT /api/v1/gear/custom/{id}
func (h *GearHandler) UpdateCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gear id required"})
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	var update storage.CustomGearUpdate
	if v, ok := raw["name"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			update.Name = &s
		}
	}
	if v, ok := raw["hashtag"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			update.Hashtag = &s
		}
	}
	if v, ok := raw["retired"]; ok {
		var b bool
		if err := json.Unmarshal(v, &b); err == nil {
			update.Retired = &b
		}
	}
	if v, ok := raw["purchase_currency"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			update.PurchaseCurrency = &s
		}
	}
	if v, ok := raw["purchase_price"]; ok {
		var p *float64
		if string(v) != "null" {
			var pv float64
			if err := json.Unmarshal(v, &pv); err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid purchase_price"})
				return
			}
			p = &pv
		}
		update.PurchasePrice = &p
	}

	// Validate updated fields
	if update.Name != nil {
		if err := ValidateGearName(*update.Name); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}
	if update.Hashtag != nil {
		if err := ValidateHashtag(*update.Hashtag); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}
	if update.PurchaseCurrency != nil {
		if err := ValidatePurchaseCurrency(*update.PurchaseCurrency); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
	}

	if update.Hashtag != nil {
		if existing, err := h.repo.GetCustomByHashtag(r.Context(), athlete.ID, *update.Hashtag); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to validate hashtag"})
			return
		} else if existing != nil && existing.ID != id {
			writeJSON(w, http.StatusConflict, ErrorResponse{Error: "hashtag already in use"})
			return
		}
	}

	updated, err := h.repo.UpdateCustom(r.Context(), athlete.ID, id, update)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if updated == nil {
		writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "custom gear not found"})
		return
	}

	count, err := h.repo.GetActivityCount(r.Context(), updated.ID)
	if err != nil {
		slog.Error("failed to get activity count for gear", "gear_id", updated.ID, "error", err)
	}
	writeJSON(w, http.StatusOK, gearToResponse(*updated, count))
}

// DeleteCustom handles DELETE /api/v1/gear/custom/{id}
func (h *GearHandler) DeleteCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "gear id required"})
		return
	}

	force := r.URL.Query().Get("force") == "true" || r.URL.Query().Get("force") == "1"
	hadActivities, err := h.repo.DeleteCustom(r.Context(), athlete.ID, id, force)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}
	if hadActivities && !force {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "gear is referenced by activities; use ?force=true to unlink and delete"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"deleted": true})
}

// MonthlyUsage handles GET /api/v1/gear/stats/monthly
func (h *GearHandler) MonthlyUsage(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	includeRetired := r.URL.Query().Get("include_retired") == "true"
	stats, err := h.repo.GetMonthlyUsage(r.Context(), athlete.ID, includeRetired)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch gear stats"})
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
