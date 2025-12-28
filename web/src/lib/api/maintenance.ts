import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { del, get, post, put } from './client'

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

export const maintenanceKeys = {
  all: ['maintenance'] as const,
  due: () => [...maintenanceKeys.all, 'due'] as const,
  gearComponents: (gearId: string) => [...maintenanceKeys.all, 'gearComponents', gearId] as const,
}

export function useMaintenanceDue(enabled = true) {
  return useQuery({
    queryKey: maintenanceKeys.due(),
    queryFn: () => get<DueComponent[]>('/maintenance/due'),
    enabled,
  })
}

export function useGearComponents(gearId: string, enabled = true) {
  return useQuery({
    queryKey: maintenanceKeys.gearComponents(gearId),
    queryFn: () => get<ComponentWithRules[]>(`/gear/${gearId}/components`),
    enabled: enabled && !!gearId,
  })
}

export function useCreateComponent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ gearId, req }: { gearId: string; req: CreateComponentRequest }) =>
      post<ComponentWithRules>(`/gear/${gearId}/components`, req),
    onSuccess: (_data, vars) => {
      queryClient.invalidateQueries({ queryKey: maintenanceKeys.gearComponents(vars.gearId) })
      queryClient.invalidateQueries({ queryKey: maintenanceKeys.due() })
    },
  })
}

export function useUpdateComponent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, req }: { id: number; req: UpdateComponentRequest }) =>
      put<ComponentWithRules>(`/components/${id}`, req),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: maintenanceKeys.gearComponents(data.gear_id) })
      queryClient.invalidateQueries({ queryKey: maintenanceKeys.due() })
    },
  })
}

export function useDeleteComponent() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id }: { id: number }) => del<{ deleted: boolean }>(`/components/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: maintenanceKeys.all })
    },
  })
}

export function useLogMaintenance() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ componentId, req }: { componentId: number; req?: LogMaintenanceRequest }) =>
      post<{ logged: boolean }>(`/components/${componentId}/maintenance`, req ?? {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: maintenanceKeys.all })
    },
  })
}
