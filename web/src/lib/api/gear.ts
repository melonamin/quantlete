// Types for gear data structures.
//
// Type Provenance:
// - Generated types: Re-exported from types.gen.ts (generated from Go structs)
// - Frontend-only types: Defined here for API-specific requirements or composite types

import type { GearResponse as GeneratedGearItem } from '@/lib/wasm/types.gen'

// ============================================================================
// Generated from Go - re-exported from types.gen.ts
// DO NOT modify these - regenerate with: just generate-ts-types
// ============================================================================
export type Gear = GeneratedGearItem
export type { GearMonthlyUsage } from '@/lib/wasm/types.gen'

// ============================================================================
// Frontend-only types - defined here for API-specific requirements
// Safe to modify as needed
// ============================================================================

// CustomGearCreateRequest - optional retired field (generated has required)
export interface CustomGearCreateRequest {
  name: string
  hashtag: string
  retired?: boolean
  purchase_price?: number | null
  purchase_currency?: string
}

// GearFilters - query parameters with optional fields and stricter order_by union
export interface GearFilters {
  include_retired?: boolean
  page?: number
  per_page?: number
  order_by?: 'name' | 'distance'
  order_dir?: 'asc' | 'desc'
}

// GearListResponse - paginated list wrapper (distinct from GearResponse in types.gen.ts which is a single item)
export interface GearListResponse {
  data: Gear[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export {
  useGear,
  useGearDetail,
  useCustomGear,
  useCreateCustomGear,
  useUpdateCustomGear,
  useDeleteCustomGear,
  useGearMonthlyUsage,
} from '@/lib/data'
