-- Heatmap queries.

-- name: GetHeatmapActivities :many
-- Get activities with polylines for heatmap visualization.
SELECT
    id,
    sport_type,
    summary_polyline,
    start_lat,
    start_lng
FROM v_heatmap_activities
WHERE athlete_id = ?1
ORDER BY start_date DESC;

-- name: GetHeatmapActivitiesBySport :many
-- Get heatmap activities filtered by sport type.
SELECT
    id,
    sport_type,
    summary_polyline,
    start_lat,
    start_lng
FROM v_heatmap_activities
WHERE athlete_id = ?1 AND sport_type = ?2
ORDER BY start_date DESC;

-- name: GetHeatmapActivitiesByDateRange :many
-- Get heatmap activities within a date range.
SELECT
    id,
    sport_type,
    summary_polyline,
    start_lat,
    start_lng
FROM v_heatmap_activities
WHERE athlete_id = ?1
    AND start_date >= ?2
    AND start_date <= ?3
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
