package migrations

import "embed"

// FS contains the embedded DuckDB migration SQL files.
//
//go:embed *.sql
var FS embed.FS
