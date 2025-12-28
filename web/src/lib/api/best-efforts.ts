import { useQuery } from '@tanstack/react-query'
import { get } from './client'

export interface BestEffortPR {
  distance_type: string
  name: string
  distance_m: number
  elapsed_time_s: number
  moving_time_s?: number
  pr_rank?: number
  activity_id: number
  activity_name: string
  sport_type: string
  start_date_local: string
}

export interface BestEffortItem extends BestEffortPR {
  start_index?: number
  end_index?: number
}

export const bestEffortsKeys = {
  all: ['bestEfforts'] as const,
  prs: (sportType?: string) => [...bestEffortsKeys.all, 'prs', sportType] as const,
  distance: (distanceType: string, sportType?: string) =>
    [...bestEffortsKeys.all, 'distance', distanceType, sportType] as const,
}

export function useBestEffortPRs(sportType?: string) {
  return useQuery({
    queryKey: bestEffortsKeys.prs(sportType),
    queryFn: () => {
      const params = sportType ? `?sport_type=${encodeURIComponent(sportType)}` : ''
      return get<BestEffortPR[]>(`/stats/best-efforts${params}`)
    },
  })
}

export function useBestEffortsForDistance(
  distanceType: string,
  sportType?: string,
  enabled = true
) {
  return useQuery({
    queryKey: bestEffortsKeys.distance(distanceType, sportType),
    queryFn: () => {
      const params = sportType ? `?sport_type=${encodeURIComponent(sportType)}` : ''
      return get<BestEffortItem[]>(
        `/stats/best-efforts/${encodeURIComponent(distanceType)}${params}`
      )
    },
    enabled: enabled && !!distanceType,
  })
}
