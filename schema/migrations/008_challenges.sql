-- Challenges schema

CREATE TABLE IF NOT EXISTS challenges (
    id TEXT PRIMARY KEY,
    athlete_id BIGINT NOT NULL REFERENCES athletes(id),
    name TEXT NOT NULL,
    slug TEXT,
    badge_url TEXT,
    completion_date DATE,
    month TEXT, -- YYYY-MM
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_challenges_athlete_id ON challenges(athlete_id);
CREATE INDEX IF NOT EXISTS idx_challenges_month ON challenges(month);

