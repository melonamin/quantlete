package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

type SegmentsHandler struct {
	svc    *services.SegmentsService
	strava *strava.Client
}

func NewSegmentsHandler(svc *services.SegmentsService, stravaClient *strava.Client) *SegmentsHandler {
	return &SegmentsHandler{svc: svc, strava: stravaClient}
}

// List handles GET /api/v1/segments
func (h *SegmentsHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	q := r.URL.Query()
	qp := pagination.ParseQueryParams(q)
	in := services.ListSegmentsInput{
		AthleteID:    athlete.ID,
		ActivityType: q.Get("activity_type"),
		Country:      q.Get("country"),
		Search:       q.Get("search"),
		KOMOnly:      q.Get("kom_only") == "true" || q.Get("kom_only") == "1",
		Page:         qp.Page,
		PerPage:      qp.PerPage,
		OrderBy:      qp.OrderBy,
		OrderDir:     qp.OrderDir,
	}
	if v := q.Get("starred"); v != "" {
		b := v == "true" || v == "1"
		in.Starred = &b
	}

	result, err := h.svc.List(r.Context(), in)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// GetByID handles GET /api/v1/segments/:id
func (h *SegmentsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid segment id"))
		return
	}

	result, err := h.svc.GetByID(r.Context(), services.GetSegmentInput{
		AthleteID: athlete.ID,
		SegmentID: id,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// ListEfforts handles GET /api/v1/segments/:id/efforts
func (h *SegmentsHandler) ListEfforts(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid segment id"))
		return
	}

	q := r.URL.Query()
	qp := pagination.ParseQueryParams(q)
	result, err := h.svc.ListEfforts(r.Context(), services.ListEffortsInput{
		AthleteID: athlete.ID,
		SegmentID: id,
		Page:      qp.Page,
		PerPage:   qp.PerPage,
		OrderBy:   qp.OrderBy,
		OrderDir:  qp.OrderDir,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

// Countries handles GET /api/v1/segments/countries
func (h *SegmentsHandler) Countries(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	result, err := h.svc.ListCountries(r.Context(), athlete.ID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}
