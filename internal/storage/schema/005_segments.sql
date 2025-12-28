-- Segments schema

CREATE TABLE IF NOT EXISTS segments (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    activity_type TEXT,
    distance DOUBLE,
    average_grade DOUBLE,
    maximum_grade DOUBLE,
    elevation_high DOUBLE,
    elevation_low DOUBLE,
    climb_category INTEGER,
    start_lat DOUBLE,
    start_lng DOUBLE,
    end_lat DOUBLE,
    end_lng DOUBLE,
    starred BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Segment efforts (personal records on segments)
CREATE TABLE IF NOT EXISTS segment_efforts (
    id BIGINT PRIMARY KEY,
    segment_id BIGINT NOT NULL REFERENCES segments(id),
    activity_id BIGINT NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    name TEXT,
    elapsed_time INTEGER,     -- seconds
    moving_time INTEGER,      -- seconds
    start_date TIMESTAMP,
    start_date_local TIMESTAMP,
    distance DOUBLE,
    average_watts DOUBLE,
    average_heartrate DOUBLE,
    max_heartrate INTEGER,
    pr_rank INTEGER,          -- NULL if not a PR, 1-3 for top 3
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for segment queries
CREATE INDEX IF NOT EXISTS idx_segment_efforts_segment_id ON segment_efforts(segment_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_activity_id ON segment_efforts(activity_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_athlete_id ON segment_efforts(athlete_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_pr_rank ON segment_efforts(pr_rank) WHERE pr_rank IS NOT NULL;
