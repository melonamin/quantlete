-- Peak power outputs derived from streams

CREATE TABLE IF NOT EXISTS power_best_efforts (
    activity_id BIGINT NOT NULL REFERENCES activities(id),
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    duration_s INTEGER NOT NULL,
    best_avg_watts DOUBLE NOT NULL,
    computed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (activity_id, duration_s)
);

CREATE INDEX IF NOT EXISTS idx_power_best_efforts_athlete_duration ON power_best_efforts(athlete_id, duration_s);

