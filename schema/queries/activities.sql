-- Activity queries.

-- name: GetActivities :many
-- Get paginated activities.
SELECT
    id,
    athlete_id,
    name,
    sport_type,
    start_date,
    start_date_local,
    timezone,
    distance,
    moving_time,
    elapsed_time,
    total_elevation_gain,
    average_speed,
    max_speed,
    average_heartrate,
    max_heartrate,
    average_watts,
    max_watts,
    weighted_average_watts,
    kilojoules,
    average_cadence,
    calories,
    suffer_score,
    gear_id,
    commute,
    workout_type,
    location_city,
    location_state,
    location_country,
    summary_polyline,
    start_lat,
    start_lng
FROM activities
WHERE athlete_id = ?1
ORDER BY start_date DESC
LIMIT ?2 OFFSET ?3;

-- name: GetActivitiesBySport :many
-- Get activities filtered by sport type.
SELECT
    id,
    athlete_id,
    name,
    sport_type,
    start_date,
    start_date_local,
    timezone,
    distance,
    moving_time,
    elapsed_time,
    total_elevation_gain,
    average_speed,
    max_speed,
    average_heartrate,
    max_heartrate,
    average_watts,
    max_watts,
    weighted_average_watts,
    kilojoules,
    average_cadence,
    calories,
    suffer_score,
    gear_id,
    commute,
    workout_type,
    location_city,
    location_state,
    location_country,
    summary_polyline,
    start_lat,
    start_lng
FROM activities
WHERE athlete_id = ?1 AND sport_type = ?2
ORDER BY start_date DESC
LIMIT ?3 OFFSET ?4;

-- name: GetActivity :one
-- Get a single activity by ID.
SELECT
    id,
    athlete_id,
    name,
    sport_type,
    start_date,
    start_date_local,
    timezone,
    distance,
    moving_time,
    elapsed_time,
    total_elevation_gain,
    average_speed,
    max_speed,
    average_heartrate,
    max_heartrate,
    average_watts,
    max_watts,
    weighted_average_watts,
    kilojoules,
    average_cadence,
    calories,
    suffer_score,
    gear_id,
    commute,
    workout_type,
    location_city,
    location_state,
    location_country,
    summary_polyline,
    start_lat,
    start_lng,
    description,
    device_name,
    embed_token,
    trainer,
    private
FROM activities
WHERE id = ?1 AND athlete_id = ?2;

-- name: CountActivities :one
-- Count total activities for an athlete.
SELECT COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1;

-- name: CountActivitiesBySport :one
-- Count activities by sport type.
SELECT COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1 AND sport_type = ?2;

-- name: GetActivitiesByDateRange :many
-- Get activities within a date range.
SELECT
    id,
    athlete_id,
    name,
    sport_type,
    start_date,
    start_date_local,
    timezone,
    distance,
    moving_time,
    elapsed_time,
    total_elevation_gain,
    average_speed,
    max_speed,
    average_heartrate,
    max_heartrate,
    average_watts,
    max_watts,
    weighted_average_watts,
    kilojoules,
    average_cadence,
    calories,
    suffer_score,
    gear_id,
    commute,
    workout_type,
    location_city,
    location_state,
    location_country,
    summary_polyline,
    start_lat,
    start_lng
FROM activities
WHERE athlete_id = ?1 AND start_date >= ?2 AND start_date <= ?3
ORDER BY start_date DESC
LIMIT ?4 OFFSET ?5;

-- name: GetActivitiesBySportAndDateRange :many
-- Get activities filtered by sport type and date range.
SELECT
    id,
    athlete_id,
    name,
    sport_type,
    start_date,
    start_date_local,
    timezone,
    distance,
    moving_time,
    elapsed_time,
    total_elevation_gain,
    average_speed,
    max_speed,
    average_heartrate,
    max_heartrate,
    average_watts,
    max_watts,
    weighted_average_watts,
    kilojoules,
    average_cadence,
    calories,
    suffer_score,
    gear_id,
    commute,
    workout_type,
    location_city,
    location_state,
    location_country,
    summary_polyline,
    start_lat,
    start_lng
FROM activities
WHERE athlete_id = ?1 AND sport_type = ?2 AND start_date >= ?3 AND start_date <= ?4
ORDER BY start_date DESC
LIMIT ?5 OFFSET ?6;

-- name: CountActivitiesByDateRange :one
-- Count activities within a date range.
SELECT COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1 AND start_date >= ?2 AND start_date <= ?3;

-- name: CountActivitiesBySportAndDateRange :one
-- Count activities by sport type and date range.
SELECT COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1 AND sport_type = ?2 AND start_date >= ?3 AND start_date <= ?4;

-- name: GetActivityStreams :many
-- Get activity streams.
SELECT
    stream_type,
    data,
    series_type,
    original_size,
    resolution
FROM activity_streams
WHERE activity_id = ?1;
