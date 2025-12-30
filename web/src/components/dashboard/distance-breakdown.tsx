import { useState } from 'react'
import { useActivities } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { BarChart } from '@/components/charts'
import { formatDistance, formatDuration } from '@/lib/format'
import { Button } from '@/components/ui/button'

interface DistanceZone {
  label: string
  min: number // in meters
  max: number // in meters
  count: number
  totalDistance: number
  totalTime: number
  totalElevation: number
}

const DISTANCE_ZONES = [
  { label: '0-5 km', min: 0, max: 5000 },
  { label: '5-10 km', min: 5000, max: 10000 },
  { label: '10-20 km', min: 10000, max: 20000 },
  { label: '20-50 km', min: 20000, max: 50000 },
  { label: '50-100 km', min: 50000, max: 100000 },
  { label: '100+ km', min: 100000, max: Infinity },
]

type ViewMode = 'chart' | 'table'

export function DistanceBreakdown() {
  const [viewMode, setViewMode] = useState<ViewMode>('chart')

  // Fetch all activities (we'll process them client-side for simplicity)
  // In a production app, this would be a dedicated backend endpoint
  const { data, isLoading } = useActivities({ page: 1, per_page: 10000 })

  const zones: DistanceZone[] = DISTANCE_ZONES.map((zone) => ({
    ...zone,
    count: 0,
    totalDistance: 0,
    totalTime: 0,
    totalElevation: 0,
  }))

  if (data?.data) {
    for (const activity of data.data) {
      const distance = activity.distance ?? 0
      for (const zone of zones) {
        if (distance >= zone.min && distance < zone.max) {
          zone.count++
          zone.totalDistance += distance
          zone.totalTime += activity.moving_time ?? 0
          zone.totalElevation += activity.total_elevation_gain ?? 0
          break
        }
      }
    }
  }

  const chartData = zones.map((zone) => ({
    label: zone.label,
    value: zone.count,
  }))

  const hasData = zones.some((z) => z.count > 0)

  return (
    <WidgetWrapper
      title="Distance Breakdown"
      isLoading={isLoading}
      action={
        <div className="flex gap-1">
          <Button
            variant={viewMode === 'chart' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-7 px-2.5 text-xs"
            onClick={() => setViewMode('chart')}
          >
            Chart
          </Button>
          <Button
            variant={viewMode === 'table' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-7 px-2.5 text-xs"
            onClick={() => setViewMode('table')}
          >
            Table
          </Button>
        </div>
      }
    >
      {!hasData ? (
        <div className="text-sm text-muted-foreground">No activities to analyze.</div>
      ) : viewMode === 'chart' ? (
        <div className="h-full min-h-0">
          <BarChart data={chartData} height="100%" showValues />
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b text-muted-foreground">
                <th className="py-2 text-left font-medium">Distance</th>
                <th className="py-2 text-right font-medium">Count</th>
                <th className="py-2 text-right font-medium">Total</th>
                <th className="py-2 text-right font-medium">Time</th>
                <th className="py-2 text-right font-medium">Elev</th>
              </tr>
            </thead>
            <tbody>
              {zones.map((zone) => (
                <tr key={zone.label} className="border-b border-border/50 last:border-0">
                  <td className="py-1.5 font-medium">{zone.label}</td>
                  <td className="py-1.5 text-right">{zone.count}</td>
                  <td className="py-1.5 text-right">{formatDistance(zone.totalDistance)}</td>
                  <td className="py-1.5 text-right">{formatDuration(zone.totalTime)}</td>
                  <td className="py-1.5 text-right">{zone.totalElevation.toLocaleString()}m</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </WidgetWrapper>
  )
}
