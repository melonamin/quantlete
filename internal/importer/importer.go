// Package importer handles importing data from Strava.
package importer

import (
	"context"
	"fmt"
	"log/slog"
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
	Status          Status    `json:"status"`
	StartedAt       time.Time `json:"started_at,omitempty"`
	CompletedAt     time.Time `json:"completed_at,omitempty"`
	TotalActivities int       `json:"total_activities"`
	ImportedCount   int       `json:"imported_count"`
	SkippedCount    int       `json:"skipped_count"`
	FailedCount     int       `json:"failed_count"`
	CurrentPage     int       `json:"current_page"`
	Error           string    `json:"error,omitempty"`
}

// Importer orchestrates data import from Strava.
type Importer struct {
	stravaClient *strava.Client
	activities   *storage.ActivityRepository
	athletes     *storage.AthleteRepository
	tokens       *storage.TokenRepository
	gear         *storage.GearRepository
	streams      *storage.StreamRepository

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
) *Importer {
	return &Importer{
		stravaClient: stravaClient,
		activities:   activities,
		athletes:     athletes,
		tokens:       tokens,
		gear:         gear,
		streams:      streams,
		progress:     Progress{Status: StatusIdle},
	}
}

// ImportOptions configures an import run.
type ImportOptions struct {
	FullSync       bool // If true, re-import all activities
	IncludeStreams bool // If true, also import stream data
}

// Start begins an import operation.
func (i *Importer) Start(ctx context.Context, opts ImportOptions) error {
	i.mu.Lock()
	if i.progress.Status == StatusRunning {
		i.mu.Unlock()
		return fmt.Errorf("import already in progress")
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	i.cancel = cancel
	i.progress = Progress{
		Status:    StatusRunning,
		StartedAt: time.Now(),
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
		if err != nil {
			slog.Error("import failed", "error", err)
			if ctx.Err() == context.Canceled {
				i.progress.Status = StatusCanceled
			} else {
				i.progress.Status = StatusFailed
				i.progress.Error = err.Error()
			}
		} else {
			i.progress.Status = StatusCompleted
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

// runImport performs the actual import.
func (i *Importer) runImport(ctx context.Context, opts ImportOptions) error {
	slog.Info("starting import", "full_sync", opts.FullSync, "include_streams", opts.IncludeStreams)

	stravaAthlete := i.stravaClient.GetAthlete()
	if stravaAthlete == nil {
		slog.Error("import failed: not authenticated")
		return fmt.Errorf("not authenticated")
	}
	slog.Info("import authenticated", "athlete_id", stravaAthlete.ID)

	// Save athlete profile
	slog.Info("saving athlete profile")
	athlete := convertAthlete(stravaAthlete)
	athlete.ID = stravaAthlete.ID
	if err := i.athletes.Upsert(ctx, athlete); err != nil {
		slog.Warn("failed to save athlete profile", "error", err)
	}
	slog.Info("athlete profile saved")

	// Import activities page by page
	slog.Info("starting activity fetch loop")
	page := 1
	perPage := 100
	seenIDs := make(map[int64]bool)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		i.mu.Lock()
		i.progress.CurrentPage = page
		i.mu.Unlock()

		slog.Info("fetching activities page", "page", page)

		activities, err := i.stravaClient.GetActivities(ctx, page, perPage)
		if err != nil {
			slog.Error("failed to fetch activities", "page", page, "error", err)
			return fmt.Errorf("fetching activities page %d: %w", page, err)
		}

		slog.Info("fetched activities", "page", page, "count", len(activities))

		if len(activities) == 0 {
			slog.Info("no more activities, import done")
			break
		}

		for _, a := range activities {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if seenIDs[a.ID] {
				continue
			}
			seenIDs[a.ID] = true

			i.mu.Lock()
			i.progress.TotalActivities++
			i.mu.Unlock()

			// Convert and store activity
			if err := i.importActivity(ctx, &a, stravaAthlete.ID, opts.IncludeStreams); err != nil {
				slog.Warn("failed to import activity", "id", a.ID, "error", err)
				i.mu.Lock()
				i.progress.FailedCount++
				i.mu.Unlock()
				continue
			}

			i.mu.Lock()
			i.progress.ImportedCount++
			i.mu.Unlock()

			// Import gear if referenced
			if a.GearID != "" {
				if err := i.importGear(ctx, a.GearID); err != nil {
					slog.Warn("failed to import gear", "id", a.GearID, "error", err)
				}
			}
		}

		if len(activities) < perPage {
			break
		}
		page++
	}

	slog.Info("import completed",
		"total", i.progress.TotalActivities,
		"imported", i.progress.ImportedCount,
		"failed", i.progress.FailedCount,
	)

	return nil
}

// importActivity imports a single activity.
func (i *Importer) importActivity(ctx context.Context, a *strava.Activity, athleteID int64, includeStreams bool) error {
	// Convert Strava activity to storage activity
	act := convertActivity(a, athleteID)

	if err := i.activities.Upsert(ctx, act); err != nil {
		return fmt.Errorf("storing activity: %w", err)
	}

	// Optionally import streams
	if includeStreams {
		if err := i.importStreams(ctx, a.ID); err != nil {
			slog.Debug("failed to import streams", "activity_id", a.ID, "error", err)
		}
	}

	return nil
}

// importStreams imports stream data for an activity.
func (i *Importer) importStreams(ctx context.Context, activityID int64) error {
	streams, err := i.stravaClient.GetActivityStreams(ctx, activityID, nil)
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

// importGear imports gear details.
func (i *Importer) importGear(ctx context.Context, gearID string) error {
	gear, err := i.stravaClient.GetGear(ctx, gearID)
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
