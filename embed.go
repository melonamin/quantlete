// Package quantlete provides the embedded web assets for the Quantlete application.
package quantlete

import "embed"

// WebAssets contains the built React application.
// This is populated during build by go:embed.
//
//go:embed web/dist/*
var WebAssets embed.FS
