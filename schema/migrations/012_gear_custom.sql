-- Custom gear + extra gear metadata

ALTER TABLE gear ADD COLUMN source TEXT DEFAULT 'strava'; -- 'strava' | 'custom'
ALTER TABLE gear ADD COLUMN hashtag TEXT;
ALTER TABLE gear ADD COLUMN purchase_price DOUBLE;
ALTER TABLE gear ADD COLUMN purchase_currency TEXT;

CREATE INDEX IF NOT EXISTS idx_gear_source ON gear(source);
CREATE INDEX IF NOT EXISTS idx_gear_hashtag ON gear(hashtag);

