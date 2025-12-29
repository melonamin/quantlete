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

export { useTrainingGoals, useUpdateTrainingGoals } from '@/lib/data'
