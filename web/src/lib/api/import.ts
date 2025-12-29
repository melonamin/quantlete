import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDataProviderStatus } from '@/lib/data/context'

// Sync run represents a single import/sync history entry.
export interface SyncRun {
  id: number
  athlete_id: number
  started_at: string
  completed_at?: string
  duration_seconds?: number

  status: 'running' | 'completed' | 'failed' | 'canceled'
  error?: string

  // Counts
  activities_total: number
  activities_imported: number
  activities_skipped: number
  gear_imported: number
  streams_imported: number
  segments_imported: number
  photos_imported: number
  failed_count: number

  // Options
  full_sync: boolean
  skip_streams: boolean
  skip_segments: boolean
  skip_best_efforts: boolean
  skip_photos: boolean

  // Watermark
  newest_activity_date?: string

  created_at: string
}

// Sync watermark for incremental sync.
export interface SyncWatermark {
  last_synced_at: string
  newest_activity_date?: string
}

// Import phases in order of execution
export type ImportPhase =
  | 'idle'
  | 'activities'
  | 'gear'
  | 'streams'
  | 'activity_details'
  | 'segment_details'
  | 'photos'
  | 'completed'

export interface ImportProgress {
  status: 'idle' | 'running' | 'completed' | 'failed' | 'canceled'
  started_at?: string
  completed_at?: string
  error?: string

  // Current phase
  phase: ImportPhase

  // Per-phase progress
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

  // Legacy fields for backward compatibility
  total_activities: number
  imported_count: number
  skipped_count: number
  failed_count: number
  current_page: number

  // ETA estimation
  remaining_api_calls: number
  estimated_eta?: string

  // Rate limit info
  rate_limit_used_15min: number
  rate_limit_limit_15min: number
  rate_limit_used_daily: number
  rate_limit_limit_daily: number

  // Rate limit waiting state
  waiting_for_rate_limit: boolean
  waiting_until?: string
  waiting_reason?: string
}

// By default, all data types are imported. Use skip_* to exclude specific types.
export interface StartImportRequest {
  full_sync?: boolean
  resume?: boolean
  skip_streams?: boolean
  skip_segments?: boolean
  skip_best_efforts?: boolean
  skip_photos?: boolean
}

export const importKeys = {
  progress: ['import', 'progress'] as const,
  history: ['import', 'history'] as const,
  watermark: ['import', 'watermark'] as const,
}

export function useImportProgress(enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: importKeys.progress,
    queryFn: async (): Promise<ImportProgress> => {
      if (!provider) throw new Error('Data provider not ready')
      return provider.getImportProgress()
    },
    refetchInterval: (query) => {
      const data = query.state.data
      if (data?.status === 'running') {
        return 1000 // Poll every second while running
      }
      return false
    },
    enabled: enabled && initialized && !error && !!provider,
  })
}

export function useStartImport() {
  const queryClient = useQueryClient()
  const { provider, initialized } = useDataProviderStatus()

  return useMutation({
    mutationFn: async (req: StartImportRequest = {}) => {
      if (!provider || !initialized) throw new Error('Data provider not ready')
      return provider.startImport(req)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: importKeys.progress })
    },
  })
}

export function useCancelImport() {
  const queryClient = useQueryClient()
  const { provider, initialized } = useDataProviderStatus()

  return useMutation({
    mutationFn: async () => {
      if (!provider || !initialized) throw new Error('Data provider not ready')
      return provider.cancelImport()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: importKeys.progress })
    },
  })
}

export function useSyncHistory(enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: importKeys.history,
    queryFn: async (): Promise<SyncRun[]> => {
      if (!provider) throw new Error('Data provider not ready')
      return provider.getSyncHistory()
    },
    enabled: enabled && initialized && !error && !!provider,
  })
}

export function useSyncWatermark(enabled = true) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: importKeys.watermark,
    queryFn: async (): Promise<SyncWatermark | null> => {
      if (!provider) throw new Error('Data provider not ready')
      return provider.getSyncWatermark()
    },
    enabled: enabled && initialized && !error && !!provider,
  })
}
