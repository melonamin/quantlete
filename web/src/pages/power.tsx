import { useState } from 'react'
import { Link } from '@tanstack/react-router'
import { usePowerStats, usePowerStatsFiltered, usePowerZones } from '@/lib/api'
import { BarChart, LineChart, PowerZonesChart } from '@/components/charts'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

function formatDurationLabel(seconds: number) {
  if (seconds < 60) return `${seconds}s`
  if (seconds === 3600) return '1h'
  return `${Math.round(seconds / 60)}m`
}

export function PowerPage() {
  const { data, isLoading, error } = usePowerStats()
  const after90 = new Date(Date.now() - 90 * 24 * 3600 * 1000).toISOString().slice(0, 10)
  const { data: last90 } = usePowerStatsFiltered({ after: after90 })
  const { data: powerZones } = usePowerZones()
  const [selectedDuration, setSelectedDuration] = useState<number>(300)

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <p className="text-destructive">Failed to load power stats.</p>
      </div>
    )
  }

  const bestByDuration = new Map<number, number>()
  for (const b of data?.best ?? []) bestByDuration.set(b.duration_s, b.watts)

  const chartData =
    data?.durations_s?.map((d) => ({
      label: formatDurationLabel(d),
      value: Math.round(bestByDuration.get(d) ?? 0),
    })) ?? []

  const historyKey = String(selectedDuration)
  const history = data?.history?.[historyKey] ?? []

  const bestCurve = (resp: typeof data) => {
    const best = new Map<number, number>()
    for (const b of resp?.best ?? []) best.set(b.duration_s, b.watts)
    return (resp?.durations_s ?? []).map((d) => ({
      x: formatDurationLabel(d),
      y: Math.round(best.get(d) ?? 0),
    }))
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Power</h1>
          <p className="text-muted-foreground">Peak power outputs and progression</p>
        </div>
        <Button asChild variant="outline">
          <Link to="/">Back</Link>
        </Button>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>All-time Best</CardTitle>
          </CardHeader>
          <CardContent>
            <BarChart data={chartData} height={260} loading={isLoading} showValues />
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>Progression</CardTitle>
            <select
              className="h-9 rounded-md border border-border bg-background px-2 text-sm"
              value={selectedDuration}
              onChange={(e) => setSelectedDuration(Number(e.target.value))}
            >
              {(data?.durations_s ?? []).map((d) => (
                <option key={d} value={d}>
                  {formatDurationLabel(d)}
                </option>
              ))}
            </select>
          </CardHeader>
          <CardContent>
            <LineChart
              series={[
                {
                  name: 'Best',
                  data: history.map((p) => ({ x: p.date, y: p.watts })),
                },
              ]}
              height={260}
              loading={isLoading}
              showLegend={false}
              showDataZoom
            />
          </CardContent>
        </Card>
      </div>

      <div className="mt-6">
        <Card>
          <CardHeader>
            <CardTitle>Power Curve Comparison</CardTitle>
          </CardHeader>
          <CardContent>
            <LineChart
              series={[
                { name: 'All-time', data: bestCurve(data) },
                { name: 'Last 90 days', data: bestCurve(last90) },
              ]}
              height={260}
              loading={isLoading}
              showLegend
              showDataZoom={false}
            />
          </CardContent>
        </Card>
      </div>

      <div className="mt-6">
        <Card>
          <CardHeader>
            <CardTitle>Power Zones</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <PowerZonesChart
              secondsByZone={powerZones?.seconds_by_zone ?? [0, 0, 0, 0, 0]}
              loading={isLoading}
            />
            {powerZones?.ftp_watts && (
              <p className="text-xs text-muted-foreground">
                Based on FTP {Math.round(powerZones.ftp_watts)}W.
              </p>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
