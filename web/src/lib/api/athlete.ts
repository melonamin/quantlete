export interface MetricPoint {
  recorded_at: string
  value: number
}

export interface FTPHistoryResponse {
  cycling: MetricPoint[]
  running: MetricPoint[]
}

export interface WeightHistoryResponse {
  points: MetricPoint[]
}

export {
  useFtpHistory,
  useUpdateFtpHistory,
  useWeightHistory,
  useUpdateWeightHistory,
} from '@/lib/data'
