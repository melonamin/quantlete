-- Training load schema (TSS/CTL/ATL/TSB)

CREATE TABLE IF NOT EXISTS activity_training_load (
    activity_id BIGINT PRIMARY KEY REFERENCES activities(id),
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    sport_type TEXT,
    method TEXT, -- 'cycling_power', 'running_pace', 'running_hr', etc
    ftp_used DOUBLE,
    normalized_power DOUBLE,
    intensity_factor DOUBLE,
    tss DOUBLE,
    computed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS daily_training_load (
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    day DATE NOT NULL,
    tss DOUBLE NOT NULL,
    ctl DOUBLE NOT NULL,
    atl DOUBLE NOT NULL,
    tsb DOUBLE NOT NULL,
    PRIMARY KEY (athlete_id, day)
);

CREATE INDEX IF NOT EXISTS idx_daily_training_load_athlete_day ON daily_training_load(athlete_id, day);

