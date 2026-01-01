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
import { getRateLimitInfo } from './ratelimit'
import { createETAEstimator, updateFromRateLimits, estimateCompletion, formatETA, type ETAState } from './eta'
import { rollingMaxAverage, isInitialized as algorithmsInitialized } from '../algorithms'
import {
  saveImportState,
  loadImportState,
  clearImportState,
  hasResumableState,
  type ImportState,
} from './import-state'

// Standard durations for power curve (matching Go backend)
const POWER_DURATIONS = [5, 10, 30, 60, 300, 480, 1200, 3600]

// Rate limiting delays between API calls (milliseconds)
const RATE_LIMIT_DELAY_SHORT_MS = 100 // Used between fast operations (activity list pages, gear)
const RATE_LIMIT_DELAY_LONG_MS = 200 // Used between heavier operations (streams, details, segments, photos)

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
  resume?: boolean
}

export interface ImportProgress {
  status: 'idle' | 'running' | 'complete' | 'error' | 'cancelled' | 'paused'
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

  // ETA estimation
  remaining_api_calls: number
  estimated_eta?: string
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
  remaining_api_calls: 0,
  estimated_eta: undefined,
}

let cancelRequested = false
let pauseRequested = false
let etaState: ETAState = createETAEstimator()

// ============================================================================
// Public API
// ============================================================================

export function getImportProgress(): ImportProgress {
  return { ...importProgress }
}

export function cancelImport(): void {
  if (importProgress.status === 'running' || importProgress.status === 'paused') {
    const wasPaused = importProgress.status === 'paused'
    cancelRequested = true
    pauseRequested = false
    clearImportState()

    // If cancelling from paused state, update sync_history immediately
    // (running state will be updated when the import loop catches the cancel)
    if (wasPaused && importProgress.sync_run_id) {
      const db = getDatabase()
      try {
        const completedAt = new Date().toISOString()
        db.exec(
          `UPDATE sync_history SET
            completed_at = ?, status = 'canceled',
            activities_total = ?, activities_imported = ?,
            streams_imported = ?, failed_count = ?
          WHERE id = ?`,
          [
            completedAt,
            importProgress.activities_total,
            importProgress.activities_done,
            importProgress.streams_done,
            importProgress.failed_count,
            importProgress.sync_run_id,
          ]
        )
      } catch (err) {
        console.error('[Import] Failed to update sync_history on cancel:', err)
      }
      importProgress.status = 'cancelled'
    }
  }
}

export function pauseImport(): void {
  if (importProgress.status === 'running') {
    pauseRequested = true
  }
}

export function resumeImport(): void {
  if (importProgress.status === 'paused') {
    // Resume by calling startImport with resume: true
    // The actual resume is handled by startImport
    pauseRequested = false
  }
}

export function isPausedImport(): boolean {
  return importProgress.status === 'paused'
}

export function hasResumableImport(): boolean {
  return hasResumableState() || importProgress.status === 'paused'
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

// ============================================================================
// ETA Helpers
// ============================================================================

/**
 * Calculate remaining API calls based on current phase and state.
 */
function calculateRemainingAPICalls(
  phase: ImportPhase,
  state: SyncState,
  options: ImportOptions
): number {
  let calls = 0

  // Add calls for remaining phases based on current phase
  switch (phase) {
    case 'activities':
      // Estimate remaining activity pages (we don't know total yet)
      calls += 5 // Conservative estimate for remaining pages
      // Fall through to count remaining phases
      calls += state.gearIds.size
      if (!options.skipStreams) calls += state.activityIds.length - importProgress.streams_done
      if (!options.skipSegments) {
        calls += state.activityIds.length - importProgress.details_done
        calls += state.segmentIdsToFetch.size - importProgress.segments_done
      }
      if (!options.skipPhotos) calls += state.activitiesWithPhotos.length - importProgress.photos_done
      break

    case 'gear':
      calls += state.gearIds.size - importProgress.gear_done
      if (!options.skipStreams) calls += state.activityIds.length - importProgress.streams_done
      if (!options.skipSegments) {
        calls += state.activityIds.length - importProgress.details_done
        calls += state.segmentIdsToFetch.size
      }
      if (!options.skipPhotos) calls += state.activitiesWithPhotos.length
      break

    case 'streams':
      calls += importProgress.streams_total - importProgress.streams_done
      if (!options.skipSegments) {
        calls += state.activityIds.length - importProgress.details_done
        calls += state.segmentIdsToFetch.size
      }
      if (!options.skipPhotos) calls += state.activitiesWithPhotos.length - importProgress.photos_done
      break

    case 'details':
      calls += importProgress.details_total - importProgress.details_done
      if (!options.skipSegments) calls += state.segmentIdsToFetch.size - importProgress.segments_done
      if (!options.skipPhotos) calls += state.activitiesWithPhotos.length - importProgress.photos_done
      break

    case 'segments':
      calls += importProgress.segments_total - importProgress.segments_done
      if (!options.skipPhotos) calls += state.activitiesWithPhotos.length - importProgress.photos_done
      break

    case 'photos':
      calls += importProgress.photos_total - importProgress.photos_done
      break
  }

  return Math.max(0, calls)
}

/**
 * Build import state for persistence.
 */
function buildImportState(
  phase: ImportPhase,
  state: SyncState,
  options: ImportOptions,
  indices: {
    activitiesLastPage?: number
    gearLastIndex?: number
    streamsLastIndex?: number
    detailsLastIndex?: number
    segmentsLastIndex?: number
    photosLastIndex?: number
  } = {}
): ImportState {
  return {
    phase,
    activityIds: state.activityIds,
    gearIds: Array.from(state.gearIds),
    segmentIdsToFetch: Array.from(state.segmentIdsToFetch),
    activitiesWithPhotos: state.activitiesWithPhotos,
    newestActivityDate: state.newestActivityDate,
    activitiesLastPage: indices.activitiesLastPage ?? 0,
    gearLastIndex: indices.gearLastIndex ?? 0,
    streamsLastIndex: indices.streamsLastIndex ?? 0,
    detailsLastIndex: indices.detailsLastIndex ?? 0,
    segmentsLastIndex: indices.segmentsLastIndex ?? 0,
    photosLastIndex: indices.photosLastIndex ?? 0,
    activitiesTotal: importProgress.activities_total,
    activitiesDone: importProgress.activities_done,
    gearTotal: importProgress.gear_total,
    gearDone: importProgress.gear_done,
    streamsTotal: importProgress.streams_total,
    streamsDone: importProgress.streams_done,
    detailsTotal: importProgress.details_total,
    detailsDone: importProgress.details_done,
    segmentsTotal: importProgress.segments_total,
    segmentsDone: importProgress.segments_done,
    photosTotal: importProgress.photos_total,
    photosDone: importProgress.photos_done,
    failedCount: importProgress.failed_count,
    options,
    startedAt: new Date().toISOString(),
    syncRunId: importProgress.sync_run_id,
  }
}

/**
 * Check if pause was requested and handle it.
 * Returns true if paused (caller should exit loop).
 * Note: Does NOT clear pauseRequested - the outer startImport() checks this flag
 * to throw 'Paused' and exit cleanly after the phase loop breaks.
 */
function checkPauseRequested(
  phase: ImportPhase,
  state: SyncState,
  options: ImportOptions,
  currentIndex: number,
  indexKey: 'activitiesLastPage' | 'gearLastIndex' | 'streamsLastIndex' | 'detailsLastIndex' | 'segmentsLastIndex' | 'photosLastIndex'
): boolean {
  if (!pauseRequested) return false

  // Build and save state for resume
  const importState = buildImportState(phase, state, options, { [indexKey]: currentIndex })
  importState.pausedAt = new Date().toISOString()
  saveImportState(importState)

  // Update progress status - pauseRequested stays true so startImport() can detect it
  importProgress.status = 'paused'

  // Update sync_history to reflect paused status
  if (importProgress.sync_run_id) {
    const db = getDatabase()
    try {
      db.exec(
        `UPDATE sync_history SET status = 'paused',
          activities_total = ?, activities_imported = ?,
          streams_imported = ?, failed_count = ?
        WHERE id = ?`,
        [
          importProgress.activities_total,
          importProgress.activities_done,
          importProgress.streams_done,
          importProgress.failed_count,
          importProgress.sync_run_id,
        ]
      )
    } catch (err) {
      console.error('[Import] Failed to update sync_history on pause:', err)
    }
  }

  console.log(`[Import] Paused at phase ${phase}, index ${currentIndex}`)
  return true
}

/**
 * Update ETA estimation based on current progress.
 */
function updateETA(phase: ImportPhase, state: SyncState, options: ImportOptions): void {
  // Update rate limit info
  const rateLimits = getRateLimitInfo()
  updateFromRateLimits(
    etaState,
    rateLimits.usage15Min,
    rateLimits.limit15Min,
    rateLimits.usageDaily,
    rateLimits.limitDaily
  )

  // Calculate remaining calls
  const remaining = calculateRemainingAPICalls(phase, state, options)
  importProgress.remaining_api_calls = remaining

  // Estimate ETA
  if (remaining > 0) {
    const durationMs = estimateCompletion(etaState, remaining)
    importProgress.estimated_eta = formatETA(durationMs)
  } else {
    importProgress.estimated_eta = undefined
  }
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
  pauseRequested = false
  etaState = createETAEstimator()
  const db = getDatabase()
  const athleteId = athlete.id
  const startedAt = new Date().toISOString()

  // Check for resumable state
  const savedState = options.resume ? loadImportState() : null
  const isResuming = savedState !== null

  // Merge options with saved options if resuming
  const effectiveOptions: ImportOptions = isResuming
    ? { ...savedState.options, ...options }
    : options

  // Always create a new sync history record (even when resuming to preserve audit trail)
  // When resuming, the previous sync run retains its terminal state (paused/failed)
  let syncRunId: number | null = null
  try {
    db.exec(
      `INSERT INTO sync_history (
        athlete_id, started_at, status, full_sync, skip_streams
      ) VALUES (?, ?, 'running', ?, ?)`,
      [athleteId, startedAt, effectiveOptions.fullSync ? 1 : 0, effectiveOptions.skipStreams ? 1 : 0]
    )
    const result = db.queryOne<{ id: number }>('SELECT last_insert_rowid() as id')
    syncRunId = result?.id ?? null

    // If resuming, update the old sync run to mark it was resumed
    if (isResuming && savedState.syncRunId) {
      db.exec(
        `UPDATE sync_history SET status = 'resumed', error = ? WHERE id = ? AND status IN ('paused', 'running')`,
        [`Resumed in sync run ${syncRunId}`, savedState.syncRunId]
      )
    }
  } catch (err) {
    console.error('[Import] Failed to create sync history record:', err)
  }

  // Initialize or restore progress
  if (isResuming) {
    console.log('[Import] Resuming from phase:', savedState.phase)
    importProgress = {
      status: 'running',
      phase: savedState.phase,
      sync_run_id: syncRunId ?? undefined,
      activities_total: savedState.activitiesTotal,
      activities_done: savedState.activitiesDone,
      gear_total: savedState.gearTotal,
      gear_done: savedState.gearDone,
      streams_total: savedState.streamsTotal,
      streams_done: savedState.streamsDone,
      details_total: savedState.detailsTotal,
      details_done: savedState.detailsDone,
      segments_total: savedState.segmentsTotal,
      segments_done: savedState.segmentsDone,
      photos_total: savedState.photosTotal,
      photos_done: savedState.photosDone,
      failed_count: savedState.failedCount,
      remaining_api_calls: 0,
      estimated_eta: undefined,
    }
  } else {
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
      remaining_api_calls: 0,
      estimated_eta: undefined,
    }
  }

  // Initialize or restore state
  const state: SyncState = isResuming
    ? {
        activityIds: savedState.activityIds,
        gearIds: new Set(savedState.gearIds),
        activitiesWithPhotos: savedState.activitiesWithPhotos,
        segmentIdsToFetch: new Set(savedState.segmentIdsToFetch),
        newestActivityDate: savedState.newestActivityDate,
      }
    : {
        activityIds: [],
        gearIds: new Set(),
        activitiesWithPhotos: [],
        segmentIdsToFetch: new Set(),
        newestActivityDate: null,
      }

  // Resume indices
  const resumeIndices = isResuming
    ? {
        activitiesLastPage: savedState.activitiesLastPage,
        gearLastIndex: savedState.gearLastIndex,
        streamsLastIndex: savedState.streamsLastIndex,
        detailsLastIndex: savedState.detailsLastIndex,
        segmentsLastIndex: savedState.segmentsLastIndex,
        photosLastIndex: savedState.photosLastIndex,
      }
    : { activitiesLastPage: 0, gearLastIndex: 0, streamsLastIndex: 0, detailsLastIndex: 0, segmentsLastIndex: 0, photosLastIndex: 0 }

  // Get existing activity IDs if not full sync
  const existingIds = new Set<number>()
  if (!effectiveOptions.fullSync && !isResuming) {
    const rows = db.query<{ id: number }>('SELECT id FROM activities WHERE athlete_id = ?', [athleteId])
    for (const row of rows) {
      existingIds.add(row.id)
    }
  }

  // Determine starting phase
  const phaseOrder: ImportPhase[] = ['activities', 'gear', 'streams', 'details', 'segments', 'photos']
  const startPhaseIndex = isResuming ? phaseOrder.indexOf(savedState.phase) : 0

  try {
    // Phase 1: Activities (skip if resuming past this phase)
    if (startPhaseIndex <= 0) {
      await runActivitiesPhase(db, athleteId, existingIds, state, effectiveOptions, startPhaseIndex === 0 ? resumeIndices.activitiesLastPage : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
    }

    // Phase 2: Gear (skip if resuming past this phase)
    if (startPhaseIndex <= 1) {
      await runGearPhase(db, athleteId, state, effectiveOptions, startPhaseIndex === 1 ? resumeIndices.gearLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
    }

    // Phase 3: Streams
    if (!effectiveOptions.skipStreams && startPhaseIndex <= 2) {
      await runStreamsPhase(db, state, effectiveOptions, startPhaseIndex === 2 ? resumeIndices.streamsLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
    }

    // Phase 4: Activity Details (segment efforts + best efforts)
    if (!effectiveOptions.skipSegments && startPhaseIndex <= 3) {
      await runActivityDetailsPhase(db, athleteId, state, effectiveOptions, startPhaseIndex === 3 ? resumeIndices.detailsLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
    }

    // Phase 5: Segment Details (for incomplete segments)
    if (!effectiveOptions.skipSegments && state.segmentIdsToFetch.size > 0 && startPhaseIndex <= 4) {
      await runSegmentDetailsPhase(db, state, effectiveOptions, startPhaseIndex === 4 ? resumeIndices.segmentsLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
    }

    // Phase 6: Photos
    if (!effectiveOptions.skipPhotos && startPhaseIndex <= 5) {
      await runPhotosPhase(db, athleteId, state, effectiveOptions, startPhaseIndex === 5 ? resumeIndices.photosLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
    }

    // Complete - clear saved state
    clearImportState()
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

    if (errorMsg === 'Paused') {
      // Pause was requested - state already saved, clear flag and return
      pauseRequested = false
      console.log('[Import] Paused')
      return
    } else if (errorMsg === 'Cancelled') {
      importProgress.status = 'cancelled'
      clearImportState()
    } else {
      importProgress.status = 'error'
      importProgress.error = errorMsg
      clearImportState()
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
  state: SyncState,
  options: ImportOptions,
  startPage = 1
): Promise<void> {
  importProgress.phase = 'activities'
  console.log(`[Import] Phase 1: Fetching activities (starting at page ${startPage || 1})`)

  let page = startPage || 1
  const perPage = 100
  let hasMore = true

  while (hasMore && !cancelRequested) {
    const activities = await stravaFetch<StravaActivity[]>(`/athlete/activities?page=${page}&per_page=${perPage}`)

    if (activities.length === 0) {
      hasMore = false
      break
    }

    importProgress.activities_total += activities.length

    let batchCount = 0
    for (const activity of activities) {
      if (cancelRequested) break

      // Track newest activity date
      if (!state.newestActivityDate || activity.start_date > state.newestActivityDate) {
        state.newestActivityDate = activity.start_date
      }

      // Collect activity ID for later phases (with memory limit)
      if (state.activityIds.length < MAX_ACTIVITY_IDS_IN_MEMORY) {
        state.activityIds.push(activity.id)
      } else if (state.activityIds.length === MAX_ACTIVITY_IDS_IN_MEMORY) {
        // Log warning once when limit is reached (don't push beyond limit)
        console.warn(`[Import] Activity ID memory limit reached (${MAX_ACTIVITY_IDS_IN_MEMORY}). ` +
          'Older activities will not have streams/details fetched in this sync. Run another sync for remaining activities.')
      }
      // Beyond limit, we still store the activity but don't track its ID for later phases

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

      // Yield to main thread periodically to prevent UI blocking
      batchCount++
      if (batchCount >= DB_BATCH_SIZE) {
        await yieldToMain()
        batchCount = 0
      }
    }

    // Update ETA after each page
    updateETA('activities', state, options)

    // Check for pause after each page (next page will be fetched on resume)
    if (checkPauseRequested('activities', state, options, page + 1, 'activitiesLastPage')) break

    page++
    await sleep(RATE_LIMIT_DELAY_SHORT_MS)
  }

  console.log(`[Import] Phase 1 complete: ${state.activityIds.length} activities, ${state.gearIds.size} unique gear`)
}

// ============================================================================
// Phase 2: Gear
// ============================================================================

async function runGearPhase(
  db: ReturnType<typeof getDatabase>,
  athleteId: number,
  state: SyncState,
  options: ImportOptions,
  startIndex = 0
): Promise<void> {
  if (state.gearIds.size === 0) {
    console.log('[Import] Phase 2: No gear to fetch')
    return
  }

  importProgress.phase = 'gear'
  importProgress.gear_total = state.gearIds.size
  console.log(`[Import] Phase 2: Fetching ${state.gearIds.size} gear items (starting at ${startIndex})`)

  const gearArray = Array.from(state.gearIds)
  for (let i = startIndex; i < gearArray.length; i++) {
    if (cancelRequested) break
    if (checkPauseRequested('gear', state, options, i, 'gearLastIndex')) break

    const gearId = gearArray[i]
    try {
      const gear = await stravaFetch<StravaGear>(`/gear/${gearId}`)
      storeGear(db, athleteId, gear)
      importProgress.gear_done++
      updateETA('gear', state, options)
      await sleep(RATE_LIMIT_DELAY_SHORT_MS)
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

async function runStreamsPhase(
  db: ReturnType<typeof getDatabase>,
  state: SyncState,
  options: ImportOptions,
  startIndex = 0
): Promise<void> {
  importProgress.phase = 'streams'
  importProgress.streams_total = state.activityIds.length
  console.log(`[Import] Phase 3: Fetching streams for ${state.activityIds.length} activities (starting at ${startIndex})`)

  for (let i = startIndex; i < state.activityIds.length; i++) {
    if (cancelRequested) break
    if (checkPauseRequested('streams', state, options, i, 'streamsLastIndex')) break

    const activityId = state.activityIds[i]

    // Check if we already have streams
    const hasStreams = db.queryOne<{ count: number }>(
      'SELECT COUNT(*) as count FROM activity_streams WHERE activity_id = ?',
      [activityId]
    )

    if (hasStreams && hasStreams.count > 0) {
      importProgress.streams_done++
      updateETA('streams', state, options)
      continue
    }

    try {
      await fetchAndStoreStreams(db, activityId)
      importProgress.streams_done++
      updateETA('streams', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch streams for activity ${activityId}:`, error)
      importProgress.streams_done++ // Still count as done (skipped)
      updateETA('streams', state, options)
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
  state: SyncState,
  options: ImportOptions,
  startIndex = 0
): Promise<void> {
  importProgress.phase = 'details'
  importProgress.details_total = state.activityIds.length
  console.log(`[Import] Phase 4: Fetching activity details for ${state.activityIds.length} activities (starting at ${startIndex})`)

  for (let i = startIndex; i < state.activityIds.length; i++) {
    if (cancelRequested) break
    if (checkPauseRequested('details', state, options, i, 'detailsLastIndex')) break

    const activityId = state.activityIds[i]

    // Check if we already have segment efforts for this activity
    const hasEfforts = db.queryOne<{ count: number }>(
      'SELECT COUNT(*) as count FROM segment_efforts WHERE activity_id = ?',
      [activityId]
    )

    if (hasEfforts && hasEfforts.count > 0) {
      importProgress.details_done++
      updateETA('details', state, options)
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
      updateETA('details', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch details for activity ${activityId}:`, error)
      importProgress.details_done++ // Still count as processed
      updateETA('details', state, options)
    }
  }

  console.log(`[Import] Phase 4 complete: ${importProgress.details_done} activities processed, ${state.segmentIdsToFetch.size} segments need detail`)
}

// ============================================================================
// Phase 5: Segment Details
// ============================================================================

async function runSegmentDetailsPhase(
  db: ReturnType<typeof getDatabase>,
  state: SyncState,
  options: ImportOptions,
  startIndex = 0
): Promise<void> {
  importProgress.phase = 'segments'
  importProgress.segments_total = state.segmentIdsToFetch.size
  console.log(`[Import] Phase 5: Fetching details for ${state.segmentIdsToFetch.size} segments (starting at ${startIndex})`)

  const segmentArray = Array.from(state.segmentIdsToFetch)
  for (let i = startIndex; i < segmentArray.length; i++) {
    if (cancelRequested) break
    if (checkPauseRequested('segments', state, options, i, 'segmentsLastIndex')) break

    const segmentId = segmentArray[i]
    try {
      const segment = await stravaFetch<StravaSegment>(`/segments/${segmentId}`)
      storeSegmentDetail(db, segment)
      importProgress.segments_done++
      updateETA('segments', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch segment ${segmentId}:`, error)
      importProgress.segments_done++ // Still count as processed
      updateETA('segments', state, options)
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
  state: SyncState,
  options: ImportOptions,
  startIndex = 0
): Promise<void> {
  if (state.activitiesWithPhotos.length === 0) {
    console.log('[Import] Phase 6: No activities with photos')
    return
  }

  importProgress.phase = 'photos'
  importProgress.photos_total = state.activitiesWithPhotos.length
  console.log(`[Import] Phase 6: Fetching photos for ${state.activitiesWithPhotos.length} activities (starting at ${startIndex})`)

  for (let i = startIndex; i < state.activitiesWithPhotos.length; i++) {
    if (cancelRequested) break
    if (checkPauseRequested('photos', state, options, i, 'photosLastIndex')) break

    const activityId = state.activitiesWithPhotos[i]

    // Check if we already have photos
    const hasPhotos = db.queryOne<{ count: number }>('SELECT COUNT(*) as count FROM photos WHERE activity_id = ?', [
      activityId,
    ])

    if (hasPhotos && hasPhotos.count > 0) {
      importProgress.photos_done++
      updateETA('photos', state, options)
      continue
    }

    try {
      await fetchAndStorePhotos(db, athleteId, activityId)
      importProgress.photos_done++
      updateETA('photos', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch photos for activity ${activityId}:`, error)
      importProgress.photos_done++ // Still count as processed
      updateETA('photos', state, options)
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

    // Compute power best efforts if watts stream exists
    const wattsStream = streams.find((s) => s.type === 'watts')
    if (wattsStream && Array.isArray(wattsStream.data) && wattsStream.data.length > 0) {
      computeAndStorePowerBests(db, activityId, wattsStream.data as number[])
    }

    return true
  } catch {
    return false
  }
}

/**
 * Compute power best efforts for an activity and store in database.
 */
function computeAndStorePowerBests(
  db: ReturnType<typeof getDatabase>,
  activityId: number,
  watts: number[]
): void {
  // Skip if algorithms not initialized
  if (!algorithmsInitialized()) {
    console.warn('[Import] Algorithms not initialized, skipping power computation')
    return
  }

  const athlete = getAthlete()
  if (!athlete) return

  for (const duration of POWER_DURATIONS) {
    // Skip if activity is shorter than duration
    if (watts.length < duration) continue

    try {
      const best = rollingMaxAverage(watts, duration)
      if (best <= 0) continue

      db.exec(
        `INSERT INTO power_best_efforts (activity_id, athlete_id, duration_s, best_avg_watts, computed_at)
         VALUES (?, ?, ?, ?, datetime('now'))
         ON CONFLICT (activity_id, duration_s) DO UPDATE SET
           best_avg_watts = EXCLUDED.best_avg_watts,
           computed_at = EXCLUDED.computed_at`,
        [activityId, athlete.id, duration, Math.round(best)]
      )
    } catch (err) {
      console.error(`[Import] Failed to compute power for duration ${duration}s:`, err)
    }
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

/**
 * Yield to the main thread to prevent UI blocking.
 * Uses setTimeout(0) to allow event loop to process other tasks.
 */
function yieldToMain(): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, 0))
}

// Batch size for database operations before yielding
const DB_BATCH_SIZE = 25

// Maximum number of activity IDs to track in memory (safety limit for very large accounts)
// Beyond this, older activity IDs are dropped from memory (they're already stored in DB)
const MAX_ACTIVITY_IDS_IN_MEMORY = 50000

/**
 * Backfill power best efforts for activities that have watts streams but no power data.
 * Call this after importing activities to compute power for existing data.
 * Yields to main thread periodically to keep UI responsive on large datasets.
 */
export async function backfillPowerBests(): Promise<number> {
  if (!algorithmsInitialized()) {
    console.warn('[Import] Algorithms not initialized, cannot backfill power')
    return 0
  }

  const athlete = getAthlete()
  if (!athlete) {
    console.warn('[Import] Not authenticated, cannot backfill power')
    return 0
  }

  const db = getDatabase()
  if (!db.isInitialized()) {
    console.warn('[Import] Database not initialized, cannot backfill power')
    return 0
  }

  // Find activities with watts stream but no power_best_efforts
  const activities = db.query<{ id: number }>(
    `SELECT DISTINCT s.activity_id as id
     FROM activity_streams s
     LEFT JOIN power_best_efforts p ON p.activity_id = s.activity_id
     WHERE s.stream_type = 'watts' AND p.activity_id IS NULL`
  )

  if (activities.length === 0) {
    console.log('[Import] No activities need power backfill')
    return 0
  }

  console.log(`[Import] Backfilling power for ${activities.length} activities`)

  let computed = 0
  let batchCount = 0
  for (const { id } of activities) {
    const stream = db.queryOne<{ data: string }>(
      `SELECT data FROM activity_streams WHERE activity_id = ? AND stream_type = 'watts'`,
      [id]
    )
    if (!stream?.data) continue

    try {
      const watts = JSON.parse(stream.data) as number[]
      if (Array.isArray(watts) && watts.length > 0) {
        computeAndStorePowerBests(db, id, watts)
        computed++
      }
    } catch (err) {
      console.error(`[Import] Failed to backfill power for activity ${id}:`, err)
    }

    // Yield to main thread periodically to prevent UI blocking
    batchCount++
    if (batchCount >= DB_BATCH_SIZE) {
      await yieldToMain()
      batchCount = 0
    }
  }

  if (computed > 0) {
    await db.persist()
  }

  console.log(`[Import] Backfilled power for ${computed} activities`)
  return computed
}
