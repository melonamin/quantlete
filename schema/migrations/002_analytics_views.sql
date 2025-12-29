-- Analytics VIEWs for shared query logic between Go server and WASM browser mode.
-- These VIEWs encapsulate complex aggregation logic that would otherwise be duplicated.

-- Daily distance aggregation for Eddington number calculation.
-- Groups activities by date and sums distance in kilometers.
CREATE VIEW IF NOT EXISTS v_daily_distances AS
SELECT
    athlete_id,
    DATE(start_date) AS date,
    SUM(distance) / 1000.0 AS distance_km
FROM activities
GROUP BY athlete_id, DATE(start_date);

-- Best effort rankings per distance type.
-- Uses window function to rank efforts, enabling easy PR lookups.
CREATE VIEW IF NOT EXISTS v_best_effort_rankings AS
SELECT
    be.athlete_id,
    be.distance_type,
    COALESCE(be.name, be.distance_type) AS name,
    be.distance_m,
    be.elapsed_time,
    be.moving_time,
    be.pr_rank,
    a.id AS activity_id,
    a.name AS activity_name,
    a.sport_type,
    a.start_date_local,
    ROW_NUMBER() OVER (
        PARTITION BY be.athlete_id, be.distance_type
        ORDER BY be.elapsed_time ASC, a.start_date_local ASC
    ) AS rn
FROM best_efforts be
JOIN activities a ON a.id = be.activity_id AND a.athlete_id = be.athlete_id;

-- Peak power rankings per duration.
-- Ranks power outputs for each time duration to find all-time bests.
CREATE VIEW IF NOT EXISTS v_power_best_rankings AS
SELECT
    p.athlete_id,
    p.duration_s,
    p.best_avg_watts,
    p.activity_id,
    a.start_date,
    a.name AS activity_name,
    ROW_NUMBER() OVER (
        PARTITION BY p.athlete_id, p.duration_s
        ORDER BY p.best_avg_watts DESC
    ) AS rn
FROM power_best_efforts p
JOIN activities a ON a.id = p.activity_id;

-- Segment statistics with effort counts.
-- Aggregates effort data per segment per athlete.
CREATE VIEW IF NOT EXISTS v_segment_stats AS
SELECT
    s.id,
    s.name,
    s.activity_type,
    s.distance,
    s.average_grade,
    s.maximum_grade,
    s.elevation_high,
    s.elevation_low,
    s.climb_category,
    s.start_lat,
    s.start_lng,
    s.end_lat,
    s.end_lng,
    s.starred,
    s.polyline,
    s.athlete_kom_rank,
    s.athlete_effort_count,
    s.athlete_pr_elapsed_time,
    s.athlete_pr_date,
    e.athlete_id,
    COUNT(e.id) AS times_completed,
    MAX(e.start_date) AS last_effort_date,
    MIN(NULLIF(e.elapsed_time, 0)) AS best_elapsed_time
FROM segments s
LEFT JOIN segment_efforts e ON e.segment_id = s.id
GROUP BY s.id, e.athlete_id;

-- Monthly activity aggregation by sport type.
-- Provides monthly totals broken down by activity type.
CREATE VIEW IF NOT EXISTS v_monthly_stats AS
SELECT
    athlete_id,
    strftime('%Y-%m', start_date) AS month,
    sport_type,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
GROUP BY athlete_id, strftime('%Y-%m', start_date), sport_type;

-- Yearly activity aggregation.
-- Provides yearly totals across all activity types.
CREATE VIEW IF NOT EXISTS v_yearly_stats AS
SELECT
    athlete_id,
    CAST(strftime('%Y', start_date) AS INTEGER) AS year,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
GROUP BY athlete_id, strftime('%Y', start_date);

-- Daily TSS aggregation for training load calculations.
-- Groups TSS values by day for CTL/ATL/TSB computation.
CREATE VIEW IF NOT EXISTS v_daily_tss AS
SELECT
    a.athlete_id,
    DATE(a.start_date_local) AS day,
    COALESCE(SUM(tl.tss), 0) AS tss
FROM activities a
LEFT JOIN activity_training_load tl ON tl.activity_id = a.id
GROUP BY a.athlete_id, DATE(a.start_date_local);

-- Weekly activity aggregation.
-- Provides weekly totals for dashboard and trends.
CREATE VIEW IF NOT EXISTS v_weekly_stats AS
SELECT
    athlete_id,
    strftime('%Y-W%W', start_date) AS week,
    strftime('%Y-%m-%d', start_date, 'weekday 0', '-6 days') AS week_start,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
GROUP BY athlete_id, strftime('%Y-W%W', start_date);

-- Sport type statistics.
-- Aggregates all-time stats by sport type.
CREATE VIEW IF NOT EXISTS v_sport_type_stats AS
SELECT
    athlete_id,
    sport_type,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time,
    COALESCE(SUM(total_elevation_gain), 0) AS total_elevation
FROM activities
GROUP BY athlete_id, sport_type;

-- Gear usage statistics.
-- Aggregates usage by gear and month.
CREATE VIEW IF NOT EXISTS v_gear_monthly_usage AS
SELECT
    g.id AS gear_id,
    g.name AS gear_name,
    g.source,
    g.hashtag,
    g.retired,
    g.purchase_price,
    g.purchase_currency,
    g.athlete_id,
    strftime('%Y-%m', a.start_date) AS month,
    COUNT(*) AS activity_count,
    COALESCE(SUM(a.distance), 0) AS distance,
    COALESCE(SUM(a.moving_time), 0) AS moving_time
FROM gear g
LEFT JOIN activities a ON a.gear_id = g.id AND a.athlete_id = g.athlete_id
GROUP BY g.id, strftime('%Y-%m', a.start_date);

-- Segment country statistics.
-- Counts segments per country for filtering.
CREATE VIEW IF NOT EXISTS v_segment_countries AS
SELECT
    athlete_id,
    country,
    COUNT(DISTINCT segment_id) AS segment_count
FROM segment_efforts
WHERE country IS NOT NULL AND country != ''
GROUP BY athlete_id, country;

-- Heatmap activity data.
-- Provides polylines and coordinates for map visualization.
CREATE VIEW IF NOT EXISTS v_heatmap_activities AS
SELECT
    id,
    athlete_id,
    sport_type,
    summary_polyline,
    COALESCE(start_lat, 0) AS start_lat,
    COALESCE(start_lng, 0) AS start_lng,
    start_date,
    commute,
    workout_type,
    location_country
FROM activities
WHERE summary_polyline IS NOT NULL AND summary_polyline != '';

-- Calendar day statistics.
-- Provides daily aggregation for calendar heatmap.
CREATE VIEW IF NOT EXISTS v_calendar_days AS
SELECT
    athlete_id,
    DATE(start_date) AS date,
    COUNT(*) AS activity_count,
    COALESCE(SUM(distance), 0) AS total_distance,
    COALESCE(SUM(moving_time), 0) AS total_time
FROM activities
GROUP BY athlete_id, DATE(start_date);
