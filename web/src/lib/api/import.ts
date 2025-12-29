import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { get, post } from './client'

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
}

export function useImportProgress(enabled = true) {
  return useQuery({
    queryKey: importKeys.progress,
    queryFn: () => get<ImportProgress>('/import/progress'),
    refetchInterval: (query) => {
      const data = query.state.data
      if (data?.status === 'running') {
        return 1000 // Poll every second while running
      }
      return false
    },
    enabled,
  })
}

export function useStartImport() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (req: StartImportRequest = {}) => post<{ message: string }>('/import/start', req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: importKeys.progress })
    },
  })
}

export function useCancelImport() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => post<{ message: string }>('/import/cancel'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: importKeys.progress })
    },
  })
}
