package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type SettingsHandler struct {
	settings *storage.SettingsRepository
	strava   *strava.Client
}

func NewSettingsHandler(settings *storage.SettingsRepository, stravaClient *strava.Client) *SettingsHandler {
	return &SettingsHandler{settings: settings, strava: stravaClient}
}

// Get handles GET /api/v1/settings
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	s, err := h.settings.Get(r.Context(), athlete.ID)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load settings"))
		return
	}
	shared.WriteSuccess(w, s)
}

// Update handles PUT /api/v1/settings
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var s storage.AthleteSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	if err := h.settings.Upsert(r.Context(), athlete.ID, s); err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to save settings"))
		return
	}

	// Return normalized settings (defaults applied) rather than echoing input.
	saved, err := h.settings.Get(r.Context(), athlete.ID)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load settings"))
		return
	}
	shared.WriteSuccess(w, saved)
}
