package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var hashtagTokenRe = regexp.MustCompile(`#([A-Za-z0-9][A-Za-z0-9_-]{0,63})`)

func normalizeHashtag(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	s = strings.ToLower(s)
	return s
}

func extractHashtags(text string) []string {
	matches := hashtagTokenRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		tag := normalizeHashtag(m[1])
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func newCustomGearID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generating id: %w", err)
	}
	return "c_" + hex.EncodeToString(b[:]), nil
}

func (r *GearRepository) GetCustomByHashtag(ctx context.Context, athleteID int64, hashtag string) (*Gear, error) {
	tag := normalizeHashtag(hashtag)
	if tag == "" {
		return nil, nil
	}

	row := r.db.QueryRowContext(ctx, `
		SELECT id, athlete_id, name, is_primary, retired, distance,
			brand_name, model_name, description,
			COALESCE(source, ''), COALESCE(hashtag, ''), purchase_price, COALESCE(purchase_currency, ''),
			created_at, updated_at
		FROM gear
		WHERE athlete_id = ?
		  AND COALESCE(source, 'strava') = 'custom'
		  AND lower(COALESCE(hashtag, '')) = lower(?)
		LIMIT 1
	`, athleteID, tag)

	var g Gear
	err := row.Scan(
		&g.ID, &g.AthleteID, &g.Name, &g.Primary, &g.Retired, &g.Distance,
		&g.BrandName, &g.ModelName, &g.Description,
		&g.Source, &g.Hashtag, &g.PurchasePrice, &g.PurchaseCurrency,
		&g.CreatedAt, &g.UpdatedAt,
	)
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *GearRepository) ListCustom(ctx context.Context, athleteID int64, includeRetired bool) ([]Gear, error) {
	query := `
		SELECT id, athlete_id, name, is_primary, retired, distance,
			brand_name, model_name, description,
			COALESCE(source, ''), COALESCE(hashtag, ''), purchase_price, COALESCE(purchase_currency, ''),
			created_at, updated_at
		FROM gear
		WHERE athlete_id = ?
		  AND COALESCE(source, 'strava') = 'custom'
	`
	args := []any{athleteID}
	if !includeRetired {
		query += " AND retired = FALSE"
	}
	query += " ORDER BY name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []Gear
	for rows.Next() {
		var g Gear
		if err := rows.Scan(
			&g.ID, &g.AthleteID, &g.Name, &g.Primary, &g.Retired, &g.Distance,
			&g.BrandName, &g.ModelName, &g.Description,
			&g.Source, &g.Hashtag, &g.PurchasePrice, &g.PurchaseCurrency,
			&g.CreatedAt, &g.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

type CustomGearCreate struct {
	Name             string
	Hashtag          string
	Retired          bool
	PurchasePrice    *float64
	PurchaseCurrency string
}

func (r *GearRepository) CreateCustom(ctx context.Context, athleteID int64, in CustomGearCreate) (*Gear, error) {
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	tag := normalizeHashtag(in.Hashtag)
	if tag == "" {
		return nil, fmt.Errorf("hashtag is required")
	}

	id, err := newCustomGearID()
	if err != nil {
		return nil, err
	}

	g := &Gear{
		ID:               id,
		AthleteID:        athleteID,
		Name:             name,
		Primary:          false,
		Retired:          in.Retired,
		Distance:         0,
		Source:           "custom",
		Hashtag:          tag,
		PurchasePrice:    in.PurchasePrice,
		PurchaseCurrency: strings.TrimSpace(in.PurchaseCurrency),
	}

	if err := r.Upsert(ctx, g); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

type CustomGearUpdate struct {
	Name             *string
	Hashtag          *string
	Retired          *bool
	PurchasePrice    **float64
	PurchaseCurrency *string
}

func (r *GearRepository) UpdateCustom(ctx context.Context, athleteID int64, id string, in CustomGearUpdate) (*Gear, error) {
	g, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if g == nil || g.AthleteID != athleteID || strings.ToLower(g.Source) != "custom" {
		return nil, nil
	}

	if in.Name != nil {
		g.Name = strings.TrimSpace(*in.Name)
	}
	if in.Hashtag != nil {
		g.Hashtag = normalizeHashtag(*in.Hashtag)
	}
	if in.Retired != nil {
		g.Retired = *in.Retired
	}
	if in.PurchaseCurrency != nil {
		g.PurchaseCurrency = strings.TrimSpace(*in.PurchaseCurrency)
	}
	if in.PurchasePrice != nil {
		g.PurchasePrice = *in.PurchasePrice
	}

	g.Source = "custom"
	if err := r.Upsert(ctx, g); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *GearRepository) DeleteCustom(ctx context.Context, athleteID int64, id string, force bool) (hadActivities bool, err error) {
	var source string
	err = r.db.QueryRowContext(ctx, `SELECT COALESCE(source, 'strava') FROM gear WHERE id = ? AND athlete_id = ?`, id, athleteID).Scan(&source)
	if isNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if strings.ToLower(source) != "custom" {
		return false, fmt.Errorf("not custom gear")
	}

	var cnt int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM activities WHERE athlete_id = ? AND gear_id = ?`, athleteID, id).Scan(&cnt); err != nil {
		return false, err
	}
	if cnt > 0 {
		hadActivities = true
		if !force {
			return hadActivities, nil
		}
		if _, err := r.db.ExecContext(ctx, `UPDATE activities SET gear_id = '' WHERE athlete_id = ? AND gear_id = ?`, athleteID, id); err != nil {
			return hadActivities, err
		}
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM gear WHERE id = ? AND athlete_id = ? AND COALESCE(source, 'strava') = 'custom'`, id, athleteID)
	return hadActivities, err
}

func (r *GearRepository) ResolveCustomGearIDFromActivityName(ctx context.Context, athleteID int64, activityName string) (string, error) {
	tags := extractHashtags(activityName)
	if len(tags) == 0 {
		return "", nil
	}
	for _, tag := range tags {
		g, err := r.GetCustomByHashtag(ctx, athleteID, tag)
		if err != nil {
			return "", err
		}
		if g != nil {
			return g.ID, nil
		}
	}
	return "", nil
}
