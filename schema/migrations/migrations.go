package migrations

import "embed"

// FS contains the embedded SQLite migration SQL files.
//
//go:embed *.sql
var FS embed.FS
