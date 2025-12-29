export interface Gear {
  id: string
  name: string
  primary: boolean
  retired: boolean
  distance: number
  brand_name?: string
  model_name?: string
  description?: string
  source: string
  hashtag?: string
  purchase_price?: number
  purchase_currency?: string
  activity_count: number
}

export interface CustomGearCreateRequest {
  name: string
  hashtag: string
  retired?: boolean
  purchase_price?: number | null
  purchase_currency?: string
}

export interface GearMonthlyUsage {
  month: string
  gear_id: string
  gear_name: string
  source: string
  hashtag?: string
  retired: boolean
  purchase_price?: number
  purchase_currency?: string
  activity_count: number
  distance: number
  moving_time: number
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
