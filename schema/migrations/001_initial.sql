-- Initial schema: athletes and authentication tokens

-- Athletes table stores Strava athlete profiles
CREATE TABLE IF NOT EXISTS athletes (
    id BIGINT PRIMARY KEY,
    username TEXT,
    firstname TEXT,
    lastname TEXT,
    city TEXT,
    state TEXT,
    country TEXT,
    sex TEXT,
    premium BOOLEAN DEFAULT FALSE,
    summit BOOLEAN DEFAULT FALSE,
    profile_medium TEXT,
    profile TEXT,
    weight DOUBLE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Auth tokens table stores OAuth tokens for each athlete
CREATE TABLE IF NOT EXISTS auth_tokens (
    athlete_id BIGINT PRIMARY KEY REFERENCES athletes(id),
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_type TEXT DEFAULT 'Bearer',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for token expiry checks
CREATE INDEX IF NOT EXISTS idx_auth_tokens_expires_at ON auth_tokens(expires_at);

