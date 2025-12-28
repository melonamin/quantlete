import { useQuery } from '@tanstack/react-query'
import { get } from './client'

export interface RewindMonth {
  month: string
  activities: number
  distance_m: number
  elevation_m: number
  prs: number
}

export interface RewindTotals {
  activities: number
  distance_m: number
  elevation_m: number
  moving_time_s: number
  kudos: number
  commute_distance_m: number
  carbon_saved_kg: number
}

export interface RewindSportTime {
  sport_type: string
  moving_time_s: number
}

export interface RewindHourCount {
  hour: number
  count: number
}

export interface RewindLocationPoint {
  lat: number
  lng: number
  count: number
}

export interface RewindStreaks {
  longest_active_days: number
  longest_rest_days: number
}

export interface RewindPhoto {
  id: string
  activity_id: number
  url: string
  thumbnail_url?: string
  caption?: string
}

export interface RewindBiggestActivity {
  activity_id: number
  name: string
  sport_type: string
  start_date_local: string
  value: number
}

export interface RewindReport {
  year: number
  range_start: string
  range_end: string
  total_days: number
  active_days: number
  rest_days: number
  totals: RewindTotals
  months?: RewindMonth[]
  moving_time_by_sport: RewindSportTime[]
  start_times_by_hour: RewindHourCount[]
  locations: RewindLocationPoint[]
  streaks: RewindStreaks
  random_photo?: RewindPhoto
  biggest: {
    longest_distance?: RewindBiggestActivity
    most_elevation?: RewindBiggestActivity
    longest_duration?: RewindBiggestActivity
  }
}

export const rewindKeys = {
  all: ['rewind'] as const,
  years: () => [...rewindKeys.all, 'years'] as const,
  report: (year: number) => [...rewindKeys.all, 'report', year] as const,
}

export function useRewindYears() {
  return useQuery({
    queryKey: rewindKeys.years(),
    queryFn: () => get<number[]>('/stats/rewind/years'),
    retry: false,
  })
}

export function useRewind(year: number, enabled = true) {
  return useQuery({
    queryKey: rewindKeys.report(year),
    queryFn: () => get<RewindReport>(`/stats/rewind?year=${encodeURIComponent(String(year))}`),
    enabled,
    retry: false,
  })
}
