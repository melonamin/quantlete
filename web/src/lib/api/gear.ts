import { useQuery } from '@tanstack/react-query'
import { get } from './client'

export interface Gear {
  id: string
  name: string
  primary: boolean
  retired: boolean
  distance: number
  brand_name?: string
  model_name?: string
  description?: string
  activity_count: number
}

export const gearKeys = {
  all: ['gear'] as const,
  list: (includeRetired?: boolean) => [...gearKeys.all, 'list', includeRetired] as const,
  detail: (id: string) => [...gearKeys.all, 'detail', id] as const,
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
