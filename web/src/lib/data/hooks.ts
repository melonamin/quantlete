/**
 * React Query hooks that use the DataProvider.
 *
 * These hooks work in both server and WASM modes by delegating
 * to the appropriate DataProvider implementation.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDataProviderStatus } from './context'
import type {
  AuthStatus,
  ActivityFilters,
  ActivitiesResponse,
  Activity,
  ActivityStream,
  DashboardData,
  SportTypeStat,
  MonthlyStat,
  YearlyStat,
  CalendarDay,
  HeatmapFilters,
  HeatmapResponse,
  EddingtonResult,
  EddingtonHistoryPoint,
  DashboardConfig,
  ImportProgress,
  StartImportRequest,
  AppSettings,
  FTPHistoryResponse,
  WeightHistoryResponse,
  HrZoneDefinition,
  PowerStatsResponse,
  HrZonesResponse,
  TrainingLoadResponse,
  DistributionSlice,
  Challenge,
  TrainingGoalsConfig,
  TrainingGoalsResponse,
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
    staleTime: 1000 * 60 * 5, // 5 minutes
    retry: false,
  })
}

// ============================================================================
// Dashboard
// ============================================================================

export function useDashboard() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'dashboard'],
    queryFn: async (): Promise<DashboardData> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getDashboard()
    },
    enabled: initialized && !error && !!provider,
    staleTime: 1000 * 60, // 1 minute
  })
}

export function useDashboardConfig() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'dashboard', 'config'],
    queryFn: async () => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getDashboardConfig()
    },
    enabled: initialized && !error && !!provider,
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

export function useChallenges(month?: string) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'challenges', month],
    queryFn: async (): Promise<Challenge[]> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getChallenges(month)
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

export function useActivities(filters: ActivityFilters) {
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
    enabled: initialized && !error && !!provider,
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
    enabled: enabled && initialized && !error && !!provider,
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

export function useImportProgress() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: ['data', 'import', 'progress'],
    queryFn: async (): Promise<ImportProgress> => {
      if (!provider) throw new Error('Provider not ready')
      return provider.getImportProgress()
    },
    enabled: initialized && !error && !!provider,
    refetchInterval: (query) => {
      const data = query.state.data
      if (data?.status === 'running') {
        return 1000 // Poll every second while running
      }
      return false
    },
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
