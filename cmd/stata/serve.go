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

	"github.com/sasha/stata/internal/api"
	"github.com/sasha/stata/internal/config"
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

	// Create router
	router := api.NewRouter(cfg, stravaClient, db)

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
