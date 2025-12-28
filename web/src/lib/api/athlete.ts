import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { get, put } from './client'

export interface MetricPoint {
  recorded_at: string // YYYY-MM-DD
  value: number
}

export interface FTPHistoryResponse {
  cycling: MetricPoint[]
  running: MetricPoint[]
}

export interface WeightHistoryResponse {
  points: MetricPoint[]
}

export const athleteKeys = {
  all: ['athlete'] as const,
  ftp: () => [...athleteKeys.all, 'ftp'] as const,
  weight: () => [...athleteKeys.all, 'weight'] as const,
}

export function useFtpHistory() {
  return useQuery({
    queryKey: athleteKeys.ftp(),
    queryFn: () => get<FTPHistoryResponse>('/athlete/ftp'),
  })
}

export function useUpdateFtpHistory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: FTPHistoryResponse) => put<FTPHistoryResponse>('/athlete/ftp', body),
    onSuccess: (data) => {
      qc.setQueryData(athleteKeys.ftp(), data)
    },
  })
}

export function useWeightHistory() {
  return useQuery({
    queryKey: athleteKeys.weight(),
    queryFn: () => get<WeightHistoryResponse>('/athlete/weight'),
  })
}

export function useUpdateWeightHistory() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (body: WeightHistoryResponse) =>
      put<WeightHistoryResponse>('/athlete/weight', body),
    onSuccess: (data) => {
      qc.setQueryData(athleteKeys.weight(), data)
    },
  })
}
