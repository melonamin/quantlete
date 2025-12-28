import { useMemo, useState } from 'react'
import { Polyline, CircleMarker, Tooltip } from 'react-leaflet'
import { BaseMap } from './base-map'
import { decodePolyline, getBounds, type TileLayer } from '@/lib/maps'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'

interface ActivityMapProps {
  polyline: string
  className?: string
  showMarkers?: boolean
  strokeColor?: string
  strokeWeight?: number
  latlngStream?: unknown
  altitudeStream?: unknown
  tileLayer?: TileLayer
}

function toLatLngArray(v: unknown): [number, number][] {
  if (!Array.isArray(v)) return []
  const out: [number, number][] = []
  for (const item of v) {
    if (!Array.isArray(item) || item.length < 2) continue
    const lat = Number(item[0])
    const lng = Number(item[1])
    if (!Number.isFinite(lat) || !Number.isFinite(lng)) continue
    out.push([lat, lng])
  }
  return out
}

function toNumberArray(v: unknown): number[] {
  if (!Array.isArray(v)) return []
  return v.map((x) => (typeof x === 'number' ? x : Number(x))).filter((n) => Number.isFinite(n))
}

function haversineMeters(a: [number, number], b: [number, number]) {
  const R = 6371000
  const toRad = (d: number) => (d * Math.PI) / 180
  const dLat = toRad(b[0] - a[0])
  const dLng = toRad(b[1] - a[1])
  const lat1 = toRad(a[0])
  const lat2 = toRad(b[0])
  const x =
    Math.sin(dLat / 2) * Math.sin(dLat / 2) +
    Math.cos(lat1) * Math.cos(lat2) * Math.sin(dLng / 2) * Math.sin(dLng / 2)
  return 2 * R * Math.atan2(Math.sqrt(x), Math.sqrt(1 - x))
}

function gradeColor(gradePct: number) {
  // Clamp to [-10, 10] and map to a 7-step diverging palette.
  const g = Math.max(-10, Math.min(10, gradePct))
  const palette = ['#2563eb', '#60a5fa', '#86efac', '#22c55e', '#facc15', '#fb923c', '#ef4444']
  const idx = Math.round(((g + 10) / 20) * (palette.length - 1))
  return palette[idx]
}

export function ActivityMap({
  polyline,
  className,
  showMarkers = true,
  strokeColor = '#fc4c02',
  strokeWeight = 3,
  latlngStream,
  altitudeStream,
  tileLayer,
}: ActivityMapProps) {
  const [elevationMode, setElevationMode] = useState(false)

  const { points, bounds, startPoint, endPoint, segments, hasElevation } = useMemo(() => {
    const latlng = toLatLngArray(latlngStream)
    const decoded = latlng.length ? latlng : decodePolyline(polyline)
    const altitude = toNumberArray(altitudeStream)
    const canColor = altitude.length === decoded.length && decoded.length > 2

    const segs: { points: [number, number][]; color: string }[] = []
    if (canColor) {
      // Group consecutive segments by color bucket to reduce Polyline count.
      let current: { points: [number, number][]; color: string } | null = null
      for (let i = 0; i < decoded.length - 1; i++) {
        const d = haversineMeters(decoded[i], decoded[i + 1])
        if (d <= 0.5) continue
        const gradePct = ((altitude[i + 1] - altitude[i]) / d) * 100
        const color = gradeColor(gradePct)
        if (!current || current.color !== color) {
          current = { points: [decoded[i], decoded[i + 1]], color }
          segs.push(current)
        } else {
          current.points.push(decoded[i + 1])
        }
      }
    }

    return {
      points: decoded,
      bounds: getBounds(decoded),
      startPoint: decoded[0],
      endPoint: decoded[decoded.length - 1],
      segments: segs,
      hasElevation: canColor,
    }
  }, [polyline, latlngStream, altitudeStream])

  if (points.length === 0) {
    return (
      <div className={cn('flex items-center justify-center bg-muted rounded-lg', className)}>
        <p className="text-muted-foreground">No route data available</p>
      </div>
    )
  }

  return (
    <div className={cn('rounded-lg overflow-hidden', className)}>
      <BaseMap bounds={bounds} tileLayer={tileLayer}>
        {hasElevation && elevationMode ? (
          segments.map((s, idx) => (
            <Polyline
              key={idx}
              positions={s.points}
              pathOptions={{
                color: s.color,
                weight: strokeWeight,
                opacity: 0.95,
              }}
            />
          ))
        ) : (
          <Polyline
            positions={points}
            pathOptions={{
              color: strokeColor,
              weight: strokeWeight,
              opacity: 0.9,
            }}
          />
        )}

        {hasElevation && (
          <div className="leaflet-top leaflet-right">
            <div className="m-2 flex flex-col gap-2 rounded-md border border-border bg-background/90 p-2 backdrop-blur">
              <Button size="sm" variant="outline" onClick={() => setElevationMode(!elevationMode)}>
                {elevationMode ? 'Flat color' : 'Elevation'}
              </Button>
              {elevationMode && (
                <div className="space-y-1">
                  <div className="text-[10px] text-muted-foreground">Grade</div>
                  <div className="h-2 w-32 overflow-hidden rounded-full bg-muted">
                    <div
                      className="h-full"
                      style={{
                        background:
                          'linear-gradient(90deg, #2563eb, #60a5fa, #86efac, #22c55e, #facc15, #fb923c, #ef4444)',
                      }}
                    />
                  </div>
                  <div className="flex justify-between text-[10px] text-muted-foreground">
                    <span>-10%</span>
                    <span>+10%</span>
                  </div>
                </div>
              )}
            </div>
          </div>
        )}

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
