-- Add intensity metric to calendar view for heatmap coloring
-- Uses suffer_score from activities (relative effort from Strava)

-- Drop and recreate the view with intensity column
DROP VIEW IF EXISTS v_calendar_days;

CREATE VIEW v_calendar_days AS
SELECT
    athlete_id,
    DATE(start_date_local) AS date,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(calories), 0) AS total_calories,
    COALESCE(SUM(suffer_score), 0) AS total_intensity
FROM activities
GROUP BY athlete_id, DATE(start_date_local);
