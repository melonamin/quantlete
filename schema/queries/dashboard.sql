-- Dashboard queries.

-- name: GetDashboardStats :one
-- Get overall athlete statistics for dashboard.
SELECT
    COUNT(*) AS total_activities,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation,
    COALESCE(SUM(calories), 0) AS total_calories
FROM activities
WHERE athlete_id = ?1;

-- name: GetDashboardStatsFromDate :one
-- Get athlete statistics from a specific date.
SELECT
    COUNT(*) AS total_activities,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
WHERE athlete_id = ?1 AND DATE(start_date_local) >= ?2;

-- name: GetStatsBySportType :many
-- Get statistics grouped by sport type.
SELECT
    sport_type,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
WHERE athlete_id = ?1
GROUP BY sport_type
ORDER BY activity_count DESC;

-- name: GetStatsBySportTypeFromDate :many
-- Get statistics grouped by sport type from a specific date.
SELECT
    sport_type,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
WHERE athlete_id = ?1 AND DATE(start_date_local) >= ?2
GROUP BY sport_type
ORDER BY total_distance DESC;

-- name: GetWeeklyStats :many
-- Get weekly statistics for the last 12 weeks.
SELECT
    week,
    week_start,
    activity_count,
    total_distance,
    total_time,
    total_elevation
FROM v_weekly_stats
WHERE athlete_id = ?1
ORDER BY week DESC
LIMIT 12;

-- name: GetRecentActivities :many
-- Get recent activities for dashboard.
SELECT
    id,
    name,
    sport_type,
    start_date,
    start_date_local,
    distance,
    moving_time,
    total_elevation_gain,
    average_speed,
    max_speed
FROM activities
WHERE athlete_id = ?1
ORDER BY start_date DESC
LIMIT ?2;

-- name: GetSportTypeStats :many
-- Get statistics by sport type.
SELECT
    sport_type,
    activity_count,
    total_distance,
    total_time,
    total_elevation
FROM v_sport_type_stats
WHERE athlete_id = ?1
ORDER BY activity_count DESC;

-- name: GetMonthlyStats :many
-- Get monthly statistics for a year.
SELECT
    month,
    sport_type,
    activity_count,
    total_distance,
    total_time,
    total_elevation
FROM v_monthly_stats
WHERE athlete_id = ?1 AND month LIKE ?2 || '%'
ORDER BY month DESC;

-- name: GetYearlyStats :many
-- Get yearly statistics.
SELECT
    year,
    activity_count,
    total_distance,
    total_time,
    total_elevation
FROM v_yearly_stats
WHERE athlete_id = ?1
ORDER BY year DESC;

-- name: GetDashboardConfig :one
-- Get dashboard config for an athlete.
SELECT config
FROM dashboard_config
WHERE athlete_id = ?1;

-- name: UpsertDashboardConfig :exec
-- Insert or update dashboard config.
INSERT INTO dashboard_config (athlete_id, config, updated_at)
VALUES (?1, ?2, ?3)
ON CONFLICT (athlete_id) DO UPDATE SET
    config = EXCLUDED.config,
    updated_at = EXCLUDED.updated_at;

-- name: GetTrainingGoalsConfig :one
-- Get training goals config for an athlete.
SELECT config
FROM training_goals
WHERE athlete_id = ?1;

-- name: UpsertTrainingGoalsConfig :exec
-- Insert or update training goals config.
INSERT INTO training_goals (athlete_id, config, updated_at)
VALUES (?1, ?2, ?3)
ON CONFLICT (athlete_id) DO UPDATE SET
    config = EXCLUDED.config,
    updated_at = EXCLUDED.updated_at;

-- name: GetWeeklyTrends :many
-- Get weekly trends for rolling N weeks with optional sport type filter.
-- Uses ISO week numbering for consistent week boundaries.
SELECT
    strftime('%Y-W%W', start_date_local) AS week,
    strftime('%Y-%m-%d', start_date_local, 'weekday 0', '-6 days') AS week_start,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
WHERE athlete_id = ?1
    AND DATE(start_date_local) >= DATE('now', '-' || ?2 || ' days')
    AND (?3 = '' OR sport_type = ?3)
GROUP BY strftime('%Y-W%W', start_date_local)
ORDER BY week ASC;

-- name: GetMonthlyComparison :many
-- Get monthly statistics grouped by year and month for cross-year comparison.
-- Returns data for specified years (or all years if year list is empty) with optional sport type filter.
-- Each row contains year, month (1-12), and aggregated metrics.
SELECT
    CAST(strftime('%Y', start_date_local) AS INTEGER) AS year,
    CAST(strftime('%m', start_date_local) AS INTEGER) AS month,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
WHERE athlete_id = ?1
    AND (?2 = '' OR sport_type = ?2)
GROUP BY year, month
ORDER BY year ASC, month ASC;
