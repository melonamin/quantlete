/**
 * Go WASM Storage Bridge
 *
 * Thin wrapper around Go WASM storage layer. All database operations
 * are handled by Go code compiled to WASM, with sql.js providing the
 * SQLite engine in the browser.
 *
 * This module shares the same underlying sql.js database with the
 * existing WasmDatabase, allowing both TypeScript and Go code to
 * access the same data.
 */

import { initializeDatabase, getDatabase, type WasmDatabaseOptions } from './db'
import { stravaFetch } from './strava'
import type { GoStorageInterface } from './go-storage.gen'
import {
  setInitialized as setWrappersInitialized,
  parseGoResult,
  callGoStorage,
  callGoStorageArray,
  callGoStorageValue,
  callGoStorageVoid,
  type GoStorageResult,
} from './go-storage-wrappers.gen'

// Import generated types - these are the source of truth
import type {
  DashboardStats,
  WeeklyStat,
  RecentActivity,
  SportTypeStat,
  MonthlyStat,
  YearlyStatOutput,
  CalendarDay,
  CalendarActivity,
  CalendarMonthSummary,
  HeatmapActivity,
  HeatmapCountryStat,
  DistributionSlice,
  EddingtonStep,
  EddingtonResult,
  EddingtonHistoryPoint,
  DashboardConfig as GenDashboardConfig,
} from './types.gen'

// Import Insight types from the canonical source
import type {
  InsightType,
  InsightSeverity,
  Insight,
  InsightsResponse,
} from '@/lib/api/stats'

// Re-export for consumers that import from this module
export type { InsightType, InsightSeverity, Insight, InsightsResponse }

// Re-export for consumers
export type { GoStorageResult }
export { parseGoResult }

// Type alias for backwards compatibility
export type YearlyStat = YearlyStatOutput

// Re-export generated types that match the API
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
  HeatmapCountryStat,
  DistributionSlice,
  EddingtonStep,
  EddingtonResult,
  EddingtonHistoryPoint,
}

// Namespace for Go WASM interop - prevents direct global pollution.
// This must match goWasmNamespace in cmd/wasm/import_strava_adapter.go.
const GO_WASM_NAMESPACE = '__quantlete_go_wasm__' as const

// Declare global Go WASM types
declare global {
  interface Window {
    _go_sqlite: SqlJsStatic
    _go_sqlite_dbs: Map<string, SqlJsDatabase>
    // Namespaced access for Go WASM - prevents XSS from easily intercepting
    [GO_WASM_NAMESPACE]: {
      stravaFetch: typeof stravaFetch
    }
  }

  // sql.js types
  interface SqlJsStatic {
    Database: new (data?: Uint8Array | string) => SqlJsDatabase
  }

  interface SqlJsDatabase {
    export(): Uint8Array
    close(): void
  }

  // Go runtime
  class Go {
    importObject: WebAssembly.Imports
    run(instance: WebAssembly.Instance): Promise<void>
  }

  // goStorage namespace registered by Go WASM - uses generated interface
  const goStorage: GoStorageInterface
}

let initialized = false

// ============================================================================
// WASM Connection Lifecycle
// ============================================================================

/**
 * Load wasm_exec.js (Go WebAssembly runtime)
 */
async function loadGoRuntime(): Promise<void> {
  if (typeof (window as unknown as { Go?: unknown }).Go !== 'undefined') {
    return
  }

  return new Promise((resolve, reject) => {
    const script = document.createElement('script')
    script.src = '/wasm/wasm_exec.js'
    script.onload = () => resolve()
    script.onerror = () => reject(new Error('Failed to load wasm_exec.js'))
    document.head.appendChild(script)
  })
}

interface GoInstance {
  importObject: WebAssembly.Imports
  run(instance: WebAssembly.Instance): Promise<void>
}

/**
 * Load Go WASM binary
 */
async function loadGoWasm(): Promise<void> {
  await loadGoRuntime()

  const Go = (window as unknown as { Go: new () => GoInstance }).Go
  const go = new Go()
  const result = await WebAssembly.instantiateStreaming(
    fetch('/wasm/quantlete.wasm'),
    go.importObject
  )
  go.run(result.instance)
}

/**
 * Initialize Go WASM storage layer
 */
export async function initGoStorage(options?: WasmDatabaseOptions): Promise<void> {
  if (initialized) return

  const wasmDb = await initializeDatabase(options)

  const sqlJs = wasmDb.getSqlJs()
  const internalDb = wasmDb.getInternalDb()

  if (!sqlJs || !internalDb) {
    throw new Error('Failed to get sql.js instances from WasmDatabase')
  }

  const OriginalDatabase = sqlJs.Database
  const wrappedSqlJs = {
    ...sqlJs,
    Database: function (data?: Uint8Array | string) {
      if (data === ':memory:' || data === 'file::memory:' || data === '') {
        return internalDb
      }
      return new OriginalDatabase(data as Uint8Array)
    } as unknown as typeof sqlJs.Database,
  }
  wrappedSqlJs.Database.prototype = OriginalDatabase.prototype

  window._go_sqlite = wrappedSqlJs as SqlJsStatic
  window._go_sqlite_dbs = new Map()
  window._go_sqlite_dbs.set(':memory:', internalDb as unknown as SqlJsDatabase)

  // Expose stravaFetch via namespaced, non-configurable object for Go WASM to call.
  // This prevents easy XSS interception compared to a simple window property.
  Object.defineProperty(window, GO_WASM_NAMESPACE, {
    value: Object.freeze({
      stravaFetch,
    }),
    writable: false,
    configurable: false,
    enumerable: false,
  })

  await loadGoWasm()

  await new Promise((resolve) => setTimeout(resolve, 100))

  const result = parseGoResult(goStorage.init(), 'init')
  if (!result.ok) {
    throw new Error(`Go storage init failed: ${result.error}`)
  }

  initialized = true
  setWrappersInitialized(true)
}

export function isInitialized(): boolean {
  return initialized
}

export async function persistDatabase(): Promise<void> {
  const db = getDatabase()
  await db.persist()
}

export async function clearDatabase(): Promise<void> {
  const db = getDatabase()
  await db.reset()
}

export function setAthleteId(id: number): void {
  callGoStorageVoid(() => goStorage.setAthleteId(id), 'setAthleteId')
}

// ============================================================================
// Auth
// ============================================================================

export interface AuthStatusResult {
  authenticated: boolean
  athlete?: {
    id: number
    username: string
    firstname: string
    lastname: string
    profile: string
  }
}

export function getAuthStatus(): AuthStatusResult {
  if (!initialized) {
    return { authenticated: false }
  }
  return parseGoResult<AuthStatusResult>(goStorage.getAuthStatus())
}

export interface FirstAthleteResult {
  id: number
  firstname: string
  lastname: string
  username?: string
  profile?: string
}

/**
 * Get the first athlete in the database.
 * Used for demo mode initialization to identify the demo user.
 */
export function getFirstAthlete(): FirstAthleteResult | null {
  if (!initialized) {
    return null
  }
  return callGoStorage<FirstAthleteResult | null>(
    () => goStorage.getFirstAthlete(),
    'getFirstAthlete'
  )
}

// ============================================================================
// Activities
// ============================================================================

export interface ActivityFilters {
  page?: number
  per_page?: number
  sport_type?: string
  after?: string
  before?: string
  gear_id?: string
  commute?: boolean
  trainer?: boolean
  search?: string
  order_by?: string
  order_dir?: string
}

export interface ActivitiesResult {
  data: Activity[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface Activity {
  id: number
  athlete_id: number
  name: string
  sport_type: string
  start_date: string
  start_date_local: string
  timezone: string
  distance: number
  moving_time: number
  elapsed_time: number
  total_elevation_gain: number
  average_speed: number
  max_speed: number
  kudos_count: number
  comment_count: number
  photo_count: number
  commute: boolean
  private: boolean
  trainer: boolean
  gear_id: string
  summary_polyline: string
  description?: string
  location_city?: string
  location_state?: string
  location_country?: string
  average_heartrate?: number
  max_heartrate?: number
  average_watts?: number
  max_watts?: number
  weighted_average_watts?: number
  kilojoules?: number
  average_cadence?: number
  calories?: number
  workout_type?: number
  start_lat?: number
  start_lng?: number
}

export const getActivities = (filters: ActivityFilters): ActivitiesResult => {
  // Convert date strings to RFC3339 for Go time.Time parsing
  const goFilters = {
    ...filters,
    after: filters.after ? `${filters.after}T00:00:00Z` : undefined,
    before: filters.before ? `${filters.before}T00:00:00Z` : undefined,
  }
  return callGoStorage<ActivitiesResult>(
    () => goStorage.getActivities(JSON.stringify(goFilters)),
    'getActivities'
  )
}

export const getActivity = (id: number): Activity =>
  callGoStorage<Activity>(
    () => goStorage.getActivity(JSON.stringify({ activity_id: id })),
    'getActivity'
  )

export interface ActivityStream {
  stream_type: string
  data: number[]
  resolution: string
  original_size: number
  series_type: string
}

export const getActivityStreams = (activityId: number): ActivityStream[] =>
  callGoStorageArray<ActivityStream>(
    () => goStorage.getActivityStreams(JSON.stringify({ activity_id: activityId })),
    'getActivityStreams'
  )

// ============================================================================
// Dashboard
// ============================================================================

// DashboardWidgetConfig with stricter width types for API use
export interface DashboardWidgetConfig {
  id: string
  width: number
  height?: number
  hidden?: boolean
  settings?: Record<string, unknown>
}

export interface DashboardConfig {
  version: number
  widgets: DashboardWidgetConfig[]
}

export function getDashboardConfig(): DashboardConfig {
  const cfg = callGoStorage<GenDashboardConfig | undefined>(
    () => goStorage.getDashboardConfig(),
    'getDashboardConfig'
  ) ?? { version: 1, widgets: [] }
  return { version: cfg.version, widgets: cfg.widgets ?? [] }
}

export function updateDashboardConfig(config: DashboardConfig): DashboardConfig {
  const cfg = callGoStorage<GenDashboardConfig | undefined>(
    () => goStorage.updateDashboardConfig(JSON.stringify(config)),
    'updateDashboardConfig'
  ) ?? { version: config.version, widgets: config.widgets }
  return { version: cfg.version, widgets: cfg.widgets ?? [] }
}

export const getDashboardStats = (): DashboardStats =>
  callGoStorage<DashboardStats>(() => goStorage.getDashboardStats(), 'getDashboardStats')

export const getWeeklyStats = (): WeeklyStat[] =>
  callGoStorageArray<WeeklyStat>(() => goStorage.getWeeklyStats(), 'getWeeklyStats')

export const getRecentActivities = (limit: number): RecentActivity[] =>
  callGoStorageArray<RecentActivity>(
    () => goStorage.getRecentActivities(JSON.stringify({ limit })),
    'getRecentActivities'
  )

export const getSportTypeStats = (): SportTypeStat[] =>
  callGoStorageArray<SportTypeStat>(() => goStorage.getSportTypeStats(), 'getSportTypeStats')

export const getMonthlyStats = (year?: number): MonthlyStat[] =>
  callGoStorageArray<MonthlyStat>(
    () => goStorage.getMonthlyStats(JSON.stringify({ year })),
    'getMonthlyStats'
  )

export const getYearlyStats = (): YearlyStat[] =>
  callGoStorageArray<YearlyStat>(() => goStorage.getYearlyStats(), 'getYearlyStats')

// ============================================================================
// Distribution Stats
// ============================================================================

export const getDaytimeDistribution = (): DistributionSlice[] =>
  callGoStorageArray<DistributionSlice>(
    () => goStorage.getDaytimeDistribution(),
    'getDaytimeDistribution'
  )

export const getWeekdayDistribution = (): DistributionSlice[] =>
  callGoStorageArray<DistributionSlice>(
    () => goStorage.getWeekdayDistribution(),
    'getWeekdayDistribution'
  )

export interface ExportStats {
  total_activities: number
  first_activity: string | null
  last_activity: string | null
}

export function getExportStats(): ExportStats {
  return (
    callGoStorage<ExportStats | undefined>(() => goStorage.getExportStats(), 'getExportStats') ?? {
      total_activities: 0,
      first_activity: null,
      last_activity: null,
    }
  )
}

// ============================================================================
// Heatmap
// ============================================================================

export interface HeatmapFilters {
  sport_type?: string
  year?: number
  commute?: boolean
  workout_type?: number
  limit?: number
  offset?: number
}

// Response type for heatmap data that includes both activities and countries
export interface HeatmapDataResponse {
  activities: HeatmapActivity[]
  countries: HeatmapCountryStat[]
}

export function getHeatmapData(filters: HeatmapFilters): HeatmapDataResponse {
  return (
    callGoStorage<HeatmapDataResponse | undefined>(
      () => goStorage.getHeatmapData(JSON.stringify(filters)),
      'getHeatmapData'
    ) ?? { activities: [], countries: [] }
  )
}

// ============================================================================
// Calendar
// ============================================================================

export function getCalendarData(year: number): CalendarDay[] {
  return (
    callGoStorage<CalendarDay[] | undefined>(
      () => goStorage.getCalendarData(JSON.stringify({ year })),
      'getCalendarData'
    ) || []
  )
}

// CalendarActivityResult extends CalendarActivity with the same fields
export type CalendarActivityResult = CalendarActivity

export function getCalendarActivities(year: number, month: number): CalendarActivityResult[] {
  const startDate = `${year}-${String(month).padStart(2, '0')}-01`
  const endMonth = month === 12 ? 1 : month + 1
  const endYear = month === 12 ? year + 1 : year
  const endDate = `${endYear}-${String(endMonth).padStart(2, '0')}-01`

  const result = getActivities({ after: startDate, before: endDate, per_page: 1000 })
  return result.data.map((a) => ({
    id: a.id,
    name: a.name,
    sport_type: a.sport_type,
    start_date: a.start_date,
    distance: a.distance,
    moving_time: a.moving_time,
    total_elevation_gain: a.total_elevation_gain,
  }))
}

export interface CalendarSummaryResult {
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

export function getCalendarSummary(year: number, month: number): CalendarSummaryResult {
  const activities = getCalendarActivities(year, month)
  return {
    year,
    month,
    activity_count: activities.length,
    total_distance: activities.reduce((sum, a) => sum + a.distance, 0),
    total_elevation_gain: activities.reduce((sum, a) => sum + a.total_elevation_gain, 0),
    total_moving_time: activities.reduce((sum, a) => sum + a.moving_time, 0),
    total_calories: 0,
    workout_count: 0,
    challenges_completed: 0,
  }
}

// ============================================================================
// Gear - Read
// ============================================================================

export interface GearFilters {
  include_retired?: boolean
  page?: number
  per_page?: number
  order_by?: string
  order_dir?: string
}

export interface GearItem {
  id: string
  athlete_id: number
  name: string
  primary: boolean
  retired: boolean
  distance: number
  brand_name?: string
  model_name?: string
  description?: string
  source: string
  hashtag?: string
  purchase_price?: number
  purchase_currency?: string
  activity_count: number
}

export interface GearResult {
  data: GearItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export const getGear = (filters?: GearFilters): GearResult =>
  callGoStorage<GearResult>(() => goStorage.getGear(JSON.stringify(filters || {})), 'getGear')

export function getGearDetail(id: string): GearItem {
  return callGoStorage<GearItem>(
    () => goStorage.getGearDetail(JSON.stringify({ gear_id: id })),
    'getGearDetail'
  )
}

// ============================================================================
// Segments - Read
// ============================================================================

export interface SegmentsFilters {
  activity_type?: string
  country?: string
  starred?: boolean
  kom_only?: boolean
  search?: string
  page?: number
  per_page?: number
  order_by?: string
  order_dir?: string
}

export interface SegmentListItem {
  id: number
  name: string
  activity_type: string
  distance: number
  average_grade: number
  maximum_grade: number
  elevation_high: number
  elevation_low: number
  climb_category: number
  start_lat?: number
  start_lng?: number
  end_lat?: number
  end_lng?: number
  starred: boolean
  polyline?: string
  times_completed: number
  last_effort_date?: string
  best_elapsed_time?: number
}

export interface SegmentsResult {
  data: SegmentListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export const getSegments = (filters?: SegmentsFilters): SegmentsResult =>
  callGoStorage<SegmentsResult>(
    () => goStorage.getSegments(JSON.stringify(filters || {})),
    'getSegments'
  )

export interface SegmentDetailResult {
  segment: SegmentListItem
  efforts: SegmentEffort[]
}

export const getSegmentDetail = (id: number): SegmentDetailResult =>
  callGoStorage<SegmentDetailResult>(
    () => goStorage.getSegmentDetail(JSON.stringify({ segment_id: id })),
    'getSegmentDetail'
  )

// ============================================================================
// Photos - Read
// ============================================================================

export interface PhotosFilters {
  sport_type?: string
  country?: string
  page?: number
  per_page?: number
}

export interface PhotoListItem {
  id: string
  activity_id: number
  url: string
  thumbnail_url?: string
  caption?: string
  created_at: string
  activity_name: string
  sport_type: string
  start_date_local: string
  location_country?: string
}

export interface PhotosFacet {
  value: string
  iso2?: string
  count: number
}

export interface PhotosResult {
  data: PhotoListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
  countries: PhotosFacet[]
  sport_types: PhotosFacet[]
}

export const getPhotos = (filters?: PhotosFilters): PhotosResult =>
  callGoStorage<PhotosResult>(() => goStorage.getPhotos(JSON.stringify(filters || {})), 'getPhotos')

export interface ActivityPhotoResult {
  id: string
  athlete_id: number
  activity_id: number
  url: string
  thumbnail_url?: string
  caption?: string
  location?: unknown
  created_at: string
}

export function getActivityPhotos(activityId: number): ActivityPhotoResult[] {
  return (
    callGoStorage<ActivityPhotoResult[] | undefined>(
      () => goStorage.getActivityPhotos(JSON.stringify({ activity_id: activityId })),
      'getActivityPhotos'
    ) || []
  )
}

// ============================================================================
// Best Efforts - Read
// ============================================================================

export interface BestEffortPRResult {
  distance_type: string
  name: string
  distance_m: number
  elapsed_time_s: number
  moving_time_s?: number
  pr_rank?: number
  activity_id: number
  activity_name: string
  sport_type: string
  start_date_local: string
}

export function getBestEffortPRs(sportType?: string): BestEffortPRResult[] {
  return (
    callGoStorage<BestEffortPRResult[] | undefined>(
      () => goStorage.getBestEffortPRs(JSON.stringify({ sport_type: sportType })),
      'getBestEffortPRs'
    ) || []
  )
}

export interface BestEffortItemResult extends BestEffortPRResult {
  start_index?: number
  end_index?: number
}

export function getBestEffortsForType(
  distanceType: string,
  sportType?: string
): BestEffortItemResult[] {
  return (
    callGoStorage<BestEffortItemResult[] | undefined>(
      () =>
        goStorage.getBestEffortsForType(
          JSON.stringify({ distance_type: distanceType, sport_type: sportType })
        ),
      'getBestEffortsForType'
    ) || []
  )
}

// ============================================================================
// Write Operations
// ============================================================================

export interface SaveActivityInput {
  id: number
  athlete_id: number
  name: string
  sport_type: string
  start_date: string
  start_date_local: string
  timezone: string
  distance: number
  moving_time: number
  elapsed_time: number
  total_elevation_gain: number
  average_speed: number
  max_speed: number
  average_heartrate?: number | null
  max_heartrate?: number | null
  average_watts?: number | null
  max_watts?: number | null
  weighted_average_watts?: number | null
  kilojoules?: number | null
  average_cadence?: number | null
  calories?: number | null
  gear_id?: string | null
  commute?: boolean
  workout_type?: number | null
  location_city?: string | null
  location_state?: string | null
  location_country?: string | null
  summary_polyline?: string | null
  start_lat?: number | null
  start_lng?: number | null
  description?: string | null
  device_name?: string | null
  trainer?: boolean
  private?: boolean
  kudos_count?: number
  photo_count?: number
}

export const saveActivity = (activity: SaveActivityInput): void =>
  callGoStorageVoid(() => goStorage.saveActivity(JSON.stringify(activity)), 'saveActivity')

export interface SaveStreamInput {
  activity_id: number
  stream_type: string
  data: unknown[]
  series_type: string
  original_size: number
  resolution: string
}

export const saveStream = (stream: SaveStreamInput): void =>
  callGoStorageVoid(() => goStorage.saveStream(JSON.stringify(stream)), 'saveStream')

export interface SaveAthleteInput {
  id: number
  username: string
  firstname: string
  lastname: string
  profile_medium?: string
  profile?: string
  city?: string
  state?: string
  country?: string
  sex?: string
  premium?: boolean
  summit?: boolean
}

export const saveAthlete = (athlete: SaveAthleteInput): void =>
  callGoStorageVoid(() => goStorage.saveAthlete(JSON.stringify(athlete)), 'saveAthlete')

export interface SaveGearInput {
  id: string
  athlete_id: number
  name: string
  primary?: boolean
  retired?: boolean
  distance?: number
  brand_name?: string | null
  model_name?: string | null
  description?: string | null
}

export const saveGear = (gear: SaveGearInput): void =>
  callGoStorageVoid(() => goStorage.saveGear(JSON.stringify(gear)), 'saveGear')

export interface SaveSegmentInput {
  id: number
  name: string
  activity_type: string
  distance: number
  average_grade: number
  maximum_grade: number
  elevation_high: number
  elevation_low: number
  climb_category: number
  start_lat?: number | null
  start_lng?: number | null
  end_lat?: number | null
  end_lng?: number | null
  starred?: boolean
  polyline?: string | null
  athlete_kom_rank?: number | null
  athlete_effort_count?: number | null
  athlete_pr_elapsed_time?: number | null
  athlete_pr_date?: string | null
}

export const saveSegment = (segment: SaveSegmentInput): void =>
  callGoStorageVoid(() => goStorage.saveSegment(JSON.stringify(segment)), 'saveSegment')

export interface SaveSegmentEffortInput {
  id: number
  segment_id: number
  activity_id: number
  athlete_id: number
  name: string
  elapsed_time: number
  moving_time: number
  start_date?: string
  start_date_local?: string
  distance: number
  average_watts?: number | null
  average_heartrate?: number | null
  max_heartrate?: number | null
  pr_rank?: number | null
}

export const saveSegmentEffort = (effort: SaveSegmentEffortInput): void =>
  callGoStorageVoid(() => goStorage.saveSegmentEffort(JSON.stringify(effort)), 'saveSegmentEffort')

export interface BestEffortInput {
  name: string
  distance_m: number
  elapsed_time: number
  moving_time?: number | null
  start_index?: number | null
  end_index?: number | null
  pr_rank?: number | null
  start_date?: string
}

export interface SaveBestEffortsInput {
  athlete_id: number
  activity_id: number
  sport_type?: string | null
  efforts: BestEffortInput[]
}

export const saveBestEfforts = (data: SaveBestEffortsInput): void =>
  callGoStorageVoid(() => goStorage.saveBestEfforts(JSON.stringify(data)), 'saveBestEfforts')

export interface SavePhotoInput {
  id: string
  athlete_id: number
  activity_id: number
  url: string
  thumbnail_url?: string
  caption?: string | null
  location?: string | null
  created_at?: string
}

export const savePhoto = (photo: SavePhotoInput): void =>
  callGoStorageVoid(() => goStorage.savePhoto(JSON.stringify(photo)), 'savePhoto')

// ============================================================================
// Power Best Efforts
// ============================================================================

export interface ComputePowerBestEffortsInput {
  activity_id: number
  athlete_id: number
}

export const computePowerBestEfforts = (input: ComputePowerBestEffortsInput): void =>
  callGoStorageVoid(
    () => goStorage.computePowerBestEfforts(JSON.stringify(input)),
    'computePowerBestEfforts'
  )

// ============================================================================
// Sync History
// ============================================================================

export interface CreateSyncRunInput {
  athlete_id: number
  full_sync?: boolean
  skip_streams?: boolean
  skip_segments?: boolean
  skip_best_efforts?: boolean
  skip_photos?: boolean
}

export interface CreateSyncRunResult {
  ok: boolean
  id: number
}

export function createSyncRun(input: CreateSyncRunInput): number {
  const result = callGoStorage<CreateSyncRunResult>(
    () => goStorage.createSyncRun(JSON.stringify(input)),
    'createSyncRun'
  )
  return (result as unknown as CreateSyncRunResult).id
}

export interface UpdateSyncRunInput {
  id: number
  status: 'canceled' | 'paused'
  activities_total?: number
  activities_imported?: number
  streams_imported?: number
  failed_count?: number
  newest_activity_date?: string | null
}

export const updateSyncRun = (input: UpdateSyncRunInput): void =>
  callGoStorageVoid(() => goStorage.updateSyncRun(JSON.stringify(input)), 'updateSyncRun')

export interface CompleteSyncRunInput {
  id: number
  status: 'completed' | 'failed' | 'canceled'
  error?: string
  activities_total?: number
  activities_imported?: number
  activities_skipped?: number
  gear_imported?: number
  streams_imported?: number
  segments_imported?: number
  photos_imported?: number
  failed_count?: number
  newest_activity_date?: string | null
}

export const completeSyncRun = (input: CompleteSyncRunInput): void =>
  callGoStorageVoid(() => goStorage.completeSyncRun(JSON.stringify(input)), 'completeSyncRun')

export interface SyncHistoryItem {
  id: number
  athlete_id: number
  started_at: string
  completed_at?: string
  duration_seconds?: number
  status: string
  error?: string
  activities_total: number
  activities_imported: number
  activities_skipped: number
  streams_imported: number
  failed_count: number
  full_sync: boolean
  skip_streams: boolean
  newest_activity_date?: string
}

export function getSyncHistory(limit = 10): SyncHistoryItem[] {
  return (
    callGoStorage<SyncHistoryItem[] | undefined>(
      () => goStorage.getSyncHistory(limit),
      'getSyncHistory'
    ) ?? []
  )
}

// ============================================================================
// Algorithms - Power
// ============================================================================

export const normalizedPower = (watts: number[]): number =>
  callGoStorageValue<number>(() => goStorage.normalizedPower(watts), 'normalizedPower')

export const rollingMaxAverage = (values: number[], windowSeconds: number): number =>
  callGoStorageValue<number>(
    () => goStorage.rollingMaxAverage(values, windowSeconds),
    'rollingMaxAverage'
  )

export const intensityFactor = (np: number, ftp: number): number =>
  callGoStorageValue<number>(() => goStorage.intensityFactor(np, ftp), 'intensityFactor')

export const trainingStressScore = (durationSeconds: number, np: number, ftp: number): number =>
  callGoStorageValue<number>(
    () => goStorage.trainingStressScore(durationSeconds, np, ftp),
    'trainingStressScore'
  )

// ============================================================================
// Algorithms - Eddington
// ============================================================================

export const eddingtonNumber = (distances: number[]): number =>
  callGoStorageValue<number>(() => goStorage.eddingtonNumber(distances), 'eddingtonNumber')

// Use the generated EddingtonStep type which has { target, rides_needed }
export function eddingtonNextSteps(
  distances: number[],
  currentE: number,
  stepsToCalculate: number
): EddingtonStep[] {
  return callGoStorage<EddingtonStep[]>(
    () => goStorage.eddingtonNextSteps(distances, currentE, stepsToCalculate),
    'eddingtonNextSteps'
  )
}

export function eddingtonHistory(distances: number[]): number[] {
  return callGoStorage<number[]>(() => goStorage.eddingtonHistory(distances), 'eddingtonHistory')
}

// ============================================================================
// Algorithms - Training Load
// ============================================================================

export interface TrainingLoadPoint {
  ctl: number
  atl: number
  tsb: number
}

export function calculateTrainingLoad(
  dailyTss: number[],
  ctlTau: number,
  atlTau: number
): TrainingLoadPoint[] {
  return callGoStorage<TrainingLoadPoint[]>(
    () => goStorage.calculateTrainingLoad(dailyTss, ctlTau, atlTau),
    'calculateTrainingLoad'
  )
}

export function calculateTrainingLoadWithInitial(
  dailyTss: number[],
  initialCtl: number,
  initialAtl: number,
  ctlTau: number,
  atlTau: number
): TrainingLoadPoint[] {
  return callGoStorage<TrainingLoadPoint[]>(
    () =>
      goStorage.calculateTrainingLoadWithInitial(dailyTss, initialCtl, initialAtl, ctlTau, atlTau),
    'calculateTrainingLoadWithInitial'
  )
}

export function predictAfterWorkout(
  currentCtl: number,
  currentAtl: number,
  plannedTss: number,
  ctlTau: number,
  atlTau: number
): TrainingLoadPoint {
  const result = callGoStorage<TrainingLoadPoint>(
    () => goStorage.predictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau),
    'predictAfterWorkout'
  )
  return { ctl: result.ctl, atl: result.atl, tsb: result.tsb } as TrainingLoadPoint
}

export const tssForTargetTsb = (
  currentCtl: number,
  currentAtl: number,
  targetTsb: number,
  ctlTau: number,
  atlTau: number
): number =>
  callGoStorageValue<number>(
    () => goStorage.tssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau),
    'tssForTargetTsb'
  )

// ============================================================================
// Eddington Data
// ============================================================================

export interface EddingtonFilters {
  sport_types?: string[]
}

// EddingtonHistoryPoint imported from types.gen.ts

export interface EddingtonDistributionDay {
  date: string
  distance: number
}

// EddingtonDataResult matches EddingtonResult from generated types with additional distribution field
export interface EddingtonDataResult {
  number: number
  history: EddingtonHistoryPoint[]
  distribution: EddingtonDistributionDay[]
  next_steps: EddingtonStep[]
}

export function getEddingtonData(filters?: EddingtonFilters): EddingtonDataResult {
  const result = callGoStorage<EddingtonDataResult>(
    () => goStorage.getEddingtonData(JSON.stringify(filters || {})),
    'getEddingtonData'
  )
  return {
    number: result.number,
    history: result.history || [],
    distribution: result.distribution || [],
    next_steps: result.next_steps || [],
  } as EddingtonDataResult
}

// ============================================================================
// Gear Monthly Usage
// ============================================================================

export interface GearMonthlyUsageFilters {
  include_retired?: boolean
}

export interface GearMonthlyUsage {
  month: string
  gear_id: string
  gear_name: string
  source: string
  retired: boolean
  activity_count: number
  distance: number
  moving_time: number
  hashtag?: string
  purchase_price?: number
  purchase_currency?: string
}

export function getGearMonthlyUsage(filters?: GearMonthlyUsageFilters): GearMonthlyUsage[] {
  return (
    callGoStorage<GearMonthlyUsage[] | undefined>(
      () => goStorage.getGearMonthlyUsage(JSON.stringify(filters || {})),
      'getGearMonthlyUsage'
    ) || []
  )
}

// ============================================================================
// Custom Gear
// ============================================================================

export interface CustomGearFilters {
  include_retired?: boolean
  page?: number
  per_page?: number
  order_by?: string
  order_dir?: string
}

export interface CustomGear {
  id: string
  athlete_id: number
  name: string
  primary: boolean
  retired: boolean
  distance: number
  source: string
  hashtag?: string
  purchase_price?: number
  purchase_currency?: string
}

export interface CustomGearResult {
  data: CustomGear[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export const getCustomGear = (filters?: CustomGearFilters): CustomGearResult =>
  callGoStorage<CustomGearResult>(
    () => goStorage.getCustomGear(JSON.stringify(filters || {})),
    'getCustomGear'
  )

export interface CreateCustomGearInput {
  name: string
  hashtag?: string
  retired?: boolean
  purchase_price?: number
  purchase_currency?: string
}

export interface CreateCustomGearResult {
  id: string
  name: string
  hashtag: string
  retired: boolean
  distance: number
}

export function createCustomGear(input: CreateCustomGearInput): CreateCustomGearResult {
  return callGoStorage<CreateCustomGearResult>(
    () => goStorage.createCustomGear(JSON.stringify(input)),
    'createCustomGear'
  )
}

export interface UpdateCustomGearInput {
  id: string
  name?: string
  hashtag?: string
  retired?: boolean
  purchase_price?: number | null
  purchase_currency?: string
}

export function updateCustomGear(input: UpdateCustomGearInput): CreateCustomGearResult {
  return callGoStorage<CreateCustomGearResult>(
    () => goStorage.updateCustomGear(JSON.stringify(input)),
    'updateCustomGear'
  )
}

export interface DeleteCustomGearResult {
  has_activities?: boolean
  message?: string
}

export function deleteCustomGear(id: string, force?: boolean): DeleteCustomGearResult {
  const result = parseGoResult<DeleteCustomGearResult>(
    goStorage.deleteCustomGear(JSON.stringify({ id, force }))
  )
  if (!result.ok && !result.has_activities) {
    throw new Error(result.error || 'Failed to delete custom gear')
  }
  return result
}

// ============================================================================
// Segment Efforts
// ============================================================================

export interface SegmentEffortsFilters {
  segment_id: number
  page?: number
  per_page?: number
  order_by?: string
  order_dir?: string
}

export interface SegmentEffort {
  id: number
  segment_id: number
  activity_id: number
  athlete_id: number
  name: string
  elapsed_time: number
  moving_time: number
  distance: number
  country?: string
  start_date?: string
  start_date_local?: string
  average_watts?: number
  average_heartrate?: number
  max_heartrate?: number
  pr_rank?: number
}

export interface SegmentEffortsResult {
  data: SegmentEffort[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export const getSegmentEfforts = (filters: SegmentEffortsFilters): SegmentEffortsResult =>
  callGoStorage<SegmentEffortsResult>(
    () => goStorage.getSegmentEfforts(JSON.stringify(filters)),
    'getSegmentEfforts'
  )

export interface SegmentCountry {
  country: string
  count: number
}

export function getSegmentCountries(): SegmentCountry[] {
  return (
    callGoStorage<SegmentCountry[] | undefined>(
      () => goStorage.getSegmentCountries(),
      'getSegmentCountries'
    ) || []
  )
}

// ============================================================================
// Wrapped
// ============================================================================

export function getWrappedYears(): number[] {
  return (
    callGoStorage<number[] | undefined>(() => goStorage.getWrappedYears(), 'getWrappedYears') || []
  )
}

export interface WrappedBiggestActivity {
  activity_id: number
  name: string
  sport_type: string
  start_date_local: string
  value: number
}

export interface WrappedPhoto {
  id: string
  activity_id: number
  url: string
  thumbnail_url: string
  caption?: string
}

export interface WrappedData {
  year: number
  range_start: string
  range_end: string
  total_days: number
  active_days: number
  rest_days: number
  totals: {
    activities: number
    distance_m: number
    elevation_m: number
    moving_time_s: number
    kudos: number
    commute_dist_m: number
    carbon_saved_kg: number
  }
  streaks: {
    longest_active_days: number
    longest_rest_days: number
  }
  months?: Array<{
    month: string
    activities: number
    distance_m: number
    elevation_m: number
    prs: number
  }>
  moving_time_by_sport?: Array<{ sport_type: string; moving_time_s: number }>
  start_times_by_hour?: Array<{ hour: number; count: number }>
  locations?: Array<{ lat: number; lng: number; count: number }>
  biggest?: {
    longest_distance?: WrappedBiggestActivity
    most_elevation?: WrappedBiggestActivity
    longest_duration?: WrappedBiggestActivity
  }
  random_photo?: WrappedPhoto
}

export function getWrapped(year?: number): WrappedData | null {
  return callGoStorage<WrappedData | null>(
    () => goStorage.getWrapped(JSON.stringify({ year: year || 0 })),
    'getWrapped'
  )
}

// ============================================================================
// Athlete Metrics (FTP/Weight)
// ============================================================================

export interface MetricEntry {
  recorded_at: string
  value: number
}

export function getFtpHistory(): MetricEntry[] {
  return (
    callGoStorage<MetricEntry[] | undefined>(() => goStorage.getFtpHistory(), 'getFtpHistory') || []
  )
}

export function getFtpRunningHistory(): MetricEntry[] {
  return (
    callGoStorage<MetricEntry[] | undefined>(
      () => goStorage.getFtpRunningHistory(),
      'getFtpRunningHistory'
    ) || []
  )
}

export function getWeightHistory(): MetricEntry[] {
  return (
    callGoStorage<MetricEntry[] | undefined>(
      () => goStorage.getWeightHistory(),
      'getWeightHistory'
    ) || []
  )
}

export const updateFtpHistory = (entries: MetricEntry[]): void =>
  callGoStorageVoid(() => goStorage.updateFtpHistory(JSON.stringify(entries)), 'updateFtpHistory')

export const updateFtpRunningHistory = (entries: MetricEntry[]): void =>
  callGoStorageVoid(
    () => goStorage.updateFtpRunningHistory(JSON.stringify(entries)),
    'updateFtpRunningHistory'
  )

export const updateWeightHistory = (entries: MetricEntry[]): void =>
  callGoStorageVoid(
    () => goStorage.updateWeightHistory(JSON.stringify(entries)),
    'updateWeightHistory'
  )

// ============================================================================
// Training Load (Database)
// ============================================================================

export interface TrainingLoadFilters {
  after?: string
  before?: string
}

export interface TrainingLoadDayData {
  day: string
  tss: number
  ctl: number
  atl: number
  tsb: number
}

export interface TrainingLoadResult {
  series: TrainingLoadDayData[]
  summary?: TrainingLoadDayData
}

export function getTrainingLoad(filters?: TrainingLoadFilters): TrainingLoadResult {
  const result = callGoStorage<TrainingLoadResult>(
    () => goStorage.getTrainingLoad(JSON.stringify(filters || {})),
    'getTrainingLoad'
  )
  return { series: result.series || [], summary: result.summary }
}

// ============================================================================
// Insights
// ============================================================================

export function getInsights(): InsightsResponse {
  const result = callGoStorage<InsightsResponse>(
    () => goStorage.getInsights(),
    'getInsights'
  )
  return { insights: result.insights ?? [] }
}

// ============================================================================
// Power Stats
// ============================================================================

export interface PowerStatsFilters {
  durations?: number[]
  after?: string
  before?: string
  sport_types?: string[]
  history_duration?: number
}

export interface PowerBest {
  duration_s: number
  watts: number
  activity_id: number
  start_date: string
}

export interface PowerHistoryPoint {
  date: string
  watts: number
}

export interface PowerStatsResult {
  best: PowerBest[]
  history: PowerHistoryPoint[]
}

export function getPowerStats(filters?: PowerStatsFilters): PowerStatsResult {
  const result = callGoStorage<PowerStatsResult>(
    () => goStorage.getPowerStats(JSON.stringify(filters || {})),
    'getPowerStats'
  )
  return { best: result.best || [], history: result.history || [] }
}

// ============================================================================
// Challenges
// ============================================================================

export interface ChallengesFilters {
  month?: string
  page?: number
  per_page?: number
}

export interface Challenge {
  id: number
  athlete_id: number
  name: string
  created_at: string
  slug?: string
  badge_url?: string
  local_badge_url?: string
  completion_date?: string
  month?: string
}

export interface ChallengesResult {
  data: Challenge[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export const getChallenges = (filters?: ChallengesFilters): ChallengesResult =>
  callGoStorage<ChallengesResult>(
    () => goStorage.getChallenges(JSON.stringify(filters || {})),
    'getChallenges'
  )

// ============================================================================
// Training Goals
// ============================================================================

export interface TrainingGoalTarget {
  distance_m?: number
  elevation_m?: number
  moving_time_s?: number
  activity_count?: number
}

export interface TrainingGoalSport {
  name: string
  sport_types: string[]
  targets: Record<string, TrainingGoalTarget>
}

export interface TrainingGoalsConfig {
  version?: number
  sports: TrainingGoalSport[]
}

export interface TrainingGoalProgress {
  distance_m: number
  elevation_m: number
  moving_time_s: number
  activity_count: number
}

export interface TrainingGoalsResult {
  config: TrainingGoalsConfig
  progress: Record<string, Record<string, TrainingGoalProgress>>
}

export function getTrainingGoals(): TrainingGoalsResult {
  const result = callGoStorage<TrainingGoalsResult>(
    () => goStorage.getTrainingGoals(),
    'getTrainingGoals'
  )
  return { config: result.config, progress: result.progress }
}

export const updateTrainingGoals = (config: TrainingGoalsConfig): void =>
  callGoStorageVoid(
    () => goStorage.updateTrainingGoals(JSON.stringify(config)),
    'updateTrainingGoals'
  )

// ============================================================================
// Maintenance
// ============================================================================

export interface MaintenanceRule {
  id: number
  component_id: number
  type: string
  threshold_value: number
}

export interface MaintenanceProgress {
  type: string
  threshold_value: number
  current_value: number
  percent: number
  due: boolean
}

export interface MaintenanceDueItem {
  id: number
  gear_id: string
  name: string
  created_at: string
  updated_at: string
  image_url?: string
  maintenance_hashtag?: string
  last_completed_at?: string
  distance_since: number
  moving_time_since: number
  days_since: number
  is_due: boolean
  rules: MaintenanceRule[]
  progress: MaintenanceProgress[]
}

export function getMaintenanceDue(): MaintenanceDueItem[] {
  return (
    callGoStorage<MaintenanceDueItem[] | undefined>(
      () => goStorage.getMaintenanceDue(),
      'getMaintenanceDue'
    ) || []
  )
}

export interface GearComponentsFilters {
  gear_id: string
  page?: number
  per_page?: number
}

export interface GearComponent {
  id: number
  gear_id: string
  name: string
  created_at: string
  updated_at: string
  image_url?: string
  maintenance_hashtag?: string
  last_completed_at?: string
  rules: MaintenanceRule[]
}

export interface GearComponentsResult {
  data: GearComponent[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export const getGearComponents = (filters: GearComponentsFilters): GearComponentsResult =>
  callGoStorage<GearComponentsResult>(
    () => goStorage.getGearComponents(JSON.stringify(filters)),
    'getGearComponents'
  )

export interface CreateComponentRule {
  type: string
  threshold_value: number
}

export interface CreateComponentInput {
  gear_id: string
  name: string
  image_url?: string
  maintenance_hashtag?: string
  rules?: CreateComponentRule[]
}

export interface CreateComponentResult {
  id: number
  gear_id: string
  name: string
}

export function createComponent(input: CreateComponentInput): CreateComponentResult {
  return callGoStorage<CreateComponentResult>(
    () => goStorage.createComponent(JSON.stringify(input)),
    'createComponent'
  )
}

export interface UpdateComponentInput {
  id: number
  name?: string
  image_url?: string
  maintenance_hashtag?: string
  rules?: CreateComponentRule[]
}

export function updateComponent(input: UpdateComponentInput): CreateComponentResult {
  return callGoStorage<CreateComponentResult>(
    () => goStorage.updateComponent(JSON.stringify(input)),
    'updateComponent'
  )
}

export const deleteComponent = (id: number): void =>
  callGoStorageVoid(() => goStorage.deleteComponent(id), 'deleteComponent')

export interface LogMaintenanceInput {
  component_id: number
  activity_id?: number
  completed_at?: string
}

export const logMaintenance = (input: LogMaintenanceInput): void =>
  callGoStorageVoid(() => goStorage.logMaintenance(JSON.stringify(input)), 'logMaintenance')

// ============================================================================
// Settings
// ============================================================================

export interface AppSettings {
  units?: string
  theme?: string
  default_sport_type?: string
  [key: string]: unknown
}

export function getAppSettings(): AppSettings {
  return (
    callGoStorage<AppSettings | undefined>(() => goStorage.getAppSettings(), 'getAppSettings') || {}
  )
}

export const updateAppSettings = (settings: AppSettings): void =>
  callGoStorageVoid(
    () => goStorage.updateAppSettings(JSON.stringify(settings)),
    'updateAppSettings'
  )

// ============================================================================
// HR Zones
// ============================================================================

export interface HRZoneDefinition {
  sport_type: string
  effective_from: string
  method: string
  zones: unknown
}

export function getHrZoneDefinitions(): HRZoneDefinition[] {
  return (
    callGoStorage<HRZoneDefinition[] | undefined>(
      () => goStorage.getHrZoneDefinitions(),
      'getHrZoneDefinitions'
    ) || []
  )
}

export interface UpsertHRZoneInput {
  sport_type: string
  effective_from: string
  method: string
  zones: unknown
}

export const upsertHrZoneDefinition = (input: UpsertHRZoneInput): void =>
  callGoStorageVoid(
    () => goStorage.upsertHrZoneDefinition(JSON.stringify(input)),
    'upsertHrZoneDefinition'
  )

export interface DeleteHRZoneInput {
  sport_type: string
  effective_from: string
}

export const deleteHrZoneDefinition = (input: DeleteHRZoneInput): void =>
  callGoStorageVoid(
    () => goStorage.deleteHrZoneDefinition(JSON.stringify(input)),
    'deleteHrZoneDefinition'
  )

// ============================================================================
// Import - Control (Bridge to Go WASM Importer)
// ============================================================================

export type GoImportPhase =
  | 'idle'
  | 'activities'
  | 'gear'
  | 'streams'
  | 'activity_details'
  | 'segment_details'
  | 'photos'
  | 'completed'
export type GoImportStatus = 'idle' | 'running' | 'completed' | 'failed' | 'canceled' | 'paused'

export interface GoImportOptions {
  full_sync?: boolean
  resume?: boolean
  skip_streams?: boolean
  skip_segments?: boolean
  skip_best_efforts?: boolean
  skip_photos?: boolean
}

export interface GoImportProgress {
  status: GoImportStatus
  phase: GoImportPhase
  activities_total: number
  activities_done: number
  gear_total: number
  gear_done: number
  streams_total: number
  streams_done: number
  details_total: number
  details_done: number
  segments_total: number
  segments_done: number
  photos_total: number
  photos_done: number
  failed_count: number
  remaining_api_calls: number
  estimated_eta?: string
  rate_limit_used_15min: number
  rate_limit_limit_15min: number
  rate_limit_used_daily: number
  rate_limit_limit_daily: number
  waiting_for_rate_limit: boolean
  waiting_until?: string
  waiting_reason?: string
  started_at?: string
  errors?: string[]
}

export interface GoImportState {
  phase: GoImportPhase
  activities_total: number
  activities_done: number
  gear_total: number
  gear_done: number
  streams_total: number
  streams_done: number
  details_total: number
  details_done: number
  segments_total: number
  segments_done: number
  photos_total: number
  photos_done: number
  failed_count: number
  skip_streams: boolean
  skip_segments: boolean
  skip_best_efforts: boolean
  skip_photos: boolean
  started_at?: string
  after_date?: string
  newest_activity_date?: string
  errors?: string[]
}

export interface GoAthlete {
  id: number
  username?: string
  first_name?: string
  last_name?: string
  profile_medium?: string
}

export const goStartImport = (options: GoImportOptions, athlete: GoAthlete): void =>
  callGoStorageVoid(
    () => goStorage.startImport(JSON.stringify(options), JSON.stringify(athlete)),
    'startImport'
  )

export const goCancelImport = (): void =>
  callGoStorageVoid(() => goStorage.cancelImport(), 'cancelImport')

export function goGetImportProgress(): GoImportProgress | null {
  return callGoStorage<GoImportProgress | null>(
    () => goStorage.getImportProgress(),
    'getImportProgress'
  )
}

export function goIsImportRunning(): boolean {
  const result = callGoStorage<{ running: boolean }>(
    () => goStorage.isImportRunning(),
    'isImportRunning'
  )
  return result.running ?? false
}

export function goGetImportState(): GoImportState | null {
  return callGoStorage<GoImportState | null>(() => goStorage.getImportState(), 'getImportState')
}

export const goClearImportState = (): void =>
  callGoStorageVoid(() => goStorage.clearImportState(), 'clearImportState')

// ============================================================================
// Strava Credentials
// ============================================================================

export interface StravaCredentials {
  client_id: string
  client_secret: string
}

export function getStravaCredentials(): StravaCredentials | null {
  if (!initialized) {
    return null
  }
  return callGoStorage<StravaCredentials | null>(
    () => goStorage.getStravaCredentials(),
    'getStravaCredentials'
  )
}

export const saveStravaCredentials = (clientId: string, clientSecret: string): void =>
  callGoStorageVoid(
    () =>
      goStorage.saveStravaCredentials(
        JSON.stringify({ client_id: clientId, client_secret: clientSecret })
      ),
    'saveStravaCredentials'
  )

export const deleteStravaCredentials = (): void =>
  callGoStorageVoid(() => goStorage.deleteStravaCredentials(), 'deleteStravaCredentials')

export function hasStravaCredentials(): boolean {
  if (!initialized) {
    return false
  }
  const result = callGoStorage<{ has_credentials: boolean } | undefined>(
    () => goStorage.hasStravaCredentials(),
    'hasStravaCredentials'
  )
  return result?.has_credentials ?? false
}
