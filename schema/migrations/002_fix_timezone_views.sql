-- Fix timezone handling: use start_date_local for activity grouping
-- This ensures activities are grouped by their local calendar date, not UTC

-- Drop existing views that use start_date (UTC)
DROP VIEW IF EXISTS v_daily_distances;
DROP VIEW IF EXISTS v_monthly_stats;
DROP VIEW IF EXISTS v_yearly_stats;
DROP VIEW IF EXISTS v_weekly_stats;
DROP VIEW IF EXISTS v_gear_monthly_usage;
DROP VIEW IF EXISTS v_calendar_days;

-- Recreate views with start_date_local for correct timezone grouping

-- Daily distance aggregation for Eddington number calculation
CREATE VIEW v_daily_distances AS
SELECT
    athlete_id,
    DATE(start_date_local) AS date,
    SUM(distance) / 1000.0 AS distance_km
FROM activities
GROUP BY athlete_id, DATE(start_date_local);

-- Monthly activity aggregation by sport type
CREATE VIEW v_monthly_stats AS
SELECT
    athlete_id,
    strftime('%Y-%m', start_date_local) AS month,
    sport_type,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
GROUP BY athlete_id, strftime('%Y-%m', start_date_local), sport_type;

-- Yearly activity aggregation
CREATE VIEW v_yearly_stats AS
SELECT
    athlete_id,
    CAST(strftime('%Y', start_date_local) AS INTEGER) AS year,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
GROUP BY athlete_id, strftime('%Y', start_date_local);

-- Weekly activity aggregation
CREATE VIEW v_weekly_stats AS
SELECT
    athlete_id,
    strftime('%Y-W%W', start_date_local) AS week,
    strftime('%Y-%m-%d', start_date_local, 'weekday 0', '-6 days') AS week_start,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
GROUP BY athlete_id, strftime('%Y-W%W', start_date_local);

-- Gear usage statistics
CREATE VIEW v_gear_monthly_usage AS
SELECT
    g.id AS gear_id,
    g.name AS gear_name,
    g.source,
    g.hashtag,
    g.retired,
    g.purchase_price,
    g.purchase_currency,
    g.athlete_id,
    strftime('%Y-%m', a.start_date_local) AS month,
    COUNT(*) AS activity_count,
    COALESCE(SUM(a.distance), 0) AS distance,
    COALESCE(SUM(a.moving_time), 0) AS moving_time
FROM gear g
LEFT JOIN activities a ON a.gear_id = g.id AND a.athlete_id = g.athlete_id
GROUP BY g.id, strftime('%Y-%m', a.start_date_local);

-- Calendar day statistics
CREATE VIEW v_calendar_days AS
SELECT
    athlete_id,
    DATE(start_date_local) AS date,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(calories), 0) AS total_calories
FROM activities
GROUP BY athlete_id, DATE(start_date_local);
