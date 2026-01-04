import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Checkbox } from '@/components/ui/checkbox'
import { Label } from '@/components/ui/label'
import type { AppSettings } from '@/lib/api/settings'

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

interface EddingtonDefinitionsEditorProps {
  settings: AppSettings | undefined
  onSave: (s: AppSettings) => void
  saving: boolean
}

export function EddingtonDefinitionsEditor({
  settings,
  onSave,
  saving,
}: EddingtonDefinitionsEditorProps) {
  const current = settings ?? DEFAULT_APP_SETTINGS
  const defs =
    current.eddington_definitions && current.eddington_definitions.length
      ? current.eddington_definitions
      : DEFAULT_EDDINGTON_DEFS

  const [name, setName] = useState('')
  const [sportTypes, setSportTypes] = useState('')
  const [showInNav, setShowInNav] = useState(true)
  const [showInWidget, setShowInWidget] = useState(true)

  const add = () => {
    if (!name.trim()) return
    const id = `e_${Math.random().toString(16).slice(2, 10)}`
    const types = sportTypes
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    const next = structuredClone(current)
    next.eddington_definitions = [
      ...(defs ?? []),
      {
        id,
        name: name.trim(),
        sport_types: types.length ? types : undefined,
        show_in_nav: showInNav,
        show_in_dashboard_widget: showInWidget,
      },
    ]
    next.version = Math.max(next.version ?? 3, 3)
    onSave(next)
    setName('')
    setSportTypes('')
  }

  const update = (
    id: string,
    patch: Partial<{
      name: string
      sport_types?: string[]
      show_in_nav?: boolean
      show_in_dashboard_widget?: boolean
    }>
  ) => {
    const next = structuredClone(current)
    next.eddington_definitions = defs.map((d) => (d.id === id ? { ...d, ...patch } : d))
    next.version = Math.max(next.version ?? 3, 3)
    onSave(next)
  }

  const remove = (id: string) => {
    const nextDefs = defs.filter((d) => d.id !== id)
    const next = structuredClone(current)
    next.eddington_definitions = nextDefs.length ? nextDefs : DEFAULT_EDDINGTON_DEFS
    next.version = Math.max(next.version ?? 3, 3)
    onSave(next)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Eddington Definitions</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <p className="text-xs text-muted-foreground">
        Create multiple Eddington definitions (e.g. Rides, Runs). Leave sport types empty to include
        all activities.
      </p>

      <div className="grid gap-2 md:grid-cols-2">
        <div className="space-y-1">
          <Label className="text-sm text-muted-foreground">Name</Label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Commutes"
          />
        </div>
        <div className="space-y-1">
          <Label className="text-sm text-muted-foreground">Sport types (optional)</Label>
          <Input
            value={sportTypes}
            onChange={(e) => setSportTypes(e.target.value)}
            placeholder="Ride,VirtualRide"
          />
        </div>
        <div className="flex items-center gap-2 pt-2">
          <Checkbox
            id="new-show-in-nav"
            checked={showInNav}
            onCheckedChange={(c) => setShowInNav(c === true)}
          />
          <Label htmlFor="new-show-in-nav" className="text-sm font-normal">
            Show in navigation
          </Label>
        </div>
        <div className="flex items-center gap-2 pt-2">
          <Checkbox
            id="new-show-in-widget"
            checked={showInWidget}
            onCheckedChange={(c) => setShowInWidget(c === true)}
          />
          <Label htmlFor="new-show-in-widget" className="text-sm font-normal">
            Show in dashboard widget
          </Label>
        </div>
      </div>

      <div className="flex justify-end">
        <Button onClick={add}>Add Definition</Button>
      </div>

      <div className="space-y-2">
        <div className="text-sm font-medium">Current definitions</div>
        <div className="space-y-2">
          {defs.map((d) => (
            <div key={d.id} className="rounded-md border border-border p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="min-w-0">
                  <div className="font-medium">{d.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {(d.sport_types && d.sport_types.length
                      ? d.sport_types.join(', ')
                      : 'All sport types') +
                      ' • ' +
                      (d.show_in_nav ? 'Nav' : 'Hidden') +
                      ' • ' +
                      (d.show_in_dashboard_widget ? 'Widget' : 'Hidden')}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => update(d.id, { show_in_nav: !d.show_in_nav })}
                  >
                    {d.show_in_nav ? 'Hide Nav' : 'Show Nav'}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      update(d.id, { show_in_dashboard_widget: !d.show_in_dashboard_widget })
                    }
                  >
                    {d.show_in_dashboard_widget ? 'Hide Widget' : 'Show Widget'}
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => remove(d.id)}>
                    Delete
                  </Button>
                </div>
              </div>

              <div className="mt-3 grid gap-2 md:grid-cols-2">
                <div className="space-y-1">
                  <Label className="text-sm text-muted-foreground">Rename</Label>
                  <Input value={d.name} onChange={(e) => update(d.id, { name: e.target.value })} />
                </div>
                <div className="space-y-1">
                  <Label className="text-sm text-muted-foreground">Sport types</Label>
                  <Input
                    value={(d.sport_types ?? []).join(',')}
                    onChange={(e) => {
                      const v = e.target.value
                      const types = v
                        .split(',')
                        .map((s) => s.trim())
                        .filter(Boolean)
                      update(d.id, { sport_types: types.length ? types : undefined })
                    }}
                    placeholder="All"
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
