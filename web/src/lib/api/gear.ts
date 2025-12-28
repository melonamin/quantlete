import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { del, get, post, put } from './client'

export interface Gear {
  id: string
  name: string
  primary: boolean
  retired: boolean
  distance: number
  brand_name?: string
  model_name?: string
  description?: string
  source: string
  hashtag?: string
  purchase_price?: number
  purchase_currency?: string
  activity_count: number
}

export const gearKeys = {
  all: ['gear'] as const,
  list: (includeRetired?: boolean) => [...gearKeys.all, 'list', includeRetired] as const,
  detail: (id: string) => [...gearKeys.all, 'detail', id] as const,
  custom: (includeRetired?: boolean) => [...gearKeys.all, 'custom', includeRetired] as const,
  monthlyUsage: (includeRetired?: boolean) =>
    [...gearKeys.all, 'monthlyUsage', includeRetired] as const,
}

export function useGear(includeRetired = true) {
  return useQuery({
    queryKey: gearKeys.list(includeRetired),
    queryFn: () => get<Gear[]>(`/gear?include_retired=${includeRetired}`),
  })
}

export function useGearDetail(id: string) {
  return useQuery({
    queryKey: gearKeys.detail(id),
    queryFn: () => get<Gear>(`/gear/${id}`),
    enabled: !!id,
  })
}

export interface CustomGearCreateRequest {
  name: string
  hashtag: string
  retired?: boolean
  purchase_price?: number | null
  purchase_currency?: string
}

export function useCustomGear(includeRetired = true) {
  return useQuery({
    queryKey: gearKeys.custom(includeRetired),
    queryFn: () => get<Gear[]>(`/gear/custom?include_retired=${includeRetired}`),
  })
}

export function useCreateCustomGear() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: CustomGearCreateRequest) => post<Gear>('/gear/custom', req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: gearKeys.all })
    },
  })
}

export function useUpdateCustomGear() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, patch }: { id: string; patch: Partial<CustomGearCreateRequest> }) =>
      put<Gear>(`/gear/custom/${id}`, patch),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: gearKeys.all })
    },
  })
}

export function useDeleteCustomGear() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, force }: { id: string; force?: boolean }) =>
      del<{ deleted: boolean }>(`/gear/custom/${id}${force ? '?force=true' : ''}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: gearKeys.all })
    },
  })
}

export interface GearMonthlyUsage {
  month: string
  gear_id: string
  gear_name: string
  source: string
  hashtag?: string
  retired: boolean
  purchase_price?: number
  purchase_currency?: string
  activity_count: number
  distance: number
  moving_time: number
}

export function useGearMonthlyUsage(includeRetired = true) {
  return useQuery({
    queryKey: gearKeys.monthlyUsage(includeRetired),
    queryFn: () => get<GearMonthlyUsage[]>(`/gear/stats/monthly?include_retired=${includeRetired}`),
  })
}
