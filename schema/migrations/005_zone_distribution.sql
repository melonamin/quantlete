-- Activity zone distribution table to store HR zone time per activity
-- Enables fast aggregation for zone trend charts without recalculating from HR streams

CREATE TABLE IF NOT EXISTS activity_zone_distributions (
    activity_id INTEGER PRIMARY KEY REFERENCES activities(id) ON DELETE CASCADE,
    zone_def_id TEXT NOT NULL,  -- "sport_type:effective_from" for invalidation tracking
    seconds_z1 INTEGER NOT NULL DEFAULT 0,
    seconds_z2 INTEGER NOT NULL DEFAULT 0,
    seconds_z3 INTEGER NOT NULL DEFAULT 0,
    seconds_z4 INTEGER NOT NULL DEFAULT 0,
    seconds_z5 INTEGER NOT NULL DEFAULT 0,
    total_seconds INTEGER NOT NULL DEFAULT 0,
    computed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Note: No index needed on activity_id - PRIMARY KEY already creates one
CREATE INDEX IF NOT EXISTS idx_zone_dist_zone_def ON activity_zone_distributions(zone_def_id);
