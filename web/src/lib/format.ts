import type { UnitSystem } from '@/stores/settings'

// Conversion constants
const METERS_PER_MILE = 1609.344
const METERS_PER_KM = 1000
const METERS_PER_FOOT = 0.3048

/**
 * Convert distance from display units to meters.
 * @param value Distance in km (metric) or miles (imperial)
 * @param unitSystem The unit system being used
 * @returns Distance in meters
 */
export function distanceToMeters(value: number, unitSystem: UnitSystem): number {
  return unitSystem === 'imperial' ? value * METERS_PER_MILE : value * METERS_PER_KM
}

/**
 * Convert distance from meters to display units.
 * @param meters Distance in meters
 * @param unitSystem The unit system being used
 * @returns Distance in km (metric) or miles (imperial)
 */
export function metersToDisplayUnit(meters: number, unitSystem: UnitSystem): number {
  return unitSystem === 'imperial' ? meters / METERS_PER_MILE : meters / METERS_PER_KM
}

/**
 * Parse duration input in HH:MM or H:MM or MM format to seconds.
 * @param value Duration string (e.g., "1:30" for 1h30m, "45" for 45m)
 * @returns Duration in seconds, or null if invalid
 */
export function parseDurationInput(value: string): number | null {
  if (!value.trim()) return null

  // Handle HH:MM format
  if (value.includes(':')) {
    const parts = value.split(':')
    if (parts.length !== 2) return null
    const hours = parseInt(parts[0], 10)
    const minutes = parseInt(parts[1], 10)
    if (isNaN(hours) || isNaN(minutes) || hours < 0 || minutes < 0 || minutes >= 60) return null
    return hours * 3600 + minutes * 60
  }

  // Handle plain minutes
  const minutes = parseFloat(value)
  if (isNaN(minutes) || minutes < 0) return null
  return Math.round(minutes * 60)
}

/**
 * Format seconds to HH:MM or MM format for input display.
 * @param seconds Duration in seconds
 * @returns Formatted duration string (e.g., "1:30" for 1h30m)
 */
export function formatDurationInput(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}`
  }
  return minutes.toString()
}

export function formatDistance(meters: number, unitSystem: UnitSystem = 'metric'): string {
  if (unitSystem === 'imperial') {
    const miles = meters / METERS_PER_MILE
    if (miles < 0.1) {
      return `${Math.round(meters / METERS_PER_FOOT)} ft`
    }
    return `${miles.toFixed(2)} mi`
  }

  if (meters < 1000) {
    return `${Math.round(meters)} m`
  }
  return `${(meters / 1000).toFixed(2)} km`
}

// Elevation formatting
export function formatElevation(meters: number, unitSystem: UnitSystem = 'metric'): string {
  if (unitSystem === 'imperial') {
    return `${Math.round(meters / METERS_PER_FOOT)} ft`
  }
  return `${Math.round(meters)} m`
}

// Speed formatting (m/s input)
export function formatSpeed(metersPerSecond: number, unitSystem: UnitSystem): string {
  if (unitSystem === 'imperial') {
    const mph = metersPerSecond * 2.23694
    return `${mph.toFixed(1)} mph`
  }
  const kmh = metersPerSecond * 3.6
  return `${kmh.toFixed(1)} km/h`
}

// Pace formatting (m/s input, for running)
export function formatPace(metersPerSecond: number, unitSystem: UnitSystem): string {
  if (metersPerSecond === 0) return '--:--'

  const secondsPerMeter = 1 / metersPerSecond

  if (unitSystem === 'imperial') {
    const secondsPerMile = secondsPerMeter * METERS_PER_MILE
    const minutes = Math.floor(secondsPerMile / 60)
    const seconds = Math.round(secondsPerMile % 60)
    return `${minutes}:${seconds.toString().padStart(2, '0')}/mi`
  }

  const secondsPerKm = secondsPerMeter * 1000
  const minutes = Math.floor(secondsPerKm / 60)
  const seconds = Math.round(secondsPerKm % 60)
  return `${minutes}:${seconds.toString().padStart(2, '0')}/km`
}

export function formatPaceFromSecondsPerKm(
  secondsPerKm: number,
  unitSystem: UnitSystem
): string {
  if (!secondsPerKm || secondsPerKm <= 0 || !isFinite(secondsPerKm)) return '-'
  const metersPerSecond = METERS_PER_KM / secondsPerKm
  return formatPace(metersPerSecond, unitSystem)
}

// Duration formatting
export function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  if (hours > 0) {
    return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }
  return `${minutes}:${secs.toString().padStart(2, '0')}`
}

export function formatDurationLong(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}

// Date formatting
export function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatDateTime(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatRelativeDate(dateStr: string): string {
  const date = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

  if (diffDays === 0) return 'Today'
  if (diffDays === 1) return 'Yesterday'
  if (diffDays < 7) return `${diffDays} days ago`
  if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`
  if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`
  return `${Math.floor(diffDays / 365)} years ago`
}

// Number formatting
export function formatNumber(value: number, decimals = 0): string {
  return value.toLocaleString(undefined, {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals,
  })
}

// Unit label helpers (for axis labels, tooltips, etc.)
export function getDistanceUnit(unitSystem: UnitSystem): string {
  return unitSystem === 'imperial' ? 'mi' : 'km'
}

export function getElevationUnit(unitSystem: UnitSystem): string {
  return unitSystem === 'imperial' ? 'ft' : 'm'
}

export function getSpeedUnit(unitSystem: UnitSystem): string {
  return unitSystem === 'imperial' ? 'mph' : 'km/h'
}

export function getPaceUnit(unitSystem: UnitSystem): string {
  return unitSystem === 'imperial' ? '/mi' : '/km'
}
