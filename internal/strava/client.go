package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/oauth2"

	"github.com/sasha/stata/internal/config"
)

const (
	authURL  = "https://www.strava.com/oauth/authorize"
	tokenURL = "https://www.strava.com/oauth/token" //nolint:gosec // G101: This is a URL endpoint, not a credential
	apiBase  = "https://www.strava.com/api/v3"
)

// RateLimitPersister is called after each API request to persist rate limit state.
type RateLimitPersister func(json string) error

// Client is a Strava API client.
type Client struct {
	cfg              *config.StravaConfig
	oauth            *oauth2.Config
	httpClient       *http.Client
	rateLimit        *RateLimiter
	rateLimitPersist RateLimitPersister
	token            *oauth2.Token
	athlete          *Athlete
	tokenMu          sync.RWMutex
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

// GetAuthURL returns the URL to redirect users to for OAuth authorization.
func (c *Client) GetAuthURL() string {
	return c.oauth.AuthCodeURL("state", oauth2.AccessTypeOffline)
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

// IsAuthenticated returns true if the client has a valid token.
func (c *Client) IsAuthenticated() bool {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token != nil && c.token.Valid()
}

// do performs an authenticated API request.
func (c *Client) do(ctx context.Context, method, path string, result any) error {
	c.tokenMu.RLock()
	token := c.token
	c.tokenMu.RUnlock()

	if token == nil {
		return fmt.Errorf("not authenticated")
	}

	// Check rate limits
	if err := c.rateLimit.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiBase+path, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	// Update rate limits from response headers
	c.rateLimit.UpdateFromHeaders(resp.Header)

	// Persist rate limit state if persister is configured
	if c.rateLimitPersist != nil {
		if json, err := c.rateLimit.ToJSON(); err == nil {
			_ = c.rateLimitPersist(json) // Best effort, don't fail request on persistence error
		}
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: token may be expired")
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("rate limited by Strava")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error: %s", resp.Status)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	return nil
}

// getAthleteWithToken fetches the athlete profile using the provided token.
func (c *Client) getAthleteWithToken(ctx context.Context, token *oauth2.Token) (*Athlete, error) {
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

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: %s", resp.Status)
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
