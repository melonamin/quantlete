package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/melonamin/quantlete/internal/config"
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

			req := httptest.NewRequest(http.MethodGet, "/api/v1/webhooks/strava", nil)
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
		name             string
		event            StravaWebhookEvent
		configSubID      int64
		wantStatusCode   int
		wantReceived     bool
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
			name: "no subscription ID configured (allows any)",
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
