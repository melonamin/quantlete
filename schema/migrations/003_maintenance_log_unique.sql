-- Add unique constraint for maintenance_log to support ON CONFLICT clause
-- This prevents duplicate maintenance entries for the same component at the same time

CREATE UNIQUE INDEX IF NOT EXISTS idx_maintenance_log_component_completed
    ON maintenance_log(component_id, completed_at);
