package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type StravaWebhookHandler struct {
	cfg        *config.Config
	importer   *importer.Importer
	strava     *strava.Client
	activities *storage.ActivityRepository
	settings   *storage.SettingsRepository
	workerSem  chan struct{} // bounds concurrent event processing
}

func NewStravaWebhookHandler(cfg *config.Config, imp *importer.Importer, activities *storage.ActivityRepository, settingsRepo *storage.SettingsRepository, stravaClient *strava.Client) *StravaWebhookHandler {
	return &StravaWebhookHandler{
		cfg:        cfg,
		importer:   imp,
		strava:     stravaClient,
		activities: activities,
		settings:   settingsRepo,
		workerSem:  make(chan struct{}, 5), // max 5 concurrent event processors
	}
}

// StravaWebhookValidationResponse matches Strava's expected response body:
// {"hub.challenge":"..."}
type StravaWebhookValidationResponse struct {
	Challenge string `json:"hub.challenge"`
}

// Validate handles GET /api/v1/webhooks/strava (subscription validation).
func (h *StravaWebhookHandler) Validate(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	if mode != "subscribe" || challenge == "" {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid validation request"})
		return
	}
	if h.cfg.Strava.WebhookVerifyToken == "" {
		writeJSON(w, http.StatusServiceUnavailable, ErrorResponse{Error: "webhook verify token not configured"})
		return
	}
	if token != h.cfg.Strava.WebhookVerifyToken {
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "invalid verify token"})
		return
	}

	writeJSON(w, http.StatusOK, StravaWebhookValidationResponse{Challenge: challenge})
}

type StravaWebhookEvent struct {
	ObjectType     string         `json:"object_type"`
	ObjectID       int64          `json:"object_id"`
	AspectType     string         `json:"aspect_type"`
	OwnerID        int64          `json:"owner_id"`
	SubscriptionID int64          `json:"subscription_id"`
	EventTime      int64          `json:"event_time"`
	Updates        map[string]any `json:"updates,omitempty"`
}

// Receive handles POST /api/v1/webhooks/strava (event delivery).
//
// Security note: Strava webhooks do not include HMAC signatures like GitHub or Stripe.
// We validate events by checking the subscription_id matches our configured subscription.
// The subscription ID is only known to Strava and this application.
func (h *StravaWebhookHandler) Receive(w http.ResponseWriter, r *http.Request) {
	var e StravaWebhookEvent
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON"})
		return
	}

	// Validate subscription ID if configured (recommended for production)
	if h.cfg.Strava.WebhookSubscriptionID != 0 && e.SubscriptionID != h.cfg.Strava.WebhookSubscriptionID {
		slog.Warn("webhook: subscription ID mismatch", "expected", h.cfg.Strava.WebhookSubscriptionID, "got", e.SubscriptionID)
		writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "invalid subscription"})
		return
	}

	// Acknowledge quickly; Strava expects a fast 2xx response.
	writeJSON(w, http.StatusOK, map[string]bool{"received": true})

	// Process asynchronously with bounded concurrency
	go func() {
		h.workerSem <- struct{}{}        // acquire
		defer func() { <-h.workerSem }() // release
		h.processEvent(e)
	}()
}

func (h *StravaWebhookHandler) processEvent(e StravaWebhookEvent) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if e.ObjectType != "activity" {
		return
	}

	athlete := h.strava.GetAthlete()
	if athlete == nil {
		slog.Info("webhook event received but not authenticated; ignoring", "object_id", e.ObjectID, "aspect_type", e.AspectType)
		return
	}
	if e.OwnerID != 0 && e.OwnerID != athlete.ID {
		slog.Info("webhook event owner mismatch; ignoring", "event_owner_id", e.OwnerID, "athlete_id", athlete.ID)
		return
	}

	if h.settings != nil {
		s, err := h.settings.Get(ctx, athlete.ID)
		if err != nil {
			slog.Warn("webhook: failed to load settings; ignoring", "error", err)
			return
		}
		if !s.Scheduler.Push.Enabled {
			return
		}
	}

	switch e.AspectType {
	case "create":
		// Trigger an incremental sync (uses watermark).
		if err := h.importer.Start(ctx, importer.ImportOptions{}); err != nil {
			slog.Info("webhook create: sync not started", "activity_id", e.ObjectID, "error", err)
		}
	case "update":
		act, err := h.strava.GetActivity(ctx, e.ObjectID)
		if err != nil {
			slog.Warn("webhook update: failed to fetch activity", "activity_id", e.ObjectID, "error", err)
			return
		}

		stored := importer.ConvertActivity(act, athlete.ID)
		if err := h.activities.Upsert(ctx, stored); err != nil {
			slog.Warn("webhook update: failed to upsert activity", "activity_id", e.ObjectID, "error", err)
		}
	case "delete":
		if err := h.activities.DeleteByID(ctx, athlete.ID, e.ObjectID); err != nil {
			slog.Warn("webhook delete: failed to delete activity", "activity_id", e.ObjectID, "error", err)
		}
	default:
		// Ignore other aspect types.
	}
}
