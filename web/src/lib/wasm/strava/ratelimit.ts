/**
 * Rate Limiter for Strava API - Browser implementation matching Go backend.
 *
 * Strava rate limits:
 * - 100 requests per 15 minutes (short-term)
 * - 1000 requests per day (daily)
 *
 * Headers returned by Strava:
 * - X-RateLimit-Limit: "100,1000" (15min, daily limits)
 * - X-RateLimit-Usage: "50,500" (15min, daily usage)
 */

import { getDatabase } from '../db'

// Rate limit state
interface RateLimitState {
  limit15Min: number
  limitDaily: number
  usage15Min: number
  usageDaily: number
  lastUpdate: number // timestamp
  windowStart15Min: number // timestamp when current 15-min window started
  dayStart: number // timestamp when current day started
}

// Default limits (Strava standard)
const DEFAULT_LIMIT_15MIN = 100
const DEFAULT_LIMIT_DAILY = 1000

// Threshold for proactive waiting (90% of limit)
const THRESHOLD_PERCENT = 0.9

// Global state
let state: RateLimitState = {
  limit15Min: DEFAULT_LIMIT_15MIN,
  limitDaily: DEFAULT_LIMIT_DAILY,
  usage15Min: 0,
  usageDaily: 0,
  lastUpdate: 0,
  windowStart15Min: Date.now(),
  dayStart: getStartOfDay(),
}

// Waiting state (exposed to UI)
let waitingUntil: Date | null = null
let waitingReason: string | null = null

function getStartOfDay(): number {
  const now = new Date()
  return new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
}

/**
 * Update rate limit state from response headers.
 */
export function updateFromHeaders(headers: Headers): void {
  // X-RateLimit-Limit: "100,1000"
  const limitHeader = headers.get('X-RateLimit-Limit')
  if (limitHeader) {
    const parts = limitHeader.split(',')
    if (parts.length >= 2) {
      const limit15 = parseInt(parts[0].trim(), 10)
      const limitDaily = parseInt(parts[1].trim(), 10)
      if (!isNaN(limit15)) state.limit15Min = limit15
      if (!isNaN(limitDaily)) state.limitDaily = limitDaily
    }
  }

  // X-RateLimit-Usage: "50,500"
  const usageHeader = headers.get('X-RateLimit-Usage')
  if (usageHeader) {
    const parts = usageHeader.split(',')
    if (parts.length >= 2) {
      const usage15 = parseInt(parts[0].trim(), 10)
      const usageDaily = parseInt(parts[1].trim(), 10)
      if (!isNaN(usage15)) state.usage15Min = usage15
      if (!isNaN(usageDaily)) state.usageDaily = usageDaily
    }
  }

  state.lastUpdate = Date.now()

  // Persist state
  persistState()
}

/**
 * Increment usage counter (when headers aren't available).
 */
export function incrementUsage(): void {
  // Check if windows have reset
  const now = Date.now()
  const fifteenMinutes = 15 * 60 * 1000

  // Reset 15-min window if expired
  if (now - state.windowStart15Min >= fifteenMinutes) {
    state.usage15Min = 0
    state.windowStart15Min = now
  }

  // Reset daily window if day changed
  const currentDayStart = getStartOfDay()
  if (currentDayStart > state.dayStart) {
    state.usageDaily = 0
    state.dayStart = currentDayStart
  }

  state.usage15Min++
  state.usageDaily++
  state.lastUpdate = now

  persistState()
}

/**
 * Check if we should wait before making a request.
 * Returns wait time in milliseconds, or 0 if no wait needed.
 */
export function getWaitTime(): number {
  const now = Date.now()
  const fifteenMinutes = 15 * 60 * 1000

  // Check if windows have reset
  if (now - state.windowStart15Min >= fifteenMinutes) {
    state.usage15Min = 0
    state.windowStart15Min = now
  }

  const currentDayStart = getStartOfDay()
  if (currentDayStart > state.dayStart) {
    state.usageDaily = 0
    state.dayStart = currentDayStart
  }

  const threshold15 = Math.floor(state.limit15Min * THRESHOLD_PERCENT)
  const thresholdDaily = Math.floor(state.limitDaily * THRESHOLD_PERCENT)

  // At or over 15-min limit
  if (state.usage15Min >= state.limit15Min) {
    const windowEnd = state.windowStart15Min + fifteenMinutes
    return Math.max(0, windowEnd - now)
  }

  // At or over daily limit
  if (state.usageDaily >= state.limitDaily) {
    const nextDay = state.dayStart + 24 * 60 * 60 * 1000
    return Math.max(0, nextDay - now)
  }

  // At 90% of 15-min limit, add small delay
  if (state.usage15Min >= threshold15) {
    return 1000 // 1 second
  }

  // At 90% of daily limit, add small delay
  if (state.usageDaily >= thresholdDaily) {
    return 1000 // 1 second
  }

  return 0
}

/**
 * Calculate wait time after receiving a 429 error.
 */
export function getRetryAfter(): number {
  const now = Date.now()
  const fifteenMinutes = 15 * 60 * 1000

  // If we know when the window started, wait for it to reset
  const windowEnd = state.windowStart15Min + fifteenMinutes
  const waitTime = Math.max(0, windowEnd - now)

  // If window already reset, wait a full 15 minutes as fallback
  return waitTime > 0 ? waitTime : fifteenMinutes
}

/**
 * Wait for rate limit to reset. Returns a promise that resolves when ready.
 */
export async function wait(): Promise<void> {
  const waitTime = getWaitTime()
  if (waitTime <= 0) return

  waitingUntil = new Date(Date.now() + waitTime)
  waitingReason =
    state.usage15Min >= state.limit15Min
      ? '15-minute rate limit reached'
      : state.usageDaily >= state.limitDaily
        ? 'Daily rate limit reached'
        : 'Approaching rate limit'

  console.log(`[RateLimit] Waiting ${Math.ceil(waitTime / 1000)}s - ${waitingReason}`)

  await sleep(waitTime)

  waitingUntil = null
  waitingReason = null
}

/**
 * Wait after receiving a 429 error.
 */
export async function waitForRetry(): Promise<void> {
  const waitTime = getRetryAfter()

  waitingUntil = new Date(Date.now() + waitTime)
  waitingReason = '429 rate limit hit, waiting for reset'

  console.log(`[RateLimit] 429 received, waiting ${Math.ceil(waitTime / 1000)}s`)

  await sleep(waitTime)

  // Reset the 15-min usage since window should have reset
  state.usage15Min = 0
  state.windowStart15Min = Date.now()

  waitingUntil = null
  waitingReason = null
}

/**
 * Get current rate limit info for UI display.
 */
export function getRateLimitInfo(): {
  usage15Min: number
  limit15Min: number
  usageDaily: number
  limitDaily: number
  waitingForRateLimit: boolean
  waitingUntil: string | undefined
  waitingReason: string | undefined
} {
  return {
    usage15Min: state.usage15Min,
    limit15Min: state.limit15Min,
    usageDaily: state.usageDaily,
    limitDaily: state.limitDaily,
    waitingForRateLimit: waitingUntil !== null,
    waitingUntil: waitingUntil?.toISOString(),
    waitingReason: waitingReason ?? undefined,
  }
}

/**
 * Persist rate limit state to database.
 */
function persistState(): void {
  try {
    const db = getDatabase()
    if (!db.isInitialized()) return

    const json = JSON.stringify(state)
    db.exec(
      `INSERT INTO app_state (key, value) VALUES ('strava_rate_limit', ?)
       ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`,
      [json]
    )
    // Don't await persist here to avoid slowing down requests
  } catch (err) {
    console.warn('[RateLimit] Failed to persist state:', err)
  }
}

/**
 * Load rate limit state from database.
 */
export function loadState(): void {
  try {
    const db = getDatabase()
    if (!db.isInitialized()) return

    const row = db.queryOne<{ value: string }>(`SELECT value FROM app_state WHERE key = 'strava_rate_limit'`)

    if (row?.value) {
      const saved = JSON.parse(row.value) as RateLimitState
      const now = Date.now()
      const fifteenMinutes = 15 * 60 * 1000

      // Restore state but decay usage if windows have reset
      state.limit15Min = saved.limit15Min || DEFAULT_LIMIT_15MIN
      state.limitDaily = saved.limitDaily || DEFAULT_LIMIT_DAILY

      // Check if 15-min window has reset
      if (now - (saved.windowStart15Min || 0) >= fifteenMinutes) {
        state.usage15Min = 0
        state.windowStart15Min = now
      } else {
        state.usage15Min = saved.usage15Min || 0
        state.windowStart15Min = saved.windowStart15Min || now
      }

      // Check if day has changed
      const currentDayStart = getStartOfDay()
      if (currentDayStart > (saved.dayStart || 0)) {
        state.usageDaily = 0
        state.dayStart = currentDayStart
      } else {
        state.usageDaily = saved.usageDaily || 0
        state.dayStart = saved.dayStart || currentDayStart
      }

      state.lastUpdate = now

      console.log('[RateLimit] Loaded state:', {
        usage15Min: state.usage15Min,
        usageDaily: state.usageDaily,
      })
    }
  } catch (err) {
    console.warn('[RateLimit] Failed to load state:', err)
  }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
