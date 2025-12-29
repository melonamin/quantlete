/**
 * WasmProvider - DataProvider implementation for WASM mode (browser-only).
 *
 * Uses:
 * - Generated SQL queries from schema/queries/*.sql (via queries.gen.ts)
 * - WASM algorithm bindings for NormalizedPower, Eddington, TrainingLoad
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
  ActivityStream,
  PowerStatsResponse,
  HrZonesResponse,
  TrainingLoadResponse,
  PowerZonesResponse,
  HrZoneDefinition,
  DistributionSlice,
  Gear,
  CustomGearCreateRequest,
  GearMonthlyUsage,
  SegmentListItem,
  SegmentCountryStat,
  SegmentDetailResponse,
  SegmentEffort,
  SegmentsFilters,
  FTPHistoryResponse,
  WeightHistoryResponse,
  BestEffortPR,
  BestEffortItem,
  RewindReport,
  PhotosListResponse,
  PhotosFilters,
  ActivityPhoto,
  Challenge,
  TrainingGoalsConfig,
  TrainingGoalsResponse,
  ComponentWithRules,
  DueComponent,
  CreateComponentRequest,
  UpdateComponentRequest,
  LogMaintenanceRequest,
  AppSettings,
  ImportProgress,
  StartImportRequest,
  ExportStats,
} from '../types'
import { WasmDatabase, initializeDatabase } from '@/lib/wasm/db'
import { queries } from '@/lib/wasm/queries.gen'
import {
  initAlgorithms,
  eddingtonNumber,
  eddingtonNextSteps,
  eddingtonHistory,
  calculateTrainingLoad,
} from '@/lib/wasm/algorithms'
import {
  loadAuth,
  isAuthenticated,
  getAthlete,
  clearAuth,
  getAuthUrl,
  exchangeCode,
  startImport as stravaStartImport,
  cancelImport as stravaCancelImport,
  getImportProgress as stravaGetImportProgress,
  getSyncHistory as stravaGetSyncHistory,
  getLatestSync as stravaGetLatestSync,
} from '@/lib/wasm/strava'
import type { SyncRun as ApiSyncRun, SyncWatermark } from '@/lib/api/import'
export class WasmProvider implements DataProvider {
  private db: WasmDatabase | null = null
  private athleteId: number | null = null

  async initialize(): Promise<void> {
    console.log('[WasmProvider] Initializing...')

    // Initialize database and WASM algorithms in parallel
    const [db] = await Promise.all([initializeDatabase(), initAlgorithms()])

    this.db = db

    // Try to load authentication from database
    const authLoaded = await loadAuth()
    if (authLoaded) {
      const athlete = getAthlete()
      this.athleteId = athlete?.id ?? null
      console.log('[WasmProvider] Loaded auth', { athleteId: this.athleteId })
    } else {
      // Fall back to checking if athlete exists
      const athleteRow = this.db.queryOne<{ id: number }>('SELECT id FROM athletes LIMIT 1')
      this.athleteId = athleteRow?.id ?? null
    }

    console.log('[WasmProvider] Ready', { athleteId: this.athleteId, authenticated: isAuthenticated() })
  }

  private assertInitialized(): WasmDatabase {
    if (!this.db) {
      throw new Error('WasmProvider not initialized. Call initialize() first.')
    }
    return this.db
  }

  private getAthleteId(): number {
    if (this.athleteId === null) {
      throw new Error('No athlete found. Please import activities first.')
    }
    return this.athleteId
  }

  private notImplemented(method: string): never {
    throw new Error(`[WasmProvider] ${method} not yet implemented. Waiting for generated queries.`)
  }

  // ============================================================================
  // Auth
  // ============================================================================
  async getAuthStatus(): Promise<AuthStatus> {
    this.assertInitialized()

    // Check if authenticated via Strava client
    if (!isAuthenticated()) {
      return { authenticated: false, athlete: undefined }
    }

    const athlete = getAthlete()
    if (!athlete) {
      return { authenticated: false, athlete: undefined }
    }

    this.athleteId = athlete.id

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
    // Token refresh is handled automatically by the Strava client
    return { success: true }
  }

  /**
   * Get OAuth URL for Strava login (WASM mode specific).
   */
  getStravaAuthUrl(redirectUri: string): string {
    return getAuthUrl(redirectUri)
  }

  /**
   * Exchange OAuth code for tokens (WASM mode specific).
   */
  async handleOAuthCallback(code: string, redirectUri: string): Promise<void> {
    await exchangeCode(code, redirectUri)
    const athlete = getAthlete()
    this.athleteId = athlete?.id ?? null
  }

  /**
   * Clear authentication (logout).
   */
  async logout(): Promise<void> {
    await clearAuth()
    this.athleteId = null
  }

  // ============================================================================
  // Activities - STUB
  // ============================================================================
  async getActivities(filters: ActivityFilters): Promise<ActivitiesResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()
    const page = filters.page ?? 1
    const perPage = filters.per_page ?? 50
    const offset = (page - 1) * perPage

    type ActivityRow = Activity & Record<string, unknown>

    let data: ActivityRow[]
    let total: number

    if (filters.sport_type) {
      data = queries.getActivitiesBySport<ActivityRow>(db, athleteId, filters.sport_type, perPage, offset)
      const countRow = queries.countActivitiesBySport<{ count: number }>(db, athleteId, filters.sport_type)
      total = countRow?.count ?? 0
    } else {
      data = queries.getActivities<ActivityRow>(db, athleteId, perPage, offset)
      const countRow = queries.countActivities<{ count: number }>(db, athleteId)
      total = countRow?.count ?? 0
    }

    return {
      data,
      total,
      page,
      per_page: perPage,
      total_pages: Math.ceil(total / perPage),
    }
  }

  async getActivity(id: number): Promise<Activity> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type ActivityRow = Activity & Record<string, unknown>
    const activity = queries.getActivity<ActivityRow>(db, id, athleteId)

    if (!activity) {
      throw new Error(`Activity ${id} not found`)
    }

    return activity
  }

  async getActivityStreams(id: number): Promise<ActivityStream[]> {
    const db = this.assertInitialized()

    type StreamRow = {
      stream_type: string
      data: string
      series_type: string
      original_size: number
      resolution: string
    } & Record<string, unknown>

    const rows = queries.getActivityStreams<StreamRow>(db, id)

    return rows.map((row) => ({
      activity_id: id,
      stream_type: row.stream_type,
      original_size: row.original_size,
      resolution: row.resolution,
      series_type: row.series_type,
      data: JSON.parse(row.data),
    }))
  }

  // ============================================================================
  // Dashboard
  // ============================================================================
  async getDashboard(): Promise<DashboardData> {
    const [stats, weekly, recent, sportTypes] = await Promise.all([
      this.getDashboardStats(),
      this.getWeeklyStats(),
      this.getRecentActivities(5),
      this.getSportTypeStats(),
    ])

    return { stats, weekly_stats: weekly, recent_activities: recent, sport_type_stats: sportTypes }
  }

  async getDashboardStats(): Promise<DashboardStats> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type StatsRow = {
      total_activities: number
      total_distance: number
      total_time: number
      total_elevation: number
    } & Record<string, unknown>

    const row = queries.getDashboardStats<StatsRow>(db, athleteId)

    // Return defaults matching DashboardStats interface
    return {
      total_activities: row?.total_activities ?? 0,
      total_distance: row?.total_distance ?? 0,
      total_moving_time: row?.total_time ?? 0,
      total_elevation_gain: row?.total_elevation ?? 0,
      total_calories: 0,
      year_activities: 0,
      year_distance: 0,
      year_moving_time: 0,
      year_elevation_gain: 0,
      month_activities: 0,
      month_distance: 0,
      month_moving_time: 0,
      month_elevation_gain: 0,
    }
  }

  async getWeeklyStats(): Promise<WeeklyStat[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type WeeklyRow = {
      week: string
      week_start: string
      activity_count: number
      total_distance: number
      total_time: number
      total_elevation: number
    } & Record<string, unknown>

    const rows = queries.getWeeklyStats<WeeklyRow>(db, athleteId)

    // Group by sport type is complex - for now return "All" sport type
    return rows.map((row) => ({
      sport_type: 'All',
      activity_count: row.activity_count,
      total_distance: row.total_distance,
      total_time: row.total_time,
      total_elevation: row.total_elevation,
    }))
  }

  async getRecentActivities(limit: number): Promise<RecentActivity[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type RecentRow = RecentActivity & Record<string, unknown>
    return queries.getRecentActivities<RecentRow>(db, athleteId, limit)
  }

  async getSportTypeStats(): Promise<SportTypeStat[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type SportRow = {
      sport_type: string
      activity_count: number
      total_distance: number
      total_time: number
      total_elevation: number
    } & Record<string, unknown>

    const rows = queries.getSportTypeStats<SportRow>(db, athleteId)

    return rows.map((row) => ({
      sport_type: row.sport_type,
      activity_count: row.activity_count,
      total_distance: row.total_distance,
      total_time: row.total_time,
      total_elevation: row.total_elevation,
    }))
  }

  async getMonthlyStats(year?: number): Promise<MonthlyStat[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()
    const yearStr = year?.toString() ?? new Date().getFullYear().toString()

    type MonthlyRow = {
      month: string
      sport_type: string
      activity_count: number
      total_distance: number
      total_time: number
      total_elevation: number
    } & Record<string, unknown>

    const rows = queries.getMonthlyStats<MonthlyRow>(db, athleteId, yearStr)

    return rows.map((row) => ({
      month: row.month,
      sport_type: row.sport_type,
      activity_count: row.activity_count,
      total_distance: row.total_distance,
      total_time: row.total_time,
      total_elevation: row.total_elevation,
    }))
  }

  async getYearlyStats(): Promise<YearlyStat[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type YearlyRow = {
      year: number
      activity_count: number
      total_distance: number
      total_time: number
      total_elevation: number
    } & Record<string, unknown>

    const rows = queries.getYearlyStats<YearlyRow>(db, athleteId)

    return rows.map((row) => ({
      year: row.year,
      activity_count: row.activity_count,
      total_distance: row.total_distance,
      total_time: row.total_time,
      total_elevation: row.total_elevation,
    }))
  }

  async getCalendarData(year: number): Promise<CalendarDay[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type CalendarRow = {
      date: string
      activity_count: number
      total_distance: number
      total_time: number
    } & Record<string, unknown>

    const rows = queries.getCalendarData<CalendarRow>(db, athleteId, year)

    return rows.map((row) => ({
      date: row.date,
      activity_count: row.activity_count,
      total_distance: row.total_distance,
      total_time: row.total_time,
    }))
  }

  async getCalendarActivities(year: number, month: number): Promise<CalendarActivity[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()
    const yearStr = year.toString()
    const monthStr = month.toString().padStart(2, '0')

    type CalendarActivityRow = CalendarActivity & Record<string, unknown>

    return queries.getCalendarActivities<CalendarActivityRow>(db, athleteId, yearStr, monthStr)
  }

  async getCalendarSummary(year: number, month: number): Promise<CalendarMonthSummary> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()
    const yearStr = year.toString()
    const monthStr = month.toString().padStart(2, '0')

    type SummaryRow = {
      activity_count: number
      total_distance: number
      total_elevation: number
      total_time: number
      total_calories: number
      workout_count: number
    } & Record<string, unknown>

    const row = queries.getCalendarSummary<SummaryRow>(db, athleteId, yearStr, monthStr)

    return {
      year,
      month,
      activity_count: row?.activity_count ?? 0,
      total_distance: row?.total_distance ?? 0,
      total_elevation_gain: row?.total_elevation ?? 0,
      total_moving_time: row?.total_time ?? 0,
      total_calories: row?.total_calories ?? 0,
      workout_count: row?.workout_count ?? 0,
      challenges_completed: 0,
    }
  }

  async getDashboardConfig(): Promise<DashboardConfig> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const row = db.queryOne<{ config: string }>(
      'SELECT config FROM dashboard_config WHERE athlete_id = ?',
      [athleteId]
    )

    if (!row) {
      return { version: 1, widgets: [] }
    }

    return JSON.parse(row.config)
  }

  async updateDashboardConfig(config: DashboardConfig): Promise<DashboardConfig> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.exec(
      `INSERT INTO dashboard_config (athlete_id, config, updated_at)
       VALUES (?, ?, datetime('now'))
       ON CONFLICT (athlete_id) DO UPDATE SET
         config = EXCLUDED.config,
         updated_at = EXCLUDED.updated_at`,
      [athleteId, JSON.stringify(config)]
    )

    await db.persist()
    return config
  }

  // ============================================================================
  // Heatmap & Eddington
  // ============================================================================
  async getHeatmapData(filters: HeatmapFilters): Promise<HeatmapResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type HeatmapRow = {
      id: number
      name: string
      sport_type: string
      summary_polyline: string
      start_lat: number
      start_lng: number
    } & Record<string, unknown>

    let activities: HeatmapRow[]

    if (filters.sport_type) {
      activities = queries.getHeatmapActivitiesBySport<HeatmapRow>(db, athleteId, filters.sport_type)
    } else if (filters.after && filters.before) {
      activities = queries.getHeatmapActivitiesByDateRange<HeatmapRow>(
        db,
        athleteId,
        filters.after,
        filters.before
      )
    } else {
      activities = queries.getHeatmapActivities<HeatmapRow>(db, athleteId)
    }

    // Get countries for filter options
    type CountryRow = { country: string; count: number } & Record<string, unknown>
    const countries = queries.getHeatmapCountries<CountryRow>(db, athleteId)

    return {
      activities: activities.map((a) => ({
        id: a.id,
        name: a.name ?? '',
        sport_type: a.sport_type,
        summary_polyline: a.summary_polyline,
        start_lat: a.start_lat,
        start_lng: a.start_lng,
      })),
      total: activities.length,
      countries: countries.map((c) => ({ country: c.country, count: c.count })),
    }
  }

  async getEddingtonData(sportType?: string): Promise<EddingtonResult> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type DayRow = { date: string; distance_km: number } & Record<string, unknown>

    // Get daily distances from database
    let days: DayRow[]
    if (sportType) {
      days = queries.getEddingtonDaysBySport<DayRow>(db, athleteId, sportType)
    } else {
      days = queries.getEddingtonDays<DayRow>(db, athleteId)
    }

    // Extract distances and calculate using WASM
    const distances = days.map((d) => d.distance_km)
    const number = eddingtonNumber(distances)
    const nextSteps = eddingtonNextSteps(distances, number, 5)

    return {
      number,
      distribution: days.map((d) => ({
        date: d.date,
        distance: d.distance_km,
      })),
      next_steps: nextSteps.map((s) => ({
        target: s.target,
        rides_needed: s.daysNeeded,
      })),
    }
  }

  async getEddingtonHistory(sportType?: string): Promise<EddingtonHistoryPoint[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type DayRow = { date: string; distance_km: number } & Record<string, unknown>

    // Get chronological daily distances
    let days: DayRow[]
    if (sportType) {
      // For filtered sport type, need chronological order
      days = db.query<DayRow>(
        `SELECT DATE(start_date) AS date, SUM(distance) / 1000.0 AS distance_km
         FROM activities
         WHERE athlete_id = ? AND sport_type = ?
         GROUP BY DATE(start_date)
         HAVING distance_km > 0
         ORDER BY date ASC`,
        [athleteId, sportType]
      )
    } else {
      days = queries.getEddingtonDaysChronological<DayRow>(db, athleteId)
    }

    // Calculate progressive Eddington numbers using WASM
    const distances = days.map((d) => d.distance_km)
    const history = eddingtonHistory(distances)

    return days.map((day, i) => ({
      date: day.date,
      number: history[i],
    }))
  }

  // ============================================================================
  // Stats & Training
  // ============================================================================
  async getPowerStats(
    _filters?: { after?: string; before?: string; sport_type?: string }
  ): Promise<PowerStatsResponse> {
    // Power stats require power best efforts table - return stub for now
    return { durations_s: [], best: [], history: {} }
  }

  async getPowerZones(): Promise<PowerZonesResponse> {
    return { seconds_by_zone: [], total_seconds: 0, bounds: [] }
  }

  async getHrZones(
    _filters?: { after?: string; before?: string; sport_type?: string }
  ): Promise<HrZonesResponse> {
    return { method: 'reserve', zones: { bounds: [] }, seconds_by_zone: [], total_seconds: 0 }
  }

  async getTrainingLoad(filters?: { after?: string; before?: string }): Promise<TrainingLoadResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    // Default to last 180 days
    const before = filters?.before ?? new Date().toISOString().slice(0, 10)
    const afterDate = new Date()
    afterDate.setDate(afterDate.getDate() - 180)
    const after = filters?.after ?? afterDate.toISOString().slice(0, 10)

    type TSSRow = { day: string; tss: number } & Record<string, unknown>
    const rows = queries.getDailyTSS<TSSRow>(db, athleteId, after, before)

    if (rows.length === 0) {
      return { series: [], summary: undefined }
    }

    // Calculate CTL/ATL/TSB using WASM
    const tssValues = rows.map((r) => r.tss)
    const loadData = calculateTrainingLoad(tssValues)

    const series = rows.map((row, i) => ({
      day: row.day,
      tss: row.tss,
      ctl: loadData[i].ctl,
      atl: loadData[i].atl,
      tsb: loadData[i].tsb,
    }))

    // Summary is the latest point
    const latest = series[series.length - 1]

    return {
      series,
      summary: latest,
    }
  }

  async getHrZoneDefinitions(): Promise<HrZoneDefinition[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type RawRow = { sport_type: string; effective_from: string; method: string; zones: string }

    return db
      .query<RawRow>(
        `SELECT sport_type, effective_from, method, zones
       FROM hr_zone_definitions
       WHERE athlete_id = ?
       ORDER BY effective_from DESC`,
        [athleteId]
      )
      .map((row): HrZoneDefinition => ({
        sport_type: row.sport_type,
        effective_from: row.effective_from,
        method: row.method,
        zones: JSON.parse(row.zones),
      }))
  }

  async upsertHrZoneDefinition(def: HrZoneDefinition): Promise<{ status: string }> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.exec(
      `INSERT INTO hr_zone_definitions (athlete_id, sport_type, effective_from, method, zones)
       VALUES (?, ?, ?, ?, ?)
       ON CONFLICT (athlete_id, sport_type, effective_from) DO UPDATE SET
         method = EXCLUDED.method,
         zones = EXCLUDED.zones`,
      [athleteId, def.sport_type, def.effective_from, def.method, JSON.stringify(def.zones)]
    )

    await db.persist()
    return { status: 'ok' }
  }

  async deleteHrZoneDefinition(
    sportType: string,
    effectiveFrom: string
  ): Promise<{ status: string }> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.delete('hr_zone_definitions', 'athlete_id = ? AND sport_type = ? AND effective_from = ?', [
      athleteId,
      sportType,
      effectiveFrom,
    ])

    await db.persist()
    return { status: 'ok' }
  }

  async getDaytimeDistribution(): Promise<DistributionSlice[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type DistRow = { hour: number; count: number } & Record<string, unknown>
    const rows = queries.getDaytimeDistribution<DistRow>(db, athleteId)

    return rows.map((r) => ({ label: r.hour.toString(), count: r.count }))
  }

  async getWeekdayDistribution(): Promise<DistributionSlice[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

    type DistRow = { weekday: number; count: number } & Record<string, unknown>
    const rows = queries.getWeekdayDistribution<DistRow>(db, athleteId)

    return rows.map((r) => ({ label: dayNames[r.weekday], count: r.count }))
  }

  // ============================================================================
  // Best Efforts
  // ============================================================================
  async getBestEffortPRs(sportType?: string): Promise<BestEffortPR[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type PRRow = {
      distance_type: string
      name: string
      distance_m: number
      elapsed_time: number
      moving_time: number
      pr_rank: number
      activity_id: number
      activity_name: string
      sport_type: string
      start_date_local: string
    } & Record<string, unknown>

    let rows: PRRow[]
    if (sportType) {
      rows = queries.getBestEffortPRsBySport<PRRow>(db, athleteId, sportType)
    } else {
      rows = queries.getBestEffortPRs<PRRow>(db, athleteId)
    }

    return rows.map((r) => ({
      distance_type: r.distance_type,
      name: r.name,
      distance_m: r.distance_m,
      elapsed_time_s: r.elapsed_time,
      moving_time_s: r.moving_time,
      pr_rank: r.pr_rank,
      activity_id: r.activity_id,
      activity_name: r.activity_name,
      sport_type: r.sport_type,
      start_date_local: r.start_date_local,
    }))
  }

  async getBestEffortsForDistance(
    distanceType: string,
    _sportType?: string
  ): Promise<BestEffortItem[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    type EffortRow = {
      distance_type: string
      name: string
      distance_m: number
      elapsed_time: number
      moving_time: number
      pr_rank: number
      activity_id: number
      activity_name: string
      sport_type: string
      start_date_local: string
      rn: number
    } & Record<string, unknown>

    const rows = queries.getBestEffortsForDistance<EffortRow>(db, athleteId, distanceType)

    return rows.map((r) => ({
      distance_type: r.distance_type,
      name: r.name,
      distance_m: r.distance_m,
      elapsed_time_s: r.elapsed_time,
      moving_time_s: r.moving_time,
      pr_rank: r.pr_rank,
      activity_id: r.activity_id,
      activity_name: r.activity_name,
      sport_type: r.sport_type,
      start_date_local: r.start_date_local,
    }))
  }

  // ============================================================================
  // Rewind - STUB
  // ============================================================================
  async getRewindYears(): Promise<number[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const rows = db.query<{ year: string }>(
      `SELECT DISTINCT strftime('%Y', start_date) as year
       FROM activities
       WHERE athlete_id = ?
       ORDER BY year DESC`,
      [athleteId]
    )

    return rows.map((r) => parseInt(r.year, 10))
  }

  async getRewind(_year: number): Promise<RewindReport> {
    this.notImplemented('getRewind')
  }

  // ============================================================================
  // Gear - STUB
  // ============================================================================
  async getGear(_includeRetired?: boolean): Promise<Gear[]> {
    this.notImplemented('getGear')
  }

  async getGearDetail(_id: string): Promise<Gear> {
    this.notImplemented('getGearDetail')
  }

  async getCustomGear(_includeRetired?: boolean): Promise<Gear[]> {
    this.notImplemented('getCustomGear')
  }

  async createCustomGear(_req: CustomGearCreateRequest): Promise<Gear> {
    this.notImplemented('createCustomGear')
  }

  async updateCustomGear(_id: string, _patch: Partial<CustomGearCreateRequest>): Promise<Gear> {
    this.notImplemented('updateCustomGear')
  }

  async deleteCustomGear(_id: string, _force?: boolean): Promise<{ deleted: boolean }> {
    this.notImplemented('deleteCustomGear')
  }

  async getGearMonthlyUsage(_includeRetired?: boolean): Promise<GearMonthlyUsage[]> {
    this.notImplemented('getGearMonthlyUsage')
  }

  // ============================================================================
  // Segments - STUB
  // ============================================================================
  async getSegments(_filters?: SegmentsFilters): Promise<SegmentListItem[]> {
    this.notImplemented('getSegments')
  }

  async getSegmentCountries(): Promise<SegmentCountryStat[]> {
    this.notImplemented('getSegmentCountries')
  }

  async getSegmentDetail(_id: number): Promise<SegmentDetailResponse> {
    this.notImplemented('getSegmentDetail')
  }

  async getSegmentEfforts(_id: number): Promise<SegmentEffort[]> {
    this.notImplemented('getSegmentEfforts')
  }

  // ============================================================================
  // Athlete (FTP/Weight)
  // ============================================================================
  async getFtpHistory(): Promise<FTPHistoryResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const cycling = db.query<{ recorded_at: string; value: number }>(
      `SELECT recorded_at, value FROM athlete_metrics
       WHERE athlete_id = ? AND metric = 'ftp_cycling_watts'
       ORDER BY recorded_at DESC`,
      [athleteId]
    )

    const running = db.query<{ recorded_at: string; value: number }>(
      `SELECT recorded_at, value FROM athlete_metrics
       WHERE athlete_id = ? AND metric = 'ftp_running_mps'
       ORDER BY recorded_at DESC`,
      [athleteId]
    )

    return { cycling, running }
  }

  async updateFtpHistory(body: FTPHistoryResponse): Promise<FTPHistoryResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.delete('athlete_metrics', "athlete_id = ? AND metric IN ('ftp_cycling_watts', 'ftp_running_mps')", [
      athleteId,
    ])

    for (const point of body.cycling) {
      db.exec(
        `INSERT INTO athlete_metrics (athlete_id, metric, value, recorded_at)
         VALUES (?, 'ftp_cycling_watts', ?, ?)`,
        [athleteId, point.value, point.recorded_at]
      )
    }

    for (const point of body.running) {
      db.exec(
        `INSERT INTO athlete_metrics (athlete_id, metric, value, recorded_at)
         VALUES (?, 'ftp_running_mps', ?, ?)`,
        [athleteId, point.value, point.recorded_at]
      )
    }

    await db.persist()
    return body
  }

  async getWeightHistory(): Promise<WeightHistoryResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const points = db.query<{ recorded_at: string; value: number }>(
      `SELECT recorded_at, value FROM athlete_metrics
       WHERE athlete_id = ? AND metric = 'weight_kg'
       ORDER BY recorded_at DESC`,
      [athleteId]
    )

    return { points }
  }

  async updateWeightHistory(body: WeightHistoryResponse): Promise<WeightHistoryResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.delete('athlete_metrics', "athlete_id = ? AND metric = 'weight_kg'", [athleteId])

    for (const point of body.points) {
      db.exec(
        `INSERT INTO athlete_metrics (athlete_id, metric, value, recorded_at)
         VALUES (?, 'weight_kg', ?, ?)`,
        [athleteId, point.value, point.recorded_at]
      )
    }

    await db.persist()
    return body
  }

  // ============================================================================
  // Photos - STUB
  // ============================================================================
  async getPhotos(_filters?: PhotosFilters): Promise<PhotosListResponse> {
    return { data: [], total: 0, page: 1, per_page: 50, total_pages: 0, countries: [], sport_types: [] }
  }

  async getActivityPhotos(_activityId: number): Promise<ActivityPhoto[]> {
    return []
  }

  // ============================================================================
  // Challenges - STUB
  // ============================================================================
  async getChallenges(_month?: string): Promise<Challenge[]> {
    return []
  }

  async importChallenges(_file: File): Promise<{ imported: number }> {
    return { imported: 0 }
  }

  // ============================================================================
  // Goals
  // ============================================================================
  async getTrainingGoals(): Promise<TrainingGoalsResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const row = db.queryOne<{ config: string }>(
      'SELECT config FROM training_goals WHERE athlete_id = ?',
      [athleteId]
    )

    const config: TrainingGoalsConfig = row ? JSON.parse(row.config) : { version: 1, sports: [] }

    return { config, progress: {} }
  }

  async updateTrainingGoals(config: TrainingGoalsConfig): Promise<TrainingGoalsConfig> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.exec(
      `INSERT INTO training_goals (athlete_id, config, updated_at)
       VALUES (?, ?, datetime('now'))
       ON CONFLICT (athlete_id) DO UPDATE SET
         config = EXCLUDED.config,
         updated_at = EXCLUDED.updated_at`,
      [athleteId, JSON.stringify(config)]
    )

    await db.persist()
    return config
  }

  // ============================================================================
  // Maintenance - STUB
  // ============================================================================
  async getMaintenanceDue(): Promise<DueComponent[]> {
    return []
  }

  async getGearComponents(_gearId: string): Promise<ComponentWithRules[]> {
    return []
  }

  async createComponent(_gearId: string, _req: CreateComponentRequest): Promise<ComponentWithRules> {
    this.notImplemented('createComponent')
  }

  async updateComponent(_id: number, _req: UpdateComponentRequest): Promise<ComponentWithRules> {
    this.notImplemented('updateComponent')
  }

  async deleteComponent(_id: number): Promise<{ deleted: boolean }> {
    return { deleted: false }
  }

  async logMaintenance(_componentId: number, _req?: LogMaintenanceRequest): Promise<{ logged: boolean }> {
    return { logged: false }
  }

  // ============================================================================
  // Settings
  // ============================================================================
  async getAppSettings(): Promise<AppSettings> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const row = db.queryOne<{ settings: string }>(
      'SELECT settings FROM athlete_settings WHERE athlete_id = ?',
      [athleteId]
    )

    if (!row) {
      return {
        version: 1,
        virtual_world_tile_layers: {},
        eddington_definitions: [],
      }
    }

    return JSON.parse(row.settings)
  }

  async updateAppSettings(settings: AppSettings): Promise<AppSettings> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.exec(
      `INSERT INTO athlete_settings (athlete_id, settings, updated_at)
       VALUES (?, ?, datetime('now'))
       ON CONFLICT (athlete_id) DO UPDATE SET
         settings = EXCLUDED.settings,
         updated_at = EXCLUDED.updated_at`,
      [athleteId, JSON.stringify(settings)]
    )

    await db.persist()
    return settings
  }

  // ============================================================================
  // Import - WASM mode uses browser-based import
  // ============================================================================
  async getImportProgress(): Promise<ImportProgress> {
    const progress = stravaGetImportProgress()

    // Map status values between browser importer and API types
    const statusMap: Record<string, ImportProgress['status']> = {
      idle: 'idle',
      running: 'running',
      complete: 'completed',
      error: 'failed',
      cancelled: 'canceled',
    }

    const status = statusMap[progress.status] ?? 'idle'

    return {
      status,
      // Current phase (WASM uses simplified single-phase import)
      phase: status === 'completed' ? 'completed' : status === 'running' ? 'activities' : 'idle',

      // Per-phase progress (mapped to activities phase for WASM)
      activities_total: progress.total,
      activities_done: progress.imported,
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

      // Legacy fields
      total_activities: progress.total,
      imported_count: progress.imported,
      skipped_count: progress.skipped,
      failed_count: progress.failed,
      current_page: 0,
      error: progress.error,

      // ETA and rate limits (not available in WASM mode)
      remaining_api_calls: 0,
      estimated_eta: undefined,
      rate_limit_used_15min: 0,
      rate_limit_limit_15min: 0,
      rate_limit_used_daily: 0,
      rate_limit_limit_daily: 0,

      // Rate limit waiting state (not applicable in WASM mode)
      waiting_for_rate_limit: false,
      waiting_until: undefined,
      waiting_reason: undefined,
    }
  }

  async startImport(req?: StartImportRequest): Promise<{ message: string }> {
    // Start import in background (non-blocking)
    // WASM importer uses inverted logic (include = !skip)
    stravaStartImport({
      fullSync: req?.full_sync,
      includeStreams: !req?.skip_streams,
    }).catch((err) => {
      console.error('[WasmProvider] Import error:', err)
    })

    return { message: 'Import started' }
  }

  async cancelImport(): Promise<{ message: string }> {
    stravaCancelImport()
    return { message: 'Import cancelled' }
  }

  async getSyncHistory(limit = 10): Promise<ApiSyncRun[]> {
    const runs = stravaGetSyncHistory(limit)

    // Map WASM SyncRun to API SyncRun format
    return runs.map((r): ApiSyncRun => ({
      id: r.id,
      athlete_id: r.athlete_id,
      started_at: r.started_at,
      completed_at: r.completed_at,
      duration_seconds: r.duration_seconds,
      status: r.status as ApiSyncRun['status'],
      error: r.error,
      activities_total: r.activities_total,
      activities_imported: r.activities_imported,
      activities_skipped: r.activities_skipped,
      gear_imported: 0, // WASM mode doesn't import gear yet
      streams_imported: r.streams_imported,
      segments_imported: 0, // WASM mode doesn't import segments yet
      photos_imported: 0, // WASM mode doesn't import photos yet
      failed_count: r.failed_count,
      full_sync: r.full_sync,
      skip_streams: r.skip_streams,
      skip_segments: true, // WASM mode always skips segments
      skip_best_efforts: true, // WASM mode always skips best efforts
      skip_photos: true, // WASM mode always skips photos
      newest_activity_date: r.newest_activity_date,
      created_at: r.started_at, // Use started_at as created_at
    }))
  }

  async getLatestSync(): Promise<ApiSyncRun | null> {
    const run = stravaGetLatestSync()
    if (!run) return null

    return {
      id: run.id,
      athlete_id: run.athlete_id,
      started_at: run.started_at,
      completed_at: run.completed_at,
      duration_seconds: run.duration_seconds,
      status: run.status as ApiSyncRun['status'],
      error: run.error,
      activities_total: run.activities_total,
      activities_imported: run.activities_imported,
      activities_skipped: run.activities_skipped,
      gear_imported: 0,
      streams_imported: run.streams_imported,
      segments_imported: 0,
      photos_imported: 0,
      failed_count: run.failed_count,
      full_sync: run.full_sync,
      skip_streams: run.skip_streams,
      skip_segments: true,
      skip_best_efforts: true,
      skip_photos: true,
      newest_activity_date: run.newest_activity_date,
      created_at: run.started_at,
    }
  }

  async getSyncWatermark(): Promise<SyncWatermark | null> {
    return null
  }

  async getExportStats(): Promise<ExportStats> {
    const db = this.assertInitialized()
    const row = db.queryOne<{ total: number; first_activity: string | null; last_activity: string | null }>(
      `SELECT
         COUNT(*) as total,
         MIN(start_date) as first_activity,
         MAX(start_date) as last_activity
       FROM activities`
    )

    return {
      total_activities: row?.total ?? 0,
      first_activity: row?.first_activity ?? null,
      last_activity: row?.last_activity ?? null,
    }
  }
}
