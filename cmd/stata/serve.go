package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/sasha/stata/internal/api"
	"github.com/sasha/stata/internal/config"
	"github.com/sasha/stata/internal/importer"
	"github.com/sasha/stata/internal/storage"
	"github.com/sasha/stata/internal/strava"
)

func newServeCmd() *cobra.Command {
	var port int
	var dev bool

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the web server",
		Long: `Start the Stata web server to serve the dashboard.

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
	db, err := storage.Open(cfg.Storage.DataDir)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = db.Close() }()

	slog.Info("database opened", "path", db.Path())

	// Run migrations
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Create Strava client
	stravaClient := strava.NewClient(&cfg.Strava)

	// Create repositories
	activityRepo := storage.NewActivityRepository(db)
	athleteRepo := storage.NewAthleteRepository(db)
	tokenRepo := storage.NewTokenRepository(db)
	gearRepo := storage.NewGearRepository(db)
	streamRepo := storage.NewStreamRepository(db)

	// Restore tokens from database
	if err := restoreAuth(context.Background(), stravaClient, tokenRepo, athleteRepo); err != nil {
		slog.Warn("failed to restore auth from database", "error", err)
	}

	// Create importer
	imp := importer.New(stravaClient, activityRepo, athleteRepo, tokenRepo, gearRepo, streamRepo)

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

		if err := server.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", "error", err)
		}
		close(done)
	}()

	slog.Info("starting server",
		"addr", addr,
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

// restoreAuth restores authentication from the database on startup.
func restoreAuth(
	ctx context.Context,
	stravaClient *strava.Client,
	tokenRepo *storage.TokenRepository,
	athleteRepo *storage.AthleteRepository,
) error {
	// Get active tokens (not expired)
	tokens, err := tokenRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("getting active tokens: %w", err)
	}

	slog.Info("checking for stored tokens", "active_count", len(tokens))

	if len(tokens) == 0 {
		slog.Info("no active tokens found in database")
		return nil
	}

	// Use the first active token (single-user app)
	storedToken := tokens[0]

	// Get the athlete
	storedAthlete, err := athleteRepo.GetByID(ctx, storedToken.AthleteID)
	if err != nil {
		return fmt.Errorf("getting athlete: %w", err)
	}
	if storedAthlete == nil {
		return fmt.Errorf("athlete not found for token")
	}

	// Convert to oauth2.Token
	token := &oauth2.Token{
		AccessToken:  storedToken.AccessToken,
		RefreshToken: storedToken.RefreshToken,
		TokenType:    storedToken.TokenType,
		Expiry:       storedToken.ExpiresAt,
	}

	// Convert to strava.Athlete
	athlete := &strava.Athlete{
		ID:            storedAthlete.ID,
		Username:      storedAthlete.Username,
		FirstName:     storedAthlete.FirstName,
		LastName:      storedAthlete.LastName,
		City:          storedAthlete.City,
		State:         storedAthlete.State,
		Country:       storedAthlete.Country,
		Sex:           storedAthlete.Sex,
		Premium:       storedAthlete.Premium,
		Summit:        storedAthlete.Summit,
		ProfileMedium: storedAthlete.ProfileMedium,
		Profile:       storedAthlete.Profile,
		Weight:        storedAthlete.Weight,
	}

	stravaClient.SetToken(token, athlete)
	slog.Info("restored auth from database", "athlete_id", athlete.ID, "expires_at", token.Expiry)

	return nil
}
