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
	var includeStreams bool

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import activities from Strava",
		Long: `Import activities from Strava API into the local database.

This command fetches all activities from your Strava account and stores
them locally for analysis. It respects Strava's rate limits and can be
interrupted and resumed.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runImport(fullSync, includeStreams)
		},
	}

	cmd.Flags().BoolVar(&fullSync, "full", false, "Perform a full sync (re-import all activities)")
	cmd.Flags().BoolVar(&includeStreams, "streams", false, "Include activity stream data (GPS, heartrate, etc.)")

	return cmd
}

func runImport(fullSync, includeStreams bool) error {
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

	// Create Strava client
	stravaClient := strava.NewClient(&cfg.Strava)

	// Check if we have a stored token
	// For now, we need to authenticate via the web interface first
	if !stravaClient.IsAuthenticated() {
		slog.Info("Not authenticated with Strava")
		slog.Info("Please start the server and authenticate via the web interface first:")
		slog.Info("  stata serve --dev")
		slog.Info("  Then visit http://localhost:8081/api/v1/auth/strava")
		return fmt.Errorf("not authenticated with Strava")
	}

	// Create importer
	imp := importer.New(stravaClient, activityRepo, athleteRepo, tokenRepo, gearRepo, streamRepo)

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
		FullSync:       fullSync,
		IncludeStreams: includeStreams,
	}

	slog.Info("Starting import",
		"full_sync", fullSync,
		"include_streams", includeStreams,
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
				slog.Info("Import progress",
					"imported", progress.ImportedCount,
					"total", progress.TotalActivities,
					"failed", progress.FailedCount,
					"page", progress.CurrentPage,
				)
			case importer.StatusCompleted:
				slog.Info("Import completed",
					"imported", progress.ImportedCount,
					"total", progress.TotalActivities,
					"failed", progress.FailedCount,
					"duration", progress.CompletedAt.Sub(progress.StartedAt).Round(time.Second),
				)
				return nil
			case importer.StatusFailed:
				return fmt.Errorf("import failed: %s", progress.Error)
			case importer.StatusCanceled:
				slog.Info("Import canceled",
					"imported", progress.ImportedCount,
					"total", progress.TotalActivities,
				)
				return nil
			}
		}
	}
}
