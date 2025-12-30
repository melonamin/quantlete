export interface Challenge {
  id: string
  name: string
  slug?: string
  badge_url?: string
  local_badge_url?: string
  completion_date?: string
  month?: string
}

export interface ChallengesFilters {
  month?: string
  page?: number
  per_page?: number
  order_by?: 'name' | 'completion_date' | 'month'
  order_dir?: 'asc' | 'desc'
}

export interface ChallengesResponse {
  data: Challenge[]
  total: number
  page: number
  per_page: number
  total_pages: number
}

export { useChallenges, useImportChallenges, useImportChallengesFromProfile } from '@/lib/data'
