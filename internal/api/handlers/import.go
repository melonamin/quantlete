package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// sseChannelBuffer is the buffer size for SSE event channels.
// With eventBatchSize=25 and ~1000 items/second throughput, this provides
// ~2.5 seconds of buffer before events are dropped. Adjust if you see
// "SSE event dropped" log messages frequently.
const sseChannelBuffer = 100

// sseKeepaliveInterval is the interval for sending keepalive comments to SSE clients.
// Most reverse proxies (nginx, cloudflare) have 60s default timeouts.
// We use 30s to provide a safety margin for network latency.
const sseKeepaliveInterval = 30 * time.Second

// ImportHandler handles import-related endpoints.
type ImportHandler struct {
	importer       *importer.Importer
	syncHistory    *storage.SyncHistoryRepository
	stravaClient   *strava.Client
	allowedOrigins map[string]bool
}

// NewImportHandler creates a new import handler.
func NewImportHandler(imp *importer.Importer, syncHistory *storage.SyncHistoryRepository, stravaClient *strava.Client) *ImportHandler {
	return &ImportHandler{
		importer:       imp,
		syncHistory:    syncHistory,
		stravaClient:   stravaClient,
		allowedOrigins: make(map[string]bool),
	}
}

// SetAllowedOrigins configures the allowed origins for CORS on the SSE endpoint.
// This should be called after creation to set the allowed origins list.
func (h *ImportHandler) SetAllowedOrigins(origins []string) {
	h.allowedOrigins = make(map[string]bool, len(origins))
	for _, o := range origins {
		h.allowedOrigins[o] = true
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

// Events handles GET /api/v1/import/events (SSE endpoint for real-time updates)
func (h *ImportHandler) Events(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	// Create channel and subscribe early so we can defer unsubscribe before any panics
	events := make(chan importer.Event, sseChannelBuffer)
	unsubscribe := h.importer.Subscribe(events)

	// Ensure cleanup happens even if handler panics
	defer func() {
		if r := recover(); r != nil {
			slog.Error("SSE handler panic", "error", r, "athlete_id", athlete.ID)
		}
		unsubscribe()
	}()

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// CORS headers for SSE (allows cross-origin EventSource)
	// Only reflect origin if it's in the allowed origins list to prevent
	// arbitrary origins from accessing authenticated SSE streams.
	origin := r.Header.Get("Origin")
	if origin != "" && h.allowedOrigins[origin] {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	// Send initial progress
	progress := h.importer.Progress()
	if err := sendSSEEvent(w, flusher, string(importer.EventSyncProgress), progress); err != nil {
		slog.Debug("SSE: failed to send initial progress", "error", err)
		return
	}

	// Start keepalive ticker to prevent connection timeouts
	ticker := time.NewTicker(sseKeepaliveInterval)
	defer ticker.Stop()

	// Stream events until client disconnects
	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			slog.Debug("SSE: client disconnected")
			return
		case <-ticker.C:
			// Send SSE comment as keepalive (: prefix indicates comment)
			if _, err := w.Write([]byte(": keepalive\n\n")); err != nil {
				slog.Debug("SSE: keepalive failed, client disconnected", "error", err)
				return
			}
			flusher.Flush()
		case event, ok := <-events:
			if !ok {
				slog.Debug("SSE: event channel closed")
				return
			}
			if err := sendSSEEvent(w, flusher, string(event.Type), event.Data); err != nil {
				slog.Debug("SSE: failed to send event", "error", err)
				return
			}
		}
	}
}

// sendSSEEvent sends a single SSE event to the client.
func sendSSEEvent(w http.ResponseWriter, f http.Flusher, eventType string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("event: " + eventType + "\ndata: " + string(jsonData) + "\n\n"))
	if err != nil {
		return err
	}
	f.Flush()
	return nil
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

// Pause handles POST /api/v1/import/pause
func (h *ImportHandler) Pause(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	if h.importer.Pause() {
		writeJSON(w, http.StatusOK, map[string]string{
			"message": "import pause requested",
		})
	} else {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "no import running to pause"})
	}
}

// Resume handles POST /api/v1/import/resume
func (h *ImportHandler) Resume(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	// Check if there's a resumable state
	if !h.importer.CanResume(r.Context()) {
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "no resumable import state found"})
		return
	}

	// Start import with resume option
	opts := importer.ImportOptions{
		Resume: true,
	}

	// Use background context since import runs asynchronously after HTTP request completes
	if err := h.importer.Start(context.Background(), opts); err != nil {
		slog.Warn("failed to resume import", "error", err)
		writeJSON(w, http.StatusConflict, ErrorResponse{Error: "failed to resume import"})
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]string{
		"message": "import resumed",
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
