-- Calendar queries.

-- name: GetCalendarData :many
-- Get calendar day data for a year.
SELECT
    date,
    activity_count,
    total_distance,
    total_time,
    total_calories,
    total_intensity
FROM v_calendar_days
WHERE athlete_id = ?1
    AND date >= ?2 || '-01-01'
    AND date <= ?2 || '-12-31'
ORDER BY date ASC;

-- name: GetCalendarDataRange :many
-- Get calendar day data for a date range (for rolling 365 view).
SELECT
    date,
    activity_count,
    total_distance,
    total_time,
    total_calories,
    total_intensity
FROM v_calendar_days
WHERE athlete_id = ?1
    AND date >= ?2
    AND date <= ?3
ORDER BY date ASC;

-- name: GetCalendarActivities :many
-- Get activities for a specific month.
SELECT
    id,
    name,
    sport_type,
    start_date,
    start_date_local,
    distance,
    moving_time,
    total_elevation_gain
FROM activities
WHERE athlete_id = ?1
    AND strftime('%Y', start_date) = ?2
    AND strftime('%m', start_date) = ?3
ORDER BY start_date ASC;

-- name: GetCalendarSummary :one
-- Get summary for a specific month.
SELECT
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(calories), 0) AS total_calories,
    COALESCE(SUM(CASE WHEN workout_type IS NOT NULL AND workout_type != 0 THEN 1 ELSE 0 END), 0) AS workout_count
FROM activities
WHERE athlete_id = ?1
    AND strftime('%Y', start_date) = ?2
    AND strftime('%m', start_date) = ?3;
