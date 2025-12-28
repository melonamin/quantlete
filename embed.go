// Package stata provides the embedded web assets for the Stata application.
package stata

import "embed"

// WebAssets contains the built React application.
// This is populated during build by go:embed.
//
//go:embed web/dist/*
var WebAssets embed.FS
