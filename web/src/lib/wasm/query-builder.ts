/**
 * Dynamic query builders for WASM mode.
 *
 * These mirror the dynamic query building logic in Go storage layer,
 * allowing WASM mode to support all the same filters as server mode.
 *
 * @see internal/storage/activities.go:216-268 for Go equivalent
 */

import type { WasmDatabase } from './db'
import type { ActivityFilters, HeatmapFilters, SegmentsFilters } from '../data/types'

interface QueryResult<T> {
  data: T[]
  total: number
}

/**
 * Activity columns matching schema/queries/activities.sql GetActivities query.
 */
const ACTIVITY_COLUMNS = `
  id,
  athlete_id,
  name,
  sport_type,
  start_date,
  start_date_local,
  timezone,
  distance,
  moving_time,
  elapsed_time,
  total_elevation_gain,
  average_speed,
  max_speed,
  average_heartrate,
  max_heartrate,
  average_watts,
  max_watts,
  weighted_average_watts,
  kilojoules,
  average_cadence,
  calories,
  suffer_score,
  gear_id,
  commute,
  workout_type,
  location_city,
  location_state,
  location_country,
  summary_polyline,
  start_lat,
  start_lng
`

/**
 * Build dynamic activity queries matching Go's logic in activities.go:216-268.
 *
 * Supports all filters:
 * - sport_type: Filter by sport type
 * - after: Activities after this date (start_date >= ?)
 * - before: Activities before this date (start_date <= ?)
 * - gear_id: Filter by gear ID
 * - commute: Filter commute activities
 * - trainer: Filter trainer/indoor activities
 * - search: Name search (LIKE %search%)
 */
export function queryActivities<T extends Record<string, unknown>>(
  db: WasmDatabase,
  athleteId: number,
  filters: ActivityFilters,
  limit: number,
  offset: number
): QueryResult<T> {
  const conditions: string[] = ['athlete_id = ?']
  const args: unknown[] = [athleteId]

  // Match Go's filter logic exactly (activities.go:226-263)
  if (filters.sport_type) {
    conditions.push('sport_type = ?')
    args.push(filters.sport_type)
  }

  if (filters.after) {
    conditions.push('start_date >= ?')
    args.push(filters.after)
  }

  if (filters.before) {
    conditions.push('start_date <= ?')
    args.push(filters.before)
  }

  if (filters.gear_id) {
    conditions.push('gear_id = ?')
    args.push(filters.gear_id)
  }

  if (filters.commute !== undefined) {
    conditions.push('commute = ?')
    args.push(filters.commute ? 1 : 0)
  }

  if (filters.trainer !== undefined) {
    conditions.push('trainer = ?')
    args.push(filters.trainer ? 1 : 0)
  }

  if (filters.search) {
    conditions.push('name LIKE ? COLLATE NOCASE')
    args.push(`%${filters.search}%`)
  }

  const where = conditions.join(' AND ')

  // Count query
  const countSql = `SELECT COUNT(*) as count FROM activities WHERE ${where}`
  const countResult = db.queryOne<{ count: number }>(countSql, args)
  const total = countResult?.count ?? 0

  // Data query with order by
  const orderBy = buildOrderClause(filters.order_by, filters.order_dir, 'start_date DESC')

  const dataSql = `
    SELECT ${ACTIVITY_COLUMNS}
    FROM activities
    WHERE ${where}
    ORDER BY ${orderBy}
    LIMIT ? OFFSET ?
  `
  const data = db.query<T>(dataSql, [...args, limit, offset])

  return { data, total }
}

/**
 * Heatmap columns for polyline data.
 * Matches the columns expected by HeatmapRow in provider.ts
 */
const HEATMAP_COLUMNS = `
  id,
  name,
  sport_type,
  start_date,
  distance,
  summary_polyline,
  start_lat,
  start_lng
`

/**
 * Build dynamic heatmap queries matching Go's logic in heatmap.go.
 *
 * Supports filters:
 * - sport_type: Filter by sport type
 * - after: Activities after this date
 * - before: Activities before this date
 * - commute: Filter commute activities
 * - workout_type: Filter by workout type
 */
export function queryHeatmapActivities<T extends Record<string, unknown>>(
  db: WasmDatabase,
  athleteId: number,
  filters: HeatmapFilters
): T[] {
  const conditions: string[] = [
    'athlete_id = ?',
    'summary_polyline IS NOT NULL',
    "summary_polyline != ''",
  ]
  const args: unknown[] = [athleteId]

  if (filters.sport_type) {
    conditions.push('sport_type = ?')
    args.push(filters.sport_type)
  }

  if (filters.after) {
    conditions.push('start_date >= ?')
    args.push(filters.after)
  }

  if (filters.before) {
    conditions.push('start_date <= ?')
    args.push(filters.before)
  }

  if (filters.commute !== undefined) {
    conditions.push('commute = ?')
    args.push(filters.commute ? 1 : 0)
  }

  if (filters.workout_type !== undefined) {
    conditions.push('workout_type = ?')
    args.push(filters.workout_type)
  }

  const where = conditions.join(' AND ')

  const sql = `
    SELECT ${HEATMAP_COLUMNS}
    FROM activities
    WHERE ${where}
  `

  return db.query<T>(sql, args)
}

// Allowed columns for ordering to prevent SQL injection
const ALLOWED_ORDER_COLUMNS = new Set([
  'start_date',
  'start_date_local',
  'name',
  'distance',
  'moving_time',
  'elapsed_time',
  'total_elevation_gain',
  'average_speed',
  'max_speed',
  'average_heartrate',
  'max_heartrate',
  'average_watts',
  'max_watts',
  'calories',
])

/**
 * Build safe ORDER BY clause with validation.
 */
function buildOrderClause(
  orderBy: string | undefined,
  orderDir: 'asc' | 'desc' | undefined,
  defaultOrder: string
): string {
  if (!orderBy || !ALLOWED_ORDER_COLUMNS.has(orderBy)) {
    return defaultOrder
  }

  const dir = orderDir === 'asc' ? 'ASC' : 'DESC'
  return `${orderBy} ${dir}`
}

/**
 * Segment columns from v_segment_stats view.
 */
const SEGMENT_COLUMNS = `
  id,
  name,
  activity_type,
  distance,
  average_grade,
  maximum_grade,
  elevation_high,
  elevation_low,
  climb_category,
  starred,
  athlete_kom_rank,
  athlete_effort_count,
  athlete_pr_elapsed_time,
  athlete_pr_date,
  times_completed,
  last_effort_date,
  best_elapsed_time
`

// Allowed columns for segment ordering
const ALLOWED_SEGMENT_ORDER_COLUMNS: Record<string, string> = {
  name: 'name',
  distance: 'distance',
  maximum_grade: 'maximum_grade',
  times_completed: 'times_completed',
  last_effort_date: 'last_effort_date',
  best_elapsed_time: 'best_elapsed_time',
}

/**
 * Build dynamic segment queries matching Go's logic in segments.go:240-356.
 *
 * Supports all filters:
 * - activity_type: Filter by activity type (Ride, Run, etc.)
 * - country: Filter by country (requires JOIN with segment_efforts)
 * - starred: Filter starred segments only
 * - kom_only: Filter to segments where athlete has KOM
 * - search: Name search (LIKE %search%)
 */
export function querySegments<T extends Record<string, unknown>>(
  db: WasmDatabase,
  athleteId: number,
  filters: SegmentsFilters | undefined,
  limit: number,
  offset: number
): QueryResult<T> {
  // For country filter, we need to JOIN with segment_efforts
  // Otherwise, we can query directly from v_segment_stats view
  const hasCountryFilter = filters?.country

  if (hasCountryFilter) {
    return querySegmentsWithCountry<T>(db, athleteId, filters!, limit, offset)
  }

  const conditions: string[] = ['athlete_id = ?']
  const args: unknown[] = [athleteId]

  // Match Go's filter logic (segments.go:248-266)
  if (filters?.activity_type) {
    conditions.push('activity_type = ?')
    args.push(filters.activity_type)
  }

  if (filters?.starred !== undefined) {
    conditions.push('starred = ?')
    args.push(filters.starred ? 1 : 0)
  }

  if (filters?.kom_only) {
    conditions.push('athlete_kom_rank = 1')
  }

  if (filters?.search) {
    conditions.push('name LIKE ? COLLATE NOCASE')
    args.push(`%${filters.search}%`)
  }

  const where = conditions.join(' AND ')

  // Count query
  const countSql = `SELECT COUNT(*) as count FROM v_segment_stats WHERE ${where}`
  const countResult = db.queryOne<{ count: number }>(countSql, args)
  const total = countResult?.count ?? 0

  // Build ORDER BY
  const orderBy = buildSegmentOrderClause(
    filters?.order_by,
    filters?.order_dir,
    'times_completed DESC, distance DESC'
  )

  const dataSql = `
    SELECT ${SEGMENT_COLUMNS}
    FROM v_segment_stats
    WHERE ${where}
    ORDER BY ${orderBy}
    LIMIT ? OFFSET ?
  `
  const data = db.query<T>(dataSql, [...args, limit, offset])

  return { data, total }
}

/**
 * Query segments with country filter (requires JOIN).
 */
function querySegmentsWithCountry<T extends Record<string, unknown>>(
  db: WasmDatabase,
  athleteId: number,
  filters: SegmentsFilters,
  limit: number,
  offset: number
): QueryResult<T> {
  const conditions: string[] = ['se.athlete_id = ?', 'se.country = ?']
  const args: unknown[] = [athleteId, filters.country]

  if (filters.activity_type) {
    conditions.push('s.activity_type = ?')
    args.push(filters.activity_type)
  }

  if (filters.starred !== undefined) {
    conditions.push('s.starred = ?')
    args.push(filters.starred ? 1 : 0)
  }

  if (filters.kom_only) {
    conditions.push('s.athlete_kom_rank = 1')
  }

  if (filters.search) {
    conditions.push('s.name LIKE ? COLLATE NOCASE')
    args.push(`%${filters.search}%`)
  }

  const where = conditions.join(' AND ')

  // Count distinct segments
  const countSql = `
    SELECT COUNT(DISTINCT s.id) as count
    FROM segments s
    JOIN segment_efforts se ON se.segment_id = s.id
    WHERE ${where}
  `
  const countResult = db.queryOne<{ count: number }>(countSql, args)
  const total = countResult?.count ?? 0

  // Build ORDER BY
  const orderBy = buildSegmentOrderClause(
    filters.order_by,
    filters.order_dir,
    'vs.times_completed DESC, s.distance DESC'
  )

  // Data query with JOIN
  const dataSql = `
    SELECT DISTINCT
      s.id,
      s.name,
      s.activity_type,
      s.distance,
      s.average_grade,
      s.maximum_grade,
      s.elevation_high,
      s.elevation_low,
      s.climb_category,
      s.starred,
      s.athlete_kom_rank,
      s.athlete_effort_count,
      s.athlete_pr_elapsed_time,
      s.athlete_pr_date,
      vs.times_completed,
      vs.last_effort_date,
      vs.best_elapsed_time
    FROM segments s
    JOIN segment_efforts se ON se.segment_id = s.id
    LEFT JOIN v_segment_stats vs ON vs.id = s.id AND vs.athlete_id = se.athlete_id
    WHERE ${where}
    ORDER BY ${orderBy}
    LIMIT ? OFFSET ?
  `
  const data = db.query<T>(dataSql, [...args, limit, offset])

  return { data, total }
}

/**
 * Build safe ORDER BY clause for segments with validation.
 */
function buildSegmentOrderClause(
  orderBy: string | undefined,
  orderDir: 'asc' | 'desc' | undefined,
  defaultOrder: string
): string {
  if (!orderBy || !ALLOWED_SEGMENT_ORDER_COLUMNS[orderBy]) {
    return defaultOrder
  }

  const column = ALLOWED_SEGMENT_ORDER_COLUMNS[orderBy]
  const dir = orderDir === 'asc' ? 'ASC' : 'DESC'

  // Handle NULL values for computed columns
  if (orderBy === 'last_effort_date' || orderBy === 'best_elapsed_time') {
    return `${column} IS NULL, ${column} ${dir}`
  }

  return `${column} ${dir}`
}
