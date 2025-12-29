-- Dashboard queries.

-- name: GetDashboardStats :one
-- Get overall athlete statistics for dashboard.
SELECT
    COUNT(*) AS total_activities,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
WHERE athlete_id = ?1;

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
