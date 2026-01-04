package importer

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/strava"
)

func TestAddError_BoundedList(t *testing.T) {
	imp := &Importer{
		state: &ImportState{},
	}

	// Add more than maxAggregatedErrors
	for i := 0; i < maxAggregatedErrors+50; i++ {
		imp.addError("error %d", i)
	}

	// Verify error list is capped
	if len(imp.state.Errors) != maxAggregatedErrors {
		t.Errorf("expected %d errors, got %d", maxAggregatedErrors, len(imp.state.Errors))
	}

	// Verify failed count still tracks all failures
	expectedFailedCount := maxAggregatedErrors + 50
	if imp.state.FailedCount != expectedFailedCount {
		t.Errorf("expected FailedCount=%d, got %d", expectedFailedCount, imp.state.FailedCount)
	}
}

func TestAddError_Sequential(t *testing.T) {
	// Note: addError is designed for single-threaded access (import goroutine only).
	// This test verifies sequential error accumulation works correctly.
	imp := &Importer{
		state: &ImportState{},
	}

	totalErrors := 200

	// Add errors sequentially (as would happen in practice)
	for i := 0; i < totalErrors; i++ {
		imp.addError("error %d", i)
	}

	// Verify all errors were tracked
	if imp.state.FailedCount != totalErrors {
		t.Errorf("expected FailedCount=%d, got %d", totalErrors, imp.state.FailedCount)
	}

	// Errors should be capped at maxAggregatedErrors
	if len(imp.state.Errors) > maxAggregatedErrors {
		t.Errorf("error list exceeded max: %d > %d", len(imp.state.Errors), maxAggregatedErrors)
	}
}

func TestUpdateProgress_CopiesState(t *testing.T) {
	// Note: State access is now single-threaded by design (import goroutine only)
	// This test verifies updateProgress() correctly copies state to progress
	imp := &Importer{
		state: &ImportState{
			Phase:           PhaseActivities,
			ActivitiesTotal: 100,
			ActivitiesDone:  50,
		},
		eta:      NewETAEstimator(),
		progress: Progress{Status: StatusRunning},
	}

	// Update state
	imp.state.ActivitiesDone = 75
	imp.state.Phase = PhaseStreams

	// Call updateProgress to copy state to progress
	imp.updateProgress()

	// Verify progress was updated
	if imp.progress.ActivitiesDone != 75 {
		t.Errorf("expected ActivitiesDone = 75, got %d", imp.progress.ActivitiesDone)
	}
	if imp.progress.Phase != PhaseStreams {
		t.Errorf("expected Phase = %s, got %s", PhaseStreams, imp.progress.Phase)
	}
}

func TestImportState_RemainingAPICalls(t *testing.T) {
	tests := []struct {
		name     string
		state    ImportState
		expected int
	}{
		{
			name: "activities phase",
			state: ImportState{
				Phase:       PhaseActivities,
				ActivityIDs: []int64{1, 2, 3},
			},
			// 10 (pages estimate) + gear + streams + details + segments + photos
			expected: 10 + 0 + 3 + 3 + 0 + 3, // 19
		},
		{
			name: "streams phase midway",
			state: ImportState{
				Phase:            PhaseStreams,
				ActivityIDs:      []int64{1, 2, 3, 4, 5},
				StreamsLastIndex: 2,
			},
			// remaining streams (5-2=3) + details (5) + segments (0) + photos (5)
			expected: 3 + 5 + 0 + 5, // 13
		},
		{
			name: "skip streams",
			state: ImportState{
				Phase:       PhaseStreams,
				ActivityIDs: []int64{1, 2, 3},
				SkipStreams: true,
			},
			// With SkipStreams=true in PhaseStreams:
			// streams (0 - skipped) + details (3) + segments (0) + photos (3)
			expected: 0 + 3 + 0 + 3, // 6
		},
		{
			name: "completed phase",
			state: ImportState{
				Phase: PhaseCompleted,
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.state.RemainingAPICalls()
			if result != tt.expected {
				t.Errorf("RemainingAPICalls() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestProgress_StatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus Status
		expectValid   bool
	}{
		{"idle", StatusIdle, true},
		{"running", StatusRunning, true},
		{"completed", StatusCompleted, true},
		{"failed", StatusFailed, true},
		{"canceled", StatusCanceled, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Progress{Status: tt.initialStatus}
			if (p.Status != "") != tt.expectValid {
				t.Errorf("unexpected status validation for %s", tt.name)
			}
		})
	}
}

func TestImportPhase_Values(t *testing.T) {
	phases := []ImportPhase{
		PhaseIdle,
		PhaseActivities,
		PhaseGear,
		PhaseStreams,
		PhaseActivityDetails,
		PhaseSegmentDetails,
		PhasePhotos,
		PhaseCompleted,
	}

	seen := make(map[ImportPhase]bool)
	for _, p := range phases {
		if seen[p] {
			t.Errorf("duplicate phase value: %s", p)
		}
		seen[p] = true

		if p == "" {
			t.Error("phase value should not be empty")
		}
	}
}

func TestFormatETA(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero", 0, "< 1 min"},
		{"seconds", 30 * time.Second, "< 1 min"},
		{"one minute", time.Minute, "1 min"},
		{"minutes and seconds", 90 * time.Second, "1 min"},
		{"two minutes", 2 * time.Minute, "2 mins"},
		{"hours", 2 * time.Hour, "2 hours"},
		{"one hour", time.Hour, "1 hour"},
		{"hours and minutes", 2*time.Hour + 30*time.Minute, "2 hours 30 mins"},
		{"one hour one min", time.Hour + time.Minute, "1 hour 1 min"},
		{"one day", 25 * time.Hour, "1 day 1 hour"},
		{"multiple days", 50 * time.Hour, "2 days 2 hours"},
		{"exact 24 hours", 24 * time.Hour, "24 hours"}, // exactly 24h not treated as day (> 24 needed)
		{"negative", -time.Second, "< 1 min"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatETA(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatETA(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestETAEstimator_Defaults(t *testing.T) {
	est := NewETAEstimator()

	// Check default values
	if est.FifteenMinLimit != 100 {
		t.Errorf("expected FifteenMinLimit=100, got %d", est.FifteenMinLimit)
	}
	if est.DailyLimit != 1000 {
		t.Errorf("expected DailyLimit=1000, got %d", est.DailyLimit)
	}
	if est.AvgCallDuration != 500*time.Millisecond {
		t.Errorf("expected AvgCallDuration=500ms, got %v", est.AvgCallDuration)
	}
}

func TestETAEstimator_ZeroRemaining(t *testing.T) {
	est := NewETAEstimator()

	// Zero remaining should give zero ETA
	eta := est.EstimateCompletion(0)
	if eta != 0 {
		t.Errorf("expected 0 ETA for zero remaining, got %v", eta)
	}

	// Negative remaining should also give zero
	eta = est.EstimateCompletion(-5)
	if eta != 0 {
		t.Errorf("expected 0 ETA for negative remaining, got %v", eta)
	}
}

func TestETAEstimator_SmallBatch(t *testing.T) {
	est := NewETAEstimator()

	// Small number of calls should fit in current window
	// With 100 limit and 0 used, 10 calls should complete quickly
	eta := est.EstimateCompletion(10)
	expected := 10 * 500 * time.Millisecond // 10 calls * 500ms avg
	if eta != expected {
		t.Errorf("expected ETA=%v for 10 calls, got %v", expected, eta)
	}
}

func TestETAEstimator_UpdateFromRateLimits(t *testing.T) {
	est := NewETAEstimator()
	now := time.Now()
	reset := now.Add(10 * time.Minute)
	dailyReset := now.Add(12 * time.Hour)

	est.UpdateFromRateLimits(50, 100, reset, 500, 1000, dailyReset)

	if est.FifteenMinUsed != 50 {
		t.Errorf("expected FifteenMinUsed=50, got %d", est.FifteenMinUsed)
	}
	if est.FifteenMinRemaining() != 50 {
		t.Errorf("expected FifteenMinRemaining=50, got %d", est.FifteenMinRemaining())
	}
	if est.DailyUsed != 500 {
		t.Errorf("expected DailyUsed=500, got %d", est.DailyUsed)
	}
	if est.DailyRemaining() != 500 {
		t.Errorf("expected DailyRemaining=500, got %d", est.DailyRemaining())
	}
}

func TestIsResourceGoneError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name: "404 not found",
			err: &strava.APIError{
				StatusCode: 404,
				Status:     "Not Found",
				Path:       "/segments/123",
			},
			expected: true,
		},
		{
			name: "403 forbidden",
			err: &strava.APIError{
				StatusCode: 403,
				Status:     "Forbidden",
				Path:       "/activities/456",
			},
			expected: true,
		},
		{
			name: "500 server error",
			err: &strava.APIError{
				StatusCode: 500,
				Status:     "Internal Server Error",
				Path:       "/athlete",
			},
			expected: false,
		},
		{
			name: "401 unauthorized",
			err: &strava.APIError{
				StatusCode: 401,
				Status:     "Unauthorized",
				Path:       "/athlete",
			},
			expected: false,
		},
		{
			name: "429 rate limited",
			err: &strava.APIError{
				StatusCode: 429,
				Status:     "Too Many Requests",
				Path:       "/activities",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isResourceGoneError(tt.err)
			if result != tt.expected {
				t.Errorf("isResourceGoneError(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestSubscribe_Basic(t *testing.T) {
	imp := &Importer{}

	ch := make(chan Event, 10)
	unsubscribe := imp.Subscribe(ch)

	// Should have one subscriber
	imp.subscribersMu.RLock()
	if len(imp.subscribers) != 1 {
		t.Errorf("expected 1 subscriber, got %d", len(imp.subscribers))
	}
	imp.subscribersMu.RUnlock()

	// Unsubscribe should remove and close channel
	unsubscribe()

	imp.subscribersMu.RLock()
	if len(imp.subscribers) != 0 {
		t.Errorf("expected 0 subscribers after unsubscribe, got %d", len(imp.subscribers))
	}
	imp.subscribersMu.RUnlock()

	// Channel should be closed
	_, ok := <-ch
	if ok {
		t.Error("channel should be closed after unsubscribe")
	}
}

func TestSubscribe_MultipleUnsubscribeCalls(t *testing.T) {
	imp := &Importer{}

	ch := make(chan Event, 10)
	unsubscribe := imp.Subscribe(ch)

	// Multiple unsubscribe calls should be safe (no panic)
	unsubscribe()
	unsubscribe()
	unsubscribe()

	imp.subscribersMu.RLock()
	if len(imp.subscribers) != 0 {
		t.Errorf("expected 0 subscribers, got %d", len(imp.subscribers))
	}
	imp.subscribersMu.RUnlock()
}

func TestSubscribe_ConcurrentUnsubscribe(t *testing.T) {
	imp := &Importer{}

	// Subscribe multiple channels
	var unsubscribes []func()
	for i := 0; i < 100; i++ {
		ch := make(chan Event, 10)
		unsubscribes = append(unsubscribes, imp.Subscribe(ch))
	}

	// Unsubscribe all concurrently - should not panic
	var wg sync.WaitGroup
	for _, unsub := range unsubscribes {
		wg.Add(1)
		go func(unsub func()) {
			defer wg.Done()
			unsub()
		}(unsub)
	}
	wg.Wait()

	imp.subscribersMu.RLock()
	if len(imp.subscribers) != 0 {
		t.Errorf("expected 0 subscribers after concurrent unsubscribe, got %d", len(imp.subscribers))
	}
	imp.subscribersMu.RUnlock()
}

func TestEmitEvent_ToSubscribers(t *testing.T) {
	imp := &Importer{}

	ch1 := make(chan Event, 10)
	ch2 := make(chan Event, 10)
	unsub1 := imp.Subscribe(ch1)
	unsub2 := imp.Subscribe(ch2)
	defer unsub1()
	defer unsub2()

	// Emit an event
	event := Event{Type: EventSyncProgress, Data: "test"}
	imp.emitEvent(event)

	// Both channels should receive the event
	select {
	case e := <-ch1:
		if e.Type != EventSyncProgress {
			t.Errorf("expected EventSyncProgress, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("ch1 did not receive event")
	}

	select {
	case e := <-ch2:
		if e.Type != EventSyncProgress {
			t.Errorf("expected EventSyncProgress, got %s", e.Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("ch2 did not receive event")
	}
}

func TestEmitEvent_DropsOnFullChannel(t *testing.T) {
	imp := &Importer{}

	// Create a channel with no buffer
	ch := make(chan Event)
	unsub := imp.Subscribe(ch)
	defer unsub()

	// Emit should not block even if channel is full
	done := make(chan bool)
	go func() {
		imp.emitEvent(Event{Type: EventSyncProgress, Data: "test"})
		done <- true
	}()

	select {
	case <-done:
		// Good - emit didn't block
	case <-time.After(100 * time.Millisecond):
		t.Error("emitEvent blocked on full channel")
	}
}

func TestShouldEmitProgress(t *testing.T) {
	imp := &Importer{}

	// Should emit on batch boundary
	if !imp.shouldEmitProgress(shared.ImportEventBatchSize) {
		t.Error("should emit at ImportEventBatchSize")
	}
	if !imp.shouldEmitProgress(shared.ImportEventBatchSize * 2) {
		t.Error("should emit at 2*ImportEventBatchSize")
	}

	// Should not emit mid-batch (when time hasn't elapsed)
	imp.eventMu.Lock()
	imp.lastEventEmitTime = time.Now()
	imp.eventMu.Unlock()

	if imp.shouldEmitProgress(1) {
		t.Error("should not emit at 1 when time hasn't elapsed")
	}

	// Should emit when time threshold exceeded
	flushInterval := time.Duration(shared.ImportEventFlushIntervalMs) * time.Millisecond
	imp.eventMu.Lock()
	imp.lastEventEmitTime = time.Now().Add(-flushInterval - time.Second)
	imp.eventMu.Unlock()

	if !imp.shouldEmitProgress(1) {
		t.Error("should emit when time threshold exceeded")
	}
}
