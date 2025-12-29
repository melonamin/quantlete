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

export { usePhotos, useActivityPhotos } from '@/lib/data'
