package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// urlEncode encodes a string for safe use in URL query parameters.
func urlEncode(s string) string {
	return url.QueryEscape(s)
}

// encodeJSON encodes data as JSON and logs any encoding errors.
func encodeJSON(w http.ResponseWriter, data any) {
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

const (
	oauthStateCookieName = "quantlete_oauth_state"
	oauthStateTTL        = 5 * time.Minute
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
	DemoMode      bool            `json:"demo_mode,omitempty"`
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

	state, err := generateStateToken()
	if err != nil {
		slog.Error("failed to generate oauth state", "error", err)
		http.Error(w, "Failed to initiate OAuth", http.StatusInternalServerError)
		return
	}

	h.setStateCookie(w, r, state)
	authURL := h.strava.GetAuthURL(state)
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// HandleCallback handles GET /api/v1/auth/strava/callback.
// Exchanges the authorization code for access tokens.
func (h *AuthHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		slog.Error("OAuth callback missing state parameter")
		http.Error(w, "missing state", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil || cookie.Value == "" || cookie.Value != state {
		slog.Error("OAuth state mismatch", "expected", cookieValueOrEmpty(cookie), "received", state, "error", err)
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}
	h.clearStateCookie(w, r)

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
		// Redirect to frontend with error message for better UX
		errorRedirect := "/settings?auth_error=" + urlEncode(err.Error())
		if h.cfg.Server.DevMode {
			errorRedirect = "http://localhost:5173" + errorRedirect
		}
		http.Redirect(w, r, errorRedirect, http.StatusTemporaryRedirect)
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
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token := h.strava.GetToken()
	athlete := h.strava.GetAthlete()

	// Demo mode: athlete exists but no token (set by demo command)
	isDemoMode := athlete != nil && token == nil

	// If token is present but expired, attempt an automatic refresh so the UI
	// stays authenticated without manual intervention.
	if token != nil && athlete != nil && !token.Valid() {
		if newToken, err := h.strava.RefreshToken(r.Context(), token); err == nil {
			h.strava.SetToken(newToken, athlete)
			_ = h.persistToken(context.Background(), newToken, athlete.ID)
			token = newToken
		}
	}

	resp := AuthStatusResponse{
		Authenticated: (token != nil && token.Valid()) || isDemoMode,
		DemoMode:      isDemoMode,
	}

	if athlete != nil {
		resp.Athlete = athlete
		if token != nil {
			resp.ExpiresAt = token.Expiry.Unix()
		}
	}

	encodeJSON(w, resp)
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
	encodeJSON(w, resp)
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
		ExpiresAt:    storage.SQLiteTime{Time: token.Expiry},
	}
	return h.tokens.Upsert(ctx, storageToken)
}

func generateStateToken() (string, error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func (h *AuthHandler) setStateCookie(w http.ResponseWriter, r *http.Request, state string) {
	secure := !h.cfg.Server.DevMode && r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(oauthStateTTL.Seconds()),
	})
}

func (h *AuthHandler) clearStateCookie(w http.ResponseWriter, r *http.Request) {
	secure := !h.cfg.Server.DevMode && r.TLS != nil
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func cookieValueOrEmpty(c *http.Cookie) string {
	if c == nil {
		return ""
	}
	return c.Value
}
