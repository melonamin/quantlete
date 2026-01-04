import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface HrZoneDefinition {
  sport_type: string
  effective_from: string
  method: string
  zones: { bounds: number[]; hr_max?: number }
}

interface HrZonesEditorProps {
  defs: HrZoneDefinition[] | undefined
  onSave: (def: HrZoneDefinition) => void
  onDelete: (p: { sport_type: string; effective_from: string }) => void
  saving: boolean
}

const DEFAULT_DEF = {
  sport_type: 'All',
  effective_from: '1970-01-01',
  method: 'percent_hrmax',
  zones: { hr_max: 190, bounds: [0.6, 0.7, 0.8, 0.9, 1.0] },
}

export function HrZonesEditor({ defs, onSave, onDelete, saving }: HrZonesEditorProps) {
  const [sportType, setSportType] = useState('All')
  const [effectiveFrom, setEffectiveFrom] = useState('1970-01-01')
  const [method, setMethod] = useState(DEFAULT_DEF.method)
  const [hrMax, setHrMax] = useState(String(DEFAULT_DEF.zones.hr_max ?? 190))
  const [bounds, setBounds] = useState(DEFAULT_DEF.zones.bounds.map(String))

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Heart Rate Zones</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <div className="grid gap-2 md:grid-cols-3">
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Sport group</label>
          <Select value={sportType} onValueChange={setSportType}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="All">All</SelectItem>
              <SelectItem value="Ride">Ride</SelectItem>
              <SelectItem value="Run">Run</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Effective from</label>
          <Input
            type="date"
            value={effectiveFrom}
            onChange={(e) => setEffectiveFrom(e.target.value)}
          />
        </div>
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Method</label>
          <Select value={method} onValueChange={setMethod}>
            <SelectTrigger>
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="percent_hrmax">% of HR max</SelectItem>
              <SelectItem value="absolute_bpm">Absolute BPM</SelectItem>
            </SelectContent>
          </Select>
        </div>
        {method === 'percent_hrmax' && (
          <div className="space-y-1">
            <label className="text-sm text-muted-foreground">HR max (bpm)</label>
            <Input inputMode="numeric" value={hrMax} onChange={(e) => setHrMax(e.target.value)} />
          </div>
        )}
      </div>
      <div className="grid gap-2 md:grid-cols-5">
        {bounds.slice(0, 5).map((v, idx) => (
          <div key={idx} className="space-y-1">
            <label className="text-sm text-muted-foreground">Z{idx + 1} max</label>
            <Input
              inputMode="decimal"
              value={v}
              onChange={(e) => {
                const next = bounds.slice()
                next[idx] = e.target.value
                setBounds(next)
              }}
            />
          </div>
        ))}
      </div>
      <div className="flex justify-end">
        <Button
          onClick={() => {
            const parsedBounds = bounds.map((b) => Number(b)).filter((n) => Number.isFinite(n))
            if (parsedBounds.length < 5) return
            onSave({
              sport_type: sportType,
              effective_from: effectiveFrom,
              method,
              zones:
                method === 'percent_hrmax'
                  ? { hr_max: Number(hrMax), bounds: parsedBounds }
                  : { bounds: parsedBounds },
            })
          }}
        >
          Save Definition
        </Button>
      </div>
      <p className="text-xs text-muted-foreground">
        For % zones, enter bounds as fractions (e.g. 0.6, 0.7, 0.8, 0.9, 1.0). For absolute zones,
        enter BPM caps.
      </p>

      {defs && defs.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-medium">Existing definitions</div>
          <div className="space-y-2">
            {defs.map((d) => (
              <div
                key={`${d.sport_type}:${d.effective_from}`}
                className="flex items-center justify-between rounded-md border border-border px-3 py-2"
              >
                <div className="text-sm">
                  <div className="font-medium">{d.sport_type}</div>
                  <div className="text-xs text-muted-foreground">
                    From {d.effective_from} • {d.method}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setSportType(d.sport_type)
                      setEffectiveFrom(d.effective_from)
                      setMethod(d.method)
                      setHrMax(String(d.zones.hr_max ?? 190))
                      setBounds(
                        (d.zones.bounds ?? DEFAULT_DEF.zones.bounds).slice(0, 5).map(String)
                      )
                    }}
                  >
                    Edit
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      onDelete({ sport_type: d.sport_type, effective_from: d.effective_from })
                    }
                  >
                    Delete
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
