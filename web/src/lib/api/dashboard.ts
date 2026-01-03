// Types for dashboard-related data structures.
// Re-exported from generated types for single source of truth.

export type {
  DashboardStats,
  WeeklyStat,
  RecentActivity,
  SportTypeStat,
  MonthlyStat,
  CalendarDay,
  CalendarActivity,
  CalendarMonthSummary,
  HeatmapActivity,
  HeatmapResponse,
  EddingtonDay,
  EddingtonStep,
  EddingtonResult,
  EddingtonHistoryPoint,
} from '@/lib/wasm/types.gen'

// Widget types with stricter width/height unions (generated uses plain number)
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

// DashboardData combines multiple dashboard stats - not in generated types
export interface DashboardData {
  stats: import('@/lib/wasm/types.gen').DashboardStats
  weekly_stats: import('@/lib/wasm/types.gen').WeeklyStat[]
  recent_activities: import('@/lib/wasm/types.gen').RecentActivity[]
  sport_type_stats: import('@/lib/wasm/types.gen').SportTypeStat[]
}

// YearlyStat - not in generated types (uses number for year)
export interface YearlyStat {
  year: number
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

// HeatmapFilters - not in generated types (query-only)
export interface HeatmapFilters {
  sport_type?: string
  after?: string
  before?: string
  commute?: boolean
  workout_type?: number
  limit?: number
  offset?: number
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
