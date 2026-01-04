package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type GoalsHandler struct {
	goals  *storage.GoalsRepository
	strava *strava.Client
}

func NewGoalsHandler(goals *storage.GoalsRepository, stravaClient *strava.Client) *GoalsHandler {
	return &GoalsHandler{
		goals:  goals,
		strava: stravaClient,
	}
}

// GetGoals handles GET /api/v1/goals
func (h *GoalsHandler) GetGoals(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	cfg, err := h.goals.GetConfig(r.Context(), athlete.ID)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to get goals config"))
		return
	}

	progress := make(map[string]map[storage.GoalPeriod]storage.GoalsProgress)
	for _, sport := range cfg.Sports {
		per := make(map[storage.GoalPeriod]storage.GoalsProgress)
		for _, p := range []storage.GoalPeriod{storage.GoalPeriodWeek, storage.GoalPeriodMonth, storage.GoalPeriodYear, storage.GoalPeriodLifetime} {
			val, err := h.goals.GetProgress(r.Context(), athlete.ID, sport.SportTypes, p)
			if err != nil {
				shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to compute goals progress"))
				return
			}
			per[p] = val
		}
		progress[sport.Name] = per
	}

	shared.WriteSuccess(w, storage.TrainingGoalsResponse{
		Config:   *cfg,
		Progress: progress,
	})
}

// UpdateGoals handles PUT /api/v1/goals
func (h *GoalsHandler) UpdateGoals(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var cfg storage.TrainingGoalsConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if len(cfg.Sports) == 0 {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("at least one sport group is required"))
		return
	}
	for _, s := range cfg.Sports {
		if s.Name == "" {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("sport group name is required"))
			return
		}
		if len(s.SportTypes) == 0 {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("sport_types is required"))
			return
		}
	}

	if err := h.goals.UpsertConfig(r.Context(), athlete.ID, cfg); err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to save goals config"))
		return
	}

	shared.WriteSuccess(w, cfg)
}
