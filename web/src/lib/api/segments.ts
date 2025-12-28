import { useQuery } from '@tanstack/react-query'
import { get } from './client'

export interface Segment {
  id: number
  name: string
  activity_type: string
  distance: number
  average_grade: number
  maximum_grade: number
  elevation_high: number
  elevation_low: number
  climb_category: number
  start_lat?: number
  start_lng?: number
  end_lat?: number
  end_lng?: number
  starred: boolean
  polyline?: string
  athlete_kom_rank?: number
  athlete_effort_count?: number
  athlete_pr_elapsed_time?: number
  athlete_pr_date?: string
}

export interface SegmentListItem extends Segment {
  times_completed: number
  last_effort_date?: string
  best_elapsed_time?: number
}

export interface SegmentEffort {
  id: number
  segment_id: number
  activity_id: number
  athlete_id: number
  name?: string
  elapsed_time: number
  moving_time: number
  start_date?: string
  start_date_local?: string
  distance: number
  average_watts?: number
  average_heartrate?: number
  max_heartrate?: number
  pr_rank?: number
  country?: string
}

export interface SegmentCountryStat {
  country: string
  iso2?: string
  count: number
}

export interface SegmentsFilters {
  activity_type?: string
  country?: string
  starred?: boolean
  kom_only?: boolean
  search?: string
  limit?: number
}

export interface SegmentDetailResponse {
  segment: Segment
  efforts: SegmentEffort[]
}

export const segmentKeys = {
  all: ['segments'] as const,
  list: (filters: SegmentsFilters) => [...segmentKeys.all, 'list', filters] as const,
  countries: () => [...segmentKeys.all, 'countries'] as const,
  detail: (id: number) => [...segmentKeys.all, 'detail', id] as const,
  efforts: (id: number) => [...segmentKeys.all, 'efforts', id] as const,
}

export function useSegments(filters: SegmentsFilters = {}) {
  return useQuery({
    queryKey: segmentKeys.list(filters),
    queryFn: () =>
      get<SegmentListItem[]>('/segments', {
        activity_type: filters.activity_type,
        country: filters.country,
        starred: filters.starred,
        kom_only: filters.kom_only,
        search: filters.search,
        limit: filters.limit,
      }),
  })
}

export function useSegmentCountries(enabled = true) {
  return useQuery({
    queryKey: segmentKeys.countries(),
    queryFn: () => get<SegmentCountryStat[]>('/segments/countries'),
    enabled,
  })
}

export function useSegmentDetail(id: number, enabled = true) {
  return useQuery({
    queryKey: segmentKeys.detail(id),
    queryFn: () => get<SegmentDetailResponse>(`/segments/${id}`),
    enabled: enabled && id > 0,
  })
}

export function useSegmentEfforts(id: number, enabled = true) {
  return useQuery({
    queryKey: segmentKeys.efforts(id),
    queryFn: () => get<SegmentEffort[]>(`/segments/${id}/efforts`),
    enabled: enabled && id > 0,
  })
}
