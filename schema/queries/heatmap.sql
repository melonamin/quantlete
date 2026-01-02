-- Heatmap queries.

-- name: GetHeatmapActivities :many
-- Get activities with polylines for heatmap visualization.
SELECT
    id,
    name,
    sport_type,
    start_date,
    COALESCE(distance, 0) AS distance,
    summary_polyline,
    COALESCE(start_lat, 0) AS start_lat,
    COALESCE(start_lng, 0) AS start_lng
FROM activities
WHERE athlete_id = ?1
    AND summary_polyline IS NOT NULL
    AND summary_polyline != ''
ORDER BY start_date DESC;

-- name: GetHeatmapActivitiesBySport :many
-- Get heatmap activities filtered by sport type.
SELECT
    id,
    name,
    sport_type,
    start_date,
    COALESCE(distance, 0) AS distance,
    summary_polyline,
    COALESCE(start_lat, 0) AS start_lat,
    COALESCE(start_lng, 0) AS start_lng
FROM activities
WHERE athlete_id = ?1
    AND sport_type = ?2
    AND summary_polyline IS NOT NULL
    AND summary_polyline != ''
ORDER BY start_date DESC;

-- name: GetHeatmapActivitiesByDateRange :many
-- Get heatmap activities within a date range.
SELECT
    id,
    name,
    sport_type,
    start_date,
    COALESCE(distance, 0) AS distance,
    summary_polyline,
    COALESCE(start_lat, 0) AS start_lat,
    COALESCE(start_lng, 0) AS start_lng
FROM activities
WHERE athlete_id = ?1
    AND start_date >= ?2
    AND start_date <= ?3
    AND summary_polyline IS NOT NULL
    AND summary_polyline != ''
ORDER BY start_date DESC;

-- name: GetHeatmapActivitiesBySportAndDateRange :many
-- Get heatmap activities filtered by sport type and date range.
SELECT
    id,
    name,
    sport_type,
    start_date,
    COALESCE(distance, 0) AS distance,
    summary_polyline,
    COALESCE(start_lat, 0) AS start_lat,
    COALESCE(start_lng, 0) AS start_lng
FROM activities
WHERE athlete_id = ?1
    AND sport_type = ?2
    AND start_date >= ?3
    AND start_date <= ?4
    AND summary_polyline IS NOT NULL
    AND summary_polyline != ''
ORDER BY start_date DESC;

-- name: GetHeatmapCountries :many
-- Get countries for heatmap filtering.
SELECT
    COALESCE(location_country, '') AS country,
    COUNT(*) AS count
FROM activities
WHERE athlete_id = ?1
    AND COALESCE(location_country, '') != ''
    AND summary_polyline IS NOT NULL
    AND summary_polyline != ''
GROUP BY location_country
ORDER BY count DESC, country ASC;
