/**
 * Re-export all API types for use by DataProvider implementations.
 */

// Core types
export type {
  Activity,
  ActivityFilters,
  ActivitiesResponse,
  PaginatedResponse,
  AuthStatus,
} from '@/lib/api/types'

// Dashboard types
export type {
  DashboardData,
  DashboardStats,
  WeeklyStat,
  RecentActivity,
  SportTypeStat,
  MonthlyStat,
  YearlyStat,
  CalendarDay,
  CalendarActivity,
  CalendarMonthSummary,
  HeatmapActivity,
  HeatmapResponse,
  HeatmapFilters,
  EddingtonDay,
  EddingtonStep,
  EddingtonResult,
  EddingtonHistoryPoint,
  DashboardConfig,
  DashboardWidgetConfig,
  WidgetWidth,
  WidgetHeight,
} from '@/lib/api/dashboard'

// Activity stream types
export type { ActivityStream } from '@/lib/api/activities'

// Stats types
export type {
  PeakPowerBest,
  PeakPowerHistoryPoint,
  PowerStatsResponse,
  HrZonesResponse,
  DailyTrainingLoadPoint,
  TrainingLoadResponse,
  PowerZonesResponse,
  HrZoneDefinition,
  DistributionSlice,
} from '@/lib/api/stats'

// Gear types
export type {
  Gear,
  CustomGearCreateRequest,
  GearMonthlyUsage,
} from '@/lib/api/gear'

// Segment types
export type {
  Segment,
  SegmentListItem,
  SegmentEffort,
  SegmentCountryStat,
  SegmentsFilters,
  SegmentDetailResponse,
} from '@/lib/api/segments'

// Athlete types
export type {
  MetricPoint,
  FTPHistoryResponse,
  WeightHistoryResponse,
} from '@/lib/api/athlete'

// Best efforts types
export type { BestEffortPR, BestEffortItem } from '@/lib/api/best-efforts'

// Rewind types
export type { RewindReport } from '@/lib/api/rewind'

// Photo types
export type {
  PhotoListItem,
  PhotosListResponse,
  PhotosFilters,
  ActivityPhoto,
} from '@/lib/api/photos'

// Challenge types
export type { Challenge } from '@/lib/api/challenges'

// Goals types
export type {
  GoalPeriod,
  TrainingGoalsConfig,
  TrainingGoalsResponse,
} from '@/lib/api/goals'

// Maintenance types
export type {
  MaintenanceRule,
  ComponentWithRules,
  DueComponent,
  CreateComponentRequest,
  UpdateComponentRequest,
  LogMaintenanceRequest,
} from '@/lib/api/maintenance'

// Settings types
export type { AppSettings } from '@/lib/api/settings'

// Import types
export type { ImportProgress, StartImportRequest, ImportPhase } from '@/lib/api/import'
