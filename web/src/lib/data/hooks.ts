/**
 * React Query hooks that use the DataProvider.
 *
 * These hooks work in both server and WASM modes by delegating
 * to the appropriate DataProvider implementation.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDataProviderStatus } from './context'
import { STALE_TIME } from '@/lib/constants'
import type {
  Activity,
  ActivityAnalysis,
  ActivityFilters,
  ActivityStream,
  ActivitiesResponse,
  ActivityPhoto,
  AppSettings,
  AuthStatus,
  BestEffortItem,
  BestEffortPR,
  CalendarActivity,
  CalendarDay,
  CalendarMonthSummary,
  Challenge,
  ChallengesFilters,
  ChallengesResponse,
  ComponentWithRules,
  ComponentsFilters,
  ComponentsResponse,
  CreateComponentRequest,
  UpdateComponentRequest,
  CredentialsStatus,
  CustomGearCreateRequest,
  DashboardConfig,
  DashboardData,
  DashboardStats,
  DistributionSlice,
  DueComponent,
  EddingtonHistoryPoint,
  EddingtonResult,
  ExportStats,
  FTPHistoryResponse,
  Gear,
  GearFilters,
  GearListResponse,
  GearMonthlyUsage,
  HeatmapFilters,
  HeatmapResponse,
  HrZoneDefinition,
  HrZonesResponse,
  ImportProgress,
  InsightsResponse,
  LogMaintenanceRequest,
  MonthlyComparisonResponse,
  MonthlyStat,
  PhotosFilters,
  PhotosListResponse,
  PowerStatsResponse,
  PowerZonesResponse,
  RecentActivity,
  WrappedReport,
  SegmentCountryStat,
  SegmentDetailResponse,
  SegmentEffort,
  SegmentEffortsFilters,
  SegmentEffortsResponse,
  SegmentsFilters,
  SegmentsResponse,
  SportTypeStat,
  StartImportRequest,
  SyncWatermark,
  TrainingGoalsConfig,
  TrainingGoalsResponse,
  TrainingLoadResponse,
  UpdateCredentialsRequest,
  WeeklyStat,
  WeeklyTrendsResponse,
  WeightHistoryResponse,
  YearlyStat,
  ZoneTrendResponse,
  ActivityWeather,
} from './types'

// ============================================================================
// Auth
// ============================================================================

export function useAuthStatus() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'auth', 'status'],
    queryFn: async (): Promise<AuthStatus> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getAuthStatus()
    },
    enabled: initialized && !error && !!provider,
    staleTime: 1000 * 30, // 30 seconds - shorter than default to ensure auth state updates quickly
    retry: false,
  })
}

export function useRefreshToken() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.refreshToken()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'auth', 'status'] })
    },
  })
}

// ============================================================================
// Dashboard
// ============================================================================

export function useDashboard() {
  const { provider, initialized, error } = useDataProviderStatus()
  const { data: auth, isLoading: authLoading } = useAuthStatus()
  const isAuthenticated = auth?.authenticated ?? false

  return useQuery({
    queryKey: ['data', 'dashboard'],
    queryFn: async (): Promise<DashboardData> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getDashboard()
    },
    enabled: initialized && !error && !!provider && !authLoading && isAuthenticated,
    staleTime: 1000 * 60, // 1 minute
  })
}

export function useDashboardStats() {
  const { provider, initialized, error } = useDataProviderStatus()
  const { data: auth, isLoading: authLoading } = useAuthStatus()
  const isAuthenticated = auth?.authenticated ?? false

  return useQuery({
    queryKey: ['data', 'dashboard', 'stats'],
    queryFn: async (): Promise<DashboardStats> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getDashboardStats()
    },
    enabled: initialized && !error && !!provider && !authLoading && isAuthenticated,
  })
}

export function useDashboardConfig() {
  const { provider, initialized, error } = useDataProviderStatus()
  const { data: auth, isLoading: authLoading } = useAuthStatus()
  const isAuthenticated = auth?.authenticated ?? false

  return useQuery({
    queryKey: ['data', 'dashboard', 'config'],
    queryFn: async () => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getDashboardConfig()
    },
    enabled: initialized && !error && !!provider && !authLoading && isAuthenticated,
  })
}

export function useUpdateDashboardConfig() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (config: DashboardConfig) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateDashboardConfig(config)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'dashboard', 'config'] })
    },
  })
}

export function useYearlyStats() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'yearly'],
    queryFn: async (): Promise<YearlyStat[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getYearlyStats()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useMonthlyStats(year?: number) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'monthly', year],
    queryFn: async (): Promise<MonthlyStat[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getMonthlyStats(year)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useSportTypeStats() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'sportTypes'],
    queryFn: async (): Promise<SportTypeStat[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getSportTypeStats()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useWeeklyStats() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'dashboard', 'weekly'],
    queryFn: async (): Promise<WeeklyStat[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getWeeklyStats()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useRecentActivities(limit = 5) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'dashboard', 'recent', limit],
    queryFn: async (): Promise<RecentActivity[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getRecentActivities(limit)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useCalendarData(year: number) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'calendar', year],
    queryFn: async (): Promise<CalendarDay[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getCalendarData(year)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useCalendarDataRange(startDate: string, endDate: string) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'calendar', 'range', startDate, endDate],
    queryFn: async (): Promise<CalendarDay[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getCalendarDataRange(startDate, endDate)
    },
    enabled: initialized && !error && !!provider && !!startDate && !!endDate,
  })
}

export function useCalendarActivities(year: number, month: number) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'calendar', year, month, 'activities'],
    queryFn: async (): Promise<CalendarActivity[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getCalendarActivities(year, month)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useCalendarSummary(year: number, month: number) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'calendar', year, month, 'summary'],
    queryFn: async (): Promise<CalendarMonthSummary> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getCalendarSummary(year, month)
    },
    enabled: initialized && !error && !!provider,
  })
}

// ============================================================================
// Stats & Training
// ============================================================================

export function usePowerStats(filters?: { after?: string; before?: string; sport_type?: string }) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'power', filters],
    queryFn: async (): Promise<PowerStatsResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getPowerStats(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function usePowerZones() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'powerZones'],
    queryFn: async (): Promise<PowerZonesResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getPowerZones()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useHrZones(filters?: { after?: string; before?: string; sport_type?: string }) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'hrZones', filters],
    queryFn: async (): Promise<HrZonesResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getHrZones(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useTrainingLoad(filters?: { after?: string; before?: string }) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'trainingLoad', filters],
    queryFn: async (): Promise<TrainingLoadResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getTrainingLoad(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useZoneTrend(weeks = 52) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'zoneTrend', weeks],
    queryFn: async (): Promise<ZoneTrendResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getZoneTrend(weeks)
    },
    enabled: initialized && !error && !!provider,
    staleTime: 1000 * 60 * 5, // 5 minutes - zone data is expensive to compute
  })
}

/**
 * Hook to fetch weekly trends data for trend analysis widgets.
 * Returns rolling N weeks of activity statistics with optional sport type filter.
 */
export function useWeeklyTrends(filters?: { weeks?: number; sport_type?: string }) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'weeklyTrends', filters?.weeks ?? 12, filters?.sport_type ?? ''],
    queryFn: async (): Promise<WeeklyTrendsResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getWeeklyTrends(filters)
    },
    enabled: initialized && !error && !!provider,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

/**
 * Hook to fetch monthly comparison data for cross-year analysis.
 * Returns monthly aggregated stats grouped by year and month with available years list.
 */
export function useMonthlyComparison(filters?: { sport_type?: string }) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'monthlyComparison', filters?.sport_type ?? ''],
    queryFn: async (): Promise<MonthlyComparisonResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getMonthlyComparison(filters)
    },
    enabled: initialized && !error && !!provider,
    staleTime: 1000 * 60 * 5, // 5 minutes
  })
}

/**
 * Hook to fetch coaching insights based on training load and activity patterns.
 * Returns insights about fatigue, recovery, streaks, and fitness trends.
 */
export function useInsights() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'insights'],
    queryFn: async (): Promise<InsightsResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getInsights()
    },
    enabled: initialized && !error && !!provider,
    staleTime: STALE_TIME.LONG,
  })
}

export function useDaytimeDistribution() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'daytime'],
    queryFn: async (): Promise<DistributionSlice[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getDaytimeDistribution()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useWeekdayDistribution() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'stats', 'weekday'],
    queryFn: async (): Promise<DistributionSlice[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getWeekdayDistribution()
    },
    enabled: initialized && !error && !!provider,
  })
}

// ============================================================================
// Challenges
// ============================================================================

export function useChallenges(filters?: ChallengesFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'challenges', filters],
    queryFn: async (): Promise<Challenge[]> => {
      if (!provider) throw new Error('Provider not ready')
      const result = await provider.getChallenges(filters)
      return result.data
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useChallengesPaginated(filters?: ChallengesFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'challenges', 'paginated', filters],
    queryFn: async (): Promise<ChallengesResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getChallenges(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

// ============================================================================
// Training Goals
// ============================================================================

export function useTrainingGoals() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'goals'],
    queryFn: async (): Promise<TrainingGoalsResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getTrainingGoals()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useUpdateTrainingGoals() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (config: TrainingGoalsConfig) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateTrainingGoals(config)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'goals'] })
    },
  })
}

// ============================================================================
// Activities
// ============================================================================

export function useActivities(filters: ActivityFilters = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'activities', filters],
    queryFn: async (): Promise<ActivitiesResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getActivities(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useActivity(id: number) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'activity', id],
    queryFn: async (): Promise<Activity> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getActivity(id)
    },
    enabled: id > 0 && initialized && !error && !!provider,
  })
}

export function useActivityStreams(id: number, enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'activity', id, 'streams'],
    queryFn: async (): Promise<ActivityStream[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getActivityStreams(id)
    },
    enabled: enabled && id > 0 && initialized && !error && !!provider,
  })
}

export function useActivityWeather(id: number, enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'activity', id, 'weather'],
    queryFn: async (): Promise<ActivityWeather | null> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getActivityWeather(id)
    },
    enabled: enabled && id > 0 && initialized && !error && !!provider,
    staleTime: Infinity, // Weather data doesn't change, cache forever
  })
}

export function useActivityAnalysis(id: number, splitUnit?: 'km' | 'mi', enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'activity', id, 'analysis', splitUnit],
    queryFn: async (): Promise<ActivityAnalysis> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getActivityAnalysis(id, splitUnit)
    },
    enabled: enabled && id > 0 && initialized && !error && !!provider,
    staleTime: STALE_TIME.MEDIUM, // Analysis is computed from streams, can be cached
  })
}

// ============================================================================
// Heatmap & Eddington
// ============================================================================

export function useHeatmap(filters: HeatmapFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'heatmap', filters],
    queryFn: async (): Promise<HeatmapResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getHeatmapData(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useEddington(sportType?: string) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'eddington', sportType],
    queryFn: async (): Promise<EddingtonResult> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getEddingtonData(sportType)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useEddingtonHistory(sportType?: string) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'eddington', 'history', sportType],
    queryFn: async (): Promise<EddingtonHistoryPoint[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getEddingtonHistory(sportType)
    },
    enabled: initialized && !error && !!provider,
  })
}

// ============================================================================
// Import
// ============================================================================

export function useImportProgress(enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()
  const { data: auth, isLoading: authLoading } = useAuthStatus()
  const isAuthenticated = auth?.authenticated ?? false

  return useQuery({
    queryKey: ['data', 'import', 'progress'],
    queryFn: async (): Promise<ImportProgress> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getImportProgress()
    },
    enabled: enabled && initialized && !error && !!provider && !authLoading && isAuthenticated,
    // Data is updated via events from useDataEvents hook
    staleTime: Infinity,
  })
}

export function useStartImport() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req?: StartImportRequest) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.startImport(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'import', 'progress'] })
      // Also refresh auth status since successful import start confirms valid auth
      queryClient.invalidateQueries({ queryKey: ['data', 'auth', 'status'] })
    },
  })
}

export function useCancelImport() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.cancelImport()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'import', 'progress'] })
    },
  })
}

export function usePauseImport() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.pauseImport()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'import', 'progress'] })
    },
  })
}

export function useResumeImport() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.resumeImport()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'import', 'progress'] })
    },
  })
}

export function useHasResumableImport() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'import', 'resumable'],
    queryFn: async (): Promise<boolean> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.hasResumableImport()
    },
    enabled: initialized && !error && !!provider,
    staleTime: 1000 * 5, // 5 seconds
  })
}

export function useSyncHistory(limit = 10, opts: { enabled?: boolean } = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'import', 'history', limit],
    queryFn: async () => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getSyncHistory(limit)
    },
    enabled: (opts.enabled ?? true) && initialized && !error && !!provider,
  })
}

export function useLatestSync() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'import', 'latest'],
    queryFn: async () => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getLatestSync()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useSyncWatermark(enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'import', 'watermark'],
    queryFn: async (): Promise<SyncWatermark | null> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getSyncWatermark()
    },
    enabled: enabled && initialized && !error && !!provider,
  })
}

export function useResetSyncWatermark() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async () => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.resetSyncWatermark()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'import', 'watermark'] })
    },
    onError: (error) => {
      console.error('Failed to reset sync watermark:', error)
    },
  })
}

// ============================================================================
// Settings
// ============================================================================

export function useAppSettings(opts: { enabled?: boolean } = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'settings'],
    queryFn: async (): Promise<AppSettings> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getAppSettings()
    },
    enabled: (opts.enabled ?? true) && initialized && !error && !!provider,
    retry: false,
  })
}

export function useUpdateAppSettings() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (settings: AppSettings) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateAppSettings(settings)
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['data', 'settings'], data)
    },
  })
}

// ============================================================================
// Athlete (FTP/Weight)
// ============================================================================

export function useFtpHistory() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'athlete', 'ftp'],
    queryFn: async (): Promise<FTPHistoryResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getFtpHistory()
    },
    enabled: initialized && !error && !!provider,
    retry: false,
  })
}

export function useUpdateFtpHistory() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (body: FTPHistoryResponse) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateFtpHistory(body)
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['data', 'athlete', 'ftp'], data)
    },
  })
}

export function useWeightHistory() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'athlete', 'weight'],
    queryFn: async (): Promise<WeightHistoryResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getWeightHistory()
    },
    enabled: initialized && !error && !!provider,
    retry: false,
  })
}

export function useUpdateWeightHistory() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (body: WeightHistoryResponse) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateWeightHistory(body)
    },
    onSuccess: (data) => {
      queryClient.setQueryData(['data', 'athlete', 'weight'], data)
    },
  })
}

// ============================================================================
// HR Zone Definitions
// ============================================================================

export function useHrZoneDefinitions(opts: { enabled?: boolean } = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'zones', 'hr', 'definitions'],
    queryFn: async (): Promise<HrZoneDefinition[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getHrZoneDefinitions()
    },
    enabled: (opts.enabled ?? true) && initialized && !error && !!provider,
    retry: false,
  })
}

export function useUpsertHrZoneDefinition() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (def: HrZoneDefinition) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.upsertHrZoneDefinition(def)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'zones', 'hr', 'definitions'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'stats', 'hrZones'] })
    },
  })
}

export function useDeleteHrZoneDefinition() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: { sport_type: string; effective_from: string }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.deleteHrZoneDefinition(params.sport_type, params.effective_from)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'zones', 'hr', 'definitions'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'stats', 'hrZones'] })
    },
  })
}

// ============================================================================
// Gear
// ============================================================================

export function useGear(filters?: GearFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', filters],
    queryFn: async (): Promise<Gear[]> => {
      if (!provider) throw new Error('Provider not ready')
      const result = await provider.getGear(filters)
      return result.data
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useGearPaginated(filters?: GearFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', 'paginated', filters],
    queryFn: async (): Promise<GearListResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getGear(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useGearDetail(id: string | null) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', 'detail', id],
    queryFn: async (): Promise<Gear> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getGearDetail(id!)
    },
    enabled: !!id && initialized && !error && !!provider,
  })
}

export function useCustomGear(filters?: GearFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', 'custom', filters],
    queryFn: async (): Promise<Gear[]> => {
      if (!provider) throw new Error('Provider not ready')
      const result = await provider.getCustomGear(filters)
      return result.data
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useCustomGearPaginated(filters?: GearFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', 'custom', 'paginated', filters],
    queryFn: async (): Promise<GearListResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getCustomGear(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useCreateCustomGear() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: CustomGearCreateRequest) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.createCustomGear(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'gear'] })
    },
  })
}

export function useUpdateCustomGear() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, patch }: { id: string; patch: Partial<CustomGearCreateRequest> }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateCustomGear(id, patch)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'gear'] })
    },
  })
}

export function useDeleteCustomGear() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ id, force }: { id: string; force?: boolean }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.deleteCustomGear(id, force)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'gear'] })
    },
  })
}

/**
 * Hook to update purchase price and currency for any gear item.
 * Works for both Strava-imported and custom gear.
 *
 * @returns Mutation for updating gear price. Pass null for price to clear it.
 *
 * @example
 * const updatePrice = useUpdateGearPrice()
 * updatePrice.mutate({ id: 'b12345', price: 499.99, currency: 'USD' })
 * updatePrice.mutate({ id: 'b12345', price: null, currency: '' }) // Clear price
 */
export function useUpdateGearPrice() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      id,
      price,
      currency,
    }: {
      id: string
      price: number | null
      currency: string
    }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateGearPrice(id, price, currency)
    },
    onSuccess: (_, { id }) => {
      // Invalidate all gear queries (list, custom, monthly usage, etc.)
      queryClient.invalidateQueries({ queryKey: ['data', 'gear'] })
      // Also invalidate the specific gear detail query to ensure it refreshes
      queryClient.invalidateQueries({ queryKey: ['data', 'gear', 'detail', id] })
    },
  })
}

export function useGearMonthlyUsage(includeRetired = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', 'monthly', includeRetired],
    queryFn: async (): Promise<GearMonthlyUsage[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getGearMonthlyUsage(includeRetired)
    },
    enabled: initialized && !error && !!provider,
  })
}

// ============================================================================
// Segments
// ============================================================================

export function useSegments(filters: SegmentsFilters = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'segments', filters],
    queryFn: async (): Promise<SegmentsResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getSegments(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useSegmentCountries(enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'segments', 'countries'],
    queryFn: async (): Promise<SegmentCountryStat[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getSegmentCountries()
    },
    enabled: enabled && initialized && !error && !!provider,
  })
}

export function useSegmentDetail(id: number | null, enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'segments', 'detail', id],
    queryFn: async (): Promise<SegmentDetailResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getSegmentDetail(id!)
    },
    enabled: enabled && !!id && initialized && !error && !!provider,
  })
}

export function useSegmentEfforts(id: number | null, filters?: SegmentEffortsFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'segments', 'efforts', id, filters],
    queryFn: async (): Promise<SegmentEffort[]> => {
      if (!provider) throw new Error('Provider not ready')
      const result = await provider.getSegmentEfforts(id!, filters)
      return result.data
    },
    enabled: !!id && initialized && !error && !!provider,
  })
}

export function useSegmentEffortsPaginated(id: number | null, filters?: SegmentEffortsFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'segments', 'efforts', 'paginated', id, filters],
    queryFn: async (): Promise<SegmentEffortsResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getSegmentEfforts(id!, filters)
    },
    enabled: !!id && initialized && !error && !!provider,
  })
}

// ============================================================================
// Maintenance
// ============================================================================

export function useMaintenanceDue(enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'maintenance', 'due'],
    queryFn: async (): Promise<DueComponent[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getMaintenanceDue()
    },
    enabled: enabled && initialized && !error && !!provider,
  })
}

export function useGearComponents(gearId: string | null, filters?: ComponentsFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', 'components', gearId, filters],
    queryFn: async (): Promise<ComponentWithRules[]> => {
      if (!provider) throw new Error('Provider not ready')
      const result = await provider.getGearComponents(gearId!, filters)
      return result.data
    },
    enabled: !!gearId && initialized && !error && !!provider,
  })
}

export function useGearComponentsPaginated(gearId: string | null, filters?: ComponentsFilters) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'gear', 'components', 'paginated', gearId, filters],
    queryFn: async (): Promise<ComponentsResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getGearComponents(gearId!, filters)
    },
    enabled: !!gearId && initialized && !error && !!provider,
  })
}

export function useCreateComponent() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: { gearId: string; body: CreateComponentRequest }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.createComponent(params.gearId, params.body)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'gear', 'components'] })
    },
  })
}

export function useUpdateComponent() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: { id: number; req: UpdateComponentRequest }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateComponent(params.id, params.req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'gear', 'components'] })
    },
  })
}

export function useDeleteComponent() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: { id: number }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.deleteComponent(params.id)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'gear', 'components'] })
    },
  })
}

export function useLogMaintenance() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: { componentId: number; body?: LogMaintenanceRequest }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.logMaintenance(params.componentId, params.body)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'gear', 'components'] })
    },
  })
}

// ============================================================================
// Photos
// ============================================================================

export function usePhotos(filters: PhotosFilters = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'photos', filters],
    queryFn: async (): Promise<PhotosListResponse> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getPhotos(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useActivityPhotos(activityId: number | null, enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'photos', 'activity', activityId],
    queryFn: async (): Promise<ActivityPhoto[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getActivityPhotos(activityId!)
    },
    enabled: enabled && !!activityId && activityId > 0 && initialized && !error && !!provider,
  })
}

// ============================================================================
// Best Efforts
// ============================================================================

export function useBestEffortPRs(sportType?: string) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'bestEfforts', 'prs', sportType],
    queryFn: async (): Promise<BestEffortPR[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getBestEffortPRs(sportType)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useBestEffortsForDistance(
  distanceType: string,
  sportType?: string,
  enabled = true
) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'bestEfforts', distanceType, sportType],
    queryFn: async (): Promise<BestEffortItem[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getBestEffortsForDistance(distanceType, sportType)
    },
    enabled: enabled && !!distanceType && initialized && !error && !!provider,
  })
}

// ============================================================================
// Wrapped
// ============================================================================

export function useWrappedYears() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'wrapped', 'years'],
    queryFn: async (): Promise<number[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getWrappedYears()
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useWrappedReport(year: number, enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'wrapped', year],
    queryFn: async (): Promise<WrappedReport> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getWrapped(year)
    },
    enabled: enabled && initialized && !error && !!provider,
  })
}

// ============================================================================
// Export
// ============================================================================

export function useExportStats() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'export', 'stats'],
    queryFn: async (): Promise<ExportStats> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getExportStats()
    },
    enabled: initialized && !error && !!provider,
  })
}

// ============================================================================
// Challenges (import)
// ============================================================================

export function useImportChallenges() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params: { file: File }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.importChallenges(params.file)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'challenges'] })
    },
  })
}

export function useImportChallengesFromProfile() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (params?: { athleteId?: string }) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.importChallengesFromProfile(params?.athleteId)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['data', 'challenges'] })
    },
  })
}

// ============================================================================
// Setup (Strava Credentials)
// ============================================================================

export function useCredentialsStatus(opts: { enabled?: boolean } = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'setup', 'credentials'],
    queryFn: async (): Promise<CredentialsStatus> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getCredentialsStatus()
    },
    enabled: (opts.enabled ?? true) && initialized && !error && !!provider,
    staleTime: 1000 * 60 * 5, // 5 minutes
    retry: false,
  })
}

export function useUpdateCredentials() {
  const { provider, initialized } = useDataProviderStatus()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (req: UpdateCredentialsRequest) => {
      if (!provider || !initialized) throw new Error('Provider not ready')
      return provider.updateCredentials(req)
    },
    onSuccess: (data) => {
      // Set cache directly with mutation result to avoid race condition with refetch
      queryClient.setQueryData(['data', 'setup', 'credentials'], data)
    },
  })
}
