/**
 * DataProvider interface - abstraction layer for data access.
 *
 * This interface is implemented by:
 * - ServerProvider: calls Go backend via HTTP
 * - WasmProvider: queries local sql.js database
 *
 * All methods return Promises to support both sync (WASM) and async (HTTP) operations.
 */

import type {
  // Core
  Activity,
  ActivityFilters,
  ActivitiesResponse,
  AuthStatus,
  // Dashboard
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
  HeatmapResponse,
  HeatmapFilters,
  EddingtonResult,
  EddingtonHistoryPoint,
  DashboardConfig,
  // Activity streams
  ActivityStream,
  // Stats
  PowerStatsResponse,
  HrZonesResponse,
  TrainingLoadResponse,
  PowerZonesResponse,
  HrZoneDefinition,
  DistributionSlice,
  // Gear
  Gear,
  CustomGearCreateRequest,
  GearMonthlyUsage,
  // Segments
  SegmentListItem,
  SegmentCountryStat,
  SegmentDetailResponse,
  SegmentEffort,
  SegmentsFilters,
  // Athlete
  FTPHistoryResponse,
  WeightHistoryResponse,
  // Best efforts
  BestEffortPR,
  BestEffortItem,
  // Rewind
  RewindReport,
  // Photos
  PhotosListResponse,
  PhotosFilters,
  ActivityPhoto,
  // Challenges
  Challenge,
  // Goals
  TrainingGoalsConfig,
  TrainingGoalsResponse,
  // Maintenance
  ComponentWithRules,
  DueComponent,
  CreateComponentRequest,
  UpdateComponentRequest,
  LogMaintenanceRequest,
  // Settings
  AppSettings,
  // Import
  ImportProgress,
  StartImportRequest,
} from './types'

export interface DataProvider {
  // ============================================================================
  // Auth
  // ============================================================================
  getAuthStatus(): Promise<AuthStatus>
  refreshToken(): Promise<{ success: boolean }>

  // ============================================================================
  // Activities
  // ============================================================================
  getActivities(filters: ActivityFilters): Promise<ActivitiesResponse>
  getActivity(id: number): Promise<Activity>
  getActivityStreams(id: number): Promise<ActivityStream[]>

  // ============================================================================
  // Dashboard
  // ============================================================================
  getDashboard(): Promise<DashboardData>
  getDashboardStats(): Promise<DashboardStats>
  getWeeklyStats(): Promise<WeeklyStat[]>
  getRecentActivities(limit: number): Promise<RecentActivity[]>
  getSportTypeStats(): Promise<SportTypeStat[]>
  getMonthlyStats(year?: number): Promise<MonthlyStat[]>
  getYearlyStats(): Promise<YearlyStat[]>
  getCalendarData(year: number): Promise<CalendarDay[]>
  getCalendarActivities(year: number, month: number): Promise<CalendarActivity[]>
  getCalendarSummary(year: number, month: number): Promise<CalendarMonthSummary>
  getDashboardConfig(): Promise<DashboardConfig>
  updateDashboardConfig(config: DashboardConfig): Promise<DashboardConfig>

  // ============================================================================
  // Heatmap & Eddington
  // ============================================================================
  getHeatmapData(filters: HeatmapFilters): Promise<HeatmapResponse>
  getEddingtonData(sportType?: string): Promise<EddingtonResult>
  getEddingtonHistory(sportType?: string): Promise<EddingtonHistoryPoint[]>

  // ============================================================================
  // Stats & Training
  // ============================================================================
  getPowerStats(filters?: {
    after?: string
    before?: string
    sport_type?: string
  }): Promise<PowerStatsResponse>
  getPowerZones(): Promise<PowerZonesResponse>
  getHrZones(filters?: {
    after?: string
    before?: string
    sport_type?: string
  }): Promise<HrZonesResponse>
  getTrainingLoad(filters?: { after?: string; before?: string }): Promise<TrainingLoadResponse>
  getHrZoneDefinitions(): Promise<HrZoneDefinition[]>
  upsertHrZoneDefinition(def: HrZoneDefinition): Promise<{ status: string }>
  deleteHrZoneDefinition(sportType: string, effectiveFrom: string): Promise<{ status: string }>
  getDaytimeDistribution(): Promise<DistributionSlice[]>
  getWeekdayDistribution(): Promise<DistributionSlice[]>

  // ============================================================================
  // Best Efforts
  // ============================================================================
  getBestEffortPRs(sportType?: string): Promise<BestEffortPR[]>
  getBestEffortsForDistance(distanceType: string, sportType?: string): Promise<BestEffortItem[]>

  // ============================================================================
  // Rewind
  // ============================================================================
  getRewindYears(): Promise<number[]>
  getRewind(year: number): Promise<RewindReport>

  // ============================================================================
  // Gear
  // ============================================================================
  getGear(includeRetired?: boolean): Promise<Gear[]>
  getGearDetail(id: string): Promise<Gear>
  getCustomGear(includeRetired?: boolean): Promise<Gear[]>
  createCustomGear(req: CustomGearCreateRequest): Promise<Gear>
  updateCustomGear(id: string, patch: Partial<CustomGearCreateRequest>): Promise<Gear>
  deleteCustomGear(id: string, force?: boolean): Promise<{ deleted: boolean }>
  getGearMonthlyUsage(includeRetired?: boolean): Promise<GearMonthlyUsage[]>

  // ============================================================================
  // Segments
  // ============================================================================
  getSegments(filters?: SegmentsFilters): Promise<SegmentListItem[]>
  getSegmentCountries(): Promise<SegmentCountryStat[]>
  getSegmentDetail(id: number): Promise<SegmentDetailResponse>
  getSegmentEfforts(id: number): Promise<SegmentEffort[]>

  // ============================================================================
  // Athlete (FTP/Weight)
  // ============================================================================
  getFtpHistory(): Promise<FTPHistoryResponse>
  updateFtpHistory(body: FTPHistoryResponse): Promise<FTPHistoryResponse>
  getWeightHistory(): Promise<WeightHistoryResponse>
  updateWeightHistory(body: WeightHistoryResponse): Promise<WeightHistoryResponse>

  // ============================================================================
  // Photos
  // ============================================================================
  getPhotos(filters?: PhotosFilters): Promise<PhotosListResponse>
  getActivityPhotos(activityId: number): Promise<ActivityPhoto[]>

  // ============================================================================
  // Challenges
  // ============================================================================
  getChallenges(month?: string): Promise<Challenge[]>
  importChallenges(file: File): Promise<{ imported: number }>

  // ============================================================================
  // Goals
  // ============================================================================
  getTrainingGoals(): Promise<TrainingGoalsResponse>
  updateTrainingGoals(config: TrainingGoalsConfig): Promise<TrainingGoalsConfig>

  // ============================================================================
  // Maintenance
  // ============================================================================
  getMaintenanceDue(): Promise<DueComponent[]>
  getGearComponents(gearId: string): Promise<ComponentWithRules[]>
  createComponent(gearId: string, req: CreateComponentRequest): Promise<ComponentWithRules>
  updateComponent(id: number, req: UpdateComponentRequest): Promise<ComponentWithRules>
  deleteComponent(id: number): Promise<{ deleted: boolean }>
  logMaintenance(componentId: number, req?: LogMaintenanceRequest): Promise<{ logged: boolean }>

  // ============================================================================
  // Settings
  // ============================================================================
  getAppSettings(): Promise<AppSettings>
  updateAppSettings(settings: AppSettings): Promise<AppSettings>

  // ============================================================================
  // Import (server mode only)
  // ============================================================================
  getImportProgress(): Promise<ImportProgress>
  startImport(req?: StartImportRequest): Promise<{ message: string }>
  cancelImport(): Promise<{ message: string }>
}
