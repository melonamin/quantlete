/**
 * Hook to automatically invalidate queries when sync imports new data.
 * Makes the UI reactive to database changes during sync.
 */

import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useImportProgress } from '@/lib/api'

export function useSyncInvalidation() {
  const queryClient = useQueryClient()
  const { data: progress } = useImportProgress()

  // Track previous values to detect changes
  const prevRef = useRef<{
    activitiesDone: number
    streamsDone: number
    detailsDone: number
    status: string
  } | null>(null)

  useEffect(() => {
    if (!progress) return

    const prev = prevRef.current
    const current = {
      activitiesDone: progress.activities_done,
      streamsDone: progress.streams_done,
      detailsDone: progress.details_done,
      status: progress.status,
    }

    // Skip first render
    if (!prev) {
      prevRef.current = current
      return
    }

    // Check if activities were imported
    if (current.activitiesDone > prev.activitiesDone) {
      // Invalidate dashboard and activity-related queries
      queryClient.invalidateQueries({ queryKey: ['data', 'dashboard'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'activities'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'calendar'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'stats'] })
    }

    // Check if streams were imported (affects training load, power stats)
    if (current.streamsDone > prev.streamsDone) {
      queryClient.invalidateQueries({ queryKey: ['data', 'training'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'power'] })
    }

    // Check if details were imported (affects segments)
    if (current.detailsDone > prev.detailsDone) {
      queryClient.invalidateQueries({ queryKey: ['data', 'segments'] })
    }

    // When sync completes, do a full refresh
    if (prev.status === 'running' && current.status === 'completed') {
      // Invalidate all data queries
      queryClient.invalidateQueries({ queryKey: ['data'] })
      queryClient.invalidateQueries({ queryKey: ['sync'] })
    }

    prevRef.current = current
  }, [progress, queryClient])
}
