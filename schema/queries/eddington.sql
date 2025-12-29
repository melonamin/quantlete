-- Eddington number queries.

-- name: GetEddingtonDays :many
-- Get daily distances for Eddington calculation.
-- Results are sorted by distance descending for the algorithm.
SELECT date, distance_km
FROM v_daily_distances
WHERE athlete_id = ?1 AND distance_km > 0
ORDER BY distance_km DESC;

-- name: GetEddingtonDaysBySport :many
-- Get daily distances filtered by sport types.
SELECT
    DATE(start_date) AS date,
    SUM(distance) / 1000.0 AS distance_km
FROM activities
WHERE athlete_id = ?1
    AND sport_type IN (/*SLICE:sport_types*/?2)
GROUP BY DATE(start_date)
HAVING distance_km > 0
ORDER BY distance_km DESC;

-- name: GetEddingtonDaysChronological :many
-- Get daily distances in chronological order (for history calculation).
SELECT date, distance_km
FROM v_daily_distances
WHERE athlete_id = ?1 AND distance_km > 0
ORDER BY date ASC;
