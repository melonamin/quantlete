import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useBestEffortPRs, useBestEffortsForDistance, type BestEffortPR } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { LineChart } from '@/components/charts'
import { formatDate, formatDuration } from '@/lib/format'

type SportFilter = 'all' | 'run' | 'ride'

const SPORT_FILTERS: { key: SportFilter; label: string; sportType?: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'run', label: 'Runs', sportType: 'Run,TrailRun,VirtualRun' },
  {
    key: 'ride',
    label: 'Rides',
    sportType: 'Ride,MountainBikeRide,GravelRide,EBikeRide,VirtualRide',
  },
]

const DISTANCES: { type: string; label: string }[] = [
  { type: '400m', label: '400m' },
  { type: '0.5mi', label: '1/2 mile' },
  { type: '1k', label: '1km' },
  { type: '1mi', label: '1 mile' },
  { type: '2mi', label: '2 mile' },
  { type: '5k', label: '5km' },
  { type: '10k', label: '10km' },
  { type: '15k', label: '15km' },
  { type: '10mi', label: '10 mile' },
  { type: '20k', label: '20km' },
  { type: 'half_marathon', label: 'Half Marathon' },
  { type: '30k', label: '30km' },
  { type: 'marathon', label: 'Marathon' },
  { type: '50k', label: '50km' },
  { type: '100k', label: '100km' },
]

function distanceLabel(distanceType: string) {
  return DISTANCES.find((d) => d.type === distanceType)?.label ?? distanceType
}

export function BestEffortsPage() {
  const [sportFilter, setSportFilter] = useState<SportFilter>('run')
  const sportType = SPORT_FILTERS.find((s) => s.key === sportFilter)?.sportType

  const { data: prs, isLoading, error } = useBestEffortPRs(sportType)
  const [activeDistance, setActiveDistance] = useState<string | null>(null)

  const prsByType = useMemo(() => {
    const m = new Map<string, BestEffortPR>()
    for (const p of prs ?? []) m.set(p.distance_type, p)
    return m
  }, [prs])

  const otherPRs = useMemo(() => {
    const standard = new Set(DISTANCES.map((d) => d.type))
    return (prs ?? [])
      .filter((p) => !standard.has(p.distance_type))
      .sort((a, b) => a.distance_m - b.distance_m)
  }, [prs])

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Best Efforts</h1>
          <p className="text-muted-foreground">Your personal records by distance</p>
        </div>
        <div className="flex gap-2">
          {SPORT_FILTERS.map((f) => (
            <Button
              key={f.key}
              variant={sportFilter === f.key ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSportFilter(f.key)}
            >
              {f.label}
            </Button>
          ))}
        </div>
      </div>

      {error ? (
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">Failed to load best efforts.</p>
          <p className="mt-2 text-sm text-muted-foreground">
            Import with “Include best efforts” enabled in{' '}
            <Link to="/settings" className="underline">
              Settings
            </Link>
            .
          </p>
        </div>
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Personal Records</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {isLoading ? (
              <div className="space-y-2">
                {Array.from({ length: 8 }).map((_, i) => (
                  <Skeleton key={i} className="h-10 w-full" />
                ))}
              </div>
            ) : (
              <div className="space-y-2">
                {DISTANCES.map((d) => {
                  const pr = prsByType.get(d.type)
                  return (
                    <button
                      key={d.type}
                      className="flex w-full items-center justify-between rounded-md border border-border px-3 py-2 text-left hover:bg-accent/40"
                      onClick={() => setActiveDistance(d.type)}
                    >
                      <div className="text-sm font-medium">{d.label}</div>
                      <div className="flex items-center gap-3 text-sm">
                        <span className="tabular-nums">
                          {pr ? formatDuration(pr.elapsed_time_s) : '—'}
                        </span>
                        {pr ? (
                          <span className="text-xs text-muted-foreground">
                            {formatDate(pr.start_date_local)}
                          </span>
                        ) : (
                          <span className="text-xs text-muted-foreground">No data</span>
                        )}
                      </div>
                    </button>
                  )
                })}

                {otherPRs.length > 0 && (
                  <div className="mt-4">
                    <div className="mb-2 text-sm font-medium">Other distances</div>
                    <div className="space-y-2">
                      {otherPRs.map((p) => (
                        <button
                          key={p.distance_type}
                          className="flex w-full items-center justify-between rounded-md border border-border px-3 py-2 text-left hover:bg-accent/40"
                          onClick={() => setActiveDistance(p.distance_type)}
                        >
                          <div className="text-sm font-medium">
                            {distanceLabel(p.distance_type)}
                          </div>
                          <div className="flex items-center gap-3 text-sm">
                            <span className="tabular-nums">{formatDuration(p.elapsed_time_s)}</span>
                            <span className="text-xs text-muted-foreground">
                              {formatDate(p.start_date_local)}
                            </span>
                          </div>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {activeDistance && (
        <DistanceModal
          distanceType={activeDistance}
          sportType={sportType}
          onClose={() => setActiveDistance(null)}
        />
      )}
    </div>
  )
}

function DistanceModal({
  distanceType,
  sportType,
  onClose,
}: {
  distanceType: string
  sportType?: string
  onClose: () => void
}) {
  const { data, isLoading, error } = useBestEffortsForDistance(
    distanceType,
    sportType,
    !!distanceType
  )

  const progression = useMemo(() => {
    const items = [...(data ?? [])].sort((a, b) =>
      a.start_date_local.localeCompare(b.start_date_local)
    )
    let best = Number.POSITIVE_INFINITY
    const pts: { x: string; y: number }[] = []
    for (const it of items) {
      if (it.elapsed_time_s < best) {
        best = it.elapsed_time_s
        pts.push({ x: it.start_date_local, y: it.elapsed_time_s })
      }
    }
    return pts
  }, [data])

  const series = useMemo(() => {
    return [
      {
        name: 'PR',
        data: progression,
        smooth: false,
        step: 'end' as const,
      },
    ]
  }, [progression])

  return (
    <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
      <div className="mx-auto mt-10 w-[calc(100%-2rem)] max-w-3xl rounded-lg border border-border bg-card p-4 shadow-lg">
        <div className="mb-3 flex items-center justify-between gap-2">
          <div>
            <div className="text-base font-semibold">{distanceLabel(distanceType)}</div>
            <div className="text-xs text-muted-foreground">Progression and all efforts</div>
          </div>
          <Button variant="ghost" size="sm" onClick={onClose}>
            Close
          </Button>
        </div>

        {error ? (
          <p className="text-sm text-destructive">Failed to load distance efforts.</p>
        ) : (
          <div className="space-y-4">
            <div className="rounded-md border border-border p-3">
              <LineChart
                series={series}
                xAxisType="time"
                yAxisLabel="Elapsed (s)"
                height={240}
                loading={isLoading}
                showLegend={false}
                showDataZoom
              />
            </div>

            <div className="rounded-md border border-border">
              <div className="grid grid-cols-12 gap-2 border-b border-border px-3 py-2 text-xs font-medium text-muted-foreground">
                <div className="col-span-3">Date</div>
                <div className="col-span-3">Time</div>
                <div className="col-span-6">Activity</div>
              </div>
              {(data ?? [])
                .slice()
                .reverse()
                .map((it) => (
                  <div
                    key={`${it.activity_id}-${it.distance_type}`}
                    className="grid grid-cols-12 gap-2 px-3 py-2 text-sm"
                  >
                    <div className="col-span-3 text-muted-foreground">
                      {formatDate(it.start_date_local)}
                    </div>
                    <div className="col-span-3 tabular-nums">
                      {formatDuration(it.elapsed_time_s)}
                    </div>
                    <div className="col-span-6 truncate">
                      <Link
                        to="/activities/$activityId"
                        params={{ activityId: String(it.activity_id) }}
                        className="hover:underline"
                      >
                        {it.activity_name}
                      </Link>
                    </div>
                  </div>
                ))}
              {!isLoading && (data ?? []).length === 0 && (
                <div className="px-3 py-6 text-center text-sm text-muted-foreground">
                  No efforts found.
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
