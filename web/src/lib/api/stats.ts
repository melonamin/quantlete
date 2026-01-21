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
  DailyTrainingLoadPoint,
  PowerZonesResponse,
  DistributionSlice,
  WeeklyZoneDistribution,
  TrainingLoadDiagnostics,
} from '@/lib/wasm/types.gen'

// ============================================================================
// Insight types - enhanced with string literal types for type safety
// The generated types use generic strings; these provide better type checking
// ============================================================================
export type InsightType = 'fatigue' | 'recovery' | 'fitness'
export type InsightSeverity = 'info' | 'warning' | 'success'

export interface Insight {
  id: string
  type: InsightType
  severity: InsightSeverity
  title: string
  description: string
}

export interface InsightsResponse {
  insights: Insight[]
}

// ============================================================================
// Composite response types - not generated from Go
// These aggregate generated types into API response structures
// ============================================================================

// PowerStatsResponse - composite type combining best efforts and history
export interface PowerStatsResponse {
  best: import('@/lib/wasm/types.gen').PeakPowerBest[]
  durations_s: number[]
  history: Record<string, import('@/lib/wasm/types.gen').PeakPowerHistoryPoint[]>
}

// TrainingLoadResponse - composite type for training load data
export interface TrainingLoadResponse {
  series: import('@/lib/wasm/types.gen').DailyTrainingLoadPoint[]
  summary?: import('@/lib/wasm/types.gen').DailyTrainingLoadPoint
  diagnostics?: import('@/lib/wasm/types.gen').TrainingLoadDiagnostics
}

// ZoneTrendResponse - zone distribution data over time
export interface ZoneTrendResponse {
  weeks: import('@/lib/wasm/types.gen').WeeklyZoneDistribution[]
}

// WeeklyTrendsResponse - rolling weekly trends data
export interface WeeklyTrendsResponse {
  weeks: import('@/lib/wasm/types.gen').WeeklyTrendPoint[]
}

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
  useZoneTrend,
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
