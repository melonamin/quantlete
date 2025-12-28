-- Activity streams schema
-- Streams contain time-series data for activities (heart rate, power, GPS, etc.)

CREATE TABLE IF NOT EXISTS activity_streams (
    activity_id BIGINT NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    stream_type TEXT NOT NULL,  -- 'time', 'distance', 'latlng', 'altitude', etc.
    original_size INTEGER,
    resolution TEXT,
    series_type TEXT,
    data JSON NOT NULL,         -- DuckDB supports JSON type natively
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (activity_id, stream_type)
);

-- Index for querying streams by activity
CREATE INDEX IF NOT EXISTS idx_streams_activity_id ON activity_streams(activity_id);
