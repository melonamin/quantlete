package config

import (
	"testing"
)

func TestLoadStravaCredentialsFromEnv(t *testing.T) {
	// Set env vars (t.Setenv automatically cleans up after test)
	t.Setenv("QUANTLETE_STRAVA_CLIENT_ID", "test-client-id-12345")
	t.Setenv("QUANTLETE_STRAVA_CLIENT_SECRET", "test-client-secret-abcdef")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Strava.ClientID != "test-client-id-12345" {
		t.Errorf("ClientID = %q, want %q", cfg.Strava.ClientID, "test-client-id-12345")
	}

	if cfg.Strava.ClientSecret != "test-client-secret-abcdef" {
		t.Errorf("ClientSecret = %q, want %q", cfg.Strava.ClientSecret, "test-client-secret-abcdef")
	}
}
