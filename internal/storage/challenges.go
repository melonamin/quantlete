package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/melonamin/quantlete/internal/pagination"
)

type Challenge struct {
	ID             string      `json:"id"`
	AthleteID      int64       `json:"athlete_id"`
	Name           string      `json:"name"`
	Slug           string      `json:"slug,omitempty"`
	BadgeURL       string      `json:"badge_url,omitempty"`
	LocalBadgeURL  string      `json:"local_badge_url,omitempty"` // Local path to downloaded badge
	CompletionDate *SQLiteTime `json:"completion_date,omitempty"`
	Month          string      `json:"month,omitempty"` // YYYY-MM
	CreatedAt      SQLiteTime  `json:"created_at"`
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
		INSERT INTO challenges (id, athlete_id, name, slug, badge_url, local_badge_url, completion_date, month, created_at)
		VALUES (?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), ?, NULLIF(?, ''), ?)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			slug = EXCLUDED.slug,
			badge_url = EXCLUDED.badge_url,
			local_badge_url = COALESCE(NULLIF(EXCLUDED.local_badge_url, ''), challenges.local_badge_url),
			completion_date = EXCLUDED.completion_date,
			month = EXCLUDED.month
	`, c.ID, c.AthleteID, c.Name, c.Slug, c.BadgeURL, c.LocalBadgeURL, c.CompletionDate, c.Month, SQLiteTime{Time: time.Now()})
	return err
}

func (r *ChallengeRepository) List(ctx context.Context, athleteID int64, month string) ([]Challenge, error) {
	query := `
		SELECT id, athlete_id, name, COALESCE(slug, ''), COALESCE(badge_url, ''), COALESCE(local_badge_url, ''), completion_date, COALESCE(month, ''), created_at
		FROM challenges
		WHERE athlete_id = ?
	`
	args := []any{athleteID}
	if month != "" {
		query += " AND month = ?"
		args = append(args, month)
	}
	query += " ORDER BY month DESC, completion_date IS NULL, completion_date DESC, name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Challenge
	for rows.Next() {
		var c Challenge
		if err := rows.Scan(&c.ID, &c.AthleteID, &c.Name, &c.Slug, &c.BadgeURL, &c.LocalBadgeURL, &c.CompletionDate, &c.Month, &c.CreatedAt); err != nil {
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

// ChallengeFilters contains pagination and filter parameters.
type ChallengeFilters struct {
	Month string
	pagination.QueryParams
}

// ChallengeListResult contains paginated challenge results.
type ChallengeListResult struct {
	Items      []Challenge
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

// ListPaginated returns a paginated list of challenges for an athlete.
func (r *ChallengeRepository) ListPaginated(ctx context.Context, athleteID int64, f ChallengeFilters) (ChallengeListResult, error) {
	// Build WHERE clause
	where := "athlete_id = ?"
	args := []any{athleteID}
	if f.Month != "" {
		where += " AND month = ?"
		args = append(args, f.Month)
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM challenges WHERE " + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil && err != sql.ErrNoRows {
		return ChallengeListResult{}, fmt.Errorf("counting challenges: %w", err)
	}

	// Normalize pagination params
	p := pagination.NewParams(f.Page, f.PerPage)

	// Build ORDER BY with validation
	validOrderBy := map[string]string{
		"name":            "name",
		"completion_date": "completion_date",
		"month":           "month",
	}
	orderBy := pagination.BuildOrderClause(f.OrderBy, f.OrderDir, validOrderBy, "month DESC, completion_date IS NULL, completion_date DESC, name ASC")
	query := fmt.Sprintf(`
		SELECT id, athlete_id, name, COALESCE(slug, ''), COALESCE(badge_url, ''), COALESCE(local_badge_url, ''), completion_date, COALESCE(month, ''), created_at
		FROM challenges
		WHERE %s
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, where, orderBy)

	args = append(args, p.PerPage, p.Offset())
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return ChallengeListResult{}, err
	}
	defer func() { _ = rows.Close() }()

	var items []Challenge
	for rows.Next() {
		var c Challenge
		if err := rows.Scan(&c.ID, &c.AthleteID, &c.Name, &c.Slug, &c.BadgeURL, &c.LocalBadgeURL, &c.CompletionDate, &c.Month, &c.CreatedAt); err != nil {
			return ChallengeListResult{}, err
		}
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return ChallengeListResult{}, err
	}

	return ChallengeListResult{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: p.TotalPages(total),
	}, nil
}
