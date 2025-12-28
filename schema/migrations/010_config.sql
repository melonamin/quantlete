-- Dashboard + settings configuration storage

CREATE TABLE IF NOT EXISTS dashboard_config (
    athlete_id BIGINT PRIMARY KEY REFERENCES athletes(id),
    config JSON NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS training_goals (
    athlete_id BIGINT PRIMARY KEY REFERENCES athletes(id),
    config JSON NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS athlete_settings (
    athlete_id BIGINT PRIMARY KEY REFERENCES athletes(id),
    settings JSON NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

