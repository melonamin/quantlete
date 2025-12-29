-- Power analysis queries.

-- name: GetPowerBests :many
-- Get best power outputs for specified durations.
SELECT
    duration_s,
    best_avg_watts AS watts,
    activity_id,
    start_date,
    activity_name
FROM v_power_best_rankings
WHERE athlete_id = ?1 AND rn = 1
ORDER BY duration_s ASC;

-- name: GetPowerBestsForDurations :many
-- Get best power for specific duration values.
SELECT
    duration_s,
    best_avg_watts AS watts,
    activity_id,
    start_date,
    activity_name
FROM v_power_best_rankings
WHERE athlete_id = ?1
    AND duration_s IN (/*SLICE:durations*/?2)
    AND rn = 1
ORDER BY duration_s ASC;

-- name: GetPowerHistory :many
-- Get power history for a specific duration (for charts).
SELECT
    duration_s,
    best_avg_watts AS watts,
    activity_id,
    start_date,
    activity_name,
    rn
FROM v_power_best_rankings
WHERE athlete_id = ?1 AND duration_s = ?2
ORDER BY start_date DESC
LIMIT 100;
