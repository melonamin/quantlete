-- Best effort queries.

-- name: GetBestEffortPRs :many
-- Get personal records for each distance type (rn = 1 means fastest).
SELECT
    distance_type,
    name,
    distance_m,
    elapsed_time,
    moving_time,
    pr_rank,
    activity_id,
    activity_name,
    sport_type,
    start_date_local
FROM v_best_effort_rankings
WHERE athlete_id = ?1 AND rn = 1
ORDER BY distance_m ASC;

-- name: GetBestEffortPRsBySport :many
-- Get personal records filtered by sport type.
SELECT
    distance_type,
    name,
    distance_m,
    elapsed_time,
    moving_time,
    pr_rank,
    activity_id,
    activity_name,
    sport_type,
    start_date_local
FROM v_best_effort_rankings
WHERE athlete_id = ?1 AND sport_type = ?2 AND rn = 1
ORDER BY distance_m ASC;

-- name: GetBestEffortsForDistance :many
-- Get all efforts for a specific distance type, ranked by time.
SELECT
    distance_type,
    name,
    distance_m,
    elapsed_time,
    moving_time,
    pr_rank,
    activity_id,
    activity_name,
    sport_type,
    start_date_local,
    rn
FROM v_best_effort_rankings
WHERE athlete_id = ?1 AND distance_type = ?2
ORDER BY elapsed_time ASC
LIMIT 100;
