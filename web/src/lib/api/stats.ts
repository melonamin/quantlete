import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { del, get, put } from './client'

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

export const statsKeys = {
  all: ['stats'] as const,
  power: (filters: { after?: string; before?: string; sport_type?: string } = {}) =>
    [...statsKeys.all, 'power', filters] as const,
  hrZones: (filters: { after?: string; before?: string; sport_type?: string } = {}) =>
    [...statsKeys.all, 'hrZones', filters] as const,
  trainingLoad: (filters: { after?: string; before?: string } = {}) =>
    [...statsKeys.all, 'trainingLoad', filters] as const,
  hrZoneDefs: () => [...statsKeys.all, 'hrZoneDefs'] as const,
  daytime: () => [...statsKeys.all, 'daytime'] as const,
  weekday: () => [...statsKeys.all, 'weekday'] as const,
  powerZones: () => [...statsKeys.all, 'powerZones'] as const,
}

export function usePowerStats() {
  return useQuery({
    queryKey: statsKeys.power(),
    queryFn: () => get<PowerStatsResponse>('/stats/power'),
  })
}

export function usePowerStatsFiltered(
  filters: { after?: string; before?: string; sport_type?: string } = {}
) {
  return useQuery({
    queryKey: statsKeys.power(filters),
    queryFn: () => {
      const params = new URLSearchParams()
      if (filters.after) params.set('after', filters.after)
      if (filters.before) params.set('before', filters.before)
      if (filters.sport_type) params.set('sport_type', filters.sport_type)
      const qs = params.toString()
      return get<PowerStatsResponse>(`/stats/power${qs ? `?${qs}` : ''}`)
    },
  })
}

export function usePowerZones() {
  return useQuery({
    queryKey: statsKeys.powerZones(),
    queryFn: () => get<PowerZonesResponse>('/stats/power-zones'),
  })
}

export function useHrZones(filters: { after?: string; before?: string; sport_type?: string } = {}) {
  return useQuery({
    queryKey: statsKeys.hrZones(filters),
    queryFn: () => {
      const params = new URLSearchParams()
      if (filters.after) params.set('after', filters.after)
      if (filters.before) params.set('before', filters.before)
      if (filters.sport_type) params.set('sport_type', filters.sport_type)
      const qs = params.toString()
      return get<HrZonesResponse>(`/stats/hr-zones${qs ? `?${qs}` : ''}`)
    },
  })
}

export function useTrainingLoad(filters: { after?: string; before?: string } = {}) {
  return useQuery({
    queryKey: statsKeys.trainingLoad(filters),
    queryFn: () => {
      const params = new URLSearchParams()
      if (filters.after) params.set('after', filters.after)
      if (filters.before) params.set('before', filters.before)
      const qs = params.toString()
      return get<TrainingLoadResponse>(`/stats/training-load${qs ? `?${qs}` : ''}`)
    },
  })
}

export function useHrZoneDefinitions(opts: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: statsKeys.hrZoneDefs(),
    queryFn: () => get<HrZoneDefinition[]>('/zones/hr'),
    enabled: opts.enabled ?? true,
    retry: false,
  })
}

export function useUpsertHrZoneDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (def: HrZoneDefinition) => put<{ status: string }>('/zones/hr', def),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: statsKeys.hrZoneDefs() })
      qc.invalidateQueries({ queryKey: ['stats', 'hrZones'] as const })
    },
  })
}

export function useDeleteHrZoneDefinition() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (params: { sport_type: string; effective_from: string }) =>
      del<{ status: string }>(
        `/zones/hr?sport_type=${encodeURIComponent(params.sport_type)}&effective_from=${encodeURIComponent(params.effective_from)}`
      ),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: statsKeys.hrZoneDefs() })
      qc.invalidateQueries({ queryKey: ['stats', 'hrZones'] as const })
    },
  })
}

export function useDaytimeDistribution() {
  return useQuery({
    queryKey: statsKeys.daytime(),
    queryFn: () => get<DistributionSlice[]>('/stats/daytime'),
  })
}

export function useWeekdayDistribution() {
  return useQuery({
    queryKey: statsKeys.weekday(),
    queryFn: () => get<DistributionSlice[]>('/stats/weekday'),
  })
}
