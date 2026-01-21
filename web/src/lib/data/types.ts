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
  EddingtonCompareItem,
  EddingtonCompareOutput,
  SportGroup,
  DashboardConfig,
  DashboardWidgetConfig,
  WidgetWidth,
  WidgetHeight,
} from '@/lib/api/dashboard'

// Activity stream types
export type { ActivityStream } from '@/lib/api/activities'

// Activity analysis types
export type {
  SplitItem,
  SplitsOutput,
  ZoneItem,
  HRZonesOutput,
  PaceBucketItem,
  PaceDistributionOutput,
  ActivityAnalysis,
} from '@/lib/api/activities'

// Stats types
export type {
  PeakPowerBest,
  PeakPowerHistoryPoint,
  PowerStatsResponse,
  HrZonesResponse,
  DailyTrainingLoadPoint,
  TrainingLoadResponse,
  ZoneTrendResponse,
  WeeklyTrendsResponse,
  MonthlyComparisonResponse,
  WeeklyZoneDistribution,
  PowerZonesResponse,
  HrZoneDefinition,
  DistributionSlice,
  InsightType,
  InsightSeverity,
  Insight,
  InsightsResponse,
} from '@/lib/api/stats'

// Gear types
export type {
  Gear,
  GearFilters,
  GearListResponse,
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
  SegmentsResponse,
  SegmentDetailResponse,
  SegmentEffortsFilters,
  SegmentEffortsResponse,
} from '@/lib/api/segments'

// Athlete types
export type { MetricPoint, FTPHistoryResponse, WeightHistoryResponse } from '@/lib/api/athlete'

// Best efforts types
export type { BestEffortPR, BestEffortItem } from '@/lib/api/best-efforts'

// Wrapped types
export type { WrappedReport } from '@/lib/api/wrapped'

// Photo types
export type {
  PhotoListItem,
  PhotosListResponse,
  PhotosFilters,
  ActivityPhoto,
} from '@/lib/api/photos'

// Challenge types
export type { Challenge, ChallengesFilters, ChallengesResponse } from '@/lib/api/challenges'

// Goals types
export type { GoalPeriod, TrainingGoalsConfig, TrainingGoalsResponse } from '@/lib/api/goals'

// Maintenance types
export type {
  MaintenanceRule,
  ComponentWithRules,
  DueComponent,
  CreateComponentRequest,
  UpdateComponentRequest,
  LogMaintenanceRequest,
  ComponentsFilters,
  ComponentsResponse,
} from '@/lib/api/maintenance'

// Settings types
export type { AppSettings } from '@/lib/api/settings'

// Import types
export type {
  ImportProgress,
  StartImportRequest,
  ImportPhase,
  SyncRun,
  SyncWatermark,
} from '@/lib/api/import'

// Export helpers
export type { ExportStats } from '@/lib/api/export'

// Weather types
export type { ActivityWeather } from '@/lib/api/weather'

// Setup types
export type { CredentialsStatus, UpdateCredentialsRequest } from '@/lib/api/setup'
