-- Photos schema

CREATE TABLE IF NOT EXISTS photos (
    id TEXT PRIMARY KEY,
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    activity_id BIGINT NOT NULL REFERENCES activities(id),
    url TEXT NOT NULL,
    thumbnail_url TEXT,
    caption TEXT,
    location JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_photos_activity_id ON photos(activity_id);
CREATE INDEX IF NOT EXISTS idx_photos_athlete_id ON photos(athlete_id);

