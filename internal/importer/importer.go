// Package importer handles importing data from Strava.
package importer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/melonamin/quantlete/internal/shared"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// Status represents the current import status.
type Status string

const (
	StatusIdle      Status = "idle"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// Event types for SSE streaming.
type EventType string

const (
	EventSyncProgress EventType = "sync:progress"
	EventSyncComplete EventType = "sync:complete"
	EventDataChanged  EventType = "data:changed"
)

// Timeout constants for async operations.
const (
	achievementDetectionTimeout = 30 * time.Second
)

// Event represents an import event for SSE streaming.
type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

// DataChangeSet represents what data types changed.
type DataChangeSet struct {
	Activities bool `json:"activities,omitempty"`
	Streams    bool `json:"streams,omitempty"`
	Segments   bool `json:"segments,omitempty"`
	Gear       bool `json:"gear,omitempty"`
	Photos     bool `json:"photos,omitempty"`
	All        bool `json:"all,omitempty"`
}

// SyncCompleteData contains data for sync complete events.
type SyncCompleteData struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Progress tracks import progress.
type Progress struct {
	Status      Status    `json:"status"`
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`

	// Current phase
	Phase ImportPhase `json:"phase"`

	// Per-phase progress
	ActivitiesTotal int `json:"activities_total"`
	ActivitiesDone  int `json:"activities_done"`
	GearTotal       int `json:"gear_total"`
	GearDone        int `json:"gear_done"`
	StreamsTotal    int `json:"streams_total"`
	StreamsDone     int `json:"streams_done"`
	DetailsTotal    int `json:"details_total"`
	DetailsDone     int `json:"details_done"`
	SegmentsTotal   int `json:"segments_total"`
	SegmentsDone    int `json:"segments_done"`
	PhotosTotal     int `json:"photos_total"`
	PhotosDone      int `json:"photos_done"`

	// Legacy fields for backward compatibility
	TotalActivities int `json:"total_activities"`
	ImportedCount   int `json:"imported_count"`
	SkippedCount    int `json:"skipped_count"`
	FailedCount     int `json:"failed_count"`
	CurrentPage     int `json:"current_page"`

	// Aggregated non-fatal errors for user visibility
	Errors            []string `json:"errors,omitempty"`
	DroppedErrorCount int      `json:"dropped_error_count,omitempty"`

	// ETA estimation
	RemainingAPICalls int    `json:"remaining_api_calls"`
	EstimatedETA      string `json:"estimated_eta,omitempty"`

	// Rate limit info
	RateLimitUsed15Min  int `json:"rate_limit_used_15min"`
	RateLimitLimit15Min int `json:"rate_limit_limit_15min"`
	RateLimitUsedDaily  int `json:"rate_limit_used_daily"`
	RateLimitLimitDaily int `json:"rate_limit_limit_daily"`

	// Rate limit waiting state
	WaitingForRateLimit bool      `json:"waiting_for_rate_limit"`
	WaitingUntil        time.Time `json:"waiting_until,omitempty"`
	WaitingReason       string    `json:"waiting_reason,omitempty"`
}

// Importer orchestrates data import from Strava.
type Importer struct {
	// Platform-agnostic interfaces for Strava API and storage.
	stravaClient StravaClient
	storage      ImportStorage

	// State management for resume capability.
	// IMPORTANT: state is only accessed from the import goroutine after Start() launches it.
	// Start() ensures single-caller by checking progress.Status == StatusRunning under mu lock.
	// The progress field (protected by mu) is the public-facing snapshot updated via updateProgress().
	// Do NOT access state from other goroutines - use Progress() to read the snapshot instead.
	stateManager *StateManager
	state        *ImportState
	eta          *ETAEstimator

	// Current sync run (for history tracking)
	currentRunID int64

	mu       sync.RWMutex // Protects progress
	progress Progress
	cancel   context.CancelFunc

	// Subscribers for SSE event streaming.
	subscribersMu sync.RWMutex
	subscribers   []chan Event

	// Time-based flush tracking (protected by eventMu).
	eventMu           sync.Mutex
	lastEventEmitTime time.Time

	// Optional notification callback called on sync complete.
	notifyOnComplete func(ctx context.Context, athleteID int64, stats SyncStats)

	// Optional achievement detection support.
	achievementStorage  AchievementStorage
	notificationHistory NotificationHistoryChecker
	achievementDetector *AchievementDetector
}

// SyncStats contains statistics about a completed sync for notifications.
type SyncStats struct {
	ActivitiesImported int
	ActivitiesUpdated  int
	Duration           time.Duration
	Achievements       []Achievement
}

// SetNotificationHandler sets a callback to be invoked when sync completes successfully.
// The callback receives the athlete ID and sync statistics.
func (i *Importer) SetNotificationHandler(fn func(ctx context.Context, athleteID int64, stats SyncStats)) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.notifyOnComplete = fn
}

// SetAchievementStorage sets the achievement storage for detecting achievements during sync.
// If set, achievements will be detected and included in SyncStats on completion.
func (i *Importer) SetAchievementStorage(achievementStorage AchievementStorage) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.achievementStorage = achievementStorage
}

// SetNotificationHistory sets the notification history checker for filtering already-notified achievements.
// If set, achievements that have already been notified will be filtered out.
func (i *Importer) SetNotificationHistory(checker NotificationHistoryChecker) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.notificationHistory = checker
}

// New creates a new importer with platform-agnostic interfaces.
// Returns an error if required dependencies are nil.
func New(
	stravaClient StravaClient,
	importStorage ImportStorage,
) (*Importer, error) {
	if stravaClient == nil {
		return nil, fmt.Errorf("stravaClient is required")
	}
	if importStorage == nil {
		return nil, fmt.Errorf("storage is required")
	}

	return &Importer{
		stravaClient: stravaClient,
		storage:      importStorage,
		stateManager: NewStateManagerFromStorage(importStorage),
		eta:          NewETAEstimator(),
		progress:     Progress{Status: StatusIdle, Phase: PhaseIdle},
	}, nil
}

// ImportOptions configures an import run.
type ImportOptions struct {
	FullSync        bool // If true, re-import all activities (ignore resume state)
	Resume          bool // If true, resume from previous state
	SkipStreams     bool // If true, skip importing stream data (default: import all)
	SkipSegments    bool // If true, skip importing segments and segment efforts (default: import all)
	SkipBestEfforts bool // If true, skip importing Strava best efforts/PRs (default: import all)
	SkipPhotos      bool // If true, skip importing activity photos (default: import all)
}

// Start begins an import operation.
func (i *Importer) Start(ctx context.Context, opts ImportOptions) error {
	i.mu.Lock()
	if i.progress.Status == StatusRunning {
		i.mu.Unlock()
		return fmt.Errorf("import already in progress")
	}

	// Get athlete ID for sync history
	athlete := i.stravaClient.GetAthlete()
	if athlete == nil || athlete.ID == 0 {
		i.mu.Unlock()
		return fmt.Errorf("not authenticated or invalid athlete")
	}
	athleteID := athlete.ID

	// Load or create state
	var state *ImportState
	var err error

	if opts.Resume && !opts.FullSync {
		state, err = i.stateManager.Load(ctx)
		if err != nil {
			i.mu.Unlock()
			return fmt.Errorf("loading import state: %w", err)
		}
		if state.Phase == PhaseCompleted || state.Phase == PhaseIdle {
			// No resume needed, start fresh
			state = nil
		}
	}

	if state == nil {
		state = &ImportState{
			Phase:           PhaseActivities,
			StartedAt:       time.Now(),
			SkipStreams:     opts.SkipStreams,
			SkipSegments:    opts.SkipSegments,
			SkipBestEfforts: opts.SkipBestEfforts,
			SkipPhotos:      opts.SkipPhotos,
		}

		// Load watermark for incremental sync (unless full sync requested)
		if !opts.FullSync {
			var wm *storage.SyncWatermark
			wm, err = i.storage.GetSyncWatermark(ctx, athleteID)
			if err != nil {
				slog.Warn("failed to load sync watermark", "error", err)
			} else if wm != nil && wm.NewestActivityDate != nil {
				state.AfterDate = wm.NewestActivityDate
				slog.Info("using incremental sync from watermark",
					"after_date", wm.NewestActivityDate.Format(time.RFC3339))
			}
		}
	}

	i.state = state

	// Start sync run record
	run, err := i.storage.StartSyncRun(ctx, athleteID, storage.SyncRunOptions{
		FullSync:        opts.FullSync,
		SkipStreams:     opts.SkipStreams,
		SkipSegments:    opts.SkipSegments,
		SkipBestEfforts: opts.SkipBestEfforts,
		SkipPhotos:      opts.SkipPhotos,
	})
	if err != nil {
		slog.Warn("failed to start sync run record", "error", err)
	} else {
		i.currentRunID = run.ID
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	i.cancel = cancel
	i.progress = Progress{
		Status:    StatusRunning,
		StartedAt: state.StartedAt,
		Phase:     state.Phase,
	}

	// Initialize achievement detector if storage is available.
	// Take snapshot before import to enable before/after comparison.
	if i.achievementStorage != nil {
		i.achievementDetector = NewAchievementDetector(slog.Default(), i.achievementStorage, athleteID)
		if i.notificationHistory != nil {
			i.achievementDetector.SetNotificationHistory(i.notificationHistory)
		}
		if snapErr := i.achievementDetector.TakeSnapshot(ctx); snapErr != nil {
			slog.Debug("failed to take achievement snapshot", "error", snapErr)
		}
	}
	i.mu.Unlock()

	// Reset event emit time under its own mutex
	i.eventMu.Lock()
	i.lastEventEmitTime = time.Now()
	i.eventMu.Unlock()

	// Run import in background
	go func() { //nolint:gosec // import runs beyond the request lifetime by design; ctx comes from the caller (see Start contract)
		defer func() {
			if r := recover(); r != nil {
				slog.Error("import panicked", "panic", r)
				i.mu.Lock()
				i.progress.Status = StatusFailed
				i.progress.Error = fmt.Sprintf("panic: %v", r)
				i.progress.CompletedAt = time.Now()
				i.mu.Unlock()
			}
		}()

		err := i.runImport(ctx, opts)
		i.mu.Lock()

		i.progress.CompletedAt = time.Now()

		// Build sync run counts
		counts := storage.SyncRunCounts{
			ActivitiesTotal:    i.state.ActivitiesTotal,
			ActivitiesImported: i.state.ActivitiesDone,
			ActivitiesSkipped:  i.state.ActivitiesTotal - i.state.ActivitiesDone,
			GearImported:       i.state.GearDone,
			StreamsImported:    i.state.StreamsDone,
			SegmentsImported:   i.state.SegmentsDone,
			PhotosImported:     i.state.PhotosDone,
			FailedCount:        i.state.FailedCount,
			NewestActivityDate: i.state.NewestActivityDate,
		}

		if err != nil {
			slog.Error("import failed", "error", err)
			if ctx.Err() == context.Canceled {
				i.progress.Status = StatusCanceled
				// Emit canceled event
				i.emitSyncComplete("canceled", "")
				// Log canceled run
				if i.currentRunID > 0 {
					if logErr := i.storage.CancelSyncRun(context.Background(), i.currentRunID, counts); logErr != nil {
						slog.Warn("failed to log canceled sync run", "error", logErr)
					}
				}
			} else {
				i.progress.Status = StatusFailed
				i.progress.Error = err.Error()
				// Emit failed event
				i.emitSyncComplete("failed", err.Error())
				// Log failed run
				if i.currentRunID > 0 {
					if logErr := i.storage.FailSyncRun(context.Background(), i.currentRunID, err.Error(), counts); logErr != nil {
						slog.Warn("failed to log failed sync run", "error", logErr)
					}
				}
			}
			i.mu.Unlock()
			return
		}

		// Success path
		i.progress.Status = StatusCompleted
		i.progress.Phase = PhaseCompleted

		// Emit completed event and data changed
		i.emitSyncComplete("completed", "")
		i.emitDataChanged(DataChangeSet{All: true})

		// Log completed run
		if i.currentRunID > 0 {
			if logErr := i.storage.CompleteSyncRun(context.Background(), i.currentRunID, counts); logErr != nil {
				slog.Warn("failed to log completed sync run", "error", logErr)
			}

			// Update watermark with newest activity date
			if i.state.NewestActivityDate != nil {
				wm := &storage.SyncWatermark{
					LastSyncedAt:       time.Now(),
					NewestActivityDate: i.state.NewestActivityDate,
				}
				if wmErr := i.storage.SetSyncWatermark(context.Background(), athleteID, wm); wmErr != nil {
					slog.Warn("failed to update sync watermark", "error", wmErr)
				} else {
					slog.Info("updated sync watermark",
						"newest_activity_date", i.state.NewestActivityDate.Format(time.RFC3339))
				}
			}
		}

		// Copy references under lock for use after unlock.
		// This pattern ensures thread-safety: we capture all needed values while
		// holding the lock, then release it before any slow I/O operations
		// (achievement detection, notifications). This prevents blocking Progress()
		// calls during potentially slow network operations.
		detector := i.achievementDetector
		activityIDs := append([]int64(nil), i.state.ActivityIDs...) // Deep copy to avoid data race
		notifyFn := i.notifyOnComplete
		startedAt := i.progress.StartedAt
		i.mu.Unlock()

		// Achievement detection and notifications run outside the lock
		// to avoid blocking other operations during potentially slow calls.

		var achievements []Achievement
		if detector != nil && len(activityIDs) > 0 {
			// Use a separate context since the import context may already be canceled.
			detectCtx, detectCancel := context.WithTimeout(context.Background(), achievementDetectionTimeout)
			var detectErr error
			achievements, detectErr = detector.DetectAchievements(detectCtx, activityIDs)
			detectCancel()
			if detectErr != nil {
				slog.Warn("failed to detect achievements", "error", detectErr)
			}
		}

		if notifyFn != nil {
			stats := SyncStats{
				ActivitiesImported: counts.ActivitiesImported,
				ActivitiesUpdated:  counts.ActivitiesTotal - counts.ActivitiesImported - counts.ActivitiesSkipped,
				Duration:           time.Since(startedAt),
				Achievements:       achievements,
			}
			// Run notification in goroutine to not block completion.
			// The handler manages its own timeout via notificationHandlerTimeout.
			go func() {
				notifyFn(context.Background(), athleteID, stats)
			}()
		}

		// Clear state on successful completion
		if clearErr := i.stateManager.Clear(context.Background()); clearErr != nil {
			slog.Warn("failed to clear import state", "error", clearErr)
		}
	}()

	return nil
}

// Cancel cancels a running import.
func (i *Importer) Cancel() {
	i.mu.RLock()
	cancel := i.cancel
	i.mu.RUnlock()

	if cancel != nil {
		cancel()
	}
}

// Pause pauses a running import. The import can be resumed later with Start(opts{Resume: true}).
// This is currently implemented as a cancel that preserves state, allowing resume.
func (i *Importer) Pause() bool {
	i.mu.RLock()
	status := i.progress.Status
	cancel := i.cancel
	i.mu.RUnlock()

	if status != StatusRunning || cancel == nil {
		return false
	}

	// Save current state before canceling so it can be resumed
	// The saveState is called periodically during import, so state is already preserved
	cancel()
	return true
}

// CanResume returns true if there is a resumable import state.
func (i *Importer) CanResume(ctx context.Context) bool {
	state, err := i.stateManager.Load(ctx)
	if err != nil {
		return false
	}
	return state != nil && state.Phase != PhaseCompleted && state.Phase != PhaseIdle
}

// Progress returns the current import progress.
func (i *Importer) Progress() Progress {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.progress
}

// Subscribe registers a channel to receive import events.
// Returns an unsubscribe function that closes the channel and removes it.
// The unsubscribe function is safe to call multiple times.
// Note: Only the importer (via unsubscribe) closes subscription channels;
// external callers must not close them directly.
func (i *Importer) Subscribe(ch chan Event) func() {
	i.subscribersMu.Lock()
	i.subscribers = append(i.subscribers, ch)
	i.subscribersMu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			i.subscribersMu.Lock()
			defer i.subscribersMu.Unlock()
			for idx, sub := range i.subscribers {
				if sub == ch {
					i.subscribers = append(i.subscribers[:idx], i.subscribers[idx+1:]...)
					close(ch)
					return
				}
			}
			// Channel not found in subscribers list - already removed by a previous call.
			// Do NOT close here to avoid double-close panic. The sync.Once ensures
			// this branch only executes if unsubscribe was somehow called after the
			// channel was already removed (shouldn't happen with proper usage).
		})
	}
}

// emitEvent sends an event to all subscribers (non-blocking, best-effort delivery).
// Events are dropped without blocking if a subscriber's channel is full.
// This prevents slow subscribers from blocking the import process.
func (i *Importer) emitEvent(event Event) {
	i.subscribersMu.RLock()
	defer i.subscribersMu.RUnlock()

	for _, ch := range i.subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, drop event (non-blocking to avoid backpressure)
			slog.Debug("SSE event dropped: subscriber channel full", "type", event.Type)
		}
	}
}

// emitProgress sends a sync:progress event with current progress.
func (i *Importer) emitProgress() {
	progress := i.Progress()
	i.emitEvent(Event{
		Type: EventSyncProgress,
		Data: progress,
	})
	i.eventMu.Lock()
	i.lastEventEmitTime = time.Now()
	i.eventMu.Unlock()
}

// shouldEmitProgress checks if we should emit a progress event based on batch count or time elapsed.
// Returns true if either:
// - itemsDone is a multiple of shared.ImportEventBatchSize (batch threshold)
// - More than shared.ImportEventFlushIntervalMs has elapsed since last emit (time threshold)
func (i *Importer) shouldEmitProgress(itemsDone int) bool {
	if itemsDone%shared.ImportEventBatchSize == 0 {
		return true
	}
	i.eventMu.Lock()
	elapsed := time.Since(i.lastEventEmitTime)
	i.eventMu.Unlock()
	return elapsed >= time.Duration(shared.ImportEventFlushIntervalMs)*time.Millisecond
}

// emitDataChanged sends a data:changed event.
func (i *Importer) emitDataChanged(changes DataChangeSet) {
	i.emitEvent(Event{
		Type: EventDataChanged,
		Data: changes,
	})
}

// emitSyncComplete sends a sync:complete event.
func (i *Importer) emitSyncComplete(status, errMsg string) {
	i.emitEvent(Event{
		Type: EventSyncComplete,
		Data: SyncCompleteData{
			Status: status,
			Error:  errMsg,
		},
	})
}

// runImport performs the actual import using a phased approach.
func (i *Importer) runImport(ctx context.Context, opts ImportOptions) error {
	slog.Info("starting phased import",
		"full_sync", opts.FullSync,
		"resume", opts.Resume,
		"skip_streams", opts.SkipStreams,
		"skip_segments", opts.SkipSegments,
		"skip_best_efforts", opts.SkipBestEfforts,
		"skip_photos", opts.SkipPhotos,
		"resuming_from_phase", i.state.Phase,
	)

	athlete := i.stravaClient.GetAthlete()
	if athlete == nil {
		slog.Error("import failed: not authenticated")
		return fmt.Errorf("not authenticated")
	}
	slog.Info("import authenticated", "athlete_id", athlete.ID)

	// Save athlete profile via storage interface
	if err := i.storage.SaveAthlete(ctx, athlete); err != nil {
		slog.Warn("failed to save athlete profile", "error", err)
	}

	// Phase 1: Activity list (enables dashboard immediately)
	if i.state.Phase == PhaseActivities {
		slog.Info("Phase 1: Importing activity list")
		if err := i.runActivitiesPhase(ctx, athlete.ID); err != nil {
			return err
		}
		i.state.Phase = PhaseGear
		i.saveState(ctx)
		i.emitDataChanged(DataChangeSet{Activities: true})
		i.emitProgress()
	}

	// Phase 2: Gear (quick, needed for gear stats)
	if i.state.Phase == PhaseGear {
		slog.Info("Phase 2: Importing gear")
		if err := i.runGearPhase(ctx, athlete.ID); err != nil {
			return err
		}
		i.state.Phase = PhaseStreams
		i.saveState(ctx)
		i.emitDataChanged(DataChangeSet{Gear: true})
		i.emitProgress()
	}

	// Phase 3: Streams (enables training load, power analysis)
	if i.state.Phase == PhaseStreams {
		if !i.state.SkipStreams {
			slog.Info("Phase 3: Importing streams")
			if err := i.runStreamsPhase(ctx); err != nil {
				return err
			}
		} else {
			slog.Info("Phase 3: Skipping streams (disabled)")
		}
		i.state.Phase = PhaseActivityDetails
		i.saveState(ctx)
		i.emitDataChanged(DataChangeSet{Streams: true})
		i.emitProgress()
	}

	// Phase 4: Activity Details (best efforts + segment efforts)
	if i.state.Phase == PhaseActivityDetails {
		if !i.state.SkipBestEfforts || !i.state.SkipSegments {
			slog.Info("Phase 4: Importing activity details (best efforts + segments)")
			if err := i.runActivityDetailsPhase(ctx, athlete.ID); err != nil {
				return err
			}
		} else {
			slog.Info("Phase 4: Skipping activity details (best efforts and segments disabled)")
		}
		i.state.Phase = PhaseSegmentDetails
		i.saveState(ctx)
		i.emitDataChanged(DataChangeSet{Segments: true})
		i.emitProgress()
	}

	// Phase 5: Segment Details (for segments missing full data)
	if i.state.Phase == PhaseSegmentDetails {
		if !i.state.SkipSegments && len(i.state.SegmentIDsToFetch) > 0 {
			slog.Info("Phase 5: Importing segment details", "count", len(i.state.SegmentIDsToFetch))
			if err := i.runSegmentDetailsPhase(ctx); err != nil {
				return err
			}
		} else {
			slog.Info("Phase 5: Skipping segment details")
		}
		i.state.Phase = PhasePhotos
		i.saveState(ctx)
		i.emitDataChanged(DataChangeSet{Segments: true})
		i.emitProgress()
	}

	// Phase 6: Photos (cosmetic, last)
	if i.state.Phase == PhasePhotos {
		if !i.state.SkipPhotos {
			slog.Info("Phase 6: Importing photos")
			if err := i.runPhotosPhase(ctx, athlete.ID); err != nil {
				return err
			}
		} else {
			slog.Info("Phase 6: Skipping photos (disabled)")
		}
		i.state.Phase = PhaseCompleted
		i.saveState(ctx)
		i.emitDataChanged(DataChangeSet{Photos: true})
		i.emitProgress()
	}

	slog.Info("import completed",
		"activities", i.state.ActivitiesDone,
		"gear", i.state.GearDone,
		"streams", i.state.StreamsDone,
		"details", i.state.DetailsDone,
		"segments", i.state.SegmentsDone,
		"photos", i.state.PhotosDone,
		"failed", i.state.FailedCount,
	)

	return nil
}

// saveState persists the current import state.
// Failures are logged and tracked in aggregated errors so users know resume may not work.
func (i *Importer) saveState(ctx context.Context) {
	if err := i.stateManager.Save(ctx, i.state); err != nil {
		slog.Warn("failed to save import state", "error", err)
		i.addError("state persistence failed (resume may not work): %v", err)
	}
}

// updateProgress updates the progress from current state.
// Called from the import goroutine to snapshot state into the progress field
// which can be safely read by other goroutines via Progress().
//
// Thread safety: i.state is only accessed from this goroutine (see struct comment),
// so reading i.state.Errors here is safe. The copy is made before taking the lock,
// then assigned to i.progress.Errors under lock. Progress() reads under the same lock.
func (i *Importer) updateProgress() {
	// Copy errors slice to avoid sharing the backing array
	var errors []string
	if len(i.state.Errors) > 0 {
		errors = make([]string, len(i.state.Errors))
		copy(errors, i.state.Errors)
	}

	remaining := i.state.RemainingAPICalls()

	i.mu.Lock()
	defer i.mu.Unlock()

	i.progress.Phase = i.state.Phase
	i.progress.ActivitiesTotal = i.state.ActivitiesTotal
	i.progress.ActivitiesDone = i.state.ActivitiesDone
	i.progress.GearTotal = i.state.GearTotal
	i.progress.GearDone = i.state.GearDone
	i.progress.StreamsTotal = i.state.StreamsTotal
	i.progress.StreamsDone = i.state.StreamsDone
	i.progress.DetailsTotal = i.state.DetailsTotal
	i.progress.DetailsDone = i.state.DetailsDone
	i.progress.SegmentsTotal = i.state.SegmentsTotal
	i.progress.SegmentsDone = i.state.SegmentsDone
	i.progress.PhotosTotal = i.state.PhotosTotal
	i.progress.PhotosDone = i.state.PhotosDone
	i.progress.FailedCount = i.state.FailedCount

	// Legacy fields
	i.progress.TotalActivities = i.state.ActivitiesTotal
	i.progress.ImportedCount = i.state.ActivitiesDone

	// ETA estimation
	i.progress.RemainingAPICalls = remaining
	i.progress.EstimatedETA = FormatETA(i.eta.EstimateCompletion(remaining))

	// Copy aggregated errors
	i.progress.Errors = errors
	i.progress.DroppedErrorCount = i.state.DroppedErrorCount
}

// maxAggregatedErrors limits the number of errors stored to prevent unbounded memory growth.
const maxAggregatedErrors = 100

// addError records a non-fatal error for user visibility.
// These errors are aggregated and displayed to the user at the end of the sync.
// Called only from the import goroutine, so no locking is needed.
func (i *Importer) addError(format string, args ...any) {
	i.state.FailedCount++
	// Cap error list to prevent unbounded memory growth
	if len(i.state.Errors) >= maxAggregatedErrors {
		i.state.DroppedErrorCount++
		return
	}
	msg := fmt.Sprintf(format, args...)
	i.state.Errors = append(i.state.Errors, msg)
}

// waitForRateLimit handles rate limit errors by waiting and retrying.
// It updates progress to show the waiting state so the UI can display a countdown.
// Uses a ticker to allow for periodic cancellation checks instead of blocking.
func (i *Importer) waitForRateLimit(ctx context.Context, err error) error {
	rle, ok := strava.IsRateLimitError(err)
	if !ok {
		return err
	}

	slog.Info("rate limited by Strava, waiting for reset",
		"wait_duration", rle.RetryAfter.Round(time.Second),
		"reset_at", rle.ResetAt)

	// Update progress to show waiting state
	i.mu.Lock()
	i.progress.WaitingForRateLimit = true
	i.progress.WaitingUntil = rle.ResetAt
	i.progress.WaitingReason = fmt.Sprintf("Rate limit exceeded. Waiting %v for reset.", rle.RetryAfter.Round(time.Second))
	i.mu.Unlock()
	i.emitProgress() // Notify UI of rate limit wait

	// Use ticker for periodic cancellation checks instead of blocking time.After
	deadline := time.Now().Add(rle.RetryAfter)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			i.clearWaitingState()
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				i.clearWaitingState()
				i.emitProgress() // Notify UI that wait is over
				return nil
			}
			// Update waiting progress periodically
			remaining := time.Until(deadline).Round(time.Second)
			i.mu.Lock()
			i.progress.WaitingReason = fmt.Sprintf("Rate limit exceeded. Waiting %v for reset.", remaining)
			i.mu.Unlock()
			i.emitProgress() // Send countdown update to UI
		}
	}
}

// clearWaitingState clears the rate limit waiting state from progress.
func (i *Importer) clearWaitingState() {
	i.mu.Lock()
	i.progress.WaitingForRateLimit = false
	i.progress.WaitingUntil = time.Time{}
	i.progress.WaitingReason = ""
	i.mu.Unlock()
}

// withRetry wraps an API operation with rate limit retry logic.
func (i *Importer) withRetry(ctx context.Context, op func() error) error {
	for {
		err := op()
		if err == nil {
			return nil
		}

		// Check if it's a rate limit error
		if _, ok := strava.IsRateLimitError(err); ok {
			if waitErr := i.waitForRateLimit(ctx, err); waitErr != nil {
				return waitErr
			}
			// Retry the operation after waiting
			continue
		}

		// Non-rate-limit error, return it
		return err
	}
}

// runActivitiesPhase imports all activity metadata.
// Activities are fetched from Strava newest-first (descending by start_date),
// so users see their fresh data first during sync.
func (i *Importer) runActivitiesPhase(ctx context.Context, athleteID int64) error {
	page := i.state.ActivitiesLastPage
	if page == 0 {
		page = 1
	}
	perPage := strava.MaxActivitiesPerPage
	seenIDs := make(map[int64]bool)

	// Mark seen IDs from existing state
	for _, id := range i.state.ActivityIDs {
		seenIDs[id] = true
	}

	// Log watermark usage
	if i.state.AfterDate != nil {
		slog.Info("incremental sync: fetching activities after watermark",
			"after_date", i.state.AfterDate.Format(time.RFC3339))
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		i.state.ActivitiesLastPage = page
		i.mu.Lock()
		i.progress.CurrentPage = page
		i.mu.Unlock()

		slog.Debug("fetching activities page", "page", page, "after_date", i.state.AfterDate)

		var activities []Activity
		err := i.withRetry(ctx, func() error {
			var fetchErr error
			activities, fetchErr = i.stravaClient.GetActivities(ctx, GetActivitiesOptions{
				Page:    page,
				PerPage: perPage,
				After:   i.state.AfterDate,
			})
			return fetchErr
		})
		if err != nil {
			return fmt.Errorf("fetching activities page %d: %w", page, err)
		}

		if len(activities) == 0 {
			break
		}

		for idx := range activities {
			a := &activities[idx]
			if seenIDs[a.ID] {
				continue
			}
			seenIDs[a.ID] = true
			i.state.ActivityIDs = append(i.state.ActivityIDs, a.ID)
			i.state.ActivitiesTotal++

			// Hashtag-based custom gear linking
			if a.GearID == "" {
				if gearID, err := i.storage.ResolveCustomGearID(ctx, athleteID, a.Name); err == nil && gearID != "" {
					a.GearID = gearID
				}
			}

			// Store activity using platform-agnostic storage interface
			if err := i.storage.SaveActivity(ctx, athleteID, a); err != nil {
				slog.Warn("failed to store activity", "id", a.ID, "error", err)
				i.addError("failed to store activity %d: %v", a.ID, err)
				continue
			}

			i.state.ActivitiesDone++

			// Emit progress event periodically (batch or time-based)
			if i.shouldEmitProgress(i.state.ActivitiesDone) {
				i.updateProgress()
				i.emitProgress()
			}

			// Track newest activity date for watermark
			// Truncate to second precision to match Strava API (Unix timestamp)
			if i.state.NewestActivityDate == nil || a.StartDate.After(*i.state.NewestActivityDate) {
				t := a.StartDate.Truncate(time.Second)
				i.state.NewestActivityDate = &t
			}

			// Collect gear IDs
			if a.GearID != "" && !contains(i.state.GearIDs, a.GearID) {
				i.state.GearIDs = append(i.state.GearIDs, a.GearID)
			}

			// Log maintenance from hashtags
			if err := i.storage.LogMaintenanceFromHashtags(ctx, athleteID, a.ID, a.StartDateLocal, a.Name); err != nil {
				slog.Debug("failed to log maintenance from hashtags", "activity_id", a.ID, "error", err)
			}
		}

		i.state.GearTotal = len(i.state.GearIDs)
		i.state.StreamsTotal = len(i.state.ActivityIDs)
		i.state.DetailsTotal = len(i.state.ActivityIDs)
		i.state.PhotosTotal = len(i.state.ActivityIDs)

		i.updateProgress()
		i.saveState(ctx)

		if len(activities) < perPage {
			break
		}
		page++
	}

	slog.Info("activities phase complete", "total", i.state.ActivitiesTotal, "imported", i.state.ActivitiesDone)
	return nil
}

// runGearPhase imports all gear.
//
//nolint:dupl // Similar loop pattern to runPhotosPhase but different state fields
func (i *Importer) runGearPhase(ctx context.Context, athleteID int64) error {
	for idx := i.state.GearLastIndex; idx < len(i.state.GearIDs); idx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		gearID := i.state.GearIDs[idx]
		if err := i.importGearItem(ctx, athleteID, gearID); err != nil {
			slog.Warn("failed to import gear", "id", gearID, "error", err)
			i.addError("failed to import gear %s: %v", gearID, err)
		} else {
			i.state.GearDone++
		}

		i.state.GearLastIndex = idx + 1
		i.updateProgress()

		// Save state periodically (every 10 items)
		if idx%10 == 0 {
			i.saveState(ctx)
		}
	}

	i.saveState(ctx)
	slog.Info("gear phase complete", "total", i.state.GearTotal, "imported", i.state.GearDone)
	return nil
}

// runStreamsPhase imports streams for all activities.
func (i *Importer) runStreamsPhase(ctx context.Context) error {
	for idx := i.state.StreamsLastIndex; idx < len(i.state.ActivityIDs); idx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		activityID := i.state.ActivityIDs[idx]
		if err := i.importActivityStreams(ctx, activityID); err != nil {
			slog.Debug("failed to import streams", "activity_id", activityID, "error", err)
			i.addError("failed to import streams for activity %d: %v", activityID, err)
		} else {
			i.state.StreamsDone++
		}

		i.state.StreamsLastIndex = idx + 1
		i.updateProgress()

		// Emit progress event periodically (batch or time-based)
		if i.shouldEmitProgress(i.state.StreamsDone) {
			i.emitProgress()
			// Also emit data changed for incremental updates
			i.emitDataChanged(DataChangeSet{Streams: true})
		}

		// Save state periodically
		if idx%50 == 0 {
			i.saveState(ctx)
		}
	}

	i.saveState(ctx)
	slog.Info("streams phase complete", "total", i.state.StreamsTotal, "imported", i.state.StreamsDone)
	return nil
}

// runActivityDetailsPhase fetches activity details for best efforts and segments.
func (i *Importer) runActivityDetailsPhase(ctx context.Context, athleteID int64) error {
	for idx := i.state.DetailsLastIndex; idx < len(i.state.ActivityIDs); idx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		activityID := i.state.ActivityIDs[idx]

		// Fetch activity detail with retry for rate limits
		var detail *Activity
		err := i.withRetry(ctx, func() error {
			var fetchErr error
			detail, fetchErr = i.stravaClient.GetActivity(ctx, activityID)
			return fetchErr
		})
		if err != nil {
			// 404/403 are "soft" errors - activity may have been deleted or made private
			if isResourceGoneError(err) {
				slog.Debug("activity no longer accessible (deleted or private)", "activity_id", activityID)
				i.state.DetailsDone++ // Count as processed (skipped)
			} else {
				slog.Debug("failed to fetch activity detail", "activity_id", activityID, "error", err)
				i.addError("failed to fetch activity detail %d: %v", activityID, err)
			}
			i.state.DetailsLastIndex = idx + 1
			continue
		}

		// Import best efforts
		if !i.state.SkipBestEfforts {
			if err := i.storage.SaveBestEfforts(ctx, athleteID, activityID, detail.SportType, detail.BestEfforts); err != nil {
				slog.Debug("failed to import best efforts", "activity_id", activityID, "error", err)
				i.addError("failed to import best efforts for activity %d: %v", activityID, err)
			}
		}

		// Import segment efforts and collect segment IDs
		if !i.state.SkipSegments {
			for effortIdx := range detail.SegmentEfforts {
				effort := &detail.SegmentEfforts[effortIdx]
				seg := &effort.Segment

				// Skip segments with invalid zero ID
				if seg.ID == 0 {
					slog.Debug("skipping segment with zero ID", "activity_id", activityID)
					continue
				}

				// Store segment and effort atomically
				if err := i.storage.SaveSegmentWithEffort(ctx, athleteID, activityID, effort, detail.LocationCountry); err != nil {
					slog.Debug("failed to store segment with effort", "segment_id", seg.ID, "effort_id", effort.ID, "error", err)
					i.addError("failed to store segment %d with effort %d: %v", seg.ID, effort.ID, err)
				}

				// Check if we need to fetch full segment detail later
				if shouldFetchSegmentDetail(seg) {
					if !containsInt64(i.state.SegmentIDsToFetch, seg.ID) {
						i.state.SegmentIDsToFetch = append(i.state.SegmentIDsToFetch, seg.ID)
					}
				}
			}
		}

		i.state.DetailsDone++
		i.state.DetailsLastIndex = idx + 1
		i.state.SegmentsTotal = len(i.state.SegmentIDsToFetch)
		i.updateProgress()

		// Save state periodically
		if idx%50 == 0 {
			i.saveState(ctx)
		}
	}

	// Deduplicate segment IDs
	i.state.SegmentIDsToFetch = uniqueInt64(i.state.SegmentIDsToFetch)
	i.state.SegmentsTotal = len(i.state.SegmentIDsToFetch)

	i.saveState(ctx)
	slog.Info("activity details phase complete", "total", i.state.DetailsTotal, "imported", i.state.DetailsDone, "segments_to_fetch", len(i.state.SegmentIDsToFetch))
	return nil
}

// runSegmentDetailsPhase fetches full segment details.
func (i *Importer) runSegmentDetailsPhase(ctx context.Context) error {
	for idx := i.state.SegmentsLastIndex; idx < len(i.state.SegmentIDsToFetch); idx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		segmentID := i.state.SegmentIDsToFetch[idx]

		var seg *Segment
		fetchErr := i.withRetry(ctx, func() error {
			var err error
			seg, err = i.stravaClient.GetSegment(ctx, segmentID)
			return err
		})
		if fetchErr != nil {
			// 404/403 are "soft" errors - segment may have been deleted or made private
			if isResourceGoneError(fetchErr) {
				slog.Debug("segment no longer accessible (deleted or private)", "segment_id", segmentID)
				i.state.SegmentsDone++ // Count as processed (skipped)
			} else {
				slog.Debug("failed to fetch segment detail", "segment_id", segmentID, "error", fetchErr)
				i.addError("failed to fetch segment detail %d: %v", segmentID, fetchErr)
			}
		} else {
			if err := i.storage.SaveSegment(ctx, seg); err != nil {
				slog.Debug("failed to store segment", "segment_id", segmentID, "error", err)
				i.addError("failed to store segment %d: %v", segmentID, err)
			} else {
				i.state.SegmentsDone++
			}
		}

		i.state.SegmentsLastIndex = idx + 1
		i.updateProgress()

		// Save state periodically
		if idx%20 == 0 {
			i.saveState(ctx)
		}
	}

	i.saveState(ctx)
	slog.Info("segment details phase complete", "total", i.state.SegmentsTotal, "imported", i.state.SegmentsDone)
	return nil
}

// runPhotosPhase imports photos for all activities.
//
//nolint:dupl // Similar loop pattern to runGearPhase but different state fields
func (i *Importer) runPhotosPhase(ctx context.Context, athleteID int64) error {
	for idx := i.state.PhotosLastIndex; idx < len(i.state.ActivityIDs); idx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		activityID := i.state.ActivityIDs[idx]
		if err := i.importActivityPhotos(ctx, activityID, athleteID); err != nil {
			slog.Debug("failed to import photos", "activity_id", activityID, "error", err)
			i.addError("failed to import photos for activity %d: %v", activityID, err)
		} else {
			i.state.PhotosDone++
		}

		i.state.PhotosLastIndex = idx + 1
		i.updateProgress()

		// Save state periodically
		if idx%50 == 0 {
			i.saveState(ctx)
		}
	}

	i.saveState(ctx)
	slog.Info("photos phase complete", "total", i.state.PhotosTotal, "imported", i.state.PhotosDone)
	return nil
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsInt64(slice []int64, item int64) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func uniqueInt64(slice []int64) []int64 {
	seen := make(map[int64]bool)
	result := make([]int64, 0, len(slice))
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// ImportActivityByID fetches and imports a single activity by ID.
// Used for webhook-triggered updates to ensure streams, segments, best efforts, and photos are synced.
func (i *Importer) ImportActivityByID(ctx context.Context, activityID int64, opts ImportOptions) error {
	athlete := i.stravaClient.GetAthlete()
	if athlete == nil || athlete.ID == 0 {
		return fmt.Errorf("not authenticated")
	}

	// Fetch the activity from Strava
	activity, err := i.stravaClient.GetActivity(ctx, activityID)
	if err != nil {
		return fmt.Errorf("fetching activity %d: %w", activityID, err)
	}

	return i.importSingleActivity(ctx, activity, athlete.ID, opts)
}

// importSingleActivity imports a single activity (used for webhook/incremental sync).
func (i *Importer) importSingleActivity(ctx context.Context, a *Activity, athleteID int64, opts ImportOptions) error {
	actSource := a
	var detail *Activity

	// Segment/best-efforts are only available on the detailed activity response.
	needsDetail := (!opts.SkipSegments) || (!opts.SkipBestEfforts)
	if needsDetail {
		var err error
		detail, err = i.stravaClient.GetActivity(ctx, a.ID)
		if err != nil {
			slog.Debug("failed to fetch activity detail", "activity_id", a.ID, "error", err)
		} else {
			actSource = detail
		}
	}

	// Hashtag-based custom gear linking (only if Strava gear_id is empty).
	if actSource.GearID == "" {
		if gearID, err := i.storage.ResolveCustomGearID(ctx, athleteID, actSource.Name); err == nil && gearID != "" {
			actSource.GearID = gearID
		}
	}

	if err := i.storage.SaveActivity(ctx, athleteID, actSource); err != nil {
		return fmt.Errorf("storing activity: %w", err)
	}

	// Import streams unless skipped
	if !opts.SkipStreams {
		if err := i.importActivityStreams(ctx, a.ID); err != nil {
			slog.Debug("failed to import streams", "activity_id", a.ID, "error", err)
		}
	}

	// Import segments and best efforts from detail
	source := detail
	if source == nil {
		source = actSource
	}

	if !opts.SkipSegments {
		var segmentsNeedingDetail []int64

		for idx := range source.SegmentEfforts {
			effort := &source.SegmentEfforts[idx]
			seg := &effort.Segment
			if seg.ID == 0 {
				continue
			}
			if err := i.storage.SaveSegmentWithEffort(ctx, athleteID, a.ID, effort, source.LocationCountry); err != nil {
				slog.Debug("failed to store segment with effort", "segment_id", seg.ID, "effort_id", effort.ID, "error", err)
			}
			// Check if we need to fetch full segment detail (polyline, athlete stats)
			if shouldFetchSegmentDetail(seg) {
				segmentsNeedingDetail = append(segmentsNeedingDetail, seg.ID)
			}
		}

		// Fetch full segment details for segments missing data
		for _, segmentID := range segmentsNeedingDetail {
			var seg *Segment
			fetchErr := i.withRetry(ctx, func() error {
				var err error
				seg, err = i.stravaClient.GetSegment(ctx, segmentID)
				return err
			})
			if fetchErr != nil {
				if isResourceGoneError(fetchErr) {
					slog.Debug("segment no longer accessible", "segment_id", segmentID)
				} else {
					slog.Debug("failed to fetch segment detail", "segment_id", segmentID, "error", fetchErr)
				}
				continue
			}
			if err := i.storage.SaveSegment(ctx, seg); err != nil {
				slog.Debug("failed to store segment detail", "segment_id", segmentID, "error", err)
			}
		}
	}

	if !opts.SkipBestEfforts {
		if err := i.storage.SaveBestEfforts(ctx, athleteID, a.ID, source.SportType, source.BestEfforts); err != nil {
			slog.Debug("failed to import best efforts", "activity_id", a.ID, "error", err)
		}
	}

	if !opts.SkipPhotos {
		if err := i.importActivityPhotos(ctx, a.ID, athleteID); err != nil {
			slog.Debug("failed to import photos", "activity_id", a.ID, "error", err)
		}
	}

	// Hashtag-based maintenance logging.
	if err := i.storage.LogMaintenanceFromHashtags(ctx, athleteID, a.ID, actSource.StartDateLocal, actSource.Name); err != nil {
		slog.Debug("failed to log maintenance from hashtags", "activity_id", a.ID, "error", err)
	}

	return nil
}

// importActivityStreams imports streams for an activity using the interface.
// Returns an error if fetching fails or if any stream storage operation fails.
func (i *Importer) importActivityStreams(ctx context.Context, activityID int64) error {
	var streams *StreamSet
	err := i.withRetry(ctx, func() error {
		var fetchErr error
		streams, fetchErr = i.stravaClient.GetActivityStreams(ctx, activityID, nil)
		return fetchErr
	})
	if err != nil {
		return err
	}
	if streams == nil {
		return nil
	}

	// Store each stream type
	streamTypes := map[string]*Stream{
		"time":            streams.Time,
		"distance":        streams.Distance,
		"altitude":        streams.Altitude,
		"heartrate":       streams.Heartrate,
		"watts":           streams.Watts,
		"cadence":         streams.Cadence,
		"velocity_smooth": streams.VelocitySmooth,
		"latlng":          streams.Latlng,
	}

	var storageErrors []string
	for streamType, s := range streamTypes {
		if s != nil {
			if err := i.storage.SaveStream(ctx, activityID, streamType, s); err != nil {
				slog.Debug("failed to save stream", "activity_id", activityID, "stream_type", streamType, "error", err)
				storageErrors = append(storageErrors, fmt.Sprintf("%s: %v", streamType, err))
			}
		}
	}

	if len(storageErrors) > 0 {
		return fmt.Errorf("failed to save %d stream(s): %s", len(storageErrors), strings.Join(storageErrors, "; "))
	}

	return nil
}

// importActivityPhotos imports photos for an activity using the interface.
// Returns an error if fetching fails or if any photo storage operation fails.
func (i *Importer) importActivityPhotos(ctx context.Context, activityID, athleteID int64) error {
	var photos []Photo
	err := i.withRetry(ctx, func() error {
		var fetchErr error
		photos, fetchErr = i.stravaClient.GetActivityPhotos(ctx, activityID)
		return fetchErr
	})
	if err != nil {
		return err
	}

	var failedCount int
	for idx := range photos {
		p := &photos[idx]
		if err := i.storage.SavePhoto(ctx, athleteID, activityID, p); err != nil {
			slog.Debug("failed to save photo", "activity_id", activityID, "photo_id", p.UniqueID, "error", err)
			failedCount++
		}
	}

	if failedCount > 0 {
		return fmt.Errorf("failed to save %d of %d photo(s)", failedCount, len(photos))
	}

	return nil
}

// importGearItem imports gear details using the interface.
func (i *Importer) importGearItem(ctx context.Context, athleteID int64, gearID string) error {
	var gear *Gear
	err := i.withRetry(ctx, func() error {
		var fetchErr error
		gear, fetchErr = i.stravaClient.GetGear(ctx, gearID)
		return fetchErr
	})
	if err != nil {
		return err
	}

	return i.storage.SaveGear(ctx, athleteID, gear)
}

// isResourceGoneError returns true if the error indicates the resource no longer exists
// or is inaccessible (404 Not Found or 403 Forbidden). These are "soft" errors that
// shouldn't fail the entire import - the resource may have been deleted or made private.
func isResourceGoneError(err error) bool {
	if apiErr, ok := err.(*strava.APIError); ok {
		return apiErr.IsNotFound() || apiErr.IsForbidden()
	}
	return false
}

// shouldFetchSegmentDetail checks if a segment needs full detail fetch.
func shouldFetchSegmentDetail(seg *Segment) bool {
	if seg.Polyline == "" {
		return true
	}
	if seg.AthleteSegmentStats.PRElapsedTime == 0 && seg.AthleteSegmentStats.PRDate == nil && seg.AthleteSegmentStats.EffortCount == 0 && seg.AthleteSegmentStats.KOMRank == nil {
		return true
	}
	return false
}
