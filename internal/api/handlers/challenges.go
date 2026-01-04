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

	"github.com/melonamin/quantlete/internal/challenges"
	"github.com/melonamin/quantlete/internal/pagination"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// ChallengesHandler handles challenges-related endpoints.
type ChallengesHandler struct {
	svc        *services.ChallengesService
	strava     *strava.Client
	downloader *challenges.BadgeDownloader
}

// NewChallengesHandler creates a new challenges handler.
func NewChallengesHandler(svc *services.ChallengesService, stravaClient *strava.Client, dataDir string) *ChallengesHandler {
	return &ChallengesHandler{
		svc:        svc,
		strava:     stravaClient,
		downloader: challenges.NewBadgeDownloader(dataDir),
	}
}

// List handles GET /api/v1/challenges
func (h *ChallengesHandler) List(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	q := r.URL.Query()
	params := pagination.ParseQueryParams(q)

	result, err := h.svc.List(r.Context(), services.ListChallengesInput{
		AthleteID: athlete.ID,
		Month:     strings.TrimSpace(q.Get("month")),
		Page:      params.Page,
		PerPage:   params.PerPage,
		OrderBy:   params.OrderBy,
		OrderDir:  params.OrderDir,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	shared.WriteSuccess(w, result)
}

type importResponse struct {
	Imported int `json:"imported"`
}

// Import handles POST /api/v1/challenges/import
// Supports multipart HTML upload (field: file) and JSON body with { "html": "..." }.
func (h *ChallengesHandler) Import(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
		return
	}

	var htmlBytes []byte
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid multipart form"))
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("missing file field"))
			return
		}
		defer func() { _ = file.Close() }()
		b, err := io.ReadAll(io.LimitReader(file, 10<<20))
		if err != nil {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("failed to read file"))
			return
		}
		htmlBytes = b
	} else {
		var body struct {
			HTML string `json:"html"`
			URL  string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("invalid JSON"))
			return
		}
		if strings.TrimSpace(body.HTML) != "" {
			htmlBytes = []byte(body.HTML)
		} else if strings.TrimSpace(body.URL) != "" {
			// Validate URL to prevent SSRF attacks
			if err := ValidateImportURL(body.URL); err != nil {
				shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorResponse(err))
				return
			}
			resp, err := http.Get(body.URL) //nolint:gosec // URL validated by ValidateImportURL
			if err != nil {
				shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("failed to fetch url"))
				return
			}
			defer resp.Body.Close()
			b, err := io.ReadAll(io.LimitReader(resp.Body, 10<<20))
			if err != nil {
				shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("failed to read response"))
				return
			}
			htmlBytes = b
		}
	}

	htmlBytes = bytes.TrimSpace(htmlBytes)
	if len(htmlBytes) == 0 {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("no html provided"))
		return
	}

	parsed, err := challenges.ParseTrophyCaseHTML(string(htmlBytes))
	if err != nil {
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("failed to parse html"))
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
		if err := h.svc.Import(r.Context(), c); err != nil {
			handleServiceError(w, err)
			return
		}
		imported++
	}

	shared.WriteSuccess(w, importResponse{Imported: imported})
}

// ImportFromProfile handles POST /api/v1/challenges/import-profile
// Scrapes challenges from a Strava public profile page.
func (h *ChallengesHandler) ImportFromProfile(w http.ResponseWriter, r *http.Request) {
	athlete := h.strava.GetAthlete()
	if athlete == nil {
		shared.WriteJSONResponse(w, http.StatusUnauthorized, shared.ErrorMessage("not authenticated"))
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
		shared.WriteJSONResponse(w, http.StatusBadRequest, shared.ErrorMessage("failed to fetch public profile: "+err.Error()))
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
		if err := h.svc.Import(r.Context(), c); err != nil {
			handleServiceError(w, err)
			return
		}
		imported++
	}

	shared.WriteSuccess(w, importResponse{Imported: imported})
}

func ptrSQLiteTime(t *time.Time) *storage.SQLiteTime {
	if t == nil {
		return nil
	}
	return &storage.SQLiteTime{Time: *t}
}
