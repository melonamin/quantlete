-- Add CASCADE DELETE constraints and additional indexes for data integrity

-- SQLite doesn't support ALTER TABLE to add FK constraints, so we need to recreate tables
-- This migration uses temporary tables to preserve data while adding proper constraints

-- Enable foreign keys
PRAGMA foreign_keys = ON;

-- 1. Recreate activity_streams with CASCADE DELETE
CREATE TABLE IF NOT EXISTS activity_streams_new (
    activity_id INTEGER NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    stream_type TEXT NOT NULL,
    original_size INTEGER,
    resolution TEXT,
    series_type TEXT,
    data BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (activity_id, stream_type)
);

INSERT OR IGNORE INTO activity_streams_new SELECT * FROM activity_streams;
DROP TABLE IF EXISTS activity_streams;
ALTER TABLE activity_streams_new RENAME TO activity_streams;

-- 2. Recreate photos with CASCADE DELETE on activity
CREATE TABLE IF NOT EXISTS photos_new (
    id TEXT PRIMARY KEY,
    athlete_id INTEGER NOT NULL,
    activity_id INTEGER NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    thumbnail_url TEXT,
    caption TEXT,
    location BLOB,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO photos_new SELECT * FROM photos;
DROP TABLE IF EXISTS photos;
ALTER TABLE photos_new RENAME TO photos;

CREATE INDEX IF NOT EXISTS idx_photos_athlete_activity ON photos(athlete_id, activity_id);
CREATE INDEX IF NOT EXISTS idx_photos_athlete_id ON photos(athlete_id);

-- 3. Recreate best_efforts with CASCADE DELETE on activity
CREATE TABLE IF NOT EXISTS best_efforts_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    athlete_id INTEGER NOT NULL,
    activity_id INTEGER NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    sport_type TEXT NOT NULL,
    distance_type TEXT NOT NULL,
    name TEXT NOT NULL,
    distance_m REAL NOT NULL,
    elapsed_time_s INTEGER NOT NULL,
    moving_time_s INTEGER,
    start_index INTEGER,
    end_index INTEGER,
    pr_rank INTEGER,
    start_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(athlete_id, activity_id, distance_type)
);

INSERT OR IGNORE INTO best_efforts_new SELECT * FROM best_efforts;
DROP TABLE IF EXISTS best_efforts;
ALTER TABLE best_efforts_new RENAME TO best_efforts;

CREATE INDEX IF NOT EXISTS idx_best_efforts_athlete_sport_distance ON best_efforts(athlete_id, sport_type, distance_type);

-- 4. Recreate segment_efforts with CASCADE DELETE on activity and segment
CREATE TABLE IF NOT EXISTS segment_efforts_new (
    id INTEGER PRIMARY KEY,
    segment_id INTEGER NOT NULL REFERENCES segments(id) ON DELETE CASCADE,
    activity_id INTEGER NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    athlete_id INTEGER NOT NULL,
    name TEXT,
    elapsed_time INTEGER,
    moving_time INTEGER,
    start_date DATETIME,
    start_date_local DATETIME,
    distance REAL,
    average_watts REAL,
    average_heartrate REAL,
    max_heartrate INTEGER,
    pr_rank INTEGER,
    country TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO segment_efforts_new SELECT * FROM segment_efforts;
DROP TABLE IF EXISTS segment_efforts;
ALTER TABLE segment_efforts_new RENAME TO segment_efforts;

CREATE INDEX IF NOT EXISTS idx_segment_efforts_segment_id ON segment_efforts(segment_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_activity_id ON segment_efforts(activity_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_athlete_country ON segment_efforts(athlete_id, country);

-- 5. Recreate maintenance_rules with CASCADE DELETE on component
CREATE TABLE IF NOT EXISTS maintenance_rules_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    component_id INTEGER NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    metric TEXT NOT NULL CHECK(metric IN ('distance', 'time', 'interval')),
    threshold_value REAL NOT NULL,
    threshold_unit TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO maintenance_rules_new SELECT * FROM maintenance_rules;
DROP TABLE IF EXISTS maintenance_rules;
ALTER TABLE maintenance_rules_new RENAME TO maintenance_rules;

CREATE INDEX IF NOT EXISTS idx_maintenance_rules_component_id ON maintenance_rules(component_id);

-- 6. Recreate maintenance_log with CASCADE DELETE on component, SET NULL on activity
CREATE TABLE IF NOT EXISTS maintenance_log_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    component_id INTEGER NOT NULL REFERENCES components(id) ON DELETE CASCADE,
    activity_id INTEGER REFERENCES activities(id) ON DELETE SET NULL,
    completed_at DATETIME NOT NULL,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO maintenance_log_new SELECT * FROM maintenance_log;
DROP TABLE IF EXISTS maintenance_log;
ALTER TABLE maintenance_log_new RENAME TO maintenance_log;

CREATE INDEX IF NOT EXISTS idx_maintenance_log_component_id ON maintenance_log(component_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_log_activity_id ON maintenance_log(activity_id);

-- 7. Recreate components with CASCADE DELETE on gear
CREATE TABLE IF NOT EXISTS components_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gear_id TEXT NOT NULL REFERENCES gear(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    installed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    retired_at DATETIME,
    initial_distance REAL DEFAULT 0,
    initial_time INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO components_new SELECT * FROM components;
DROP TABLE IF EXISTS components;
ALTER TABLE components_new RENAME TO components;

CREATE INDEX IF NOT EXISTS idx_components_gear_id ON components(gear_id);

-- 8. Additional indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_activities_athlete_gear ON activities(athlete_id, gear_id);
CREATE INDEX IF NOT EXISTS idx_activities_athlete_commute ON activities(athlete_id, commute);
CREATE INDEX IF NOT EXISTS idx_activities_athlete_trainer ON activities(athlete_id, trainer);
