package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information - set via ldflags during build
var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("stata %s\n", version)
			fmt.Printf("  commit:     %s\n", commit)
			fmt.Printf("  built:      %s\n", buildDate)
			fmt.Printf("  go version: %s\n", runtime.Version())
			fmt.Printf("  platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
