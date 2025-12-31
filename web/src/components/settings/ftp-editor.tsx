import { useMemo, useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { LineChart } from '@/components/charts'

interface FtpHistoryPoint {
  recorded_at: string
  value: number
}

interface FtpHistory {
  cycling: FtpHistoryPoint[]
  running: FtpHistoryPoint[]
}

interface FtpEditorProps {
  value: FtpHistory | undefined
  onSave: (v: FtpHistory) => void
  saving: boolean
}

function parseISODate(d: string) {
  const t = new Date(d + 'T00:00:00')
  return Number.isFinite(t.getTime()) ? t : null
}

function formatPaceFromMps(mps: number) {
  if (mps <= 0) return '–'
  const secPerKm = 1000 / mps
  const mm = Math.floor(secPerKm / 60)
  const ss = Math.round(secPerKm % 60)
  return `${mm}:${String(ss).padStart(2, '0')}/km`
}

export function FtpEditor({ value, onSave, saving }: FtpEditorProps) {
  const [cyclingDate, setCyclingDate] = useState('')
  const [cyclingValue, setCyclingValue] = useState('')
  const [runningDate, setRunningDate] = useState('')
  const [runningValue, setRunningValue] = useState('')

  const cyclingSeries = useMemo(() => {
    const pts = (value?.cycling ?? []).filter((p) => parseISODate(p.recorded_at))
    return [
      {
        name: 'Cycling FTP (W)',
        data: pts.map((p) => ({ x: p.recorded_at, y: p.value })),
      },
    ]
  }, [value])

  const runningSeries = useMemo(() => {
    const pts = (value?.running ?? []).filter((p) => parseISODate(p.recorded_at))
    return [
      {
        name: 'Running FTP (m/s)',
        data: pts.map((p) => ({ x: p.recorded_at, y: p.value })),
      },
    ]
  }, [value])

  const addPoint = (kind: 'cycling' | 'running') => {
    if (!value) return
    const date = kind === 'cycling' ? cyclingDate : runningDate
    const raw = kind === 'cycling' ? cyclingValue : runningValue
    const v = Number(raw)
    if (!date || !Number.isFinite(v) || v <= 0) return

    const next = structuredClone(value)
    next[kind] = [...next[kind], { recorded_at: date, value: v }]
      .sort((a, b) => a.recorded_at.localeCompare(b.recorded_at))
      .filter((p, idx, arr) => idx === 0 || p.recorded_at !== arr[idx - 1].recorded_at)

    onSave(next)
    if (kind === 'cycling') {
      setCyclingDate('')
      setCyclingValue('')
    } else {
      setRunningDate('')
      setRunningValue('')
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">FTP History</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>

      <div className="space-y-3">
        <div className="grid gap-2 md:grid-cols-3">
          <Input
            type="date"
            value={cyclingDate}
            onChange={(e) => setCyclingDate(e.target.value)}
          />
          <Input
            inputMode="numeric"
            placeholder="Cycling FTP (W)"
            value={cyclingValue}
            onChange={(e) => setCyclingValue(e.target.value)}
          />
          <Button onClick={() => addPoint('cycling')}>Add</Button>
        </div>
        <LineChart series={cyclingSeries} height={200} showLegend={false} />
      </div>

      <div className="space-y-3">
        <div className="grid gap-2 md:grid-cols-3">
          <Input
            type="date"
            value={runningDate}
            onChange={(e) => setRunningDate(e.target.value)}
          />
          <Input
            inputMode="decimal"
            placeholder="Running FTP (m/s)"
            value={runningValue}
            onChange={(e) => setRunningValue(e.target.value)}
          />
          <Button onClick={() => addPoint('running')}>Add</Button>
        </div>
        <div className="text-xs text-muted-foreground">
          Store running threshold speed in m/s (tooltip pace example: {formatPaceFromMps(3.5)}).
        </div>
        <LineChart series={runningSeries} height={200} showLegend={false} />
      </div>
    </div>
  )
}
