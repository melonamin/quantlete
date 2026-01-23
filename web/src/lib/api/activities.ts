// Activity stream types - defined here for API consistency
// The generated type has data as string (JSON), but API returns parsed array

export interface ActivityStream {
  activity_id?: number
  stream_type: string
  data: number[] | string
  series_type: string
  original_size: number
  resolution: string
}

// Activity analysis types
export interface SplitItem {
  index: number
  distance_m: number
  duration_s: number
  pace_sec_km: number
  avg_hr?: number
  avg_watts?: number
  elev_gain: number
  elev_loss: number
}

export interface SplitsOutput {
  splits: SplitItem[]
  split_length_m: number
  total_splits: number
  fastest_split: number
  slowest_split: number
}

export interface ZoneItem {
  zone: number
  seconds: number
  percentage: number
  min_bpm: number
  max_bpm: number
  label: string
}

export interface HRZonesOutput {
  zones: ZoneItem[]
  total_seconds: number
  avg_hr: number
  max_hr: number
}

export interface PaceBucketItem {
  min_pace: number
  max_pace: number
  count: number
  seconds: number
  percentage: number
}

export interface PaceDistributionOutput {
  buckets: PaceBucketItem[]
  total_seconds: number
  avg_pace: number
  fastest_pace: number
  slowest_pace: number
  median_pace: number
}

export interface ActivityAnalysis {
  activity_id: number
  splits?: SplitsOutput | null
  hr_zones?: HRZonesOutput | null
  pace_distribution?: PaceDistributionOutput | null
}

export { useActivities, useActivity, useActivityStreams, useActivityAnalysis } from '@/lib/data'
