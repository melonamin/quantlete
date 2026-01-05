-- Notification history queries.

-- name: HasBeenNotified :one
-- Check if an achievement has already been notified.
SELECT COUNT(*) as count
FROM notification_history
WHERE athlete_id = ?1 AND achievement_type = ?2 AND achievement_key = ?3;

-- name: MarkNotified :exec
-- Record that an achievement has been notified.
INSERT INTO notification_history (athlete_id, achievement_type, achievement_key, notified_at)
VALUES (?1, ?2, ?3, ?4)
ON CONFLICT(athlete_id, achievement_type, achievement_key) DO NOTHING;

-- name: GetNotifiedKeys :many
-- Get all notified achievement keys for an athlete.
-- Used for batch filtering - retrieve all, then filter in Go.
SELECT achievement_type, achievement_key
FROM notification_history
WHERE athlete_id = ?1;

-- name: GetNotifiedKeysByTypes :many
-- Get notified achievement keys for specific types.
SELECT achievement_type, achievement_key
FROM notification_history
WHERE athlete_id = ?1
  AND achievement_type IN (/*SLICE:types*/?2);
