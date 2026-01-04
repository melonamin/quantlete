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
  // Weather
  ActivityWeather,
  // Stats
  PowerStatsResponse,
  HrZonesResponse,
  TrainingLoadResponse,
  PowerZonesResponse,
  HrZoneDefinition,
  DistributionSlice,
  // Gear
  Gear,
  GearFilters,
  GearListResponse,
  CustomGearCreateRequest,
  GearMonthlyUsage,
  // Segments
  SegmentCountryStat,
  SegmentDetailResponse,
  SegmentEffortsFilters,
  SegmentEffortsResponse,
  SegmentsFilters,
  SegmentsResponse,
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
  ExportStats,
  // Challenges
  ChallengesFilters,
  ChallengesResponse,
  // Goals
  TrainingGoalsConfig,
  TrainingGoalsResponse,
  // Maintenance
  ComponentWithRules,
  ComponentsFilters,
  ComponentsResponse,
  DueComponent,
  CreateComponentRequest,
  UpdateComponentRequest,
  LogMaintenanceRequest,
  // Settings
  AppSettings,
  // Import
  ImportProgress,
  StartImportRequest,
  // Setup
  CredentialsStatus,
  UpdateCredentialsRequest,
} from './types'

// Import SyncRun from API types
import type { SyncRun, SyncWatermark } from '@/lib/api/import'

// Import event types
import type { DataEventListener } from './events'

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
  getActivityWeather(id: number): Promise<ActivityWeather | null>

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
  getGear(filters?: GearFilters): Promise<GearListResponse>
  getGearDetail(id: string): Promise<Gear>
  getCustomGear(filters?: GearFilters): Promise<GearListResponse>
  createCustomGear(req: CustomGearCreateRequest): Promise<Gear>
  updateCustomGear(id: string, patch: Partial<CustomGearCreateRequest>): Promise<Gear>
  deleteCustomGear(id: string, force?: boolean): Promise<{ deleted: boolean }>
  getGearMonthlyUsage(includeRetired?: boolean): Promise<GearMonthlyUsage[]>

  // ============================================================================
  // Segments
  // ============================================================================
  getSegments(filters?: SegmentsFilters): Promise<SegmentsResponse>
  getSegmentCountries(): Promise<SegmentCountryStat[]>
  getSegmentDetail(id: number): Promise<SegmentDetailResponse>
  getSegmentEfforts(id: number, filters?: SegmentEffortsFilters): Promise<SegmentEffortsResponse>

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
  getChallenges(filters?: ChallengesFilters): Promise<ChallengesResponse>
  importChallenges(file: File): Promise<{ imported: number }>
  importChallengesFromProfile(athleteId?: string): Promise<{ imported: number }>

  // ============================================================================
  // Goals
  // ============================================================================
  getTrainingGoals(): Promise<TrainingGoalsResponse>
  updateTrainingGoals(config: TrainingGoalsConfig): Promise<TrainingGoalsConfig>

  // ============================================================================
  // Maintenance
  // ============================================================================
  getMaintenanceDue(): Promise<DueComponent[]>
  getGearComponents(gearId: string, filters?: ComponentsFilters): Promise<ComponentsResponse>
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
  // Import
  // ============================================================================
  getImportProgress(): Promise<ImportProgress>
  startImport(req?: StartImportRequest): Promise<{ message: string }>
  cancelImport(): Promise<{ message: string }>
  pauseImport(): Promise<{ message: string }>
  resumeImport(): Promise<{ message: string }>
  hasResumableImport(): boolean
  getSyncHistory(limit?: number): Promise<SyncRun[]>
  getLatestSync(): Promise<SyncRun | null>
  getSyncWatermark(): Promise<SyncWatermark | null>

  // ============================================================================
  // Export
  // ============================================================================
  getExportStats(): Promise<ExportStats>

  // ============================================================================
  // Setup (Strava Credentials)
  // ============================================================================
  getCredentialsStatus(): Promise<CredentialsStatus>
  updateCredentials(req: UpdateCredentialsRequest): Promise<CredentialsStatus>

  // ============================================================================
  // Events (Reactive Updates)
  // ============================================================================
  /**
   * Subscribe to data events for reactive UI updates.
   * @param listener Callback function to receive events
   * @returns Unsubscribe function
   */
  subscribeToEvents(listener: DataEventListener): () => void
}
