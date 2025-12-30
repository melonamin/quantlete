package handlers

import (
	"encoding/json"
	"net/http"

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
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	cfg, err := h.goals.GetConfig(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to get goals config"})
		return
	}

	progress := make(map[string]map[storage.GoalPeriod]storage.GoalsProgress)
	for _, sport := range cfg.Sports {
		per := make(map[storage.GoalPeriod]storage.GoalsProgress)
		for _, p := range []storage.GoalPeriod{storage.GoalPeriodWeek, storage.GoalPeriodMonth, storage.GoalPeriodYear, storage.GoalPeriodLifetime} {
			val, err := h.goals.GetProgress(r.Context(), athlete.ID, sport.SportTypes, p)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to compute goals progress"})
				return
			}
			per[p] = val
		}
		progress[sport.Name] = per
	}

	writeJSON(w, http.StatusOK, storage.TrainingGoalsResponse{
		Config:   *cfg,
		Progress: progress,
	})
}

// UpdateGoals handles PUT /api/v1/goals
func (h *GoalsHandler) UpdateGoals(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var cfg storage.TrainingGoalsConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}
	if len(cfg.Sports) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "at least one sport group is required"})
		return
	}
	for _, s := range cfg.Sports {
		if s.Name == "" {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "sport group name is required"})
			return
		}
		if len(s.SportTypes) == 0 {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "sport_types is required"})
			return
		}
	}

	if err := h.goals.UpsertConfig(r.Context(), athlete.ID, cfg); err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to save goals config"})
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}
