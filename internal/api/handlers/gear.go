package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

// GearHandler handles gear-related endpoints.
type GearHandler struct {
	svc    *services.GearService
	strava *strava.Client
}

// NewGearHandler creates a new gear handler.
func NewGearHandler(svc *services.GearService, stravaClient *strava.Client) *GearHandler {
	return &GearHandler{
		svc:    svc,
		strava: stravaClient,
	}
}

// List handles GET /api/v1/gear
func (h *GearHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	q := r.URL.Query()
	params := pagination.ParseQueryParams(q)

	result, err := h.svc.List(r.Context(), services.ListGearInput{
		AthleteID:      athlete.ID,
		IncludeRetired: q.Get("include_retired") == "true",
		Page:           params.Page,
		PerPage:        params.PerPage,
		OrderBy:        params.OrderBy,
		OrderDir:       params.OrderDir,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// GetByID handles GET /api/v1/gear/{id}
func (h *GearHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	result, err := h.svc.GetByID(r.Context(), services.GetGearInput{
		AthleteID: athlete.ID,
		GearID:    chi.URLParam(r, "id"),
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// ListCustom handles GET /api/v1/gear/custom
func (h *GearHandler) ListCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	q := r.URL.Query()
	params := pagination.ParseQueryParams(q)

	result, err := h.svc.ListCustom(r.Context(), services.ListGearInput{
		AthleteID:      athlete.ID,
		IncludeRetired: q.Get("include_retired") == "true",
		Page:           params.Page,
		PerPage:        params.PerPage,
		OrderBy:        params.OrderBy,
		OrderDir:       params.OrderDir,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// CreateCustom handles POST /api/v1/gear/custom
func (h *GearHandler) CreateCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var req services.CreateCustomGearInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}
	req.AthleteID = athlete.ID

	result, err := h.svc.CreateCustom(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteJSONResponse(w, http.StatusCreated, shared.SuccessResponse(result))
}

// UpdateCustom handles PUT /api/v1/gear/custom/{id}
func (h *GearHandler) UpdateCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	// Parse raw JSON to handle optional fields properly
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	input := services.UpdateCustomGearInput{
		AthleteID: athlete.ID,
		GearID:    chi.URLParam(r, "id"),
	}

	if v, ok := raw["name"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			input.Name = &s
		}
	}
	if v, ok := raw["hashtag"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			input.Hashtag = &s
		}
	}
	if v, ok := raw["retired"]; ok {
		var b bool
		if err := json.Unmarshal(v, &b); err == nil {
			input.Retired = &b
		}
	}
	if v, ok := raw["purchase_currency"]; ok {
		var s string
		if err := json.Unmarshal(v, &s); err == nil {
			input.PurchaseCurrency = &s
		}
	}
	if v, ok := raw["purchase_price"]; ok {
		var p *float64
		if string(v) != "null" {
			var pv float64
			if err := json.Unmarshal(v, &pv); err != nil {
				shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid purchase_price"))
				return
			}
			p = &pv
		}
		input.PurchasePrice = &p
	}

	result, err := h.svc.UpdateCustom(r.Context(), input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// DeleteCustom handles DELETE /api/v1/gear/custom/{id}
func (h *GearHandler) DeleteCustom(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	force := r.URL.Query().Get("force") == "true" || r.URL.Query().Get("force") == "1"

	result, err := h.svc.DeleteCustom(r.Context(), services.DeleteCustomGearInput{
		AthleteID: athlete.ID,
		GearID:    chi.URLParam(r, "id"),
		Force:     force,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// MonthlyUsage handles GET /api/v1/gear/stats/monthly
func (h *GearHandler) MonthlyUsage(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	result, err := h.svc.MonthlyUsage(r.Context(), services.MonthlyUsageInput{
		AthleteID:      athlete.ID,
		IncludeRetired: r.URL.Query().Get("include_retired") == "true",
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}
