import { useMemo } from 'react'
import { useActivities } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { formatDistance, formatDuration } from '@/lib/format'
import { Bike, Route, Clock, Mountain } from 'lucide-react'

// Known Zwift worlds based on location data
const ZWIFT_WORLDS: Record<string, string> = {
  'Watopia': 'Watopia',
  'London': 'London',
  'New York': 'New York',
  'Innsbruck': 'Innsbruck',
  'Richmond': 'Richmond',
  'Yorkshire': 'Yorkshire',
  'France': 'France',
  'Paris': 'Paris',
  'Makuri Islands': 'Makuri',
  'Scotland': 'Scotland',
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

  const worldStats = useMemo(() => {
    if (!data?.data) return []

    const stats: Record<string, WorldStats> = {}

    for (const activity of data.data) {
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
  }, [data?.data])

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
        <div className="text-sm text-muted-foreground">
          No virtual activities found.
        </div>
      ) : (
        <div className="h-full flex flex-col">
          {/* Summary stats */}
          <div className="grid grid-cols-4 gap-2 mb-4 flex-shrink-0">
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
                <div className="font-semibold">{formatDistance(totals.distance / 1000)}</div>
              </div>
            </div>
            <div className="flex items-center gap-1.5">
              <Mountain className="h-3.5 w-3.5 text-muted-foreground" />
              <div>
                <div className="text-xs text-muted-foreground">Elevation</div>
                <div className="font-semibold">{Math.round(totals.elevation).toLocaleString()}m</div>
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
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-border text-left text-xs text-muted-foreground">
                  <th className="pb-2 font-medium">World</th>
                  <th className="pb-2 font-medium text-right">Rides</th>
                  <th className="pb-2 font-medium text-right">Distance</th>
                  <th className="pb-2 font-medium text-right">Elev</th>
                </tr>
              </thead>
              <tbody>
                {worldStats.map((w) => (
                  <tr key={w.world} className="border-b border-border/50">
                    <td className="py-2">{w.world}</td>
                    <td className="py-2 text-right">{w.count}</td>
                    <td className="py-2 text-right">{formatDistance(w.distance / 1000)}</td>
                    <td className="py-2 text-right">{Math.round(w.elevation).toLocaleString()}m</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </WidgetWrapper>
  )
}
