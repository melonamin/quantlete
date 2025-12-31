import { useParams, Link } from '@tanstack/react-router'
import {
  useActivity,
  useActivityPhotos,
  useActivityStreams,
  useAppSettings,
  useAuthStatus,
} from '@/lib/api'
import { ActivityHeader, ActivityStats, WeatherBadge } from '@/components/activities'
import { ActivityMap } from '@/components/maps'
import { Checkbox } from '@/components/ui/checkbox'
import { Skeleton } from '@/components/ui/skeleton'
import { ChevronLeft } from 'lucide-react'
import { ActivityStreamProfileChart, ElevationProfileChart } from '@/components/charts'
import type { TileLayer } from '@/lib/maps'
import { Button } from '@/components/ui/button'
import { useMemo, useState } from 'react'

export function ActivityDetailPage() {
  const { activityId } = useParams({ from: '/activities/$activityId' })
  const activityNum = Number(activityId)
  const { data: authStatus } = useAuthStatus()
  const isAuthenticated = authStatus?.authenticated
  const { data: activity, isLoading, error } = useActivity(activityNum)
  const { data: streams, isLoading: streamsLoading } = useActivityStreams(activityNum)
  const { data: photos, isLoading: photosLoading } = useActivityPhotos(activityNum, activityNum > 0)
  const { data: settings } = useAppSettings({ enabled: !!isAuthenticated })

  const streamMap = useMemo(
    () =>
      (streams ?? []).reduce<Record<string, unknown>>((acc, s) => {
        acc[s.stream_type] = s.data
        return acc
      }, {}),
    [streams]
  )
  const hasStreams = !!streamMap.distance || !!streamMap.time

  const hasHr = Array.isArray(streamMap.heartrate) && streamMap.heartrate.length > 0
  const hasWatts = Array.isArray(streamMap.watts) && streamMap.watts.length > 0
  const hasCadence = Array.isArray(streamMap.cadence) && streamMap.cadence.length > 0
  const hasAltitude = Array.isArray(streamMap.altitude) && streamMap.altitude.length > 0

  const [showHr, setShowHr] = useState(true)
  const [showWatts, setShowWatts] = useState(true)
  const [showCadence, setShowCadence] = useState(true)
  const [showAltitude, setShowAltitude] = useState(true)
  const [elevGradient, setElevGradient] = useState(true)

  const virtualWorld = activity
    ? detectVirtualWorld(activity.sport_type, activity.name, activity.device_name)
    : null
  const virtualTile = virtualWorld ? settings?.virtual_world_tile_layers?.[virtualWorld] : undefined
  const tileLayer: TileLayer | undefined = virtualTile
    ? {
        name: virtualTile.name,
        url: virtualTile.url,
        attribution: virtualTile.attribution ?? '',
        maxZoom: virtualTile.max_zoom,
      }
    : undefined

  if (isLoading) {
    return <ActivityDetailSkeleton />
  }

  if (error || !activity) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="mb-4">
          <Link
            to="/activities"
            className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground"
          >
            <ChevronLeft className="h-4 w-4 mr-1" />
            Back to Activities
          </Link>
        </div>
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">
            {error ? `Failed to load activity: ${error.message}` : 'Activity not found'}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-4">
        <Link
          to="/activities"
          className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground"
        >
          <ChevronLeft className="h-4 w-4 mr-1" />
          Back to Activities
        </Link>
      </div>

      <ActivityHeader activity={activity} />
      <ActivityStats activity={activity} />

      <div className="mt-6">
        <WeatherBadge activityId={activityNum} />
      </div>

      {activity.summary_polyline && (
        <div className="mt-6">
          <ActivityMap
            polyline={activity.summary_polyline}
            latlngStream={streamMap.latlng}
            altitudeStream={streamMap.altitude}
            tileLayer={tileLayer}
            className="h-[400px]"
          />
          {virtualWorld && !tileLayer && (
            <p className="mt-2 text-xs text-muted-foreground">
              Virtual world detected: {virtualWorld}. Configure a tile layer in Settings to switch
              maps.
            </p>
          )}
        </div>
      )}

      {hasStreams && (
        <div className="mt-6 grid gap-6 lg:grid-cols-2">
          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-3 flex flex-wrap items-center justify-between gap-3">
              <h2 className="text-sm font-medium">Activity Streams</h2>
              <div className="flex flex-wrap items-center gap-3 text-xs text-muted-foreground">
                <label className="flex items-center gap-1">
                  <Checkbox
                    checked={showHr}
                    onCheckedChange={(checked) => setShowHr(!!checked)}
                    disabled={!hasHr}
                  />
                  HR
                </label>
                <label className="flex items-center gap-1">
                  <Checkbox
                    checked={showWatts}
                    onCheckedChange={(checked) => setShowWatts(!!checked)}
                    disabled={!hasWatts}
                  />
                  Power
                </label>
                <label className="flex items-center gap-1">
                  <Checkbox
                    checked={showCadence}
                    onCheckedChange={(checked) => setShowCadence(!!checked)}
                    disabled={!hasCadence}
                  />
                  Cadence
                </label>
                <label className="flex items-center gap-1">
                  <Checkbox
                    checked={showAltitude}
                    onCheckedChange={(checked) => setShowAltitude(!!checked)}
                    disabled={!hasAltitude}
                  />
                  Elev
                </label>
              </div>
            </div>
            <ActivityStreamProfileChart
              streams={streamMap}
              loading={streamsLoading}
              enabledSeries={{
                heartrate: showHr,
                watts: showWatts,
                cadence: showCadence,
                altitude: showAltitude,
              }}
            />
          </div>
          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-3 flex items-center justify-between gap-3">
              <h2 className="text-sm font-medium">Elevation Profile</h2>
              <label className="flex items-center gap-2 text-xs text-muted-foreground">
                <Checkbox
                  checked={elevGradient}
                  onCheckedChange={(checked) => setElevGradient(!!checked)}
                  disabled={!hasAltitude}
                />
                Gradient
              </label>
            </div>
            <ElevationProfileChart
              distance={streamMap.distance}
              altitude={streamMap.altitude}
              gradient={elevGradient}
              loading={streamsLoading}
            />
          </div>
        </div>
      )}

      {(photosLoading || (photos && photos.length > 0)) && (
        <div className="mt-6 rounded-lg border border-border bg-card p-4">
          <div className="mb-3 flex items-center justify-between gap-4">
            <h2 className="text-sm font-medium">Photos</h2>
            <Button variant="ghost" size="sm" asChild>
              <Link to="/photos">View all</Link>
            </Button>
          </div>
          {photosLoading ? (
            <div className="grid grid-cols-3 gap-2 sm:grid-cols-4 lg:grid-cols-6">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="aspect-square w-full" />
              ))}
            </div>
          ) : (
            <div className="grid grid-cols-3 gap-2 sm:grid-cols-4 lg:grid-cols-6">
              {(photos ?? []).slice(0, 12).map((p) => (
                <a
                  key={p.id}
                  href={p.url}
                  target="_blank"
                  rel="noreferrer"
                  className="block overflow-hidden rounded-md border border-border"
                >
                  <img
                    src={p.thumbnail_url || p.url}
                    alt={p.caption || activity.name}
                    loading="lazy"
                    className="aspect-square w-full object-cover"
                  />
                </a>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

function detectVirtualWorld(sportType: string, name: string, deviceName?: string) {
  const isVirtual =
    sportType.includes('Virtual') ||
    (deviceName ? deviceName.toLowerCase().includes('zwift') : false)
  if (!isVirtual) return null
  const hay = `${name}`.toLowerCase()
  const worlds: { key: string; match: string[] }[] = [
    { key: 'Watopia', match: ['watopia'] },
    { key: 'London', match: ['london'] },
    { key: 'New York', match: ['new york', 'nyc'] },
    { key: 'Innsbruck', match: ['innsbruck'] },
    { key: 'Richmond', match: ['richmond'] },
    { key: 'France', match: ['france'] },
    { key: 'Makuri Islands', match: ['makuri'] },
    { key: 'Scotland', match: ['scotland'] },
    { key: 'Yumezi', match: ['yumezi'] },
    { key: 'Crit City', match: ['crit city', 'crit'] },
  ]
  for (const w of worlds) {
    if (w.match.some((m) => hay.includes(m))) return w.key
  }
  return null
}

function ActivityDetailSkeleton() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-4">
        <Skeleton className="h-4 w-32" />
      </div>
      <div className="mb-6 flex items-start gap-4">
        <Skeleton className="h-12 w-12 rounded-lg" />
        <div className="space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-48" />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-24 rounded-lg" />
        ))}
      </div>
    </div>
  )
}
