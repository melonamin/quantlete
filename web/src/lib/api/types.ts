// API Response types.
//
// These types define the shape expected by frontend components.
// They extend the generated query types with additional fields used by the UI.
//
// IMPORTANT: These types must be kept in sync with their counterparts in:
// - @/lib/wasm/types.gen.ts (auto-generated from Go structs)
// - Go storage types in internal/storage/*.go
//
// When modifying these types, verify they still match the actual API responses
// and update the generated types if the underlying Go structs have changed.

// Activity - extends generated GetActivityRow with additional UI fields
export interface Activity {
  id: number
  athlete_id: number
  name: string
  sport_type: string
  start_date: string
  start_date_local: string
  timezone: string
  distance: number
  moving_time: number
  elapsed_time: number
  total_elevation_gain: number
  average_speed: number
  max_speed: number
  average_heartrate?: number | null
  max_heartrate?: number | null
  average_watts?: number | null
  max_watts?: number | null
  weighted_average_watts?: number | null
  kilojoules?: number | null
  average_cadence?: number | null
  calories?: number | null
  suffer_score?: number | null
  gear_id: string
  commute: boolean
  workout_type?: number | null
  location_city: string
  location_state: string
  location_country: string
  summary_polyline: string
  start_lat?: number | null
  start_lng?: number | null
  description?: string
  device_name?: string
  embed_token?: string
  trainer?: boolean
  private?: boolean
  // Additional UI fields not in generated type
  elev_high?: number | null
  elev_low?: number | null
  kudos_count?: number
  comment_count?: number
  photo_count?: number
}

// Generic paginated response template - used across API
export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export type ActivitiesResponse = PaginatedResponse<Activity>

// ActivityFilters - manual with optional fields (generated GetActivitiesRequest has required fields)
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

// AuthStatus - manual (generated uses different structure)
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
