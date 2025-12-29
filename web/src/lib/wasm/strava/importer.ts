/**
 * Browser Activity Importer - Fetches activities from Strava and stores locally.
 */

import { stravaFetch, getAthlete } from './client'
import { getDatabase } from '../db'

export interface ImportOptions {
  fullSync?: boolean
  includeStreams?: boolean
}

export interface ImportProgress {
  status: 'idle' | 'running' | 'complete' | 'error' | 'cancelled'
  total: number
  imported: number
  skipped: number
  failed: number
  current_activity?: string
  error?: string
  streams_imported?: number
  sync_run_id?: number
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

// Strava API types (simplified)
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
  map?: {
    summary_polyline?: string
  }
  start_latlng?: [number, number]
  description?: string
  device_name?: string
  embed_token?: string
  trainer?: boolean
  private?: boolean
}

interface StravaStream {
  type: string
  data: unknown[]
  series_type: string
  original_size: number
  resolution: string
}

let importProgress: ImportProgress = {
  status: 'idle',
  total: 0,
  imported: 0,
  skipped: 0,
  failed: 0,
  streams_imported: 0,
}

let cancelRequested = false

/**
 * Get current import progress.
 */
export function getImportProgress(): ImportProgress {
  return { ...importProgress }
}

/**
 * Cancel the running import.
 */
export function cancelImport(): void {
  if (importProgress.status === 'running') {
    cancelRequested = true
  }
}

/**
 * Get sync history for the current athlete.
 */
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

/**
 * Get the latest sync run for the current athlete.
 */
export function getLatestSync(): SyncRun | null {
  const history = getSyncHistory(1)
  return history.length > 0 ? history[0] : null
}

/**
 * Start importing activities from Strava.
 */
export async function startImport(options: ImportOptions = {}): Promise<void> {
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
  let syncRunId: number | null = null
  let newestActivityDate: string | null = null
  let streamsImported = 0

  // Create sync history record
  try {
    db.exec(
      `INSERT INTO sync_history (
        athlete_id, started_at, status, full_sync, skip_streams
      ) VALUES (?, ?, 'running', ?, ?)`,
      [athleteId, startedAt, options.fullSync ? 1 : 0, options.includeStreams ? 0 : 1]
    )
    const result = db.queryOne<{ id: number }>('SELECT last_insert_rowid() as id')
    syncRunId = result?.id ?? null
  } catch (err) {
    console.error('[Import] Failed to create sync history record:', err)
  }

  importProgress = {
    status: 'running',
    total: 0,
    imported: 0,
    skipped: 0,
    failed: 0,
    streams_imported: 0,
    sync_run_id: syncRunId ?? undefined,
  }

  try {
    // Get existing activity IDs if not full sync
    const existingIds = new Set<number>()
    if (!options.fullSync) {
      const rows = db.query<{ id: number }>('SELECT id FROM activities WHERE athlete_id = ?', [
        athleteId,
      ])
      for (const row of rows) {
        existingIds.add(row.id)
      }
    }

    let page = 1
    const perPage = 100
    let hasMore = true

    while (hasMore && !cancelRequested) {
      // Fetch activities page
      const activities = await stravaFetch<StravaActivity[]>(
        `/athlete/activities?page=${page}&per_page=${perPage}`
      )

      if (activities.length === 0) {
        hasMore = false
        break
      }

      importProgress.total += activities.length

      for (const activity of activities) {
        if (cancelRequested) break

        importProgress.current_activity = activity.name

        // Track newest activity date
        if (!newestActivityDate || activity.start_date > newestActivityDate) {
          newestActivityDate = activity.start_date
        }

        // Skip if already imported
        if (existingIds.has(activity.id)) {
          importProgress.skipped++
          continue
        }

        try {
          // Store activity
          await storeActivity(db, athleteId, activity)

          // Fetch streams (enabled by default)
          if (options.includeStreams !== false) {
            const streamsFetched = await fetchAndStoreStreams(db, activity.id)
            if (streamsFetched) {
              streamsImported++
              importProgress.streams_imported = streamsImported
            }
          }

          importProgress.imported++
        } catch (error) {
          console.error(`Failed to import activity ${activity.id}:`, error)
          importProgress.failed++
        }
      }

      page++

      // Small delay to avoid rate limiting
      await sleep(100)
    }

    if (cancelRequested) {
      importProgress.status = 'cancelled'
    } else {
      importProgress.status = 'complete'
    }

    await db.persist()

    // Update sync history record
    if (syncRunId) {
      const completedAt = new Date().toISOString()
      const durationSeconds = Math.floor(
        (new Date(completedAt).getTime() - new Date(startedAt).getTime()) / 1000
      )
      db.exec(
        `UPDATE sync_history SET
          completed_at = ?,
          duration_seconds = ?,
          status = ?,
          activities_total = ?,
          activities_imported = ?,
          activities_skipped = ?,
          streams_imported = ?,
          failed_count = ?,
          newest_activity_date = ?
        WHERE id = ?`,
        [
          completedAt,
          durationSeconds,
          importProgress.status === 'complete' ? 'completed' : 'canceled',
          importProgress.total,
          importProgress.imported,
          importProgress.skipped,
          streamsImported,
          importProgress.failed,
          newestActivityDate,
          syncRunId,
        ]
      )
      await db.persist()
    }
  } catch (error) {
    importProgress.status = 'error'
    importProgress.error = error instanceof Error ? error.message : 'Unknown error'

    // Update sync history with failure
    if (syncRunId) {
      const completedAt = new Date().toISOString()
      const durationSeconds = Math.floor(
        (new Date(completedAt).getTime() - new Date(startedAt).getTime()) / 1000
      )
      db.exec(
        `UPDATE sync_history SET
          completed_at = ?,
          duration_seconds = ?,
          status = 'failed',
          error = ?,
          activities_total = ?,
          activities_imported = ?,
          activities_skipped = ?,
          streams_imported = ?,
          failed_count = ?
        WHERE id = ?`,
        [
          completedAt,
          durationSeconds,
          importProgress.error,
          importProgress.total,
          importProgress.imported,
          importProgress.skipped,
          streamsImported,
          importProgress.failed,
          syncRunId,
        ]
      )
      try {
        await db.persist()
      } catch {
        // Ignore persist error during failure handling
      }
    }

    throw error
  }
}

/**
 * Store an activity in the database.
 */
async function storeActivity(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  activity: StravaActivity
): Promise<void> {
  db.exec(
    `INSERT INTO activities (
      id, athlete_id, name, sport_type, start_date, start_date_local,
      timezone, distance, moving_time, elapsed_time, total_elevation_gain,
      average_speed, max_speed, average_heartrate, max_heartrate,
      average_watts, max_watts, weighted_average_watts, kilojoules,
      average_cadence, calories, suffer_score, gear_id, commute,
      workout_type, location_city, location_state, location_country,
      summary_polyline, start_lat, start_lng, description, device_name,
      embed_token, trainer, private
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    ON CONFLICT (id) DO UPDATE SET
      name = EXCLUDED.name,
      sport_type = EXCLUDED.sport_type,
      distance = EXCLUDED.distance,
      moving_time = EXCLUDED.moving_time,
      elapsed_time = EXCLUDED.elapsed_time,
      total_elevation_gain = EXCLUDED.total_elevation_gain,
      average_speed = EXCLUDED.average_speed,
      max_speed = EXCLUDED.max_speed,
      average_heartrate = EXCLUDED.average_heartrate,
      max_heartrate = EXCLUDED.max_heartrate,
      average_watts = EXCLUDED.average_watts,
      max_watts = EXCLUDED.max_watts,
      weighted_average_watts = EXCLUDED.weighted_average_watts,
      kilojoules = EXCLUDED.kilojoules,
      average_cadence = EXCLUDED.average_cadence,
      calories = EXCLUDED.calories,
      suffer_score = EXCLUDED.suffer_score,
      gear_id = EXCLUDED.gear_id,
      summary_polyline = EXCLUDED.summary_polyline`,
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
    ]
  )
}

/**
 * Fetch and store activity streams.
 * Returns true if streams were successfully fetched and stored.
 */
async function fetchAndStoreStreams(
  db: ReturnType<typeof getDatabase>,
  activityId: number
): Promise<boolean> {
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
         ON CONFLICT (activity_id, stream_type) DO UPDATE SET
           data = EXCLUDED.data`,
        [
          activityId,
          stream.type,
          JSON.stringify(stream.data),
          stream.series_type,
          stream.original_size,
          stream.resolution,
        ]
      )
    }
    return true
  } catch {
    // Streams may not be available for all activities
    return false
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
