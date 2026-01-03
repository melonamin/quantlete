// Types for rewind (year-in-review) data structures.
//
// Type Provenance:
// - Generated types: Re-exported from types.gen.ts (generated from Go structs)
// - Frontend-only types: Defined here for specific TS requirements

// ============================================================================
// Generated from Go - re-exported from types.gen.ts
// DO NOT modify these - regenerate with: just generate-ts-types
// ============================================================================
export type {
  RewindMonth,
  RewindTotals,
  RewindSportTime,
  RewindHourCount,
  RewindLocationPoint,
  RewindStreaks,
  RewindPhoto,
  RewindBiggestActivity,
} from '@/lib/wasm/types.gen'

// ============================================================================
// Frontend-only types - defined here for specific TS requirements
// Safe to modify as needed
// ============================================================================

// RewindBiggest - uses undefined instead of null for optional fields
// (generated type uses | null for Go pointer types)
export interface RewindBiggest {
  longest_distance?: import('@/lib/wasm/types.gen').RewindBiggestActivity
  most_elevation?: import('@/lib/wasm/types.gen').RewindBiggestActivity
  longest_duration?: import('@/lib/wasm/types.gen').RewindBiggestActivity
}

// RewindReport - uses custom RewindBiggest type with undefined optionals
export interface RewindReport {
  year: number
  range_start: string
  range_end: string
  total_days: number
  active_days: number
  rest_days: number
  totals: import('@/lib/wasm/types.gen').RewindTotals
  months?: import('@/lib/wasm/types.gen').RewindMonth[]
  moving_time_by_sport: import('@/lib/wasm/types.gen').RewindSportTime[]
  start_times_by_hour: import('@/lib/wasm/types.gen').RewindHourCount[]
  locations: import('@/lib/wasm/types.gen').RewindLocationPoint[]
  streaks: import('@/lib/wasm/types.gen').RewindStreaks
  random_photo?: import('@/lib/wasm/types.gen').RewindPhoto
  biggest: RewindBiggest
}

export { useRewindYears, useRewindReport as useRewind } from '@/lib/data'
