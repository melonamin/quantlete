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
  getRateLimitInfo,
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
  // Activities
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
  // Rewind
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

  async getRewind(year: number): Promise<RewindReport> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const yearStr = String(year)
    const rangeStart = `${yearStr}-01-01`
    const rangeEnd = `${yearStr}-12-31`

    // Get totals for the year
    const totals = db.queryOne<{
      activities: number
      distance_m: number
      elevation_m: number
      moving_time_s: number
      kudos: number
      commute_distance_m: number
    }>(
      `SELECT
        COUNT(*) as activities,
        COALESCE(SUM(distance), 0) as distance_m,
        COALESCE(SUM(total_elevation_gain), 0) as elevation_m,
        COALESCE(SUM(moving_time), 0) as moving_time_s,
        COALESCE(SUM(kudos_count), 0) as kudos,
        COALESCE(SUM(CASE WHEN commute = 1 THEN distance ELSE 0 END), 0) as commute_distance_m
      FROM activities
      WHERE athlete_id = ? AND strftime('%Y', start_date) = ?`,
      [athleteId, yearStr]
    )

    // Calculate carbon saved (assuming 0.21 kg CO2/km for car)
    const carbonSavedKg = ((totals?.commute_distance_m || 0) / 1000) * 0.21

    // Get active days count
    const activeDays = db.queryOne<{ count: number }>(
      `SELECT COUNT(DISTINCT DATE(start_date)) as count
       FROM activities
       WHERE athlete_id = ? AND strftime('%Y', start_date) = ?`,
      [athleteId, yearStr]
    )

    // Calculate total days and rest days
    const totalDays = year === new Date().getFullYear()
      ? Math.floor((Date.now() - new Date(rangeStart).getTime()) / (1000 * 60 * 60 * 24)) + 1
      : 365 + (year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0) ? 1 : 0)
    const active = activeDays?.count || 0
    const restDays = totalDays - active

    // Monthly stats
    const months = db.query<{
      month: string
      activities: number
      distance_m: number
      elevation_m: number
    }>(
      `SELECT
        strftime('%Y-%m', start_date) as month,
        COUNT(*) as activities,
        COALESCE(SUM(distance), 0) as distance_m,
        COALESCE(SUM(total_elevation_gain), 0) as elevation_m
      FROM activities
      WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
      GROUP BY month
      ORDER BY month`,
      [athleteId, yearStr]
    )

    // Get PR counts per month from best_efforts (pr_rank = 1)
    const prsByMonth = db.query<{ month: string; prs: number }>(
      `SELECT
        strftime('%Y-%m', a.start_date) as month,
        COUNT(*) as prs
      FROM best_efforts be
      JOIN activities a ON a.id = be.activity_id
      WHERE a.athlete_id = ? AND strftime('%Y', a.start_date) = ? AND be.pr_rank = 1
      GROUP BY month`,
      [athleteId, yearStr]
    )
    const prsMap = new Map(prsByMonth.map((p) => [p.month, p.prs]))

    const monthsWithPrs = months.map((m) => ({
      month: m.month,
      activities: m.activities,
      distance_m: m.distance_m,
      elevation_m: m.elevation_m,
      prs: prsMap.get(m.month) || 0,
    }))

    // Moving time by sport
    const sportTimes = db.query<{ sport_type: string; moving_time_s: number }>(
      `SELECT sport_type, COALESCE(SUM(moving_time), 0) as moving_time_s
       FROM activities
       WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
       GROUP BY sport_type
       ORDER BY moving_time_s DESC`,
      [athleteId, yearStr]
    )

    // Start times by hour
    const hourCounts = db.query<{ hour: number; count: number }>(
      `SELECT CAST(strftime('%H', start_date_local) AS INTEGER) as hour, COUNT(*) as count
       FROM activities
       WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
       GROUP BY hour
       ORDER BY hour`,
      [athleteId, yearStr]
    )

    // Locations (start points with counts)
    const locations = db.query<{ lat: number; lng: number; count: number }>(
      `SELECT
        ROUND(start_lat, 2) as lat,
        ROUND(start_lng, 2) as lng,
        COUNT(*) as count
      FROM activities
      WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
        AND start_lat IS NOT NULL AND start_lng IS NOT NULL
      GROUP BY lat, lng
      ORDER BY count DESC
      LIMIT 100`,
      [athleteId, yearStr]
    )

    // Streaks calculation
    const activityDates = db.query<{ date: string }>(
      `SELECT DISTINCT DATE(start_date) as date
       FROM activities
       WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
       ORDER BY date`,
      [athleteId, yearStr]
    )

    const dateSet = new Set(activityDates.map((d) => d.date))
    let longestActiveDays = 0
    let longestRestDays = 0
    let currentActiveStreak = 0
    let currentRestStreak = 0

    // Iterate through all days of the year
    const startDate = new Date(rangeStart)
    const endDate = new Date(Math.min(new Date(rangeEnd).getTime(), Date.now()))

    for (let d = new Date(startDate); d <= endDate; d.setDate(d.getDate() + 1)) {
      const dateStr = d.toISOString().slice(0, 10)
      if (dateSet.has(dateStr)) {
        currentActiveStreak++
        longestActiveDays = Math.max(longestActiveDays, currentActiveStreak)
        if (currentRestStreak > 0) {
          longestRestDays = Math.max(longestRestDays, currentRestStreak)
          currentRestStreak = 0
        }
      } else {
        currentRestStreak++
        longestRestDays = Math.max(longestRestDays, currentRestStreak)
        if (currentActiveStreak > 0) {
          longestActiveDays = Math.max(longestActiveDays, currentActiveStreak)
          currentActiveStreak = 0
        }
      }
    }

    // Biggest activities
    const longestDistance = db.queryOne<{
      activity_id: number
      name: string
      sport_type: string
      start_date_local: string
      value: number
    }>(
      `SELECT id as activity_id, name, sport_type, start_date_local, distance as value
       FROM activities
       WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
       ORDER BY distance DESC
       LIMIT 1`,
      [athleteId, yearStr]
    )

    const mostElevation = db.queryOne<{
      activity_id: number
      name: string
      sport_type: string
      start_date_local: string
      value: number
    }>(
      `SELECT id as activity_id, name, sport_type, start_date_local, total_elevation_gain as value
       FROM activities
       WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
       ORDER BY total_elevation_gain DESC
       LIMIT 1`,
      [athleteId, yearStr]
    )

    const longestDuration = db.queryOne<{
      activity_id: number
      name: string
      sport_type: string
      start_date_local: string
      value: number
    }>(
      `SELECT id as activity_id, name, sport_type, start_date_local, moving_time as value
       FROM activities
       WHERE athlete_id = ? AND strftime('%Y', start_date) = ?
       ORDER BY moving_time DESC
       LIMIT 1`,
      [athleteId, yearStr]
    )

    // Random photo from the year
    const randomPhoto = db.queryOne<{
      id: string
      activity_id: number
      url: string
      caption: string | null
    }>(
      `SELECT p.id, p.activity_id, p.url, p.caption
       FROM photos p
       JOIN activities a ON a.id = p.activity_id
       WHERE a.athlete_id = ? AND strftime('%Y', a.start_date) = ?
       ORDER BY RANDOM()
       LIMIT 1`,
      [athleteId, yearStr]
    )

    return {
      year,
      range_start: rangeStart,
      range_end: rangeEnd,
      total_days: totalDays,
      active_days: active,
      rest_days: restDays,
      totals: {
        activities: totals?.activities || 0,
        distance_m: totals?.distance_m || 0,
        elevation_m: totals?.elevation_m || 0,
        moving_time_s: totals?.moving_time_s || 0,
        kudos: totals?.kudos || 0,
        commute_distance_m: totals?.commute_distance_m || 0,
        carbon_saved_kg: carbonSavedKg,
      },
      months: monthsWithPrs,
      moving_time_by_sport: sportTimes,
      start_times_by_hour: hourCounts,
      locations,
      streaks: {
        longest_active_days: longestActiveDays,
        longest_rest_days: longestRestDays,
      },
      random_photo: randomPhoto
        ? {
            id: randomPhoto.id,
            activity_id: randomPhoto.activity_id,
            url: randomPhoto.url,
            caption: randomPhoto.caption ?? undefined,
          }
        : undefined,
      biggest: {
        longest_distance: longestDistance ?? undefined,
        most_elevation: mostElevation ?? undefined,
        longest_duration: longestDuration ?? undefined,
      },
    }
  }

  // ============================================================================
  // Gear
  // ============================================================================
  async getGear(includeRetired = true): Promise<Gear[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const retiredFilter = includeRetired ? '' : 'AND g.retired = 0'
    const rows = db.query<{
      id: string
      name: string
      is_primary: number
      retired: number
      distance: number
      brand_name: string | null
      model_name: string | null
      description: string | null
      source: string
      hashtag: string | null
      purchase_price: number | null
      purchase_currency: string | null
      activity_count: number
    }>(
      `SELECT g.id, g.name, g.is_primary, g.retired, g.distance,
              g.brand_name, g.model_name, g.description, g.source,
              g.hashtag, g.purchase_price, g.purchase_currency,
              COUNT(a.id) as activity_count
       FROM gear g
       LEFT JOIN activities a ON a.gear_id = g.id AND a.athlete_id = g.athlete_id
       WHERE g.athlete_id = ? ${retiredFilter}
       GROUP BY g.id
       ORDER BY g.distance DESC`,
      [athleteId]
    )

    return rows.map((r) => ({
      id: r.id,
      name: r.name,
      primary: r.is_primary === 1,
      retired: r.retired === 1,
      distance: r.distance,
      brand_name: r.brand_name ?? undefined,
      model_name: r.model_name ?? undefined,
      description: r.description ?? undefined,
      source: r.source,
      hashtag: r.hashtag ?? undefined,
      purchase_price: r.purchase_price ?? undefined,
      purchase_currency: r.purchase_currency ?? undefined,
      activity_count: r.activity_count,
    }))
  }

  async getGearDetail(id: string): Promise<Gear> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const row = db.queryOne<{
      id: string
      name: string
      is_primary: number
      retired: number
      distance: number
      brand_name: string | null
      model_name: string | null
      description: string | null
      source: string
      hashtag: string | null
      purchase_price: number | null
      purchase_currency: string | null
      activity_count: number
    }>(
      `SELECT g.id, g.name, g.is_primary, g.retired, g.distance,
              g.brand_name, g.model_name, g.description, g.source,
              g.hashtag, g.purchase_price, g.purchase_currency,
              COUNT(a.id) as activity_count
       FROM gear g
       LEFT JOIN activities a ON a.gear_id = g.id AND a.athlete_id = g.athlete_id
       WHERE g.id = ? AND g.athlete_id = ?
       GROUP BY g.id`,
      [id, athleteId]
    )

    if (!row) {
      throw new Error(`Gear not found: ${id}`)
    }

    return {
      id: row.id,
      name: row.name,
      primary: row.is_primary === 1,
      retired: row.retired === 1,
      distance: row.distance,
      brand_name: row.brand_name ?? undefined,
      model_name: row.model_name ?? undefined,
      description: row.description ?? undefined,
      source: row.source,
      hashtag: row.hashtag ?? undefined,
      purchase_price: row.purchase_price ?? undefined,
      purchase_currency: row.purchase_currency ?? undefined,
      activity_count: row.activity_count,
    }
  }

  async getCustomGear(includeRetired = true): Promise<Gear[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const retiredFilter = includeRetired ? '' : 'AND g.retired = 0'
    const rows = db.query<{
      id: string
      name: string
      is_primary: number
      retired: number
      distance: number
      brand_name: string | null
      model_name: string | null
      description: string | null
      source: string
      hashtag: string | null
      purchase_price: number | null
      purchase_currency: string | null
      activity_count: number
    }>(
      `SELECT g.id, g.name, g.is_primary, g.retired, g.distance,
              g.brand_name, g.model_name, g.description, g.source,
              g.hashtag, g.purchase_price, g.purchase_currency,
              COUNT(a.id) as activity_count
       FROM gear g
       LEFT JOIN activities a ON a.gear_id = g.id AND a.athlete_id = g.athlete_id
       WHERE g.athlete_id = ? AND g.source = 'custom' ${retiredFilter}
       GROUP BY g.id
       ORDER BY g.name`,
      [athleteId]
    )

    return rows.map((r) => ({
      id: r.id,
      name: r.name,
      primary: r.is_primary === 1,
      retired: r.retired === 1,
      distance: r.distance,
      brand_name: r.brand_name ?? undefined,
      model_name: r.model_name ?? undefined,
      description: r.description ?? undefined,
      source: r.source,
      hashtag: r.hashtag ?? undefined,
      purchase_price: r.purchase_price ?? undefined,
      purchase_currency: r.purchase_currency ?? undefined,
      activity_count: r.activity_count,
    }))
  }

  async createCustomGear(req: CustomGearCreateRequest): Promise<Gear> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const id = `custom_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`

    db.exec(
      `INSERT INTO gear (id, athlete_id, name, hashtag, retired, purchase_price, purchase_currency, source)
       VALUES (?, ?, ?, ?, ?, ?, ?, 'custom')`,
      [id, athleteId, req.name, req.hashtag, req.retired ? 1 : 0, req.purchase_price ?? null, req.purchase_currency ?? null]
    )
    await db.persist()

    return this.getGearDetail(id)
  }

  async updateCustomGear(id: string, patch: Partial<CustomGearCreateRequest>): Promise<Gear> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const sets: string[] = []
    const values: unknown[] = []

    if (patch.name !== undefined) {
      sets.push('name = ?')
      values.push(patch.name)
    }
    if (patch.hashtag !== undefined) {
      sets.push('hashtag = ?')
      values.push(patch.hashtag)
    }
    if (patch.retired !== undefined) {
      sets.push('retired = ?')
      values.push(patch.retired ? 1 : 0)
    }
    if (patch.purchase_price !== undefined) {
      sets.push('purchase_price = ?')
      values.push(patch.purchase_price)
    }
    if (patch.purchase_currency !== undefined) {
      sets.push('purchase_currency = ?')
      values.push(patch.purchase_currency)
    }

    if (sets.length > 0) {
      sets.push('updated_at = CURRENT_TIMESTAMP')
      values.push(id, athleteId)
      db.exec(
        `UPDATE gear SET ${sets.join(', ')} WHERE id = ? AND athlete_id = ? AND source = 'custom'`,
        values
      )
      await db.persist()
    }

    return this.getGearDetail(id)
  }

  async deleteCustomGear(id: string, _force?: boolean): Promise<{ deleted: boolean }> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    db.exec(
      `DELETE FROM gear WHERE id = ? AND athlete_id = ? AND source = 'custom'`,
      [id, athleteId]
    )
    await db.persist()

    return { deleted: true }
  }

  async getGearMonthlyUsage(includeRetired = true): Promise<GearMonthlyUsage[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const retiredFilter = includeRetired ? '' : 'AND g.retired = 0'
    const rows = db.query<{
      month: string
      gear_id: string
      gear_name: string
      source: string
      hashtag: string | null
      retired: number
      purchase_price: number | null
      purchase_currency: string | null
      activity_count: number
      distance: number
      moving_time: number
    }>(
      `SELECT
         strftime('%Y-%m', a.start_date) AS month,
         g.id AS gear_id,
         g.name AS gear_name,
         g.source,
         g.hashtag,
         g.retired,
         g.purchase_price,
         g.purchase_currency,
         COUNT(*) AS activity_count,
         COALESCE(SUM(a.distance), 0) AS distance,
         COALESCE(SUM(a.moving_time), 0) AS moving_time
       FROM gear g
       JOIN activities a ON a.gear_id = g.id AND a.athlete_id = g.athlete_id
       WHERE g.athlete_id = ? ${retiredFilter}
       GROUP BY month, g.id
       ORDER BY month DESC, g.name`,
      [athleteId]
    )

    return rows.map((r) => ({
      month: r.month,
      gear_id: r.gear_id,
      gear_name: r.gear_name,
      source: r.source,
      hashtag: r.hashtag ?? undefined,
      retired: r.retired === 1,
      purchase_price: r.purchase_price ?? undefined,
      purchase_currency: r.purchase_currency ?? undefined,
      activity_count: r.activity_count,
      distance: r.distance,
      moving_time: r.moving_time,
    }))
  }

  // ============================================================================
  // Segments
  // ============================================================================
  async getSegments(filters?: SegmentsFilters): Promise<SegmentListItem[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    // Build query based on filters
    let rows: Array<{
      id: number
      name: string
      activity_type: string
      distance: number
      average_grade: number
      maximum_grade: number
      elevation_high: number
      elevation_low: number
      climb_category: number
      starred: number
      athlete_kom_rank: number | null
      athlete_effort_count: number | null
      athlete_pr_elapsed_time: number | null
      athlete_pr_date: string | null
      times_completed: number
      last_effort_date: string | null
      best_elapsed_time: number | null
    }>

    if (filters?.country) {
      rows = queries.getSegmentsByCountry(db, athleteId, filters.country)
    } else {
      rows = queries.getSegments(db, athleteId)
    }

    // Apply additional filters in-memory
    let result = rows.map((r) => ({
      id: r.id,
      name: r.name,
      activity_type: r.activity_type,
      distance: r.distance,
      average_grade: r.average_grade,
      maximum_grade: r.maximum_grade,
      elevation_high: r.elevation_high,
      elevation_low: r.elevation_low,
      climb_category: r.climb_category,
      starred: Boolean(r.starred),
      athlete_kom_rank: r.athlete_kom_rank ?? undefined,
      athlete_effort_count: r.athlete_effort_count ?? undefined,
      athlete_pr_elapsed_time: r.athlete_pr_elapsed_time ?? undefined,
      athlete_pr_date: r.athlete_pr_date ?? undefined,
      times_completed: r.times_completed || 0,
      last_effort_date: r.last_effort_date ?? undefined,
      best_elapsed_time: r.best_elapsed_time ?? undefined,
    }))

    // Activity type filter
    if (filters?.activity_type) {
      result = result.filter((s) => s.activity_type === filters.activity_type)
    }

    // Starred filter
    if (filters?.starred) {
      result = result.filter((s) => s.starred)
    }

    // KOM only filter
    if (filters?.kom_only) {
      result = result.filter((s) => s.athlete_kom_rank === 1)
    }

    // Search filter
    if (filters?.search) {
      const searchLower = filters.search.toLowerCase()
      result = result.filter((s) => s.name.toLowerCase().includes(searchLower))
    }

    // Limit
    if (filters?.limit && filters.limit > 0) {
      result = result.slice(0, filters.limit)
    }

    return result
  }

  async getSegmentCountries(): Promise<SegmentCountryStat[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const rows = queries.getSegmentCountries<{
      country: string
      segment_count: number
    }>(db, athleteId)

    return rows.map((r) => ({
      country: r.country,
      count: r.segment_count,
    }))
  }

  async getSegmentDetail(id: number): Promise<SegmentDetailResponse> {
    const db = this.assertInitialized()
    this.getAthleteId() // Ensure authenticated

    const segment = queries.getSegmentDetail<{
      id: number
      name: string
      activity_type: string
      distance: number
      average_grade: number
      maximum_grade: number
      elevation_high: number
      elevation_low: number
      climb_category: number
      start_lat: number | null
      start_lng: number | null
      end_lat: number | null
      end_lng: number | null
      starred: number
      polyline: string | null
      athlete_kom_rank: number | null
      athlete_effort_count: number | null
      athlete_pr_elapsed_time: number | null
      athlete_pr_date: string | null
    }>(db, id)

    if (!segment) {
      throw new Error(`Segment ${id} not found`)
    }

    const efforts = await this.getSegmentEfforts(id)

    return {
      segment: {
        id: segment.id,
        name: segment.name,
        activity_type: segment.activity_type,
        distance: segment.distance,
        average_grade: segment.average_grade,
        maximum_grade: segment.maximum_grade,
        elevation_high: segment.elevation_high,
        elevation_low: segment.elevation_low,
        climb_category: segment.climb_category,
        start_lat: segment.start_lat ?? undefined,
        start_lng: segment.start_lng ?? undefined,
        end_lat: segment.end_lat ?? undefined,
        end_lng: segment.end_lng ?? undefined,
        starred: Boolean(segment.starred),
        polyline: segment.polyline ?? undefined,
        athlete_kom_rank: segment.athlete_kom_rank ?? undefined,
        athlete_effort_count: segment.athlete_effort_count ?? undefined,
        athlete_pr_elapsed_time: segment.athlete_pr_elapsed_time ?? undefined,
        athlete_pr_date: segment.athlete_pr_date ?? undefined,
      },
      efforts,
    }
  }

  async getSegmentEfforts(id: number): Promise<SegmentEffort[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const rows = queries.getSegmentEfforts<{
      id: number
      segment_id: number
      activity_id: number
      athlete_id: number
      name: string | null
      elapsed_time: number
      moving_time: number
      start_date: string | null
      start_date_local: string | null
      distance: number
      average_watts: number | null
      average_heartrate: number | null
      max_heartrate: number | null
      pr_rank: number | null
    }>(db, id, athleteId)

    return rows.map((r) => ({
      id: r.id,
      segment_id: r.segment_id,
      activity_id: r.activity_id,
      athlete_id: r.athlete_id,
      name: r.name ?? undefined,
      elapsed_time: r.elapsed_time,
      moving_time: r.moving_time,
      start_date: r.start_date ?? undefined,
      start_date_local: r.start_date_local ?? undefined,
      distance: r.distance,
      average_watts: r.average_watts ?? undefined,
      average_heartrate: r.average_heartrate ?? undefined,
      max_heartrate: r.max_heartrate ?? undefined,
      pr_rank: r.pr_rank ?? undefined,
    }))
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
  // Photos
  // ============================================================================
  async getPhotos(filters?: PhotosFilters): Promise<PhotosListResponse> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const page = filters?.page ?? 1
    const perPage = filters?.per_page ?? 50
    const offset = (page - 1) * perPage

    // Build WHERE clause based on filters
    const conditions = ['p.athlete_id = ?']
    const params: unknown[] = [athleteId]

    if (filters?.sport_type) {
      conditions.push('a.sport_type = ?')
      params.push(filters.sport_type)
    }

    if (filters?.country) {
      conditions.push('a.location_country = ?')
      params.push(filters.country)
    }

    const whereClause = conditions.join(' AND ')

    // Get total count
    const countRow = db.queryOne<{ count: number }>(
      `SELECT COUNT(*) as count
       FROM photos p
       JOIN activities a ON a.id = p.activity_id
       WHERE ${whereClause}`,
      params
    )
    const total = countRow?.count ?? 0
    const totalPages = Math.ceil(total / perPage)

    // Get paginated photos
    const rows = db.query<{
      id: string
      activity_id: number
      url: string
      thumbnail_url: string | null
      caption: string | null
      created_at: string
      activity_name: string
      sport_type: string
      start_date_local: string
      location_country: string | null
    }>(
      `SELECT
        p.id,
        p.activity_id,
        p.url,
        p.thumbnail_url,
        p.caption,
        p.created_at,
        a.name as activity_name,
        a.sport_type,
        a.start_date_local,
        a.location_country
       FROM photos p
       JOIN activities a ON a.id = p.activity_id
       WHERE ${whereClause}
       ORDER BY a.start_date DESC
       LIMIT ? OFFSET ?`,
      [...params, perPage, offset]
    )

    // Get country facets
    const countryFacets = db.query<{ country: string; count: number }>(
      `SELECT a.location_country as country, COUNT(*) as count
       FROM photos p
       JOIN activities a ON a.id = p.activity_id
       WHERE p.athlete_id = ? AND a.location_country IS NOT NULL AND a.location_country != ''
       GROUP BY a.location_country
       ORDER BY count DESC`,
      [athleteId]
    )

    // Get sport type facets
    const sportFacets = db.query<{ sport_type: string; count: number }>(
      `SELECT a.sport_type, COUNT(*) as count
       FROM photos p
       JOIN activities a ON a.id = p.activity_id
       WHERE p.athlete_id = ?
       GROUP BY a.sport_type
       ORDER BY count DESC`,
      [athleteId]
    )

    return {
      data: rows.map((r) => ({
        id: r.id,
        activity_id: r.activity_id,
        url: r.url,
        thumbnail_url: r.thumbnail_url ?? undefined,
        caption: r.caption ?? undefined,
        created_at: r.created_at,
        activity_name: r.activity_name,
        sport_type: r.sport_type,
        start_date_local: r.start_date_local,
        location_country: r.location_country ?? undefined,
      })),
      total,
      page,
      per_page: perPage,
      total_pages: totalPages,
      countries: countryFacets.map((f) => ({ value: f.country, count: f.count })),
      sport_types: sportFacets.map((f) => ({ value: f.sport_type, count: f.count })),
    }
  }

  async getActivityPhotos(activityId: number): Promise<ActivityPhoto[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const rows = db.query<{
      id: string
      athlete_id: number
      activity_id: number
      url: string
      thumbnail_url: string | null
      caption: string | null
      location: string | null
      created_at: string
    }>(
      `SELECT id, athlete_id, activity_id, url, thumbnail_url, caption, location, created_at
       FROM photos
       WHERE activity_id = ? AND athlete_id = ?
       ORDER BY created_at`,
      [activityId, athleteId]
    )

    return rows.map((r) => ({
      id: r.id,
      athlete_id: r.athlete_id,
      activity_id: r.activity_id,
      url: r.url,
      thumbnail_url: r.thumbnail_url ?? undefined,
      caption: r.caption ?? undefined,
      location: r.location ? JSON.parse(r.location) : undefined,
      created_at: r.created_at,
    }))
  }

  // ============================================================================
  // Challenges
  // ============================================================================
  async getChallenges(month?: string): Promise<Challenge[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    let rows: Array<{
      id: string
      name: string
      slug: string | null
      badge_url: string | null
      completion_date: string | null
      month: string | null
    }>

    if (month) {
      rows = db.query(
        `SELECT id, name, slug, badge_url, completion_date, month
         FROM challenges
         WHERE athlete_id = ? AND month = ?
         ORDER BY completion_date DESC`,
        [athleteId, month]
      )
    } else {
      rows = db.query(
        `SELECT id, name, slug, badge_url, completion_date, month
         FROM challenges
         WHERE athlete_id = ?
         ORDER BY completion_date DESC`,
        [athleteId]
      )
    }

    return rows.map((r) => ({
      id: r.id,
      name: r.name,
      slug: r.slug ?? undefined,
      badge_url: r.badge_url ?? undefined,
      completion_date: r.completion_date ?? undefined,
      month: r.month ?? undefined,
    }))
  }

  async importChallenges(file: File): Promise<{ imported: number }> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    const text = await file.text()
    let challenges: Array<{
      id?: string
      name: string
      slug?: string
      badge_url?: string
      completion_date?: string
      month?: string
    }>

    try {
      challenges = JSON.parse(text)
      if (!Array.isArray(challenges)) {
        throw new Error('Expected an array of challenges')
      }
    } catch {
      throw new Error('Invalid JSON file')
    }

    let imported = 0
    for (const challenge of challenges) {
      if (!challenge.name) continue

      const id = challenge.id || `${athleteId}-${challenge.name}-${challenge.completion_date || Date.now()}`
      const month = challenge.month || (challenge.completion_date ? challenge.completion_date.slice(0, 7) : null)

      db.exec(
        `INSERT INTO challenges (id, athlete_id, name, slug, badge_url, completion_date, month)
         VALUES (?, ?, ?, ?, ?, ?, ?)
         ON CONFLICT (id) DO UPDATE SET
           name = EXCLUDED.name,
           slug = EXCLUDED.slug,
           badge_url = EXCLUDED.badge_url,
           completion_date = EXCLUDED.completion_date,
           month = EXCLUDED.month`,
        [id, athleteId, challenge.name, challenge.slug ?? null, challenge.badge_url ?? null, challenge.completion_date ?? null, month]
      )
      imported++
    }

    await db.persist()
    return { imported }
  }

  async importChallengesFromProfile(_athleteId?: string): Promise<{ imported: number }> {
    throw new Error('Importing challenges from public profile is not available in WASM mode')
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
  // Maintenance
  // ============================================================================

  private async getComponentWithRules(componentId: number): Promise<ComponentWithRules | null> {
    const db = this.assertInitialized()

    const component = db.queryOne<{
      id: number
      gear_id: string
      name: string
      image_url: string | null
      maintenance_hashtag: string | null
      created_at: string
      updated_at: string
    }>(
      'SELECT id, gear_id, name, image_url, maintenance_hashtag, created_at, updated_at FROM components WHERE id = ?',
      [componentId]
    )

    if (!component) return null

    const rules = db.query<{
      id: number
      component_id: number
      type: string
      threshold_value: number
      created_at: string
      updated_at: string
    }>('SELECT id, component_id, type, threshold_value, created_at, updated_at FROM maintenance_rules WHERE component_id = ?', [
      componentId,
    ])

    const lastLog = db.queryOne<{ completed_at: string }>(
      'SELECT completed_at FROM maintenance_log WHERE component_id = ? ORDER BY completed_at DESC LIMIT 1',
      [componentId]
    )

    return {
      id: component.id,
      gear_id: component.gear_id,
      name: component.name,
      image_url: component.image_url ?? undefined,
      maintenance_hashtag: component.maintenance_hashtag ?? undefined,
      created_at: component.created_at,
      updated_at: component.updated_at,
      last_completed_at: lastLog?.completed_at,
      rules: rules.map((r) => ({
        id: r.id,
        component_id: r.component_id,
        type: r.type as 'distance_m' | 'time_s' | 'days',
        threshold_value: r.threshold_value,
        created_at: r.created_at,
        updated_at: r.updated_at,
      })),
    }
  }

  async getMaintenanceDue(): Promise<DueComponent[]> {
    const db = this.assertInitialized()
    const athleteId = this.getAthleteId()

    // Get all components for this athlete's gear
    const components = db.query<{
      id: number
      gear_id: string
      name: string
      image_url: string | null
      maintenance_hashtag: string | null
      created_at: string
      updated_at: string
    }>(
      `SELECT c.id, c.gear_id, c.name, c.image_url, c.maintenance_hashtag, c.created_at, c.updated_at
       FROM components c
       JOIN gear g ON g.id = c.gear_id
       WHERE g.athlete_id = ?`,
      [athleteId]
    )

    const dueComponents: DueComponent[] = []

    for (const comp of components) {
      // Get rules for this component
      const rules = db.query<{
        id: number
        component_id: number
        type: string
        threshold_value: number
        created_at: string
        updated_at: string
      }>('SELECT id, component_id, type, threshold_value, created_at, updated_at FROM maintenance_rules WHERE component_id = ?', [
        comp.id,
      ])

      if (rules.length === 0) continue

      // Get last maintenance date
      const lastLog = db.queryOne<{ completed_at: string }>(
        'SELECT completed_at FROM maintenance_log WHERE component_id = ? ORDER BY completed_at DESC LIMIT 1',
        [comp.id]
      )
      const lastCompletedAt = lastLog?.completed_at

      // Calculate stats since last maintenance
      let distanceSince = 0
      let movingTimeSince = 0
      let daysSince = 0

      if (lastCompletedAt) {
        // Get activity stats since last maintenance for this gear
        const stats = db.queryOne<{ distance: number; moving_time: number }>(
          `SELECT COALESCE(SUM(distance), 0) as distance, COALESCE(SUM(moving_time), 0) as moving_time
           FROM activities
           WHERE athlete_id = ? AND gear_id = ? AND start_date > ?`,
          [athleteId, comp.gear_id, lastCompletedAt]
        )
        distanceSince = stats?.distance || 0
        movingTimeSince = stats?.moving_time || 0
        daysSince = Math.floor((Date.now() - new Date(lastCompletedAt).getTime()) / (1000 * 60 * 60 * 24))
      } else {
        // No maintenance logged - calculate from all time for this gear
        const stats = db.queryOne<{ distance: number; moving_time: number; first_date: string | null }>(
          `SELECT COALESCE(SUM(distance), 0) as distance, COALESCE(SUM(moving_time), 0) as moving_time, MIN(start_date) as first_date
           FROM activities
           WHERE athlete_id = ? AND gear_id = ?`,
          [athleteId, comp.gear_id]
        )
        distanceSince = stats?.distance || 0
        movingTimeSince = stats?.moving_time || 0
        if (stats?.first_date) {
          daysSince = Math.floor((Date.now() - new Date(stats.first_date).getTime()) / (1000 * 60 * 60 * 24))
        }
      }

      // Calculate progress for each rule
      const progress = rules.map((rule) => {
        let currentValue = 0
        if (rule.type === 'distance_m') currentValue = distanceSince
        else if (rule.type === 'time_s') currentValue = movingTimeSince
        else if (rule.type === 'days') currentValue = daysSince

        const percent = rule.threshold_value > 0 ? Math.min(100, (currentValue / rule.threshold_value) * 100) : 0

        return {
          type: rule.type as 'distance_m' | 'time_s' | 'days',
          threshold_value: rule.threshold_value,
          current_value: currentValue,
          percent,
          due: percent >= 100,
        }
      })

      const isDue = progress.some((p) => p.due)

      dueComponents.push({
        id: comp.id,
        gear_id: comp.gear_id,
        name: comp.name,
        image_url: comp.image_url ?? undefined,
        maintenance_hashtag: comp.maintenance_hashtag ?? undefined,
        created_at: comp.created_at,
        updated_at: comp.updated_at,
        last_completed_at: lastCompletedAt,
        rules: rules.map((r) => ({
          id: r.id,
          component_id: r.component_id,
          type: r.type as 'distance_m' | 'time_s' | 'days',
          threshold_value: r.threshold_value,
          created_at: r.created_at,
          updated_at: r.updated_at,
        })),
        distance_since: distanceSince,
        moving_time_since: movingTimeSince,
        days_since: daysSince,
        progress,
        is_due: isDue,
      })
    }

    // Return only due components, sorted by most overdue first
    return dueComponents
      .filter((c) => c.is_due)
      .sort((a, b) => {
        const aMax = Math.max(...a.progress.map((p) => p.percent))
        const bMax = Math.max(...b.progress.map((p) => p.percent))
        return bMax - aMax
      })
  }

  async getGearComponents(gearId: string): Promise<ComponentWithRules[]> {
    const db = this.assertInitialized()

    const components = db.query<{ id: number }>(
      'SELECT id FROM components WHERE gear_id = ?',
      [gearId]
    )

    const result: ComponentWithRules[] = []
    for (const comp of components) {
      const withRules = await this.getComponentWithRules(comp.id)
      if (withRules) result.push(withRules)
    }

    return result
  }

  async createComponent(gearId: string, req: CreateComponentRequest): Promise<ComponentWithRules> {
    const db = this.assertInitialized()
    const now = new Date().toISOString()

    // Insert component
    db.exec(
      `INSERT INTO components (gear_id, name, image_url, maintenance_hashtag, created_at, updated_at)
       VALUES (?, ?, ?, ?, ?, ?)`,
      [gearId, req.name, req.image_url ?? null, req.maintenance_hashtag ?? null, now, now]
    )

    const result = db.queryOne<{ id: number }>('SELECT last_insert_rowid() as id')
    const componentId = result!.id

    // Insert rules
    if (req.rules && req.rules.length > 0) {
      for (const rule of req.rules) {
        db.exec(
          `INSERT INTO maintenance_rules (component_id, type, threshold_value, created_at, updated_at)
           VALUES (?, ?, ?, ?, ?)`,
          [componentId, rule.type, rule.threshold_value, now, now]
        )
      }
    }

    await db.persist()

    const component = await this.getComponentWithRules(componentId)
    if (!component) throw new Error('Failed to create component')
    return component
  }

  async updateComponent(id: number, req: UpdateComponentRequest): Promise<ComponentWithRules> {
    const db = this.assertInitialized()
    const now = new Date().toISOString()

    // Update component fields
    const updates: string[] = ['updated_at = ?']
    const params: unknown[] = [now]

    if (req.name !== undefined) {
      updates.push('name = ?')
      params.push(req.name)
    }
    if (req.image_url !== undefined) {
      updates.push('image_url = ?')
      params.push(req.image_url)
    }
    if (req.maintenance_hashtag !== undefined) {
      updates.push('maintenance_hashtag = ?')
      params.push(req.maintenance_hashtag)
    }

    params.push(id)
    db.exec(`UPDATE components SET ${updates.join(', ')} WHERE id = ?`, params)

    // Replace rules if provided
    if (req.rules !== undefined) {
      db.exec('DELETE FROM maintenance_rules WHERE component_id = ?', [id])
      for (const rule of req.rules) {
        db.exec(
          `INSERT INTO maintenance_rules (component_id, type, threshold_value, created_at, updated_at)
           VALUES (?, ?, ?, ?, ?)`,
          [id, rule.type, rule.threshold_value, now, now]
        )
      }
    }

    await db.persist()

    const component = await this.getComponentWithRules(id)
    if (!component) throw new Error('Component not found')
    return component
  }

  async deleteComponent(id: number): Promise<{ deleted: boolean }> {
    const db = this.assertInitialized()

    // Delete rules and logs first (foreign key)
    db.exec('DELETE FROM maintenance_rules WHERE component_id = ?', [id])
    db.exec('DELETE FROM maintenance_log WHERE component_id = ?', [id])
    db.exec('DELETE FROM components WHERE id = ?', [id])

    await db.persist()
    return { deleted: true }
  }

  async logMaintenance(componentId: number, req?: LogMaintenanceRequest): Promise<{ logged: boolean }> {
    const db = this.assertInitialized()

    const completedAt = req?.completed_at || new Date().toISOString()
    const activityId = req?.activity_id ?? null

    db.exec(
      `INSERT INTO maintenance_log (component_id, activity_id, completed_at, created_at)
       VALUES (?, ?, ?, datetime('now'))
       ON CONFLICT (component_id, completed_at) DO NOTHING`,
      [componentId, activityId, completedAt]
    )

    await db.persist()
    return { logged: true }
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

    // Map phase names (WASM uses 'details'/'segments', API uses 'activity_details'/'segment_details')
    const phaseMap: Record<string, ImportProgress['phase']> = {
      idle: 'idle',
      activities: 'activities',
      gear: 'gear',
      streams: 'streams',
      details: 'activity_details',
      segments: 'segment_details',
      photos: 'photos',
      complete: 'completed',
    }

    const status = statusMap[progress.status] ?? 'idle'
    const phase = phaseMap[progress.phase] ?? 'idle'
    const rateLimitInfo = getRateLimitInfo()

    return {
      status,
      phase,

      // Per-phase progress (direct mapping from WASM importer)
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
      error: progress.error,

      // Rate limit info from WASM rate limiter
      remaining_api_calls: 0,
      estimated_eta: undefined,
      rate_limit_used_15min: rateLimitInfo.usage15Min,
      rate_limit_limit_15min: rateLimitInfo.limit15Min,
      rate_limit_used_daily: rateLimitInfo.usageDaily,
      rate_limit_limit_daily: rateLimitInfo.limitDaily,

      // Rate limit waiting state
      waiting_for_rate_limit: rateLimitInfo.waitingForRateLimit,
      waiting_until: rateLimitInfo.waitingUntil,
      waiting_reason: rateLimitInfo.waitingReason,
    }
  }

  async startImport(req?: StartImportRequest): Promise<{ message: string }> {
    // Start import in background (non-blocking)
    stravaStartImport({
      fullSync: req?.full_sync,
      skipStreams: req?.skip_streams,
      skipSegments: req?.skip_segments,
      skipPhotos: req?.skip_photos,
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
    const latestSync = await this.getLatestSync()
    if (!latestSync || latestSync.status !== 'completed') {
      return null
    }

    return {
      last_synced_at: latestSync.completed_at || latestSync.started_at,
      newest_activity_date: latestSync.newest_activity_date,
    }
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
