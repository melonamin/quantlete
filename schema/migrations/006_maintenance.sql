-- Gear maintenance schema

CREATE SEQUENCE IF NOT EXISTS components_id_seq;
CREATE SEQUENCE IF NOT EXISTS maintenance_rules_id_seq;

CREATE TABLE IF NOT EXISTS components (
    id BIGINT PRIMARY KEY DEFAULT nextval('components_id_seq'),
    gear_id TEXT NOT NULL REFERENCES gear(id),
    name TEXT NOT NULL,
    image_url TEXT,
    maintenance_hashtag TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS maintenance_rules (
    id BIGINT PRIMARY KEY DEFAULT nextval('maintenance_rules_id_seq'),
    component_id BIGINT NOT NULL REFERENCES components(id),
    type TEXT NOT NULL, -- 'distance_m', 'time_s', 'days'
    threshold_value DOUBLE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS maintenance_log (
    component_id BIGINT NOT NULL REFERENCES components(id),
    activity_id BIGINT REFERENCES activities(id),
    completed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (component_id, completed_at)
);

CREATE INDEX IF NOT EXISTS idx_components_gear_id ON components(gear_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_rules_component_id ON maintenance_rules(component_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_log_component_id ON maintenance_log(component_id);

