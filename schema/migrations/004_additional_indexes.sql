-- Additional indexes for query performance

-- Photos: Composite index for athlete activity lookups
CREATE INDEX IF NOT EXISTS idx_photos_athlete_activity ON photos(athlete_id, activity_id);

-- Segment efforts: Composite index for athlete country filtering
CREATE INDEX IF NOT EXISTS idx_segment_efforts_athlete_country ON segment_efforts(athlete_id, country);

-- Maintenance log: Index for activity lookups
CREATE INDEX IF NOT EXISTS idx_maintenance_log_activity_id ON maintenance_log(activity_id);

-- Challenges: Composite index for athlete and completion date filtering
CREATE INDEX IF NOT EXISTS idx_challenges_athlete_completion ON challenges(athlete_id, completion_date);
