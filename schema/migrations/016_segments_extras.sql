-- Segment extras needed for maps + filtering

ALTER TABLE segments ADD COLUMN polyline TEXT;
ALTER TABLE segment_efforts ADD COLUMN country TEXT;

CREATE INDEX IF NOT EXISTS idx_segment_efforts_country ON segment_efforts(country);

