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
  // Pagination
  page?: number
  per_page?: number
  // Sorting
  order_by?: 'name' | 'distance' | 'maximum_grade' | 'times_completed' | 'last_effort_date' | 'best_elapsed_time'
  order_dir?: 'asc' | 'desc'
}

export interface SegmentsResponse {
  data: SegmentListItem[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export interface SegmentDetailResponse {
  segment: Segment
  efforts: SegmentEffort[]
}

export interface SegmentEffortsFilters {
  page?: number
  per_page?: number
  order_by?: 'start_date' | 'elapsed_time'
  order_dir?: 'asc' | 'desc'
}

export interface SegmentEffortsResponse {
  data: SegmentEffort[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export {
  useSegments,
  useSegmentCountries,
  useSegmentDetail,
  useSegmentEfforts,
} from '@/lib/data'
