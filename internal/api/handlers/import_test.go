package handlers

import (
	"net/http/httptest"
	"testing"
)

// mockFlusher implements http.Flusher for testing.
type mockFlusher struct {
	*httptest.ResponseRecorder
	flushCount int
}

func (m *mockFlusher) Flush() {
	m.flushCount++
}

func TestSendSSEEvent(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		data      interface{}
		wantBody  string
	}{
		{
			name:      "sync:progress event",
			eventType: "sync:progress",
			data:      map[string]int{"activities_done": 10, "activities_total": 100},
			wantBody:  "event: sync:progress\ndata: {\"activities_done\":10,\"activities_total\":100}\n\n",
		},
		{
			name:      "sync:complete event",
			eventType: "sync:complete",
			data:      map[string]string{"status": "completed"},
			wantBody:  "event: sync:complete\ndata: {\"status\":\"completed\"}\n\n",
		},
		{
			name:      "data:changed event",
			eventType: "data:changed",
			data:      map[string]bool{"activities": true, "streams": false},
			wantBody:  "event: data:changed\ndata: {\"activities\":true,\"streams\":false}\n\n",
		},
		{
			name:      "empty data",
			eventType: "sync:complete",
			data:      map[string]string{},
			wantBody:  "event: sync:complete\ndata: {}\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			flusher := &mockFlusher{ResponseRecorder: rr}

			err := sendSSEEvent(rr, flusher, tt.eventType, tt.data)
			if err != nil {
				t.Fatalf("sendSSEEvent() error = %v", err)
			}

			if got := rr.Body.String(); got != tt.wantBody {
				t.Errorf("sendSSEEvent() body = %q, want %q", got, tt.wantBody)
			}

			if flusher.flushCount != 1 {
				t.Errorf("sendSSEEvent() flush count = %d, want 1", flusher.flushCount)
			}
		})
	}
}

func TestSendSSEEvent_InvalidJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	flusher := &mockFlusher{ResponseRecorder: rr}

	// Functions cannot be marshaled to JSON
	err := sendSSEEvent(rr, flusher, "test", func() {})
	if err == nil {
		t.Error("sendSSEEvent() with unmarshalable data should return error")
	}
}
