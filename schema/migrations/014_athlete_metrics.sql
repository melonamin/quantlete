-- Athlete measurement history (FTP, weight, etc.)

CREATE TABLE IF NOT EXISTS athlete_metrics (
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    metric TEXT NOT NULL, -- 'ftp_cycling_watts' | 'ftp_running_mps' | 'weight_kg'
    value DOUBLE NOT NULL,
    recorded_at TIMESTAMP NOT NULL,
    PRIMARY KEY (athlete_id, metric, recorded_at)
);

CREATE INDEX IF NOT EXISTS idx_athlete_metrics_athlete_metric ON athlete_metrics(athlete_id, metric);
CREATE INDEX IF NOT EXISTS idx_athlete_metrics_recorded_at ON athlete_metrics(recorded_at);

