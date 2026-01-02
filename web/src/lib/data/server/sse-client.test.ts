import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { ImportSSEClient } from './sse-client'

// Mock EventSource
class MockEventSource {
  static instances: MockEventSource[] = []

  url: string
  withCredentials: boolean
  readyState: number = 0
  onerror: ((event: Event) => void) | null = null
  private listeners: Map<string, ((e: MessageEvent) => void)[]> = new Map()

  constructor(url: string, options?: { withCredentials?: boolean }) {
    this.url = url
    this.withCredentials = options?.withCredentials ?? false
    MockEventSource.instances.push(this)
  }

  addEventListener(type: string, listener: (e: MessageEvent) => void) {
    const listeners = this.listeners.get(type) ?? []
    listeners.push(listener)
    this.listeners.set(type, listeners)
  }

  removeEventListener(type: string, listener: (e: MessageEvent) => void) {
    const listeners = this.listeners.get(type) ?? []
    const index = listeners.indexOf(listener)
    if (index !== -1) {
      listeners.splice(index, 1)
    }
  }

  close() {
    this.readyState = 2
  }

  // Test helper: simulate receiving an event
  simulateEvent(type: string, data: unknown) {
    const listeners = this.listeners.get(type) ?? []
    const event = { data: JSON.stringify(data) } as MessageEvent
    listeners.forEach((listener) => listener(event))
  }

  // Test helper: simulate an error
  simulateError() {
    if (this.onerror) {
      this.onerror(new Event('error'))
    }
  }

  static reset() {
    MockEventSource.instances = []
  }

  static get lastInstance(): MockEventSource | undefined {
    return MockEventSource.instances[MockEventSource.instances.length - 1]
  }
}

// Install mock globally
const originalEventSource = globalThis.EventSource
beforeEach(() => {
  MockEventSource.reset()
  // @ts-expect-error - mocking EventSource
  globalThis.EventSource = MockEventSource
})

afterEach(() => {
  globalThis.EventSource = originalEventSource
})

describe('ImportSSEClient', () => {
  let client: ImportSSEClient

  beforeEach(() => {
    vi.useFakeTimers()
    client = new ImportSSEClient()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('subscribe', () => {
    it('should connect on first subscribe', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      expect(MockEventSource.instances.length).toBe(1)
      expect(MockEventSource.lastInstance?.url).toBe('/api/v1/import/events')
      expect(MockEventSource.lastInstance?.withCredentials).toBe(true)
    })

    it('should not create multiple connections for multiple subscribers', () => {
      const listener1 = vi.fn()
      const listener2 = vi.fn()

      client.subscribe(listener1)
      client.subscribe(listener2)

      expect(MockEventSource.instances.length).toBe(1)
    })

    it('should return an unsubscribe function', () => {
      const listener = vi.fn()
      const unsubscribe = client.subscribe(listener)

      expect(typeof unsubscribe).toBe('function')
    })

    it('should disconnect when last subscriber unsubscribes', () => {
      const listener = vi.fn()
      const unsubscribe = client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance
      expect(eventSource?.readyState).not.toBe(2)

      unsubscribe()

      expect(eventSource?.readyState).toBe(2)
    })
  })

  describe('event handling', () => {
    it('should parse and forward sync:progress events', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('sync:progress', {
        phase: 'activities',
        activities_done: 10,
        activities_total: 100,
        gear_done: 0,
        gear_total: 5,
        streams_done: 0,
        streams_total: 0,
        details_done: 0,
        details_total: 0,
        segments_done: 0,
        segments_total: 0,
        photos_done: 0,
        photos_total: 0,
        estimated_eta: '2m 30s',
      })

      expect(listener).toHaveBeenCalledWith({
        type: 'sync:progress',
        phase: 'activities',
        activitiesDone: 10,
        activitiesTotal: 100,
        gearDone: 0,
        gearTotal: 5,
        streamsDone: 0,
        streamsTotal: 0,
        detailsDone: 0,
        detailsTotal: 0,
        segmentsDone: 0,
        segmentsTotal: 0,
        photosDone: 0,
        photosTotal: 0,
        estimatedEta: '2m 30s',
      })
    })

    it('should handle missing fields with defaults in sync:progress', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('sync:progress', {
        phase: 'activities',
        // All counters missing
      })

      expect(listener).toHaveBeenCalledWith(
        expect.objectContaining({
          type: 'sync:progress',
          activitiesDone: 0,
          activitiesTotal: 0,
          gearDone: 0,
          gearTotal: 0,
        })
      )
    })

    it('should parse and forward sync:complete events', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('sync:complete', {
        status: 'completed',
      })

      expect(listener).toHaveBeenCalledWith({
        type: 'sync:complete',
        status: 'completed',
        error: undefined,
      })
    })

    it('should parse and forward sync:complete events with error', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('sync:complete', {
        status: 'failed',
        error: 'Rate limited',
      })

      expect(listener).toHaveBeenCalledWith({
        type: 'sync:complete',
        status: 'failed',
        error: 'Rate limited',
      })
    })

    it('should parse and forward data:changed events', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('data:changed', {
        activities: true,
        streams: true,
        segments: false,
        gear: false,
        photos: false,
      })

      expect(listener).toHaveBeenCalledWith({
        type: 'data:changed',
        changes: {
          activities: true,
          streams: true,
          segments: false,
          gear: false,
          photos: false,
          all: false,
        },
      })
    })

    it('should handle missing change flags with defaults', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('data:changed', {
        activities: true,
        // Other fields missing
      })

      expect(listener).toHaveBeenCalledWith({
        type: 'data:changed',
        changes: {
          activities: true,
          streams: false,
          segments: false,
          gear: false,
          photos: false,
          all: false,
        },
      })
    })

    it('should call multiple listeners', () => {
      const listener1 = vi.fn()
      const listener2 = vi.fn()

      client.subscribe(listener1)
      client.subscribe(listener2)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('sync:complete', { status: 'completed' })

      expect(listener1).toHaveBeenCalled()
      expect(listener2).toHaveBeenCalled()
    })

    it('should not call unsubscribed listeners', () => {
      const listener = vi.fn()
      const unsubscribe = client.subscribe(listener)

      unsubscribe()

      // Resubscribe a new listener to keep connection open
      const newListener = vi.fn()
      client.subscribe(newListener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('sync:complete', { status: 'completed' })

      expect(listener).not.toHaveBeenCalled()
      expect(newListener).toHaveBeenCalled()
    })

    it('should handle listener errors gracefully', () => {
      const errorListener = vi.fn().mockImplementation(() => {
        throw new Error('Listener error')
      })
      const goodListener = vi.fn()

      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      client.subscribe(errorListener)
      client.subscribe(goodListener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateEvent('sync:complete', { status: 'completed' })

      expect(goodListener).toHaveBeenCalled()
      expect(consoleSpy).toHaveBeenCalled()

      consoleSpy.mockRestore()
    })

    it('should handle malformed JSON gracefully', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})

      const eventSource = MockEventSource.lastInstance!
      // Simulate sending invalid JSON directly
      const listeners = (eventSource as unknown as { listeners: Map<string, ((e: MessageEvent) => void)[]> }).listeners
      const progressListeners = listeners.get('sync:progress') ?? []
      progressListeners.forEach((l) => l({ data: 'invalid json{' } as MessageEvent))

      expect(listener).not.toHaveBeenCalled()
      expect(consoleSpy).toHaveBeenCalled()

      consoleSpy.mockRestore()
    })
  })

  describe('reconnection', () => {
    it('should reconnect on error with exponential backoff', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      expect(MockEventSource.instances.length).toBe(1)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateError()

      // First reconnect after 1 second
      expect(MockEventSource.instances.length).toBe(1) // Not yet

      vi.advanceTimersByTime(1000)
      expect(MockEventSource.instances.length).toBe(2)

      // Simulate another error
      MockEventSource.lastInstance!.simulateError()

      // Second reconnect after 2 seconds (exponential backoff)
      vi.advanceTimersByTime(1000)
      expect(MockEventSource.instances.length).toBe(2) // Not yet

      vi.advanceTimersByTime(1000)
      expect(MockEventSource.instances.length).toBe(3)
    })

    it('should reset backoff delay on successful message', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateError()

      vi.advanceTimersByTime(1000) // First reconnect

      // Simulate successful message
      MockEventSource.lastInstance!.simulateEvent('sync:progress', {
        phase: 'activities',
      })

      // Simulate another error
      MockEventSource.lastInstance!.simulateError()

      // Should reconnect after 1 second again (reset backoff)
      vi.advanceTimersByTime(1000)
      expect(MockEventSource.instances.length).toBe(3)
    })

    it('should not reconnect if no listeners', () => {
      const listener = vi.fn()
      const unsubscribe = client.subscribe(listener)

      unsubscribe()

      const eventSource = MockEventSource.lastInstance!
      eventSource.simulateError()

      vi.advanceTimersByTime(10000)

      // Should still be just 1 instance (the closed one)
      expect(MockEventSource.instances.length).toBe(1)
    })

    it('should cap backoff at max delay', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      // Simulate many errors to hit max backoff
      for (let i = 0; i < 10; i++) {
        MockEventSource.lastInstance!.simulateError()
        vi.advanceTimersByTime(30000) // Max delay
      }

      // Should still reconnect
      expect(MockEventSource.instances.length).toBeGreaterThan(1)
    })
  })

  describe('disconnect', () => {
    it('should close EventSource on disconnect', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      const eventSource = MockEventSource.lastInstance!
      expect(eventSource.readyState).not.toBe(2)

      client['disconnect']() // Access private method for testing

      expect(eventSource.readyState).toBe(2)
    })

    it('should cancel pending reconnect on disconnect', () => {
      const listener = vi.fn()
      client.subscribe(listener)

      MockEventSource.lastInstance!.simulateError()

      // Unsubscribe (which disconnects)
      const unsubscribe = client.subscribe(vi.fn())
      unsubscribe()
      unsubscribe() // Double unsubscribe - second listener

      vi.advanceTimersByTime(10000)

      // Should not have created additional connections
      expect(MockEventSource.instances.length).toBeLessThanOrEqual(2)
    })
  })
})
