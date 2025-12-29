import { Bike, Footprints, Waves, Snowflake, Activity } from 'lucide-react'

export type SportCategory = 'ride' | 'run' | 'walk' | 'swim' | 'winter' | 'other'

const sportCategoryMap: Record<string, SportCategory> = {
  // Rides
  Ride: 'ride',
  MountainBikeRide: 'ride',
  GravelRide: 'ride',
  EBikeRide: 'ride',
  EMountainBikeRide: 'ride',
  VirtualRide: 'ride',
  Velomobile: 'ride',
  Handcycle: 'ride',

  // Runs
  Run: 'run',
  TrailRun: 'run',
  VirtualRun: 'run',

  // Walks
  Walk: 'walk',
  Hike: 'walk',

  // Swimming
  Swim: 'swim',

  // Winter sports
  AlpineSki: 'winter',
  BackcountrySki: 'winter',
  NordicSki: 'winter',
  Snowboard: 'winter',
  Snowshoe: 'winter',
  IceSkate: 'winter',
}

export function getSportCategory(sportType: string): SportCategory {
  return sportCategoryMap[sportType] || 'other'
}

export function getSportColor(sportType: string): string {
  const category = getSportCategory(sportType)
  const colors: Record<SportCategory, string> = {
    ride: 'bg-sport-ride',
    run: 'bg-sport-run',
    walk: 'bg-sport-walk',
    swim: 'bg-sport-swim',
    winter: 'bg-sport-winter',
    other: 'bg-sport-other',
  }
  return colors[category]
}

export function getSportTextColor(sportType: string): string {
  const category = getSportCategory(sportType)
  const colors: Record<SportCategory, string> = {
    ride: 'text-sport-ride',
    run: 'text-sport-run',
    walk: 'text-sport-walk',
    swim: 'text-sport-swim',
    winter: 'text-sport-winter',
    other: 'text-sport-other',
  }
  return colors[category]
}

export function getSportIcon(sportType: string) {
  const category = getSportCategory(sportType)
  const icons: Record<SportCategory, typeof Activity> = {
    ride: Bike,
    run: Footprints,
    walk: Footprints,
    swim: Waves,
    winter: Snowflake,
    other: Activity,
  }
  return icons[category]
}

export function formatSportType(sportType: string): string {
  return sportType.replace(/([A-Z])/g, ' $1').trim()
}

// Hex colors for inline styles (when Tailwind classes aren't suitable)
export function getSportHexColor(sportType: string): string {
  const category = getSportCategory(sportType)
  const colors: Record<SportCategory, string> = {
    ride: '#f97316', // orange-500
    run: '#22c55e', // green-500
    walk: '#a855f7', // purple-500
    swim: '#3b82f6', // blue-500
    winter: '#06b6d4', // cyan-500
    other: '#6b7280', // gray-500
  }
  return colors[category]
}

export function getSportEmoji(sportType: string): string {
  const category = getSportCategory(sportType)
  const emojis: Record<SportCategory, string> = {
    ride: '🚴',
    run: '🏃',
    walk: '🚶',
    swim: '🏊',
    winter: '⛷️',
    other: '🏋️',
  }
  return emojis[category]
}
