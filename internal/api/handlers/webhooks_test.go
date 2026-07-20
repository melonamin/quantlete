package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

func TestStravaWebhookHandler_Validate(t *testing.T) {
	tests := []struct {
		name           string
		mode           string
		verifyToken    string
		challenge      string
		configToken    string
		wantStatusCode int
		wantChallenge  string
	}{
		{
			name:           "valid subscription request",
			mode:           "subscribe",
			verifyToken:    "test-token",
			challenge:      "test-challenge-123",
			configToken:    "test-token",
			wantStatusCode: http.StatusOK,
			wantChallenge:  "test-challenge-123",
		},
		{
			name:           "invalid mode",
			mode:           "invalid",
			verifyToken:    "test-token",
			challenge:      "test-challenge",
			configToken:    "test-token",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "empty mode",
			mode:           "",
			verifyToken:    "test-token",
			challenge:      "test-challenge",
			configToken:    "test-token",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "missing challenge",
			mode:           "subscribe",
			verifyToken:    "test-token",
			challenge:      "",
			configToken:    "test-token",
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "token mismatch",
			mode:           "subscribe",
			verifyToken:    "wrong-token",
			challenge:      "test-challenge",
			configToken:    "test-token",
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "webhook token not configured",
			mode:           "subscribe",
			verifyToken:    "test-token",
			challenge:      "test-challenge",
			configToken:    "",
			wantStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Strava: config.StravaConfig{
					WebhookVerifyToken: tt.configToken,
				},
			}

			handler := NewStravaWebhookHandler(cfg, nil, nil, nil, nil)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/strava", http.NoBody)
			q := req.URL.Query()
			q.Set("hub.mode", tt.mode)
			q.Set("hub.verify_token", tt.verifyToken)
			q.Set("hub.challenge", tt.challenge)
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler.Validate(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("Validate() status = %d, want %d", rr.Code, tt.wantStatusCode)
			}

			if tt.wantStatusCode == http.StatusOK && tt.wantChallenge != "" {
				var resp StravaWebhookValidationResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if resp.Challenge != tt.wantChallenge {
					t.Errorf("Validate() challenge = %q, want %q", resp.Challenge, tt.wantChallenge)
				}
			}
		})
	}
}

func TestStravaWebhookHandler_Receive(t *testing.T) {
	tests := []struct {
		name           string
		event          StravaWebhookEvent
		configSubID    int64
		wantStatusCode int
		wantReceived   bool
	}{
		{
			name: "valid activity create event",
			event: StravaWebhookEvent{
				ObjectType:     "activity",
				ObjectID:       12345,
				AspectType:     "create",
				OwnerID:        67890,
				SubscriptionID: 11111,
				EventTime:      1234567890,
			},
			configSubID:    11111,
			wantStatusCode: http.StatusOK,
			wantReceived:   true,
		},
		{
			name: "subscription ID mismatch",
			event: StravaWebhookEvent{
				ObjectType:     "activity",
				ObjectID:       12345,
				AspectType:     "create",
				OwnerID:        67890,
				SubscriptionID: 22222,
				EventTime:      1234567890,
			},
			configSubID:    11111,
			wantStatusCode: http.StatusForbidden,
			wantReceived:   false,
		},
		{
			name: "no subscription ID configured (acknowledges and drops)",
			event: StravaWebhookEvent{
				ObjectType:     "activity",
				ObjectID:       12345,
				AspectType:     "create",
				OwnerID:        67890,
				SubscriptionID: 99999,
				EventTime:      1234567890,
			},
			configSubID:    0,
			wantStatusCode: http.StatusOK,
			wantReceived:   true,
		},
		{
			name: "activity update event",
			event: StravaWebhookEvent{
				ObjectType:     "activity",
				ObjectID:       12345,
				AspectType:     "update",
				OwnerID:        67890,
				SubscriptionID: 11111,
				Updates:        map[string]any{"title": "New Title"},
			},
			configSubID:    11111,
			wantStatusCode: http.StatusOK,
			wantReceived:   true,
		},
		{
			name: "activity delete event",
			event: StravaWebhookEvent{
				ObjectType:     "activity",
				ObjectID:       12345,
				AspectType:     "delete",
				OwnerID:        67890,
				SubscriptionID: 11111,
			},
			configSubID:    11111,
			wantStatusCode: http.StatusOK,
			wantReceived:   true,
		},
		{
			name: "athlete object type (ignored)",
			event: StravaWebhookEvent{
				ObjectType:     "athlete",
				ObjectID:       67890,
				AspectType:     "update",
				OwnerID:        67890,
				SubscriptionID: 11111,
			},
			configSubID:    11111,
			wantStatusCode: http.StatusOK,
			wantReceived:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Strava: config.StravaConfig{
					WebhookSubscriptionID: tt.configSubID,
				},
			}

			handler := NewStravaWebhookHandler(cfg, nil, nil, nil, nil)

			body, err := json.Marshal(tt.event)
			if err != nil {
				t.Fatalf("failed to marshal event: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/strava", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Receive(rr, req)

			if rr.Code != tt.wantStatusCode {
				t.Errorf("Receive() status = %d, want %d", rr.Code, tt.wantStatusCode)
			}

			if tt.wantReceived {
				var resp map[string]bool
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if !resp["received"] {
					t.Error("Receive() should return {received: true}")
				}
			}
		})
	}
}

func TestStravaWebhookHandler_Receive_InvalidJSON(t *testing.T) {
	cfg := &config.Config{}
	handler := NewStravaWebhookHandler(cfg, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/strava", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Receive(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Receive() with invalid JSON status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestStravaWebhookHandler_CreateImportContextOutlivesHandler(t *testing.T) {
	const (
		athleteID      = int64(67890)
		subscriptionID = int64(11111)
	)

	importClient := newBlockingWebhookImportClient(athleteID)
	imp := newWebhookTestImporter(t, importClient)
	stravaClient := newAuthenticatedWebhookStravaClient(athleteID)
	handler := NewStravaWebhookHandler(&config.Config{
		Strava: config.StravaConfig{WebhookSubscriptionID: subscriptionID},
	}, imp, nil, nil, stravaClient)

	recorder := receiveWebhookEvent(t, handler, StravaWebhookEvent{
		ObjectType:     "activity",
		ObjectID:       12345,
		AspectType:     "create",
		OwnerID:        athleteID,
		SubscriptionID: subscriptionID,
	})
	if recorder.Code != http.StatusOK {
		t.Fatalf("Receive() status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var importCtx context.Context
	select {
	case importCtx = <-importClient.started:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook create did not start an import")
	}

	if got := imp.Progress().Status; got != importer.StatusRunning {
		t.Fatalf("import status = %q, want %q", got, importer.StatusRunning)
	}
	select {
	case <-importCtx.Done():
		t.Fatalf("import context was canceled after handler returned: %v", importCtx.Err())
	case <-time.After(100 * time.Millisecond):
	}

	close(importClient.release)
	waitForWebhookImportToStop(t, imp)
}

func TestStravaWebhookHandler_UnconfiguredSubscriptionDropsEvents(t *testing.T) {
	const (
		athleteID  = int64(67890)
		activityID = int64(12345)
	)

	db := newWebhookTestDB(t)
	activities := storage.NewActivityRepository(db)
	seedWebhookActivity(t, db, activities, athleteID, activityID)

	importClient := newBlockingWebhookImportClient(athleteID)
	imp := newWebhookTestImporter(t, importClient)
	handler := NewStravaWebhookHandler(
		&config.Config{},
		imp,
		activities,
		nil,
		newAuthenticatedWebhookStravaClient(athleteID),
	)

	for _, aspectType := range []string{"delete", "create"} {
		recorder := receiveWebhookEvent(t, handler, StravaWebhookEvent{
			ObjectType:     "activity",
			ObjectID:       activityID,
			AspectType:     aspectType,
			OwnerID:        athleteID,
			SubscriptionID: 99999,
		})
		if recorder.Code != http.StatusOK {
			t.Fatalf("Receive(%s) status = %d, want %d", aspectType, recorder.Code, http.StatusOK)
		}
	}

	select {
	case <-importClient.started:
		t.Fatal("unverified create event started an import")
	case <-time.After(100 * time.Millisecond):
	}
	if got := imp.Progress().Status; got != importer.StatusIdle {
		t.Errorf("import status = %q, want %q", got, importer.StatusIdle)
	}
	if activity, err := activities.GetByID(context.Background(), activityID); err != nil {
		t.Fatalf("get activity after unverified delete: %v", err)
	} else if activity == nil {
		t.Fatal("unverified delete event removed the activity")
	}
}

func TestStravaWebhookHandler_DeleteRequiresVerifiedSubscription(t *testing.T) {
	const (
		athleteID      = int64(67890)
		activityID     = int64(12345)
		subscriptionID = int64(11111)
	)

	db := newWebhookTestDB(t)
	activities := storage.NewActivityRepository(db)
	seedWebhookActivity(t, db, activities, athleteID, activityID)
	handler := NewStravaWebhookHandler(
		&config.Config{Strava: config.StravaConfig{WebhookSubscriptionID: subscriptionID}},
		nil,
		activities,
		nil,
		newAuthenticatedWebhookStravaClient(athleteID),
	)

	mismatch := receiveWebhookEvent(t, handler, StravaWebhookEvent{
		ObjectType:     "activity",
		ObjectID:       activityID,
		AspectType:     "delete",
		OwnerID:        athleteID,
		SubscriptionID: subscriptionID + 1,
	})
	if mismatch.Code != http.StatusForbidden {
		t.Fatalf("mismatched subscription status = %d, want %d", mismatch.Code, http.StatusForbidden)
	}
	if activity, err := activities.GetByID(context.Background(), activityID); err != nil {
		t.Fatalf("get activity after mismatched delete: %v", err)
	} else if activity == nil {
		t.Fatal("mismatched subscription delete removed the activity")
	}

	verified := receiveWebhookEvent(t, handler, StravaWebhookEvent{
		ObjectType:     "activity",
		ObjectID:       activityID,
		AspectType:     "delete",
		OwnerID:        athleteID,
		SubscriptionID: subscriptionID,
	})
	if verified.Code != http.StatusOK {
		t.Fatalf("verified subscription status = %d, want %d", verified.Code, http.StatusOK)
	}
	waitForWebhookCondition(t, "verified delete to remove activity", func() bool {
		var count int
		err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM activities WHERE id = ?", activityID).Scan(&count)
		return err == nil && count == 0
	})
}

func TestStravaWebhookHandler_ReceivePromptWhenWorkersSaturated(t *testing.T) {
	const (
		athleteID      = int64(67890)
		subscriptionID = int64(11111)
	)

	importClient := newBlockingWebhookImportClient(athleteID)
	imp := newWebhookTestImporter(t, importClient)
	handler := NewStravaWebhookHandler(
		&config.Config{Strava: config.StravaConfig{WebhookSubscriptionID: subscriptionID}},
		imp,
		nil,
		nil,
		newAuthenticatedWebhookStravaClient(athleteID),
	)
	for range cap(handler.workerSem) {
		handler.workerSem <- struct{}{}
	}

	startedAt := time.Now()
	recorder := receiveWebhookEvent(t, handler, StravaWebhookEvent{
		ObjectType:     "activity",
		ObjectID:       12345,
		AspectType:     "create",
		OwnerID:        athleteID,
		SubscriptionID: subscriptionID,
	})
	elapsed := time.Since(startedAt)
	if recorder.Code != http.StatusOK {
		t.Fatalf("Receive() status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Receive() took %v with saturated workers, want at most 500ms", elapsed)
	}

	// Free a slot and verify the saturated event was dropped rather than left in
	// a parked goroutine waiting to start later.
	<-handler.workerSem
	select {
	case <-importClient.started:
		close(importClient.release)
		t.Fatal("saturated webhook event remained queued in a goroutine")
	case <-time.After(100 * time.Millisecond):
	}
}

type blockingWebhookImportClient struct {
	athlete *importer.Athlete
	started chan context.Context
	release chan struct{}
}

func newBlockingWebhookImportClient(athleteID int64) *blockingWebhookImportClient {
	return &blockingWebhookImportClient{
		athlete: &importer.Athlete{ID: athleteID},
		started: make(chan context.Context, 1),
		release: make(chan struct{}),
	}
}

func (c *blockingWebhookImportClient) GetAthlete() *importer.Athlete {
	return c.athlete
}

func (c *blockingWebhookImportClient) GetActivities(ctx context.Context, _ importer.GetActivitiesOptions) ([]importer.Activity, error) {
	c.started <- ctx
	select {
	case <-c.release:
		return nil, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (c *blockingWebhookImportClient) GetActivity(context.Context, int64) (*importer.Activity, error) {
	return nil, nil
}

func (c *blockingWebhookImportClient) GetActivityStreams(context.Context, int64, []string) (*importer.StreamSet, error) {
	return nil, nil
}

func (c *blockingWebhookImportClient) GetGear(context.Context, string) (*importer.Gear, error) {
	return nil, nil
}

func (c *blockingWebhookImportClient) GetSegment(context.Context, int64) (*importer.Segment, error) {
	return nil, nil
}

func (c *blockingWebhookImportClient) GetActivityPhotos(context.Context, int64) ([]importer.Photo, error) {
	return nil, nil
}

func (c *blockingWebhookImportClient) RateLimitInfo() importer.RateLimitInfo {
	return importer.RateLimitInfo{}
}

func newWebhookTestImporter(t *testing.T, client importer.StravaClient) *importer.Importer {
	t.Helper()
	storageAdapter := importer.NewServerStorageAdapter(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	imp, err := importer.New(client, storageAdapter)
	if err != nil {
		t.Fatalf("create test importer: %v", err)
	}
	return imp
}

func newAuthenticatedWebhookStravaClient(athleteID int64) *strava.Client {
	client := strava.NewClient(&config.StravaConfig{})
	client.SetToken(nil, &strava.Athlete{ID: athleteID})
	return client
}

func newWebhookTestDB(t *testing.T) *storage.DB {
	t.Helper()
	db, err := storage.OpenInMemory()
	if err != nil {
		t.Fatalf("open in-memory database: %v", err)
	}
	if err := db.Migrate(); err != nil {
		_ = db.Close()
		t.Fatalf("migrate in-memory database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func seedWebhookActivity(t *testing.T, db *storage.DB, activities *storage.ActivityRepository, athleteID, activityID int64) {
	t.Helper()
	if err := storage.NewAthleteRepository(db).Upsert(context.Background(), &storage.Athlete{ID: athleteID}); err != nil {
		t.Fatalf("seed athlete: %v", err)
	}
	now := storage.SQLiteTime{Time: time.Now().UTC()}
	if err := activities.Upsert(context.Background(), &storage.Activity{
		ID:             activityID,
		AthleteID:      athleteID,
		Name:           "Webhook test activity",
		SportType:      "Run",
		StartDate:      now,
		StartDateLocal: now,
	}); err != nil {
		t.Fatalf("seed activity: %v", err)
	}
}

func receiveWebhookEvent(t *testing.T, handler *StravaWebhookHandler, event StravaWebhookEvent) *httptest.ResponseRecorder {
	t.Helper()
	body, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal webhook event: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/strava", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.Receive(recorder, req)
	return recorder
}

func waitForWebhookImportToStop(t *testing.T, imp *importer.Importer) {
	t.Helper()
	waitForWebhookCondition(t, "import to stop", func() bool {
		return imp.Progress().Status != importer.StatusRunning
	})
}

func waitForWebhookCondition(t *testing.T, description string, condition func() bool) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if condition() {
			return
		}
		select {
		case <-deadline.C:
			t.Fatalf("timed out waiting for %s", description)
		case <-ticker.C:
		}
	}
}
