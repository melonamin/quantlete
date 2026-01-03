// Types for statistics-related data structures.
//
// Type Provenance:
// - Generated types: Re-exported from types.gen.ts (generated from Go structs)
// - Frontend-only types: Defined here with specific TS requirements or composite types

import { usePowerStats as dataUsePowerStats } from '@/lib/data'

// ============================================================================
// Generated from Go - re-exported from types.gen.ts
// DO NOT modify these - regenerate with: just generate-ts-types
// ============================================================================
export type {
  PeakPowerBest,
  PeakPowerHistoryPoint,
  PowerStatsResponse,
  DailyTrainingLoadPoint,
  TrainingLoadResponse,
  PowerZonesResponse,
  DistributionSlice,
} from '@/lib/wasm/types.gen'

// ============================================================================
// Frontend-only types - defined here for specific TS requirements
// Safe to modify as needed
// ============================================================================

// HrZonesResponse - manual definition for strongly-typed nested zone structure
// (generated type has zones: unknown)
export interface HrZonesResponse {
  method: string
  zones: { bounds: number[]; hr_max?: number }
  seconds_by_zone: number[]
  total_seconds: number
}

// HrZoneDefinition - manual definition for strongly-typed nested zone structure
// (generated type has zones: unknown)
export interface HrZoneDefinition {
  sport_type: string
  effective_from: string
  method: string
  zones: { bounds: number[]; hr_max?: number }
}

export {
  usePowerStats,
  usePowerZones,
  useHrZones,
  useTrainingLoad,
  useHrZoneDefinitions,
  useUpsertHrZoneDefinition,
  useDeleteHrZoneDefinition,
  useDaytimeDistribution,
  useWeekdayDistribution,
} from '@/lib/data'

export function usePowerStatsFiltered(filters?: {
  after?: string
  before?: string
  sport_type?: string
}) {
  return dataUsePowerStats(filters)
}
