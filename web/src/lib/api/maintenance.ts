export interface MaintenanceRule {
  id: number
  component_id: number
  type: 'distance_m' | 'time_s' | 'days'
  threshold_value: number
  created_at: string
  updated_at: string
}

export interface ComponentWithRules {
  id: number
  gear_id: string
  name: string
  image_url?: string
  maintenance_hashtag?: string
  created_at: string
  updated_at: string
  last_completed_at?: string
  rules: MaintenanceRule[]
}

export interface RuleProgress {
  type: 'distance_m' | 'time_s' | 'days'
  threshold_value: number
  current_value: number
  percent: number
  due: boolean
}

export interface DueComponent extends ComponentWithRules {
  distance_since: number
  moving_time_since: number
  days_since: number
  progress: RuleProgress[]
  is_due: boolean
}

export interface ComponentRuleInput {
  type: 'distance_m' | 'time_s' | 'days'
  threshold_value: number
}

export interface CreateComponentRequest {
  name: string
  image_url?: string
  maintenance_hashtag?: string
  rules?: ComponentRuleInput[]
}

export interface UpdateComponentRequest {
  name?: string
  image_url?: string
  maintenance_hashtag?: string
  rules?: ComponentRuleInput[]
}

export interface LogMaintenanceRequest {
  activity_id?: number
  completed_at?: string
}

export {
  useMaintenanceDue,
  useGearComponents,
  useCreateComponent,
  useUpdateComponent,
  useDeleteComponent,
  useLogMaintenance,
} from '@/lib/data'
