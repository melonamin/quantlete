-- Gear schema

CREATE TABLE IF NOT EXISTS gear (
    id TEXT PRIMARY KEY,
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    name TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    retired BOOLEAN DEFAULT FALSE,
    distance DOUBLE DEFAULT 0,        -- meters
    brand_name TEXT,
    model_name TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for looking up gear by athlete
CREATE INDEX IF NOT EXISTS idx_gear_athlete_id ON gear(athlete_id);
CREATE INDEX IF NOT EXISTS idx_gear_retired ON gear(retired);

