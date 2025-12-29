/**
 * Browser Activity Importer - Fetches activities from Strava and stores locally.
 *
 * Uses a phased approach matching the Go backend:
 * Phase 1: Activities - fetch all activity metadata
 * Phase 2: Gear - fetch gear details
 * Phase 3: Streams - fetch activity streams
 * Phase 4: Activity Details - fetch detailed activities with segment efforts
 * Phase 5: Segment Details - fetch full segment info for incomplete segments
 * Phase 6: Photos - fetch activity photos
 */

import { stravaFetch, getAthlete } from './client'
import { getDatabase } from '../db'

// ============================================================================
// Types
// ============================================================================

export type ImportPhase =
  | 'idle'
  | 'activities'
  | 'gear'
  | 'streams'
  | 'details'
  | 'segments'
  | 'photos'
  | 'complete'

export interface ImportOptions {
  fullSync?: boolean
  skipStreams?: boolean
  skipSegments?: boolean
  skipPhotos?: boolean
}

export interface ImportProgress {
  status: 'idle' | 'running' | 'complete' | 'error' | 'cancelled'
  phase: ImportPhase
  error?: string
  sync_run_id?: number

  // Per-phase progress (matches API format)
  activities_total: number
  activities_done: number
  gear_total: number
  gear_done: number
  streams_total: number
  streams_done: number
  details_total: number
  details_done: number
  segments_total: number
  segments_done: number
  photos_total: number
  photos_done: number

  // Legacy/aggregate fields
  failed_count: number
}

export interface SyncRun {
  id: number
  athlete_id: number
  started_at: string
  completed_at?: string
  duration_seconds?: number
  status: string
  error?: string
  activities_total: number
  activities_imported: number
  activities_skipped: number
  streams_imported: number
  failed_count: number
  full_sync: boolean
  skip_streams: boolean
  newest_activity_date?: string
}

// Internal state for tracking sync progress across phases
interface SyncState {
  activityIds: number[]
  gearIds: Set<string>
  activitiesWithPhotos: number[]
  segmentIdsToFetch: Set<number>
  newestActivityDate: string | null
}

// Strava API types
interface StravaActivity {
  id: number
  name: string
  sport_type: string
  type: string
  start_date: string
  start_date_local: string
  timezone: string
  distance: number
  moving_time: number
  elapsed_time: number
  total_elevation_gain: number
  average_speed: number
  max_speed: number
  average_heartrate?: number
  max_heartrate?: number
  average_watts?: number
  max_watts?: number
  weighted_average_watts?: number
  kilojoules?: number
  average_cadence?: number
  calories?: number
  suffer_score?: number
  gear_id?: string
  commute?: boolean
  workout_type?: number
  location_city?: string
  location_state?: string
  location_country?: string
  map?: { summary_polyline?: string }
  start_latlng?: [number, number]
  description?: string
  device_name?: string
  embed_token?: string
  trainer?: boolean
  private?: boolean
  kudos_count?: number
  photo_count?: number
}

interface StravaGear {
  id: string
  name: string
  primary: boolean
  retired: boolean
  distance: number
  brand_name?: string
  model_name?: string
  description?: string
}

interface StravaStream {
  type: string
  data: unknown[]
  series_type: string
  original_size: number
  resolution: string
}

interface StravaSegment {
  id: number
  name: string
  activity_type: string
  distance: number
  average_grade: number
  maximum_grade: number
  elevation_high: number
  elevation_low: number
  climb_category: number
  start_latlng?: [number, number]
  end_latlng?: [number, number]
  starred: boolean
  map?: { polyline?: string }
  athlete_segment_stats?: {
    pr_elapsed_time?: number
    pr_date?: string
    effort_count?: number
  }
}

interface StravaSegmentEffort {
  id: number
  segment: StravaSegment
  name: string
  activity: { id: number }
  athlete: { id: number }
  elapsed_time: number
  moving_time: number
  start_date: string
  start_date_local: string
  distance: number
  average_watts?: number
  average_heartrate?: number
  max_heartrate?: number
  pr_rank?: number
}

interface StravaBestEffort {
  id: number
  name: string
  elapsed_time: number
  moving_time: number
  start_date: string
  start_date_local: string
  distance: number
  pr_rank?: number
  start_index?: number
  end_index?: number
}

interface StravaDetailedActivity {
  id: number
  segment_efforts?: StravaSegmentEffort[]
  best_efforts?: StravaBestEffort[]
}

interface StravaPhoto {
  unique_id: string
  activity_id: number
  urls: { [key: string]: string }
  caption?: string
  location?: [number, number]
  created_at: string
}

// ============================================================================
// Global State
// ============================================================================

let importProgress: ImportProgress = {
  status: 'idle',
  phase: 'idle',
  activities_total: 0,
  activities_done: 0,
  gear_total: 0,
  gear_done: 0,
  streams_total: 0,
  streams_done: 0,
  details_total: 0,
  details_done: 0,
  segments_total: 0,
  segments_done: 0,
  photos_total: 0,
  photos_done: 0,
  failed_count: 0,
}

let cancelRequested = false

// ============================================================================
// Public API
// ============================================================================

export function getImportProgress(): ImportProgress {
  return { ...importProgress }
}

export function cancelImport(): void {
  if (importProgress.status === 'running') {
    cancelRequested = true
  }
}

export function getSyncHistory(limit = 10): SyncRun[] {
  const db = getDatabase()
  const athlete = getAthlete()
  if (!athlete) return []

  const rows = db.query<{
    id: number
    athlete_id: number
    started_at: string
    completed_at: string | null
    duration_seconds: number | null
    status: string
    error: string | null
    activities_total: number
    activities_imported: number
    activities_skipped: number
    streams_imported: number
    failed_count: number
    full_sync: number
    skip_streams: number
    newest_activity_date: string | null
  }>(
    `SELECT id, athlete_id, started_at, completed_at, duration_seconds,
            status, error, activities_total, activities_imported,
            activities_skipped, streams_imported, failed_count,
            full_sync, skip_streams, newest_activity_date
     FROM sync_history
     WHERE athlete_id = ?
     ORDER BY started_at DESC
     LIMIT ?`,
    [athlete.id, limit]
  )

  return rows.map((r) => ({
    id: r.id,
    athlete_id: r.athlete_id,
    started_at: r.started_at,
    completed_at: r.completed_at ?? undefined,
    duration_seconds: r.duration_seconds ?? undefined,
    status: r.status,
    error: r.error ?? undefined,
    activities_total: r.activities_total,
    activities_imported: r.activities_imported,
    activities_skipped: r.activities_skipped,
    streams_imported: r.streams_imported,
    failed_count: r.failed_count,
    full_sync: r.full_sync === 1,
    skip_streams: r.skip_streams === 1,
    newest_activity_date: r.newest_activity_date ?? undefined,
  }))
}

export function getLatestSync(): SyncRun | null {
  const history = getSyncHistory(1)
  return history.length > 0 ? history[0] : null
}

/**
 * Start importing activities from Strava using phased approach.
 */
export async function startImport(options: ImportOptions = {}): Promise<void> {
  console.log('[Import] Starting phased import with options:', JSON.stringify(options))

  if (importProgress.status === 'running') {
    throw new Error('Import already running')
  }

  const athlete = getAthlete()
  if (!athlete) {
    throw new Error('Not authenticated')
  }

  cancelRequested = false
  const db = getDatabase()
  const athleteId = athlete.id
  const startedAt = new Date().toISOString()

  // Create sync history record
  let syncRunId: number | null = null
  try {
    db.exec(
      `INSERT INTO sync_history (
        athlete_id, started_at, status, full_sync, skip_streams
      ) VALUES (?, ?, 'running', ?, ?)`,
      [athleteId, startedAt, options.fullSync ? 1 : 0, options.skipStreams ? 1 : 0]
    )
    const result = db.queryOne<{ id: number }>('SELECT last_insert_rowid() as id')
    syncRunId = result?.id ?? null
  } catch (err) {
    console.error('[Import] Failed to create sync history record:', err)
  }

  // Initialize progress
  importProgress = {
    status: 'running',
    phase: 'activities',
    sync_run_id: syncRunId ?? undefined,
    activities_total: 0,
    activities_done: 0,
    gear_total: 0,
    gear_done: 0,
    streams_total: 0,
    streams_done: 0,
    details_total: 0,
    details_done: 0,
    segments_total: 0,
    segments_done: 0,
    photos_total: 0,
    photos_done: 0,
    failed_count: 0,
  }

  // Initialize state
  const state: SyncState = {
    activityIds: [],
    gearIds: new Set(),
    activitiesWithPhotos: [],
    segmentIdsToFetch: new Set(),
    newestActivityDate: null,
  }

  // Get existing activity IDs if not full sync
  const existingIds = new Set<number>()
  if (!options.fullSync) {
    const rows = db.query<{ id: number }>('SELECT id FROM activities WHERE athlete_id = ?', [athleteId])
    for (const row of rows) {
      existingIds.add(row.id)
    }
  }

  try {
    // Phase 1: Activities
    await runActivitiesPhase(db, athleteId, existingIds, state)
    if (cancelRequested) throw new Error('Cancelled')
    await db.persist()

    // Phase 2: Gear
    await runGearPhase(db, athleteId, state)
    if (cancelRequested) throw new Error('Cancelled')
    await db.persist()

    // Phase 3: Streams
    if (!options.skipStreams) {
      await runStreamsPhase(db, state)
      if (cancelRequested) throw new Error('Cancelled')
      await db.persist()
    }

    // Phase 4: Activity Details (segment efforts + best efforts)
    if (!options.skipSegments) {
      await runActivityDetailsPhase(db, athleteId, state)
      if (cancelRequested) throw new Error('Cancelled')
      await db.persist()
    }

    // Phase 5: Segment Details (for incomplete segments)
    if (!options.skipSegments && state.segmentIdsToFetch.size > 0) {
      await runSegmentDetailsPhase(db, state)
      if (cancelRequested) throw new Error('Cancelled')
      await db.persist()
    }

    // Phase 6: Photos
    if (!options.skipPhotos) {
      await runPhotosPhase(db, athleteId, state)
      if (cancelRequested) throw new Error('Cancelled')
      await db.persist()
    }

    // Complete
    importProgress.status = 'complete'
    importProgress.phase = 'complete'

    // Update sync history
    if (syncRunId) {
      const completedAt = new Date().toISOString()
      const durationSeconds = Math.floor((new Date(completedAt).getTime() - new Date(startedAt).getTime()) / 1000)
      db.exec(
        `UPDATE sync_history SET
          completed_at = ?, duration_seconds = ?, status = 'completed',
          activities_total = ?, activities_imported = ?, activities_skipped = ?,
          streams_imported = ?, failed_count = ?, newest_activity_date = ?
        WHERE id = ?`,
        [
          completedAt,
          durationSeconds,
          importProgress.activities_total,
          importProgress.activities_done,
          0, // skipped is calculated differently now
          importProgress.streams_done,
          importProgress.failed_count,
          state.newestActivityDate,
          syncRunId,
        ]
      )
      await db.persist()
    }
  } catch (error) {
    const errorMsg = error instanceof Error ? error.message : 'Unknown error'

    if (errorMsg === 'Cancelled') {
      importProgress.status = 'cancelled'
    } else {
      importProgress.status = 'error'
      importProgress.error = errorMsg
    }

    // Update sync history with failure/cancellation
    if (syncRunId) {
      const completedAt = new Date().toISOString()
      const durationSeconds = Math.floor((new Date(completedAt).getTime() - new Date(startedAt).getTime()) / 1000)
      db.exec(
        `UPDATE sync_history SET
          completed_at = ?, duration_seconds = ?, status = ?,
          error = ?, activities_total = ?, activities_imported = ?,
          activities_skipped = ?, streams_imported = ?, failed_count = ?
        WHERE id = ?`,
        [
          completedAt,
          durationSeconds,
          importProgress.status === 'cancelled' ? 'canceled' : 'failed',
          importProgress.error ?? null,
          importProgress.activities_total,
          importProgress.activities_done,
          0,
          importProgress.streams_done,
          importProgress.failed_count,
          syncRunId,
        ]
      )
      try {
        await db.persist()
      } catch {
        // Ignore persist error during failure handling
      }
    }

    if (errorMsg !== 'Cancelled') {
      throw error
    }
  }
}

// ============================================================================
// Phase 1: Activities
// ============================================================================

async function runActivitiesPhase(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  existingIds: Set<number>,
  state: SyncState
): Promise<void> {
  importProgress.phase = 'activities'
  console.log('[Import] Phase 1: Fetching activities')

  let page = 1
  const perPage = 100
  let hasMore = true

  while (hasMore && !cancelRequested) {
    const activities = await stravaFetch<StravaActivity[]>(`/athlete/activities?page=${page}&per_page=${perPage}`)

    if (activities.length === 0) {
      hasMore = false
      break
    }

    importProgress.activities_total += activities.length

    for (const activity of activities) {
      if (cancelRequested) break

      // Track newest activity date
      if (!state.newestActivityDate || activity.start_date > state.newestActivityDate) {
        state.newestActivityDate = activity.start_date
      }

      // Collect activity ID for later phases
      state.activityIds.push(activity.id)

      // Collect gear ID
      if (activity.gear_id) {
        state.gearIds.add(activity.gear_id)
      }

      // Track activities with photos
      if (activity.photo_count && activity.photo_count > 0) {
        state.activitiesWithPhotos.push(activity.id)
      }

      const activityExists = existingIds.has(activity.id)

      try {
        if (!activityExists) {
          storeActivity(db, athleteId, activity)
        }
        // Count as done regardless of new or skipped
        importProgress.activities_done++
      } catch (error) {
        console.error(`Failed to store activity ${activity.id}:`, error)
        importProgress.failed_count++
      }
    }

    page++
    await sleep(100) // Rate limiting
  }

  console.log(`[Import] Phase 1 complete: ${state.activityIds.length} activities, ${state.gearIds.size} unique gear`)
}

// ============================================================================
// Phase 2: Gear
// ============================================================================

async function runGearPhase(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  state: SyncState
): Promise<void> {
  if (state.gearIds.size === 0) {
    console.log('[Import] Phase 2: No gear to fetch')
    return
  }

  importProgress.phase = 'gear'
  importProgress.gear_total = state.gearIds.size
  console.log(`[Import] Phase 2: Fetching ${state.gearIds.size} gear items`)

  for (const gearId of state.gearIds) {
    if (cancelRequested) break

    try {
      const gear = await stravaFetch<StravaGear>(`/gear/${gearId}`)
      storeGear(db, athleteId, gear)
      importProgress.gear_done++
      await sleep(100)
    } catch (error) {
      console.error(`Failed to fetch gear ${gearId}:`, error)
      importProgress.failed_count++
    }
  }

  console.log('[Import] Phase 2 complete')
}

// ============================================================================
// Phase 3: Streams
// ============================================================================

async function runStreamsPhase(db: ReturnType<typeof getDatabase>, state: SyncState): Promise<void> {
  importProgress.phase = 'streams'
  importProgress.streams_total = state.activityIds.length
  console.log(`[Import] Phase 3: Fetching streams for ${state.activityIds.length} activities`)

  for (const activityId of state.activityIds) {
    if (cancelRequested) break

    // Check if we already have streams
    const hasStreams = db.queryOne<{ count: number }>(
      'SELECT COUNT(*) as count FROM activity_streams WHERE activity_id = ?',
      [activityId]
    )

    if (hasStreams && hasStreams.count > 0) {
      importProgress.streams_done++
      continue
    }

    try {
      await fetchAndStoreStreams(db, activityId)
      importProgress.streams_done++
      await sleep(200)
    } catch (error) {
      console.error(`Failed to fetch streams for activity ${activityId}:`, error)
      importProgress.streams_done++ // Still count as done (skipped)
    }
  }

  console.log(`[Import] Phase 3 complete: ${importProgress.streams_done} streams`)
}

// ============================================================================
// Phase 4: Activity Details (segment efforts + best efforts)
// ============================================================================

async function runActivityDetailsPhase(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  state: SyncState
): Promise<void> {
  importProgress.phase = 'details'
  importProgress.details_total = state.activityIds.length
  console.log(`[Import] Phase 4: Fetching activity details for ${state.activityIds.length} activities`)

  for (const activityId of state.activityIds) {
    if (cancelRequested) break

    // Check if we already have segment efforts for this activity
    const hasEfforts = db.queryOne<{ count: number }>(
      'SELECT COUNT(*) as count FROM segment_efforts WHERE activity_id = ?',
      [activityId]
    )

    if (hasEfforts && hasEfforts.count > 0) {
      importProgress.details_done++
      continue
    }

    try {
      const detailed = await stravaFetch<StravaDetailedActivity>(`/activities/${activityId}`)

      // Store segment efforts
      if (detailed.segment_efforts && detailed.segment_efforts.length > 0) {
        for (const effort of detailed.segment_efforts) {
          const seg = effort.segment

          // Store segment (basic info from effort)
          storeSegmentFromEffort(db, seg)

          // Check if segment needs full detail fetch
          if (shouldFetchSegmentDetail(seg)) {
            state.segmentIdsToFetch.add(seg.id)
          }

          // Store segment effort
          storeSegmentEffort(db, athleteId, activityId, effort)
        }
      }

      // Store best efforts
      if (detailed.best_efforts && detailed.best_efforts.length > 0) {
        for (const effort of detailed.best_efforts) {
          storeBestEffort(db, athleteId, activityId, effort)
        }
      }

      importProgress.details_done++
      await sleep(200)
    } catch (error) {
      console.error(`Failed to fetch details for activity ${activityId}:`, error)
      importProgress.details_done++ // Still count as processed
    }
  }

  console.log(`[Import] Phase 4 complete: ${importProgress.details_done} activities processed, ${state.segmentIdsToFetch.size} segments need detail`)
}

// ============================================================================
// Phase 5: Segment Details
// ============================================================================

async function runSegmentDetailsPhase(db: ReturnType<typeof getDatabase>, state: SyncState): Promise<void> {
  importProgress.phase = 'segments'
  importProgress.segments_total = state.segmentIdsToFetch.size
  console.log(`[Import] Phase 5: Fetching details for ${state.segmentIdsToFetch.size} segments`)

  for (const segmentId of state.segmentIdsToFetch) {
    if (cancelRequested) break

    try {
      const segment = await stravaFetch<StravaSegment>(`/segments/${segmentId}`)
      storeSegmentDetail(db, segment)
      importProgress.segments_done++
      await sleep(200)
    } catch (error) {
      console.error(`Failed to fetch segment ${segmentId}:`, error)
      importProgress.segments_done++ // Still count as processed
    }
  }

  console.log(`[Import] Phase 5 complete: ${importProgress.segments_done} segments`)
}

// ============================================================================
// Phase 6: Photos
// ============================================================================

async function runPhotosPhase(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  state: SyncState
): Promise<void> {
  if (state.activitiesWithPhotos.length === 0) {
    console.log('[Import] Phase 6: No activities with photos')
    return
  }

  importProgress.phase = 'photos'
  importProgress.photos_total = state.activitiesWithPhotos.length
  console.log(`[Import] Phase 6: Fetching photos for ${state.activitiesWithPhotos.length} activities`)

  for (const activityId of state.activitiesWithPhotos) {
    if (cancelRequested) break

    // Check if we already have photos
    const hasPhotos = db.queryOne<{ count: number }>('SELECT COUNT(*) as count FROM photos WHERE activity_id = ?', [
      activityId,
    ])

    if (hasPhotos && hasPhotos.count > 0) {
      importProgress.photos_done++
      continue
    }

    try {
      await fetchAndStorePhotos(db, athleteId, activityId)
      importProgress.photos_done++
      await sleep(200)
    } catch (error) {
      console.error(`Failed to fetch photos for activity ${activityId}:`, error)
      importProgress.photos_done++ // Still count as processed
    }
  }

  console.log(`[Import] Phase 6 complete: ${importProgress.photos_done} activities processed`)
}

// ============================================================================
// Storage Functions
// ============================================================================

function storeActivity(db: ReturnType<typeof getDatabase>, athleteId: number, activity: StravaActivity): void {
  db.exec(
    `INSERT INTO activities (
      id, athlete_id, name, sport_type, start_date, start_date_local,
      timezone, distance, moving_time, elapsed_time, total_elevation_gain,
      average_speed, max_speed, average_heartrate, max_heartrate,
      average_watts, max_watts, weighted_average_watts, kilojoules,
      average_cadence, calories, suffer_score, gear_id, commute,
      workout_type, location_city, location_state, location_country,
      summary_polyline, start_lat, start_lng, description, device_name,
      embed_token, trainer, private, kudos_count, photo_count
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT (id) DO UPDATE SET
      name = EXCLUDED.name, sport_type = EXCLUDED.sport_type,
      distance = EXCLUDED.distance, moving_time = EXCLUDED.moving_time,
      elapsed_time = EXCLUDED.elapsed_time, total_elevation_gain = EXCLUDED.total_elevation_gain,
      average_speed = EXCLUDED.average_speed, max_speed = EXCLUDED.max_speed,
      average_heartrate = EXCLUDED.average_heartrate, max_heartrate = EXCLUDED.max_heartrate,
      average_watts = EXCLUDED.average_watts, max_watts = EXCLUDED.max_watts,
      weighted_average_watts = EXCLUDED.weighted_average_watts, kilojoules = EXCLUDED.kilojoules,
      average_cadence = EXCLUDED.average_cadence, calories = EXCLUDED.calories,
      suffer_score = EXCLUDED.suffer_score, gear_id = EXCLUDED.gear_id,
      summary_polyline = EXCLUDED.summary_polyline,
      kudos_count = EXCLUDED.kudos_count, photo_count = EXCLUDED.photo_count`,
    [
      activity.id,
      athleteId,
      activity.name,
      activity.sport_type || activity.type,
      activity.start_date,
      activity.start_date_local,
      activity.timezone,
      activity.distance,
      activity.moving_time,
      activity.elapsed_time,
      activity.total_elevation_gain,
      activity.average_speed,
      activity.max_speed,
      activity.average_heartrate ?? null,
      activity.max_heartrate ?? null,
      activity.average_watts ?? null,
      activity.max_watts ?? null,
      activity.weighted_average_watts ?? null,
      activity.kilojoules ?? null,
      activity.average_cadence ?? null,
      activity.calories ?? null,
      activity.suffer_score ?? null,
      activity.gear_id ?? null,
      activity.commute ? 1 : 0,
      activity.workout_type ?? null,
      activity.location_city ?? null,
      activity.location_state ?? null,
      activity.location_country ?? null,
      activity.map?.summary_polyline ?? null,
      activity.start_latlng?.[0] ?? null,
      activity.start_latlng?.[1] ?? null,
      activity.description ?? null,
      activity.device_name ?? null,
      activity.embed_token ?? null,
      activity.trainer ? 1 : 0,
      activity.private ? 1 : 0,
      activity.kudos_count ?? 0,
      activity.photo_count ?? 0,
    ]
  )
}

function storeGear(db: ReturnType<typeof getDatabase>, athleteId: number, gear: StravaGear): void {
  db.exec(
    `INSERT INTO gear (id, athlete_id, name, is_primary, retired, distance, brand_name, model_name, description, source)
     VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'strava')
     ON CONFLICT (id) DO UPDATE SET
       name = EXCLUDED.name, is_primary = EXCLUDED.is_primary, retired = EXCLUDED.retired,
       distance = EXCLUDED.distance, brand_name = EXCLUDED.brand_name,
       model_name = EXCLUDED.model_name, description = EXCLUDED.description`,
    [
      gear.id,
      athleteId,
      gear.name,
      gear.primary ? 1 : 0,
      gear.retired ? 1 : 0,
      gear.distance,
      gear.brand_name ?? null,
      gear.model_name ?? null,
      gear.description ?? null,
    ]
  )
}

async function fetchAndStoreStreams(db: ReturnType<typeof getDatabase>, activityId: number): Promise<boolean> {
  try {
    const streamTypes = 'time,distance,latlng,altitude,heartrate,cadence,watts,temp'
    const streams = await stravaFetch<StravaStream[]>(
      `/activities/${activityId}/streams?keys=${streamTypes}&key_by_type=false`
    )

    if (!streams || streams.length === 0) {
      return false
    }

    for (const stream of streams) {
      db.exec(
        `INSERT INTO activity_streams (activity_id, stream_type, data, series_type, original_size, resolution)
         VALUES (?, ?, ?, ?, ?, ?)
         ON CONFLICT (activity_id, stream_type) DO UPDATE SET data = EXCLUDED.data`,
        [activityId, stream.type, JSON.stringify(stream.data), stream.series_type, stream.original_size, stream.resolution]
      )
    }
    return true
  } catch {
    return false
  }
}

function storeSegmentFromEffort(db: ReturnType<typeof getDatabase>, seg: StravaSegment): void {
  db.exec(
    `INSERT INTO segments (
      id, name, activity_type, distance, average_grade, maximum_grade,
      elevation_high, elevation_low, climb_category, start_lat, start_lng,
      end_lat, end_lng, starred
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT (id) DO UPDATE SET
      name = EXCLUDED.name, starred = EXCLUDED.starred`,
    [
      seg.id,
      seg.name,
      seg.activity_type,
      seg.distance,
      seg.average_grade,
      seg.maximum_grade,
      seg.elevation_high,
      seg.elevation_low,
      seg.climb_category,
      seg.start_latlng?.[0] ?? null,
      seg.start_latlng?.[1] ?? null,
      seg.end_latlng?.[0] ?? null,
      seg.end_latlng?.[1] ?? null,
      seg.starred ? 1 : 0,
    ]
  )
}

function shouldFetchSegmentDetail(seg: StravaSegment): boolean {
  // Fetch detail if missing polyline or athlete stats
  return !seg.map?.polyline || !seg.athlete_segment_stats
}

function storeSegmentDetail(db: ReturnType<typeof getDatabase>, seg: StravaSegment): void {
  db.exec(
    `UPDATE segments SET
      polyline = ?,
      athlete_kom_rank = ?,
      athlete_pr_elapsed_time = ?,
      athlete_pr_date = ?,
      athlete_effort_count = ?
    WHERE id = ?`,
    [
      seg.map?.polyline ?? null,
      null, // KOM rank not available from this endpoint
      seg.athlete_segment_stats?.pr_elapsed_time ?? null,
      seg.athlete_segment_stats?.pr_date ?? null,
      seg.athlete_segment_stats?.effort_count ?? null,
      seg.id,
    ]
  )
}

function storeSegmentEffort(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  activityId: number,
  effort: StravaSegmentEffort
): void {
  db.exec(
    `INSERT INTO segment_efforts (
      id, segment_id, activity_id, athlete_id, name,
      elapsed_time, moving_time, start_date, start_date_local,
      distance, average_watts, average_heartrate, max_heartrate, pr_rank
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT (id) DO NOTHING`,
    [
      effort.id,
      effort.segment.id,
      activityId,
      athleteId,
      effort.name,
      effort.elapsed_time,
      effort.moving_time,
      effort.start_date,
      effort.start_date_local,
      effort.distance,
      effort.average_watts ?? null,
      effort.average_heartrate ?? null,
      effort.max_heartrate ?? null,
      effort.pr_rank ?? null,
    ]
  )
}

function storeBestEffort(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  activityId: number,
  effort: StravaBestEffort
): void {
  // Map effort name to distance type
  const distanceType = effort.name.toLowerCase().replace(/\s+/g, '_')

  db.exec(
    `INSERT INTO best_efforts (
      athlete_id, activity_id, sport_type, distance_type, distance_m,
      elapsed_time, start_index, end_index, start_date, pr_rank
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT DO NOTHING`,
    [
      athleteId,
      activityId,
      null, // sport_type determined from activity
      distanceType,
      effort.distance,
      effort.elapsed_time,
      effort.start_index ?? null,
      effort.end_index ?? null,
      effort.start_date,
      effort.pr_rank ?? null,
    ]
  )
}

async function fetchAndStorePhotos(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  activityId: number
): Promise<number> {
  try {
    const photos = await stravaFetch<StravaPhoto[]>(`/activities/${activityId}/photos?size=600`)

    if (!photos || photos.length === 0) {
      return 0
    }

    for (const photo of photos) {
      const url = photo.urls['600'] || photo.urls['100'] || Object.values(photo.urls)[0]
      const thumbnailUrl = photo.urls['100'] || url

      if (!url) continue

      db.exec(
        `INSERT INTO photos (id, athlete_id, activity_id, url, thumbnail_url, caption, location, created_at)
         VALUES (?, ?, ?, ?, ?, ?, ?, ?)
         ON CONFLICT (id) DO UPDATE SET
           url = EXCLUDED.url, thumbnail_url = EXCLUDED.thumbnail_url, caption = EXCLUDED.caption`,
        [
          photo.unique_id,
          athleteId,
          activityId,
          url,
          thumbnailUrl,
          photo.caption ?? null,
          photo.location ? JSON.stringify(photo.location) : null,
          photo.created_at,
        ]
      )
    }
    return photos.length
  } catch {
    return 0
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
