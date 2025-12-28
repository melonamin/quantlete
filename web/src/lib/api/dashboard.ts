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

// Query keys
export const dashboardKeys = {
  all: ['dashboard'] as const,
  data: () => [...dashboardKeys.all, 'data'] as const,
  stats: () => [...dashboardKeys.all, 'stats'] as const,
  weekly: () => [...dashboardKeys.all, 'weekly'] as const,
  recent: (limit?: number) => [...dashboardKeys.all, 'recent', limit] as const,
  sports: () => [...dashboardKeys.all, 'sports'] as const,
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
