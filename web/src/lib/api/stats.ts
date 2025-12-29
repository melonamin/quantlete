import { usePowerStats as dataUsePowerStats } from '@/lib/data'

export interface PeakPowerBest {
  duration_s: number
  watts: number
  activity_id: number
  start_date: string
}

export interface PeakPowerHistoryPoint {
  date: string
  watts: number
}

export interface PowerStatsResponse {
  durations_s: number[]
  best: PeakPowerBest[]
  history: Record<string, PeakPowerHistoryPoint[]>
}

export interface HrZonesResponse {
  method: string
  zones: { bounds: number[]; hr_max?: number }
  seconds_by_zone: number[]
  total_seconds: number
}

export interface DailyTrainingLoadPoint {
  day: string
  tss: number
  ctl: number
  atl: number
  tsb: number
}

export interface TrainingLoadResponse {
  series: DailyTrainingLoadPoint[]
  summary?: DailyTrainingLoadPoint
}

export interface PowerZonesResponse {
  ftp_watts?: number
  seconds_by_zone: number[]
  total_seconds: number
  bounds: number[]
}

export interface HrZoneDefinition {
  sport_type: string
  effective_from: string
  method: string
  zones: { bounds: number[]; hr_max?: number }
}

export interface DistributionSlice {
  label: string
  count: number
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
