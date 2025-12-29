package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Gear represents a stored gear record.
type Gear struct {
	ID               string
	AthleteID        int64
	Name             string
	Primary          bool
	Retired          bool
	Distance         float64
	BrandName        string
	ModelName        string
	Description      string
	Source           string
	Hashtag          string
	PurchasePrice    *float64
	PurchaseCurrency string
	CreatedAt        SQLiteTime
	UpdatedAt        SQLiteTime
}

// GearRepository handles gear persistence.
type GearRepository struct {
	db *DB
}

// NewGearRepository creates a new gear repository.
func NewGearRepository(db *DB) *GearRepository {
	return &GearRepository{db: db}
}

// Upsert inserts or updates gear using INSERT ON CONFLICT to avoid race conditions.
func (r *GearRepository) Upsert(ctx context.Context, g *Gear) error {
	now := SQLiteTime{Time: time.Now()}

	source := g.Source
	if source == "" {
		source = "strava"
	}
	var hashtag any
	if g.Hashtag != "" {
		hashtag = g.Hashtag
	}
	var currency any
	if g.PurchaseCurrency != "" {
		currency = g.PurchaseCurrency
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gear (
			id, athlete_id, name, is_primary, retired, distance,
			brand_name, model_name, description,
			source, hashtag, purchase_price, purchase_currency,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			athlete_id = excluded.athlete_id,
			name = excluded.name,
			is_primary = excluded.is_primary,
			retired = excluded.retired,
			distance = excluded.distance,
			brand_name = excluded.brand_name,
			model_name = excluded.model_name,
			description = excluded.description,
			source = COALESCE(NULLIF(excluded.source, ''), gear.source),
			hashtag = COALESCE(excluded.hashtag, gear.hashtag),
			purchase_price = COALESCE(excluded.purchase_price, gear.purchase_price),
			purchase_currency = COALESCE(NULLIF(excluded.purchase_currency, ''), gear.purchase_currency),
			updated_at = excluded.updated_at
	`,
		g.ID,
		g.AthleteID,
		g.Name,
		g.Primary,
		g.Retired,
		g.Distance,
		g.BrandName,
		g.ModelName,
		g.Description,
		source,
		hashtag,
		g.PurchasePrice,
		currency,
		now,
		now,
	)
	return err
}

// GetByID retrieves gear by ID.
func (r *GearRepository) GetByID(ctx context.Context, id string) (*Gear, error) {
	row := r.db.QueryRow(`
		SELECT id, athlete_id, name, is_primary, retired, distance,
			brand_name, model_name, description,
			COALESCE(source, ''), COALESCE(hashtag, ''), purchase_price, COALESCE(purchase_currency, ''),
			created_at, updated_at
		FROM gear WHERE id = ?
	`, id)

	var g Gear
	err := row.Scan(
		&g.ID, &g.AthleteID, &g.Name, &g.Primary, &g.Retired, &g.Distance,
		&g.BrandName, &g.ModelName, &g.Description,
		&g.Source, &g.Hashtag, &g.PurchasePrice, &g.PurchaseCurrency,
		&g.CreatedAt, &g.UpdatedAt,
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
			brand_name, model_name, description,
			COALESCE(source, ''), COALESCE(hashtag, ''), purchase_price, COALESCE(purchase_currency, ''),
			created_at, updated_at
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
			&g.BrandName, &g.ModelName, &g.Description,
			&g.Source, &g.Hashtag, &g.PurchasePrice, &g.PurchaseCurrency,
			&g.CreatedAt, &g.UpdatedAt,
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
