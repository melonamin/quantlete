import { useQuery } from '@tanstack/react-query'
import { get } from './client'

export interface PhotoListItem {
  id: string
  activity_id: number
  url: string
  thumbnail_url?: string
  caption?: string
  created_at: string
  activity_name: string
  sport_type: string
  start_date_local: string
  location_country?: string
}

export interface PhotosFacet {
  value: string
  iso2?: string
  count: number
}

export interface PhotosListResponse {
  data: PhotoListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
  countries: PhotosFacet[]
  sport_types: PhotosFacet[]
}

export interface PhotosFilters {
  sport_type?: string
  country?: string
  page?: number
  per_page?: number
}

export interface ActivityPhoto {
  id: string
  athlete_id: number
  activity_id: number
  url: string
  thumbnail_url?: string
  caption?: string
  location?: unknown
  created_at: string
}

export const photoKeys = {
  all: ['photos'] as const,
  list: (filters: PhotosFilters) => [...photoKeys.all, 'list', filters] as const,
  activity: (activityId: number) => [...photoKeys.all, 'activity', activityId] as const,
}

export function usePhotos(filters: PhotosFilters = {}) {
  return useQuery({
    queryKey: photoKeys.list(filters),
    queryFn: () =>
      get<PhotosListResponse>('/photos', {
        sport_type: filters.sport_type,
        country: filters.country,
        page: filters.page,
        per_page: filters.per_page,
      }),
  })
}

export function useActivityPhotos(activityId: number, enabled = true) {
  return useQuery({
    queryKey: photoKeys.activity(activityId),
    queryFn: () => get<ActivityPhoto[]>(`/activities/${activityId}/photos`),
    enabled: enabled && activityId > 0,
  })
}
