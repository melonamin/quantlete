package storage

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"

	"github.com/melonamin/quantlete/schema/migrations"
)

var migrationsFS = migrations.FS

// Migration represents a single database migration.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// Migrate runs all pending migrations.
func (db *DB) Migrate() error {
	// Create migrations tracking table
	if err := db.createMigrationsTable(); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	// Load all migrations
	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	// Get applied migrations
	applied, err := db.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("getting applied migrations: %w", err)
	}

	// Run pending migrations
	for _, m := range migrations {
		if applied[m.Version] {
			continue
		}

		slog.Info("applying migration", "version", m.Version, "name", m.Name)

		if err := db.runMigration(m); err != nil {
			return fmt.Errorf("applying migration %d (%s): %w", m.Version, m.Name, err)
		}

		slog.Info("migration applied", "version", m.Version)
	}

	return nil
}

// createMigrationsTable creates the migrations tracking table if it doesn't exist.
func (db *DB) createMigrationsTable() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

// getAppliedMigrations returns a set of applied migration versions.
func (db *DB) getAppliedMigrations() (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM _migrations")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// runMigration executes a single migration.
func (db *DB) runMigration(m Migration) error {
	// Execute migration SQL
	if _, err := db.Exec(m.SQL); err != nil {
		return err
	}

	// Record migration as applied
	_, err := db.Exec(
		"INSERT INTO _migrations (version, name) VALUES (?, ?)",
		m.Version, m.Name,
	)
	return err
}

// loadMigrations loads all migration files from the embedded filesystem.
func loadMigrations() ([]Migration, error) {
	var migrations []Migration

	err := fs.WalkDir(migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".sql") {
			return nil
		}

		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		filename := filepath.Base(path)
		version, name, err := parseMigrationFilename(filename)
		if err != nil {
			return fmt.Errorf("parsing filename %s: %w", filename, err)
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			SQL:     string(content),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseMigrationFilename extracts version and name from a migration filename.
// Expected format: 001_initial.sql
func parseMigrationFilename(filename string) (int, string, error) {
	name := strings.TrimSuffix(filename, ".sql")
	parts := strings.SplitN(name, "_", 2)

	if len(parts) != 2 {
		return 0, "", fmt.Errorf("invalid migration filename format: %s", filename)
	}

	var version int
	if _, err := fmt.Sscanf(parts[0], "%d", &version); err != nil {
		return 0, "", fmt.Errorf("invalid version number: %s", parts[0])
	}

	return version, parts[1], nil
}

// MigrationStatus returns the current migration status.
type MigrationStatus struct {
	CurrentVersion int
	LatestVersion  int
	Pending        int
}

// Status returns the current migration status.
func (db *DB) MigrationStatus() (MigrationStatus, error) {
	migrations, err := loadMigrations()
	if err != nil {
		return MigrationStatus{}, err
	}

	applied, err := db.getAppliedMigrations()
	if err != nil {
		return MigrationStatus{}, err
	}

	var current int
	for v := range applied {
		if v > current {
			current = v
		}
	}

	var latest int
	if len(migrations) > 0 {
		latest = migrations[len(migrations)-1].Version
	}

	pending := 0
	for _, m := range migrations {
		if !applied[m.Version] {
			pending++
		}
	}

	return MigrationStatus{
		CurrentVersion: current,
		LatestVersion:  latest,
		Pending:        pending,
	}, nil
}
