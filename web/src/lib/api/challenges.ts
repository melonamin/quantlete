export interface Challenge {
  id: string
  name: string
  slug?: string
  badge_url?: string
  local_badge_url?: string
  completion_date?: string
  month?: string
}

export { useChallenges, useImportChallenges, useImportChallengesFromProfile } from '@/lib/data'
