// Types for wrapped (year-in-review) data structures.
//
// Type Provenance:
// - Generated types: Re-exported from types.gen.ts (generated from Go structs)
// - Frontend-only types: Defined here for specific TS requirements

// ============================================================================
// Generated from Go - re-exported from types.gen.ts
// DO NOT modify these - regenerate with: just generate-ts-types
// ============================================================================
export type {
  WrappedMonth,
  WrappedTotals,
  WrappedSportTime,
  WrappedHourCount,
  WrappedLocationPoint,
  WrappedStreaks,
  WrappedPhoto,
  WrappedBiggestActivity,
} from '@/lib/wasm/types.gen'

// ============================================================================
// Frontend-only types - defined here for specific TS requirements
// Safe to modify as needed
// ============================================================================

// WrappedBiggest - uses undefined instead of null for optional fields
// (generated type uses | null for Go pointer types)
export interface WrappedBiggest {
  longest_distance?: import('@/lib/wasm/types.gen').WrappedBiggestActivity
  most_elevation?: import('@/lib/wasm/types.gen').WrappedBiggestActivity
  longest_duration?: import('@/lib/wasm/types.gen').WrappedBiggestActivity
}

// WrappedReport - uses custom WrappedBiggest type with undefined optionals
export interface WrappedReport {
  year: number
  range_start: string
  range_end: string
  total_days: number
  active_days: number
  rest_days: number
  totals: import('@/lib/wasm/types.gen').WrappedTotals
  months?: import('@/lib/wasm/types.gen').WrappedMonth[]
  moving_time_by_sport: import('@/lib/wasm/types.gen').WrappedSportTime[]
  start_times_by_hour: import('@/lib/wasm/types.gen').WrappedHourCount[]
  locations: import('@/lib/wasm/types.gen').WrappedLocationPoint[]
  streaks: import('@/lib/wasm/types.gen').WrappedStreaks
  random_photo?: import('@/lib/wasm/types.gen').WrappedPhoto
  biggest: WrappedBiggest
}

export { useWrappedYears, useWrappedReport as useWrapped } from '@/lib/data'
