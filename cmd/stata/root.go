package main

import (
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stata",
		Short: "Statistics for Strava - Analytics dashboard for your activities",
		Long: `Stata is a self-hosted analytics dashboard for Strava athletes.
It provides comprehensive statistics, visualizations, and insights
into your training data.

Run 'stata serve' to start the web server, or use other commands
for data import and management.`,
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newServeCmd())
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newDemoCmd())

	return cmd
}
