import { useMemo } from 'react'
import { Polyline, CircleMarker, Tooltip } from 'react-leaflet'
import { BaseMap } from './base-map'
import { decodePolyline, getBounds } from '@/lib/maps'
import { cn } from '@/lib/utils'

function gradeColor(gradePct: number) {
  const g = Math.max(-10, Math.min(20, gradePct))
  const palette = ['#22c55e', '#86efac', '#facc15', '#fb923c', '#ef4444']
  const idx = Math.round(((g + 10) / 30) * (palette.length - 1))
  return palette[Math.max(0, Math.min(palette.length - 1, idx))]
}

export function SegmentMap({
  polyline,
  className,
  strokeColor,
  strokeWeight = 4,
  averageGrade,
  maximumGrade,
  climbCategory,
}: {
  polyline: string
  className?: string
  strokeColor?: string
  strokeWeight?: number
  averageGrade?: number
  maximumGrade?: number
  climbCategory?: number
}) {
  const { points, bounds, startPoint, endPoint } = useMemo(() => {
    const decoded = decodePolyline(polyline)
    return {
      points: decoded,
      bounds: getBounds(decoded),
      startPoint: decoded[0],
      endPoint: decoded[decoded.length - 1],
    }
  }, [polyline])

  const color =
    strokeColor ?? (typeof averageGrade === 'number' ? gradeColor(averageGrade) : '#3b82f6')

  if (points.length === 0) {
    return (
      <div className={cn('flex items-center justify-center bg-muted rounded-lg', className)}>
        <p className="text-muted-foreground">No route data available</p>
      </div>
    )
  }

  return (
    <div className={cn('relative rounded-lg overflow-hidden', className)}>
      <BaseMap bounds={bounds}>
        <Polyline
          positions={points}
          pathOptions={{
            color: color,
            weight: strokeWeight,
            opacity: 0.9,
          }}
        />
        {startPoint && (
          <CircleMarker
            center={startPoint}
            radius={7}
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
        {endPoint && (
          <CircleMarker
            center={endPoint}
            radius={7}
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

      {(typeof averageGrade === 'number' ||
        typeof maximumGrade === 'number' ||
        typeof climbCategory === 'number') && (
        <div className="absolute right-2 top-2 rounded-md border border-border bg-background/90 px-2 py-1 text-xs text-muted-foreground shadow backdrop-blur">
          <div className="flex items-center gap-2">
            {typeof averageGrade === 'number' && <span>Avg {averageGrade.toFixed(1)}%</span>}
            {typeof maximumGrade === 'number' && <span>Max {maximumGrade.toFixed(1)}%</span>}
            {typeof climbCategory === 'number' && climbCategory > 0 && (
              <span>Cat {climbCategory}</span>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
