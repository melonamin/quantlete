/**
 * SSE Client for import events from the Go backend.
 *
 * Subscribes to the /api/v1/import/events endpoint and
 * converts SSE events to DataEvents for the frontend.
 */

import type {
  DataEvent,
  DataEventListener,
  SyncProgressEvent,
  SyncCompleteEvent,
  DataChangedEvent,
} from '../events'
import type { ImportPhase } from '../types'

/**
 * SSE client for real-time import updates from the server.
 */
export class ImportSSEClient {
  private eventSource: EventSource | null = null
  private listeners: Set<DataEventListener> = new Set()
  private reconnectTimeout: ReturnType<typeof setTimeout> | null = null
  private reconnectDelay = 1000
  private readonly maxReconnectDelay = 30000
  private readonly baseUrl = '/api/v1/import/events'

  /**
   * Connect to the SSE endpoint.
   */
  connect(): void {
    if (this.eventSource) return

    this.eventSource = new EventSource(this.baseUrl, { withCredentials: true })

    // Handle sync:progress events
    this.eventSource.addEventListener('sync:progress', (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data)
        const event: SyncProgressEvent = {
          type: 'sync:progress',
          status: data.status ?? 'idle',
          phase: data.phase as ImportPhase,
          activitiesDone: data.activities_done ?? 0,
          activitiesTotal: data.activities_total ?? 0,
          gearDone: data.gear_done ?? 0,
          gearTotal: data.gear_total ?? 0,
          streamsDone: data.streams_done ?? 0,
          streamsTotal: data.streams_total ?? 0,
          detailsDone: data.details_done ?? 0,
          detailsTotal: data.details_total ?? 0,
          segmentsDone: data.segments_done ?? 0,
          segmentsTotal: data.segments_total ?? 0,
          photosDone: data.photos_done ?? 0,
          photosTotal: data.photos_total ?? 0,
          estimatedEta: data.estimated_eta,
        }
        this.handleEvent(event)
        this.reconnectDelay = 1000 // Reset on successful message
      } catch (err) {
        console.error('[SSE] Failed to parse sync:progress event:', err)
      }
    })

    // Handle sync:complete events
    this.eventSource.addEventListener('sync:complete', (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data)
        const event: SyncCompleteEvent = {
          type: 'sync:complete',
          status: data.status as 'completed' | 'failed' | 'canceled' | 'paused',
          error: data.error,
        }
        this.handleEvent(event)
      } catch (err) {
        console.error('[SSE] Failed to parse sync:complete event:', err)
      }
    })

    // Handle data:changed events
    this.eventSource.addEventListener('data:changed', (e: MessageEvent) => {
      try {
        const data = JSON.parse(e.data)
        const event: DataChangedEvent = {
          type: 'data:changed',
          changes: {
            activities: data.activities ?? false,
            streams: data.streams ?? false,
            segments: data.segments ?? false,
            gear: data.gear ?? false,
            photos: data.photos ?? false,
            all: data.all ?? false,
          },
        }
        this.handleEvent(event)
      } catch (err) {
        console.error('[SSE] Failed to parse data:changed event:', err)
      }
    })

    // Handle errors and schedule reconnection
    this.eventSource.onerror = () => {
      console.warn('[SSE] Connection error, will attempt to reconnect')
      this.disconnect()
      this.scheduleReconnect()
    }
  }

  /**
   * Disconnect from the SSE endpoint.
   */
  disconnect(): void {
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
  }

  /**
   * Schedule a reconnection attempt with exponential backoff.
   */
  private scheduleReconnect(): void {
    if (this.listeners.size === 0) {
      // No listeners, don't reconnect
      return
    }

    this.reconnectTimeout = setTimeout(() => {
      this.reconnectDelay = Math.min(this.reconnectDelay * 2, this.maxReconnectDelay)
      this.connect()
    }, this.reconnectDelay)
  }

  /**
   * Subscribe to data events.
   * Automatically connects when first listener is added.
   * @returns Unsubscribe function
   */
  subscribe(listener: DataEventListener): () => void {
    this.listeners.add(listener)
    if (this.listeners.size === 1) {
      this.connect()
    }
    return () => {
      this.listeners.delete(listener)
      if (this.listeners.size === 0) {
        this.disconnect()
      }
    }
  }

  /**
   * Dispatch an event to all listeners.
   */
  private handleEvent(event: DataEvent): void {
    for (const listener of this.listeners) {
      try {
        listener(event)
      } catch (err) {
        console.error('[SSE] Listener error:', err)
      }
    }
  }
}
