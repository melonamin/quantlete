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

import { initializeDatabase, getDatabase } from './db'

// Declare global Go WASM types
declare global {
  interface Window {
    _go_sqlite: SqlJsStatic
    _go_sqlite_dbs: Map<string, SqlJsDatabase>
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

  // goStorage namespace registered by Go WASM
  const goStorage: {
    // Initialization
    init(): string
    setAthleteId(id: number): string
    exportDb(): Uint8Array

    // Auth
    getAuthStatus(): string

    // Activities - Read
    getActivities(filtersJSON: string): string
    getActivity(id: number): string
    getActivityStreams(activityId: number): string

    // Activities - Write
    saveActivity(activityJSON: string): string
    saveStream(streamJSON: string): string

    // Athlete - Write
    saveAthlete(athleteJSON: string): string

    // Gear - Write
    saveGear(gearJSON: string): string

    // Segments - Write
    saveSegment(segmentJSON: string): string
    saveSegmentEffort(effortJSON: string): string

    // Best Efforts - Write
    saveBestEfforts(dataJSON: string): string

    // Photos - Write
    savePhoto(photoJSON: string): string

    // Power - Write (compute and store)
    computePowerBestEfforts(dataJSON: string): string

    // Sync History - Write
    createSyncRun(dataJSON: string): string
    updateSyncRun(dataJSON: string): string
    completeSyncRun(dataJSON: string): string

    // Dashboard
    getDashboardStats(): string
    getWeeklyStats(): string
    getRecentActivities(limit: number): string
    getSportTypeStats(): string
    getMonthlyStats(year?: number): string
    getYearlyStats(): string
    getDaytimeDistribution(): string
    getWeekdayDistribution(): string
    getExportStats(): string

    // Heatmap
    getHeatmapData(filtersJSON: string): string

    // Gear - Read
    getGear(filtersJSON: string): string
    getGearDetail(id: string): string

    // Segments - Read
    getSegments(filtersJSON: string): string
    getSegmentDetail(id: number): string

    // Photos - Read
    getPhotos(filtersJSON: string): string
    getActivityPhotos(activityId: number): string

    // Calendar - Read
    getCalendarData(year: number): string

    // Best Efforts - Read
    getBestEffortPRs(sportType?: string): string
    getBestEffortsForType(distanceType: string, sportType?: string): string

    // Sync History - Read
    getSyncHistory(limit?: number): string

    // Algorithms - Power
    normalizedPower(watts: number[]): string
    rollingMaxAverage(values: number[], windowSeconds: number): string
    intensityFactor(np: number, ftp: number): string
    trainingStressScore(durationSeconds: number, np: number, ftp: number): string

    // Algorithms - Eddington
    eddingtonNumber(distances: number[]): string
    eddingtonNextSteps(distances: number[], currentE: number, stepsToCalculate: number): string
    eddingtonHistory(distances: number[]): string

    // Algorithms - Training Load
    calculateTrainingLoad(dailyTss: number[], ctlTau: number, atlTau: number): string
    calculateTrainingLoadWithInitial(
      dailyTss: number[],
      initialCtl: number,
      initialAtl: number,
      ctlTau: number,
      atlTau: number,
    ): string
    predictAfterWorkout(
      currentCtl: number,
      currentAtl: number,
      plannedTss: number,
      ctlTau: number,
      atlTau: number,
    ): string
    tssForTargetTsb(
      currentCtl: number,
      currentAtl: number,
      targetTsb: number,
      ctlTau: number,
      atlTau: number,
    ): string

    // Eddington
    getEddingtonData(filtersJSON: string): string

    // Gear Stats
    getGearMonthlyUsage(filtersJSON: string): string

    // Segment Efforts
    getSegmentEfforts(filtersJSON: string): string
    getSegmentCountries(): string

    // Rewind
    getRewindYears(): string
    getRewind(year: number): string

    // Athlete Metrics (FTP/Weight)
    getFtpHistory(): string
    getWeightHistory(): string
    updateFtpHistory(entriesJSON: string): string
    updateWeightHistory(entriesJSON: string): string

    // Training Load
    getTrainingLoad(filtersJSON: string): string

    // Power Stats
    getPowerStats(filtersJSON: string): string

    // Challenges
    getChallenges(filtersJSON: string): string

    // Training Goals
    getTrainingGoals(): string
    updateTrainingGoals(configJSON: string): string

    // Maintenance
    getMaintenanceDue(): string
    getGearComponents(filtersJSON: string): string
    createComponent(componentJSON: string): string
    updateComponent(componentJSON: string): string
    deleteComponent(id: number): string
    logMaintenance(logJSON: string): string

    // Settings
    getAppSettings(): string
    updateAppSettings(settingsJSON: string): string

    // Custom Gear
    getCustomGear(filtersJSON: string): string
    createCustomGear(gearJSON: string): string
    updateCustomGear(gearJSON: string): string
    deleteCustomGear(deleteJSON: string): string

    // HR Zones
    getHrZoneDefinitions(): string
    upsertHrZoneDefinition(zoneJSON: string): string
    deleteHrZoneDefinition(deleteJSON: string): string
  }
}

export interface GoStorageResult<T = unknown> {
  ok: boolean
  error?: string
  message?: string
  data?: T
}

let initialized = false

/**
 * Require goStorage to be available. Throws if WASM hasn't loaded.
 * Use this to fail fast with a clear error message.
 */
function requireGoStorage(operation: string): typeof goStorage {
  if (typeof goStorage === 'undefined' || goStorage === null) {
    throw new Error(
      `goStorage is not available for ${operation}. ` +
        'Ensure WASM has been loaded and initGoStorage() has been called.',
    )
  }
  return goStorage
}

/**
 * Parse Go WASM result - all Go functions return JSON strings.
 * @param result - JSON string returned by Go WASM function
 * @param operation - Name of the operation for error context
 */
export function parseGoResult<T>(result: string, operation?: string): GoStorageResult<T> & T {
  try {
    return JSON.parse(result)
  } catch {
    const ctx = operation ? ` (${operation})` : ''
    return { ok: false, error: `Failed to parse Go result${ctx}: ${result}` } as GoStorageResult<T> & T
  }
}

// ============================================================================
// WASM Connection Lifecycle
// ============================================================================
//
// The Go WASM storage layer follows this lifecycle:
//
// 1. INITIALIZATION (initGoStorage):
//    a. Load wasm_exec.js (Go WebAssembly runtime)
//    b. Initialize sql.js for the SQLite engine
//    c. Load and instantiate quantlete.wasm
//    d. Run the Go main() which registers goStorage methods
//    e. Call goStorage.init() to create the database and run migrations
//    f. Restore database from OPFS if available
//
// 2. USAGE:
//    - All goStorage.* methods are available after initialization
//    - The database is in-memory (sql.js handles the underlying storage)
//    - Changes are persisted to OPFS periodically via persistToOpfs()
//    - Use requireGoStorage(operation) to safely access goStorage
//
// 3. PERSISTENCE:
//    - Call persistToOpfs() to save database state to OPFS
//    - The database is exported as a binary blob via goStorage.exportDb()
//    - OPFS storage survives browser restarts
//
// 4. TEARDOWN:
//    - Currently, there is no explicit teardown/close method
//    - The database is in-memory and will be garbage collected when the page unloads
//    - Always call persistToOpfs() before page unload to save state
//    - Consider calling persistToOpfs() in a beforeunload handler
//
// NOTE: The Go WASM does not currently expose a db.Close() method.
// This is intentional as sql.js databases are in-memory and don't require
// explicit cleanup. If you need to reinitialize, reload the page.
// ============================================================================

/**
 * Load wasm_exec.js (Go WebAssembly runtime)
 */
async function loadGoRuntime(): Promise<void> {
  // Check if Go class already exists (script already loaded)
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

/**
 * Load Go WASM binary
 */
async function loadGoWasm(): Promise<void> {
  // First load the Go runtime (wasm_exec.js)
  await loadGoRuntime()

  const Go = (window as unknown as { Go: new () => GoInstance }).Go
  const go = new Go()
  const result = await WebAssembly.instantiateStreaming(
    fetch('/wasm/quantlete.wasm'),
    go.importObject,
  )
  // Don't await - Go needs to keep running
  go.run(result.instance)
}

// Go instance type from wasm_exec.js
interface GoInstance {
  importObject: WebAssembly.Imports
  run(instance: WebAssembly.Instance): Promise<void>
}

/**
 * Initialize Go WASM storage layer
 *
 * This shares the same database with the existing WasmDatabase,
 * so both TypeScript (importer) and Go WASM (queries) can access the same data.
 */
export async function initGoStorage(): Promise<void> {
  if (initialized) return

  console.log('[go-storage] Initializing...')

  // 1. Initialize the existing WasmDatabase (handles OPFS, migrations)
  const wasmDb = await initializeDatabase()
  console.log('[go-storage] WasmDatabase ready')

  // 2. Get the internal sql.js instances
  const sqlJs = wasmDb.getSqlJs()
  const internalDb = wasmDb.getInternalDb()

  if (!sqlJs || !internalDb) {
    throw new Error('Failed to get sql.js instances from WasmDatabase')
  }

  // 3. Set up globals for go-sqlite3-js
  // Wrap Database constructor to handle :memory: correctly
  const OriginalDatabase = sqlJs.Database
  const wrappedSqlJs = {
    ...sqlJs,
    Database: function (data?: Uint8Array | string) {
      if (data === ':memory:' || data === 'file::memory:' || data === '') {
        console.log('[go-storage] Returning shared database for :memory:')
        // Return the existing database instead of creating a new one
        return internalDb
      }
      return new OriginalDatabase(data as Uint8Array)
    } as unknown as typeof sqlJs.Database,
  }
  wrappedSqlJs.Database.prototype = OriginalDatabase.prototype

  window._go_sqlite = wrappedSqlJs as SqlJsStatic
  window._go_sqlite_dbs = new Map()
  // Pre-register the shared database
  window._go_sqlite_dbs.set(':memory:', internalDb as unknown as SqlJsDatabase)

  console.log('[go-storage] sql.js globals configured')

  // 4. Load and run Go WASM
  await loadGoWasm()
  console.log('[go-storage] Go WASM loaded')

  // Wait a moment for Go to register functions
  await new Promise((resolve) => setTimeout(resolve, 100))

  // 5. Initialize Go storage (should recognize existing database)
  const gs = requireGoStorage('init')
  const result = parseGoResult(gs.init(), 'init')
  if (!result.ok) {
    throw new Error(`Go storage init failed: ${result.error}`)
  }

  initialized = true
  console.log('[go-storage] Initialization complete')
}

/**
 * Check if storage is initialized
 */
export function isInitialized(): boolean {
  return initialized
}

/**
 * Persist the database to OPFS
 */
export async function persistDatabase(): Promise<void> {
  const db = getDatabase()
  await db.persist()
}

/**
 * Reset the database
 */
export async function clearDatabase(): Promise<void> {
  const db = getDatabase()
  await db.reset()
}

/**
 * Set the current athlete ID for queries
 */
export function setAthleteId(id: number): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.setAthleteId(id))
  if (!result.ok) {
    throw new Error(`Failed to set athlete ID: ${result.error}`)
  }
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

export function getActivities(filters: ActivityFilters): ActivitiesResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<ActivitiesResult>(goStorage.getActivities(JSON.stringify(filters)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get activities')
  }
  return result
}

export function getActivity(id: number): Activity {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<Activity>(goStorage.getActivity(id))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get activity')
  }
  return result
}

export interface ActivityStream {
  type: string
  data: number[]
  resolution: string
}

export function getActivityStreams(activityId: number): ActivityStream[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<ActivityStream[]>(goStorage.getActivityStreams(activityId))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get activity streams')
  }
  // Arrays are wrapped in {ok: true, data: [...]}
  return result.data ?? []
}

// ============================================================================
// Dashboard
// ============================================================================

export interface DashboardStats {
  total_activities: number
  total_distance: number
  total_moving_time: number
  total_elevation_gain: number
  total_kudos: number
  total_photos: number
  first_activity_date?: string
  last_activity_date?: string
  sport_types: string[]
}

export function getDashboardStats(): DashboardStats {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<DashboardStats>(goStorage.getDashboardStats())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get dashboard stats')
  }
  return result
}

export interface WeeklyStat {
  sport_type: string
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export function getWeeklyStats(): WeeklyStat[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<WeeklyStat[]>(goStorage.getWeeklyStats())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get weekly stats')
  }
  // Arrays are wrapped in {ok: true, data: [...]}
  return result.data ?? []
}

export interface RecentActivity {
  id: number
  name: string
  sport_type: string
  start_date: string
  distance: number
  moving_time: number
  elevation_gain: number
  summary_polyline?: string
}

export function getRecentActivities(limit: number): RecentActivity[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<RecentActivity[]>(goStorage.getRecentActivities(limit))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get recent activities')
  }
  // Arrays are wrapped in {ok: true, data: [...]}
  return result.data ?? []
}

export interface SportTypeStat {
  sport_type: string
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export function getSportTypeStats(): SportTypeStat[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<SportTypeStat[]>(goStorage.getSportTypeStats())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get sport type stats')
  }
  // Arrays are wrapped in {ok: true, data: [...]}
  return result.data ?? []
}

export interface MonthlyStat {
  month: string
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export function getMonthlyStats(year?: number): MonthlyStat[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<MonthlyStat[]>(goStorage.getMonthlyStats(year))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get monthly stats')
  }
  // Arrays are wrapped in {ok: true, data: [...]}
  return result.data ?? []
}

export interface YearlyStat {
  year: number
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

export function getYearlyStats(): YearlyStat[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<YearlyStat[]>(goStorage.getYearlyStats())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get yearly stats')
  }
  // Arrays are wrapped in {ok: true, data: [...]}
  return result.data ?? []
}

// ============================================================================
// Distribution Stats
// ============================================================================

export interface DistributionSlice {
  label: string
  count: number
}

export function getDaytimeDistribution(): DistributionSlice[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<DistributionSlice[]>(goStorage.getDaytimeDistribution())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get daytime distribution')
  }
  return result.data ?? []
}

export function getWeekdayDistribution(): DistributionSlice[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<DistributionSlice[]>(goStorage.getWeekdayDistribution())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get weekday distribution')
  }
  return result.data ?? []
}

export interface ExportStats {
  total_activities: number
  first_activity: string | null
  last_activity: string | null
}

export function getExportStats(): ExportStats {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<ExportStats>(goStorage.getExportStats())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get export stats')
  }
  return result.data ?? { total_activities: 0, first_activity: null, last_activity: null }
}

// ============================================================================
// Heatmap
// ============================================================================

export interface HeatmapFilters {
  sport_type?: string
  year?: number
  commute?: boolean
}

export interface HeatmapActivity {
  id: number
  name: string
  sport_type: string
  start_date: string
  distance: number
  summary_polyline: string
  start_lat: number
  start_lng: number
}

export function getHeatmapData(filters: HeatmapFilters): HeatmapActivity[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<HeatmapActivity[]>(goStorage.getHeatmapData(JSON.stringify(filters)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get heatmap data')
  }
  // Arrays are wrapped in {ok: true, data: [...]}
  return result.data ?? []
}

// ============================================================================
// Calendar
// ============================================================================

export interface CalendarDay {
  date: string
  activity_count: number
  total_distance: number
  total_time: number
  total_calories: number
}

export function getCalendarData(year: number): CalendarDay[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: CalendarDay[] }>(goStorage.getCalendarData(year))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get calendar data')
  }
  return result.data || []
}

export interface CalendarActivityResult {
  id: number
  name: string
  sport_type: string
  start_date: string
  distance: number
  moving_time: number
  total_elevation_gain: number
}

export function getCalendarActivities(year: number, month: number): CalendarActivityResult[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  // Use getActivities with date filters for this month
  const startDate = `${year}-${String(month).padStart(2, '0')}-01`
  const endMonth = month === 12 ? 1 : month + 1
  const endYear = month === 12 ? year + 1 : year
  const endDate = `${endYear}-${String(endMonth).padStart(2, '0')}-01`

  const result = getActivities({
    after: startDate,
    before: endDate,
    per_page: 1000,
  })

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  // Calculate summary from activities for this month
  const activities = getCalendarActivities(year, month)

  return {
    year,
    month,
    activity_count: activities.length,
    total_distance: activities.reduce((sum, a) => sum + a.distance, 0),
    total_elevation_gain: activities.reduce((sum, a) => sum + a.total_elevation_gain, 0),
    total_moving_time: activities.reduce((sum, a) => sum + a.moving_time, 0),
    total_calories: 0, // Not available in activity data
    workout_count: 0, // Would need workout_type filter
    challenges_completed: 0, // Would need challenges data
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

export function getGear(filters?: GearFilters): GearResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<GearResult>(goStorage.getGear(JSON.stringify(filters || {})))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get gear')
  }
  return result
}

export function getGearDetail(id: string): GearItem {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: GearItem }>(goStorage.getGearDetail(id))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get gear detail')
  }
  return result.data as GearItem
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

export function getSegments(filters?: SegmentsFilters): SegmentsResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<SegmentsResult>(goStorage.getSegments(JSON.stringify(filters || {})))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get segments')
  }
  return result
}

export interface SegmentDetailResult {
  segment: SegmentListItem
  efforts: SegmentEffort[]
}

export function getSegmentDetail(id: number): SegmentDetailResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<SegmentDetailResult>(goStorage.getSegmentDetail(id))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get segment detail')
  }
  return result
}

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

export function getPhotos(filters?: PhotosFilters): PhotosResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<PhotosResult>(goStorage.getPhotos(JSON.stringify(filters || {})))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get photos')
  }
  return result
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: ActivityPhotoResult[] }>(goStorage.getActivityPhotos(activityId))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get activity photos')
  }
  return result.data || []
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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: BestEffortPRResult[] }>(goStorage.getBestEffortPRs(sportType))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get best effort PRs')
  }
  return result.data || []
}

export interface BestEffortItemResult extends BestEffortPRResult {
  start_index?: number
  end_index?: number
}

export function getBestEffortsForType(distanceType: string, sportType?: string): BestEffortItemResult[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: BestEffortItemResult[] }>(
    goStorage.getBestEffortsForType(distanceType, sportType),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get best efforts for type')
  }
  return result.data || []
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

export function saveActivity(activity: SaveActivityInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('saveActivity')
  const result = parseGoResult(gs.saveActivity(JSON.stringify(activity)), 'saveActivity')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save activity')
  }
}

export interface SaveStreamInput {
  activity_id: number
  stream_type: string
  data: unknown[]
  series_type: string
  original_size: number
  resolution: string
}

export function saveStream(stream: SaveStreamInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('saveStream')
  const result = parseGoResult(gs.saveStream(JSON.stringify(stream)), 'saveStream')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save stream')
  }
}

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

export function saveAthlete(athlete: SaveAthleteInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('saveAthlete')
  const result = parseGoResult(gs.saveAthlete(JSON.stringify(athlete)), 'saveAthlete')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save athlete')
  }
}

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

export function saveGear(gear: SaveGearInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('saveGear')
  const result = parseGoResult(gs.saveGear(JSON.stringify(gear)), 'saveGear')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save gear')
  }
}

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

export function saveSegment(segment: SaveSegmentInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('saveSegment')
  const result = parseGoResult(gs.saveSegment(JSON.stringify(segment)), 'saveSegment')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save segment')
  }
}

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

export function saveSegmentEffort(effort: SaveSegmentEffortInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('saveSegmentEffort')
  const result = parseGoResult(gs.saveSegmentEffort(JSON.stringify(effort)), 'saveSegmentEffort')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save segment effort')
  }
}

export interface BestEffortInput {
  distance_type: string
  name?: string
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

export function saveBestEfforts(data: SaveBestEffortsInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('saveBestEfforts')
  const result = parseGoResult(gs.saveBestEfforts(JSON.stringify(data)), 'saveBestEfforts')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save best efforts')
  }
}

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

export function savePhoto(photo: SavePhotoInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('savePhoto')
  const result = parseGoResult(gs.savePhoto(JSON.stringify(photo)), 'savePhoto')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to save photo')
  }
}

// ============================================================================
// Power Best Efforts
// ============================================================================

export interface ComputePowerBestEffortsInput {
  activity_id: number
  athlete_id: number
}

/**
 * Compute and store power best efforts for an activity.
 * Uses the Go storage layer to read the watts stream and compute rolling max averages.
 * This replaces the TypeScript algorithms.wasm computation.
 */
export function computePowerBestEfforts(input: ComputePowerBestEffortsInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const gs = requireGoStorage('computePowerBestEfforts')
  const result = parseGoResult(gs.computePowerBestEfforts(JSON.stringify(input)), 'computePowerBestEfforts')
  if (!result.ok) {
    throw new Error(result.error || 'Failed to compute power best efforts')
  }
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<CreateSyncRunResult>(goStorage.createSyncRun(JSON.stringify(input)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to create sync run')
  }
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

export function updateSyncRun(input: UpdateSyncRunInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.updateSyncRun(JSON.stringify(input)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to update sync run')
  }
}

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

export function completeSyncRun(input: CompleteSyncRunInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.completeSyncRun(JSON.stringify(input)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to complete sync run')
  }
}

// ============================================================================
// Algorithms - Power
// ============================================================================

/**
 * Calculate normalized power from power data.
 * Uses 30-second rolling average, then takes the 4th root of the mean of 4th powers.
 */
export function normalizedPower(watts: number[]): number {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ value: number }>(goStorage.normalizedPower(watts))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate normalized power')
  }
  return (result as { value: number }).value
}

/**
 * Find the maximum rolling average over a given window.
 * Used for peak power calculations (e.g., 5s peak, 1min peak, 5min peak).
 */
export function rollingMaxAverage(values: number[], windowSeconds: number): number {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ value: number }>(goStorage.rollingMaxAverage(values, windowSeconds))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate rolling max average')
  }
  return (result as { value: number }).value
}

/**
 * Calculate intensity factor: IF = NP / FTP
 */
export function intensityFactor(np: number, ftp: number): number {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ value: number }>(goStorage.intensityFactor(np, ftp))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate intensity factor')
  }
  return (result as { value: number }).value
}

/**
 * Calculate Training Stress Score (TSS)
 */
export function trainingStressScore(durationSeconds: number, np: number, ftp: number): number {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ value: number }>(
    goStorage.trainingStressScore(durationSeconds, np, ftp),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate TSS')
  }
  return (result as { value: number }).value
}

// ============================================================================
// Algorithms - Eddington
// ============================================================================

/**
 * Calculate the Eddington number from daily distances (in km).
 * The Eddington number E is the largest number such that you have cycled
 * at least E km on at least E days.
 */
export function eddingtonNumber(distances: number[]): number {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ value: number }>(goStorage.eddingtonNumber(distances))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate Eddington number')
  }
  return (result as { value: number }).value
}

export interface EddingtonNextStep {
  target: number
  days_needed: number
}

/**
 * Calculate how many more days of riding are needed for each of the next Eddington numbers.
 */
export function eddingtonNextSteps(
  distances: number[],
  currentE: number,
  stepsToCalculate: number,
): EddingtonNextStep[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: EddingtonNextStep[] }>(
    goStorage.eddingtonNextSteps(distances, currentE, stepsToCalculate),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate Eddington next steps')
  }
  return (result as { data: EddingtonNextStep[] }).data
}

/**
 * Calculate progressive Eddington numbers over time.
 * Given distances in chronological order, returns E for each day.
 */
export function eddingtonHistory(distances: number[]): number[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: number[] }>(goStorage.eddingtonHistory(distances))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate Eddington history')
  }
  return (result as { data: number[] }).data
}

// ============================================================================
// Algorithms - Training Load
// ============================================================================

export interface TrainingLoadPoint {
  ctl: number
  atl: number
  tsb: number
}

/**
 * Calculate training load metrics from daily TSS values.
 * Uses EWMA: new_value = old_value + (tss - old_value) * (1 / tau)
 * Default tau values: CTL = 42 days, ATL = 7 days
 */
export function calculateTrainingLoad(
  dailyTss: number[],
  ctlTau: number,
  atlTau: number,
): TrainingLoadPoint[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: TrainingLoadPoint[] }>(
    goStorage.calculateTrainingLoad(dailyTss, ctlTau, atlTau),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate training load')
  }
  return (result as { data: TrainingLoadPoint[] }).data
}

/**
 * Calculate training load starting from existing CTL/ATL values.
 * Useful for continuing from a known state.
 */
export function calculateTrainingLoadWithInitial(
  dailyTss: number[],
  initialCtl: number,
  initialAtl: number,
  ctlTau: number,
  atlTau: number,
): TrainingLoadPoint[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: TrainingLoadPoint[] }>(
    goStorage.calculateTrainingLoadWithInitial(dailyTss, initialCtl, initialAtl, ctlTau, atlTau),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate training load with initial')
  }
  return (result as { data: TrainingLoadPoint[] }).data
}

/**
 * Calculate predicted TSB after a planned workout.
 */
export function predictAfterWorkout(
  currentCtl: number,
  currentAtl: number,
  plannedTss: number,
  ctlTau: number,
  atlTau: number,
): TrainingLoadPoint {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<TrainingLoadPoint>(
    goStorage.predictAfterWorkout(currentCtl, currentAtl, plannedTss, ctlTau, atlTau),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to predict after workout')
  }
  return { ctl: result.ctl, atl: result.atl, tsb: result.tsb } as TrainingLoadPoint
}

/**
 * Calculate the TSS needed to reach a target TSB.
 */
export function tssForTargetTsb(
  currentCtl: number,
  currentAtl: number,
  targetTsb: number,
  ctlTau: number,
  atlTau: number,
): number {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ value: number }>(
    goStorage.tssForTargetTsb(currentCtl, currentAtl, targetTsb, ctlTau, atlTau),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to calculate TSS for target TSB')
  }
  return (result as { value: number }).value
}

// ============================================================================
// Eddington Data
// ============================================================================

export interface EddingtonFilters {
  sport_types?: string[]
}

export interface EddingtonHistoryPoint {
  date: string
  number: number
}

export interface EddingtonDataResult {
  number: number
  history: EddingtonHistoryPoint[]
}

export function getEddingtonData(filters?: EddingtonFilters): EddingtonDataResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<EddingtonDataResult>(
    goStorage.getEddingtonData(JSON.stringify(filters || {})),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get Eddington data')
  }
  return { number: result.number, history: result.history } as EddingtonDataResult
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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: GearMonthlyUsage[] }>(
    goStorage.getGearMonthlyUsage(JSON.stringify(filters || {})),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get gear monthly usage')
  }
  return result.data || []
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

export function getCustomGear(filters?: CustomGearFilters): CustomGearResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<CustomGearResult>(
    goStorage.getCustomGear(JSON.stringify(filters || {})),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get custom gear')
  }
  return result
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: CreateCustomGearResult }>(
    goStorage.createCustomGear(JSON.stringify(input)),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to create custom gear')
  }
  return result.data as CreateCustomGearResult
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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: CreateCustomGearResult }>(
    goStorage.updateCustomGear(JSON.stringify(input)),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to update custom gear')
  }
  return result.data as CreateCustomGearResult
}

export interface DeleteCustomGearResult {
  has_activities?: boolean
  message?: string
}

export function deleteCustomGear(id: string, force?: boolean): DeleteCustomGearResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<DeleteCustomGearResult>(
    goStorage.deleteCustomGear(JSON.stringify({ id, force })),
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

export function getSegmentEfforts(filters: SegmentEffortsFilters): SegmentEffortsResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<SegmentEffortsResult>(
    goStorage.getSegmentEfforts(JSON.stringify(filters)),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get segment efforts')
  }
  return result
}

export interface SegmentCountry {
  country: string
  count: number
}

export function getSegmentCountries(): SegmentCountry[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: SegmentCountry[] }>(goStorage.getSegmentCountries())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get segment countries')
  }
  return result.data || []
}

// ============================================================================
// Rewind
// ============================================================================

export function getRewindYears(): number[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: number[] }>(goStorage.getRewindYears())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get rewind years')
  }
  return result.data || []
}

export interface RewindBiggestActivity {
  activity_id: number
  name: string
  sport_type: string
  start_date_local: string
  value: number
}

export interface RewindPhoto {
  id: string
  activity_id: number
  url: string
  thumbnail_url: string
  caption?: string
}

export interface RewindData {
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
  moving_time_by_sport?: Array<{
    sport_type: string
    moving_time_s: number
  }>
  start_times_by_hour?: Array<{
    hour: number
    count: number
  }>
  locations?: Array<{
    lat: number
    lng: number
    count: number
  }>
  biggest?: {
    longest_distance?: RewindBiggestActivity
    most_elevation?: RewindBiggestActivity
    longest_duration?: RewindBiggestActivity
  }
  random_photo?: RewindPhoto
}

export function getRewind(year?: number): RewindData | null {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: RewindData | null }>(goStorage.getRewind(year || 0))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get rewind')
  }
  return result.data || null
}

// ============================================================================
// Athlete Metrics (FTP/Weight)
// ============================================================================

export interface MetricEntry {
  recorded_at: string
  value: number
}

export function getFtpHistory(): MetricEntry[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: MetricEntry[] }>(goStorage.getFtpHistory())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get FTP history')
  }
  return result.data || []
}

export function getWeightHistory(): MetricEntry[] {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: MetricEntry[] }>(goStorage.getWeightHistory())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get weight history')
  }
  return result.data || []
}

export function updateFtpHistory(entries: MetricEntry[]): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.updateFtpHistory(JSON.stringify(entries)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to update FTP history')
  }
}

export function updateWeightHistory(entries: MetricEntry[]): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.updateWeightHistory(JSON.stringify(entries)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to update weight history')
  }
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<TrainingLoadResult>(
    goStorage.getTrainingLoad(JSON.stringify(filters || {})),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get training load')
  }
  return { series: result.series || [], summary: result.summary }
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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<PowerStatsResult>(
    goStorage.getPowerStats(JSON.stringify(filters || {})),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get power stats')
  }
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

export function getChallenges(filters?: ChallengesFilters): ChallengesResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<ChallengesResult>(
    goStorage.getChallenges(JSON.stringify(filters || {})),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get challenges')
  }
  return result
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<TrainingGoalsResult>(goStorage.getTrainingGoals())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get training goals')
  }
  return { config: result.config, progress: result.progress }
}

export function updateTrainingGoals(config: TrainingGoalsConfig): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.updateTrainingGoals(JSON.stringify(config)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to update training goals')
  }
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: MaintenanceDueItem[] }>(goStorage.getMaintenanceDue())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get maintenance due')
  }
  return result.data || []
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

export function getGearComponents(filters: GearComponentsFilters): GearComponentsResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<GearComponentsResult>(
    goStorage.getGearComponents(JSON.stringify(filters)),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get gear components')
  }
  return result
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: CreateComponentResult }>(
    goStorage.createComponent(JSON.stringify(input)),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to create component')
  }
  return result.data as CreateComponentResult
}

export interface UpdateComponentInput {
  id: number
  name?: string
  image_url?: string
  maintenance_hashtag?: string
  rules?: CreateComponentRule[]
}

export function updateComponent(input: UpdateComponentInput): CreateComponentResult {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: CreateComponentResult }>(
    goStorage.updateComponent(JSON.stringify(input)),
  )
  if (!result.ok) {
    throw new Error(result.error || 'Failed to update component')
  }
  return result.data as CreateComponentResult
}

export function deleteComponent(id: number): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.deleteComponent(id))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to delete component')
  }
}

export interface LogMaintenanceInput {
  component_id: number
  activity_id?: number
  completed_at?: string
}

export function logMaintenance(input: LogMaintenanceInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.logMaintenance(JSON.stringify(input)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to log maintenance')
  }
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: AppSettings }>(goStorage.getAppSettings())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get app settings')
  }
  return result.data || {}
}

export function updateAppSettings(settings: AppSettings): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.updateAppSettings(JSON.stringify(settings)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to update app settings')
  }
}

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
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult<{ data: HRZoneDefinition[] }>(goStorage.getHrZoneDefinitions())
  if (!result.ok) {
    throw new Error(result.error || 'Failed to get HR zone definitions')
  }
  return result.data || []
}

export interface UpsertHRZoneInput {
  sport_type: string
  effective_from: string
  method: string
  zones: unknown
}

export function upsertHrZoneDefinition(input: UpsertHRZoneInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.upsertHrZoneDefinition(JSON.stringify(input)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to upsert HR zone definition')
  }
}

export interface DeleteHRZoneInput {
  sport_type: string
  effective_from: string
}

export function deleteHrZoneDefinition(input: DeleteHRZoneInput): void {
  if (!initialized) {
    throw new Error('Go storage not initialized')
  }
  const result = parseGoResult(goStorage.deleteHrZoneDefinition(JSON.stringify(input)))
  if (!result.ok) {
    throw new Error(result.error || 'Failed to delete HR zone definition')
  }
}
