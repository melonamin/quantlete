// Package importer handles importing data from Strava.
package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
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
	stravaClient *strava.Client
	activities   *storage.ActivityRepository
	athletes     *storage.AthleteRepository
	tokens       *storage.TokenRepository
	gear         *storage.GearRepository
	streams      *storage.StreamRepository
	segments     *storage.SegmentRepository
	bestEfforts  *storage.BestEffortsRepository
	maintenance  *storage.MaintenanceRepository
	photos       *storage.PhotoRepository
	syncHistory  *storage.SyncHistoryRepository

	// State management for resume capability
	stateManager *StateManager
	state        *ImportState
	eta          *ETAEstimator

	// Current sync run (for history tracking)
	currentRunID int64

	segmentCacheMu sync.Mutex
	segmentCache   map[int64]*strava.Segment

	mu       sync.RWMutex
	progress Progress
	cancel   context.CancelFunc
}

// New creates a new importer.
func New(
	stravaClient *strava.Client,
	activities *storage.ActivityRepository,
	athletes *storage.AthleteRepository,
	tokens *storage.TokenRepository,
	gear *storage.GearRepository,
	streams *storage.StreamRepository,
	segments *storage.SegmentRepository,
	bestEfforts *storage.BestEffortsRepository,
	maintenance *storage.MaintenanceRepository,
	photos *storage.PhotoRepository,
	appState *storage.AppStateRepository,
	syncHistory *storage.SyncHistoryRepository,
) *Importer {
	return &Importer{
		stravaClient: stravaClient,
		activities:   activities,
		athletes:     athletes,
		tokens:       tokens,
		gear:         gear,
		streams:      streams,
		segments:     segments,
		bestEfforts:  bestEfforts,
		maintenance:  maintenance,
		photos:       photos,
		syncHistory:  syncHistory,
		stateManager: NewStateManager(appState),
		eta:          NewETAEstimator(),
		segmentCache: make(map[int64]*strava.Segment),
		progress:     Progress{Status: StatusIdle, Phase: PhaseIdle},
	}
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
	if athlete == nil {
		i.mu.Unlock()
		return fmt.Errorf("not authenticated")
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
		if !opts.FullSync && i.syncHistory != nil {
			wm, err := i.syncHistory.GetWatermark(ctx, athleteID)
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
	if i.syncHistory != nil {
		run, err := i.syncHistory.StartRun(ctx, athleteID, storage.SyncRunOptions{
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
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	i.cancel = cancel
	i.progress = Progress{
		Status:    StatusRunning,
		StartedAt: state.StartedAt,
		Phase:     state.Phase,
	}
	i.mu.Unlock()

	// Run import in background
	go func() {
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
		defer i.mu.Unlock()

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
				// Log canceled run
				if i.syncHistory != nil && i.currentRunID > 0 {
					if logErr := i.syncHistory.CancelRun(context.Background(), i.currentRunID, counts); logErr != nil {
						slog.Warn("failed to log canceled sync run", "error", logErr)
					}
				}
			} else {
				i.progress.Status = StatusFailed
				i.progress.Error = err.Error()
				// Log failed run
				if i.syncHistory != nil && i.currentRunID > 0 {
					if logErr := i.syncHistory.FailRun(context.Background(), i.currentRunID, err.Error(), counts); logErr != nil {
						slog.Warn("failed to log failed sync run", "error", logErr)
					}
				}
			}
		} else {
			i.progress.Status = StatusCompleted
			i.progress.Phase = PhaseCompleted

			// Log completed run
			if i.syncHistory != nil && i.currentRunID > 0 {
				if logErr := i.syncHistory.CompleteRun(context.Background(), i.currentRunID, counts); logErr != nil {
					slog.Warn("failed to log completed sync run", "error", logErr)
				}

				// Update watermark with newest activity date
				if i.state.NewestActivityDate != nil {
					wm := &storage.SyncWatermark{
						LastSyncedAt:       time.Now(),
						NewestActivityDate: i.state.NewestActivityDate,
					}
					if wmErr := i.syncHistory.SetWatermark(context.Background(), athleteID, wm); wmErr != nil {
						slog.Warn("failed to update sync watermark", "error", wmErr)
					} else {
						slog.Info("updated sync watermark",
							"newest_activity_date", i.state.NewestActivityDate.Format(time.RFC3339))
					}
				}
			}

			// Clear state on successful completion
			if clearErr := i.stateManager.Clear(context.Background()); clearErr != nil {
				slog.Warn("failed to clear import state", "error", clearErr)
			}
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

// Progress returns the current import progress.
func (i *Importer) Progress() Progress {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.progress
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

	stravaAthlete := i.stravaClient.GetAthlete()
	if stravaAthlete == nil {
		slog.Error("import failed: not authenticated")
		return fmt.Errorf("not authenticated")
	}
	slog.Info("import authenticated", "athlete_id", stravaAthlete.ID)

	// Save athlete profile
	athlete := convertAthlete(stravaAthlete)
	athlete.ID = stravaAthlete.ID
	if err := i.athletes.Upsert(ctx, athlete); err != nil {
		slog.Warn("failed to save athlete profile", "error", err)
	}

	// Phase 1: Activity list (enables dashboard immediately)
	if i.state.Phase == PhaseActivities {
		slog.Info("Phase 1: Importing activity list")
		if err := i.runActivitiesPhase(ctx, stravaAthlete.ID); err != nil {
			return err
		}
		i.state.Phase = PhaseGear
		i.saveState(ctx)
	}

	// Phase 2: Gear (quick, needed for gear stats)
	if i.state.Phase == PhaseGear {
		slog.Info("Phase 2: Importing gear")
		if err := i.runGearPhase(ctx, stravaAthlete.ID); err != nil {
			return err
		}
		i.state.Phase = PhaseStreams
		i.saveState(ctx)
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
	}

	// Phase 4: Activity Details (best efforts + segment efforts)
	if i.state.Phase == PhaseActivityDetails {
		if !i.state.SkipBestEfforts || !i.state.SkipSegments {
			slog.Info("Phase 4: Importing activity details (best efforts + segments)")
			if err := i.runActivityDetailsPhase(ctx, stravaAthlete.ID); err != nil {
				return err
			}
		} else {
			slog.Info("Phase 4: Skipping activity details (best efforts and segments disabled)")
		}
		i.state.Phase = PhaseSegmentDetails
		i.saveState(ctx)
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
	}

	// Phase 6: Photos (cosmetic, last)
	if i.state.Phase == PhasePhotos {
		if !i.state.SkipPhotos {
			slog.Info("Phase 6: Importing photos")
			if err := i.runPhotosPhase(ctx, stravaAthlete.ID); err != nil {
				return err
			}
		} else {
			slog.Info("Phase 6: Skipping photos (disabled)")
		}
		i.state.Phase = PhaseCompleted
		i.saveState(ctx)
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
func (i *Importer) saveState(ctx context.Context) {
	if err := i.stateManager.Save(ctx, i.state); err != nil {
		slog.Warn("failed to save import state", "error", err)
	}
}

// updateProgress updates the progress from current state.
func (i *Importer) updateProgress() {
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
	remaining := i.state.RemainingAPICalls()
	i.progress.RemainingAPICalls = remaining
	i.progress.EstimatedETA = FormatETA(i.eta.EstimateCompletion(remaining))
}

// waitForRateLimit handles rate limit errors by waiting and retrying.
// It updates progress to show the waiting state so the UI can display a countdown.
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

	// Wait for the rate limit to reset
	select {
	case <-ctx.Done():
		i.clearWaitingState()
		return ctx.Err()
	case <-time.After(rle.RetryAfter):
		i.clearWaitingState()
		return nil
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

		var activities []strava.Activity
		err := i.withRetry(ctx, func() error {
			var fetchErr error
			activities, fetchErr = i.stravaClient.GetActivitiesWithOptions(ctx, strava.GetActivitiesOptions{
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

		for _, a := range activities {
			if seenIDs[a.ID] {
				continue
			}
			seenIDs[a.ID] = true
			i.state.ActivityIDs = append(i.state.ActivityIDs, a.ID)
			i.state.ActivitiesTotal++

			// Convert and store activity
			act := convertActivity(&a, athleteID)

			// Hashtag-based custom gear linking
			if act.GearID == "" {
				if gearID, err := i.gear.ResolveCustomGearIDFromActivityName(ctx, athleteID, act.Name); err == nil && gearID != "" {
					act.GearID = gearID
				}
			}

			if err := i.activities.Upsert(ctx, act); err != nil {
				slog.Warn("failed to store activity", "id", a.ID, "error", err)
				i.state.FailedCount++
				continue
			}

			i.state.ActivitiesDone++

			// Track newest activity date for watermark
			if i.state.NewestActivityDate == nil || a.StartDate.After(*i.state.NewestActivityDate) {
				t := a.StartDate
				i.state.NewestActivityDate = &t
			}

			// Collect gear IDs
			if a.GearID != "" && !contains(i.state.GearIDs, a.GearID) {
				i.state.GearIDs = append(i.state.GearIDs, a.GearID)
			}

			// Log maintenance from hashtags
			if i.maintenance != nil {
				if _, err := i.maintenance.LogFromActivityHashtags(ctx, athleteID, act.ID, act.StartDateLocal.Time, act.Name); err != nil {
					slog.Debug("failed to log maintenance from hashtags", "activity_id", act.ID, "error", err)
				}
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
func (i *Importer) runGearPhase(ctx context.Context, athleteID int64) error {
	for idx := i.state.GearLastIndex; idx < len(i.state.GearIDs); idx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		gearID := i.state.GearIDs[idx]
		if err := i.importGearItem(ctx, gearID); err != nil {
			slog.Warn("failed to import gear", "id", gearID, "error", err)
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
		if err := i.importStreamsForActivity(ctx, activityID); err != nil {
			slog.Debug("failed to import streams", "activity_id", activityID, "error", err)
		} else {
			i.state.StreamsDone++
		}

		i.state.StreamsLastIndex = idx + 1
		i.updateProgress()

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
		var detail *strava.Activity
		err := i.withRetry(ctx, func() error {
			var fetchErr error
			detail, fetchErr = i.stravaClient.GetActivity(ctx, activityID)
			return fetchErr
		})
		if err != nil {
			slog.Debug("failed to fetch activity detail", "activity_id", activityID, "error", err)
			i.state.DetailsLastIndex = idx + 1
			continue
		}

		// Import best efforts
		if !i.state.SkipBestEfforts {
			if err := i.importBestEfforts(ctx, detail, athleteID); err != nil {
				slog.Debug("failed to import best efforts", "activity_id", activityID, "error", err)
			}
		}

		// Import segment efforts and collect segment IDs
		if !i.state.SkipSegments {
			for _, effort := range detail.SegmentEfforts {
				seg := effort.Segment

				// Store segment effort
				if err := i.storeSegmentEffort(ctx, detail, &effort, athleteID); err != nil {
					slog.Debug("failed to store segment effort", "effort_id", effort.ID, "error", err)
				}

				// Check if we need to fetch full segment detail
				if shouldFetchSegmentDetail(seg) {
					if !containsInt64(i.state.SegmentIDsToFetch, seg.ID) {
						i.state.SegmentIDsToFetch = append(i.state.SegmentIDsToFetch, seg.ID)
					}
				} else {
					// We have enough data, store segment now
					if err := i.storeSegmentFromEffort(ctx, &seg); err != nil {
						slog.Debug("failed to store segment", "segment_id", seg.ID, "error", err)
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

		var seg *strava.Segment
		fetchErr := i.withRetry(ctx, func() error {
			var err error
			seg, err = i.stravaClient.GetSegment(ctx, segmentID)
			return err
		})
		if fetchErr != nil {
			slog.Debug("failed to fetch segment detail", "segment_id", segmentID, "error", fetchErr)
		} else if err := i.storeSegment(ctx, seg); err != nil {
			slog.Debug("failed to store segment", "segment_id", segmentID, "error", err)
		} else {
			i.state.SegmentsDone++
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
func (i *Importer) runPhotosPhase(ctx context.Context, athleteID int64) error {
	for idx := i.state.PhotosLastIndex; idx < len(i.state.ActivityIDs); idx++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		activityID := i.state.ActivityIDs[idx]
		if err := i.importPhotosForActivity(ctx, activityID, athleteID); err != nil {
			slog.Debug("failed to import photos", "activity_id", activityID, "error", err)
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

// importActivity imports a single activity (used for webhook/incremental sync).
func (i *Importer) importActivity(ctx context.Context, a *strava.Activity, athleteID int64, opts ImportOptions) error {
	actSource := a
	var detail *strava.Activity

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

	// Convert Strava activity to storage activity
	act := convertActivity(actSource, athleteID)

	// Hashtag-based custom gear linking (only if Strava gear_id is empty).
	if act.GearID == "" {
		if gearID, err := i.gear.ResolveCustomGearIDFromActivityName(ctx, athleteID, act.Name); err == nil && gearID != "" {
			act.GearID = gearID
		}
	}

	if err := i.activities.Upsert(ctx, act); err != nil {
		return fmt.Errorf("storing activity: %w", err)
	}

	// Import streams unless skipped
	if !opts.SkipStreams {
		if err := i.importStreams(ctx, a.ID); err != nil {
			slog.Debug("failed to import streams", "activity_id", a.ID, "error", err)
		}
	}

	if !opts.SkipSegments {
		source := detail
		if source == nil {
			source = actSource
		}
		if err := i.importSegments(ctx, source, athleteID); err != nil {
			slog.Debug("failed to import segments", "activity_id", a.ID, "error", err)
		}
	}

	if !opts.SkipBestEfforts {
		source := detail
		if source == nil {
			source = actSource
		}
		if err := i.importBestEfforts(ctx, source, athleteID); err != nil {
			slog.Debug("failed to import best efforts", "activity_id", a.ID, "error", err)
		}
	}

	if !opts.SkipPhotos && i.photos != nil {
		if err := i.importPhotos(ctx, act.ID, athleteID); err != nil {
			slog.Debug("failed to import photos", "activity_id", act.ID, "error", err)
		}
	}

	// Hashtag-based maintenance logging.
	if i.maintenance != nil {
		if _, err := i.maintenance.LogFromActivityHashtags(ctx, athleteID, act.ID, act.StartDateLocal.Time, act.Name); err != nil {
			slog.Debug("failed to log maintenance from hashtags", "activity_id", act.ID, "error", err)
		}
	}

	return nil
}

func canonicalBestEffortDistanceType(distanceM float64, name string) (distanceType string, canonicalM float64) {
	// Prefer matching by distance (tolerant to minor rounding).
	type candidate struct {
		Type string
		M    float64
	}
	candidates := []candidate{
		{Type: "400m", M: 400},
		{Type: "0.5mi", M: 804.672},
		{Type: "1k", M: 1000},
		{Type: "1mi", M: 1609.344},
		{Type: "2mi", M: 3218.688},
		{Type: "5k", M: 5000},
		{Type: "10k", M: 10000},
		{Type: "15k", M: 15000},
		{Type: "10mi", M: 16093.44},
		{Type: "20k", M: 20000},
		{Type: "half_marathon", M: 21097.5},
		{Type: "30k", M: 30000},
		{Type: "marathon", M: 42195},
		{Type: "50k", M: 50000},
		{Type: "100k", M: 100000},
	}

	if distanceM > 0 {
		for _, c := range candidates {
			// Accept within 1% or 25m.
			tol := 25.0
			if c.M*0.01 > tol {
				tol = c.M * 0.01
			}
			if distanceM >= c.M-tol && distanceM <= c.M+tol {
				return c.Type, c.M
			}
		}
	}

	// Fallback: best-effort name or rounded meters.
	n := strings.ToLower(strings.TrimSpace(name))
	n = strings.ReplaceAll(n, " ", "_")
	n = strings.ReplaceAll(n, "/", "_")
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.Trim(n, "_")
	if n != "" {
		return n, distanceM
	}
	return fmt.Sprintf("m_%d", int(distanceM+0.5)), distanceM
}

func (i *Importer) importBestEfforts(ctx context.Context, a *strava.Activity, athleteID int64) error {
	if i.bestEfforts == nil {
		return nil
	}
	if a == nil || len(a.BestEfforts) == 0 {
		return nil
	}

	var out []storage.BestEffort
	for _, be := range a.BestEfforts {
		if be.ElapsedTime <= 0 || be.Distance <= 0 {
			continue
		}
		dt, canonM := canonicalBestEffortDistanceType(be.Distance, be.Name)
		start := storage.SQLiteTime{Time: be.StartDate}
		moving := be.MovingTime
		out = append(out, storage.BestEffort{
			AthleteID:    athleteID,
			ActivityID:   a.ID,
			SportType:    a.SportType,
			DistanceType: dt,
			Name:         be.Name,
			DistanceM:    canonM,
			ElapsedTimeS: be.ElapsedTime,
			MovingTimeS:  &moving,
			StartIndex:   be.StartIndex,
			EndIndex:     be.EndIndex,
			PRRank:       be.PRRank,
			StartDate:    &start,
		})
	}

	return i.bestEfforts.ReplaceForActivity(ctx, athleteID, a.ID, a.SportType, out)
}

func (i *Importer) getSegment(ctx context.Context, segmentID int64) (*strava.Segment, error) {
	i.segmentCacheMu.Lock()
	if cached, ok := i.segmentCache[segmentID]; ok && cached != nil {
		i.segmentCacheMu.Unlock()
		return cached, nil
	}
	i.segmentCacheMu.Unlock()

	seg, err := i.stravaClient.GetSegment(ctx, segmentID)
	if err != nil {
		return nil, err
	}

	i.segmentCacheMu.Lock()
	i.segmentCache[segmentID] = seg
	i.segmentCacheMu.Unlock()

	return seg, nil
}

func shouldFetchSegmentDetail(seg strava.Segment) bool {
	if seg.Map.Polyline == "" {
		return true
	}
	if seg.AthleteSegmentStats.PRElapsedTime == 0 && seg.AthleteSegmentStats.PRDate == nil && seg.AthleteSegmentStats.EffortCount == 0 && seg.AthleteSegmentStats.KOMRank == nil {
		return true
	}
	return false
}

func ptrFloat64(v float64) *float64 {
	if v == 0 {
		return nil
	}
	return &v
}

func ptrInt(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

func ptrSQLiteTime(t *time.Time) *storage.SQLiteTime {
	if t == nil {
		return nil
	}
	return &storage.SQLiteTime{Time: *t}
}

func floatSliceLatLng(latlng []float64) (lat *float64, lng *float64) {
	if len(latlng) < 2 {
		return nil, nil
	}
	latV := latlng[0]
	lngV := latlng[1]
	return &latV, &lngV
}

func (i *Importer) importSegments(ctx context.Context, a *strava.Activity, athleteID int64) error {
	if i.segments == nil {
		return nil
	}
	if a == nil || len(a.SegmentEfforts) == 0 {
		return nil
	}

	for _, effort := range a.SegmentEfforts {
		seg := effort.Segment
		segDetail := &seg
		if shouldFetchSegmentDetail(seg) {
			if full, err := i.getSegment(ctx, seg.ID); err == nil && full != nil {
				segDetail = full
			}
		}

		startLat, startLng := floatSliceLatLng(segDetail.StartLatlng)
		endLat, endLng := floatSliceLatLng(segDetail.EndLatlng)

		var effortCount *int
		if segDetail.AthleteSegmentStats.EffortCount >= 0 {
			v := segDetail.AthleteSegmentStats.EffortCount
			effortCount = &v
		}
		var prElapsed *int
		if segDetail.AthleteSegmentStats.PRElapsedTime > 0 {
			v := segDetail.AthleteSegmentStats.PRElapsedTime
			prElapsed = &v
		}

		s := &storage.Segment{
			ID:                   segDetail.ID,
			Name:                 segDetail.Name,
			ActivityType:         segDetail.ActivityType,
			Distance:             segDetail.Distance,
			AverageGrade:         segDetail.AverageGrade,
			MaximumGrade:         segDetail.MaximumGrade,
			ElevationHigh:        segDetail.ElevationHigh,
			ElevationLow:         segDetail.ElevationLow,
			ClimbCategory:        segDetail.ClimbCategory,
			StartLat:             startLat,
			StartLng:             startLng,
			EndLat:               endLat,
			EndLng:               endLng,
			Starred:              segDetail.Starred,
			Polyline:             segDetail.Map.Polyline,
			AthleteKOMRank:       segDetail.AthleteSegmentStats.KOMRank,
			AthleteEffortCount:   effortCount,
			AthletePRElapsedTime: prElapsed,
			AthletePRDate:        ptrSQLiteTime(segDetail.AthleteSegmentStats.PRDate),
		}
		if err := i.segments.UpsertSegment(ctx, s); err != nil {
			return fmt.Errorf("upserting segment %d: %w", segDetail.ID, err)
		}

		startDate := storage.SQLiteTime{Time: effort.StartDate}
		startDateLocal := storage.SQLiteTime{Time: effort.StartDateLocal}
		e := &storage.SegmentEffort{
			ID:               effort.ID,
			SegmentID:        segDetail.ID,
			ActivityID:       a.ID,
			AthleteID:        athleteID,
			Name:             effort.Name,
			ElapsedTime:      effort.ElapsedTime,
			MovingTime:       effort.MovingTime,
			StartDate:        &startDate,
			StartDateLocal:   &startDateLocal,
			Distance:         effort.Distance,
			AverageWatts:     ptrFloat64(effort.AverageWatts),
			AverageHeartrate: ptrFloat64(effort.AverageHeartrate),
			MaxHeartrate:     ptrInt(int(effort.MaxHeartrate)),
			PRRank:           effort.PRRank,
			Country:          a.LocationCountry,
		}
		if err := i.segments.UpsertEffort(ctx, e); err != nil {
			return fmt.Errorf("upserting segment effort %d: %w", effort.ID, err)
		}
	}

	return nil
}

func (i *Importer) importPhotos(ctx context.Context, activityID int64, athleteID int64) error {
	var photos []strava.ActivityPhoto
	err := i.withRetry(ctx, func() error {
		var fetchErr error
		photos, fetchErr = i.stravaClient.GetActivityPhotos(ctx, activityID)
		return fetchErr
	})
	if err != nil {
		return err
	}

	for _, p := range photos {
		url, thumb := bestPhotoURLs(p.URLs)
		if url == "" {
			continue
		}

		id := p.UniqueID
		if id == "" {
			id = strconv.FormatInt(p.ID, 10)
		}

		var loc json.RawMessage
		if len(p.Location) > 0 {
			if b, err := json.Marshal(p.Location); err == nil {
				loc = b
			}
		}

		err := i.photos.Upsert(ctx, &storage.Photo{
			ID:           id,
			AthleteID:    athleteID,
			ActivityID:   activityID,
			URL:          url,
			ThumbnailURL: thumb,
			Caption:      p.Caption,
			Location:     loc,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func bestPhotoURLs(urls map[string]string) (best string, thumb string) {
	if len(urls) == 0 {
		return "", ""
	}
	type kv struct {
		k int
		v string
	}
	var items []kv
	for k, v := range urls {
		if strings.TrimSpace(v) == "" {
			continue
		}
		if n, err := strconv.Atoi(k); err == nil {
			items = append(items, kv{k: n, v: v})
		}
	}
	if len(items) == 0 {
		for _, v := range urls {
			if strings.TrimSpace(v) != "" {
				return v, ""
			}
		}
		return "", ""
	}

	sort.Slice(items, func(i, j int) bool { return items[i].k < items[j].k })
	thumb = items[0].v
	best = items[len(items)-1].v
	return best, thumb
}

// importStreams imports stream data for an activity.
func (i *Importer) importStreams(ctx context.Context, activityID int64) error {
	var streams *strava.StreamSet
	err := i.withRetry(ctx, func() error {
		var fetchErr error
		streams, fetchErr = i.stravaClient.GetActivityStreams(ctx, activityID, nil)
		return fetchErr
	})
	if err != nil {
		return err
	}

	// Store each stream type
	if streams.Time != nil {
		if err := i.storeStream(ctx, activityID, streams.Time); err != nil {
			return err
		}
	}
	if streams.Distance != nil {
		if err := i.storeStream(ctx, activityID, streams.Distance); err != nil {
			return err
		}
	}
	if streams.Altitude != nil {
		if err := i.storeStream(ctx, activityID, streams.Altitude); err != nil {
			return err
		}
	}
	if streams.Heartrate != nil {
		if err := i.storeStream(ctx, activityID, streams.Heartrate); err != nil {
			return err
		}
	}
	if streams.Watts != nil {
		if err := i.storeStream(ctx, activityID, streams.Watts); err != nil {
			return err
		}
	}
	if streams.Cadence != nil {
		if err := i.storeStream(ctx, activityID, streams.Cadence); err != nil {
			return err
		}
	}
	if streams.VelocitySmooth != nil {
		if err := i.storeStream(ctx, activityID, streams.VelocitySmooth); err != nil {
			return err
		}
	}
	if streams.Latlng != nil {
		if err := i.storeStream(ctx, activityID, streams.Latlng); err != nil {
			return err
		}
	}

	return nil
}

// storeStream stores a single stream.
func (i *Importer) storeStream(ctx context.Context, activityID int64, s *strava.Stream) error {
	if s == nil || len(s.Data) == 0 {
		return nil
	}

	stream := &storage.ActivityStream{
		ActivityID:   activityID,
		StreamType:   s.Type,
		OriginalSize: s.OriginalSize,
		Resolution:   s.Resolution,
		SeriesType:   s.SeriesType,
	}

	// Encode data as JSON
	data, err := encodeStreamData(s.Data)
	if err != nil {
		return err
	}
	stream.Data = data

	return i.streams.Upsert(ctx, stream)
}

// importGearItem imports gear details (alias for importGear).
func (i *Importer) importGearItem(ctx context.Context, gearID string) error {
	return i.importGear(ctx, gearID)
}

// importStreamsForActivity imports streams for a single activity.
func (i *Importer) importStreamsForActivity(ctx context.Context, activityID int64) error {
	return i.importStreams(ctx, activityID)
}

// importPhotosForActivity imports photos for a single activity.
func (i *Importer) importPhotosForActivity(ctx context.Context, activityID int64, athleteID int64) error {
	return i.importPhotos(ctx, activityID, athleteID)
}

// storeSegmentEffort stores a segment effort from an activity detail.
func (i *Importer) storeSegmentEffort(ctx context.Context, a *strava.Activity, effort *strava.SegmentEffort, athleteID int64) error {
	if i.segments == nil {
		return nil
	}

	startDate := storage.SQLiteTime{Time: effort.StartDate}
	startDateLocal := storage.SQLiteTime{Time: effort.StartDateLocal}
	e := &storage.SegmentEffort{
		ID:               effort.ID,
		SegmentID:        effort.Segment.ID,
		ActivityID:       a.ID,
		AthleteID:        athleteID,
		Name:             effort.Name,
		ElapsedTime:      effort.ElapsedTime,
		MovingTime:       effort.MovingTime,
		StartDate:        &startDate,
		StartDateLocal:   &startDateLocal,
		Distance:         effort.Distance,
		AverageWatts:     ptrFloat64(effort.AverageWatts),
		AverageHeartrate: ptrFloat64(effort.AverageHeartrate),
		MaxHeartrate:     ptrInt(int(effort.MaxHeartrate)),
		PRRank:           effort.PRRank,
		Country:          a.LocationCountry,
	}
	return i.segments.UpsertEffort(ctx, e)
}

// storeSegmentFromEffort stores a segment from the effort's embedded segment data.
func (i *Importer) storeSegmentFromEffort(ctx context.Context, seg *strava.Segment) error {
	if i.segments == nil {
		return nil
	}

	startLat, startLng := floatSliceLatLng(seg.StartLatlng)
	endLat, endLng := floatSliceLatLng(seg.EndLatlng)

	var effortCount *int
	if seg.AthleteSegmentStats.EffortCount >= 0 {
		v := seg.AthleteSegmentStats.EffortCount
		effortCount = &v
	}
	var prElapsed *int
	if seg.AthleteSegmentStats.PRElapsedTime > 0 {
		v := seg.AthleteSegmentStats.PRElapsedTime
		prElapsed = &v
	}

	s := &storage.Segment{
		ID:                   seg.ID,
		Name:                 seg.Name,
		ActivityType:         seg.ActivityType,
		Distance:             seg.Distance,
		AverageGrade:         seg.AverageGrade,
		MaximumGrade:         seg.MaximumGrade,
		ElevationHigh:        seg.ElevationHigh,
		ElevationLow:         seg.ElevationLow,
		ClimbCategory:        seg.ClimbCategory,
		StartLat:             startLat,
		StartLng:             startLng,
		EndLat:               endLat,
		EndLng:               endLng,
		Starred:              seg.Starred,
		Polyline:             seg.Map.Polyline,
		AthleteKOMRank:       seg.AthleteSegmentStats.KOMRank,
		AthleteEffortCount:   effortCount,
		AthletePRElapsedTime: prElapsed,
		AthletePRDate:        ptrSQLiteTime(seg.AthleteSegmentStats.PRDate),
	}
	return i.segments.UpsertSegment(ctx, s)
}

// storeSegment stores a segment from full segment detail.
func (i *Importer) storeSegment(ctx context.Context, seg *strava.Segment) error {
	return i.storeSegmentFromEffort(ctx, seg)
}

// importGear imports gear details.
func (i *Importer) importGear(ctx context.Context, gearID string) error {
	var gear *strava.Gear
	err := i.withRetry(ctx, func() error {
		var fetchErr error
		gear, fetchErr = i.stravaClient.GetGear(ctx, gearID)
		return fetchErr
	})
	if err != nil {
		return err
	}

	athlete := i.stravaClient.GetAthlete()
	if athlete == nil {
		return fmt.Errorf("not authenticated")
	}

	g := &storage.Gear{
		ID:          gear.ID,
		AthleteID:   athlete.ID,
		Name:        gear.Name,
		Primary:     gear.Primary,
		Retired:     gear.Retired,
		Distance:    gear.Distance,
		BrandName:   gear.BrandName,
		ModelName:   gear.ModelName,
		Description: gear.Description,
	}

	return i.gear.Upsert(ctx, g)
}
