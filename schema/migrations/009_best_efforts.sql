-- Best efforts schema

CREATE TABLE IF NOT EXISTS best_efforts (
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    activity_id BIGINT NOT NULL REFERENCES activities(id),
    sport_type TEXT,
    distance_type TEXT NOT NULL,
    distance_m DOUBLE NOT NULL,
    elapsed_time INTEGER NOT NULL, -- seconds
    start_index INTEGER,
    end_index INTEGER,
    start_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (athlete_id, activity_id, distance_type)
);

CREATE INDEX IF NOT EXISTS idx_best_efforts_athlete_distance ON best_efforts(athlete_id, distance_type);
CREATE INDEX IF NOT EXISTS idx_best_efforts_distance_time ON best_efforts(distance_type, elapsed_time);

