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
import {
  saveImportState,
  loadImportState,
  clearImportState,
  hasResumableState,
  type ImportState,
} from './import-state'
import {
  DataEventEmitter,
  createSyncProgressEvent,
  createSyncCompleteEvent,
  createDataChangedEvent,
  type DataEventListener,
} from '@/lib/data/events'
import { ImportEventBatchSize, ImportEventFlushIntervalMs } from '@/lib/shared/constants.gen'
import {
  saveActivity as goSaveActivity,
  saveStream as goSaveStream,
  saveGear as goSaveGear,
  saveSegment as goSaveSegment,
  saveSegmentEffort as goSaveSegmentEffort,
  saveBestEfforts as goSaveBestEfforts,
  savePhoto as goSavePhoto,
  computePowerBestEfforts as goComputePowerBestEfforts,
  createSyncRun as goCreateSyncRun,
  updateSyncRun as goUpdateSyncRun,
  completeSyncRun as goCompleteSyncRun,
  getSyncHistory as goGetSyncHistory,
} from '../go-storage'

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
  | 'activity_details'
  | 'segment_details'
  | 'photos'
  | 'completed'

export interface ImportOptions {
  fullSync?: boolean
  skipStreams?: boolean
  skipSegments?: boolean
  skipBestEfforts?: boolean
  skipPhotos?: boolean
  resume?: boolean
}

export interface ImportProgress {
  status: 'idle' | 'running' | 'completed' | 'failed' | 'canceled' | 'paused'
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

// Event emitter for reactive updates
const importEventEmitter = new DataEventEmitter()

// Track last event emission time for time-based flush
let lastEventEmitTime = 0

// ============================================================================
// Public API
// ============================================================================

export function getImportProgress(): ImportProgress {
  return { ...importProgress }
}

/**
 * Subscribe to import events for reactive UI updates.
 * @returns Unsubscribe function
 */
export function subscribeToImportEvents(listener: DataEventListener): () => void {
  return importEventEmitter.subscribe(listener)
}

/**
 * Emit a sync progress event with current state.
 */
function emitProgressEvent(): void {
  const event = createSyncProgressEvent(importProgress.phase, {
    activities_done: importProgress.activities_done,
    activities_total: importProgress.activities_total,
    gear_done: importProgress.gear_done,
    gear_total: importProgress.gear_total,
    streams_done: importProgress.streams_done,
    streams_total: importProgress.streams_total,
    details_done: importProgress.details_done,
    details_total: importProgress.details_total,
    segments_done: importProgress.segments_done,
    segments_total: importProgress.segments_total,
    photos_done: importProgress.photos_done,
    photos_total: importProgress.photos_total,
    estimated_eta: importProgress.estimated_eta,
  })
  importEventEmitter.emit(event)
  lastEventEmitTime = Date.now()
}

/**
 * Check if we should emit a progress event based on batch count or time elapsed.
 * Emits if either:
 * - itemsDone is a multiple of ImportEventBatchSize (batch threshold)
 * - More than ImportEventFlushIntervalMs has elapsed since last emit (time threshold)
 */
function shouldEmitProgress(itemsDone: number): boolean {
  if (itemsDone % ImportEventBatchSize === 0) {
    return true
  }
  const now = Date.now()
  if (now - lastEventEmitTime >= ImportEventFlushIntervalMs) {
    return true
  }
  return false
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
      try {
        goCompleteSyncRun({
          id: importProgress.sync_run_id,
          status: 'canceled',
          activities_total: importProgress.activities_total,
          activities_imported: importProgress.activities_done,
          streams_imported: importProgress.streams_done,
          failed_count: importProgress.failed_count,
        })
      } catch (err) {
        console.error('[Import] Failed to update sync_history on cancel:', err)
      }
      importProgress.status = 'canceled'
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
  const athlete = getAthlete()
  if (!athlete) return []

  try {
    const items = goGetSyncHistory(limit)
    return items.map((r) => ({
      id: r.id,
      athlete_id: r.athlete_id,
      started_at: r.started_at,
      completed_at: r.completed_at,
      duration_seconds: r.duration_seconds,
      status: r.status,
      error: r.error,
      activities_total: r.activities_total,
      activities_imported: r.activities_imported,
      activities_skipped: r.activities_skipped,
      streams_imported: r.streams_imported,
      failed_count: r.failed_count,
      full_sync: r.full_sync,
      skip_streams: r.skip_streams,
      newest_activity_date: r.newest_activity_date,
    }))
  } catch (err) {
    console.error('[Import] Failed to get sync history:', err)
    return []
  }
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

    case 'activity_details':
      calls += importProgress.details_total - importProgress.details_done
      if (!options.skipSegments) calls += state.segmentIdsToFetch.size - importProgress.segments_done
      if (!options.skipPhotos) calls += state.activitiesWithPhotos.length - importProgress.photos_done
      break

    case 'segment_details':
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
    try {
      goUpdateSyncRun({
        id: importProgress.sync_run_id,
        status: 'paused',
        activities_total: importProgress.activities_total,
        activities_imported: importProgress.activities_done,
        streams_imported: importProgress.streams_done,
        failed_count: importProgress.failed_count,
      })
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
  lastEventEmitTime = Date.now() // Reset for time-based flush
  const db = getDatabase()
  const athleteId = athlete.id

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
    syncRunId = goCreateSyncRun({
      athlete_id: athleteId,
      full_sync: effectiveOptions.fullSync ?? false,
      skip_streams: effectiveOptions.skipStreams ?? false,
      skip_segments: effectiveOptions.skipSegments ?? false,
      skip_best_efforts: effectiveOptions.skipBestEfforts ?? false,
      skip_photos: effectiveOptions.skipPhotos ?? false,
    })

    // If resuming, mark the old sync run as resumed
    if (isResuming && savedState.syncRunId) {
      try {
        goCompleteSyncRun({
          id: savedState.syncRunId,
          status: 'canceled',
          error: `Resumed in sync run ${syncRunId}`,
        })
      } catch (err) {
        console.error('[Import] Failed to mark previous sync as canceled:', err)
      }
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
  const phaseOrder: ImportPhase[] = ['activities', 'gear', 'streams', 'activity_details', 'segment_details', 'photos']
  const startPhaseIndex = isResuming ? phaseOrder.indexOf(savedState.phase) : 0

  try {
    // Phase 1: Activities (skip if resuming past this phase)
    if (startPhaseIndex <= 0) {
      await runActivitiesPhase(db, athleteId, existingIds, state, effectiveOptions, startPhaseIndex === 0 ? resumeIndices.activitiesLastPage : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
      // Emit data changed event for activities
      importEventEmitter.emit(createDataChangedEvent({ activities: true }))
      emitProgressEvent()
    }

    // Phase 2: Gear (skip if resuming past this phase)
    if (startPhaseIndex <= 1) {
      await runGearPhase(db, athleteId, state, effectiveOptions, startPhaseIndex === 1 ? resumeIndices.gearLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
      // Emit data changed event for gear
      importEventEmitter.emit(createDataChangedEvent({ gear: true }))
      emitProgressEvent()
    }

    // Phase 3: Streams
    if (!effectiveOptions.skipStreams && startPhaseIndex <= 2) {
      await runStreamsPhase(db, state, effectiveOptions, startPhaseIndex === 2 ? resumeIndices.streamsLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
      // Emit data changed event for streams
      importEventEmitter.emit(createDataChangedEvent({ streams: true }))
      emitProgressEvent()
    }

    // Phase 4: Activity Details (segment efforts + best efforts)
    if (!effectiveOptions.skipSegments && startPhaseIndex <= 3) {
      await runActivityDetailsPhase(db, athleteId, state, effectiveOptions, startPhaseIndex === 3 ? resumeIndices.detailsLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
      // Emit data changed event for segments
      importEventEmitter.emit(createDataChangedEvent({ segments: true }))
      emitProgressEvent()
    }

    // Phase 5: Segment Details (for incomplete segments)
    if (!effectiveOptions.skipSegments && state.segmentIdsToFetch.size > 0 && startPhaseIndex <= 4) {
      await runSegmentDetailsPhase(db, state, effectiveOptions, startPhaseIndex === 4 ? resumeIndices.segmentsLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
      // Segment details just updates existing segments
      emitProgressEvent()
    }

    // Phase 6: Photos
    if (!effectiveOptions.skipPhotos && startPhaseIndex <= 5) {
      await runPhotosPhase(db, athleteId, state, effectiveOptions, startPhaseIndex === 5 ? resumeIndices.photosLastIndex : 0)
      if (cancelRequested) throw new Error('Cancelled')
      if (pauseRequested) throw new Error('Paused')
      await db.persist()
      // Emit data changed event for photos
      importEventEmitter.emit(createDataChangedEvent({ photos: true }))
      emitProgressEvent()
    }

    // Complete - clear saved state
    clearImportState()
    importProgress.status = 'completed'
    importProgress.phase = 'completed'

    // Emit sync complete event
    importEventEmitter.emit(createSyncCompleteEvent('completed'))
    importEventEmitter.emit(createDataChangedEvent({ all: true }))

    // Update sync history
    if (syncRunId) {
      try {
        goCompleteSyncRun({
          id: syncRunId,
          status: 'completed',
          activities_total: importProgress.activities_total,
          activities_imported: importProgress.activities_done,
          streams_imported: importProgress.streams_done,
          failed_count: importProgress.failed_count,
          newest_activity_date: state.newestActivityDate ?? undefined,
        })
      } catch (err) {
        console.error('[Import] Failed to complete sync run:', err)
      }
      await db.persist()
    }
  } catch (error) {
    const errorMsg = error instanceof Error ? error.message : 'Unknown error'

    if (errorMsg === 'Paused') {
      // Pause was requested - state already saved, clear flag and return
      pauseRequested = false
      console.log('[Import] Paused')
      // Emit paused event
      importEventEmitter.emit(createSyncCompleteEvent('paused'))
      return
    } else if (errorMsg === 'Cancelled') {
      importProgress.status = 'canceled'
      clearImportState()
      // Emit canceled event
      importEventEmitter.emit(createSyncCompleteEvent('canceled'))
    } else {
      importProgress.status = 'failed'
      importProgress.error = errorMsg
      clearImportState()
      // Emit failed event
      importEventEmitter.emit(createSyncCompleteEvent('failed', errorMsg))
    }

    // Update sync history with failure/cancellation
    if (syncRunId) {
      try {
        goCompleteSyncRun({
          id: syncRunId,
          status: importProgress.status === 'canceled' ? 'canceled' : 'failed',
          error: importProgress.error ?? undefined,
          activities_total: importProgress.activities_total,
          activities_imported: importProgress.activities_done,
          streams_imported: importProgress.streams_done,
          failed_count: importProgress.failed_count,
        })
      } catch (err) {
        console.error('[Import] Failed to record sync failure:', err)
      }
      try {
        await db.persist()
      } catch (err) {
        console.error('[Import] Failed to persist database during failure handling:', err)
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
  _db: ReturnType<typeof getDatabase>,
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
          storeActivity(athleteId, activity)
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

      // Emit progress event periodically (batch or time-based)
      if (shouldEmitProgress(importProgress.activities_done)) {
        emitProgressEvent()
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
  _db: ReturnType<typeof getDatabase>,
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
      storeGear(athleteId, gear)
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
      await fetchAndStoreStreams(activityId)
      importProgress.streams_done++
      updateETA('streams', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch streams for activity ${activityId}:`, error)
      importProgress.streams_done++ // Still count as done (skipped)
      updateETA('streams', state, options)
    }

    // Emit progress event periodically (batch or time-based)
    if (shouldEmitProgress(importProgress.streams_done)) {
      emitProgressEvent()
      // Also emit data changed event periodically for incremental updates
      importEventEmitter.emit(createDataChangedEvent({ streams: true }))
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
  importProgress.phase = 'activity_details'
  importProgress.details_total = state.activityIds.length
  console.log(`[Import] Phase 4: Fetching activity details for ${state.activityIds.length} activities (starting at ${startIndex})`)

  for (let i = startIndex; i < state.activityIds.length; i++) {
    if (cancelRequested) break
    if (checkPauseRequested('activity_details', state, options, i, 'detailsLastIndex')) break

    const activityId = state.activityIds[i]

    // Check if we already have segment efforts for this activity
    const hasEfforts = db.queryOne<{ count: number }>(
      'SELECT COUNT(*) as count FROM segment_efforts WHERE activity_id = ?',
      [activityId]
    )

    if (hasEfforts && hasEfforts.count > 0) {
      importProgress.details_done++
      updateETA('activity_details', state, options)
      continue
    }

    try {
      const detailed = await stravaFetch<StravaDetailedActivity>(`/activities/${activityId}`)

      // Store segment efforts
      if (detailed.segment_efforts && detailed.segment_efforts.length > 0) {
        for (const effort of detailed.segment_efforts) {
          const seg = effort.segment

          // Store segment (basic info from effort)
          storeSegmentFromEffort(seg)

          // Check if segment needs full detail fetch
          if (shouldFetchSegmentDetail(seg)) {
            state.segmentIdsToFetch.add(seg.id)
          }

          // Store segment effort
          storeSegmentEffort(athleteId, activityId, effort)
        }
      }

      // Store best efforts (unless skipped)
      if (!options.skipBestEfforts && detailed.best_efforts && detailed.best_efforts.length > 0) {
        for (const effort of detailed.best_efforts) {
          storeBestEffort(athleteId, activityId, effort)
        }
      }

      importProgress.details_done++
      updateETA('activity_details', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch details for activity ${activityId}:`, error)
      importProgress.details_done++ // Still count as processed
      updateETA('activity_details', state, options)
    }

    // Emit progress event periodically (batch or time-based)
    if (shouldEmitProgress(importProgress.details_done)) {
      emitProgressEvent()
    }
  }

  console.log(`[Import] Phase 4 complete: ${importProgress.details_done} activities processed, ${state.segmentIdsToFetch.size} segments need detail`)
}

// ============================================================================
// Phase 5: Segment Details
// ============================================================================

async function runSegmentDetailsPhase(
  _db: ReturnType<typeof getDatabase>,
  state: SyncState,
  options: ImportOptions,
  startIndex = 0
): Promise<void> {
  importProgress.phase = 'segment_details'
  importProgress.segments_total = state.segmentIdsToFetch.size
  console.log(`[Import] Phase 5: Fetching details for ${state.segmentIdsToFetch.size} segments (starting at ${startIndex})`)

  const segmentArray = Array.from(state.segmentIdsToFetch)
  for (let i = startIndex; i < segmentArray.length; i++) {
    if (cancelRequested) break
    if (checkPauseRequested('segment_details', state, options, i, 'segmentsLastIndex')) break

    const segmentId = segmentArray[i]
    try {
      const segment = await stravaFetch<StravaSegment>(`/segments/${segmentId}`)
      storeSegmentDetail(segment)
      importProgress.segments_done++
      updateETA('segment_details', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch segment ${segmentId}:`, error)
      importProgress.segments_done++ // Still count as processed
      updateETA('segment_details', state, options)
    }

    // Emit progress event periodically (batch or time-based)
    if (shouldEmitProgress(importProgress.segments_done)) {
      emitProgressEvent()
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
      await fetchAndStorePhotos(athleteId, activityId)
      importProgress.photos_done++
      updateETA('photos', state, options)
      await sleep(RATE_LIMIT_DELAY_LONG_MS)
    } catch (error) {
      console.error(`Failed to fetch photos for activity ${activityId}:`, error)
      importProgress.photos_done++ // Still count as processed
      updateETA('photos', state, options)
    }

    // Emit progress event periodically (batch or time-based)
    if (shouldEmitProgress(importProgress.photos_done)) {
      emitProgressEvent()
    }
  }

  console.log(`[Import] Phase 6 complete: ${importProgress.photos_done} activities processed`)
}

// ============================================================================
// Storage Functions
// ============================================================================

/**
 * Store activity in database. Returns true on success, false on failure.
 */
function storeActivity(athleteId: number, activity: StravaActivity): boolean {
  try {
    goSaveActivity({
      id: activity.id,
      athlete_id: athleteId,
      name: activity.name,
      sport_type: activity.sport_type || activity.type,
      start_date: activity.start_date,
      start_date_local: activity.start_date_local,
      timezone: activity.timezone,
      distance: activity.distance,
      moving_time: activity.moving_time,
      elapsed_time: activity.elapsed_time,
      total_elevation_gain: activity.total_elevation_gain,
      average_speed: activity.average_speed,
      max_speed: activity.max_speed,
      average_heartrate: activity.average_heartrate,
      max_heartrate: activity.max_heartrate,
      average_watts: activity.average_watts,
      max_watts: activity.max_watts,
      weighted_average_watts: activity.weighted_average_watts,
      kilojoules: activity.kilojoules,
      average_cadence: activity.average_cadence,
      calories: activity.calories,
      gear_id: activity.gear_id,
      commute: activity.commute ?? false,
      workout_type: activity.workout_type,
      location_city: activity.location_city,
      location_state: activity.location_state,
      location_country: activity.location_country,
      summary_polyline: activity.map?.summary_polyline,
      start_lat: activity.start_latlng?.[0],
      start_lng: activity.start_latlng?.[1],
      description: activity.description,
      device_name: activity.device_name,
      trainer: activity.trainer ?? false,
      private: activity.private ?? false,
      kudos_count: activity.kudos_count ?? 0,
      photo_count: activity.photo_count ?? 0,
    })
    return true
  } catch (err) {
    console.error(`[Import] Failed to store activity ${activity.id}:`, err)
    return false
  }
}

/**
 * Store gear in database. Returns true on success, false on failure.
 */
function storeGear(athleteId: number, gear: StravaGear): boolean {
  try {
    goSaveGear({
      id: gear.id,
      athlete_id: athleteId,
      name: gear.name,
      primary: gear.primary,
      retired: gear.retired,
      distance: gear.distance,
      brand_name: gear.brand_name,
      model_name: gear.model_name,
      description: gear.description,
    })
    return true
  } catch (err) {
    console.error(`[Import] Failed to store gear ${gear.id}:`, err)
    return false
  }
}

async function fetchAndStoreStreams(activityId: number): Promise<boolean> {
  try {
    const streamTypes = 'time,distance,latlng,altitude,heartrate,cadence,watts,temp'
    const streams = await stravaFetch<StravaStream[]>(
      `/activities/${activityId}/streams?keys=${streamTypes}&key_by_type=false`
    )

    if (!streams || streams.length === 0) {
      return false
    }

    for (const stream of streams) {
      goSaveStream({
        activity_id: activityId,
        stream_type: stream.type,
        data: stream.data,
        series_type: stream.series_type,
        original_size: stream.original_size,
        resolution: stream.resolution,
      })
    }

    // Compute power best efforts if watts stream exists
    // Go WASM reads the stream from DB and computes the values
    const wattsStream = streams.find((s) => s.type === 'watts')
    if (wattsStream && Array.isArray(wattsStream.data) && wattsStream.data.length > 0) {
      computeAndStorePowerBests(activityId)
    }

    return true
  } catch {
    return false
  }
}

/**
 * Compute power best efforts for an activity and store in database.
 * Delegates to Go WASM which handles reading the watts stream and computing rolling max averages.
 */
function computeAndStorePowerBests(activityId: number): void {
  const athlete = getAthlete()
  if (!athlete) return

  try {
    goComputePowerBestEfforts({
      activity_id: activityId,
      athlete_id: athlete.id,
    })
  } catch (err) {
    console.error(`[Import] Failed to compute power for activity ${activityId}:`, err)
  }
}

/**
 * Store segment from effort (basic info). Returns true on success, false on failure.
 */
function storeSegmentFromEffort(seg: StravaSegment): boolean {
  try {
    goSaveSegment({
      id: seg.id,
      name: seg.name,
      activity_type: seg.activity_type,
      distance: seg.distance,
      average_grade: seg.average_grade,
      maximum_grade: seg.maximum_grade,
      elevation_high: seg.elevation_high,
      elevation_low: seg.elevation_low,
      climb_category: seg.climb_category,
      start_lat: seg.start_latlng?.[0],
      start_lng: seg.start_latlng?.[1],
      end_lat: seg.end_latlng?.[0],
      end_lng: seg.end_latlng?.[1],
      starred: seg.starred,
    })
    return true
  } catch (err) {
    console.error(`[Import] Failed to store segment ${seg.id}:`, err)
    return false
  }
}

function shouldFetchSegmentDetail(seg: StravaSegment): boolean {
  // Fetch detail if missing polyline or athlete stats
  return !seg.map?.polyline || !seg.athlete_segment_stats
}

/**
 * Store segment with full detail (polyline, athlete stats). Returns true on success, false on failure.
 */
function storeSegmentDetail(seg: StravaSegment): boolean {
  try {
    goSaveSegment({
      id: seg.id,
      name: seg.name,
      activity_type: seg.activity_type,
      distance: seg.distance,
      average_grade: seg.average_grade,
      maximum_grade: seg.maximum_grade,
      elevation_high: seg.elevation_high,
      elevation_low: seg.elevation_low,
      climb_category: seg.climb_category,
      start_lat: seg.start_latlng?.[0],
      start_lng: seg.start_latlng?.[1],
      end_lat: seg.end_latlng?.[0],
      end_lng: seg.end_latlng?.[1],
      starred: seg.starred,
      polyline: seg.map?.polyline,
      athlete_pr_elapsed_time: seg.athlete_segment_stats?.pr_elapsed_time,
      athlete_pr_date: seg.athlete_segment_stats?.pr_date,
      athlete_effort_count: seg.athlete_segment_stats?.effort_count,
    })
    return true
  } catch (err) {
    console.error(`[Import] Failed to store segment detail ${seg.id}:`, err)
    return false
  }
}

/**
 * Store segment effort. Returns true on success, false on failure.
 */
function storeSegmentEffort(
  athleteId: number,
  activityId: number,
  effort: StravaSegmentEffort,
): boolean {
  try {
    goSaveSegmentEffort({
      id: effort.id,
      segment_id: effort.segment.id,
      activity_id: activityId,
      athlete_id: athleteId,
      name: effort.name,
      elapsed_time: effort.elapsed_time,
      moving_time: effort.moving_time,
      start_date: effort.start_date,
      start_date_local: effort.start_date_local,
      distance: effort.distance,
      average_watts: effort.average_watts,
      average_heartrate: effort.average_heartrate,
      max_heartrate: effort.max_heartrate,
      pr_rank: effort.pr_rank,
    })
    return true
  } catch (err) {
    console.error(`[Import] Failed to store segment effort ${effort.id}:`, err)
    return false
  }
}

/**
 * Store best effort. Returns true on success, false on failure.
 * Note: Go WASM handles canonicalization of distance_type from name and distance_m.
 */
function storeBestEffort(
  athleteId: number,
  activityId: number,
  effort: StravaBestEffort,
): boolean {
  try {
    goSaveBestEfforts({
      athlete_id: athleteId,
      activity_id: activityId,
      efforts: [
        {
          name: effort.name,
          distance_m: effort.distance,
          elapsed_time: effort.elapsed_time,
          start_index: effort.start_index,
          end_index: effort.end_index,
          start_date: effort.start_date,
          pr_rank: effort.pr_rank,
        },
      ],
    })
    return true
  } catch (err) {
    console.error(`[Import] Failed to store best effort ${effort.name} for activity ${activityId}:`, err)
    return false
  }
}

async function fetchAndStorePhotos(
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

      goSavePhoto({
        id: photo.unique_id,
        athlete_id: athleteId,
        activity_id: activityId,
        url: url,
        thumbnail_url: thumbnailUrl,
        caption: photo.caption,
        location: photo.location ? JSON.stringify(photo.location) : undefined,
        created_at: photo.created_at,
      })
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
    try {
      // Go WASM reads the watts stream from DB and computes power
      computeAndStorePowerBests(id)
      computed++
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
