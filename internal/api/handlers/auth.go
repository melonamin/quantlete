package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/sasha/stata/internal/config"
	"github.com/sasha/stata/internal/strava"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	cfg    *config.Config
	strava *strava.Client
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(cfg *config.Config, stravaClient *strava.Client) *AuthHandler {
	return &AuthHandler{
		cfg:    cfg,
		strava: stravaClient,
	}
}

// AuthStatusResponse represents the auth status response.
type AuthStatusResponse struct {
	Authenticated bool            `json:"authenticated"`
	Athlete       *strava.Athlete `json:"athlete,omitempty"`
	ExpiresAt     int64           `json:"expires_at,omitempty"`
}

// InitiateOAuth handles GET /api/v1/auth/strava.
// Redirects the user to Strava's OAuth authorization page.
func (h *AuthHandler) InitiateOAuth(w http.ResponseWriter, r *http.Request) {
	if h.cfg.Strava.ClientID == "" {
		http.Error(w, "Strava client ID not configured", http.StatusServiceUnavailable)
		return
	}

	authURL := h.strava.GetAuthURL()
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleCallback handles GET /api/v1/auth/strava/callback.
// Exchanges the authorization code for access tokens.
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		errMsg := r.URL.Query().Get("error")
		if errMsg == "" {
			errMsg = "no authorization code received"
		}
		slog.Error("OAuth callback error", "error", errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	token, athlete, err := h.strava.ExchangeCode(r.Context(), code)
	if err != nil {
		slog.Error("failed to exchange code", "error", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	slog.Info("OAuth successful", "athlete_id", athlete.ID, "athlete_name", athlete.FirstName+" "+athlete.LastName)

	// Store the token (in-memory for now, will be persisted to DB later)
	h.strava.SetToken(token, athlete)

	// Redirect to dashboard
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// Status handles GET /api/v1/auth/status.
// Returns the current authentication status.
func (h *AuthHandler) Status(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := h.strava.GetToken()
	athlete := h.strava.GetAthlete()

	resp := AuthStatusResponse{
		Authenticated: token != nil && token.Valid(),
	}

	if resp.Authenticated && athlete != nil {
		resp.Athlete = athlete
		resp.ExpiresAt = token.Expiry.Unix()
	}

	_ = json.NewEncoder(w).Encode(resp)
}

// RefreshToken handles POST /api/v1/auth/refresh.
// Forces a token refresh.
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	token := h.strava.GetToken()
	if token == nil {
		http.Error(w, "Not authenticated", http.StatusUnauthorized)
		return
	}

	newToken, err := h.strava.RefreshToken(r.Context(), token)
	if err != nil {
		slog.Error("failed to refresh token", "error", err)
		http.Error(w, "Failed to refresh token", http.StatusInternalServerError)
		return
	}

	h.strava.SetToken(newToken, h.strava.GetAthlete())

	w.Header().Set("Content-Type", "application/json")
	resp := AuthStatusResponse{
		Authenticated: true,
		Athlete:       h.strava.GetAthlete(),
		ExpiresAt:     newToken.Expiry.Unix(),
	}
	_ = json.NewEncoder(w).Encode(resp)
}
