package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/shared"
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
	devMode        bool
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

// SetOriginValidation configures origin validation for CORS on the SSE endpoint.
// In dev mode, uses an explicit allowlist. In production mode, dynamically validates
// that Origin matches the request's Host header (for self-hosted flexibility).
func (h *ImportHandler) SetOriginValidation(devMode bool, devOrigins []string) {
	h.devMode = devMode
	h.allowedOrigins = make(map[string]bool, len(devOrigins))
	for _, o := range devOrigins {
		h.allowedOrigins[o] = true
	}
}

// isOriginAllowed checks if the Origin header matches the request's Host.
// This is safe for self-hosted apps where the user controls their proxy config.
func isOriginAllowed(origin string, r *http.Request) bool {
	// Extract host from origin (e.g., "http://localhost:8082" -> "localhost:8082")
	originHost := origin
	if idx := strings.Index(origin, "://"); idx != -1 {
		originHost = origin[idx+3:]
	}
	// Remove trailing slash if present
	originHost = strings.TrimSuffix(originHost, "/")

	// Get the Host header from the request
	requestHost := r.Host

	// Compare hosts (case-insensitive)
	return strings.EqualFold(originHost, requestHost)
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
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var req StartImportRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid request body"))
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
		shared.WriteJSONResponse(w, http.StatusConflict, shared.ErrorMessage("import already running or failed to start"))
		return
	}

	shared.WriteMessage(w, "import started")
}

// Progress handles GET /api/v1/import/progress
func (h *ImportHandler) Progress(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
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
	// In dev mode: check against explicit allowlist
	// In production: validate Origin matches request Host (for self-hosted flexibility)
	origin := r.Header.Get("Origin")
	if origin != "" {
		allowed := false
		if h.devMode {
			allowed = h.allowedOrigins[origin]
		} else {
			allowed = isOriginAllowed(origin, r)
		}
		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
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
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	h.importer.Cancel()
	shared.WriteMessage(w, "import cancellation requested")
}

// Pause handles POST /api/v1/import/pause
func (h *ImportHandler) Pause(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	if h.importer.Pause() {
		shared.WriteMessage(w, "import pause requested")
	} else {
		shared.WriteJSONResponse(w, http.StatusConflict, shared.ErrorMessage("no import running to pause"))
	}
}

// Resume handles POST /api/v1/import/resume
func (h *ImportHandler) Resume(w http.ResponseWriter, r *http.Request) {
	// Verify authentication
	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	// Check if there's a resumable state
	if !h.importer.CanResume(r.Context()) {
		shared.WriteJSONResponse(w, http.StatusConflict, shared.ErrorMessage("no resumable import state found"))
		return
	}

	// Start import with resume option
	opts := importer.ImportOptions{
		Resume: true,
	}

	// Use background context since import runs asynchronously after HTTP request completes
	if err := h.importer.Start(context.Background(), opts); err != nil {
		slog.Warn("failed to resume import", "error", err)
		shared.WriteJSONResponse(w, http.StatusConflict, shared.ErrorMessage("failed to resume import"))
		return
	}

	shared.WriteMessage(w, "import resumed")
}

// History handles GET /api/v1/import/history
func (h *ImportHandler) History(w http.ResponseWriter, r *http.Request) {
	if h.syncHistory == nil {
		shared.WriteSuccess(w, []storage.SyncRun{})
		return
	}

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	runs, err := h.syncHistory.GetLatest(r.Context(), athlete.ID, 20)
	if err != nil {
		slog.Error("failed to load sync history", "error", err, "athlete_id", athlete.ID)
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load sync history"))
		return
	}

	if runs == nil {
		runs = []storage.SyncRun{}
	}

	shared.WriteSuccess(w, runs)
}

// Watermark handles GET /api/v1/import/watermark
func (h *ImportHandler) Watermark(w http.ResponseWriter, r *http.Request) {
	if h.syncHistory == nil {
		shared.WriteSuccess(w, nil)
		return
	}

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	wm, err := h.syncHistory.GetWatermark(r.Context(), athlete.ID)
	if err != nil {
		slog.Error("failed to load watermark", "error", err, "athlete_id", athlete.ID)
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to load watermark"))
		return
	}

	shared.WriteSuccess(w, wm)
}

// ResetWatermark handles DELETE /api/v1/import/watermark
// This clears the sync watermark so the next import will be a full sync.
//
// Note: This endpoint is not rate-limited. In a multi-user environment,
// consider adding rate limiting to prevent abuse (repeated resets could
// trigger expensive full syncs and hit Strava API rate limits).
func (h *ImportHandler) ResetWatermark(w http.ResponseWriter, r *http.Request) {
	if h.syncHistory == nil {
		shared.WriteMessage(w, "watermark reset")
		return
	}

	athlete := h.stravaClient.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	if err := h.syncHistory.ClearWatermark(r.Context(), athlete.ID); err != nil {
		slog.Error("failed to reset watermark", "error", err, "athlete_id", athlete.ID)
		shared.WriteJSONResponse(w, http.StatusInternalServerError, shared.ErrorMessage("failed to reset watermark"))
		return
	}

	slog.Info("sync watermark reset", "athlete_id", athlete.ID)
	shared.WriteMessage(w, "watermark reset - next sync will be full")
}
