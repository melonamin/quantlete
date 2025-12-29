export interface ActivityStream {
  activity_id: number
  stream_type: string
  original_size: number
  resolution: string
  series_type: string
  data: unknown
}

export {
  useActivities,
  useActivity,
  useActivityStreams,
} from '@/lib/data'
