/**
 * Import State Persistence - Enables pause/resume for browser imports.
 *
 * Saves import state to app_state table so imports can resume after pause or refresh.
 */

import { getDatabase } from '../db'
import type { ImportPhase, ImportOptions } from './importer'

const IMPORT_STATE_KEY = 'import_state'

/**
 * Persisted import state for resume capability.
 */
export interface ImportState {
  // Current phase
  phase: ImportPhase

  // Collected IDs from activities phase
  activityIds: number[]
  gearIds: string[]
  segmentIdsToFetch: number[]
  activitiesWithPhotos: number[]
  newestActivityDate: string | null

  // Last processed index per phase (for resume within phase)
  activitiesLastPage: number
  gearLastIndex: number
  streamsLastIndex: number
  detailsLastIndex: number
  segmentsLastIndex: number
  photosLastIndex: number

  // Progress counters
  activitiesTotal: number
  activitiesDone: number
  gearTotal: number
  gearDone: number
  streamsTotal: number
  streamsDone: number
  detailsTotal: number
  detailsDone: number
  segmentsTotal: number
  segmentsDone: number
  photosTotal: number
  photosDone: number
  failedCount: number

  // Import options
  options: ImportOptions

  // Timestamps
  startedAt: string
  pausedAt?: string
  syncRunId?: number
}

/**
 * Save import state to database.
 */
export function saveImportState(state: ImportState): void {
  const db = getDatabase()
  if (!db.isInitialized()) return

  try {
    db.exec(
      `INSERT INTO app_state (key, value) VALUES (?, ?)
       ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
      [IMPORT_STATE_KEY, JSON.stringify(state)]
    )
  } catch (err) {
    console.error('[ImportState] Failed to save state:', err)
  }
}

/**
 * Map old phase names to new phase names for backward compatibility.
 * This allows resuming imports that were saved with old phase names.
 */
function normalizePhase(phase: string): string {
  const phaseMapping: Record<string, string> = {
    details: 'activity_details',
    segments: 'segment_details',
  }
  return phaseMapping[phase] ?? phase
}

/**
 * Validate and normalize loaded import state.
 * Returns null if state is invalid or corrupted.
 */
function validateImportState(parsed: unknown): ImportState | null {
  if (!parsed || typeof parsed !== 'object') return null

  const state = parsed as Record<string, unknown>

  // Required fields with type checks
  if (typeof state.phase !== 'string') return null
  if (!Array.isArray(state.activityIds)) return null
  if (!Array.isArray(state.gearIds)) return null
  if (!Array.isArray(state.segmentIdsToFetch)) return null
  if (!Array.isArray(state.activitiesWithPhotos)) return null

  // Normalize phase name (handle old -> new phase name migration)
  const normalizedPhase = normalizePhase(state.phase as string)

  // Validate phase is resumable
  const validPhases = ['activities', 'gear', 'streams', 'activity_details', 'segment_details', 'photos']
  if (!validPhases.includes(normalizedPhase)) {
    console.warn(`[ImportState] Unknown phase "${state.phase}", resetting import state`)
    return null
  }

  // Normalize numeric fields with defaults
  const normalizeNumber = (val: unknown, fallback: number): number =>
    typeof val === 'number' && !isNaN(val) ? val : fallback

  return {
    phase: normalizedPhase as ImportState['phase'],
    activityIds: state.activityIds.filter((id): id is number => typeof id === 'number'),
    gearIds: state.gearIds.filter((id): id is string => typeof id === 'string'),
    segmentIdsToFetch: state.segmentIdsToFetch.filter((id): id is number => typeof id === 'number'),
    activitiesWithPhotos: state.activitiesWithPhotos.filter((id): id is number => typeof id === 'number'),
    newestActivityDate: typeof state.newestActivityDate === 'string' ? state.newestActivityDate : null,
    activitiesLastPage: normalizeNumber(state.activitiesLastPage, 0),
    gearLastIndex: normalizeNumber(state.gearLastIndex, 0),
    streamsLastIndex: normalizeNumber(state.streamsLastIndex, 0),
    detailsLastIndex: normalizeNumber(state.detailsLastIndex, 0),
    segmentsLastIndex: normalizeNumber(state.segmentsLastIndex, 0),
    photosLastIndex: normalizeNumber(state.photosLastIndex, 0),
    activitiesTotal: normalizeNumber(state.activitiesTotal, 0),
    activitiesDone: normalizeNumber(state.activitiesDone, 0),
    gearTotal: normalizeNumber(state.gearTotal, 0),
    gearDone: normalizeNumber(state.gearDone, 0),
    streamsTotal: normalizeNumber(state.streamsTotal, 0),
    streamsDone: normalizeNumber(state.streamsDone, 0),
    detailsTotal: normalizeNumber(state.detailsTotal, 0),
    detailsDone: normalizeNumber(state.detailsDone, 0),
    segmentsTotal: normalizeNumber(state.segmentsTotal, 0),
    segmentsDone: normalizeNumber(state.segmentsDone, 0),
    photosTotal: normalizeNumber(state.photosTotal, 0),
    photosDone: normalizeNumber(state.photosDone, 0),
    failedCount: normalizeNumber(state.failedCount, 0),
    options: typeof state.options === 'object' && state.options !== null ? (state.options as ImportState['options']) : {},
    startedAt: typeof state.startedAt === 'string' ? state.startedAt : new Date().toISOString(),
    pausedAt: typeof state.pausedAt === 'string' ? state.pausedAt : undefined,
    syncRunId: typeof state.syncRunId === 'number' ? state.syncRunId : undefined,
  }
}

/**
 * Load import state from database.
 */
export function loadImportState(): ImportState | null {
  const db = getDatabase()
  if (!db.isInitialized()) return null

  try {
    const row = db.queryOne<{ value: string }>(`SELECT value FROM app_state WHERE key = ?`, [IMPORT_STATE_KEY])

    if (!row?.value) return null

    const parsed = JSON.parse(row.value)
    const state = validateImportState(parsed)

    if (!state) {
      console.warn('[ImportState] Invalid or corrupted state, clearing')
      clearImportState()
      return null
    }

    return state
  } catch (err) {
    console.error('[ImportState] Failed to load state:', err)
    clearImportState()
    return null
  }
}

/**
 * Clear import state from database (on completion or cancel).
 */
export function clearImportState(): void {
  const db = getDatabase()
  if (!db.isInitialized()) return

  try {
    db.exec(`DELETE FROM app_state WHERE key = ?`, [IMPORT_STATE_KEY])
  } catch (err) {
    console.error('[ImportState] Failed to clear state:', err)
  }
}

/**
 * Check if there's a resumable import state.
 */
export function hasResumableState(): boolean {
  return loadImportState() !== null
}
