import { useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { LineChart } from '@/components/charts'

interface WeightHistoryPoint {
  recorded_at: string
  value: number
}

interface WeightHistory {
  points: WeightHistoryPoint[]
}

interface WeightEditorProps {
  value: WeightHistory | undefined
  onSave: (v: WeightHistory) => void
  saving: boolean
}

function parseISODate(d: string) {
  const t = new Date(d + 'T00:00:00')
  return Number.isFinite(t.getTime()) ? t : null
}

export function WeightEditor({ value, onSave, saving }: WeightEditorProps) {
  const [date, setDate] = useState('')
  const [weight, setWeight] = useState('')

  const series = useMemo(() => {
    const pts = (value?.points ?? []).filter((p) => parseISODate(p.recorded_at))
    return [
      {
        name: 'Weight (kg)',
        data: pts.map((p) => ({ x: p.recorded_at, y: p.value })),
      },
    ]
  }, [value])

  const addPoint = () => {
    if (!value) return
    const v = Number(weight)
    if (!date || !Number.isFinite(v) || v <= 0) return
    const next = structuredClone(value)
    next.points = [...next.points, { recorded_at: date, value: v }]
      .sort((a, b) => a.recorded_at.localeCompare(b.recorded_at))
      .filter((p, idx, arr) => idx === 0 || p.recorded_at !== arr[idx - 1].recorded_at)
    onSave(next)
    setDate('')
    setWeight('')
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Weight History</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <div className="grid gap-2 md:grid-cols-3">
        <input
          type="date"
          className="h-9 rounded-md border border-border bg-background px-2"
          value={date}
          onChange={(e) => setDate(e.target.value)}
        />
        <input
          inputMode="decimal"
          placeholder="Weight (kg)"
          className="h-9 rounded-md border border-border bg-background px-2"
          value={weight}
          onChange={(e) => setWeight(e.target.value)}
        />
        <Button onClick={addPoint}>Add</Button>
      </div>
      <LineChart series={series} height={200} showLegend={false} />
    </div>
  )
}
