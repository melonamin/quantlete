import { useState } from 'react'
import { useActivities } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { BarChart } from '@/components/charts'
import { formatDuration } from '@/lib/format'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

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
  const { formatDistance, formatElevation } = useFormattedMetrics()

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

  // Use Array.isArray for defensive check against unexpected data shapes
  const activities = Array.isArray(data?.data) ? data.data : []
  for (const activity of activities) {
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
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Distance</TableHead>
              <TableHead className="text-right">Count</TableHead>
              <TableHead className="text-right">Total</TableHead>
              <TableHead className="text-right">Time</TableHead>
              <TableHead className="text-right">Elev</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {zones.map((zone) => (
              <TableRow key={zone.label}>
                <TableCell className="font-medium">{zone.label}</TableCell>
                <TableCell className="text-right">{zone.count}</TableCell>
                <TableCell className="text-right">{formatDistance(zone.totalDistance)}</TableCell>
                <TableCell className="text-right">{formatDuration(zone.totalTime)}</TableCell>
                <TableCell className="text-right">
                  {formatElevation(zone.totalElevation)}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </WidgetWrapper>
  )
}
