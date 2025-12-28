import type { UnitSystem } from '@/stores/settings'

// Distance formatting
const METERS_PER_MILE = 1609.344
const METERS_PER_FOOT = 0.3048

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
