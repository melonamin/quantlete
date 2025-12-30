package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// ImportHandler handles import-related endpoints.
type ImportHandler struct {
	importer     *importer.Importer
	syncHistory  *storage.SyncHistoryRepository
	stravaClient *strava.Client
}

// NewImportHandler creates a new import handler.
func NewImportHandler(imp *importer.Importer, syncHistory *storage.SyncHistoryRepository, stravaClient *strava.Client) *ImportHandler {
	return &ImportHandler{
		importer:     imp,
		syncHistory:  syncHistory,
		stravaClient: stravaClient,
	}
}

// StartImportRequest represents a request to start an import.
// By default, all data types are imported. Use skip_* fields to exclude specific types.
type StartImportRequest struct {
	FullSync        bool `json:"full_sync"`
	Resume          bool `json:"resume"`
	SkipStreams     bool `json:"skip_streams"`
	SkipSegments    bool `json:"skip_segments"`
	SkipBestEfforts bool `json:"skip_best_efforts"`
	SkipPhotos      bool `json:"skip_photos"`
}

// Start handles POST /api/v1/import/start
func (h *ImportHandler) Start(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var req StartImportRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
			return
		}
	}

	opts := importer.ImportOptions{
		FullSync:        req.FullSync,
		Resume:          req.Resume,
		SkipStreams:     req.SkipStreams,
		SkipSegments:    req.SkipSegments,
		SkipBestEfforts: req.SkipBestEfforts,
		SkipPhotos:      req.SkipPhotos,
	}

	// Use background context since import runs asynchronously after HTTP request completes
	if err := h.importer.Start(context.Background(), opts); err != nil {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "import already running or failed to start"})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "import started",
	})
}

// Progress handles GET /api/v1/import/progress
func (h *ImportHandler) Progress(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	progress := h.importer.Progress()
	writeJSON(w, http.StatusOK, progress)
}

// Cancel handles POST /api/v1/import/cancel
func (h *ImportHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	h.importer.Cancel()
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "import cancellation requested",
	})
}

// History handles GET /api/v1/import/history
func (h *ImportHandler) History(w http.ResponseWriter, r *http.Request) {
	if h.syncHistory == nil {
		writeJSON(w, http.StatusOK, []storage.SyncRun{})
		return
	}

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	runs, err := h.syncHistory.GetLatest(r.Context(), athlete.ID, 20)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	if runs == nil {
		runs = []storage.SyncRun{}
	}

	writeJSON(w, http.StatusOK, runs)
}

// Watermark handles GET /api/v1/import/watermark
func (h *ImportHandler) Watermark(w http.ResponseWriter, r *http.Request) {
	if h.syncHistory == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	wm, err := h.syncHistory.GetWatermark(r.Context(), athlete.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, wm)
}
