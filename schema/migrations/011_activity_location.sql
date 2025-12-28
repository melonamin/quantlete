-- Activity location fields used for country stats and filters

ALTER TABLE activities ADD COLUMN location_city TEXT;
ALTER TABLE activities ADD COLUMN location_state TEXT;
ALTER TABLE activities ADD COLUMN location_country TEXT;

CREATE INDEX IF NOT EXISTS idx_activities_location_country ON activities(location_country);

