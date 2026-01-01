package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

const (
	appStateKeyClientID     = "strava_client_id"
	appStateKeyClientSecret = "strava_client_secret" //nolint:gosec // G101: This is a key name, not a credential

	// Strava client ID is a numeric string, typically 5-6 digits.
	clientIDMinLen = 4
	clientIDMaxLen = 20
	// Strava client secret is a 40-character hex string.
	clientSecretLen = 40
)

var (
	clientIDPattern     = regexp.MustCompile(`^\d+$`)
	clientSecretPattern = regexp.MustCompile(`^[a-f0-9]+$`)
)

// SetupHandler handles setup and configuration endpoints.
type SetupHandler struct {
	cfg      *config.Config
	appState *storage.AppStateRepository
	strava   *strava.Client
}

// NewSetupHandler creates a new setup handler.
func NewSetupHandler(
	cfg *config.Config,
	appState *storage.AppStateRepository,
	stravaClient *strava.Client,
) *SetupHandler {
	return &SetupHandler{
		cfg:      cfg,
		appState: appState,
		strava:   stravaClient,
	}
}

// CredentialsStatusResponse represents the credentials status response.
type CredentialsStatusResponse struct {
	Configured  bool   `json:"configured"`
	ClientID    string `json:"client_id,omitempty"`
	Source      string `json:"source,omitempty"` // "env" or "database"
	RedirectURI string `json:"redirect_uri"`
}

// UpdateCredentialsRequest represents the request to update credentials.
type UpdateCredentialsRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// GetCredentialsStatus returns the current credentials configuration status.
// GET /api/v1/setup/credentials
func (h *SetupHandler) GetCredentialsStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	response := CredentialsStatusResponse{
		RedirectURI: h.cfg.Strava.RedirectURI,
	}

	// Check if credentials are configured via environment variables
	if h.cfg.Strava.ClientID != "" && h.cfg.Strava.ClientSecret != "" {
		response.Configured = true
		response.ClientID = maskClientID(h.cfg.Strava.ClientID)
		response.Source = "env"
	} else {
		// Check database for credentials
		clientID, err := h.appState.Get(ctx, appStateKeyClientID)
		if err != nil {
			slog.Error("failed to get client ID from database", "error", err)
			http.Error(w, "Failed to check credentials status", http.StatusInternalServerError)
			return
		}

		clientSecret, err := h.appState.Get(ctx, appStateKeyClientSecret)
		if err != nil {
			slog.Error("failed to get client secret from database", "error", err)
			http.Error(w, "Failed to check credentials status", http.StatusInternalServerError)
			return
		}

		if clientID != "" && clientSecret != "" {
			response.Configured = true
			response.ClientID = maskClientID(clientID)
			response.Source = "database"
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// UpdateCredentials stores new Strava credentials in the database.
// PUT /api/v1/setup/credentials
func (h *SetupHandler) UpdateCredentials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req UpdateCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := validateCredentials(req.ClientID, req.ClientSecret); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Store in database
	if err := h.appState.Set(ctx, appStateKeyClientID, req.ClientID); err != nil {
		slog.Error("failed to store client ID", "error", err)
		http.Error(w, "Failed to save credentials", http.StatusInternalServerError)
		return
	}

	if err := h.appState.Set(ctx, appStateKeyClientSecret, req.ClientSecret); err != nil {
		slog.Error("failed to store client secret", "error", err)
		http.Error(w, "Failed to save credentials", http.StatusInternalServerError)
		return
	}

	// Update the Strava client with new credentials
	h.strava.UpdateCredentials(req.ClientID, req.ClientSecret, h.cfg.Strava.RedirectURI)

	slog.Info("strava credentials updated via API")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]bool{"success": true}); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// validateCredentials validates Strava client credentials format.
func validateCredentials(clientID, clientSecret string) error {
	clientID = strings.TrimSpace(clientID)
	clientSecret = strings.TrimSpace(clientSecret)

	if clientID == "" {
		return &validationError{"client_id is required"}
	}
	if len(clientID) < clientIDMinLen || len(clientID) > clientIDMaxLen {
		return &validationError{"client_id must be between 4 and 20 characters"}
	}
	if !clientIDPattern.MatchString(clientID) {
		return &validationError{"client_id must be numeric"}
	}

	if clientSecret == "" {
		return &validationError{"client_secret is required"}
	}
	if len(clientSecret) != clientSecretLen {
		return &validationError{"client_secret must be exactly 40 characters"}
	}
	if !clientSecretPattern.MatchString(clientSecret) {
		return &validationError{"client_secret must be a hexadecimal string"}
	}

	return nil
}

// validationError is a simple error type for validation failures.
type validationError struct {
	message string
}

func (e *validationError) Error() string {
	return e.message
}

// maskClientID returns a masked version of the client ID for display.
// Shows only the first 2 characters to minimize information exposure.
func maskClientID(clientID string) string {
	if len(clientID) <= 2 {
		return "****"
	}
	return clientID[:2] + "******"
}

// LoadCredentialsFromDB loads credentials from database into the config and Strava client.
// This should be called on startup if env vars are not set.
//
// NOTE: This function mutates cfg.Strava fields. This is intentional so that
// GetCredentialsStatus can detect the credential source correctly. The config
// mutation happens only once at startup, before any handlers are served.
func LoadCredentialsFromDB(
	ctx context.Context,
	appState *storage.AppStateRepository,
	cfg *config.Config,
	stravaClient *strava.Client,
) error {
	// If env vars are already set, don't override
	if cfg.Strava.ClientID != "" && cfg.Strava.ClientSecret != "" {
		slog.Debug("using strava credentials from environment variables")
		return nil
	}

	// Attempt to load from database
	clientID, err := appState.Get(ctx, appStateKeyClientID)
	if err != nil {
		return err
	}

	clientSecret, err := appState.Get(ctx, appStateKeyClientSecret)
	if err != nil {
		return err
	}

	if clientID == "" || clientSecret == "" {
		slog.Debug("no strava credentials found in database")
		return nil
	}

	// Update the Strava client with database credentials.
	// We intentionally do NOT set cfg.Strava fields here so that
	// GetCredentialsStatus correctly identifies the source as "database".
	stravaClient.UpdateCredentials(clientID, clientSecret, cfg.Strava.RedirectURI)

	slog.Info("loaded strava credentials from database")
	return nil
}
