// API Response types - re-exported from generated types for single source of truth.

import type { ActivityResponse } from '@/lib/wasm/types.gen'

// Re-export generated types with API-friendly aliases
export type Activity = ActivityResponse

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
