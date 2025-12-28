import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { get, put } from './client'

export type GoalPeriod = 'week' | 'month' | 'year' | 'lifetime'

export interface GoalsMetricTargets {
  distance_m?: number
  elevation_m?: number
  moving_time_s?: number
}

export interface GoalsSportConfig {
  name: string
  sport_types: string[]
  targets: Partial<Record<GoalPeriod, GoalsMetricTargets>>
}

export interface TrainingGoalsConfig {
  version: number
  sports: GoalsSportConfig[]
}

export interface GoalsProgress {
  distance_m: number
  elevation_m: number
  moving_time_s: number
  activity_count: number
}

export interface TrainingGoalsResponse {
  config: TrainingGoalsConfig
  progress: Record<string, Record<GoalPeriod, GoalsProgress>>
}

export const goalsKeys = {
  all: ['goals'] as const,
}

export function useTrainingGoals() {
  return useQuery({
    queryKey: goalsKeys.all,
    queryFn: () => get<TrainingGoalsResponse>('/goals'),
  })
}

export function useUpdateTrainingGoals() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (cfg: TrainingGoalsConfig) => put<TrainingGoalsConfig>('/goals', cfg),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: goalsKeys.all })
    },
  })
}
