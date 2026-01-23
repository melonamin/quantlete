import { useMemo } from 'react'
import { useActivities } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatDuration } from '@/lib/format'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { Bike, Route, Clock, Mountain } from 'lucide-react'

// Known Zwift worlds based on location data
const ZWIFT_WORLDS: Record<string, string> = {
  Watopia: 'Watopia',
  London: 'London',
  'New York': 'New York',
  Innsbruck: 'Innsbruck',
  Richmond: 'Richmond',
  Yorkshire: 'Yorkshire',
  France: 'France',
  Paris: 'Paris',
  'Makuri Islands': 'Makuri',
  Scotland: 'Scotland',
}

interface WorldStats {
  world: string
  count: number
  distance: number
  elevation: number
  time: number
}

export function ZwiftStats() {
  const { data, isLoading } = useActivities({
    sport_type: 'VirtualRide,VirtualRun',
    per_page: 1000,
  })
  const { formatDistance, formatElevation } = useFormattedMetrics()

  const worldStats = useMemo(() => {
    // Use Array.isArray for defensive check against unexpected data shapes
    const activities = Array.isArray(data?.data) ? data.data : []
    if (activities.length === 0) return []

    const stats: Record<string, WorldStats> = {}

    for (const activity of activities) {
      // Try to detect world from location
      let world = 'Unknown'
      const city = activity.location_city || ''
      const country = activity.location_country || ''
      const location = `${city} ${country}`.toLowerCase()

      for (const [key, label] of Object.entries(ZWIFT_WORLDS)) {
        if (location.includes(key.toLowerCase())) {
          world = label
          break
        }
      }

      // Also check device name for platform hints
      const deviceName = (activity.device_name || '').toLowerCase()
      if (deviceName.includes('zwift') && world === 'Unknown') {
        world = 'Zwift (Unknown World)'
      }
      if (deviceName.includes('rouvy') || location.includes('rouvy')) {
        world = 'Rouvy'
      }
      if (deviceName.includes('mywhoosh') || location.includes('mywhoosh')) {
        world = 'MyWhoosh'
      }

      if (!stats[world]) {
        stats[world] = { world, count: 0, distance: 0, elevation: 0, time: 0 }
      }

      stats[world].count++
      stats[world].distance += activity.distance
      stats[world].elevation += activity.total_elevation_gain
      stats[world].time += activity.moving_time
    }

    return Object.values(stats).sort((a, b) => b.distance - a.distance)
  }, [data])

  const totals = useMemo(() => {
    return worldStats.reduce(
      (acc, w) => ({
        count: acc.count + w.count,
        distance: acc.distance + w.distance,
        elevation: acc.elevation + w.elevation,
        time: acc.time + w.time,
      }),
      { count: 0, distance: 0, elevation: 0, time: 0 }
    )
  }, [worldStats])

  return (
    <WidgetWrapper title="Virtual Riding" isLoading={isLoading}>
      {worldStats.length === 0 ? (
        <div className="text-sm text-muted-foreground">No virtual activities found.</div>
      ) : (
        <div className="h-full flex flex-col">
          {/* Summary stats */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2 mb-4 flex-shrink-0">
            <div className="flex items-center gap-1.5">
              <Bike className="h-3.5 w-3.5 text-muted-foreground" />
              <div>
                <div className="text-xs text-muted-foreground">Workouts</div>
                <div className="font-semibold">{totals.count}</div>
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              <Route className="h-3.5 w-3.5 text-muted-foreground" />
              <div>
                <div className="text-xs text-muted-foreground">Distance</div>
                <div className="font-semibold">{formatDistance(totals.distance)}</div>
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              <Mountain className="h-3.5 w-3.5 text-muted-foreground" />
              <div>
                <div className="text-xs text-muted-foreground">Elevation</div>
                <div className="font-semibold">{formatElevation(totals.elevation)}</div>
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              <Clock className="h-3.5 w-3.5 text-muted-foreground" />
              <div>
                <div className="text-xs text-muted-foreground">Time</div>
                <div className="font-semibold">{formatDuration(totals.time)}</div>
              </div>
            </div>
          </div>

          {/* World breakdown */}
          <div className="flex-1 min-h-0 overflow-y-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>World</TableHead>
                  <TableHead className="text-right">Rides</TableHead>
                  <TableHead className="text-right">Distance</TableHead>
                  <TableHead className="text-right">Elev</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {worldStats.map((w) => (
                  <TableRow key={w.world}>
                    <TableCell>{w.world}</TableCell>
                    <TableCell className="text-right">{w.count}</TableCell>
                    <TableCell className="text-right">{formatDistance(w.distance)}</TableCell>
                    <TableCell className="text-right">{formatElevation(w.elevation)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>
      )}
    </WidgetWrapper>
  )
}
