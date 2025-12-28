package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sasha/stata/internal/challenges"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

type ChallengesHandler struct {
	repo   *storage.ChallengeRepository
	strava *strava.Client
}

func NewChallengesHandler(repo *storage.ChallengeRepository, stravaClient *strava.Client) *ChallengesHandler {
	return &ChallengesHandler{repo: repo, strava: stravaClient}
}

type challengeResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug,omitempty"`
	BadgeURL       string  `json:"badge_url,omitempty"`
	CompletionDate *string `json:"completion_date,omitempty"` // YYYY-MM-DD
	Month          string  `json:"month,omitempty"`
}

func challengeToResponse(c storage.Challenge) challengeResponse {
	var completion *string
	if c.CompletionDate != nil && !c.CompletionDate.IsZero() {
		v := c.CompletionDate.Format("2006-01-02")
		completion = &v
	}
	return challengeResponse{
		ID:             c.ID,
		Name:           c.Name,
		Slug:           c.Slug,
		BadgeURL:       c.BadgeURL,
		CompletionDate: completion,
		Month:          c.Month,
	}
}

// List handles GET /api/v1/challenges
func (h *ChallengesHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	month := strings.TrimSpace(r.URL.Query().Get("month"))
	items, err := h.repo.List(r.Context(), athlete.ID, month)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch challenges"})
		return
	}
	out := make([]challengeResponse, 0, len(items))
	for _, c := range items {
		out = append(out, challengeToResponse(c))
	}
	writeJSON(w, http.StatusOK, out)
}

type importResponse struct {
	Imported int `json:"imported"`
}

// Import handles POST /api/v1/challenges/import
// Supports multipart HTML upload (field: file) and JSON body with { "html": "..." }.
func (h *ChallengesHandler) Import(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	var htmlBytes []byte
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid multipart form"})
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "missing file field"})
			return
		}
		defer func() { _ = file.Close() }()
		b, err := io.ReadAll(io.LimitReader(file, 10<<20))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "failed to read file"})
			return
		}
		htmlBytes = b
	} else {
		var body struct {
			HTML string `json:"html"`
			URL  string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
			return
		}
		if strings.TrimSpace(body.HTML) != "" {
			htmlBytes = []byte(body.HTML)
		} else if strings.TrimSpace(body.URL) != "" {
			resp, err := http.Get(body.URL) //nolint:gosec // user-provided URL for optional import
			if err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "failed to fetch url"})
				return
			}
			defer resp.Body.Close()
			b, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
			if err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "failed to read response"})
				return
			}
			htmlBytes = b
		}
	}

	htmlBytes = bytes.TrimSpace(htmlBytes)
	if len(htmlBytes) == 0 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "no html provided"})
		return
	}

	parsed, err := challenges.ParseTrophyCaseHTML(string(htmlBytes))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "failed to parse html"})
		return
	}

	imported := 0
	for _, p := range parsed {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		c := &storage.Challenge{
			AthleteID:      athlete.ID,
			Name:           name,
			Slug:           strings.TrimSpace(p.Slug),
			BadgeURL:       strings.TrimSpace(p.BadgeURL),
			CompletionDate: ptrSQLiteTime(p.CompletionDate),
			Month:          strings.TrimSpace(p.Month),
			CreatedAt:      storage.SQLiteTime{Time: time.Now()},
		}
		if err := h.repo.Upsert(r.Context(), c); err != nil {
			writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to store challenges"})
			return
		}
		imported++
	}

	writeJSON(w, http.StatusOK, importResponse{Imported: imported})
}

func ptrSQLiteTime(t *time.Time) *storage.SQLiteTime {
	if t == nil {
		return nil
	}
	return &storage.SQLiteTime{Time: *t}
}
