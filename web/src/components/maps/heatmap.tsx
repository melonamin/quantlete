import { useMemo } from 'react'
import { Polyline } from 'react-leaflet'
import { BaseMap } from './base-map'
import { decodePolyline, getBounds } from '@/lib/maps'
import { tileLayers } from '@/lib/maps'
import { cn } from '@/lib/utils'
import type { HeatmapActivity } from '@/lib/api'
import { getSportHexColor } from '@/lib/sport-types'

interface HeatmapProps {
  activities: HeatmapActivity[]
  className?: string
  strokeWeight?: number
  colorByActivity?: boolean
}

export function Heatmap({
  activities,
  className,
  strokeWeight = 2,
  colorByActivity = true,
}: HeatmapProps) {
  const { decodedRoutes, bounds } = useMemo(() => {
    const routes = activities
      .filter((a) => a.summary_polyline)
      .map((activity) => ({
        id: activity.id,
        sportType: activity.sport_type,
        points: decodePolyline(activity.summary_polyline),
      }))
      .filter((r) => r.points.length > 0)

    // Calculate overall bounds
    const allPoints = routes.flatMap((r) => r.points)
    const calculatedBounds = getBounds(allPoints)

    return {
      decodedRoutes: routes,
      bounds: allPoints.length > 0 ? calculatedBounds : undefined,
    }
  }, [activities])

  if (activities.length === 0) {
    return (
      <div
        className={cn(
          'flex items-center justify-center bg-muted rounded-lg',
          className
        )}
      >
        <p className="text-muted-foreground">
          No activities with GPS data available
        </p>
      </div>
    )
  }

  return (
    <div className={cn('rounded-lg overflow-hidden', className)}>
      <BaseMap bounds={bounds} tileLayer={tileLayers.cartoDark}>
        {decodedRoutes.map((route) => (
          <Polyline
            key={route.id}
            positions={route.points}
            pathOptions={{
              color: colorByActivity
                ? getSportHexColor(route.sportType)
                : '#fc4c02',
              weight: strokeWeight,
              opacity: 0.6,
            }}
          />
        ))}
      </BaseMap>
    </div>
  )
}
