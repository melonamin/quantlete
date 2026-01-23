-- Rewind queries.

-- name: GetRewindYears :many
-- Get available years for rewind.
SELECT DISTINCT strftime('%Y', start_date_local) AS year
FROM activities
WHERE athlete_id = ?1
ORDER BY year DESC;

-- name: GetRewindDateRange :one
-- Get min/max activity dates for rewind.
SELECT
    MIN(date(start_date_local)) AS min_day,
    MAX(date(start_date_local)) AS max_day
FROM activities
WHERE athlete_id = ?1;

-- name: GetRewindTotals :one
-- Get totals for a date range.
SELECT
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS distance_m,
    COALESCE(SUM(total_elevation_gain), 0) AS elevation_m,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(kudos_count), 0) AS kudos_count,
    COALESCE(SUM(CASE WHEN commute = 1 AND sport_type LIKE '%Ride%' THEN distance ELSE 0 END), 0) AS commute_distance_m
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3;

-- name: GetRewindActiveDays :one
-- Count active days for a date range.
SELECT COUNT(DISTINCT date(start_date_local)) AS count
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3;

-- name: GetRewindMonths :many
-- Get monthly totals for a date range.
SELECT
    strftime('%Y-%m', start_date_local) AS month,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS distance_m,
    COALESCE(SUM(total_elevation_gain), 0) AS elevation_m
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
GROUP BY month
ORDER BY month ASC;

-- name: GetRewindPRsByMonth :many
-- Count new PRs by month (best so far).
WITH w AS (
    SELECT
        be.distance_type,
        a.start_date_local AS dt,
        be.elapsed_time,
        MIN(be.elapsed_time) OVER (
            PARTITION BY be.distance_type
            ORDER BY a.start_date_local ASC
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) AS best_so_far
    FROM best_efforts be
    JOIN activities a
        ON a.id = be.activity_id AND a.athlete_id = be.athlete_id
    WHERE be.athlete_id = ?1
        AND date(a.start_date_local) < ?3
),
 m AS (
    SELECT
        dt,
        best_so_far,
        LAG(best_so_far) OVER (PARTITION BY distance_type ORDER BY dt ASC) AS prev_best
    FROM w
 )
SELECT
    strftime('%Y-%m', dt) AS month,
    COUNT(*) AS count
FROM m
WHERE date(dt) >= ?2
    AND (prev_best IS NULL OR best_so_far < prev_best)
GROUP BY month
ORDER BY month ASC;

-- name: GetRewindMovingTimeBySport :many
-- Get moving time totals by sport.
SELECT
    sport_type,
    COALESCE(SUM(moving_time), 0) AS moving_time
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
GROUP BY sport_type
ORDER BY moving_time DESC;

-- name: GetRewindStartTimesByHour :many
-- Get start times by hour.
SELECT
    CAST(strftime('%H', start_date_local) AS INTEGER) AS hour,
    COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
GROUP BY hour
ORDER BY hour ASC;

-- name: GetRewindLocations :many
-- Get activity start locations buckets.
SELECT
    COALESCE(ROUND(start_lat, 2), 0) AS start_lat,
    COALESCE(ROUND(start_lng, 2), 0) AS start_lng,
    COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1
    AND start_lat IS NOT NULL
    AND start_lng IS NOT NULL
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
GROUP BY start_lat, start_lng
ORDER BY count DESC
LIMIT 2000;

-- name: GetRewindActiveDayList :many
-- Get active day list for streak calculations.
SELECT DISTINCT date(start_date_local) AS day
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
ORDER BY day ASC;

-- name: GetRewindBiggestDistance :one
-- Get biggest activity by distance.
SELECT
    id,
    name,
    sport_type,
    start_date_local,
    distance
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
ORDER BY distance DESC
LIMIT 1;

-- name: GetRewindBiggestElevation :one
-- Get biggest activity by elevation gain.
SELECT
    id,
    name,
    sport_type,
    start_date_local,
    total_elevation_gain
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
ORDER BY total_elevation_gain DESC
LIMIT 1;

-- name: GetRewindBiggestDuration :one
-- Get biggest activity by moving time.
SELECT
    id,
    name,
    sport_type,
    start_date_local,
    moving_time
FROM activities
WHERE athlete_id = ?1
    AND date(start_date_local) >= ?2
    AND date(start_date_local) < ?3
ORDER BY moving_time DESC
LIMIT 1;

-- name: GetRewindRandomPhoto :one
-- Get a random photo in the date range.
SELECT
    p.id,
    p.activity_id,
    p.url,
    COALESCE(p.thumbnail_url, '') AS thumbnail_url,
    COALESCE(p.caption, '') AS caption
FROM photos p
JOIN activities a ON a.id = p.activity_id AND a.athlete_id = ?1
WHERE date(a.start_date_local) >= ?2
    AND date(a.start_date_local) < ?3
ORDER BY random()
LIMIT 1;
