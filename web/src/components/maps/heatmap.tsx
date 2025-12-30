import { useMemo, useState, useCallback, memo } from 'react'
import { Polyline, Popup, useMapEvents, CircleMarker, Tooltip } from 'react-leaflet'
import type { LeafletMouseEvent, Map as LeafletMap } from 'leaflet'
import { BaseMap } from './base-map'
import { decodePolyline, getBounds } from '@/lib/maps'
import { tileLayers } from '@/lib/maps'
import { cn } from '@/lib/utils'
import type { HeatmapActivity } from '@/lib/api'
import { getSportHexColor, getSportEmoji } from '@/lib/sport-types'
import { Link } from '@tanstack/react-router'
import { formatDistance, formatRelativeDate } from '@/lib/format'

interface HeatmapProps {
  activities: HeatmapActivity[]
  className?: string
  strokeWeight?: number
  colorByActivity?: boolean
  onMapReady?: (map: LeafletMap) => void
}

// Calculate distance between two lat/lng points in meters (Haversine formula)
function haversineDistance(
  lat1: number,
  lng1: number,
  lat2: number,
  lng2: number
): number {
  const R = 6371000 // Earth radius in meters
  const dLat = ((lat2 - lat1) * Math.PI) / 180
  const dLng = ((lng2 - lng1) * Math.PI) / 180
  const a =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos((lat1 * Math.PI) / 180) *
      Math.cos((lat2 * Math.PI) / 180) *
      Math.sin(dLng / 2) *
      Math.sin(dLng / 2)
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
  return R * c
}

interface NearbyResult {
  activity: HeatmapActivity
  distance: number
}

interface ClickedPoint {
  lat: number
  lng: number
  nearby: NearbyResult[]
}

function MapClickHandler({
  activities,
  onClickResult,
  radius = 500, // 500m default radius
}: {
  activities: HeatmapActivity[]
  onClickResult: (result: ClickedPoint | null) => void
  radius?: number
}) {
  useMapEvents({
    click: (e: LeafletMouseEvent) => {
      const { lat, lng } = e.latlng

      // Find activities with start points within radius
      const nearby: NearbyResult[] = activities
        .filter((a) => a.start_lat && a.start_lng)
        .map((activity) => ({
          activity,
          distance: haversineDistance(lat, lng, activity.start_lat, activity.start_lng),
        }))
        .filter((r) => r.distance <= radius)
        .sort((a, b) => a.distance - b.distance)
        .slice(0, 10) // Limit to 10 nearest

      if (nearby.length > 0) {
        onClickResult({ lat, lng, nearby })
      } else {
        onClickResult(null)
      }
    },
  })
  return null
}

// Memoized polyline component for performance
const RoutePolyline = memo(function RoutePolyline({
  route,
  activity,
  isHovered,
  colorByActivity,
  strokeWeight,
  onHover,
  onLeave,
}: {
  route: { id: number; sportType: string; points: [number, number][] }
  activity: HeatmapActivity
  isHovered: boolean
  colorByActivity: boolean
  strokeWeight: number
  onHover: () => void
  onLeave: () => void
}) {
  const color = colorByActivity ? getSportHexColor(route.sportType) : '#fc4c02'

  return (
    <Polyline
      positions={route.points}
      pathOptions={{
        color,
        weight: isHovered ? strokeWeight + 2 : strokeWeight,
        opacity: isHovered ? 0.9 : 0.6,
      }}
      eventHandlers={{
        mouseover: onHover,
        mouseout: onLeave,
      }}
    >
      {isHovered && (
        <Tooltip permanent direction="top" className="heatmap-route-tooltip">
          <div className="text-sm">
            <div className="font-medium flex items-center gap-1">
              <span>{getSportEmoji(activity.sport_type)}</span>
              <span className="truncate max-w-[180px]">{activity.name}</span>
            </div>
            <div className="text-xs text-muted-foreground">
              {formatRelativeDate(activity.start_date)} • {formatDistance(activity.distance / 1000)}
            </div>
          </div>
        </Tooltip>
      )}
    </Polyline>
  )
})

export function Heatmap({
  activities,
  className,
  strokeWeight = 2,
  colorByActivity = true,
  onMapReady,
}: HeatmapProps) {
  const [clickedPoint, setClickedPoint] = useState<ClickedPoint | null>(null)
  const [hoveredActivityId, setHoveredActivityId] = useState<number | null>(null)

  // Create activity lookup map for quick access
  const activityMap = useMemo(() => {
    return new Map(activities.map((a) => [a.id, a]))
  }, [activities])

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

  // Memoized hover handlers - use a Map to cache callbacks by id
  // This prevents creating new function instances on each render
  const hoverHandlers = useMemo(() => {
    const handlers = new Map<number, () => void>()
    decodedRoutes.forEach((route) => {
      handlers.set(route.id, () => setHoveredActivityId(route.id))
    })
    return handlers
  }, [decodedRoutes])

  const handleMouseLeave = useCallback(() => {
    setHoveredActivityId(null)
  }, [])

  const handleClickResult = useCallback((result: ClickedPoint | null) => {
    setClickedPoint(result)
  }, [])

  if (activities.length === 0) {
    return (
      <div className={cn('flex items-center justify-center bg-muted rounded-lg', className)}>
        <p className="text-muted-foreground">No activities with GPS data available</p>
      </div>
    )
  }

  return (
    <div className={cn('rounded-lg overflow-hidden', className)}>
      <BaseMap bounds={bounds} tileLayer={tileLayers.cartoDark} onMapReady={onMapReady}>
        <MapClickHandler
          activities={activities}
          onClickResult={handleClickResult}
        />

        {decodedRoutes.map((route) => {
          const activity = activityMap.get(route.id)
          if (!activity) return null
          const onHover = hoverHandlers.get(route.id)
          if (!onHover) return null
          return (
            <RoutePolyline
              key={route.id}
              route={route}
              activity={activity}
              isHovered={hoveredActivityId === route.id}
              colorByActivity={colorByActivity}
              strokeWeight={strokeWeight}
              onHover={onHover}
              onLeave={handleMouseLeave}
            />
          )
        })}

        {/* Click marker with nearby activities popup */}
        {clickedPoint && (
          <CircleMarker
            center={[clickedPoint.lat, clickedPoint.lng]}
            radius={8}
            pathOptions={{
              color: '#22c55e',
              fillColor: '#22c55e',
              fillOpacity: 0.8,
              weight: 2,
            }}
          >
            <Popup>
              <div className="min-w-[200px] max-w-[280px]">
                <div className="mb-2 font-semibold text-sm">
                  {clickedPoint.nearby.length} activit{clickedPoint.nearby.length === 1 ? 'y' : 'ies'} nearby
                </div>
                <div className="space-y-2 max-h-[200px] overflow-y-auto">
                  {clickedPoint.nearby.map(({ activity, distance }) => (
                    <Link
                      key={activity.id}
                      to="/activities/$activityId"
                      params={{ activityId: String(activity.id) }}
                      className="flex items-center gap-2 p-2 rounded hover:bg-muted/50 transition-colors text-sm"
                    >
                      <span className="text-base">{getSportEmoji(activity.sport_type)}</span>
                      <div className="flex-1 min-w-0">
                        <div className="font-medium truncate text-foreground">
                          {activity.name || `#${activity.id}`}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {formatDistance(distance / 1000)} away • {activity.sport_type}
                        </div>
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            </Popup>
          </CircleMarker>
        )}
      </BaseMap>
    </div>
  )
}

// Helper to calculate bounds for a set of activities
export function getActivitiesBounds(activities: HeatmapActivity[]): [[number, number], [number, number]] | undefined {
  const points = activities
    .filter((a) => a.start_lat && a.start_lng)
    .map((a) => [a.start_lat, a.start_lng] as [number, number])

  if (points.length === 0) return undefined
  return getBounds(points)
}
