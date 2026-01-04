import { useEffect, type ReactNode } from 'react'
import { MapContainer, TileLayer, useMap } from 'react-leaflet'
import type { LatLngBoundsExpression, LatLngExpression, Map } from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { defaultTileLayer, type TileLayer as TileLayerConfig } from '@/lib/maps'

interface BaseMapProps {
  center?: LatLngExpression
  zoom?: number
  bounds?: LatLngBoundsExpression
  className?: string
  children?: ReactNode
  tileLayer?: TileLayerConfig
  onMapReady?: (map: Map) => void
}

function MapEvents({ onMapReady }: { onMapReady?: (map: Map) => void }) {
  const map = useMap()

  useEffect(() => {
    if (onMapReady) {
      onMapReady(map)
    }
  }, [map, onMapReady])

  return null
}

function FitBounds({ bounds }: { bounds: LatLngBoundsExpression }) {
  const map = useMap()

  useEffect(() => {
    map.fitBounds(bounds, { padding: [20, 20] })
  }, [map, bounds])

  return null
}

export function BaseMap({
  center = [51.505, -0.09],
  zoom = 13,
  bounds,
  className = 'h-[400px] w-full',
  children,
  tileLayer = defaultTileLayer,
  onMapReady,
}: BaseMapProps) {
  return (
    <MapContainer
      center={center}
      zoom={zoom}
      className={className}
      scrollWheelZoom={true}
      style={{ height: '100%', width: '100%' }}
    >
      <TileLayer
        attribution={tileLayer.attribution}
        url={tileLayer.url}
        maxZoom={tileLayer.maxZoom}
        crossOrigin="anonymous"
      />
      {bounds && <FitBounds bounds={bounds} />}
      {onMapReady && <MapEvents onMapReady={onMapReady} />}
      {children}
    </MapContainer>
  )
}
