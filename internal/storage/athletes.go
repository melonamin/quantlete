package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Athlete represents a stored athlete record.
type Athlete struct {
	ID            int64
	Username      string
	FirstName     string
	LastName      string
	City          string
	State         string
	Country       string
	Sex           string
	Premium       bool
	Summit        bool
	ProfileMedium string
	Profile       string
	Weight        float64
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// AuthToken represents stored OAuth tokens.
type AuthToken struct {
	AthleteID    int64
	AccessToken  string
	RefreshToken string
	TokenType    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AthleteRepository handles athlete persistence.
type AthleteRepository struct {
	db *DB
}

// NewAthleteRepository creates a new athlete repository.
func NewAthleteRepository(db *DB) *AthleteRepository {
	return &AthleteRepository{db: db}
}

// Upsert inserts or updates an athlete.
func (r *AthleteRepository) Upsert(ctx context.Context, a *Athlete) error {
	_, err := r.db.Exec(`
		INSERT INTO athletes (
			id, username, firstname, lastname, city, state, country,
			sex, premium, summit, profile_medium, profile, weight,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			username = EXCLUDED.username,
			firstname = EXCLUDED.firstname,
			lastname = EXCLUDED.lastname,
			city = EXCLUDED.city,
			state = EXCLUDED.state,
			country = EXCLUDED.country,
			sex = EXCLUDED.sex,
			premium = EXCLUDED.premium,
			summit = EXCLUDED.summit,
			profile_medium = EXCLUDED.profile_medium,
			profile = EXCLUDED.profile,
			weight = EXCLUDED.weight,
			updated_at = EXCLUDED.updated_at
	`,
		a.ID, a.Username, a.FirstName, a.LastName, a.City, a.State, a.Country,
		a.Sex, a.Premium, a.Summit, a.ProfileMedium, a.Profile, a.Weight,
		time.Now(), time.Now(),
	)
	return err
}

// GetByID retrieves an athlete by ID.
func (r *AthleteRepository) GetByID(ctx context.Context, id int64) (*Athlete, error) {
	row := r.db.QueryRow(`
		SELECT id, username, firstname, lastname, city, state, country,
			sex, premium, summit, profile_medium, profile, weight,
			created_at, updated_at
		FROM athletes WHERE id = ?
	`, id)

	var a Athlete
	err := row.Scan(
		&a.ID, &a.Username, &a.FirstName, &a.LastName, &a.City, &a.State, &a.Country,
		&a.Sex, &a.Premium, &a.Summit, &a.ProfileMedium, &a.Profile, &a.Weight,
		&a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning athlete: %w", err)
	}

	return &a, nil
}

// GetAll retrieves all athletes.
func (r *AthleteRepository) GetAll(ctx context.Context) ([]Athlete, error) {
	rows, err := r.db.Query(`
		SELECT id, username, firstname, lastname, city, state, country,
			sex, premium, summit, profile_medium, profile, weight,
			created_at, updated_at
		FROM athletes
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var athletes []Athlete
	for rows.Next() {
		var a Athlete
		if err := rows.Scan(
			&a.ID, &a.Username, &a.FirstName, &a.LastName, &a.City, &a.State, &a.Country,
			&a.Sex, &a.Premium, &a.Summit, &a.ProfileMedium, &a.Profile, &a.Weight,
			&a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning athlete: %w", err)
		}
		athletes = append(athletes, a)
	}

	return athletes, rows.Err()
}

// TokenRepository handles OAuth token persistence.
type TokenRepository struct {
	db *DB
}

// NewTokenRepository creates a new token repository.
func NewTokenRepository(db *DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Upsert inserts or updates an auth token.
// Uses DELETE + INSERT because DuckDB doesn't allow updating indexed columns in UPSERT.
func (r *TokenRepository) Upsert(ctx context.Context, t *AuthToken) error {
	// Delete existing token for this athlete
	_, err := r.db.Exec("DELETE FROM auth_tokens WHERE athlete_id = ?", t.AthleteID)
	if err != nil {
		return fmt.Errorf("deleting existing token: %w", err)
	}

	// Insert new token
	_, err = r.db.Exec(`
		INSERT INTO auth_tokens (
			athlete_id, access_token, refresh_token, token_type,
			expires_at, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		t.AthleteID, t.AccessToken, t.RefreshToken, t.TokenType,
		t.ExpiresAt, time.Now(), time.Now(),
	)
	return err
}

// GetByAthleteID retrieves a token by athlete ID.
func (r *TokenRepository) GetByAthleteID(ctx context.Context, athleteID int64) (*AuthToken, error) {
	row := r.db.QueryRow(`
		SELECT athlete_id, access_token, refresh_token, token_type,
			expires_at, created_at, updated_at
		FROM auth_tokens WHERE athlete_id = ?
	`, athleteID)

	var t AuthToken
	err := row.Scan(
		&t.AthleteID, &t.AccessToken, &t.RefreshToken, &t.TokenType,
		&t.ExpiresAt, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning token: %w", err)
	}

	return &t, nil
}

// GetActive retrieves tokens that haven't expired yet.
func (r *TokenRepository) GetActive(ctx context.Context) ([]AuthToken, error) {
	rows, err := r.db.Query(`
		SELECT athlete_id, access_token, refresh_token, token_type,
			expires_at, created_at, updated_at
		FROM auth_tokens
		WHERE expires_at > ?
		ORDER BY athlete_id
	`, time.Now())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tokens []AuthToken
	for rows.Next() {
		var t AuthToken
		if err := rows.Scan(
			&t.AthleteID, &t.AccessToken, &t.RefreshToken, &t.TokenType,
			&t.ExpiresAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning token: %w", err)
		}
		tokens = append(tokens, t)
	}

	return tokens, rows.Err()
}

// Delete removes a token by athlete ID.
func (r *TokenRepository) Delete(ctx context.Context, athleteID int64) error {
	_, err := r.db.Exec("DELETE FROM auth_tokens WHERE athlete_id = ?", athleteID)
	return err
}
