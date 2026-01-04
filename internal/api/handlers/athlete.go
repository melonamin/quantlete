package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type AthleteHandler struct {
	metrics *storage.AthleteMetricsRepository
	strava  *strava.Client
}

func NewAthleteHandler(metrics *storage.AthleteMetricsRepository, stravaClient *strava.Client) *AthleteHandler {
	return &AthleteHandler{
		metrics: metrics,
		strava:  stravaClient,
	}
}

type MetricPointDTO struct {
	RecordedAt string  `json:"recorded_at"` // YYYY-MM-DD
	Value      float64 `json:"value"`
}

func parseMetricPoints(points []MetricPointDTO) ([]storage.AthleteMetricPoint, error) {
	out := make([]storage.AthleteMetricPoint, 0, len(points))
	for _, p := range points {
		t, err := time.Parse("2006-01-02", p.RecordedAt)
		if err != nil {
			// allow RFC3339 too
			t, err = time.Parse(time.RFC3339, p.RecordedAt)
		}
		if err != nil {
			return nil, err
		}
		out = append(out, storage.AthleteMetricPoint{RecordedAt: storage.SQLiteTime{Time: t}, Value: p.Value})
	}
	return out, nil
}

func formatMetricPoints(points []storage.AthleteMetricPoint) []MetricPointDTO {
	out := make([]MetricPointDTO, 0, len(points))
	for _, p := range points {
		out = append(out, MetricPointDTO{
			RecordedAt: p.RecordedAt.Format("2006-01-02"),
			Value:      p.Value,
		})
	}
	return out
}

type FTPHistoryResponse struct {
	Cycling []MetricPointDTO `json:"cycling"`
	Running []MetricPointDTO `json:"running"`
}

// GetFTP handles GET /api/v1/athlete/ftp
func (h *AthleteHandler) GetFTP(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	c, err := h.metrics.List(r.Context(), athlete.ID, "ftp_cycling_watts")
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load ftp history"))
		return
	}
	run, err := h.metrics.List(r.Context(), athlete.ID, "ftp_running_mps")
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load ftp history"))
		return
	}

	shared.WriteSuccess(w, FTPHistoryResponse{
		Cycling: formatMetricPoints(c),
		Running: formatMetricPoints(run),
	})
}

type UpdateFTPRequest struct {
	Cycling []MetricPointDTO `json:"cycling"`
	Running []MetricPointDTO `json:"running"`
}

// UpdateFTP handles PUT /api/v1/athlete/ftp
func (h *AthleteHandler) UpdateFTP(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var req UpdateFTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	c, err := parseMetricPoints(req.Cycling)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid cycling points"))
		return
	}
	run, err := parseMetricPoints(req.Running)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid running points"))
		return
	}

	if err := h.metrics.Replace(r.Context(), athlete.ID, "ftp_cycling_watts", c); err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to save cycling ftp"))
		return
	}
	if err := h.metrics.Replace(r.Context(), athlete.ID, "ftp_running_mps", run); err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to save running ftp"))
		return
	}

	shared.WriteSuccess(w, FTPHistoryResponse{
		Cycling: formatMetricPoints(c),
		Running: formatMetricPoints(run),
	})
}

type WeightHistoryResponse struct {
	Points []MetricPointDTO `json:"points"`
}

// GetWeight handles GET /api/v1/athlete/weight
func (h *AthleteHandler) GetWeight(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	points, err := h.metrics.List(r.Context(), athlete.ID, "weight_kg")
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load weight history"))
		return
	}

	shared.WriteSuccess(w, WeightHistoryResponse{Points: formatMetricPoints(points)})
}

type UpdateWeightRequest struct {
	Points []MetricPointDTO `json:"points"`
}

// UpdateWeight handles PUT /api/v1/athlete/weight
func (h *AthleteHandler) UpdateWeight(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var req UpdateWeightRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
		return
	}

	points, err := parseMetricPoints(req.Points)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid points"))
		return
	}

	if err := h.metrics.Replace(r.Context(), athlete.ID, "weight_kg", points); err != nil {
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to save weight"))
		return
	}

	shared.WriteSuccess(w, WeightHistoryResponse{Points: formatMetricPoints(points)})
}
