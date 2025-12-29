-- Segment queries.

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
