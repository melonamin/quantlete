// API Response types

export interface Activity {
  id: number
  name: string
  description?: string
  sport_type: string
  start_date: string
  start_date_local: string
  timezone?: string
  distance: number
  moving_time: number
  elapsed_time: number
  total_elevation_gain: number
  elev_high?: number
  elev_low?: number
  average_speed: number
  max_speed: number
  average_heartrate?: number
  max_heartrate?: number
  average_watts?: number
  max_watts?: number
  weighted_average_watts?: number
  kilojoules?: number
  average_cadence?: number
  calories?: number
  kudos_count: number
  comment_count: number
  photo_count: number
  commute: boolean
  private: boolean
  trainer: boolean
  workout_type?: number
  device_name?: string
  gear_id?: string
  start_lat?: number
  start_lng?: number
  end_lat?: number
  end_lng?: number
  summary_polyline?: string
  location_city?: string
  location_country?: string
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export type ActivitiesResponse = PaginatedResponse<Activity>

export interface ActivityFilters {
  sport_type?: string
  after?: string
  before?: string
  gear_id?: string
  search?: string
  commute?: boolean
  trainer?: boolean
  page?: number
  per_page?: number
  order_by?: string
  order_dir?: 'asc' | 'desc'
}

export interface AuthStatus {
  authenticated: boolean
  demo_mode?: boolean
  athlete?: {
    id: number
    username: string
    firstname: string
    lastname: string
    profile: string
    city?: string
    country?: string
  }
}

export interface ErrorResponse {
  error: string
}
