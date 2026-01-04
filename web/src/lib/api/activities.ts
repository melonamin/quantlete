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

export { useActivities, useActivity, useActivityStreams } from '@/lib/data'
