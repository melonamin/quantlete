import { useQuery } from '@tanstack/react-query'
import { useDataProviderStatus } from '@/lib/data/context'
import type { ActivitiesResponse, Activity, ActivityFilters } from './types'

export const activityKeys = {
  all: ['activities'] as const,
  lists: () => [...activityKeys.all, 'list'] as const,
  list: (filters: ActivityFilters) => [...activityKeys.lists(), filters] as const,
  details: () => [...activityKeys.all, 'detail'] as const,
  detail: (id: number) => [...activityKeys.details(), id] as const,
  streams: (id: number) => [...activityKeys.all, 'streams', id] as const,
}

export interface ActivityStream {
  activity_id: number
  stream_type: string
  original_size: number
  resolution: string
  series_type: string
  data: unknown
}

export function useActivities(filters: ActivityFilters = {}) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: activityKeys.list(filters),
    queryFn: async (): Promise<ActivitiesResponse> => {
      if (!provider) throw new Error('Data provider not ready')
      return provider.getActivities(filters)
    },
    enabled: initialized && !error && !!provider,
  })
}

export function useActivity(id: number) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: activityKeys.detail(id),
    queryFn: async (): Promise<Activity> => {
      if (!provider) throw new Error('Data provider not ready')
      return provider.getActivity(id)
    },
    enabled: id > 0 && initialized && !error && !!provider,
  })
}

export function useActivityStreams(id: number) {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: activityKeys.streams(id),
    queryFn: async (): Promise<ActivityStream[]> => {
      if (!provider) throw new Error('Data provider not ready')
      return provider.getActivityStreams(id)
    },
    enabled: id > 0 && initialized && !error && !!provider,
  })
}
