import { useMemo, useState } from 'react'
import { useTrainingGoals, useUpdateTrainingGoals, type GoalPeriod } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { formatDistance, formatDuration } from '@/lib/format'

const PERIODS: { key: GoalPeriod; label: string }[] = [
  { key: 'week', label: 'Week' },
  { key: 'month', label: 'Month' },
  { key: 'year', label: 'Year' },
  { key: 'lifetime', label: 'Lifetime' },
]

function toNumberOrUndefined(v: string) {
  const n = Number(v)
  return Number.isFinite(n) ? n : undefined
}

export function TrainingGoals() {
  const { data, isLoading } = useTrainingGoals()
  const update = useUpdateTrainingGoals()

  const sports = data?.config.sports ?? []
  const [sportName, setSportName] = useState<string | null>(null)
  const [period, setPeriod] = useState<GoalPeriod>('week')
  const [showEdit, setShowEdit] = useState(false)

  const activeSport = useMemo(() => {
    if (!sports.length) return null
    if (sportName) return sports.find((s) => s.name === sportName) ?? sports[0]
    return sports[0]
  }, [sports, sportName])

  const progress = activeSport ? data?.progress?.[activeSport.name]?.[period] : undefined
  const targets = activeSport?.targets?.[period]

  const hasAnyTarget =
    (targets?.distance_m ?? 0) > 0 ||
    (targets?.elevation_m ?? 0) > 0 ||
    (targets?.moving_time_s ?? 0) > 0

  return (
    <WidgetWrapper
      title="Training Goals"
      action={
        <Button variant="ghost" size="sm" onClick={() => setShowEdit(true)}>
          Configure
        </Button>
      }
      isLoading={isLoading}
    >
      {!activeSport ? (
        <p className="text-sm text-muted-foreground">No goals configured yet.</p>
      ) : (
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            <select
              className="h-9 rounded-md border border-border bg-background px-2 text-sm"
              value={activeSport.name}
              onChange={(e) => setSportName(e.target.value)}
            >
              {sports.map((s) => (
                <option key={s.name} value={s.name}>
                  {s.name}
                </option>
              ))}
            </select>
            <div className="flex flex-wrap gap-1">
              {PERIODS.map((p) => (
                <Button
                  key={p.key}
                  variant={period === p.key ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setPeriod(p.key)}
                >
                  {p.label}
                </Button>
              ))}
            </div>
          </div>

          {!hasAnyTarget ? (
            <div className="rounded-md border border-border bg-muted/30 p-3">
              <p className="text-sm text-muted-foreground">
                Set targets to track progress for this period.
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              <GoalRow
                label="Distance"
                current={progress?.distance_m ?? 0}
                target={targets?.distance_m}
                formatCurrent={(v) => formatDistance(v)}
              />
              <GoalRow
                label="Elevation"
                current={progress?.elevation_m ?? 0}
                target={targets?.elevation_m}
                formatCurrent={(v) => `${Math.round(v)} m`}
              />
              <GoalRow
                label="Moving Time"
                current={progress?.moving_time_s ?? 0}
                target={targets?.moving_time_s}
                formatCurrent={(v) => formatDuration(v)}
              />
            </div>
          )}
        </div>
      )}

      {showEdit && data?.config && activeSport && (
        <EditGoalsDialog
          sportName={activeSport.name}
          currentTargets={activeSport.targets?.[period]}
          period={period}
          onClose={() => setShowEdit(false)}
          onSave={(t) => {
            const next = structuredClone(data.config)
            const sport = next.sports.find((s) => s.name === activeSport.name)
            if (!sport) return
            sport.targets = sport.targets ?? {}
            sport.targets[period] = t
            update.mutate(next, { onSuccess: () => setShowEdit(false) })
          }}
        />
      )}
    </WidgetWrapper>
  )
}

function GoalRow({
  label,
  current,
  target,
  formatCurrent,
}: {
  label: string
  current: number
  target: number | undefined
  formatCurrent: (v: number) => string
}) {
  if (!target || target <= 0) {
    return (
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="text-muted-foreground">Not set</span>
      </div>
    )
  }

  const pct = Math.min(100, Math.round((current / target) * 100))

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-sm">
        <span className="text-muted-foreground">{label}</span>
        <span className="font-medium">
          {formatCurrent(current)} /{' '}
          {label === 'Distance'
            ? formatDistance(target)
            : label === 'Moving Time'
              ? formatDuration(target)
              : `${Math.round(target)} m`}
          <span className="ml-2 text-muted-foreground">({pct}%)</span>
        </span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
        <div className="h-full bg-primary transition-all" style={{ width: `${pct}%` }} />
      </div>
    </div>
  )
}

function EditGoalsDialog({
  sportName,
  period,
  currentTargets,
  onClose,
  onSave,
}: {
  sportName: string
  period: GoalPeriod
  currentTargets: { distance_m?: number; elevation_m?: number; moving_time_s?: number } | undefined
  onClose: () => void
  onSave: (targets: { distance_m?: number; elevation_m?: number; moving_time_s?: number }) => void
}) {
  const [distanceKm, setDistanceKm] = useState(
    currentTargets?.distance_m
      ? String(Math.round((currentTargets.distance_m / 1000) * 10) / 10)
      : ''
  )
  const [elevationM, setElevationM] = useState(
    currentTargets?.elevation_m ? String(Math.round(currentTargets.elevation_m)) : ''
  )
  const [timeH, setTimeH] = useState(
    currentTargets?.moving_time_s
      ? String(Math.round((currentTargets.moving_time_s / 3600) * 10) / 10)
      : ''
  )

  return (
    <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
      <div className="mx-auto mt-20 w-full max-w-lg rounded-lg border border-border bg-card p-4 shadow-lg">
        <div className="mb-3 flex items-center justify-between">
          <h2 className="text-base font-semibold">
            {sportName} – {PERIODS.find((p) => p.key === period)?.label ?? period}
          </h2>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>
        <div className="space-y-3">
          <Field label="Distance (km)">
            <input
              className="h-9 w-full rounded-md border border-border bg-background px-2"
              inputMode="decimal"
              value={distanceKm}
              onChange={(e) => setDistanceKm(e.target.value)}
              placeholder="e.g. 150"
            />
          </Field>
          <Field label="Elevation (m)">
            <input
              className="h-9 w-full rounded-md border border-border bg-background px-2"
              inputMode="numeric"
              value={elevationM}
              onChange={(e) => setElevationM(e.target.value)}
              placeholder="e.g. 2000"
            />
          </Field>
          <Field label="Moving time (hours)">
            <input
              className="h-9 w-full rounded-md border border-border bg-background px-2"
              inputMode="decimal"
              value={timeH}
              onChange={(e) => setTimeH(e.target.value)}
              placeholder="e.g. 6.5"
            />
          </Field>
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button
            onClick={() => {
              onSave({
                distance_m: distanceKm ? (toNumberOrUndefined(distanceKm) ?? 0) * 1000 : undefined,
                elevation_m: elevationM ? toNumberOrUndefined(elevationM) : undefined,
                moving_time_s: timeH
                  ? Math.round((toNumberOrUndefined(timeH) ?? 0) * 3600)
                  : undefined,
              })
            }}
          >
            Save
          </Button>
        </div>
      </div>
    </div>
  )
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="space-y-1">
      <label className="text-sm text-muted-foreground">{label}</label>
      {children}
    </div>
  )
}
