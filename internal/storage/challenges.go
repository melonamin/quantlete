package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Challenge struct {
	ID             string     `json:"id"`
	AthleteID      int64      `json:"athlete_id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug,omitempty"`
	BadgeURL       string     `json:"badge_url,omitempty"`
	CompletionDate *SQLiteTime `json:"completion_date,omitempty"`
	Month          string     `json:"month,omitempty"` // YYYY-MM
	CreatedAt            SQLiteTime  `json:"created_at"`
}

type ChallengeRepository struct {
	db *DB
}

func NewChallengeRepository(db *DB) *ChallengeRepository {
	return &ChallengeRepository{db: db}
}

func (r *ChallengeRepository) Upsert(ctx context.Context, c *Challenge) error {
	if c.ID == "" {
		if c.Slug != "" {
			c.ID = c.Slug
		} else {
			c.ID = uuid.NewString()
		}
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO challenges (id, athlete_id, name, slug, badge_url, completion_date, month, created_at)
		VALUES (?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, NULLIF(?, ''), ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			badge_url = EXCLUDED.badge_url,
			completion_date = EXCLUDED.completion_date,
			month = EXCLUDED.month
	`, c.ID, c.AthleteID, c.Name, c.Slug, c.BadgeURL, c.CompletionDate, c.Month, SQLiteTime{Time: time.Now()})
	return err
}

func (r *ChallengeRepository) List(ctx context.Context, athleteID int64, month string) ([]Challenge, error) {
	query := `
		SELECT id, athlete_id, name, COALESCE(slug, ''), COALESCE(badge_url, ''), completion_date, COALESCE(month, ''), created_at
		FROM challenges
		WHERE athlete_id = ?
	`
	args := []any{athleteID}
	if month != "" {
		query += " AND month = ?"
		args = append(args, month)
	}
	query += " ORDER BY month DESC, completion_date DESC NULLS LAST, name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Challenge
	for rows.Next() {
		var c Challenge
		if err := rows.Scan(&c.ID, &c.AthleteID, &c.Name, &c.Slug, &c.BadgeURL, &c.CompletionDate, &c.Month, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ChallengeRepository) Count(ctx context.Context, athleteID int64) (int, error) {
	var cnt int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM challenges WHERE athlete_id = ?`, athleteID).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("counting challenges: %w", err)
	}
	return cnt, nil
}
