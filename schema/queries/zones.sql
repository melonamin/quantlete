-- Zone distribution queries.

-- name: UpsertActivityZoneDistribution :exec
-- Insert or update activity zone distribution.
INSERT INTO activity_zone_distributions (
    activity_id, zone_def_id, seconds_z1, seconds_z2, seconds_z3,
    seconds_z4, seconds_z5, total_seconds, computed_at
) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)
ON CONFLICT(activity_id) DO UPDATE SET
    zone_def_id = EXCLUDED.zone_def_id,
    seconds_z1 = EXCLUDED.seconds_z1,
    seconds_z2 = EXCLUDED.seconds_z2,
    seconds_z3 = EXCLUDED.seconds_z3,
    seconds_z4 = EXCLUDED.seconds_z4,
    seconds_z5 = EXCLUDED.seconds_z5,
    total_seconds = EXCLUDED.total_seconds,
    computed_at = EXCLUDED.computed_at;

-- name: GetWeeklyZoneDistribution :many
-- Returns weekly zone distribution for the last N weeks.
SELECT
    strftime('%Y-W%W', a.start_date_local) AS week,
    COALESCE(SUM(zd.seconds_z1), 0) AS z1,
    COALESCE(SUM(zd.seconds_z2), 0) AS z2,
    COALESCE(SUM(zd.seconds_z3), 0) AS z3,
    COALESCE(SUM(zd.seconds_z4), 0) AS z4,
    COALESCE(SUM(zd.seconds_z5), 0) AS z5,
    COALESCE(SUM(zd.total_seconds), 0) AS total
FROM activity_zone_distributions zd
JOIN activities a ON a.id = zd.activity_id
WHERE a.athlete_id = ?1
  AND a.start_date_local >= date('now', '-' || ?2 || ' weeks')
GROUP BY week
ORDER BY week;

-- name: GetActivityIDsWithoutZoneDistribution :many
-- Find activities with HR streams but no zone distribution.
SELECT a.id
FROM activities a
JOIN activity_streams s ON s.activity_id = a.id AND s.type = 'heartrate'
LEFT JOIN activity_zone_distributions zd ON zd.activity_id = a.id
WHERE a.athlete_id = ?1 AND zd.activity_id IS NULL;

-- name: DeleteZoneDistributionsForSportType :exec
-- Delete zone distributions for activities matching sport type and date range.
-- Used when zone definitions change to trigger recomputation.
DELETE FROM activity_zone_distributions
WHERE activity_id IN (
    SELECT a.id FROM activities a
    WHERE a.athlete_id = ?1
    AND a.sport_type = ?2
    AND a.start_date_local >= ?3
);

-- name: GetActivityZoneDistribution :one
-- Get zone distribution for a specific activity.
SELECT
    activity_id, zone_def_id, seconds_z1, seconds_z2, seconds_z3,
    seconds_z4, seconds_z5, total_seconds, computed_at
FROM activity_zone_distributions
WHERE activity_id = ?1;

-- name: GetActivityForZoneComputation :one
-- Get minimal activity data needed for zone computation.
SELECT id, sport_type, start_date
FROM activities
WHERE id = ?1;
