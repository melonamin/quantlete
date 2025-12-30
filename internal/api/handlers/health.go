package handlers

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/melonamin/quantlete"
)

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status string `json:"status"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthCheck handles GET /api/v1/health.
func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := HealthResponse{Status: "ok"}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode health response", "error", err)
	}
}

// NotImplemented returns a 501 Not Implemented response.
func NotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":"not implemented"}`))
}

// ServeFrontend serves the embedded React frontend.
func ServeFrontend(w http.ResponseWriter, r *http.Request) {
	if err := ensureFrontendAssets(); err != nil {
		slog.Error("unable to initialize embedded frontend", "error", err)
		http.Error(w, "frontend not available", http.StatusInternalServerError)
		return
	}

	requestPath := strings.TrimPrefix(path.Clean(urlPathOrRoot(r)), "/")
	if requestPath == "" {
		serveIndexHTML(w, r)
		return
	}

	if serveEmbeddedFile(w, r, requestPath) {
		return
	}

	serveIndexHTML(w, r)
}

var (
	frontendOnce     sync.Once
	frontendInitErr  error
	frontendFS       fs.FS
	frontendIndex    []byte
	frontendIndexMod time.Time
)

func ensureFrontendAssets() error {
	frontendOnce.Do(func() {
		sub, err := fs.Sub(quantlete.WebAssets, "web/dist")
		if err != nil {
			frontendInitErr = err
			return
		}
		frontendFS = sub

		indexBytes, err := fs.ReadFile(frontendFS, "index.html")
		if err != nil {
			frontendInitErr = err
			return
		}
		frontendIndex = indexBytes
		frontendIndexMod = time.Now()
	})
	return frontendInitErr
}

func serveEmbeddedFile(w http.ResponseWriter, r *http.Request, filePath string) bool {
	if frontendFS == nil {
		return false
	}

	data, err := fs.ReadFile(frontendFS, filePath)
	if err != nil {
		return false
	}

	modTime := time.Now()
	if info, statErr := fs.Stat(frontendFS, filePath); statErr == nil && !info.IsDir() {
		modTime = info.ModTime()
	}

	http.ServeContent(w, r, filePath, modTime, bytes.NewReader(data))
	return true
}

func serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	if len(frontendIndex) == 0 {
		http.Error(w, "frontend not available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	http.ServeContent(w, r, "index.html", frontendIndexMod, bytes.NewReader(frontendIndex))
}

func urlPathOrRoot(r *http.Request) string {
	if r == nil || r.URL == nil {
		return "/"
	}
	return r.URL.Path
}
