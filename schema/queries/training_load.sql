-- Training load queries.

-- name: GetDailyTSS :many
-- Get daily TSS values for a date range.
SELECT day, tss
FROM v_daily_tss
WHERE athlete_id = ?1 AND day >= ?2 AND day <= ?3
ORDER BY day ASC;

-- name: GetAllDailyTSS :many
-- Get all daily TSS values for an athlete.
SELECT day, tss
FROM v_daily_tss
WHERE athlete_id = ?1
ORDER BY day ASC;

-- name: GetActivityTSS :one
-- Get TSS for a specific activity.
SELECT tss, normalized_power, intensity_factor
FROM activity_training_load
WHERE activity_id = ?1;
