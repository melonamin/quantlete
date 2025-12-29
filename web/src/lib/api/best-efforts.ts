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

export { useBestEffortPRs, useBestEffortsForDistance } from '@/lib/data'
