package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/melonamin/quantlete/internal/api"
	"github.com/melonamin/quantlete/internal/api/handlers"
	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/importer"
	"github.com/melonamin/quantlete/internal/scheduler"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

// notificationHandlerTimeout is the timeout for notification handlers.
const notificationHandlerTimeout = 30 * time.Second

func newServeCmd() *cobra.Command {
	var port int
	var dev bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web server",
		Long: `Start the Quantlete web server to serve the dashboard.

In development mode (--dev), the server expects the React dev server
to be running separately and will proxy API requests.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runServe(port, dev)
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 0, "Port to listen on (overrides config)")
	cmd.Flags().BoolVar(&dev, "dev", false, "Run in development mode")

	return cmd
}

func runServe(port int, dev bool) error {
	// Load configuration
	cfg, err := config.LoadWithOverrides(port, dev)
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

	// Open database
	db, err := storage.Open(cfg.Storage.DataDir, cfg.Storage.DBFile)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = db.Close() }()

	slog.Info("database opened", "path", db.Path())

	// Run migrations
	if err = db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Set redirect URI from external URL if not explicitly configured
	if cfg.Strava.RedirectURI == "" || cfg.Strava.RedirectURI == "http://localhost:8081/api/v1/auth/strava/callback" {
		cfg.Strava.RedirectURI = cfg.Server.GetOAuthCallbackURL()
		slog.Debug("using external URL for OAuth callback", "redirect_uri", cfg.Strava.RedirectURI)
	}

	// Create Strava client
	stravaClient := strava.NewClient(&cfg.Strava)

	// Check for demo mode early
	appStateRepo := storage.NewAppStateRepository(db)
	demoMode, err := appStateRepo.Get(context.Background(), storage.AppStateDemoMode)
	if err != nil {
		slog.Warn("failed to check demo mode", "error", err)
	}
	if demoMode == "true" {
		slog.Info("Running in demo mode - Strava API calls disabled")
		if err = loadDemoAthlete(context.Background(), stravaClient, appStateRepo, storage.NewAthleteRepository(db)); err != nil {
			slog.Warn("failed to load demo athlete", "error", err)
		}
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
	settingsRepo := storage.NewSettingsRepository(db)
	statsRepo := storage.NewStatsRepository(db)

	// Create repositories for achievement detection
	metricsRepo := storage.NewAthleteMetricsRepository(db)
	zonesRepo := storage.NewZonesRepository(db)
	powerRepo := storage.NewPowerRepository(db, streamRepo)
	trainingLoadRepo := storage.NewTrainingLoadRepository(db, streamRepo, metricsRepo, zonesRepo)
	achievementRepo := storage.NewAchievementRepository(db, statsRepo, gearRepo, powerRepo, trainingLoadRepo)

	// Only do Strava auth setup if not in demo mode
	if demoMode != "true" {
		// Load credentials from database if not set via environment variables
		if err = handlers.LoadCredentialsFromDB(context.Background(), appStateRepo, cfg, stravaClient); err != nil {
			slog.Warn("failed to load credentials from database", "error", err)
		}

		// Restore rate limit state from database
		if err = restoreRateLimitState(context.Background(), stravaClient, appStateRepo); err != nil {
			slog.Warn("failed to restore rate limit state", "error", err)
		}

		// Set up rate limit persistence
		stravaClient.SetRateLimitPersister(func(json string) error {
			return appStateRepo.Set(context.Background(), storage.AppStateStravaRateLimit, json)
		})

		// Set up token persistence for automatic refresh.
		stravaClient.SetTokenPersister(func(token *oauth2.Token, athleteID int64) error {
			if token == nil {
				return nil
			}
			return tokenRepo.Upsert(context.Background(), &storage.AuthToken{
				AthleteID:    athleteID,
				AccessToken:  token.AccessToken,
				RefreshToken: token.RefreshToken,
				TokenType:    token.TokenType,
				ExpiresAt:    storage.SQLiteTime{Time: token.Expiry},
			})
		})

		// Restore tokens from database
		if err = restoreAuth(context.Background(), stravaClient, tokenRepo, athleteRepo); err != nil {
			slog.Warn("failed to restore auth from database", "error", err)
		}
	}

	// Create sync history repository
	syncHistoryRepo := storage.NewSyncHistoryRepository(db, appStateRepo)

	// Create adapters for platform-agnostic importer
	stravaAdapter := importer.NewServerStravaAdapter(stravaClient)
	storageAdapter := importer.NewServerStorageAdapter(
		athleteRepo,
		activityRepo,
		streamRepo,
		gearRepo,
		segmentRepo,
		bestEffortsRepo,
		photoRepo,
		maintenanceRepo,
		syncHistoryRepo,
		appStateRepo,
	)

	// Create importer
	imp, err := importer.New(stravaAdapter, storageAdapter)
	if err != nil {
		return fmt.Errorf("creating importer: %w", err)
	}

	// Wire up achievement detection
	achievementAdapter := importer.NewAchievementStorageAdapter(achievementRepo)
	imp.SetAchievementStorage(achievementAdapter)

	// Wire up notification history for deduplication
	notificationHistoryRepo := storage.NewNotificationHistoryRepository(db)
	notificationHistoryAdapter := importer.NewNotificationHistoryAdapter(notificationHistoryRepo)
	imp.SetNotificationHistory(notificationHistoryAdapter)

	// Create services for scheduler (maintenance checks, notifications)
	registry := services.NewServiceRegistry(db, logger)

	// Wire up import completion and achievement notifications.
	// Note: The passed ctx may already be canceled, so we create a fresh context.
	imp.SetNotificationHandler(func(_ context.Context, athleteID int64, stats importer.SyncStats) {
		notifyCtx, cancel := context.WithTimeout(context.Background(), notificationHandlerTimeout)
		defer cancel()

		// Send import complete notification
		if err := registry.NotificationService.NotifyImportComplete(notifyCtx, athleteID, services.ImportStats{
			ActivitiesImported: stats.ActivitiesImported,
			ActivitiesUpdated:  stats.ActivitiesUpdated,
			Duration:           stats.Duration,
		}); err != nil {
			slog.Warn("failed to send import notification", "error", err)
		}

		// Send achievement notifications (batched)
		if len(stats.Achievements) > 0 {
			achievements := make([]services.Achievement, len(stats.Achievements))
			for i, a := range stats.Achievements {
				achievements[i] = services.Achievement{
					Type:          string(a.Type),
					SubType:       a.SubType,
					Title:         a.Title,
					Value:         a.Description,
					PreviousValue: "",
				}
			}
			if err := registry.NotificationService.NotifyAchievements(notifyCtx, athleteID, achievements); err != nil {
				slog.Warn("failed to send achievement notification", "error", err)
			} else {
				// Only mark achievements as notified after successful notification.
				// This ensures transient failures don't permanently suppress achievements.
				if err := notificationHistoryAdapter.MarkNotifiedBatch(notifyCtx, athleteID, stats.Achievements); err != nil {
					slog.Warn("failed to mark achievements as notified", "error", err)
				}
			}
		}
	})

	// Create scheduler (periodic sync, maintenance checks, etc.)
	sched := scheduler.New(logger, stravaClient, settingsRepo, imp, registry.MaintenanceService, registry.NotificationService, statsRepo)
	schedulerCtx, schedulerCancel := context.WithCancel(context.Background())
	defer schedulerCancel()
	if err := sched.Start(schedulerCtx); err != nil {
		return fmt.Errorf("starting scheduler: %w", err)
	}

	// Create router
	router := api.NewRouter(cfg, stravaClient, db, imp)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Graceful shutdown
	done := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		<-sigCh

		slog.Info("shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.WriteTimeout)
		defer cancel()

		schedulerCancel()
		if err := sched.Stop(ctx); err != nil {
			slog.Warn("scheduler stop error", "error", err)
		}

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		close(done)
	}()

	slog.Info("starting server",
		"addr", addr,
		"external_url", cfg.Server.GetExternalURL(),
		"dev", cfg.Server.DevMode,
		"strava_configured", cfg.Strava.ClientID != "",
	)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	slog.Info("server stopped")
	return nil
}
