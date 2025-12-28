package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Gear represents a stored gear record.
type Gear struct {
	ID          string
	AthleteID   int64
	Name        string
	Primary     bool
	Retired     bool
	Distance    float64
	BrandName   string
	ModelName   string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// GearRepository handles gear persistence.
type GearRepository struct {
	db *DB
}

// NewGearRepository creates a new gear repository.
func NewGearRepository(db *DB) *GearRepository {
	return &GearRepository{db: db}
}

// Upsert inserts or updates gear.
// Uses DELETE + INSERT because DuckDB doesn't allow updating indexed columns in UPSERT.
func (r *GearRepository) Upsert(ctx context.Context, g *Gear) error {
	// Delete existing gear
	_, err := r.db.Exec("DELETE FROM gear WHERE id = ?", g.ID)
	if err != nil {
		return fmt.Errorf("deleting existing gear: %w", err)
	}

	// Insert gear
	_, err = r.db.Exec(`
		INSERT INTO gear (
			id, athlete_id, name, is_primary, retired, distance,
			brand_name, model_name, description, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		g.ID, g.AthleteID, g.Name, g.Primary, g.Retired, g.Distance,
		g.BrandName, g.ModelName, g.Description, time.Now(), time.Now(),
	)
	return err
}

// GetByID retrieves gear by ID.
func (r *GearRepository) GetByID(ctx context.Context, id string) (*Gear, error) {
	row := r.db.QueryRow(`
		SELECT id, athlete_id, name, is_primary, retired, distance,
			brand_name, model_name, description, created_at, updated_at
		FROM gear WHERE id = ?
	`, id)

	var g Gear
	err := row.Scan(
		&g.ID, &g.AthleteID, &g.Name, &g.Primary, &g.Retired, &g.Distance,
		&g.BrandName, &g.ModelName, &g.Description, &g.CreatedAt, &g.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning gear: %w", err)
	}

	return &g, nil
}

// GetByAthleteID retrieves all gear for an athlete.
func (r *GearRepository) GetByAthleteID(ctx context.Context, athleteID int64, includeRetired bool) ([]Gear, error) {
	query := `
		SELECT id, athlete_id, name, is_primary, retired, distance,
			brand_name, model_name, description, created_at, updated_at
		FROM gear
		WHERE athlete_id = ?
	`
	if !includeRetired {
		query += " AND retired = FALSE"
	}
	query += " ORDER BY is_primary DESC, name ASC"

	rows, err := r.db.Query(query, athleteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var gear []Gear
	for rows.Next() {
		var g Gear
		if err := rows.Scan(
			&g.ID, &g.AthleteID, &g.Name, &g.Primary, &g.Retired, &g.Distance,
			&g.BrandName, &g.ModelName, &g.Description, &g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning gear: %w", err)
		}
		gear = append(gear, g)
	}

	return gear, rows.Err()
}

// GetTotalDistance returns the total distance for a piece of gear.
func (r *GearRepository) GetTotalDistance(ctx context.Context, gearID string) (float64, error) {
	var distance float64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(distance), 0)
		FROM activities
		WHERE gear_id = ?
	`, gearID).Scan(&distance)
	return distance, err
}

// GetActivityCount returns the number of activities for a piece of gear.
func (r *GearRepository) GetActivityCount(ctx context.Context, gearID string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM activities
		WHERE gear_id = ?
	`, gearID).Scan(&count)
	return count, err
}
