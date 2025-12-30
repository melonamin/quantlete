// Types for dashboard-related data structures. Hook implementations are provided
// by the shared data layer to support both server and WASM deployments.

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
  month: string
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
  date: string
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
  total_elevation_gain: number
}

export interface CalendarMonthSummary {
  year: number
  month: number
  activity_count: number
  total_distance: number
  total_elevation_gain: number
  total_moving_time: number
  total_calories: number
  workout_count: number
  challenges_completed: number
}

export interface HeatmapActivity {
  id: number
  name: string
  sport_type: string
  start_date: string
  distance: number
  summary_polyline: string
  start_lat: number
  start_lng: number
}

export interface HeatmapResponse {
  activities: HeatmapActivity[]
  total: number
  countries?: { country: string; iso2?: string; count: number }[]
}

export interface HeatmapFilters {
  sport_type?: string
  after?: string
  before?: string
  commute?: boolean
  workout_type?: number
}

export interface EddingtonDay {
  date: string
  distance: number
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

export interface EddingtonHistoryPoint {
  date: string
  number: number
}

export type WidgetWidth = 4 | 6 | 8 | 12
export type WidgetHeight = 1 | 2 | 3

export interface DashboardWidgetConfig {
  id: string
  width: WidgetWidth
  height?: WidgetHeight
  hidden: boolean
  settings?: Record<string, unknown>
}

export interface DashboardConfig {
  version: number
  widgets: DashboardWidgetConfig[]
}

export {
  useDashboard,
  useDashboardStats,
  useWeeklyStats,
  useRecentActivities,
  useSportTypeStats,
  useMonthlyStats,
  useYearlyStats,
  useCalendarData,
  useCalendarActivities,
  useCalendarSummary,
  useHeatmap,
  useEddington,
  useEddingtonHistory,
  useDashboardConfig,
  useUpdateDashboardConfig,
} from '@/lib/data'
