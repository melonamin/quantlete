/**
 * Data Events - Event system for reactive UI updates.
 *
 * This module provides an event-driven system for notifying the UI
 * when data changes in the database. Used by both WASM and Server modes.
 */

import type { ImportPhase } from './types'

// ============================================================================
// Event Types
// ============================================================================

/**
 * Sync progress event - emitted periodically during sync.
 * Contains current progress counters for all phases.
 */
export interface SyncProgressEvent {
  type: 'sync:progress'
  status: 'idle' | 'running' | 'completed' | 'failed' | 'canceled' | 'paused'
  phase: ImportPhase
  activitiesDone: number
  activitiesTotal: number
  gearDone: number
  gearTotal: number
  streamsDone: number
  streamsTotal: number
  detailsDone: number
  detailsTotal: number
  segmentsDone: number
  segmentsTotal: number
  photosDone: number
  photosTotal: number
  estimatedEta?: string
  // Rate limit waiting state
  waitingForRateLimit?: boolean
  waitingUntil?: string
  waitingReason?: string
}

/**
 * Sync complete event - emitted when sync finishes.
 */
export interface SyncCompleteEvent {
  type: 'sync:complete'
  status: 'completed' | 'failed' | 'canceled' | 'paused'
  error?: string
}

/**
 * Data changed event - emitted when data is modified.
 * Used to trigger query invalidation.
 */
export interface DataChangedEvent {
  type: 'data:changed'
  changes: DataChangeSet
}

/**
 * Set of data types that changed.
 */
export interface DataChangeSet {
  activities?: boolean
  streams?: boolean
  segments?: boolean
  gear?: boolean
  photos?: boolean
  all?: boolean
}

/**
 * Union of all data event types.
 */
export type DataEvent = SyncProgressEvent | SyncCompleteEvent | DataChangedEvent

/**
 * Listener function type for data events.
 */
export type DataEventListener = (event: DataEvent) => void

// ============================================================================
// Event Emitter
// ============================================================================

/**
 * Simple event emitter for data events.
 * Thread-safe for single-threaded JavaScript execution.
 */
export class DataEventEmitter {
  private listeners: Set<DataEventListener> = new Set()

  /**
   * Subscribe to data events.
   * @returns Unsubscribe function
   */
  subscribe(listener: DataEventListener): () => void {
    this.listeners.add(listener)
    return () => this.listeners.delete(listener)
  }

  /**
   * Emit an event to all listeners.
   */
  emit(event: DataEvent): void {
    for (const listener of this.listeners) {
      try {
        listener(event)
      } catch (err) {
        console.error('[DataEvents] Listener error:', err)
      }
    }
  }

  /**
   * Clear all listeners.
   */
  clear(): void {
    this.listeners.clear()
  }

  /**
   * Get the number of active listeners.
   */
  get listenerCount(): number {
    return this.listeners.size
  }
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Create a sync progress event from import progress data.
 */
export function createSyncProgressEvent(
  status: SyncProgressEvent['status'],
  phase: ImportPhase,
  progress: {
    activities_done: number
    activities_total: number
    gear_done: number
    gear_total: number
    streams_done: number
    streams_total: number
    details_done: number
    details_total: number
    segments_done: number
    segments_total: number
    photos_done: number
    photos_total: number
    estimated_eta?: string
    waiting_for_rate_limit?: boolean
    waiting_until?: string
    waiting_reason?: string
  }
): SyncProgressEvent {
  return {
    type: 'sync:progress',
    status,
    phase,
    activitiesDone: progress.activities_done,
    activitiesTotal: progress.activities_total,
    gearDone: progress.gear_done,
    gearTotal: progress.gear_total,
    streamsDone: progress.streams_done,
    streamsTotal: progress.streams_total,
    detailsDone: progress.details_done,
    detailsTotal: progress.details_total,
    segmentsDone: progress.segments_done,
    segmentsTotal: progress.segments_total,
    photosDone: progress.photos_done,
    photosTotal: progress.photos_total,
    estimatedEta: progress.estimated_eta,
    waitingForRateLimit: progress.waiting_for_rate_limit,
    waitingUntil: progress.waiting_until,
    waitingReason: progress.waiting_reason,
  }
}

/**
 * Create a sync complete event.
 */
export function createSyncCompleteEvent(
  status: 'completed' | 'failed' | 'canceled' | 'paused',
  error?: string
): SyncCompleteEvent {
  return {
    type: 'sync:complete',
    status,
    error,
  }
}

/**
 * Create a data changed event.
 */
export function createDataChangedEvent(changes: DataChangeSet): DataChangedEvent {
  return {
    type: 'data:changed',
    changes,
  }
}
