import { useEffect, useRef } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useImportProgress } from '@/lib/api'

/**
 * SyncRefresh watches the import progress and automatically invalidates
 * cached data when a sync completes so pages refresh without manual reloads.
 */
export function SyncRefresh() {
  const queryClient = useQueryClient()
  const { data: progress } = useImportProgress()
  const prevStatus = useRef(progress?.status)
  const lastCompletionToken = useRef(progress?.completed_at ?? null)

  useEffect(() => {
    const prev = prevStatus.current
    const current = progress?.status

    if (
      prev === 'running' &&
      current === 'completed' &&
      progress?.completed_at &&
      lastCompletionToken.current !== progress.completed_at
    ) {
      // Refresh every query that uses the shared data layer namespace.
      queryClient.invalidateQueries({
        predicate: (query) => Array.isArray(query.queryKey) && query.queryKey[0] === 'data',
      })

      lastCompletionToken.current = progress.completed_at
    }

    prevStatus.current = current
  }, [progress?.status, progress?.completed_at, queryClient])

  return null
}
