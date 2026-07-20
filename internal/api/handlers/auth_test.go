package handlers

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/oauth2"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/strava"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestAuthHandlerHandleCallbackExchangeErrorRedirect(t *testing.T) {
	const rawError = "sensitive token exchange details"

	tests := []struct {
		name         string
		devMode      bool
		wantLocation string
	}{
		{name: "production", devMode: false, wantLocation: "/oauth/callback?error=token_exchange_failed"},
		{name: "dev mode redirects to Vite server", devMode: true, wantLocation: "http://localhost:5173/oauth/callback?error=token_exchange_failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Server: config.ServerConfig{DevMode: tt.devMode}}
			stravaClient := strava.NewClient(&config.StravaConfig{
				ClientID:     "client-id",
				ClientSecret: "client-secret",
			})
			handler := NewAuthHandler(cfg, stravaClient, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/strava/callback?state=test-state&code=bad-code", http.NoBody)
			req.AddCookie(&http.Cookie{Name: oauthStateCookieName, Value: "test-state"})
			httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusBadRequest,
					Status:     "400 Bad Request",
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"error":"` + rawError + `"}`)),
					Request:    r,
				}, nil
			})}
			req = req.WithContext(context.WithValue(req.Context(), oauth2.HTTPClient, httpClient))

			rr := httptest.NewRecorder()
			handler.HandleCallback(rr, req)

			if rr.Code != http.StatusTemporaryRedirect {
				t.Fatalf("HandleCallback() status = %d, want %d", rr.Code, http.StatusTemporaryRedirect)
			}
			if got := rr.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("HandleCallback() Location = %q, want %q", got, tt.wantLocation)
			}
			if strings.Contains(rr.Header().Get("Location"), rawError) || strings.Contains(rr.Body.String(), rawError) {
				t.Error("HandleCallback() redirect exposed the raw token exchange error")
			}
		})
	}
}

func TestIsRequestSecure(t *testing.T) {
	tests := []struct {
		name      string
		tls       *tls.ConnectionState
		forwarded string
		want      bool
	}{
		{name: "TLS", tls: &tls.ConnectionState{}, want: true},
		{name: "forwarded HTTPS", forwarded: "https", want: true},
		{name: "plain HTTP", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://quantlete.test", http.NoBody)
			req.TLS = tt.tls
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-Proto", tt.forwarded)
			}

			if got := isRequestSecure(req); got != tt.want {
				t.Errorf("isRequestSecure() = %t, want %t", got, tt.want)
			}
		})
	}
}
