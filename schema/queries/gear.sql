-- Gear queries.

-- name: UpsertGear :exec
-- Insert or update gear.
INSERT INTO gear (
    id, athlete_id, name, is_primary, retired, distance,
    brand_name, model_name, description, source, hashtag,
    purchase_price, purchase_currency
) VALUES (?1, ?2, ?3, ?4, ?5, ?6, ?7, ?8, ?9, ?10, ?11, ?12, ?13)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    is_primary = EXCLUDED.is_primary,
    retired = EXCLUDED.retired,
    distance = EXCLUDED.distance,
    brand_name = EXCLUDED.brand_name,
    model_name = EXCLUDED.model_name,
    description = EXCLUDED.description,
    source = EXCLUDED.source,
    hashtag = EXCLUDED.hashtag,
    purchase_price = EXCLUDED.purchase_price,
    purchase_currency = EXCLUDED.purchase_currency;

-- name: GetGearByID :one
-- Get gear by ID.
SELECT
    id, athlete_id, name, is_primary, retired, distance,
    brand_name, model_name, description,
    COALESCE(source, '') AS source,
    COALESCE(hashtag, '') AS hashtag,
    purchase_price,
    COALESCE(purchase_currency, '') AS purchase_currency,
    created_at, updated_at
FROM gear
WHERE id = ?1;

-- name: GetGear :many
-- Get all gear for an athlete.
SELECT
    g.id,
    g.athlete_id,
    g.name,
    g.is_primary,
    g.retired,
    g.distance,
    g.brand_name,
    g.model_name,
    g.description,
    COALESCE(g.source, '') AS source,
    COALESCE(g.hashtag, '') AS hashtag,
    g.purchase_price,
    COALESCE(g.purchase_currency, '') AS purchase_currency,
    g.created_at,
    g.updated_at
FROM gear g
WHERE g.athlete_id = ?1
ORDER BY g.is_primary DESC, g.name;

-- name: GetActiveGear :many
-- Get non-retired gear for an athlete.
SELECT
    g.id,
    g.athlete_id,
    g.name,
    g.is_primary,
    g.retired,
    g.distance,
    g.brand_name,
    g.model_name,
    g.description,
    COALESCE(g.source, '') AS source,
    COALESCE(g.hashtag, '') AS hashtag,
    g.purchase_price,
    COALESCE(g.purchase_currency, '') AS purchase_currency,
    g.created_at,
    g.updated_at
FROM gear g
WHERE g.athlete_id = ?1 AND g.retired = 0
ORDER BY g.is_primary DESC, g.name;

-- name: GetGearDetail :one
-- Get detailed gear information.
SELECT
    g.id,
    g.name,
    g.is_primary,
    g.retired,
    g.distance,
    g.brand_name,
    g.model_name,
    g.description,
    COALESCE(g.source, '') AS source,
    COALESCE(g.hashtag, '') AS hashtag,
    g.purchase_price,
    COALESCE(g.purchase_currency, '') AS purchase_currency,
    (SELECT COUNT(*) FROM activities WHERE gear_id = g.id) AS activity_count
FROM gear g
WHERE g.id = ?1 AND g.athlete_id = ?2;

-- name: GetCustomGear :many
-- Get custom (user-created) gear.
SELECT
    g.id,
    g.name,
    g.is_primary,
    g.retired,
    g.distance,
    g.brand_name,
    g.model_name,
    g.description,
    COALESCE(g.source, '') AS source,
    COALESCE(g.hashtag, '') AS hashtag,
    g.purchase_price,
    COALESCE(g.purchase_currency, '') AS purchase_currency,
    (SELECT COUNT(*) FROM activities WHERE gear_id = g.id) AS activity_count
FROM gear g
WHERE g.athlete_id = ?1 AND g.source = 'custom'
ORDER BY g.name;

-- name: GetGearMonthlyUsage :many
-- Get monthly usage statistics for gear.
SELECT
    gear_id,
    gear_name,
    source,
    hashtag,
    retired,
    purchase_price,
    purchase_currency,
    month,
    activity_count,
    distance,
    moving_time
FROM v_gear_monthly_usage
WHERE athlete_id = ?1 AND month IS NOT NULL
ORDER BY month DESC, gear_name;

-- name: GetGearTotalDistance :one
-- Get total distance for a piece of gear.
SELECT COALESCE(SUM(distance), 0) AS total_distance
FROM activities
WHERE gear_id = ?1;

-- name: GetGearActivityCount :one
-- Get activity count for a piece of gear.
SELECT COUNT(*) AS count
FROM activities
WHERE gear_id = ?1;
