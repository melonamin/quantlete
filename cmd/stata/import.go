package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Import activities from Strava",
		Long: `Import activities from Strava API into the local database.

This command fetches all activities from your Strava account and stores
them locally for analysis. It respects Strava's rate limits and can be
interrupted and resumed.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println("Import not yet implemented. Please configure Strava OAuth first.")
			return nil
		},
	}
}
