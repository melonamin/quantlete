package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/sasha/stata/internal/config"
	"github.com/sasha/stata/internal/importer"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

func newImportCmd() *cobra.Command {
	var fullSync bool
	var resume bool
	var skipStreams bool
	var skipSegments bool
	var skipBestEfforts bool
	var skipPhotos bool

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import activities from Strava",
		Long: `Import activities from Strava API into the local database.

This command fetches all activities from your Strava account and stores
them locally for analysis. It respects Strava's rate limits and can be
interrupted and resumed.

By default, all data types are imported. Use --skip-* flags to exclude
specific data types if needed.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runImport(fullSync, resume, skipStreams, skipSegments, skipBestEfforts, skipPhotos)
		},
	}

	cmd.Flags().BoolVar(&fullSync, "full", false, "Perform a full sync (re-import all activities)")
	cmd.Flags().BoolVar(&resume, "resume", false, "Resume from previous interrupted import")
	cmd.Flags().BoolVar(&skipStreams, "skip-streams", false, "Skip importing activity stream data (GPS, heartrate, etc.)")
	cmd.Flags().BoolVar(&skipSegments, "skip-segments", false, "Skip importing segment efforts and segment details")
	cmd.Flags().BoolVar(&skipBestEfforts, "skip-best-efforts", false, "Skip importing Strava best efforts/PRs")
	cmd.Flags().BoolVar(&skipPhotos, "skip-photos", false, "Skip importing activity photos")

	return cmd
}

func runImport(fullSync, resume, skipStreams, skipSegments, skipBestEfforts, skipPhotos bool) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Setup logging
	logLevel := slog.LevelInfo
	if cfg.Log.Level == "debug" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Check Strava configuration
	if cfg.Strava.ClientID == "" || cfg.Strava.ClientSecret == "" {
		return fmt.Errorf("strava credentials not configured, set STATA_STRAVA_CLIENT_ID and STATA_STRAVA_CLIENT_SECRET environment variables")
	}

	// Open database
	db, err := storage.Open(cfg.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Run migrations
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Create repositories
	activityRepo := storage.NewActivityRepository(db)
	athleteRepo := storage.NewAthleteRepository(db)
	tokenRepo := storage.NewTokenRepository(db)
	gearRepo := storage.NewGearRepository(db)
	streamRepo := storage.NewStreamRepository(db)
	segmentRepo := storage.NewSegmentRepository(db)
	bestEffortsRepo := storage.NewBestEffortsRepository(db)
	maintenanceRepo := storage.NewMaintenanceRepository(db)
	photoRepo := storage.NewPhotoRepository(db)
	appStateRepo := storage.NewAppStateRepository(db)

	// Create Strava client
	stravaClient := strava.NewClient(&cfg.Strava)

	// Restore rate limit state from database
	if stateJSON, err := appStateRepo.Get(context.Background(), "strava_rate_limit"); err == nil && stateJSON != "" {
		if err := stravaClient.RateLimiter().LoadFromJSON(stateJSON); err != nil {
			slog.Warn("failed to restore rate limit state", "error", err)
		} else {
			status := stravaClient.RateLimiter().Status()
			slog.Info("restored rate limit state",
				"usage_15min", status.Usage15Min,
				"limit_15min", status.Limit15Min,
				"usage_daily", status.UsageDaily,
				"limit_daily", status.LimitDaily,
			)
		}
	}

	// Set up rate limit persistence
	stravaClient.SetRateLimitPersister(func(json string) error {
		return appStateRepo.Set(context.Background(), "strava_rate_limit", json)
	})

	// Check if we have a stored token
	// For now, we need to authenticate via the web interface first
	if !stravaClient.IsAuthenticated() {
		slog.Info("Not authenticated with Strava")
		slog.Info("Please start the server and authenticate via the web interface first:")
		slog.Info("  stata serve --dev")
		slog.Info("  Then visit http://localhost:8081/api/v1/auth/strava")
		return fmt.Errorf("not authenticated with Strava")
	}

	// Create sync history repository
	syncHistoryRepo := storage.NewSyncHistoryRepository(db, appStateRepo)

	// Create importer
	imp := importer.New(stravaClient, activityRepo, athleteRepo, tokenRepo, gearRepo, streamRepo, segmentRepo, bestEffortsRepo, maintenanceRepo, photoRepo, appStateRepo, syncHistoryRepo)

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("Interrupt received, stopping import...")
		imp.Cancel()
		cancel()
	}()

	// Start import
	opts := importer.ImportOptions{
		FullSync:        fullSync,
		Resume:          resume,
		SkipStreams:     skipStreams,
		SkipSegments:    skipSegments,
		SkipBestEfforts: skipBestEfforts,
		SkipPhotos:      skipPhotos,
	}

	slog.Info("Starting import",
		"full_sync", fullSync,
		"resume", resume,
		"skip_streams", skipStreams,
		"skip_segments", skipSegments,
		"skip_best_efforts", skipBestEfforts,
		"skip_photos", skipPhotos,
	)

	if err := imp.Start(ctx, opts); err != nil {
		return fmt.Errorf("starting import: %w", err)
	}

	// Wait for completion
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			progress := imp.Progress()

			switch progress.Status {
			case importer.StatusRunning:
				// Show phase-specific progress
				var phaseProgress string
				switch progress.Phase {
				case importer.PhaseActivities:
					phaseProgress = fmt.Sprintf("activities %d/%d", progress.ActivitiesDone, progress.ActivitiesTotal)
				case importer.PhaseGear:
					phaseProgress = fmt.Sprintf("gear %d/%d", progress.GearDone, progress.GearTotal)
				case importer.PhaseStreams:
					phaseProgress = fmt.Sprintf("streams %d/%d", progress.StreamsDone, progress.StreamsTotal)
				case importer.PhaseActivityDetails:
					phaseProgress = fmt.Sprintf("details %d/%d", progress.DetailsDone, progress.DetailsTotal)
				case importer.PhaseSegmentDetails:
					phaseProgress = fmt.Sprintf("segments %d/%d", progress.SegmentsDone, progress.SegmentsTotal)
				case importer.PhasePhotos:
					phaseProgress = fmt.Sprintf("photos %d/%d", progress.PhotosDone, progress.PhotosTotal)
				default:
					phaseProgress = string(progress.Phase)
				}

				slog.Info("Import progress",
					"phase", progress.Phase,
					"progress", phaseProgress,
					"failed", progress.FailedCount,
					"eta", progress.EstimatedETA,
				)
			case importer.StatusCompleted:
				slog.Info("Import completed",
					"activities", progress.ActivitiesDone,
					"gear", progress.GearDone,
					"streams", progress.StreamsDone,
					"segments", progress.SegmentsDone,
					"photos", progress.PhotosDone,
					"failed", progress.FailedCount,
					"duration", progress.CompletedAt.Sub(progress.StartedAt).Round(time.Second),
				)
				return nil
			case importer.StatusFailed:
				return fmt.Errorf("import failed: %s", progress.Error)
			case importer.StatusCanceled:
				slog.Info("Import canceled",
					"phase", progress.Phase,
					"activities", progress.ActivitiesDone,
					"failed", progress.FailedCount,
				)
				return nil
			}
		}
	}
}
