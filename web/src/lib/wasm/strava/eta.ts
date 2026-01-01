/**
 * ETA Estimator - Estimates import completion time based on rate limits.
 *
 * Ported from internal/importer/eta.go
 */

export interface ETAState {
  fifteenMinLimit: number
  fifteenMinUsed: number
  fifteenMinReset: number | null // timestamp
  dailyLimit: number
  dailyUsed: number
  dailyReset: number | null // timestamp
  avgCallDuration: number // milliseconds
}

/**
 * Create a new ETA estimator with default values.
 */
export function createETAEstimator(): ETAState {
  return {
    fifteenMinLimit: 100,
    fifteenMinUsed: 0,
    fifteenMinReset: null,
    dailyLimit: 1000,
    dailyUsed: 0,
    dailyReset: null,
    avgCallDuration: 500, // 500ms average
  }
}

/**
 * Update the estimator from rate limit info.
 */
export function updateFromRateLimits(
  state: ETAState,
  usage15Min: number,
  limit15Min: number,
  usageDaily: number,
  limitDaily: number
): void {
  state.fifteenMinUsed = usage15Min
  state.fifteenMinLimit = limit15Min
  state.dailyUsed = usageDaily
  state.dailyLimit = limitDaily
}

/**
 * Get remaining calls in current 15-min window.
 */
function fifteenMinRemaining(state: ETAState): number {
  return state.fifteenMinLimit - state.fifteenMinUsed
}

/**
 * Get remaining calls in current day.
 */
function dailyRemaining(state: ETAState): number {
  return state.dailyLimit - state.dailyUsed
}

/**
 * Estimate time to complete given number of API calls.
 * Returns duration in milliseconds.
 */
export function estimateCompletion(state: ETAState, remainingCalls: number): number {
  if (remainingCalls <= 0) {
    return 0
  }

  const now = Date.now()
  const fifteenMin = 15 * 60 * 1000
  const fifteenMinRemain = fifteenMinRemaining(state)
  const dailyRemain = dailyRemaining(state)

  // If we can complete in current 15-min window
  if (remainingCalls <= fifteenMinRemain && remainingCalls <= dailyRemain) {
    return remainingCalls * state.avgCallDuration
  }

  // Calculate time accounting for rate limit windows
  let totalDuration = 0
  let callsLeft = remainingCalls
  let dailyRem = dailyRemain

  // Use remaining calls in current 15-min window
  if (fifteenMinRemain > 0 && dailyRemain > 0) {
    const usable = Math.min(fifteenMinRemain, dailyRemain, callsLeft)
    totalDuration += usable * state.avgCallDuration
    callsLeft -= usable
    dailyRem -= usable
  }

  if (callsLeft <= 0) {
    return totalDuration
  }

  // Wait for current 15-min window to reset
  if (state.fifteenMinReset && state.fifteenMinReset > now) {
    totalDuration += state.fifteenMinReset - now
  } else {
    totalDuration += fifteenMin
  }

  // Calculate full 15-min windows needed
  const callsPerWindow = Math.min(state.fifteenMinLimit, dailyRem)
  if (callsPerWindow <= 0) {
    // Daily limit exhausted, need to wait for daily reset
    if (state.dailyReset && state.dailyReset > now) {
      return totalDuration + (state.dailyReset - now)
    }
    return totalDuration + 24 * 60 * 60 * 1000
  }

  const fullWindows = Math.floor(callsLeft / callsPerWindow)
  totalDuration += fullWindows * fifteenMin
  callsLeft -= fullWindows * callsPerWindow

  // Partial final window
  if (callsLeft > 0) {
    totalDuration += callsLeft * state.avgCallDuration
  }

  return totalDuration
}

/**
 * Format the ETA duration as human-readable string.
 */
export function formatETA(durationMs: number): string {
  if (durationMs <= 0) {
    return '< 1 min'
  }

  const hours = Math.floor(durationMs / (60 * 60 * 1000))
  const minutes = Math.floor((durationMs % (60 * 60 * 1000)) / (60 * 1000))

  if (hours > 24) {
    const days = Math.floor(hours / 24)
    const remainingHours = hours % 24
    if (remainingHours > 0) {
      return `${formatPlural(days, 'day')} ${formatPlural(remainingHours, 'hour')}`
    }
    return formatPlural(days, 'day')
  }

  if (hours > 0) {
    if (minutes > 0) {
      return `${formatPlural(hours, 'hour')} ${formatPlural(minutes, 'min')}`
    }
    return formatPlural(hours, 'hour')
  }

  if (minutes > 0) {
    return formatPlural(minutes, 'min')
  }

  return '< 1 min'
}

function formatPlural(n: number, unit: string): string {
  if (n === 1) {
    return `1 ${unit}`
  }
  return `${n} ${unit}s`
}
