-- Notification history table to track which achievements have been notified
-- This prevents duplicate notifications when the same achievements are detected during re-import

CREATE TABLE IF NOT EXISTS notification_history (
    id INTEGER PRIMARY KEY,
    athlete_id INTEGER NOT NULL,
    achievement_type TEXT NOT NULL,
    achievement_key TEXT NOT NULL,
    notified_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Unique constraint to prevent duplicate notifications
CREATE UNIQUE INDEX IF NOT EXISTS idx_notification_history_unique
    ON notification_history(athlete_id, achievement_type, achievement_key);

-- Index for efficient lookups by athlete
CREATE INDEX IF NOT EXISTS idx_notification_history_athlete
    ON notification_history(athlete_id);
