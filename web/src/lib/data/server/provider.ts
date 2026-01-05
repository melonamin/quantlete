/**
 * ServerProvider - DataProvider implementation for server mode.
 *
 * Delegates all operations to the Go backend via HTTP REST API.
 * This is a thin wrapper around the existing API client.
 */

import { get, post, put, del } from '@/lib/api/client'
import type { DataProvider } from '../provider'
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
  // Wrapped
  WrappedReport,
  // Photos
  PhotosListResponse,
  PhotosFilters,
  ActivityPhoto,
  // Challenges
  ChallengesFilters,
  ChallengesResponse,
  // Goals
  TrainingGoalsConfig,
  TrainingGoalsResponse,
  // Export
  ExportStats,
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
} from '../types'
import type { SyncRun, SyncWatermark } from '@/lib/api/import'
import type { DataEventListener } from '../events'
import { ImportSSEClient } from './sse-client'

export class ServerProvider implements DataProvider {
  private sseClient = new ImportSSEClient()
  // ============================================================================
  // Auth
  // ============================================================================
  async getAuthStatus(): Promise<AuthStatus> {
    return get<AuthStatus>('/auth/status')
  }

  async refreshToken(): Promise<{ success: boolean }> {
    return post<{ success: boolean }>('/auth/refresh')
  }

  // ============================================================================
  // Activities
  // ============================================================================
  async getActivities(filters: ActivityFilters): Promise<ActivitiesResponse> {
    return get<ActivitiesResponse>('/activities', {
      sport_type: filters.sport_type,
      after: filters.after,
      before: filters.before,
      gear_id: filters.gear_id,
      search: filters.search,
      commute: filters.commute,
      trainer: filters.trainer,
      page: filters.page,
      per_page: filters.per_page,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })
  }

  async getActivity(id: number): Promise<Activity> {
    return get<Activity>(`/activities/${id}`)
  }

  async getActivityStreams(id: number): Promise<ActivityStream[]> {
    return get<ActivityStream[]>(`/activities/${id}/streams`)
  }

  async getActivityWeather(id: number): Promise<ActivityWeather | null> {
    try {
      return await get<ActivityWeather>(`/activities/${id}/weather`)
    } catch (error) {
      // Weather data not available is not a user-facing error, but log for debugging
      console.debug('Weather data not available for activity', id, error)
      return null
    }
  }

  // ============================================================================
  // Dashboard
  // ============================================================================
  async getDashboard(): Promise<DashboardData> {
    return get<DashboardData>('/dashboard')
  }

  async getDashboardStats(): Promise<DashboardStats> {
    return get<DashboardStats>('/dashboard/stats')
  }

  async getWeeklyStats(): Promise<WeeklyStat[]> {
    return get<WeeklyStat[]>('/dashboard/weekly')
  }

  async getRecentActivities(limit: number): Promise<RecentActivity[]> {
    return get<RecentActivity[]>(`/dashboard/recent?limit=${limit}`)
  }

  async getSportTypeStats(): Promise<SportTypeStat[]> {
    return get<SportTypeStat[]>('/dashboard/sports')
  }

  async getMonthlyStats(year?: number): Promise<MonthlyStat[]> {
    const url = year ? `/dashboard/monthly?year=${year}` : '/dashboard/monthly'
    return get<MonthlyStat[]>(url)
  }

  async getYearlyStats(): Promise<YearlyStat[]> {
    return get<YearlyStat[]>('/dashboard/yearly')
  }

  async getCalendarData(year: number): Promise<CalendarDay[]> {
    return get<CalendarDay[]>(`/dashboard/calendar?year=${year}`)
  }

  async getCalendarActivities(year: number, month: number): Promise<CalendarActivity[]> {
    return get<CalendarActivity[]>(`/dashboard/calendar/activities?year=${year}&month=${month}`)
  }

  async getCalendarSummary(year: number, month: number): Promise<CalendarMonthSummary> {
    return get<CalendarMonthSummary>(`/dashboard/calendar/summary?year=${year}&month=${month}`)
  }

  async getDashboardConfig(): Promise<DashboardConfig> {
    return get<DashboardConfig>('/dashboard/config')
  }

  async updateDashboardConfig(config: DashboardConfig): Promise<DashboardConfig> {
    return put<DashboardConfig>('/dashboard/config', config)
  }

  // ============================================================================
  // Heatmap & Eddington
  // ============================================================================
  async getHeatmapData(filters: HeatmapFilters): Promise<HeatmapResponse> {
    const params = new URLSearchParams()
    if (filters.sport_type) params.set('sport_type', filters.sport_type)
    if (filters.after) params.set('after', filters.after)
    if (filters.before) params.set('before', filters.before)
    if (filters.commute !== undefined) params.set('commute', String(filters.commute))
    if (filters.workout_type !== undefined) params.set('workout_type', String(filters.workout_type))
    const qs = params.toString()
    return get<HeatmapResponse>(`/stats/heatmap${qs ? `?${qs}` : ''}`)
  }

  async getEddingtonData(sportType?: string): Promise<EddingtonResult> {
    const params = sportType ? `?sport_type=${sportType}` : ''
    return get<EddingtonResult>(`/stats/eddington${params}`)
  }

  async getEddingtonHistory(sportType?: string): Promise<EddingtonHistoryPoint[]> {
    const params = sportType ? `?sport_type=${sportType}` : ''
    return get<EddingtonHistoryPoint[]>(`/stats/eddington/history${params}`)
  }

  // ============================================================================
  // Stats & Training
  // ============================================================================
  async getPowerStats(
    filters: { after?: string; before?: string; sport_type?: string } = {}
  ): Promise<PowerStatsResponse> {
    const params = new URLSearchParams()
    if (filters.after) params.set('after', filters.after)
    if (filters.before) params.set('before', filters.before)
    if (filters.sport_type) params.set('sport_type', filters.sport_type)
    const qs = params.toString()
    return get<PowerStatsResponse>(`/stats/power${qs ? `?${qs}` : ''}`)
  }

  async getPowerZones(): Promise<PowerZonesResponse> {
    return get<PowerZonesResponse>('/stats/power-zones')
  }

  async getHrZones(
    filters: { after?: string; before?: string; sport_type?: string } = {}
  ): Promise<HrZonesResponse> {
    const params = new URLSearchParams()
    if (filters.after) params.set('after', filters.after)
    if (filters.before) params.set('before', filters.before)
    if (filters.sport_type) params.set('sport_type', filters.sport_type)
    const qs = params.toString()
    return get<HrZonesResponse>(`/stats/hr-zones${qs ? `?${qs}` : ''}`)
  }

  async getTrainingLoad(
    filters: { after?: string; before?: string } = {}
  ): Promise<TrainingLoadResponse> {
    const params = new URLSearchParams()
    if (filters.after) params.set('after', filters.after)
    if (filters.before) params.set('before', filters.before)
    const qs = params.toString()
    return get<TrainingLoadResponse>(`/stats/training-load${qs ? `?${qs}` : ''}`)
  }

  async getHrZoneDefinitions(): Promise<HrZoneDefinition[]> {
    return get<HrZoneDefinition[]>('/zones/hr')
  }

  async upsertHrZoneDefinition(def: HrZoneDefinition): Promise<{ status: string }> {
    return put<{ status: string }>('/zones/hr', def)
  }

  async deleteHrZoneDefinition(
    sportType: string,
    effectiveFrom: string
  ): Promise<{ status: string }> {
    return del<{ status: string }>(
      `/zones/hr?sport_type=${encodeURIComponent(sportType)}&effective_from=${encodeURIComponent(effectiveFrom)}`
    )
  }

  async getDaytimeDistribution(): Promise<DistributionSlice[]> {
    return get<DistributionSlice[]>('/stats/daytime')
  }

  async getWeekdayDistribution(): Promise<DistributionSlice[]> {
    return get<DistributionSlice[]>('/stats/weekday')
  }

  // ============================================================================
  // Best Efforts
  // ============================================================================
  async getBestEffortPRs(sportType?: string): Promise<BestEffortPR[]> {
    const params = sportType ? `?sport_type=${encodeURIComponent(sportType)}` : ''
    return get<BestEffortPR[]>(`/stats/best-efforts${params}`)
  }

  async getBestEffortsForDistance(
    distanceType: string,
    sportType?: string
  ): Promise<BestEffortItem[]> {
    const params = sportType ? `?sport_type=${encodeURIComponent(sportType)}` : ''
    return get<BestEffortItem[]>(`/stats/best-efforts/${encodeURIComponent(distanceType)}${params}`)
  }

  // ============================================================================
  // Wrapped
  // ============================================================================
  async getWrappedYears(): Promise<number[]> {
    return get<number[]>('/stats/wrapped/years')
  }

  async getWrapped(year: number): Promise<WrappedReport> {
    return get<WrappedReport>(`/stats/wrapped?year=${encodeURIComponent(String(year))}`)
  }

  // ============================================================================
  // Gear
  // ============================================================================
  async getGear(filters: GearFilters = {}): Promise<GearListResponse> {
    return get<GearListResponse>('/gear', {
      include_retired: filters.include_retired,
      page: filters.page,
      per_page: filters.per_page,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })
  }

  async getGearDetail(id: string): Promise<Gear> {
    return get<Gear>(`/gear/${id}`)
  }

  async getCustomGear(filters: GearFilters = {}): Promise<GearListResponse> {
    return get<GearListResponse>('/gear/custom', {
      include_retired: filters.include_retired,
      page: filters.page,
      per_page: filters.per_page,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })
  }

  async createCustomGear(req: CustomGearCreateRequest): Promise<Gear> {
    return post<Gear>('/gear/custom', req)
  }

  async updateCustomGear(id: string, patch: Partial<CustomGearCreateRequest>): Promise<Gear> {
    return put<Gear>(`/gear/custom/${id}`, patch)
  }

  async deleteCustomGear(id: string, force?: boolean): Promise<{ deleted: boolean }> {
    return del<{ deleted: boolean }>(`/gear/custom/${id}${force ? '?force=true' : ''}`)
  }

  async getGearMonthlyUsage(includeRetired = true): Promise<GearMonthlyUsage[]> {
    return get<GearMonthlyUsage[]>(`/gear/stats/monthly?include_retired=${includeRetired}`)
  }

  // ============================================================================
  // Segments
  // ============================================================================
  async getSegments(filters: SegmentsFilters = {}): Promise<SegmentsResponse> {
    return get<SegmentsResponse>('/segments', {
      activity_type: filters.activity_type,
      country: filters.country,
      starred: filters.starred,
      kom_only: filters.kom_only,
      search: filters.search,
      page: filters.page,
      per_page: filters.per_page,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })
  }

  async getSegmentCountries(): Promise<SegmentCountryStat[]> {
    return get<SegmentCountryStat[]>('/segments/countries')
  }

  async getSegmentDetail(id: number): Promise<SegmentDetailResponse> {
    return get<SegmentDetailResponse>(`/segments/${id}`)
  }

  async getSegmentEfforts(
    id: number,
    filters: SegmentEffortsFilters = {}
  ): Promise<SegmentEffortsResponse> {
    return get<SegmentEffortsResponse>(`/segments/${id}/efforts`, {
      page: filters.page,
      per_page: filters.per_page,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })
  }

  // ============================================================================
  // Athlete (FTP/Weight)
  // ============================================================================
  async getFtpHistory(): Promise<FTPHistoryResponse> {
    return get<FTPHistoryResponse>('/athlete/ftp')
  }

  async updateFtpHistory(body: FTPHistoryResponse): Promise<FTPHistoryResponse> {
    return put<FTPHistoryResponse>('/athlete/ftp', body)
  }

  async getWeightHistory(): Promise<WeightHistoryResponse> {
    return get<WeightHistoryResponse>('/athlete/weight')
  }

  async updateWeightHistory(body: WeightHistoryResponse): Promise<WeightHistoryResponse> {
    return put<WeightHistoryResponse>('/athlete/weight', body)
  }

  // ============================================================================
  // Photos
  // ============================================================================
  async getPhotos(filters: PhotosFilters = {}): Promise<PhotosListResponse> {
    return get<PhotosListResponse>('/photos', {
      sport_type: filters.sport_type,
      country: filters.country,
      page: filters.page,
      per_page: filters.per_page,
    })
  }

  async getActivityPhotos(activityId: number): Promise<ActivityPhoto[]> {
    return get<ActivityPhoto[]>(`/activities/${activityId}/photos`)
  }

  // ============================================================================
  // Challenges
  // ============================================================================
  async getChallenges(filters: ChallengesFilters = {}): Promise<ChallengesResponse> {
    return get<ChallengesResponse>('/challenges', {
      month: filters.month,
      page: filters.page,
      per_page: filters.per_page,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })
  }

  async importChallenges(file: File): Promise<{ imported: number }> {
    const form = new FormData()
    form.append('file', file)
    const response = await fetch('/api/v1/challenges/import', {
      method: 'POST',
      credentials: 'include',
      body: form,
      headers: { Accept: 'application/json' },
    })
    const body = await response.json().catch(() => null)
    if (!response.ok) {
      throw new Error(body?.error || `Request failed: ${response.status}`)
    }
    // Unwrap envelope response
    if (body && typeof body === 'object' && 'ok' in body) {
      if (!body.ok) {
        throw new Error(body.error || 'Unknown error')
      }
      return body.data ?? body
    }
    return body
  }

  async importChallengesFromProfile(athleteId?: string): Promise<{ imported: number }> {
    const params = new URLSearchParams()
    if (athleteId) {
      params.set('athlete_id', athleteId)
    }
    const url = `/api/v1/challenges/import-profile${params.toString() ? `?${params}` : ''}`
    const response = await fetch(url, {
      method: 'POST',
      credentials: 'include',
      headers: { Accept: 'application/json' },
    })
    const body = await response.json().catch(() => null)
    if (!response.ok) {
      throw new Error(body?.error || `Request failed: ${response.status}`)
    }
    // Unwrap envelope response
    if (body && typeof body === 'object' && 'ok' in body) {
      if (!body.ok) {
        throw new Error(body.error || 'Unknown error')
      }
      return body.data ?? body
    }
    return body
  }

  // ============================================================================
  // Goals
  // ============================================================================
  async getTrainingGoals(): Promise<TrainingGoalsResponse> {
    return get<TrainingGoalsResponse>('/goals')
  }

  async updateTrainingGoals(config: TrainingGoalsConfig): Promise<TrainingGoalsConfig> {
    return put<TrainingGoalsConfig>('/goals', config)
  }

  // ============================================================================
  // Maintenance
  // ============================================================================
  async getMaintenanceDue(): Promise<DueComponent[]> {
    return get<DueComponent[]>('/maintenance/due')
  }

  async getGearComponents(
    gearId: string,
    filters: ComponentsFilters = {}
  ): Promise<ComponentsResponse> {
    return get<ComponentsResponse>(`/gear/${gearId}/components`, {
      page: filters.page,
      per_page: filters.per_page,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })
  }

  async createComponent(gearId: string, req: CreateComponentRequest): Promise<ComponentWithRules> {
    return post<ComponentWithRules>(`/gear/${gearId}/components`, req)
  }

  async updateComponent(id: number, req: UpdateComponentRequest): Promise<ComponentWithRules> {
    return put<ComponentWithRules>(`/components/${id}`, req)
  }

  async deleteComponent(id: number): Promise<{ deleted: boolean }> {
    return del<{ deleted: boolean }>(`/components/${id}`)
  }

  async logMaintenance(
    componentId: number,
    req?: LogMaintenanceRequest
  ): Promise<{ logged: boolean }> {
    return post<{ logged: boolean }>(`/components/${componentId}/maintenance`, req ?? {})
  }

  // ============================================================================
  // Settings
  // ============================================================================
  async getAppSettings(): Promise<AppSettings> {
    return get<AppSettings>('/settings')
  }

  async updateAppSettings(settings: AppSettings): Promise<AppSettings> {
    return put<AppSettings>('/settings', settings)
  }

  // ============================================================================
  // Import
  // ============================================================================
  async getImportProgress(): Promise<ImportProgress> {
    return get<ImportProgress>('/import/progress')
  }

  async startImport(req: StartImportRequest = {}): Promise<{ message: string }> {
    return post<{ message: string }>('/import/start', req)
  }

  async cancelImport(): Promise<{ message: string }> {
    return post<{ message: string }>('/import/cancel')
  }

  async pauseImport(): Promise<{ message: string }> {
    return post<{ message: string }>('/import/pause')
  }

  async resumeImport(): Promise<{ message: string }> {
    return post<{ message: string }>('/import/resume')
  }

  hasResumableImport(): boolean {
    // Server mode doesn't track resumable state client-side
    return false
  }

  async getSyncHistory(limit = 10): Promise<SyncRun[]> {
    return get<SyncRun[]>('/import/history', { limit })
  }

  async getLatestSync(): Promise<SyncRun | null> {
    const history = await this.getSyncHistory(1)
    return history.length > 0 ? history[0] : null
  }

  async getSyncWatermark(): Promise<SyncWatermark | null> {
    return get<SyncWatermark | null>('/import/watermark')
  }

  async resetSyncWatermark(): Promise<{ message: string }> {
    return del<{ message: string }>('/import/watermark')
  }

  async getExportStats(): Promise<ExportStats> {
    return get<ExportStats>('/export/stats')
  }

  // ============================================================================
  // Setup (Strava Credentials)
  // ============================================================================
  async getCredentialsStatus(): Promise<CredentialsStatus> {
    return get<CredentialsStatus>('/setup/credentials')
  }

  async updateCredentials(req: UpdateCredentialsRequest): Promise<CredentialsStatus> {
    return put<CredentialsStatus>('/setup/credentials', req)
  }

  // ============================================================================
  // Events (Reactive Updates)
  // ============================================================================
  subscribeToEvents(listener: DataEventListener): () => void {
    return this.sseClient.subscribe(listener)
  }
}
