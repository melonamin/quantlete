-- Gear queries.

-- name: GetGear :many
-- Get all gear for an athlete.
SELECT
    g.id,
    g.name,
    g.is_primary,
    g.retired,
    g.distance,
    g.brand_name,
    g.model_name,
    g.description,
    g.source,
    g.hashtag,
    g.purchase_price,
    g.purchase_currency,
    (SELECT COUNT(*) FROM activities WHERE gear_id = g.id) AS activity_count
FROM gear g
WHERE g.athlete_id = ?1
ORDER BY g.name;

-- name: GetActiveGear :many
-- Get non-retired gear for an athlete.
SELECT
    g.id,
    g.name,
    g.is_primary,
    g.retired,
    g.distance,
    g.brand_name,
    g.model_name,
    g.description,
    g.source,
    g.hashtag,
    g.purchase_price,
    g.purchase_currency,
    (SELECT COUNT(*) FROM activities WHERE gear_id = g.id) AS activity_count
FROM gear g
WHERE g.athlete_id = ?1 AND g.retired = 0
ORDER BY g.name;

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
    g.source,
    g.hashtag,
    g.purchase_price,
    g.purchase_currency,
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
    g.source,
    g.hashtag,
    g.purchase_price,
    g.purchase_currency,
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
