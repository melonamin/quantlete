-- Activities schema

CREATE TABLE IF NOT EXISTS activities (
    id BIGINT PRIMARY KEY,
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    name TEXT NOT NULL,
    description TEXT,
    sport_type TEXT NOT NULL,
    start_date TIMESTAMP NOT NULL,
    start_date_local TIMESTAMP NOT NULL,
    timezone TEXT,

    -- Metrics
    distance DOUBLE DEFAULT 0,           -- meters
    moving_time INTEGER DEFAULT 0,       -- seconds
    elapsed_time INTEGER DEFAULT 0,      -- seconds
    total_elevation_gain DOUBLE DEFAULT 0, -- meters
    elev_high DOUBLE,
    elev_low DOUBLE,

    -- Speed
    average_speed DOUBLE DEFAULT 0,      -- m/s
    max_speed DOUBLE DEFAULT 0,          -- m/s

    -- Heart rate
    average_heartrate DOUBLE,
    max_heartrate INTEGER,

    -- Power
    average_watts DOUBLE,
    max_watts INTEGER,
    weighted_average_watts INTEGER,
    kilojoules DOUBLE,

    -- Cadence
    average_cadence DOUBLE,

    -- Other
    calories DOUBLE,
    kudos_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    photo_count INTEGER DEFAULT 0,
    commute BOOLEAN DEFAULT FALSE,
    private BOOLEAN DEFAULT FALSE,
    trainer BOOLEAN DEFAULT FALSE,
    workout_type INTEGER,
    device_name TEXT,
    gear_id TEXT,

    -- Location
    start_lat DOUBLE,
    start_lng DOUBLE,
    end_lat DOUBLE,
    end_lng DOUBLE,

    -- Map data (polylines)
    polyline TEXT,
    summary_polyline TEXT,

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for common queries
CREATE INDEX IF NOT EXISTS idx_activities_athlete_id ON activities(athlete_id);
CREATE INDEX IF NOT EXISTS idx_activities_sport_type ON activities(sport_type);
CREATE INDEX IF NOT EXISTS idx_activities_start_date ON activities(start_date DESC);
CREATE INDEX IF NOT EXISTS idx_activities_gear_id ON activities(gear_id);
CREATE INDEX IF NOT EXISTS idx_activities_commute ON activities(commute);

-- Composite index for filtered queries
CREATE INDEX IF NOT EXISTS idx_activities_athlete_sport_date
    ON activities(athlete_id, sport_type, start_date DESC);
