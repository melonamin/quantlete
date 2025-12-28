package storage

import (
	"context"
	"math"
	"time"
)

type EddingtonHistoryPoint struct {
	Date   string `json:"date"`   // YYYY-MM-DD
	Number int    `json:"number"` // Eddington number achieved on/after this date
}

// GetEddingtonHistory returns milestone points where the Eddington number increases.
func (r *StatsRepository) GetEddingtonHistory(ctx context.Context, athleteID int64, sportTypes []string) ([]EddingtonHistoryPoint, error) {
	query := `
		SELECT
			strftime('%Y-%m-%d', start_date_local) AS day,
			SUM(distance) / 1000.0 AS distance_km
		FROM activities
		WHERE athlete_id = ?
	`
	args := []any{athleteID}
	if len(sportTypes) > 0 {
		placeholders := make([]string, len(sportTypes))
		for i, st := range sportTypes {
			placeholders[i] = "?"
			args = append(args, st)
		}
		query += " AND sport_type IN (" + JoinStrings(placeholders, ",") + ")"
	}
	query += `
		GROUP BY day
		HAVING distance_km > 0
		ORDER BY day ASC
	`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	type dayRow struct {
		Day time.Time
		KM  float64
	}
	var days []dayRow
	for rows.Next() {
		var d dayRow
		var dayStr string
		if err := rows.Scan(&dayStr, &d.KM); err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			continue
		}
		d.Day = t
		days = append(days, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// countsGE[k] = number of days with distance_km >= k (k in whole km).
	countsGE := make([]int, 1) // index 0 unused
	currentE := 0
	var out []EddingtonHistoryPoint

	for _, d := range days {
		maxK := int(math.Floor(d.KM + 1e-9))
		if maxK <= 0 {
			continue
		}
		if len(countsGE) <= maxK {
			next := make([]int, maxK+1)
			copy(next, countsGE)
			countsGE = next
		}
		for k := 1; k <= maxK; k++ {
			countsGE[k]++
		}

		for currentE+1 < len(countsGE) && countsGE[currentE+1] >= currentE+1 {
			currentE++
			out = append(out, EddingtonHistoryPoint{
				Date:   d.Day.Format("2006-01-02"),
				Number: currentE,
			})
		}
	}

	return out, nil
}
