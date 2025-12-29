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
  importProgress = {
    status: 'running',
    total: 0,
    imported: 0,
    skipped: 0,
    failed: 0,
  }

  const db = getDatabase()
  const athleteId = athlete.id

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

        // Skip if already imported
        if (existingIds.has(activity.id)) {
          importProgress.skipped++
          continue
        }

        try {
          // Store activity
          await storeActivity(db, athleteId, activity)

          // Optionally fetch streams
          if (options.includeStreams) {
            await fetchAndStoreStreams(db, activity.id)
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
  } catch (error) {
    importProgress.status = 'error'
    importProgress.error = error instanceof Error ? error.message : 'Unknown error'
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
 */
async function fetchAndStoreStreams(
  db: ReturnType<typeof getDatabase>,
  activityId: number
): Promise<void> {
  try {
    const streamTypes = 'time,distance,latlng,altitude,heartrate,cadence,watts,temp'
    const streams = await stravaFetch<StravaStream[]>(
      `/activities/${activityId}/streams?keys=${streamTypes}&key_by_type=false`
    )

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
  } catch {
    // Streams may not be available for all activities
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
