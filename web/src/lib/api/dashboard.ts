import { useQuery } from '@tanstack/react-query'
import { get } from './client'

// Types
export interface DashboardStats {
  total_activities: number
  total_distance: number
  total_moving_time: number
  total_elevation_gain: number
  total_calories: number
  year_activities: number
  year_distance: number
  year_moving_time: number
  year_elevation_gain: number
  month_activities: number
  month_distance: number
  month_moving_time: number
  month_elevation_gain: number
}

export interface WeeklyStat {
  sport_type: string
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export interface RecentActivity {
  id: number
  name: string
  sport_type: string
  start_date: string
  distance: number
  moving_time: number
  elevation_gain: number
  summary_polyline?: string
}

export interface SportTypeStat {
  sport_type: string
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export interface DashboardData {
  stats: DashboardStats
  weekly_stats: WeeklyStat[]
  recent_activities: RecentActivity[]
  sport_type_stats: SportTypeStat[]
}

export interface MonthlyStat {
  month: string // YYYY-MM
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export interface YearlyStat {
  year: number
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export interface CalendarDay {
  date: string // YYYY-MM-DD
  activity_count: number
  total_distance: number
}

export interface CalendarActivity {
  id: number
  name: string
  sport_type: string
  start_date: string
  distance: number
  moving_time: number
}

export interface HeatmapActivity {
  id: number
  sport_type: string
  summary_polyline: string
  start_lat: number
  start_lng: number
}

export interface HeatmapResponse {
  activities: HeatmapActivity[]
  total: number
}

export interface HeatmapFilters {
  sport_type?: string
  after?: string
  before?: string
  commute?: boolean
}

export interface EddingtonDay {
  date: string
  distance: number // km
}

export interface EddingtonStep {
  target: number
  rides_needed: number
}

export interface EddingtonResult {
  number: number
  distribution: EddingtonDay[]
  next_steps: EddingtonStep[]
}

// Query keys
export const dashboardKeys = {
  all: ['dashboard'] as const,
  data: () => [...dashboardKeys.all, 'data'] as const,
  stats: () => [...dashboardKeys.all, 'stats'] as const,
  weekly: () => [...dashboardKeys.all, 'weekly'] as const,
  recent: (limit?: number) => [...dashboardKeys.all, 'recent', limit] as const,
  sports: () => [...dashboardKeys.all, 'sports'] as const,
  monthly: (year?: number) => [...dashboardKeys.all, 'monthly', year] as const,
  yearly: () => [...dashboardKeys.all, 'yearly'] as const,
  calendar: (year: number) => [...dashboardKeys.all, 'calendar', year] as const,
  calendarActivities: (year: number, month: number) => [...dashboardKeys.all, 'calendarActivities', year, month] as const,
  heatmap: (filters: HeatmapFilters) => [...dashboardKeys.all, 'heatmap', filters] as const,
  eddington: (sportType?: string) => [...dashboardKeys.all, 'eddington', sportType] as const,
}

// Hooks
export function useDashboard() {
  return useQuery({
    queryKey: dashboardKeys.data(),
    queryFn: () => get<DashboardData>('/dashboard'),
  })
}

export function useDashboardStats() {
  return useQuery({
    queryKey: dashboardKeys.stats(),
    queryFn: () => get<DashboardStats>('/dashboard/stats'),
  })
}

export function useWeeklyStats() {
  return useQuery({
    queryKey: dashboardKeys.weekly(),
    queryFn: () => get<WeeklyStat[]>('/dashboard/weekly'),
  })
}

export function useRecentActivities(limit = 5) {
  return useQuery({
    queryKey: dashboardKeys.recent(limit),
    queryFn: () => get<RecentActivity[]>(`/dashboard/recent?limit=${limit}`),
  })
}

export function useSportTypeStats() {
  return useQuery({
    queryKey: dashboardKeys.sports(),
    queryFn: () => get<SportTypeStat[]>('/dashboard/sports'),
  })
}

export function useMonthlyStats(year?: number) {
  return useQuery({
    queryKey: dashboardKeys.monthly(year),
    queryFn: () => get<MonthlyStat[]>(year ? `/dashboard/monthly?year=${year}` : '/dashboard/monthly'),
  })
}

export function useYearlyStats() {
  return useQuery({
    queryKey: dashboardKeys.yearly(),
    queryFn: () => get<YearlyStat[]>('/dashboard/yearly'),
  })
}

export function useCalendarData(year: number) {
  return useQuery({
    queryKey: dashboardKeys.calendar(year),
    queryFn: () => get<CalendarDay[]>(`/dashboard/calendar?year=${year}`),
  })
}

export function useCalendarActivities(year: number, month: number) {
  return useQuery({
    queryKey: dashboardKeys.calendarActivities(year, month),
    queryFn: () => get<CalendarActivity[]>(`/dashboard/calendar/activities?year=${year}&month=${month}`),
  })
}

export function useHeatmapData(filters: HeatmapFilters = {}) {
  return useQuery({
    queryKey: dashboardKeys.heatmap(filters),
    queryFn: () => {
      const params = new URLSearchParams()
      if (filters.sport_type) params.set('sport_type', filters.sport_type)
      if (filters.after) params.set('after', filters.after)
      if (filters.before) params.set('before', filters.before)
      if (filters.commute !== undefined) params.set('commute', String(filters.commute))

      const queryString = params.toString()
      return get<HeatmapResponse>(`/stats/heatmap${queryString ? `?${queryString}` : ''}`)
    },
  })
}

export function useEddingtonData(sportType?: string) {
  return useQuery({
    queryKey: dashboardKeys.eddington(sportType),
    queryFn: () => {
      const params = sportType ? `?sport_type=${sportType}` : ''
      return get<EddingtonResult>(`/stats/eddington${params}`)
    },
  })
}
