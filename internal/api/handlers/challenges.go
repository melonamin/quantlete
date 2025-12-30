package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sasha/stata/internal/challenges"
	"github.com/sasha/stata/internal/pagination"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

type challengesListResponse struct {
	Data       []challengeResponse `json:"data"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PerPage    int                 `json:"per_page"`
	TotalPages int                 `json:"total_pages"`
}

type ChallengesHandler struct {
	repo       *storage.ChallengeRepository
	strava     *strava.Client
	downloader *challenges.BadgeDownloader
}

func NewChallengesHandler(repo *storage.ChallengeRepository, stravaClient *strava.Client, dataDir string) *ChallengesHandler {
	return &ChallengesHandler{
		repo:       repo,
		strava:     stravaClient,
		downloader: challenges.NewBadgeDownloader(dataDir),
	}
}

type challengeResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug,omitempty"`
	BadgeURL       string  `json:"badge_url,omitempty"`
	LocalBadgeURL  string  `json:"local_badge_url,omitempty"`
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
		LocalBadgeURL:  c.LocalBadgeURL,
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

	q := r.URL.Query()
	f := storage.ChallengeFilters{
		Month:       strings.TrimSpace(q.Get("month")),
		QueryParams: pagination.ParseQueryParams(q),
	}

	result, err := h.repo.ListPaginated(r.Context(), athlete.ID, f)
	if err != nil {
		slog.Error("failed to list challenges", "error", err, "athlete_id", athlete.ID, "filters", f)
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "failed to fetch challenges"})
		return
	}

	out := make([]challengeResponse, 0, len(result.Items))
	for _, c := range result.Items {
		out = append(out, challengeToResponse(c))
	}

	writeJSON(w, http.StatusOK, challengesListResponse{
		Data:       out,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
	})
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
			// Validate URL to prevent SSRF attacks
			if err := ValidateImportURL(body.URL); err != nil {
				writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
				return
			}
			resp, err := http.Get(body.URL) //nolint:gosec // URL validated by ValidateImportURL
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

		// Generate deterministic ID based on completion date and name
		completionDate := time.Now()
		if p.CompletionDate != nil {
			completionDate = *p.CompletionDate
		}
		challengeID := challenges.GenerateChallengeID(completionDate, name)

		// Download badge image locally
		var localBadgeURL string
		if badgeURL := strings.TrimSpace(p.BadgeURL); badgeURL != "" {
			if path, err := h.downloader.Download(badgeURL); err != nil {
				slog.Warn("failed to download challenge badge", "challenge", name, "error", err)
			} else {
				localBadgeURL = path
			}
		}

		c := &storage.Challenge{
			ID:             challengeID,
			AthleteID:      athlete.ID,
			Name:           name,
			Slug:           strings.TrimSpace(p.Slug),
			BadgeURL:       strings.TrimSpace(p.BadgeURL),
			LocalBadgeURL:  localBadgeURL,
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

// ImportFromProfile handles POST /api/v1/challenges/import-profile
// Scrapes challenges from a Strava public profile page.
func (h *ChallengesHandler) ImportFromProfile(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "not authenticated"})
		return
	}

	// Get athlete ID from query param or use the logged-in athlete's ID
	athleteID := strconv.FormatInt(athlete.ID, 10)
	if qID := strings.TrimSpace(r.URL.Query().Get("athlete_id")); qID != "" {
		athleteID = qID
	}

	parsed, err := challenges.FetchPublicProfile(athleteID)
	if err != nil {
		slog.Error("failed to fetch public profile", "error", err, "athlete_id", athleteID)
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "failed to fetch public profile: " + err.Error()})
		return
	}

	imported := 0
	for _, p := range parsed {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}

		// Generate deterministic ID based on completion date and name
		challengeID := challenges.GenerateChallengeID(p.CompletedOn, name)

		// Download badge image locally
		var localBadgeURL string
		if badgeURL := strings.TrimSpace(p.BadgeURL); badgeURL != "" {
			if path, err := h.downloader.Download(badgeURL); err != nil {
				slog.Warn("failed to download challenge badge", "challenge", name, "error", err)
			} else {
				localBadgeURL = path
			}
		}

		month := p.CompletedOn.Format("2006-01")
		c := &storage.Challenge{
			ID:             challengeID,
			AthleteID:      athlete.ID,
			Name:           name,
			Slug:           strings.TrimSpace(p.Slug),
			BadgeURL:       strings.TrimSpace(p.BadgeURL),
			LocalBadgeURL:  localBadgeURL,
			CompletionDate: &storage.SQLiteTime{Time: p.CompletedOn},
			Month:          month,
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
