-- Zone definitions (heart rate)

CREATE TABLE IF NOT EXISTS hr_zone_definitions (
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    sport_type TEXT NOT NULL, -- 'Run', 'Ride', 'All'
    effective_from DATE NOT NULL,
    method TEXT NOT NULL, -- 'absolute_bpm' | 'percent_hrmax'
    zones JSON NOT NULL,
    PRIMARY KEY (athlete_id, sport_type, effective_from)
);

CREATE INDEX IF NOT EXISTS idx_hr_zones_athlete_sport ON hr_zone_definitions(athlete_id, sport_type);

