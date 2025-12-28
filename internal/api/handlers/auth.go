package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/sasha/stata/internal/config"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	cfg      *config.Config
	strava   *strava.Client
	tokens   *storage.TokenRepository
	athletes *storage.AthleteRepository
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(
	cfg *config.Config,
	stravaClient *strava.Client,
	tokens *storage.TokenRepository,
	athletes *storage.AthleteRepository,
) *AuthHandler {
	return &AuthHandler{
		cfg:      cfg,
		strava:   stravaClient,
		tokens:   tokens,
		athletes: athletes,
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

	// Persist token and athlete to database
	if err := h.persistAuth(r.Context(), token, athlete); err != nil {
		slog.Error("failed to persist auth", "error", err)
		// Continue anyway - in-memory storage will still work for this session
	} else {
		slog.Info("auth persisted to database", "athlete_id", athlete.ID, "expires_at", token.Expiry)
	}

	// Store in memory for immediate use
	h.strava.SetToken(token, athlete)

	// Redirect to frontend - in dev mode redirect to Vite dev server
	redirectURL := "/settings"
	if h.cfg.Server.DevMode {
		redirectURL = "http://localhost:5173/settings"
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
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

	athlete := h.strava.GetAthlete()
	h.strava.SetToken(newToken, athlete)

	// Persist refreshed token
	if athlete != nil {
		if err := h.persistToken(r.Context(), newToken, athlete.ID); err != nil {
			slog.Error("failed to persist refreshed token", "error", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	resp := AuthStatusResponse{
		Authenticated: true,
		Athlete:       athlete,
		ExpiresAt:     newToken.Expiry.Unix(),
	}
	_ = json.NewEncoder(w).Encode(resp)
}

// persistAuth saves the athlete and token to the database.
func (h *AuthHandler) persistAuth(ctx context.Context, token *oauth2.Token, athlete *strava.Athlete) error {
	// Save athlete profile
	storageAthlete := &storage.Athlete{
		ID:            athlete.ID,
		Username:      athlete.Username,
		FirstName:     athlete.FirstName,
		LastName:      athlete.LastName,
		City:          athlete.City,
		State:         athlete.State,
		Country:       athlete.Country,
		Sex:           athlete.Sex,
		Premium:       athlete.Premium,
		Summit:        athlete.Summit,
		ProfileMedium: athlete.ProfileMedium,
		Profile:       athlete.Profile,
		Weight:        athlete.Weight,
	}
	if err := h.athletes.Upsert(ctx, storageAthlete); err != nil {
		return err
	}

	// Save token
	return h.persistToken(ctx, token, athlete.ID)
}

// persistToken saves the OAuth token to the database.
func (h *AuthHandler) persistToken(ctx context.Context, token *oauth2.Token, athleteID int64) error {
	storageToken := &storage.AuthToken{
		AthleteID:    athleteID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresAt:    token.Expiry,
	}
	return h.tokens.Upsert(ctx, storageToken)
}
