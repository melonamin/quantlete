import { useEffect, useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useRewind, useRewindYears } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { BarChart, DonutChart, LineChart, WorldLocationsChart } from '@/components/charts'
import { formatDistance, formatDurationLong, formatElevation } from '@/lib/format'

function pctChange(current: number, baseline: number) {
  if (!Number.isFinite(current) || !Number.isFinite(baseline) || baseline === 0) return null
  return ((current - baseline) / baseline) * 100
}

export function RewindPage() {
  const nowYear = new Date().getFullYear()
  const { data: years, isLoading: yearsLoading } = useRewindYears()
  const availableYears = useMemo(() => {
    const ys = years && years.length ? years : [nowYear]
    return [nowYear, ...ys.filter((y) => y !== nowYear)]
  }, [years, nowYear])

  const [year, setYear] = useState(nowYear)
  const [compareYear, setCompareYear] = useState<number | null>(null)

  useEffect(() => {
    if (!availableYears.length) return
    if (year === 0) return
    if (availableYears.includes(year)) return
    setYear(availableYears[0])
  }, [availableYears, year])

  const { data: report, isLoading, isFetching, error } = useRewind(year)
  const { data: compare, isLoading: compareLoading } = useRewind(
    compareYear ?? 0,
    compareYear !== null
  )

  const months = report?.months ?? []
  const monthLabels = months.map((m) => m.month.slice(5))
  const activitiesData = months.map((m, idx) => ({ label: monthLabels[idx], value: m.activities }))
  const distanceData = months.map((m, idx) => ({
    label: monthLabels[idx],
    value: Math.round((m.distance_m / 1000) * 10) / 10,
  }))
  const elevationData = months.map((m, idx) => ({
    label: monthLabels[idx],
    value: Math.round(m.elevation_m),
  }))
  const prsSeries = useMemo(() => {
    return [
      {
        name: 'PRs',
        data: months.map((m, idx) => ({ x: monthLabels[idx], y: m.prs })),
        smooth: false,
      },
    ]
  }, [months, monthLabels])

  const startTimesSeries = useMemo(() => {
    const hours = report?.start_times_by_hour ?? []
    return [
      {
        name: 'Starts',
        data: hours.map((h) => ({ x: String(h.hour), y: h.count })),
        smooth: false,
      },
    ]
  }, [report])

  const movingTimeSlices = useMemo(() => {
    return (report?.moving_time_by_sport ?? [])
      .filter((s) => s.moving_time_s > 0)
      .slice(0, 10)
      .map((s) => ({ name: s.sport_type, value: s.moving_time_s }))
  }, [report])

  const activeRestSlices = useMemo(() => {
    if (!report) return []
    return [
      { name: 'Active days', value: report.active_days },
      { name: 'Rest days', value: report.rest_days },
    ]
  }, [report])

  const showMonths = year > 0

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Rewind</h1>
          <p className="text-muted-foreground">Year in review</p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <select
            className="h-9 rounded-md border border-border bg-background px-2 text-sm"
            value={String(year)}
            onChange={(e) => setYear(Number(e.target.value))}
            disabled={yearsLoading}
          >
            <option value="0">All time</option>
            {availableYears.map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>

          <select
            className="h-9 rounded-md border border-border bg-background px-2 text-sm"
            value={compareYear === null ? '' : String(compareYear)}
            onChange={(e) => setCompareYear(e.target.value ? Number(e.target.value) : null)}
          >
            <option value="">No comparison</option>
            <option value="0">Compare to all time</option>
            {availableYears.map((y) => (
              <option key={y} value={y}>
                Compare to {y}
              </option>
            ))}
          </select>
        </div>
      </div>

      {yearsLoading || isLoading || isFetching ? (
        <div className="space-y-4">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-80 w-full" />
        </div>
      ) : error || !report ? (
        <div className="rounded-lg border border-border bg-card p-8 text-center">
          <p className="text-muted-foreground">
            No activity data available for {year === 0 ? 'this period' : year}.
          </p>
          <p className="text-sm text-muted-foreground mt-1">
            Import activities to see your year in review.
          </p>
        </div>
      ) : (
        <div className="space-y-6">
          <div className="grid gap-4 md:grid-cols-3">
            <MetricCard title="Activities" value={String(report.totals.activities)} />
            <MetricCard title="Distance" value={formatDistance(report.totals.distance_m)} />
            <MetricCard title="Elevation" value={formatElevation(report.totals.elevation_m)} />
            <MetricCard
              title="Moving Time"
              value={formatDurationLong(report.totals.moving_time_s)}
            />
            <MetricCard title="Kudos Received" value={String(report.totals.kudos)} />
            <MetricCard
              title="CO₂ Saved (est.)"
              value={`${report.totals.carbon_saved_kg.toFixed(1)} kg`}
            />
          </div>

          <Card>
            <CardHeader>
              <CardTitle>Summary</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 md:grid-cols-2">
              <div className="rounded-md border border-border p-3">
                <div className="mb-2 text-sm font-medium">Active vs rest days</div>
                <DonutChart data={activeRestSlices} height={240} loading={false} />
                <div className="mt-2 text-xs text-muted-foreground">
                  {report.active_days} active / {report.total_days} total days • longest streak{' '}
                  {report.streaks.longest_active_days} days
                </div>
              </div>
              <div className="rounded-md border border-border p-3">
                <div className="mb-2 text-sm font-medium">Moving time by sport</div>
                <DonutChart data={movingTimeSlices} height={240} loading={false} />
              </div>
            </CardContent>
          </Card>

          {showMonths && (
            <div className="grid gap-4 md:grid-cols-2">
              <Card>
                <CardHeader>
                  <CardTitle>Activities by Month</CardTitle>
                </CardHeader>
                <CardContent>
                  <BarChart data={activitiesData} height={260} />
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Distance by Month (km)</CardTitle>
                </CardHeader>
                <CardContent>
                  <BarChart data={distanceData} height={260} />
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Elevation by Month (m)</CardTitle>
                </CardHeader>
                <CardContent>
                  <BarChart data={elevationData} height={260} />
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle>Personal Records by Month</CardTitle>
                </CardHeader>
                <CardContent>
                  <LineChart series={prsSeries} height={260} showLegend={false} />
                </CardContent>
              </Card>
            </div>
          )}

          <Card>
            <CardHeader>
              <CardTitle>Start Times</CardTitle>
            </CardHeader>
            <CardContent>
              <LineChart series={startTimesSeries} height={260} showLegend={false} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader className="flex flex-row items-center justify-between gap-2">
              <CardTitle>Locations</CardTitle>
              <span className="text-xs text-muted-foreground">Effect scatter (lat/lng)</span>
            </CardHeader>
            <CardContent>
              <WorldLocationsChart points={report.locations ?? []} height={380} />
            </CardContent>
          </Card>

          <div className="grid gap-4 md:grid-cols-3">
            <BiggestCard
              title="Longest Distance"
              item={report.biggest.longest_distance}
              format={(v) => formatDistance(v)}
            />
            <BiggestCard
              title="Most Elevation"
              item={report.biggest.most_elevation}
              format={(v) => formatElevation(v)}
            />
            <BiggestCard
              title="Longest Duration"
              item={report.biggest.longest_duration}
              format={(v) => formatDurationLong(Math.round(v))}
            />
          </div>

          {report.random_photo && (
            <Card>
              <CardHeader className="flex flex-row items-center justify-between gap-2">
                <CardTitle>Random Photo</CardTitle>
                <Button variant="outline" size="sm" asChild>
                  <Link
                    to="/activities/$activityId"
                    params={{ activityId: String(report.random_photo.activity_id) }}
                  >
                    View activity
                  </Link>
                </Button>
              </CardHeader>
              <CardContent>
                <img
                  src={report.random_photo.thumbnail_url || report.random_photo.url}
                  alt={report.random_photo.caption || 'Random photo'}
                  className="max-h-[420px] w-full rounded-md object-cover"
                  loading="lazy"
                />
                {report.random_photo.caption && (
                  <p className="mt-2 text-sm text-muted-foreground">
                    {report.random_photo.caption}
                  </p>
                )}
              </CardContent>
            </Card>
          )}

          {compareYear !== null && compare && (
            <Card>
              <CardHeader>
                <CardTitle>Comparison</CardTitle>
              </CardHeader>
              <CardContent>
                <ComparisonTable current={report} baseline={compare} loading={compareLoading} />
              </CardContent>
            </Card>
          )}
        </div>
      )}
    </div>
  )
}

function MetricCard({ title, value }: { title: string; value: string }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent className="text-2xl font-bold">{value}</CardContent>
    </Card>
  )
}

function BiggestCard({
  title,
  item,
  format,
}: {
  title: string
  item: { activity_id: number; name: string; start_date_local: string; value: number } | undefined
  format: (v: number) => string
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-2">
        {!item ? (
          <p className="text-sm text-muted-foreground">No data</p>
        ) : (
          <>
            <div className="text-xl font-bold">{format(item.value)}</div>
            <Link
              to="/activities/$activityId"
              params={{ activityId: String(item.activity_id) }}
              className="block truncate text-sm hover:underline"
            >
              {item.name}
            </Link>
            <div className="text-xs text-muted-foreground">
              {item.start_date_local.slice(0, 10)}
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}

function ComparisonTable({
  current,
  baseline,
  loading,
}: {
  current: {
    year: number
    totals: {
      activities: number
      distance_m: number
      elevation_m: number
      moving_time_s: number
      kudos: number
      carbon_saved_kg: number
    }
  }
  baseline: {
    year: number
    totals: {
      activities: number
      distance_m: number
      elevation_m: number
      moving_time_s: number
      kudos: number
      carbon_saved_kg: number
    }
  }
  loading: boolean
}) {
  if (loading) {
    return <Skeleton className="h-24 w-full" />
  }

  const rows = [
    { label: 'Activities', cur: current.totals.activities, base: baseline.totals.activities },
    {
      label: 'Distance (m)',
      cur: Math.round(current.totals.distance_m),
      base: Math.round(baseline.totals.distance_m),
    },
    {
      label: 'Elevation (m)',
      cur: Math.round(current.totals.elevation_m),
      base: Math.round(baseline.totals.elevation_m),
    },
    {
      label: 'Moving time (s)',
      cur: current.totals.moving_time_s,
      base: baseline.totals.moving_time_s,
    },
    { label: 'Kudos', cur: current.totals.kudos, base: baseline.totals.kudos },
    {
      label: 'CO₂ saved (kg)',
      cur: Math.round(current.totals.carbon_saved_kg * 10) / 10,
      base: Math.round(baseline.totals.carbon_saved_kg * 10) / 10,
    },
  ]

  return (
    <div className="rounded-md border border-border">
      <div className="grid grid-cols-12 gap-2 border-b border-border px-3 py-2 text-xs font-medium text-muted-foreground">
        <div className="col-span-4">Metric</div>
        <div className="col-span-3 text-right">Current</div>
        <div className="col-span-3 text-right">Baseline</div>
        <div className="col-span-2 text-right">Δ</div>
      </div>
      {rows.map((r) => {
        const pct = pctChange(r.cur, r.base)
        return (
          <div key={r.label} className="grid grid-cols-12 gap-2 px-3 py-2 text-sm">
            <div className="col-span-4">{r.label}</div>
            <div className="col-span-3 text-right tabular-nums">{r.cur}</div>
            <div className="col-span-3 text-right tabular-nums text-muted-foreground">{r.base}</div>
            <div className="col-span-2 text-right tabular-nums">
              {pct === null ? '—' : `${pct > 0 ? '+' : ''}${pct.toFixed(1)}%`}
            </div>
          </div>
        )
      })}
    </div>
  )
}
