-- Segment queries.

-- name: UpsertSegment :exec
-- Insert or update a segment.
INSERT INTO segments (
    id, name, activity_type, distance, average_grade, maximum_grade,
    elevation_high, elevation_low, climb_category,
    start_lat, start_lng, end_lat, end_lng,
    starred, polyline,
    athlete_kom_rank, athlete_effort_count, athlete_pr_elapsed_time, athlete_pr_date,
    updated_at
) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13, ?14, ?15, ?16, ?17, ?18, ?19, ?20)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    activity_type = EXCLUDED.activity_type,
    distance = EXCLUDED.distance,
    average_grade = EXCLUDED.average_grade,
    maximum_grade = EXCLUDED.maximum_grade,
    elevation_high = EXCLUDED.elevation_high,
    elevation_low = EXCLUDED.elevation_low,
    climb_category = EXCLUDED.climb_category,
    start_lat = EXCLUDED.start_lat,
    start_lng = EXCLUDED.start_lng,
    end_lat = EXCLUDED.end_lat,
    end_lng = EXCLUDED.end_lng,
    starred = EXCLUDED.starred,
    polyline = EXCLUDED.polyline,
    athlete_kom_rank = EXCLUDED.athlete_kom_rank,
    athlete_effort_count = EXCLUDED.athlete_effort_count,
    athlete_pr_elapsed_time = EXCLUDED.athlete_pr_elapsed_time,
    athlete_pr_date = EXCLUDED.athlete_pr_date,
    updated_at = EXCLUDED.updated_at;

-- name: UpsertSegmentEffort :exec
-- Insert or update a segment effort.
INSERT INTO segment_efforts (
    id, segment_id, activity_id, athlete_id,
    name, elapsed_time, moving_time,
    start_date, start_date_local,
    distance, average_watts, average_heartrate, max_heartrate,
    pr_rank, country
) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13, ?14, ?15)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    elapsed_time = EXCLUDED.elapsed_time,
    moving_time = EXCLUDED.moving_time,
    start_date = EXCLUDED.start_date,
    start_date_local = EXCLUDED.start_date_local,
    distance = EXCLUDED.distance,
    average_watts = EXCLUDED.average_watts,
    average_heartrate = EXCLUDED.average_heartrate,
    max_heartrate = EXCLUDED.max_heartrate,
    pr_rank = EXCLUDED.pr_rank,
    country = EXCLUDED.country;

-- name: GetSegmentByID :one
-- Get segment by ID.
SELECT
    id, name, activity_type, distance, average_grade, maximum_grade,
    elevation_high, elevation_low, climb_category,
    start_lat, start_lng, end_lat, end_lng,
    starred, COALESCE(polyline, '') AS polyline,
    athlete_kom_rank, athlete_effort_count, athlete_pr_elapsed_time, athlete_pr_date,
    created_at, updated_at
FROM segments
WHERE id = ?1;

-- name: CountSegmentsByAthlete :one
-- Count unique segments for an athlete.
SELECT COUNT(DISTINCT s.id) AS count
FROM segments s
JOIN segment_efforts e ON e.segment_id = s.id
WHERE e.athlete_id = ?1;

-- name: GetSegmentEffortCountries :many
-- Get distinct countries from segment efforts.
SELECT DISTINCT COALESCE(country, '') AS country
FROM segment_efforts
WHERE athlete_id = ?1 AND COALESCE(country, '') != ''
ORDER BY country ASC;

-- name: GetSegmentCountryStats :many
-- Get segment counts by country.
SELECT COALESCE(country, '') AS country, COUNT(DISTINCT segment_id) AS count
FROM segment_efforts
WHERE athlete_id = ?1 AND COALESCE(country, '') != ''
GROUP BY country
ORDER BY count DESC, country ASC;

-- name: GetSegments :many
-- Get all segments for an athlete with effort statistics.
SELECT
    id,
    name,
    activity_type,
    distance,
    average_grade,
    maximum_grade,
    elevation_high,
    elevation_low,
    climb_category,
    starred,
    athlete_kom_rank,
    athlete_effort_count,
    athlete_pr_elapsed_time,
    athlete_pr_date,
    times_completed,
    last_effort_date,
    best_elapsed_time
FROM v_segment_stats
WHERE athlete_id = ?1
ORDER BY name;

-- name: GetSegmentsByCountry :many
-- Get segments filtered by country.
SELECT DISTINCT
    s.id,
    s.name,
    s.activity_type,
    s.distance,
    s.average_grade,
    s.maximum_grade,
    s.elevation_high,
    s.elevation_low,
    s.climb_category,
    s.starred,
    s.athlete_kom_rank,
    s.athlete_effort_count,
    s.athlete_pr_elapsed_time,
    s.athlete_pr_date,
    vs.times_completed,
    vs.last_effort_date,
    vs.best_elapsed_time
FROM segments s
JOIN segment_efforts se ON se.segment_id = s.id
LEFT JOIN v_segment_stats vs ON vs.id = s.id AND vs.athlete_id = se.athlete_id
WHERE se.athlete_id = ?1 AND se.country = ?2
ORDER BY s.name;

-- name: GetSegmentCountries :many
-- Get segment countries with counts.
SELECT country, segment_count
FROM v_segment_countries
WHERE athlete_id = ?1
ORDER BY segment_count DESC;

-- name: GetSegmentDetail :one
-- Get detailed segment information.
SELECT
    id,
    name,
    activity_type,
    distance,
    average_grade,
    maximum_grade,
    elevation_high,
    elevation_low,
    climb_category,
    start_lat,
    start_lng,
    end_lat,
    end_lng,
    starred,
    polyline,
    athlete_kom_rank,
    athlete_effort_count,
    athlete_pr_elapsed_time,
    athlete_pr_date
FROM segments
WHERE id = ?1;

-- name: GetSegmentEfforts :many
-- Get all efforts for a segment by an athlete.
SELECT
    id,
    segment_id,
    activity_id,
    athlete_id,
    name,
    elapsed_time,
    moving_time,
    start_date,
    start_date_local,
    distance,
    average_watts,
    average_heartrate,
    max_heartrate,
    pr_rank
FROM segment_efforts
WHERE segment_id = ?1 AND athlete_id = ?2
ORDER BY elapsed_time ASC;
