import { useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { AppSettings } from '@/lib/api/settings'

const VIRTUAL_WORLDS = [
  'Watopia',
  'London',
  'New York',
  'Innsbruck',
  'Richmond',
  'France',
  'Makuri Islands',
  'Scotland',
  'Yumezi',
  'Crit City',
] as const

const DEFAULT_EDDINGTON_DEFS = [
  {
    id: 'all',
    name: 'All activities',
    sport_types: undefined,
    show_in_nav: true,
    show_in_dashboard_widget: true,
  },
  {
    id: 'rides',
    name: 'Rides',
    sport_types: ['Ride', 'MountainBikeRide', 'GravelRide', 'EBikeRide', 'VirtualRide'],
    show_in_nav: true,
    show_in_dashboard_widget: true,
  },
  {
    id: 'runs',
    name: 'Runs',
    sport_types: ['Run', 'TrailRun', 'VirtualRun'],
    show_in_nav: true,
    show_in_dashboard_widget: true,
  },
]

const DEFAULT_APP_SETTINGS: AppSettings = {
  version: 3,
  virtual_world_tile_layers: {},
  eddington_definitions: DEFAULT_EDDINGTON_DEFS,
  scheduler: {
    version: 2,
    pull: { enabled: false, schedule: 'midnight' },
    push: { enabled: false },
  },
}

interface VirtualWorldTilesEditorProps {
  settings: AppSettings | undefined
  onSave: (s: AppSettings) => void
  saving: boolean
}

export function VirtualWorldTilesEditor({
  settings,
  onSave,
  saving,
}: VirtualWorldTilesEditorProps) {
  const current = useMemo(() => {
    const merged = settings ?? DEFAULT_APP_SETTINGS
    return {
      ...merged,
      version: merged.version ?? 3,
      virtual_world_tile_layers: merged.virtual_world_tile_layers ?? {},
      eddington_definitions:
        merged.eddington_definitions && merged.eddington_definitions.length
          ? merged.eddington_definitions
          : DEFAULT_EDDINGTON_DEFS,
    }
  }, [settings])
  const [world, setWorld] = useState<string>(VIRTUAL_WORLDS[0])
  const [url, setUrl] = useState('')
  const [attribution, setAttribution] = useState('')
  const [maxZoom, setMaxZoom] = useState('18')

  const addOrUpdate = () => {
    if (!url) return
    const next = structuredClone(current)
    next.virtual_world_tile_layers[world] = {
      name: world,
      url,
      attribution: attribution || undefined,
      max_zoom: maxZoom ? Number(maxZoom) : undefined,
    }
    next.version = Math.max(next.version ?? 3, 3)
    onSave(next)
    setUrl('')
    setAttribution('')
  }

  const remove = (w: string) => {
    const next = structuredClone(current)
    delete next.virtual_world_tile_layers[w]
    next.version = Math.max(next.version ?? 3, 3)
    onSave(next)
  }

  const entries = Object.entries(current.virtual_world_tile_layers ?? {})

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Virtual World Maps</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <p className="text-xs text-muted-foreground">
        Add tile URLs for virtual worlds (Zwift/Rouvy/MyWhoosh). When a virtual activity is
        detected, the map switches to the matching tile layer.
      </p>
      <div className="grid gap-2 md:grid-cols-2">
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">World</label>
          <Select value={world} onValueChange={setWorld}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {VIRTUAL_WORLDS.map((w) => (
                <SelectItem key={w} value={w}>
                  {w}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Max zoom</label>
          <Input
            value={maxZoom}
            onChange={(e) => setMaxZoom(e.target.value)}
          />
        </div>
        <div className="space-y-1 md:col-span-2">
          <label className="text-sm text-muted-foreground">Tile URL template</label>
          <Input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/tiles/{z}/{x}/{y}.png"
          />
        </div>
        <div className="space-y-1 md:col-span-2">
          <label className="text-sm text-muted-foreground">Attribution (optional)</label>
          <Input
            value={attribution}
            onChange={(e) => setAttribution(e.target.value)}
          />
        </div>
      </div>
      <div className="flex justify-end">
        <Button onClick={addOrUpdate}>Save Tile Layer</Button>
      </div>

      {entries.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-medium">Configured worlds</div>
          <div className="space-y-2">
            {entries.map(([k, v]) => (
              <div key={k} className="rounded-md border border-border p-3">
                <div className="flex items-center justify-between gap-2">
                  <div className="min-w-0">
                    <div className="font-medium">{k}</div>
                    <div className="truncate text-xs text-muted-foreground">{v.url}</div>
                  </div>
                  <Button variant="outline" size="sm" onClick={() => remove(k)}>
                    Remove
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
