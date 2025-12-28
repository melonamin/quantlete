import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { get, put } from './client'

export interface VirtualWorldTileLayer {
  name: string
  url: string
  attribution?: string
  max_zoom?: number
}

export interface EddingtonDefinition {
  id: string
  name: string
  sport_types?: string[]
  show_in_nav?: boolean
  show_in_dashboard_widget?: boolean
}

export interface AppSettings {
  version: number
  virtual_world_tile_layers: Record<string, VirtualWorldTileLayer>
  eddington_definitions?: EddingtonDefinition[]
}

export const settingsKeys = {
  all: ['settings'] as const,
}

export function useAppSettings(opts: { enabled?: boolean } = {}) {
  return useQuery({
    queryKey: settingsKeys.all,
    queryFn: () => get<AppSettings>('/settings'),
    enabled: opts.enabled ?? true,
    retry: false,
  })
}

export function useUpdateAppSettings() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (settings: AppSettings) => put<AppSettings>('/settings', settings),
    onSuccess: (data) => {
      qc.setQueryData(settingsKeys.all, data)
    },
  })
}
