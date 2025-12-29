-- Distribution queries.

-- name: GetDaytimeDistribution :many
-- Get activity distribution by hour of day.
SELECT
    CAST(strftime('%H', start_date_local) AS INTEGER) AS hour,
    COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1
GROUP BY hour
ORDER BY hour;

-- name: GetWeekdayDistribution :many
-- Get activity distribution by day of week.
-- Sunday = 0, Saturday = 6
SELECT
    CAST(strftime('%w', start_date_local) AS INTEGER) AS weekday,
    COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1
GROUP BY weekday
ORDER BY weekday;

-- name: GetMonthlyDistribution :many
-- Get activity distribution by month.
SELECT
    CAST(strftime('%m', start_date_local) AS INTEGER) AS month,
    COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1
GROUP BY month
ORDER BY month;
