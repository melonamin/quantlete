package storage

import (
	"context"
)

type GearMonthlyUsage struct {
	Month            string   `json:"month"` // YYYY-MM
	GearID           string   `json:"gear_id"`
	GearName         string   `json:"gear_name"`
	Source           string   `json:"source"`
	Hashtag          string   `json:"hashtag,omitempty"`
	Retired          bool     `json:"retired"`
	PurchasePrice    *float64 `json:"purchase_price,omitempty"`
	PurchaseCurrency string   `json:"purchase_currency,omitempty"`
	ActivityCount    int      `json:"activity_count"`
	Distance         float64  `json:"distance"`
	MovingTime       int      `json:"moving_time"`
}

func (r *GearRepository) GetMonthlyUsage(ctx context.Context, athleteID int64, includeRetired bool) ([]GearMonthlyUsage, error) {
	query := `
		SELECT
			strftime(a.start_date, '%Y-%m') AS month,
			a.gear_id AS gear_id,
			COALESCE(g.name, '') AS gear_name,
			COALESCE(g.source, 'strava') AS source,
			COALESCE(g.hashtag, '') AS hashtag,
			COALESCE(g.retired, FALSE) AS retired,
			g.purchase_price AS purchase_price,
			COALESCE(g.purchase_currency, '') AS purchase_currency,
			COUNT(*) AS activity_count,
			COALESCE(SUM(a.distance), 0) AS distance,
			COALESCE(SUM(a.moving_time), 0) AS moving_time
		FROM activities a
		LEFT JOIN gear g ON g.id = a.gear_id
		WHERE a.athlete_id = ?
		  AND COALESCE(a.gear_id, '') != ''
	`
	args := []any{athleteID}
	if !includeRetired {
		query += " AND COALESCE(g.retired, FALSE) = FALSE"
	}
	query += `
		GROUP BY month, gear_id, gear_name, source, hashtag, retired, purchase_price, purchase_currency
		ORDER BY month ASC, gear_name ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []GearMonthlyUsage
	for rows.Next() {
		var s GearMonthlyUsage
		if err := rows.Scan(
			&s.Month,
			&s.GearID,
			&s.GearName,
			&s.Source,
			&s.Hashtag,
			&s.Retired,
			&s.PurchasePrice,
			&s.PurchaseCurrency,
			&s.ActivityCount,
			&s.Distance,
			&s.MovingTime,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
