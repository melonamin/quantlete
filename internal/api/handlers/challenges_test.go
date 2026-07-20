package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/strava"
)

type repeatingReader struct{}

func (repeatingReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

func TestChallengesHandlerImportRejectsOversizedMultipartBody(t *testing.T) {
	const boundary = "test-boundary"
	preamble := "--" + boundary + "\r\n" +
		"Content-Disposition: form-data; name=\"file\"; filename=\"challenges.html\"\r\n" +
		"Content-Type: text/html\r\n\r\n"
	body := io.MultiReader(
		strings.NewReader(preamble),
		io.LimitReader(repeatingReader{}, challengeImportBodyLimit+1),
	)

	stravaClient := strava.NewClient(&config.StravaConfig{})
	stravaClient.SetToken(nil, &strava.Athlete{ID: 1})
	handler := NewChallengesHandler(nil, stravaClient, t.TempDir())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/challenges/import", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary="+boundary)
	rr := httptest.NewRecorder()

	handler.Import(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Import() status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}
