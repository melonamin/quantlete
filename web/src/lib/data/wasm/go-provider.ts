/**
 * GoWasmProvider - DataProvider implementation using Go WASM storage layer.
 *
 * This provider delegates all storage operations to Go code compiled to WASM,
 * eliminating the need for TypeScript query implementations.
 */

import type { DataProvider } from '../provider'
import type {
  Activity,
  ActivityFilters,
  ActivitiesResponse,
  AuthStatus,
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
  WidgetWidth,
  WidgetHeight,
  ActivityStream,
  ActivityWeather,
  PowerStatsResponse,
  HrZonesResponse,
  TrainingLoadResponse,
  PowerZonesResponse,
  HrZoneDefinition,
  DistributionSlice,
  Gear,
  GearFilters,
  GearListResponse,
  CustomGearCreateRequest,
  GearMonthlyUsage,
  SegmentCountryStat,
  SegmentDetailResponse,
  SegmentEffortsFilters,
  SegmentEffortsResponse,
  SegmentsFilters,
  SegmentsResponse,
  FTPHistoryResponse,
  WeightHistoryResponse,
  BestEffortPR,
  BestEffortItem,
  RewindReport,
  PhotosListResponse,
  PhotosFilters,
  ActivityPhoto,
  ChallengesFilters,
  ChallengesResponse,
  TrainingGoalsConfig,
  TrainingGoalsResponse,
  ComponentWithRules,
  ComponentsFilters,
  ComponentsResponse,
  DueComponent,
  CreateComponentRequest,
  UpdateComponentRequest,
  LogMaintenanceRequest,
  AppSettings,
  ImportProgress,
  StartImportRequest,
  ExportStats,
  CredentialsStatus,
  UpdateCredentialsRequest,
} from '../types'

import type { DataEventListener } from '../events'
import {
  createSyncProgressEvent,
  createSyncCompleteEvent,
  createDataChangedEvent,
} from '../events'
import { UnsupportedFeatureError } from '../errors'
import type { SyncRun, SyncWatermark } from '@/lib/api/import'

import * as goStorage from '@/lib/wasm/go-storage'
import { loadAuth, isAuthenticated, getAthlete, getAuthUrl, exchangeCode } from '@/lib/wasm/strava'
import { getCredentials, saveCredentials, hasCredentials } from '@/lib/wasm/strava/credentials'

// Global callbacks for Go importer events - set up during subscribeToEvents
let importProgressCallback: ((progressJson: string) => void) | null = null
let importCompleteCallback: ((resultJson: string) => void) | null = null

// Register global callbacks that Go WASM can call
if (typeof window !== 'undefined') {
  ;(window as Window & { onImportProgress?: (json: string) => void }).onImportProgress = (
    json: string
  ) => {
    if (importProgressCallback) {
      importProgressCallback(json)
    }
  }
  ;(window as Window & { onImportComplete?: (json: string) => void }).onImportComplete = (
    json: string
  ) => {
    if (importCompleteCallback) {
      importCompleteCallback(json)
    }
  }
}

export interface GoWasmProviderOptions {
  /** Demo mode: load bundled database, skip OAuth */
  demoMode?: boolean
  /** URL to bundled demo database (required if demoMode is true) */
  demoDatabaseUrl?: string
}

export class GoWasmProvider implements DataProvider {
  private athleteId: number | null = null
  private options: GoWasmProviderOptions
  private demoAthlete: { id: number; firstname: string; lastname: string } | null = null

  constructor(options: GoWasmProviderOptions = {}) {
    this.options = options
  }

  async initialize(): Promise<void> {
    // Initialize Go WASM storage with options
    await goStorage.initGoStorage({
      demoMode: this.options.demoMode,
      demoDatabaseUrl: this.options.demoDatabaseUrl,
    })

    if (this.options.demoMode) {
      // Demo mode: load athlete via repository (no auth tokens needed)
      const athlete = goStorage.getFirstAthlete()
      if (!athlete?.id) {
        throw new Error('Demo database is empty or corrupted: no athlete found')
      }
      this.athleteId = athlete.id
      this.demoAthlete = {
        id: athlete.id,
        firstname: athlete.firstname || 'Demo',
        lastname: athlete.lastname || 'User',
      }
      goStorage.setAthleteId(this.athleteId)
    } else {
      // Normal mode: try to load authentication
      const authLoaded = await loadAuth()
      if (authLoaded) {
        const athlete = getAthlete()
        if (athlete?.id) {
          this.athleteId = athlete.id
          goStorage.setAthleteId(athlete.id)
        }
      }
    }
  }

  private assertInitialized(): void {
    if (!goStorage.isInitialized()) {
      throw new Error('GoWasmProvider not initialized. Call initialize() first.')
    }
  }

  private getAthleteId(): number {
    if (this.athleteId === null) {
      throw new Error('No athlete found. Please import activities first.')
    }
    return this.athleteId
  }

  // ============================================================================
  // Auth
  // ============================================================================
  async getAuthStatus(): Promise<AuthStatus> {
    this.assertInitialized()

    // Demo mode: always authenticated with demo athlete
    if (this.options.demoMode && this.demoAthlete) {
      return {
        authenticated: true,
        demo_mode: true,
        athlete: {
          id: this.demoAthlete.id,
          username: 'demo_user',
          firstname: this.demoAthlete.firstname,
          lastname: this.demoAthlete.lastname,
          profile: '',
        },
      }
    }

    if (!isAuthenticated()) {
      return { authenticated: false, athlete: undefined }
    }

    const athlete = getAthlete()
    if (!athlete) {
      return { authenticated: false, athlete: undefined }
    }

    this.athleteId = athlete.id
    goStorage.setAthleteId(athlete.id)

    return {
      authenticated: true,
      athlete: {
        id: athlete.id,
        username: athlete.username,
        firstname: athlete.firstname,
        lastname: athlete.lastname,
        profile: athlete.profile,
      },
    }
  }

  async refreshToken(): Promise<{ success: boolean }> {
    return { success: true }
  }

  getStravaAuthUrl(redirectUri: string): string {
    return getAuthUrl(redirectUri)
  }

  async handleOAuthCallback(code: string, redirectUri: string): Promise<void> {
    await exchangeCode(code, redirectUri)
    const athlete = getAthlete()
    if (athlete?.id) {
      this.athleteId = athlete.id
      goStorage.setAthleteId(athlete.id)
    }
  }

  // ============================================================================
  // Activities
  // ============================================================================
  async getActivities(filters: ActivityFilters): Promise<ActivitiesResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getActivities({
      page: filters.page,
      per_page: filters.per_page,
      sport_type: filters.sport_type,
      after: filters.after,
      before: filters.before,
      gear_id: filters.gear_id,
      commute: filters.commute,
      trainer: filters.trainer,
      search: filters.search,
      order_by: filters.order_by,
      order_dir: filters.order_dir,
    })

    return {
      data: result.data as Activity[],
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
    }
  }

  async getActivity(id: number): Promise<Activity> {
    this.assertInitialized()
    return goStorage.getActivity(id) as Activity
  }

  async getActivityStreams(id: number): Promise<ActivityStream[]> {
    this.assertInitialized()
    const streams = goStorage.getActivityStreams(id)
    return streams.map((s) => ({
      activity_id: id,
      stream_type: s.stream_type,
      original_size: s.original_size,
      series_type: s.series_type,
      resolution: s.resolution,
      data: s.data,
    })) as ActivityStream[]
  }

  async getActivityWeather(_id: number): Promise<ActivityWeather | null> {
    // Weather lookup is server-only: requires external API calls (Open-Meteo)
    // that cannot work in browser WASM mode due to CORS and API key exposure.
    return null
  }

  // ============================================================================
  // Dashboard
  // ============================================================================
  async getDashboard(): Promise<DashboardData> {
    this.assertInitialized()
    this.getAthleteId()

    const [dashStats, weeklyStats, recentActivities, sportStats] = await Promise.all([
      this.getDashboardStats(),
      this.getWeeklyStats(),
      this.getRecentActivities(5),
      this.getSportTypeStats(),
    ])

    return {
      stats: dashStats,
      weekly_stats: weeklyStats,
      recent_activities: recentActivities,
      sport_type_stats: sportStats,
    }
  }

  async getDashboardStats(): Promise<DashboardStats> {
    this.assertInitialized()
    this.getAthleteId()
    const stats = goStorage.getDashboardStats()
    return {
      total_activities: stats.total_activities,
      total_distance: stats.total_distance,
      total_moving_time: stats.total_moving_time,
      total_elevation_gain: stats.total_elevation_gain,
      total_calories: stats.total_calories ?? 0,
      year_activities: stats.year_activities ?? 0,
      year_distance: stats.year_distance ?? 0,
      year_moving_time: stats.year_moving_time ?? 0,
      year_elevation_gain: stats.year_elevation_gain ?? 0,
      month_activities: stats.month_activities ?? 0,
      month_distance: stats.month_distance ?? 0,
      month_moving_time: stats.month_moving_time ?? 0,
      month_elevation_gain: stats.month_elevation_gain ?? 0,
    }
  }

  async getWeeklyStats(): Promise<WeeklyStat[]> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getWeeklyStats()
  }

  async getRecentActivities(limit: number): Promise<RecentActivity[]> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getRecentActivities(limit)
  }

  async getSportTypeStats(): Promise<SportTypeStat[]> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getSportTypeStats()
  }

  async getMonthlyStats(year?: number): Promise<MonthlyStat[]> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getMonthlyStats(year)
  }

  async getYearlyStats(): Promise<YearlyStat[]> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getYearlyStats()
  }

  async getCalendarData(year: number): Promise<CalendarDay[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getCalendarData(year)
    return result.map((d) => ({
      date: d.date,
      activity_count: d.activity_count,
      total_distance: d.total_distance,
      total_time: d.total_time,
      total_calories: d.total_calories,
    }))
  }

  async getCalendarActivities(year: number, month: number): Promise<CalendarActivity[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getCalendarActivities(year, month)
    return result.map((a) => ({
      id: a.id,
      name: a.name,
      sport_type: a.sport_type,
      start_date: a.start_date,
      distance: a.distance,
      moving_time: a.moving_time,
      total_elevation_gain: a.total_elevation_gain,
    }))
  }

  async getCalendarSummary(year: number, month: number): Promise<CalendarMonthSummary> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getCalendarSummary(year, month)
    return {
      year: result.year,
      month: result.month,
      activity_count: result.activity_count,
      total_distance: result.total_distance,
      total_elevation_gain: result.total_elevation_gain,
      total_moving_time: result.total_moving_time,
      total_calories: result.total_calories,
      workout_count: result.workout_count,
      challenges_completed: result.challenges_completed,
    }
  }

  async getDashboardConfig(): Promise<DashboardConfig> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getDashboardConfig()
    return {
      version: result.version,
      widgets: result.widgets.map((w) => ({
        id: w.id,
        width: w.width as WidgetWidth,
        height: w.height as WidgetHeight | undefined,
        hidden: w.hidden ?? false,
        settings: w.settings,
      })),
    }
  }

  async updateDashboardConfig(config: DashboardConfig): Promise<DashboardConfig> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.updateDashboardConfig(config)
    return {
      version: result.version,
      widgets: result.widgets.map((w) => ({
        id: w.id,
        width: w.width as WidgetWidth,
        height: w.height as WidgetHeight | undefined,
        hidden: w.hidden ?? false,
        settings: w.settings,
      })),
    }
  }

  // ============================================================================
  // Heatmap & Eddington
  // ============================================================================
  async getHeatmapData(filters: HeatmapFilters): Promise<HeatmapResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getHeatmapData({
      sport_type: filters.sport_type,
      commute: filters.commute,
      workout_type: filters.workout_type,
      limit: filters.limit,
      offset: filters.offset,
    })

    // Handle both old format (array) and new format (object with activities and countries)
    const activities = Array.isArray(result) ? result : (result.activities ?? [])
    const countries = Array.isArray(result) ? [] : (result.countries ?? [])

    return {
      activities: activities.map((a) => ({
        id: a.id,
        name: a.name,
        sport_type: a.sport_type,
        start_date: a.start_date,
        distance: a.distance,
        summary_polyline: a.summary_polyline,
        start_lat: a.start_lat,
        start_lng: a.start_lng,
      })),
      total: activities.length,
      countries: countries.map((c: { country: string; iso2?: string; count: number }) => ({
        country: c.country,
        iso2: c.iso2,
        count: c.count,
      })),
    }
  }

  async getEddingtonData(sportType?: string): Promise<EddingtonResult> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getEddingtonData({ sport_types: sportType ? [sportType] : undefined })
    return {
      number: result.number,
      distribution: result.distribution.map((d) => ({
        date: d.date,
        distance: d.distance,
      })),
      next_steps: result.next_steps.map((s) => ({
        target: s.target,
        rides_needed: s.rides_needed,
      })),
    }
  }

  async getEddingtonHistory(sportType?: string): Promise<EddingtonHistoryPoint[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getEddingtonData({ sport_types: sportType ? [sportType] : undefined })
    return result.history.map((h) => ({
      date: h.date,
      number: h.number,
    }))
  }

  // ============================================================================
  // Stats & Training
  // ============================================================================
  async getPowerStats(filters?: {
    after?: string
    before?: string
    sport_type?: string
  }): Promise<PowerStatsResponse> {
    this.assertInitialized()
    this.getAthleteId()

    // Default history duration is 300s (5 min) - matches Go backend default
    const historyDuration = 300

    const result = goStorage.getPowerStats({
      after: filters?.after,
      before: filters?.before,
      sport_types: filters?.sport_type ? [filters.sport_type] : undefined,
      history_duration: historyDuration,
    })

    // Build durations from best array
    const durationsSet = new Set(result.best.map((b) => b.duration_s))
    const durations = Array.from(durationsSet).sort((a, b) => a - b)

    // History is returned for a single duration (historyDuration)
    // Key all history points under that duration
    const historyByDuration: Record<string, { date: string; watts: number }[]> = {}
    if (result.history.length > 0) {
      historyByDuration[String(historyDuration)] = result.history.map((point) => ({
        date: point.date,
        watts: point.watts,
      }))
    }

    return {
      durations_s: durations,
      best: result.best.map((b) => ({
        duration_s: b.duration_s,
        watts: b.watts,
        activity_id: b.activity_id,
        start_date: b.start_date,
      })),
      history: historyByDuration,
    }
  }

  async getPowerZones(): Promise<PowerZonesResponse> {
    // Power zone calculation is partially supported: we can compute zone bounds from FTP,
    // but seconds_by_zone requires processing all activity power streams which is
    // computationally expensive and server-only. Zone bounds are calculated here.
    this.assertInitialized()
    this.getAthleteId()

    const ftpHistory = goStorage.getFtpHistory()
    const latestFtp = ftpHistory.length > 0 ? ftpHistory[ftpHistory.length - 1].value : undefined

    // Standard 7-zone power model based on FTP
    const bounds = latestFtp
      ? [
          0,
          Math.round(latestFtp * 0.55),
          Math.round(latestFtp * 0.75),
          Math.round(latestFtp * 0.9),
          Math.round(latestFtp * 1.05),
          Math.round(latestFtp * 1.2),
          Math.round(latestFtp * 1.5),
        ]
      : [0, 100, 150, 200, 250, 300, 400]

    return {
      ftp_watts: latestFtp,
      seconds_by_zone: [], // Server-only: requires stream processing
      total_seconds: 0,
      bounds,
    }
  }

  async getHrZones(_filters?: {
    after?: string
    before?: string
    sport_type?: string
  }): Promise<HrZonesResponse> {
    // HR zone calculation is partially supported: zone bounds are returned,
    // but seconds_by_zone requires processing all activity HR streams which is
    // computationally expensive and server-only.
    this.assertInitialized()
    this.getAthleteId()

    return {
      method: 'percentage',
      zones: { bounds: [0, 60, 70, 80, 90, 100] },
      seconds_by_zone: [], // Server-only: requires stream processing
      total_seconds: 0,
    }
  }

  async getTrainingLoad(filters?: {
    after?: string
    before?: string
  }): Promise<TrainingLoadResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getTrainingLoad({
      after: filters?.after,
      before: filters?.before,
    })

    return {
      series: result.series.map((s) => ({
        day: s.day,
        tss: s.tss,
        ctl: s.ctl,
        atl: s.atl,
        tsb: s.tsb,
      })),
      summary: result.summary
        ? {
            day: result.summary.day,
            tss: result.summary.tss,
            ctl: result.summary.ctl,
            atl: result.summary.atl,
            tsb: result.summary.tsb,
          }
        : undefined,
    }
  }

  async getHrZoneDefinitions(): Promise<HrZoneDefinition[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getHrZoneDefinitions()
    return result.map((d) => ({
      sport_type: d.sport_type,
      effective_from: d.effective_from,
      method: d.method,
      zones: d.zones as { bounds: number[]; hr_max?: number },
    }))
  }

  async upsertHrZoneDefinition(def: HrZoneDefinition): Promise<{ status: string }> {
    this.assertInitialized()
    this.getAthleteId()

    goStorage.upsertHrZoneDefinition({
      sport_type: def.sport_type,
      effective_from: def.effective_from,
      method: def.method,
      zones: def.zones,
    })
    return { status: 'ok' }
  }

  async deleteHrZoneDefinition(
    sportType: string,
    effectiveFrom: string
  ): Promise<{ status: string }> {
    this.assertInitialized()
    this.getAthleteId()

    goStorage.deleteHrZoneDefinition({
      sport_type: sportType,
      effective_from: effectiveFrom,
    })
    return { status: 'ok' }
  }

  async getDaytimeDistribution(): Promise<DistributionSlice[]> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getDaytimeDistribution()
  }

  async getWeekdayDistribution(): Promise<DistributionSlice[]> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getWeekdayDistribution()
  }

  // ============================================================================
  // Best Efforts
  // ============================================================================
  async getBestEffortPRs(sportType?: string): Promise<BestEffortPR[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getBestEffortPRs(sportType)
    return result.map((e) => ({
      distance_type: e.distance_type,
      name: e.name,
      distance_m: e.distance_m,
      elapsed_time_s: e.elapsed_time_s,
      moving_time_s: e.moving_time_s,
      pr_rank: e.pr_rank,
      activity_id: e.activity_id,
      activity_name: e.activity_name,
      sport_type: e.sport_type,
      start_date_local: e.start_date_local,
    }))
  }

  async getBestEffortsForDistance(
    distanceType: string,
    sportType?: string
  ): Promise<BestEffortItem[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getBestEffortsForType(distanceType, sportType)
    return result.map((e) => ({
      distance_type: e.distance_type,
      name: e.name,
      distance_m: e.distance_m,
      elapsed_time_s: e.elapsed_time_s,
      moving_time_s: e.moving_time_s,
      pr_rank: e.pr_rank,
      activity_id: e.activity_id,
      activity_name: e.activity_name,
      sport_type: e.sport_type,
      start_date_local: e.start_date_local,
      start_index: e.start_index,
      end_index: e.end_index,
    }))
  }

  // ============================================================================
  // Rewind
  // ============================================================================
  async getRewindYears(): Promise<number[]> {
    this.assertInitialized()
    this.getAthleteId()

    return goStorage.getRewindYears()
  }

  async getRewind(year: number): Promise<RewindReport> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getRewind(year)
    if (!result) {
      throw new Error(`No rewind data for year ${year}`)
    }

    return {
      year: result.year,
      range_start: result.range_start,
      range_end: result.range_end,
      total_days: result.total_days,
      active_days: result.active_days,
      rest_days: result.rest_days,
      totals: {
        activities: result.totals.activities,
        distance_m: result.totals.distance_m,
        elevation_m: result.totals.elevation_m,
        moving_time_s: result.totals.moving_time_s,
        kudos: result.totals.kudos,
        commute_distance_m: result.totals.commute_dist_m,
        carbon_saved_kg: result.totals.carbon_saved_kg,
      },
      months: result.months?.map((m) => ({
        month: m.month,
        activities: m.activities,
        distance_m: m.distance_m,
        elevation_m: m.elevation_m,
        prs: m.prs,
      })),
      moving_time_by_sport:
        result.moving_time_by_sport?.map((s) => ({
          sport_type: s.sport_type,
          moving_time_s: s.moving_time_s,
        })) || [],
      start_times_by_hour:
        result.start_times_by_hour?.map((h) => ({
          hour: h.hour,
          count: h.count,
        })) || [],
      locations:
        result.locations?.map((l) => ({
          lat: l.lat,
          lng: l.lng,
          count: l.count,
        })) || [],
      streaks: {
        longest_active_days: result.streaks.longest_active_days,
        longest_rest_days: result.streaks.longest_rest_days,
      },
      random_photo: result.random_photo
        ? {
            id: result.random_photo.id,
            activity_id: result.random_photo.activity_id,
            url: result.random_photo.url,
            thumbnail_url: result.random_photo.thumbnail_url,
            caption: result.random_photo.caption,
          }
        : undefined,
      biggest: {
        longest_distance: result.biggest?.longest_distance
          ? {
              activity_id: result.biggest.longest_distance.activity_id,
              name: result.biggest.longest_distance.name,
              sport_type: result.biggest.longest_distance.sport_type,
              start_date_local: result.biggest.longest_distance.start_date_local,
              value: result.biggest.longest_distance.value,
            }
          : undefined,
        most_elevation: result.biggest?.most_elevation
          ? {
              activity_id: result.biggest.most_elevation.activity_id,
              name: result.biggest.most_elevation.name,
              sport_type: result.biggest.most_elevation.sport_type,
              start_date_local: result.biggest.most_elevation.start_date_local,
              value: result.biggest.most_elevation.value,
            }
          : undefined,
        longest_duration: result.biggest?.longest_duration
          ? {
              activity_id: result.biggest.longest_duration.activity_id,
              name: result.biggest.longest_duration.name,
              sport_type: result.biggest.longest_duration.sport_type,
              start_date_local: result.biggest.longest_duration.start_date_local,
              value: result.biggest.longest_duration.value,
            }
          : undefined,
      },
    }
  }

  // ============================================================================
  // Gear
  // ============================================================================
  async getGear(filters?: GearFilters): Promise<GearListResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getGear({
      include_retired: filters?.include_retired,
      page: filters?.page,
      per_page: filters?.per_page,
      order_by: filters?.order_by,
      order_dir: filters?.order_dir,
    })

    return {
      data: result.data.map((g) => ({
        id: g.id,
        name: g.name,
        primary: g.primary,
        retired: g.retired,
        distance: g.distance,
        brand_name: g.brand_name,
        model_name: g.model_name,
        description: g.description,
        source: g.source,
        hashtag: g.hashtag,
        purchase_price: g.purchase_price,
        purchase_currency: g.purchase_currency,
        activity_count: g.activity_count,
      })),
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
    }
  }

  async getGearDetail(id: string): Promise<Gear> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getGearDetail(id)
    return {
      id: result.id,
      name: result.name,
      primary: result.primary,
      retired: result.retired,
      distance: result.distance,
      brand_name: result.brand_name,
      model_name: result.model_name,
      description: result.description,
      source: result.source,
      hashtag: result.hashtag,
      purchase_price: result.purchase_price,
      purchase_currency: result.purchase_currency,
      activity_count: result.activity_count,
    }
  }

  async getCustomGear(filters?: GearFilters): Promise<GearListResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getCustomGear({
      include_retired: filters?.include_retired,
      page: filters?.page,
      per_page: filters?.per_page,
      order_by: filters?.order_by,
      order_dir: filters?.order_dir,
    })

    return {
      data: result.data.map((g) => ({
        id: g.id,
        name: g.name,
        primary: g.primary,
        retired: g.retired,
        distance: g.distance,
        source: g.source,
        hashtag: g.hashtag,
        purchase_price: g.purchase_price,
        purchase_currency: g.purchase_currency,
        activity_count: 0, // Custom gear may not have this
      })),
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
    }
  }

  async createCustomGear(req: CustomGearCreateRequest): Promise<Gear> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.createCustomGear({
      name: req.name,
      hashtag: req.hashtag,
      retired: req.retired,
      purchase_price: req.purchase_price ?? undefined,
      purchase_currency: req.purchase_currency,
    })

    return {
      id: result.id,
      name: result.name,
      primary: false,
      retired: result.retired,
      distance: result.distance,
      source: 'custom',
      hashtag: result.hashtag,
      activity_count: 0,
    }
  }

  async updateCustomGear(id: string, patch: Partial<CustomGearCreateRequest>): Promise<Gear> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.updateCustomGear({
      id,
      name: patch.name,
      hashtag: patch.hashtag,
      retired: patch.retired,
      purchase_price: patch.purchase_price,
      purchase_currency: patch.purchase_currency,
    })

    return {
      id: result.id,
      name: result.name,
      primary: false,
      retired: result.retired,
      distance: result.distance,
      source: 'custom',
      hashtag: result.hashtag,
      activity_count: 0,
    }
  }

  async deleteCustomGear(id: string, force?: boolean): Promise<{ deleted: boolean }> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.deleteCustomGear(id, force)
    if (result.has_activities && !force) {
      throw new Error(result.message || 'Gear has activities, use force to delete')
    }
    return { deleted: true }
  }

  async getGearMonthlyUsage(includeRetired?: boolean): Promise<GearMonthlyUsage[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getGearMonthlyUsage({ include_retired: includeRetired })
    return result.map((u) => ({
      month: u.month,
      gear_id: u.gear_id,
      gear_name: u.gear_name,
      source: u.source,
      hashtag: u.hashtag,
      retired: u.retired,
      purchase_price: u.purchase_price,
      purchase_currency: u.purchase_currency,
      activity_count: u.activity_count,
      distance: u.distance,
      moving_time: u.moving_time,
    }))
  }

  // ============================================================================
  // Segments
  // ============================================================================
  async getSegments(filters?: SegmentsFilters): Promise<SegmentsResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getSegments({
      activity_type: filters?.activity_type,
      country: filters?.country,
      starred: filters?.starred,
      kom_only: filters?.kom_only,
      search: filters?.search,
      page: filters?.page,
      per_page: filters?.per_page,
      order_by: filters?.order_by,
      order_dir: filters?.order_dir,
    })

    return {
      data: result.data.map((s) => ({
        id: s.id,
        name: s.name,
        activity_type: s.activity_type,
        distance: s.distance,
        average_grade: s.average_grade,
        maximum_grade: s.maximum_grade,
        elevation_high: s.elevation_high,
        elevation_low: s.elevation_low,
        climb_category: s.climb_category,
        start_lat: s.start_lat,
        start_lng: s.start_lng,
        end_lat: s.end_lat,
        end_lng: s.end_lng,
        starred: s.starred,
        polyline: s.polyline,
        times_completed: s.times_completed,
        last_effort_date: s.last_effort_date,
        best_elapsed_time: s.best_elapsed_time,
      })),
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
    }
  }

  async getSegmentCountries(): Promise<SegmentCountryStat[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getSegmentCountries()
    return result.map((c) => ({
      country: c.country,
      count: c.count,
    }))
  }

  async getSegmentDetail(id: number): Promise<SegmentDetailResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getSegmentDetail(id)
    return {
      segment: {
        id: result.segment.id,
        name: result.segment.name,
        activity_type: result.segment.activity_type,
        distance: result.segment.distance,
        average_grade: result.segment.average_grade,
        maximum_grade: result.segment.maximum_grade,
        elevation_high: result.segment.elevation_high,
        elevation_low: result.segment.elevation_low,
        climb_category: result.segment.climb_category,
        start_lat: result.segment.start_lat,
        start_lng: result.segment.start_lng,
        end_lat: result.segment.end_lat,
        end_lng: result.segment.end_lng,
        starred: result.segment.starred,
        polyline: result.segment.polyline,
      },
      efforts: result.efforts.map((e) => ({
        id: e.id,
        segment_id: e.segment_id,
        activity_id: e.activity_id,
        athlete_id: e.athlete_id,
        name: e.name,
        elapsed_time: e.elapsed_time,
        moving_time: e.moving_time,
        start_date: e.start_date,
        start_date_local: e.start_date_local,
        distance: e.distance,
        average_watts: e.average_watts,
        average_heartrate: e.average_heartrate,
        max_heartrate: e.max_heartrate,
        pr_rank: e.pr_rank,
        country: e.country,
      })),
    }
  }

  async getSegmentEfforts(
    id: number,
    filters?: SegmentEffortsFilters
  ): Promise<SegmentEffortsResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getSegmentEfforts({
      segment_id: id,
      page: filters?.page,
      per_page: filters?.per_page,
      order_by: filters?.order_by,
      order_dir: filters?.order_dir,
    })

    return {
      data: result.data.map((e) => ({
        id: e.id,
        segment_id: e.segment_id,
        activity_id: e.activity_id,
        athlete_id: e.athlete_id,
        name: e.name,
        elapsed_time: e.elapsed_time,
        moving_time: e.moving_time,
        start_date: e.start_date,
        start_date_local: e.start_date_local,
        distance: e.distance,
        average_watts: e.average_watts,
        average_heartrate: e.average_heartrate,
        max_heartrate: e.max_heartrate,
        pr_rank: e.pr_rank,
        country: e.country,
      })),
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
    }
  }

  // ============================================================================
  // Athlete FTP/Weight
  // ============================================================================
  async getFtpHistory(): Promise<FTPHistoryResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const cyclingResult = goStorage.getFtpHistory()
    const runningResult = goStorage.getFtpRunningHistory()
    return {
      cycling: cyclingResult.map((e) => ({
        recorded_at: e.recorded_at,
        value: e.value,
      })),
      running: runningResult.map((e) => ({
        recorded_at: e.recorded_at,
        value: e.value,
      })),
    }
  }

  async updateFtpHistory(body: FTPHistoryResponse): Promise<FTPHistoryResponse> {
    this.assertInitialized()
    this.getAthleteId()

    // Update cycling and running FTP separately
    goStorage.updateFtpHistory(
      body.cycling.map((e) => ({ recorded_at: e.recorded_at, value: e.value }))
    )
    goStorage.updateFtpRunningHistory(
      body.running.map((e) => ({ recorded_at: e.recorded_at, value: e.value }))
    )
    return body
  }

  async getWeightHistory(): Promise<WeightHistoryResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getWeightHistory()
    return {
      points: result.map((e) => ({
        recorded_at: e.recorded_at,
        value: e.value,
      })),
    }
  }

  async updateWeightHistory(body: WeightHistoryResponse): Promise<WeightHistoryResponse> {
    this.assertInitialized()
    this.getAthleteId()

    goStorage.updateWeightHistory(
      body.points.map((e) => ({
        recorded_at: e.recorded_at,
        value: e.value,
      }))
    )
    return body
  }

  // ============================================================================
  // Photos
  // ============================================================================
  async getPhotos(filters?: PhotosFilters): Promise<PhotosListResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getPhotos({
      sport_type: filters?.sport_type,
      country: filters?.country,
      page: filters?.page,
      per_page: filters?.per_page,
    })

    return {
      data: result.data.map((p) => ({
        id: p.id,
        activity_id: p.activity_id,
        url: p.url,
        thumbnail_url: p.thumbnail_url,
        caption: p.caption,
        created_at: p.created_at,
        activity_name: p.activity_name,
        sport_type: p.sport_type,
        start_date_local: p.start_date_local,
        location_country: p.location_country,
      })),
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
      countries: result.countries.map((c) => ({
        value: c.value,
        iso2: c.iso2,
        count: c.count,
      })),
      sport_types: result.sport_types.map((s) => ({
        value: s.value,
        count: s.count,
      })),
    }
  }

  async getActivityPhotos(activityId: number): Promise<ActivityPhoto[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getActivityPhotos(activityId)
    return result.map((p) => ({
      id: p.id,
      athlete_id: p.athlete_id,
      activity_id: p.activity_id,
      url: p.url,
      thumbnail_url: p.thumbnail_url,
      caption: p.caption,
      location: p.location,
      created_at: p.created_at,
    }))
  }

  // ============================================================================
  // Challenges
  // ============================================================================
  async getChallenges(filters?: ChallengesFilters): Promise<ChallengesResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getChallenges({
      month: filters?.month,
      page: filters?.page,
      per_page: filters?.per_page,
    })

    return {
      data: result.data.map((c) => ({
        id: String(c.id),
        name: c.name,
        slug: c.slug,
        badge_url: c.badge_url,
        local_badge_url: c.local_badge_url,
        completion_date: c.completion_date,
        month: c.month,
      })),
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
    }
  }

  async importChallenges(_file: File): Promise<{ imported: number }> {
    // File import not supported in WASM mode
    throw new UnsupportedFeatureError(
      'challengeImport',
      'Challenge import is only available in server mode.'
    )
  }

  async importChallengesFromProfile(_athleteId?: string): Promise<{ imported: number }> {
    // Profile import not supported in WASM mode
    throw new UnsupportedFeatureError(
      'challengeImport',
      'Challenge import is only available in server mode.'
    )
  }

  // ============================================================================
  // Goals
  // ============================================================================
  async getTrainingGoals(): Promise<TrainingGoalsResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getTrainingGoals()
    return {
      config: {
        version: result.config.version || 1,
        sports: result.config.sports.map((s) => ({
          name: s.name,
          sport_types: s.sport_types,
          targets: s.targets as Record<
            string,
            { distance_m?: number; elevation_m?: number; moving_time_s?: number }
          >,
        })),
      },
      progress: result.progress as Record<
        string,
        Record<
          string,
          { distance_m: number; elevation_m: number; moving_time_s: number; activity_count: number }
        >
      >,
    }
  }

  async updateTrainingGoals(config: TrainingGoalsConfig): Promise<TrainingGoalsConfig> {
    this.assertInitialized()
    this.getAthleteId()

    goStorage.updateTrainingGoals({
      sports: config.sports.map((s) => ({
        name: s.name,
        sport_types: s.sport_types,
        targets: s.targets,
      })),
    })
    return config
  }

  // ============================================================================
  // Maintenance
  // ============================================================================
  async getMaintenanceDue(): Promise<DueComponent[]> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getMaintenanceDue()
    return result.map((c) => ({
      id: c.id,
      gear_id: c.gear_id,
      name: c.name,
      image_url: c.image_url,
      maintenance_hashtag: c.maintenance_hashtag,
      created_at: c.created_at,
      updated_at: c.updated_at,
      last_completed_at: c.last_completed_at,
      rules: c.rules.map((r) => ({
        id: r.id,
        component_id: r.component_id,
        type: r.type as 'distance_m' | 'time_s' | 'days',
        threshold_value: r.threshold_value,
        created_at: c.created_at,
        updated_at: c.updated_at,
      })),
      distance_since: c.distance_since,
      moving_time_since: c.moving_time_since,
      days_since: c.days_since,
      progress: c.progress.map((p) => ({
        type: p.type as 'distance_m' | 'time_s' | 'days',
        threshold_value: p.threshold_value,
        current_value: p.current_value,
        percent: p.percent,
        due: p.due,
      })),
      is_due: c.is_due,
    }))
  }

  async getGearComponents(
    gearId: string,
    filters?: ComponentsFilters
  ): Promise<ComponentsResponse> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.getGearComponents({
      gear_id: gearId,
      page: filters?.page,
      per_page: filters?.per_page,
    })

    return {
      data: result.data.map((c) => ({
        id: c.id,
        gear_id: c.gear_id,
        name: c.name,
        image_url: c.image_url,
        maintenance_hashtag: c.maintenance_hashtag,
        created_at: c.created_at,
        updated_at: c.updated_at,
        last_completed_at: c.last_completed_at,
        rules: c.rules.map((r) => ({
          id: r.id,
          component_id: r.component_id,
          type: r.type as 'distance_m' | 'time_s' | 'days',
          threshold_value: r.threshold_value,
          created_at: c.created_at,
          updated_at: c.updated_at,
        })),
      })),
      total: result.total,
      page: result.page,
      per_page: result.per_page,
      total_pages: result.total_pages,
    }
  }

  async createComponent(gearId: string, req: CreateComponentRequest): Promise<ComponentWithRules> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.createComponent({
      gear_id: gearId,
      name: req.name,
      image_url: req.image_url,
      maintenance_hashtag: req.maintenance_hashtag,
      rules: req.rules?.map((r) => ({
        type: r.type,
        threshold_value: r.threshold_value,
      })),
    })

    return {
      id: result.id,
      gear_id: result.gear_id,
      name: result.name,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      rules: [],
    }
  }

  async updateComponent(id: number, req: UpdateComponentRequest): Promise<ComponentWithRules> {
    this.assertInitialized()
    this.getAthleteId()

    const result = goStorage.updateComponent({
      id,
      name: req.name,
      image_url: req.image_url,
      maintenance_hashtag: req.maintenance_hashtag,
      rules: req.rules?.map((r) => ({
        type: r.type,
        threshold_value: r.threshold_value,
      })),
    })

    return {
      id: result.id,
      gear_id: result.gear_id,
      name: result.name,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      rules: [],
    }
  }

  async deleteComponent(id: number): Promise<{ deleted: boolean }> {
    this.assertInitialized()
    this.getAthleteId()

    goStorage.deleteComponent(id)
    return { deleted: true }
  }

  async logMaintenance(
    componentId: number,
    req?: LogMaintenanceRequest
  ): Promise<{ logged: boolean }> {
    this.assertInitialized()
    this.getAthleteId()

    goStorage.logMaintenance({
      component_id: componentId,
      activity_id: req?.activity_id,
      completed_at: req?.completed_at,
    })
    return { logged: true }
  }

  // ============================================================================
  // Settings
  // ============================================================================
  async getAppSettings(): Promise<AppSettings> {
    this.assertInitialized()

    const result = goStorage.getAppSettings()
    return result as unknown as AppSettings
  }

  async updateAppSettings(settings: AppSettings): Promise<AppSettings> {
    this.assertInitialized()

    goStorage.updateAppSettings(settings as unknown as goStorage.AppSettings)
    return settings
  }

  // ============================================================================
  // Import (delegates to Strava importer)
  // ============================================================================
  async getImportProgress(): Promise<ImportProgress> {
    this.assertInitialized()

    const progress = goStorage.goGetImportProgress()
    if (!progress) {
      // No import running - return idle state
      return {
        status: 'idle',
        phase: 'idle',
        activities_total: 0,
        activities_done: 0,
        gear_total: 0,
        gear_done: 0,
        streams_total: 0,
        streams_done: 0,
        details_total: 0,
        details_done: 0,
        segments_total: 0,
        segments_done: 0,
        photos_total: 0,
        photos_done: 0,
        total_activities: 0,
        imported_count: 0,
        skipped_count: 0,
        failed_count: 0,
        current_page: 0,
        remaining_api_calls: 0,
        rate_limit_used_15min: 0,
        rate_limit_limit_15min: 100,
        rate_limit_used_daily: 0,
        rate_limit_limit_daily: 1000,
        waiting_for_rate_limit: false,
      }
    }

    // Adapt Go importer progress to API format
    return {
      status: progress.status,
      phase: progress.phase as ImportProgress['phase'],
      activities_total: progress.activities_total,
      activities_done: progress.activities_done,
      gear_total: progress.gear_total,
      gear_done: progress.gear_done,
      streams_total: progress.streams_total,
      streams_done: progress.streams_done,
      details_total: progress.details_total,
      details_done: progress.details_done,
      segments_total: progress.segments_total,
      segments_done: progress.segments_done,
      photos_total: progress.photos_total,
      photos_done: progress.photos_done,
      // Legacy fields
      total_activities: progress.activities_total,
      imported_count: progress.activities_done,
      skipped_count: 0,
      failed_count: progress.failed_count,
      current_page: 0,
      remaining_api_calls: progress.remaining_api_calls,
      estimated_eta: progress.estimated_eta,
      rate_limit_used_15min: progress.rate_limit_used_15min,
      rate_limit_limit_15min: progress.rate_limit_limit_15min,
      rate_limit_used_daily: progress.rate_limit_used_daily,
      rate_limit_limit_daily: progress.rate_limit_limit_daily,
      waiting_for_rate_limit: progress.waiting_for_rate_limit,
    }
  }

  async startImport(req?: StartImportRequest): Promise<{ message: string }> {
    this.assertInitialized()

    const athlete = getAthlete()
    if (!athlete) {
      throw new Error('Not authenticated. Please log in first.')
    }

    goStorage.goStartImport(
      {
        full_sync: req?.full_sync,
        resume: req?.resume,
        skip_streams: req?.skip_streams,
        skip_segments: req?.skip_segments,
        skip_best_efforts: req?.skip_best_efforts,
        skip_photos: req?.skip_photos,
      },
      {
        id: athlete.id,
        username: athlete.username,
        first_name: athlete.firstname,
        last_name: athlete.lastname,
        profile_medium: athlete.profile,
      }
    )
    return { message: 'Import started' }
  }

  async cancelImport(): Promise<{ message: string }> {
    this.assertInitialized()
    goStorage.goCancelImport()
    return { message: 'Import cancelled' }
  }

  async pauseImport(): Promise<{ message: string }> {
    // Go importer uses cancel for pause - state is persisted for resume
    this.assertInitialized()
    goStorage.goCancelImport()
    return { message: 'Import paused' }
  }

  async resumeImport(): Promise<{ message: string }> {
    return this.startImport({ resume: true })
  }

  hasResumableImport(): boolean {
    if (!goStorage.isInitialized()) {
      return false
    }
    const state = goStorage.goGetImportState()
    return state !== null && state.phase !== 'idle' && state.phase !== 'completed'
  }

  async getSyncHistory(limit?: number): Promise<SyncRun[]> {
    this.assertInitialized()
    const history = goStorage.getSyncHistory(limit ?? 10)
    return history as unknown as SyncRun[]
  }

  async getLatestSync(): Promise<SyncRun | null> {
    this.assertInitialized()
    const history = goStorage.getSyncHistory(1)
    return history.length > 0 ? (history[0] as unknown as SyncRun) : null
  }

  async getSyncWatermark(): Promise<SyncWatermark | null> {
    // Not implemented - would need to track in import state
    return null
  }

  // ============================================================================
  // Export
  // ============================================================================
  async getExportStats(): Promise<ExportStats> {
    this.assertInitialized()
    this.getAthleteId()
    return goStorage.getExportStats()
  }

  // ============================================================================
  // Setup (Strava Credentials)
  // ============================================================================
  async getCredentialsStatus(): Promise<CredentialsStatus> {
    // Demo mode: credentials are always "configured" (not needed)
    if (this.options.demoMode) {
      return {
        configured: true,
        client_id: 'demo',
        source: 'browser',
        redirect_uri: '',
      }
    }

    const creds = getCredentials()
    return {
      configured: hasCredentials(),
      client_id: creds?.clientId,
      source: 'browser',
      redirect_uri: window.location.origin + '/auth/callback',
    }
  }

  async updateCredentials(req: UpdateCredentialsRequest): Promise<CredentialsStatus> {
    saveCredentials(req.client_id, req.client_secret)
    return {
      configured: true,
      client_id: req.client_id,
      source: 'browser',
      redirect_uri: window.location.origin + '/auth/callback',
    }
  }

  // ============================================================================
  // Events
  // ============================================================================
  subscribeToEvents(listener: DataEventListener): () => void {
    // Set up callbacks for Go importer events
    importProgressCallback = (progressJson: string) => {
      try {
        const progress = JSON.parse(progressJson) as goStorage.GoImportProgress
        listener(
          createSyncProgressEvent(progress.phase, {
            activities_done: progress.activities_done,
            activities_total: progress.activities_total,
            gear_done: progress.gear_done,
            gear_total: progress.gear_total,
            streams_done: progress.streams_done,
            streams_total: progress.streams_total,
            details_done: progress.details_done,
            details_total: progress.details_total,
            segments_done: progress.segments_done,
            segments_total: progress.segments_total,
            photos_done: progress.photos_done,
            photos_total: progress.photos_total,
            estimated_eta: progress.estimated_eta,
          })
        )
      } catch {
        // Ignore parse errors
      }
    }

    importCompleteCallback = (resultJson: string) => {
      try {
        const result = JSON.parse(resultJson) as { success: boolean; error?: string }
        listener(
          createSyncCompleteEvent(result.success ? 'completed' : 'failed', result.error)
        )
        // Also emit data changed event on successful completion
        if (result.success) {
          listener(createDataChangedEvent({ activities: true, all: true }))
        }
      } catch {
        // Ignore parse errors
      }
    }

    // Return unsubscribe function
    return () => {
      importProgressCallback = null
      importCompleteCallback = null
    }
  }

  // ============================================================================
  // Database Persistence
  // ============================================================================
  async persistDatabase(): Promise<void> {
    await goStorage.persistDatabase()
  }

  async clearDatabase(): Promise<void> {
    await goStorage.clearDatabase()
  }
}
