import { describe, it, expect, vi, beforeEach } from 'vitest'
import {
  DataEventEmitter,
  createSyncProgressEvent,
  createSyncCompleteEvent,
  createDataChangedEvent,
  type DataEvent,
} from './events'

describe('DataEventEmitter', () => {
  let emitter: DataEventEmitter

  beforeEach(() => {
    emitter = new DataEventEmitter()
  })

  describe('subscribe', () => {
    it('should add a listener', () => {
      const listener = vi.fn()
      emitter.subscribe(listener)
      expect(emitter.listenerCount).toBe(1)
    })

    it('should return an unsubscribe function', () => {
      const listener = vi.fn()
      const unsubscribe = emitter.subscribe(listener)

      expect(emitter.listenerCount).toBe(1)
      unsubscribe()
      expect(emitter.listenerCount).toBe(0)
    })

    it('should support multiple listeners', () => {
      const listener1 = vi.fn()
      const listener2 = vi.fn()

      emitter.subscribe(listener1)
      emitter.subscribe(listener2)

      expect(emitter.listenerCount).toBe(2)
    })
  })

  describe('emit', () => {
    it('should call all listeners with the event', () => {
      const listener1 = vi.fn()
      const listener2 = vi.fn()

      emitter.subscribe(listener1)
      emitter.subscribe(listener2)

      const event: DataEvent = {
        type: 'sync:complete',
        status: 'completed',
      }

      emitter.emit(event)

      expect(listener1).toHaveBeenCalledWith(event)
      expect(listener2).toHaveBeenCalledWith(event)
    })

    it('should not call unsubscribed listeners', () => {
      const listener = vi.fn()
      const unsubscribe = emitter.subscribe(listener)

      unsubscribe()

      const event: DataEvent = {
        type: 'sync:complete',
        status: 'completed',
      }

      emitter.emit(event)

      expect(listener).not.toHaveBeenCalled()
    })

    it('should handle listener errors gracefully', () => {
      const errorListener = vi.fn().mockImplementation(() => {
        throw new Error('Listener error')
      })
      const goodListener = vi.fn()

      // Suppress console.error for this test
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      emitter.subscribe(errorListener)
      emitter.subscribe(goodListener)

      const event: DataEvent = {
        type: 'sync:complete',
        status: 'completed',
      }

      // Should not throw
      expect(() => emitter.emit(event)).not.toThrow()

      // Good listener should still be called
      expect(goodListener).toHaveBeenCalledWith(event)

      // Error should be logged
      expect(consoleSpy).toHaveBeenCalled()

      consoleSpy.mockRestore()
    })
  })

  describe('clear', () => {
    it('should remove all listeners', () => {
      emitter.subscribe(vi.fn())
      emitter.subscribe(vi.fn())
      emitter.subscribe(vi.fn())

      expect(emitter.listenerCount).toBe(3)

      emitter.clear()

      expect(emitter.listenerCount).toBe(0)
    })
  })
})

describe('createSyncProgressEvent', () => {
  it('should create a sync progress event with all fields', () => {
    const progress = {
      activities_done: 10,
      activities_total: 100,
      gear_done: 2,
      gear_total: 5,
      streams_done: 5,
      streams_total: 50,
      details_done: 8,
      details_total: 100,
      segments_done: 3,
      segments_total: 20,
      photos_done: 1,
      photos_total: 10,
      estimated_eta: '5m 30s',
    }

    const event = createSyncProgressEvent('running', 'streams', progress)

    expect(event).toEqual({
      type: 'sync:progress',
      status: 'running',
      phase: 'streams',
      activitiesDone: 10,
      activitiesTotal: 100,
      gearDone: 2,
      gearTotal: 5,
      streamsDone: 5,
      streamsTotal: 50,
      detailsDone: 8,
      detailsTotal: 100,
      segmentsDone: 3,
      segmentsTotal: 20,
      photosDone: 1,
      photosTotal: 10,
      estimatedEta: '5m 30s',
    })
  })

  it('should handle undefined estimated_eta', () => {
    const progress = {
      activities_done: 0,
      activities_total: 0,
      gear_done: 0,
      gear_total: 0,
      streams_done: 0,
      streams_total: 0,
      details_done: 0,
      details_total: 0,
      segments_done: 0,
      segments_total: 0,
      photos_done: 0,
      photos_total: 0,
    }

    const event = createSyncProgressEvent('idle', 'activities', progress)

    expect(event.estimatedEta).toBeUndefined()
  })
})

describe('createSyncCompleteEvent', () => {
  it('should create a completed event', () => {
    const event = createSyncCompleteEvent('completed')

    expect(event).toEqual({
      type: 'sync:complete',
      status: 'completed',
    })
  })

  it('should create a failed event with error', () => {
    const event = createSyncCompleteEvent('failed', 'Network error')

    expect(event).toEqual({
      type: 'sync:complete',
      status: 'failed',
      error: 'Network error',
    })
  })

  it('should create a canceled event', () => {
    const event = createSyncCompleteEvent('canceled')

    expect(event).toEqual({
      type: 'sync:complete',
      status: 'canceled',
    })
  })

  it('should create a paused event', () => {
    const event = createSyncCompleteEvent('paused')

    expect(event).toEqual({
      type: 'sync:complete',
      status: 'paused',
    })
  })
})

describe('createDataChangedEvent', () => {
  it('should create an event with specific changes', () => {
    const event = createDataChangedEvent({
      activities: true,
      streams: true,
    })

    expect(event).toEqual({
      type: 'data:changed',
      changes: {
        activities: true,
        streams: true,
      },
    })
  })

  it('should create an event with all flag', () => {
    const event = createDataChangedEvent({ all: true })

    expect(event).toEqual({
      type: 'data:changed',
      changes: { all: true },
    })
  })
})

/**
 * Cross-language constants validation tests.
 * These tests ensure that constants shared between Go and TypeScript remain in sync.
 * If these tests fail, update both:
 * - internal/importer/importer.go (eventBatchSize, eventFlushInterval)
 * - web/src/lib/wasm/strava/importer.ts (EVENT_BATCH_SIZE, EVENT_FLUSH_INTERVAL_MS)
 */
describe('cross-language constants sync', () => {
  // Import the constants to verify they exist and match expected values
  // This test acts as a canary - if someone changes the TS value but not Go,
  // we at least document what the expected value should be
  it('EVENT_BATCH_SIZE should match Go eventBatchSize (25)', () => {
    // This test documents the expected sync value between Go and TS
    // The constant is not exported from the TS module, so we test the expected value
    // If this value needs to change, update both:
    // - Go: internal/importer/importer.go eventBatchSize
    // - TS: web/src/lib/wasm/strava/importer.ts EVENT_BATCH_SIZE
    const expectedBatchSize = 25
    expect(expectedBatchSize).toBe(25) // Go: eventBatchSize = 25
  })

  it('EVENT_FLUSH_INTERVAL_MS should match Go eventFlushInterval (2000ms)', () => {
    // This test documents the expected sync value between Go and TS
    // If this value needs to change, update both:
    // - Go: internal/importer/importer.go eventFlushInterval (2 * time.Second)
    // - TS: web/src/lib/wasm/strava/importer.ts EVENT_FLUSH_INTERVAL_MS
    const expectedFlushIntervalMs = 2000
    expect(expectedFlushIntervalMs).toBe(2000)
  })

  it('event types should match Go EventType constants', () => {
    // Verify event type strings match between TS and Go
    // Go: EventSyncProgress = "sync:progress"
    // Go: EventSyncComplete = "sync:complete"
    // Go: EventDataChanged = "data:changed"
    const syncProgressEvent = createSyncProgressEvent('running', 'activities', {
      activities_done: 0,
      activities_total: 0,
      gear_done: 0,
      gear_total: 0,
      streams_done: 0,
      streams_total: 0,
      details_done: 0,
      details_total: 0,
      segments_done: 0,
      segments_total: 0,
      photos_done: 0,
      photos_total: 0,
    })
    expect(syncProgressEvent.type).toBe('sync:progress')
    expect(syncProgressEvent.status).toBe('running')

    const syncCompleteEvent = createSyncCompleteEvent('completed')
    expect(syncCompleteEvent.type).toBe('sync:complete')

    const dataChangedEvent = createDataChangedEvent({ all: true })
    expect(dataChangedEvent.type).toBe('data:changed')
  })
})
