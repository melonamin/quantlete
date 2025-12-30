package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/melonamin/quantlete/internal/config"
	"github.com/melonamin/quantlete/internal/demo"
	"github.com/melonamin/quantlete/internal/storage"
)

func newDemoCmd() *cobra.Command {
	var activities int
	var months int
	var athleteName string

	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Generate demo data for the dashboard",
		Long: `Generate realistic fake data for demonstrating the Quantlete dashboard.

This command creates a sample athlete with activities, gear, segments,
and training load data. It wipes any existing data before generating.

Examples:
  quantlete demo                              # Generate 100 activities over 12 months
  quantlete demo --activities=200 --months=24 # Generate 200 activities over 2 years
  quantlete demo --athlete="Jane Doe"         # Custom athlete name`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runDemo(activities, months, athleteName)
		},
	}

	cmd.Flags().IntVar(&activities, "activities", 100, "Number of activities to generate")
	cmd.Flags().IntVar(&months, "months", 12, "Number of months to spread activities across")
	cmd.Flags().StringVar(&athleteName, "athlete", "Demo User", "Name of the demo athlete")

	return cmd
}

func runDemo(numActivities, months int, athleteName string) error {
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

	// Open database
	db, err := storage.Open(cfg.Storage.DataDir, cfg.Storage.DBFile)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Run migrations
	if err := db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Create demo generator
	gen := demo.NewGenerator(db)

	// Configure options
	opts := demo.Options{
		NumActivities: numActivities,
		Months:        months,
		AthleteName:   athleteName,
	}

	slog.Info("Generating demo data",
		"activities", numActivities,
		"months", months,
		"athlete", athleteName,
	)

	// Generate demo data
	ctx := context.Background()
	if err := gen.Generate(ctx, opts); err != nil {
		return fmt.Errorf("generating demo data: %w", err)
	}

	slog.Info("Demo data generated successfully")
	return nil
}
