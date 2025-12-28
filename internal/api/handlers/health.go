package handlers

import (
	"encoding/json"
	"net/http"
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
	_ = json.NewEncoder(w).Encode(resp)
}

// NotImplemented returns a 501 Not Implemented response.
func NotImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_, _ = w.Write([]byte(`{"error":"not implemented"}`))
}

// ServeFrontend serves the React frontend placeholder.
func ServeFrontend(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Stata</title></head>
<body>
<h1>Stata - Statistics for Strava</h1>
<p>Web UI will be served here. Run React dev server separately during development.</p>
<p><a href="/api/v1/health">API Health Check</a></p>
<p><a href="/api/v1/auth/strava">Connect with Strava</a></p>
</body>
</html>`))
}
