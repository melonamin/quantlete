package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Photo struct {
	ID           string          `json:"id"`
	AthleteID    int64           `json:"athlete_id"`
	ActivityID   int64           `json:"activity_id"`
	URL          string          `json:"url"`
	ThumbnailURL string          `json:"thumbnail_url,omitempty"`
	Caption      string          `json:"caption,omitempty"`
	Location     json.RawMessage `json:"location,omitempty"`
	CreatedAt    SQLiteTime      `json:"created_at"`
}

type PhotoRepository struct {
	db *DB
}

func NewPhotoRepository(db *DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) Upsert(ctx context.Context, p *Photo) error {
	// Handle location JSON - pass nil for empty location to avoid DuckDB JSON type issues
	var location any
	if len(p.Location) > 0 {
		location = string(p.Location)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO photos (id, athlete_id, activity_id, url, thumbnail_url, caption, location, created_at)
		VALUES (?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)
		ON CONFLICT (id) DO UPDATE SET
			athlete_id = EXCLUDED.athlete_id,
			activity_id = EXCLUDED.activity_id,
			url = EXCLUDED.url,
			thumbnail_url = EXCLUDED.thumbnail_url,
			caption = EXCLUDED.caption,
			location = EXCLUDED.location
	`, p.ID, p.AthleteID, p.ActivityID, p.URL, p.ThumbnailURL, p.Caption, location, time.Now())
	return err
}

func (r *PhotoRepository) ListByActivity(ctx context.Context, athleteID int64, activityID int64) ([]Photo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, athlete_id, activity_id, url, COALESCE(thumbnail_url, ''), COALESCE(caption, ''), COALESCE(location, ''), created_at
		FROM photos
		WHERE athlete_id = ? AND activity_id = ?
		ORDER BY created_at ASC, id ASC
	`, athleteID, activityID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Photo
	for rows.Next() {
		var p Photo
		var loc string
		if err := rows.Scan(&p.ID, &p.AthleteID, &p.ActivityID, &p.URL, &p.ThumbnailURL, &p.Caption, &loc, &p.CreatedAt); err != nil {
			return nil, err
		}
		if strings.TrimSpace(loc) != "" {
			p.Location = json.RawMessage(loc)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type PhotoListFilters struct {
	SportTypes []string
	Country    string
}

type PhotoListItem struct {
	Photo
	ActivityName    string     `json:"activity_name"`
	SportType       string     `json:"sport_type"`
	StartDateLocal  SQLiteTime `json:"start_date_local"`
	LocationCountry string     `json:"location_country,omitempty"`
}

type PhotoFacetCount struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

type PhotoListResult struct {
	Items      []PhotoListItem   `json:"items"`
	Total      int               `json:"total"`
	Countries  []PhotoFacetCount `json:"countries"`
	SportTypes []PhotoFacetCount `json:"sport_types"`
}

func (r *PhotoRepository) List(ctx context.Context, athleteID int64, f PhotoListFilters, page, perPage int) (*PhotoListResult, error) {
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 || perPage > 200 {
		perPage = 60
	}
	offset := (page - 1) * perPage

	base := `
		FROM photos p
		JOIN activities a ON a.id = p.activity_id
		WHERE p.athlete_id = ?
	`
	args := []any{athleteID}

	if len(f.SportTypes) > 0 {
		placeholders := make([]string, 0, len(f.SportTypes))
		for _, s := range f.SportTypes {
			placeholders = append(placeholders, "?")
			args = append(args, s)
		}
		base += fmt.Sprintf(" AND a.sport_type IN (%s)", strings.Join(placeholders, ","))
	}
	if strings.TrimSpace(f.Country) != "" {
		base += " AND COALESCE(a.location_country, '') = ?"
		args = append(args, f.Country)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) "+base, args...).Scan(&total); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			p.id, p.athlete_id, p.activity_id, p.url, COALESCE(p.thumbnail_url, ''), COALESCE(p.caption, ''), COALESCE(p.location, ''), p.created_at,
			a.name, a.sport_type, a.start_date_local, COALESCE(a.location_country, '')
	`+base+`
		ORDER BY a.start_date_local DESC, p.created_at DESC
		LIMIT ? OFFSET ?
	`, append(args, perPage, offset)...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	items := make([]PhotoListItem, 0, perPage)
	for rows.Next() {
		var it PhotoListItem
		var loc string
		if err := rows.Scan(
			&it.ID, &it.AthleteID, &it.ActivityID, &it.URL, &it.ThumbnailURL, &it.Caption, &loc, &it.CreatedAt,
			&it.ActivityName, &it.SportType, &it.StartDateLocal, &it.LocationCountry,
		); err != nil {
			return nil, err
		}
		if strings.TrimSpace(loc) != "" {
			it.Location = json.RawMessage(loc)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	countries, err := r.facets(ctx, athleteID, "COALESCE(a.location_country, '')", "COALESCE(a.location_country, '') != ''")
	if err != nil {
		return nil, err
	}
	sportTypes, err := r.facets(ctx, athleteID, "a.sport_type", "a.sport_type != ''")
	if err != nil {
		return nil, err
	}

	return &PhotoListResult{
		Items:      items,
		Total:      total,
		Countries:  countries,
		SportTypes: sportTypes,
	}, nil
}

func (r *PhotoRepository) facets(ctx context.Context, athleteID int64, expr string, where string) ([]PhotoFacetCount, error) {
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT %s AS value, COUNT(*) AS count
		FROM photos p
		JOIN activities a ON a.id = p.activity_id
		WHERE p.athlete_id = ? AND %s
		GROUP BY value
		ORDER BY count DESC, value ASC
	`, expr, where), athleteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []PhotoFacetCount
	for rows.Next() {
		var c PhotoFacetCount
		if err := rows.Scan(&c.Value, &c.Count); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PhotoRepository) CountForActivity(ctx context.Context, activityID int64) (int, error) {
	var cnt int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM photos WHERE activity_id = ?`, activityID).Scan(&cnt)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	return cnt, nil
}
