package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sasha/stata/internal/importer"
)

// ImportHandler handles import-related endpoints.
type ImportHandler struct {
	importer *importer.Importer
}

// NewImportHandler creates a new import handler.
func NewImportHandler(imp *importer.Importer) *ImportHandler {
	return &ImportHandler{importer: imp}
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
	var req StartImportRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err.Error() != "EOF" {
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
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "import started",
	})
}

// Progress handles GET /api/v1/import/progress
func (h *ImportHandler) Progress(w http.ResponseWriter, r *http.Request) {
	progress := h.importer.Progress()
	writeJSON(w, http.StatusOK, progress)
}

// Cancel handles POST /api/v1/import/cancel
func (h *ImportHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	h.importer.Cancel()
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "import cancellation requested",
	})
}
