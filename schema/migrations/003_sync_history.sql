-- Sync History: Track import/sync runs and watermarks
-- Migration 003

--------------------------------------------------------------------------------
-- Sync History Table
--------------------------------------------------------------------------------

CREATE TABLE IF NOT EXISTS sync_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    athlete_id INTEGER NOT NULL REFERENCES athletes(id),

    -- Timing
    started_at TEXT NOT NULL,
    completed_at TEXT,
    duration_seconds INTEGER,

    -- Status
    status TEXT NOT NULL DEFAULT 'running', -- running, completed, failed, canceled
    error TEXT,

    -- Counts
    activities_total INTEGER DEFAULT 0,
    activities_imported INTEGER DEFAULT 0,
    activities_skipped INTEGER DEFAULT 0,
    gear_imported INTEGER DEFAULT 0,
    streams_imported INTEGER DEFAULT 0,
    segments_imported INTEGER DEFAULT 0,
    photos_imported INTEGER DEFAULT 0,
    failed_count INTEGER DEFAULT 0,

    -- Options used
    full_sync INTEGER DEFAULT 0,
    skip_streams INTEGER DEFAULT 0,
    skip_segments INTEGER DEFAULT 0,
    skip_best_efforts INTEGER DEFAULT 0,
    skip_photos INTEGER DEFAULT 0,

    -- Watermark: newest activity date synced in this run
    newest_activity_date TEXT,

    created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sync_history_athlete_id ON sync_history(athlete_id);
CREATE INDEX IF NOT EXISTS idx_sync_history_started_at ON sync_history(started_at DESC);
CREATE INDEX IF NOT EXISTS idx_sync_history_status ON sync_history(status);

--------------------------------------------------------------------------------
-- Sync Watermark (stored in app_state, but documenting the key here)
--------------------------------------------------------------------------------
-- Key: "sync_watermark:{athlete_id}"
-- Value: JSON {"last_synced_at": "2024-12-28T10:30:00Z", "newest_activity_date": "2024-12-27T15:00:00Z"}
