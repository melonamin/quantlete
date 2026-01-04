/**
 * useDataEvents Hook
 *
 * Subscribes to data events from the DataProvider and invalidates
 * TanStack Query caches accordingly. This enables reactive UI updates
 * without polling.
 */

import { useEffect, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useDataProviderStatus } from '@/lib/data/context'
import type { DataEvent, DataChangeSet } from '@/lib/data/events'

/**
 * Hook that subscribes to data events and handles query invalidation.
 * Should be called once at the app root level.
 */
export function useDataEvents(): void {
  const queryClient = useQueryClient()
  const { provider, initialized, error } = useDataProviderStatus()

  const invalidateForChanges = useCallback(
    (changes: DataChangeSet) => {
      if (changes.all) {
        queryClient.invalidateQueries({ queryKey: ['data'] })
        return
      }
      if (changes.activities) {
        queryClient.invalidateQueries({ queryKey: ['data', 'activities'] })
        queryClient.invalidateQueries({ queryKey: ['data', 'dashboard'] })
        queryClient.invalidateQueries({ queryKey: ['data', 'calendar'] })
        queryClient.invalidateQueries({ queryKey: ['data', 'stats'] })
      }
      if (changes.streams) {
        queryClient.invalidateQueries({ queryKey: ['data', 'stats', 'power'] })
        queryClient.invalidateQueries({ queryKey: ['data', 'stats', 'trainingLoad'] })
      }
      if (changes.segments) {
        queryClient.invalidateQueries({ queryKey: ['data', 'segments'] })
      }
      if (changes.gear) {
        queryClient.invalidateQueries({ queryKey: ['data', 'gear'] })
      }
      if (changes.photos) {
        queryClient.invalidateQueries({ queryKey: ['data', 'photos'] })
      }
    },
    [queryClient]
  )

  useEffect(() => {
    if (!initialized || error || !provider) {
      return
    }

    const handleEvent = (event: DataEvent) => {
      switch (event.type) {
        case 'sync:progress':
          // Update import progress cache directly (avoids refetch)
          queryClient.setQueryData(['data', 'import', 'progress'], (old: unknown) => ({
            ...(old as object),
            status: 'running',
            phase: event.phase,
            activities_done: event.activitiesDone,
            activities_total: event.activitiesTotal,
            gear_done: event.gearDone,
            gear_total: event.gearTotal,
            streams_done: event.streamsDone,
            streams_total: event.streamsTotal,
            details_done: event.detailsDone,
            details_total: event.detailsTotal,
            segments_done: event.segmentsDone,
            segments_total: event.segmentsTotal,
            photos_done: event.photosDone,
            photos_total: event.photosTotal,
            estimated_eta: event.estimatedEta,
            waiting_for_rate_limit: event.waitingForRateLimit,
            waiting_until: event.waitingUntil,
            waiting_reason: event.waitingReason,
          }))
          break

        case 'sync:complete':
          // Update status in cache
          queryClient.setQueryData(['data', 'import', 'progress'], (old: unknown) => ({
            ...(old as object),
            status: event.status === 'completed' ? 'completed' : event.status,
            error: event.error,
          }))
          // Invalidate all data after sync completes
          invalidateForChanges({ all: true })
          // Also invalidate sync history
          queryClient.invalidateQueries({ queryKey: ['data', 'import', 'history'] })
          break

        case 'data:changed':
          invalidateForChanges(event.changes)
          break
      }
    }

    return provider.subscribeToEvents(handleEvent)
  }, [provider, initialized, error, queryClient, invalidateForChanges])
}
