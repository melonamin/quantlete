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
  scheduler: SchedulerSettings
  enable_public_badges?: boolean
}

export type PullSchedule = 'midnight' | 'hourly' | 'every_6_hours'

export interface SchedulerSettings {
  version: number
  pull: {
    enabled: boolean
    schedule: PullSchedule
  }
  push: {
    enabled: boolean
  }
}

export { useAppSettings, useUpdateAppSettings } from '@/lib/data'
