import { useMemo } from 'react'
import { Polyline, CircleMarker, Tooltip } from 'react-leaflet'
import { BaseMap } from './base-map'
import { decodePolyline, getBounds } from '@/lib/maps'
import { cn } from '@/lib/utils'

interface ActivityMapProps {
  polyline: string
  className?: string
  showMarkers?: boolean
  strokeColor?: string
  strokeWeight?: number
}

export function ActivityMap({
  polyline,
  className,
  showMarkers = true,
  strokeColor = '#fc4c02',
  strokeWeight = 3,
}: ActivityMapProps) {
  const { points, bounds, startPoint, endPoint } = useMemo(() => {
    const decoded = decodePolyline(polyline)
    return {
      points: decoded,
      bounds: getBounds(decoded),
      startPoint: decoded[0],
      endPoint: decoded[decoded.length - 1],
    }
  }, [polyline])

  if (points.length === 0) {
    return (
      <div
        className={cn(
          'flex items-center justify-center bg-muted rounded-lg',
          className
        )}
      >
        <p className="text-muted-foreground">No route data available</p>
      </div>
    )
  }

  return (
    <div className={cn('rounded-lg overflow-hidden', className)}>
      <BaseMap bounds={bounds}>
        <Polyline
          positions={points}
          pathOptions={{
            color: strokeColor,
            weight: strokeWeight,
            opacity: 0.9,
          }}
        />
        {showMarkers && startPoint && (
          <CircleMarker
            center={startPoint}
            radius={8}
            pathOptions={{
              color: '#ffffff',
              fillColor: '#22c55e',
              fillOpacity: 1,
              weight: 2,
            }}
          >
            <Tooltip permanent={false}>Start</Tooltip>
          </CircleMarker>
        )}
        {showMarkers && endPoint && (
          <CircleMarker
            center={endPoint}
            radius={8}
            pathOptions={{
              color: '#ffffff',
              fillColor: '#ef4444',
              fillOpacity: 1,
              weight: 2,
            }}
          >
            <Tooltip permanent={false}>End</Tooltip>
          </CircleMarker>
        )}
      </BaseMap>
    </div>
  )
}
