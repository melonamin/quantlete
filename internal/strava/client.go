package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/melonamin/quantlete/internal/config"
)

const (
	authURL  = "https://www.strava.com/oauth/authorize"
	tokenURL = "https://www.strava.com/oauth/token" //nolint:gosec // G101: This is a URL endpoint, not a credential
	apiBase  = "https://www.strava.com/api/v3"
)

// APIError represents an error response from the Strava API with status code details.
type APIError struct {
	StatusCode int
	Status     string
	Path       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Strava API error %d (%s) for %s", e.StatusCode, e.Status, e.Path)
}

// IsNotFound returns true if this is a 404 error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsForbidden returns true if this is a 403 error.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsServerError returns true if this is a 5xx error.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// RateLimitPersister is called after each API request to persist rate limit state.
type RateLimitPersister func(json string) error

// TokenPersister is called after an automatic token refresh to persist the new token.
type TokenPersister func(token *oauth2.Token, athleteID int64) error

// Client is a Strava API client.
type Client struct {
	cfg              *config.StravaConfig
	oauth            *oauth2.Config
	httpClient       *http.Client
	rateLimit        *RateLimiter
	rateLimitPersist RateLimitPersister
	tokenPersist     TokenPersister
	token            *oauth2.Token
	athlete          *Athlete
	tokenMu          sync.RWMutex
	refreshMu        sync.Mutex // serializes token refresh operations
}

// NewClient creates a new Strava API client.
func NewClient(cfg *config.StravaConfig) *Client {
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       []string{"activity:read_all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	return &Client{
		cfg:        cfg,
		oauth:      oauthCfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		rateLimit:  NewRateLimiter(),
	}
}

// UpdateCredentials updates the OAuth configuration with new credentials at runtime.
// This is used when credentials are configured via the API instead of environment variables.
func (c *Client) UpdateCredentials(clientID, clientSecret, redirectURI string) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	c.cfg.ClientID = clientID
	c.cfg.ClientSecret = clientSecret
	if redirectURI != "" {
		c.cfg.RedirectURI = redirectURI
	}

	c.oauth = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  c.cfg.RedirectURI,
		Scopes:       []string{"activity:read_all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}
}

// GetAuthURL returns the URL to redirect users to for OAuth authorization.
func (c *Client) GetAuthURL(state string) string {
	if state == "" {
		state = "state"
	}
	return c.oauth.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges an authorization code for access tokens.
func (c *Client) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, *Athlete, error) {
	token, err := c.oauth.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("exchanging code: %w", err)
	}

	// The athlete info is included in the token response from Strava
	// We need to make an additional request to get it
	athlete, err := c.getAthleteWithToken(ctx, token)
	if err != nil {
		return nil, nil, fmt.Errorf("getting athlete: %w", err)
	}

	return token, athlete, nil
}

// RefreshToken refreshes an expired token.
func (c *Client) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	src := c.oauth.TokenSource(ctx, token)
	newToken, err := src.Token()
	if err != nil {
		return nil, fmt.Errorf("refreshing token: %w", err)
	}
	return newToken, nil
}

// SetToken stores the current token and athlete.
func (c *Client) SetToken(token *oauth2.Token, athlete *Athlete) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.token = token
	c.athlete = athlete
}

// GetToken returns the current token.
func (c *Client) GetToken() *oauth2.Token {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token
}

// GetAthlete returns the current athlete.
func (c *Client) GetAthlete() *Athlete {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.athlete
}

// RateLimiter returns the rate limiter for state persistence.
func (c *Client) RateLimiter() *RateLimiter {
	return c.rateLimit
}

// SetRateLimitPersister sets the callback for persisting rate limit state.
func (c *Client) SetRateLimitPersister(p RateLimitPersister) {
	c.rateLimitPersist = p
}

// SetTokenPersister sets the callback for persisting refreshed tokens.
func (c *Client) SetTokenPersister(p TokenPersister) {
	c.tokenPersist = p
}

// IsAuthenticated returns true if the client has a valid token.
func (c *Client) IsAuthenticated() bool {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token != nil && c.token.Valid()
}

// persistToken persists the token using the configured persister.
// Returns an error if persistence fails, allowing the caller to decide how to handle it.
// Token persistence is critical: if it fails, the next restart will lose the token.
func (c *Client) persistToken(token *oauth2.Token, athleteID int64) error {
	if c.tokenPersist == nil {
		return nil
	}
	if err := c.tokenPersist(token, athleteID); err != nil {
		return fmt.Errorf("token persistence failed for athlete %d: %w", athleteID, err)
	}
	return nil
}

// tokenRefreshTimeout is the maximum time allowed for a token refresh operation.
const tokenRefreshTimeout = 30 * time.Second

// forceRefreshToken performs a mutex-protected token refresh, used when a 401
// is received after the initial token validation passed (e.g., token expired
// server-side between validation and request).
func (c *Client) forceRefreshToken(ctx context.Context) (*oauth2.Token, error) {
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	// Add timeout for the refresh operation to prevent indefinite blocking.
	ctx, cancel := context.WithTimeout(ctx, tokenRefreshTimeout)
	defer cancel()

	c.tokenMu.RLock()
	currentToken := c.token
	athlete := c.athlete
	c.tokenMu.RUnlock()

	if athlete == nil {
		return nil, fmt.Errorf("not authenticated")
	}

	newToken, err := c.RefreshToken(ctx, currentToken)
	if err != nil {
		return nil, err
	}

	c.SetToken(newToken, athlete)
	if err := c.persistToken(newToken, athlete.ID); err != nil {
		// Log error prominently - token refresh succeeded but if app restarts,
		// the old token on disk will be loaded, requiring re-authentication.
		// We intentionally don't return an error here because:
		// 1. The token refresh succeeded - the current session works fine
		// 2. Returning an error would fail the API request even though it can proceed
		// 3. This is a "best-effort" persistence; the error is logged for operators
		slog.Error("CRITICAL: token persistence failed after refresh - re-authentication may be required after restart",
			"athlete_id", athlete.ID, "error", err)
	}
	return newToken, nil
}

func (c *Client) ensureValidToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	if token == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	if token.Valid() {
		return token, nil
	}

	// Serialize refresh operations to prevent concurrent refresh attempts
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()

	// Add timeout for the refresh operation to prevent indefinite blocking.
	ctx, cancel := context.WithTimeout(ctx, tokenRefreshTimeout)
	defer cancel()

	// Re-check after acquiring lock — another goroutine may have refreshed already
	c.tokenMu.RLock()
	currentToken := c.token
	athlete := c.athlete
	c.tokenMu.RUnlock()

	if athlete == nil {
		return nil, fmt.Errorf("not authenticated")
	}
	if currentToken != nil && currentToken.Valid() {
		return currentToken, nil
	}

	newToken, err := c.RefreshToken(ctx, currentToken)
	if err != nil {
		return nil, err
	}

	c.SetToken(newToken, athlete)
	if err := c.persistToken(newToken, athlete.ID); err != nil {
		// Log error prominently - token refresh succeeded but if app restarts,
		// the old token on disk will be loaded, requiring re-authentication.
		// We intentionally don't return an error here because:
		// 1. The token refresh succeeded - the current session works fine
		// 2. Returning an error would fail the API request even though it can proceed
		// 3. This is a "best-effort" persistence; the error is logged for operators
		slog.Error("CRITICAL: token persistence failed after refresh - re-authentication may be required after restart",
			"athlete_id", athlete.ID, "error", err)
	}
	return newToken, nil
}

// do performs an authenticated API request.
func (c *Client) do(ctx context.Context, method, path string, result any) error {
	c.tokenMu.RLock()
	token := c.token
	athlete := c.athlete
	c.tokenMu.RUnlock()

	var err error
	token, err = c.ensureValidToken(ctx, token)
	if err != nil {
		return err
	}

	// Check rate limits
	if err = c.rateLimit.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit: %w", err)
	}

	doRequest := func(accessToken string) (*http.Response, error) {
		req, reqErr := http.NewRequestWithContext(ctx, method, apiBase+path, http.NoBody)
		if reqErr != nil {
			return nil, fmt.Errorf("creating request: %w", reqErr)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		return c.httpClient.Do(req)
	}

	resp, err := doRequest(token.AccessToken)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// If we got a 401, try a single refresh+retry using mutex-protected refresh.
	if resp.StatusCode == http.StatusUnauthorized && athlete != nil {
		_ = resp.Body.Close()
		newToken, refreshErr := c.forceRefreshToken(ctx)
		if refreshErr != nil {
			return fmt.Errorf("unauthorized: token refresh failed: %w", refreshErr)
		}
		resp, err = doRequest(newToken.AccessToken)
		if err != nil {
			return fmt.Errorf("executing request: %w", err)
		}
		defer resp.Body.Close()
	}

	// Update rate limits from response headers
	c.rateLimit.UpdateFromHeaders(resp.Header)

	// Persist rate limit state if persister is configured
	if c.rateLimitPersist != nil {
		if jsonStr, err := c.rateLimit.ToJSON(); err == nil {
			_ = c.rateLimitPersist(jsonStr) // Best effort, don't fail request on persistence error
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return c.rateLimit.CreateRateLimitError()
	}

	if resp.StatusCode >= 400 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Path:       path,
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// getAthleteWithToken fetches the athlete profile using the provided token.
// This is called during OAuth callback, so it fails fast on rate limits rather than waiting.
func (c *Client) getAthleteWithToken(ctx context.Context, token *oauth2.Token) (*Athlete, error) {
	// Check if we're already at rate limit - fail fast for OAuth flow
	if !c.rateLimit.CanMakeRequest() {
		waitTime := c.rateLimit.TimeUntil15MinReset()
		return nil, fmt.Errorf("rate limited, please try again in %v", waitTime.Round(time.Minute))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+"/athlete", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limits from response headers
	c.rateLimit.UpdateFromHeaders(resp.Header)

	// Persist rate limit state if persister is configured
	if c.rateLimitPersist != nil {
		if jsonStr, err := c.rateLimit.ToJSON(); err == nil {
			_ = c.rateLimitPersist(jsonStr)
		}
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		waitTime := c.rateLimit.TimeUntil15MinReset()
		return nil, fmt.Errorf("rate limited by Strava, please try again in %v", waitTime.Round(time.Minute))
	}

	if resp.StatusCode >= 400 {
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Path:       "/athlete",
		}
	}

	var athlete Athlete
	if err := json.NewDecoder(resp.Body).Decode(&athlete); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &athlete, nil
}

// GetAthlete fetches the authenticated athlete's profile.
func (c *Client) FetchAthlete(ctx context.Context) (*Athlete, error) {
	var athlete Athlete
	if err := c.do(ctx, http.MethodGet, "/athlete", &athlete); err != nil {
		return nil, err
	}
	return &athlete, nil
}
