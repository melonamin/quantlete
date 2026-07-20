package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/melonamin/quantlete/internal/config"
)

func TestServeDataFile(t *testing.T) {
	dataDir := t.TempDir()
	challengeDir := filepath.Join(dataDir, "challenges")
	if err := os.MkdirAll(challengeDir, 0o750); err != nil {
		t.Fatalf("create challenge directory: %v", err)
	}
	const fileContents = "legitimate badge"
	if err := os.WriteFile(filepath.Join(challengeDir, "badge.svg"), []byte(fileContents), 0o600); err != nil {
		t.Fatalf("write badge fixture: %v", err)
	}

	router := &Router{cfg: &config.Config{Storage: config.StorageConfig{DataDir: dataDir}}}
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "traversal",
			path:       "/files/../../etc/passwd",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "legitimate file",
			path:       "/files/challenges/badge.svg",
			wantStatus: http.StatusOK,
			wantBody:   fileContents,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			router.serveDataFile(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("serveDataFile() status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rr.Body.String() != tt.wantBody {
				t.Errorf("serveDataFile() body = %q, want %q", rr.Body.String(), tt.wantBody)
			}
		})
	}
}
