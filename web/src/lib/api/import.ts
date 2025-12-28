import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { get, post } from './client'

export interface ImportProgress {
  status: 'idle' | 'running' | 'completed' | 'failed' | 'canceled'
  started_at?: string
  completed_at?: string
  total_activities: number
  imported_count: number
  skipped_count: number
  failed_count: number
  current_page: number
  error?: string
}

export interface StartImportRequest {
  full_sync?: boolean
  include_streams?: boolean
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
    mutationFn: (req: StartImportRequest = {}) =>
      post<{ message: string }>('/import/start', req),
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
