-- Stata: Complete SQLite Schema
-- Consolidated from migrations 001-019

--------------------------------------------------------------------------------
-- Athletes & Authentication
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS athletes (
    id INTEGER PRIMARY KEY,
    username TEXT,
    firstname TEXT,
    lastname TEXT,
    city TEXT,
    state TEXT,
    country TEXT,
    sex TEXT,
    premium INTEGER DEFAULT 0,
    summit INTEGER DEFAULT 0,
    profile_medium TEXT,
    profile TEXT,
    weight REAL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS auth_tokens (
    athlete_id INTEGER PRIMARY KEY REFERENCES athletes(id),
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_type TEXT DEFAULT 'Bearer',
    expires_at TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires_at ON auth_tokens(expires_at);

--------------------------------------------------------------------------------
-- Activities
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS activities (
    id INTEGER PRIMARY KEY,
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    name TEXT NOT NULL,
    description TEXT,
    sport_type TEXT NOT NULL,
    start_date TEXT NOT NULL,
    start_date_local TEXT NOT NULL,
    timezone TEXT,

    -- Location
    location_city TEXT,
    location_state TEXT,
    location_country TEXT,

    -- Metrics
    distance REAL DEFAULT 0,
    moving_time INTEGER DEFAULT 0,
    elapsed_time INTEGER DEFAULT 0,
    total_elevation_gain REAL DEFAULT 0,
    elev_high REAL,
    elev_low REAL,

    -- Speed
    average_speed REAL DEFAULT 0,
    max_speed REAL DEFAULT 0,

    -- Heart rate
    average_heartrate REAL,
    max_heartrate INTEGER,

    -- Power
    average_watts REAL,
    max_watts INTEGER,
    weighted_average_watts INTEGER,
    kilojoules REAL,

    -- Cadence
    average_cadence REAL,

    -- Other
    calories REAL,
    kudos_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    photo_count INTEGER DEFAULT 0,
    commute INTEGER DEFAULT 0,
    private INTEGER DEFAULT 0,
    trainer INTEGER DEFAULT 0,
    workout_type INTEGER,
    device_name TEXT,
    gear_id TEXT,

    -- Coordinates
    start_lat REAL,
    start_lng REAL,
    end_lat REAL,
    end_lng REAL,

    -- Map data
    polyline TEXT,
    summary_polyline TEXT,

    -- Metadata
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activities_athlete_id ON activities(athlete_id);
CREATE INDEX IF NOT EXISTS idx_activities_sport_type ON activities(sport_type);
CREATE INDEX IF NOT EXISTS idx_activities_start_date ON activities(start_date DESC);
CREATE INDEX IF NOT EXISTS idx_activities_gear_id ON activities(gear_id);
CREATE INDEX IF NOT EXISTS idx_activities_commute ON activities(commute);
CREATE INDEX IF NOT EXISTS idx_activities_location_country ON activities(location_country);
CREATE INDEX IF NOT EXISTS idx_activities_athlete_sport_date ON activities(athlete_id, sport_type, start_date DESC);

--------------------------------------------------------------------------------
-- Activity Streams
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS activity_streams (
    activity_id INTEGER NOT NULL REFERENCES activities(id),
    stream_type TEXT NOT NULL,
    original_size INTEGER,
    resolution TEXT,
    series_type TEXT,
    data TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (activity_id, stream_type)
);

CREATE INDEX IF NOT EXISTS idx_streams_activity_id ON activity_streams(activity_id);

--------------------------------------------------------------------------------
-- Gear & Maintenance
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS gear (
    id TEXT PRIMARY KEY,
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    name TEXT NOT NULL,
    is_primary INTEGER DEFAULT 0,
    retired INTEGER DEFAULT 0,
    distance REAL DEFAULT 0,
    brand_name TEXT,
    model_name TEXT,
    description TEXT,
    source TEXT DEFAULT 'strava',
    hashtag TEXT,
    purchase_price REAL,
    purchase_currency TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_gear_athlete_id ON gear(athlete_id);
CREATE INDEX IF NOT EXISTS idx_gear_retired ON gear(retired);
CREATE INDEX IF NOT EXISTS idx_gear_source ON gear(source);
CREATE INDEX IF NOT EXISTS idx_gear_hashtag ON gear(hashtag);

CREATE TABLE IF NOT EXISTS components (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gear_id TEXT NOT NULL REFERENCES gear(id),
    name TEXT NOT NULL,
    image_url TEXT,
    maintenance_hashtag TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS maintenance_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    component_id INTEGER NOT NULL REFERENCES components(id),
    type TEXT NOT NULL,
    threshold_value REAL NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS maintenance_log (
    component_id INTEGER NOT NULL REFERENCES components(id),
    activity_id INTEGER REFERENCES activities(id),
    completed_at TEXT NOT NULL,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (component_id, completed_at)
);

CREATE INDEX IF NOT EXISTS idx_components_gear_id ON components(gear_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_rules_component_id ON maintenance_rules(component_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_log_component_id ON maintenance_log(component_id);

--------------------------------------------------------------------------------
-- Segments
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS segments (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    activity_type TEXT,
    distance REAL,
    average_grade REAL,
    maximum_grade REAL,
    elevation_high REAL,
    elevation_low REAL,
    climb_category INTEGER,
    start_lat REAL,
    start_lng REAL,
    end_lat REAL,
    end_lng REAL,
    starred INTEGER DEFAULT 0,
    polyline TEXT,
    athlete_kom_rank INTEGER,
    athlete_effort_count INTEGER,
    athlete_pr_elapsed_time INTEGER,
    athlete_pr_date TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS segment_efforts (
    id INTEGER PRIMARY KEY,
    segment_id INTEGER NOT NULL REFERENCES segments(id),
    activity_id INTEGER NOT NULL REFERENCES activities(id),
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    name TEXT,
    elapsed_time INTEGER,
    moving_time INTEGER,
    start_date TEXT,
    start_date_local TEXT,
    distance REAL,
    average_watts REAL,
    average_heartrate REAL,
    max_heartrate INTEGER,
    pr_rank INTEGER,
    country TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_segment_efforts_segment_id ON segment_efforts(segment_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_activity_id ON segment_efforts(activity_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_athlete_id ON segment_efforts(athlete_id);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_pr_rank ON segment_efforts(pr_rank);
CREATE INDEX IF NOT EXISTS idx_segment_efforts_country ON segment_efforts(country);

--------------------------------------------------------------------------------
-- Photos
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS photos (
    id TEXT PRIMARY KEY,
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    activity_id INTEGER NOT NULL REFERENCES activities(id),
    url TEXT NOT NULL,
    thumbnail_url TEXT,
    caption TEXT,
    location TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_photos_activity_id ON photos(activity_id);
CREATE INDEX IF NOT EXISTS idx_photos_athlete_id ON photos(athlete_id);

--------------------------------------------------------------------------------
-- Challenges
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS challenges (
    id TEXT PRIMARY KEY,
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    name TEXT NOT NULL,
    slug TEXT,
    badge_url TEXT,
    completion_date TEXT,
    month TEXT,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_challenges_athlete_id ON challenges(athlete_id);
CREATE INDEX IF NOT EXISTS idx_challenges_month ON challenges(month);

--------------------------------------------------------------------------------
-- Best Efforts
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS best_efforts (
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    activity_id INTEGER NOT NULL REFERENCES activities(id),
    sport_type TEXT,
    distance_type TEXT NOT NULL,
    distance_m REAL NOT NULL,
    elapsed_time INTEGER NOT NULL,
    start_index INTEGER,
    end_index INTEGER,
    start_date TEXT,
    name TEXT,
    pr_rank INTEGER,
    moving_time INTEGER,
    created_at TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (athlete_id, activity_id, distance_type)
);

CREATE INDEX IF NOT EXISTS idx_best_efforts_athlete_distance ON best_efforts(athlete_id, distance_type);
CREATE INDEX IF NOT EXISTS idx_best_efforts_distance_time ON best_efforts(distance_type, elapsed_time);
CREATE INDEX IF NOT EXISTS idx_best_efforts_pr_rank ON best_efforts(athlete_id, pr_rank);

--------------------------------------------------------------------------------
-- Configuration & Settings
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS dashboard_config (
    athlete_id INTEGER PRIMARY KEY REFERENCES athletes(id),
    config TEXT NOT NULL,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS training_goals (
    athlete_id INTEGER PRIMARY KEY REFERENCES athletes(id),
    config TEXT NOT NULL,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS athlete_settings (
    athlete_id INTEGER PRIMARY KEY REFERENCES athletes(id),
    settings TEXT NOT NULL,
    updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

--------------------------------------------------------------------------------
-- Training Load (TSS/CTL/ATL/TSB)
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS activity_training_load (
    activity_id INTEGER PRIMARY KEY REFERENCES activities(id),
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    sport_type TEXT,
    method TEXT,
    ftp_used REAL,
    normalized_power REAL,
    intensity_factor REAL,
    tss REAL,
    computed_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS daily_training_load (
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    day TEXT NOT NULL,
    tss REAL NOT NULL,
    ctl REAL NOT NULL,
    atl REAL NOT NULL,
    tsb REAL NOT NULL,
    PRIMARY KEY (athlete_id, day)
);

CREATE INDEX IF NOT EXISTS idx_daily_training_load_athlete_day ON daily_training_load(athlete_id, day);

--------------------------------------------------------------------------------
-- Athlete Metrics (FTP, Weight history)
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS athlete_metrics (
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    metric TEXT NOT NULL,
    value REAL NOT NULL,
    recorded_at TEXT NOT NULL,
    PRIMARY KEY (athlete_id, metric, recorded_at)
);

CREATE INDEX IF NOT EXISTS idx_athlete_metrics_athlete_metric ON athlete_metrics(athlete_id, metric);
CREATE INDEX IF NOT EXISTS idx_athlete_metrics_recorded_at ON athlete_metrics(recorded_at);

--------------------------------------------------------------------------------
-- HR Zone Definitions
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS hr_zone_definitions (
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    sport_type TEXT NOT NULL,
    effective_from TEXT NOT NULL,
    method TEXT NOT NULL,
    zones TEXT NOT NULL,
    PRIMARY KEY (athlete_id, sport_type, effective_from)
);

CREATE INDEX IF NOT EXISTS idx_hr_zones_athlete_sport ON hr_zone_definitions(athlete_id, sport_type);

--------------------------------------------------------------------------------
-- Power Best Efforts
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS power_best_efforts (
    activity_id INTEGER NOT NULL REFERENCES activities(id),
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),
    duration_s INTEGER NOT NULL,
    best_avg_watts REAL NOT NULL,
    computed_at TEXT DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (activity_id, duration_s)
);

CREATE INDEX IF NOT EXISTS idx_power_best_efforts_athlete_duration ON power_best_efforts(athlete_id, duration_s);
