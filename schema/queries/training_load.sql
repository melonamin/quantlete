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

-- name: GetActivityTSSByAthlete :one
-- Get TSS for a specific activity by athlete.
SELECT COALESCE(tss, 0) AS tss
FROM activity_training_load
WHERE athlete_id = ?1 AND activity_id = ?2;

-- name: GetTrainingLoadSummary :one
-- Get the latest daily training load point for an athlete.
SELECT day, tss, ctl, atl, tsb
FROM daily_training_load
WHERE athlete_id = ?1
ORDER BY day DESC
LIMIT 1;

-- name: UpsertActivityTrainingLoad :exec
-- Insert or update activity training load.
INSERT INTO activity_training_load (activity_id, athlete_id, sport_type, method, ftp_used, normalized_power, intensity_factor, tss, computed_at)
VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9)
ON CONFLICT (activity_id) DO UPDATE SET
    method = EXCLUDED.method,
    ftp_used = EXCLUDED.ftp_used,
    normalized_power = EXCLUDED.normalized_power,
    intensity_factor = EXCLUDED.intensity_factor,
    tss = EXCLUDED.tss,
    computed_at = EXCLUDED.computed_at;

-- name: UpsertDailyTrainingLoad :exec
-- Insert or update daily training load.
INSERT INTO daily_training_load (athlete_id, day, tss, ctl, atl, tsb)
VALUES (?1, ?2, ?3, ?4, ?5, ?6)
ON CONFLICT (athlete_id, day) DO UPDATE SET
    tss = EXCLUDED.tss,
    ctl = EXCLUDED.ctl,
    atl = EXCLUDED.atl,
    tsb = EXCLUDED.tsb;
