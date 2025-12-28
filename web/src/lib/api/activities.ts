import { useQuery } from '@tanstack/react-query'
import { get } from './client'
import type { ActivitiesResponse, Activity, ActivityFilters } from './types'

export const activityKeys = {
  all: ['activities'] as const,
  lists: () => [...activityKeys.all, 'list'] as const,
  list: (filters: ActivityFilters) => [...activityKeys.lists(), filters] as const,
  details: () => [...activityKeys.all, 'detail'] as const,
  detail: (id: number) => [...activityKeys.details(), id] as const,
  streams: (id: number) => [...activityKeys.all, 'streams', id] as const,
}

export function useActivities(filters: ActivityFilters = {}) {
  return useQuery({
    queryKey: activityKeys.list(filters),
    queryFn: () =>
      get<ActivitiesResponse>('/activities', {
        sport_type: filters.sport_type,
        after: filters.after,
        before: filters.before,
        gear_id: filters.gear_id,
        search: filters.search,
        commute: filters.commute,
        trainer: filters.trainer,
        page: filters.page,
        per_page: filters.per_page,
        order_by: filters.order_by,
        order_dir: filters.order_dir,
      }),
  })
}

export function useActivity(id: number) {
  return useQuery({
    queryKey: activityKeys.detail(id),
    queryFn: () => get<Activity>(`/activities/${id}`),
    enabled: id > 0,
  })
}

export interface ActivityStream {
  activity_id: number
  stream_type: string
  original_size: number
  resolution: string
  series_type: string
  data: unknown
}

export function useActivityStreams(id: number) {
  return useQuery({
    queryKey: activityKeys.streams(id),
    queryFn: () => get<ActivityStream[]>(`/activities/${id}/streams`),
    enabled: id > 0,
  })
}
