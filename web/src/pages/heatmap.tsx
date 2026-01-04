import { useHeatmap, type HeatmapCountryStat } from '@/lib/api'
import { Heatmap, getActivitiesBounds } from '@/components/maps'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { Link } from '@tanstack/react-router'
import { useMemo, useState, useCallback, useRef } from 'react'
import type { Map as LeafletMap } from 'leaflet'
import { Filter, X, ChevronDown, Globe, MapPin } from 'lucide-react'
import { cn } from '@/lib/utils'

export function HeatmapPage() {
  const [sportType, setSportType] = useState('')
  const [after, setAfter] = useState('')
  const [before, setBefore] = useState('')
  const [commute, setCommute] = useState<'all' | 'yes' | 'no'>('all')
  const [workoutType, setWorkoutType] = useState<string>('')
  const [selectedCountry, setSelectedCountry] = useState<string>('')
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [showCountries, setShowCountries] = useState(false)
  const mapRef = useRef<LeafletMap | null>(null)

  const apiFilters = useMemo(() => {
    return {
      sport_type: sportType || undefined,
      after: after || undefined,
      before: before || undefined,
      commute: commute === 'all' ? undefined : commute === 'yes',
      workout_type: workoutType ? Number(workoutType) : undefined,
    }
  }, [sportType, after, before, commute, workoutType])

  const { data, isLoading, error } = useHeatmap(apiFilters)

  const countryStats = data?.countries ?? []
  const uniqueCountries = countryStats.length
  const worldCoveragePct = uniqueCountries > 0 ? Math.round((uniqueCountries / 195) * 100) : 0

  // Handle map ready callback
  const handleMapReady = useCallback((map: LeafletMap) => {
    mapRef.current = map
  }, [])

  // Handle country selection with flyTo
  const handleCountrySelect = useCallback(
    (country: string) => {
      setSelectedCountry(country)
      setShowCountries(false)

      if (!mapRef.current || !data?.activities) return

      if (country === '') {
        // Reset to all activities
        const bounds = getActivitiesBounds(data.activities)
        if (bounds) {
          mapRef.current.flyToBounds(bounds, { padding: [20, 20], duration: 1 })
        }
        return
      }

      const bounds = getActivitiesBounds(data.activities)
      if (bounds) {
        mapRef.current.flyToBounds(bounds, { padding: [20, 20], duration: 1 })
      }
    },
    [data]
  )

  const hasActiveFilters = sportType || after || before || commute !== 'all' || workoutType

  const clearFilters = () => {
    setSportType('')
    setAfter('')
    setBefore('')
    setCommute('all')
    setWorkoutType('')
    setSelectedCountry('')
  }

  return (
    <div className="flex h-[calc(100vh-3.5rem)] flex-col">
      {/* Filters Panel - Stacked Layout */}
      <div className="border-b border-border bg-background px-4 py-3 space-y-3">
        {/* Header row */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-xl font-bold">Heatmap</h1>
            <p className="text-sm text-muted-foreground">Visualize all your activities on a map</p>
          </div>
          {data && (
            <div className="text-sm text-muted-foreground">
              {data.total} {data.total === 1 ? 'route' : 'routes'}
            </div>
          )}
        </div>

        {/* Quick filters row */}
        <div className="flex flex-wrap items-center gap-2">
          {/* Sport type buttons */}
          <ToggleGroup
            type="single"
            value={sportType}
            onValueChange={(value) => setSportType(value)}
            variant="outline"
            size="sm"
          >
            <ToggleGroupItem value="">All</ToggleGroupItem>
            {['Ride', 'Run', 'Walk', 'Swim'].map((type) => (
              <ToggleGroupItem key={type} value={type}>
                {type}
              </ToggleGroupItem>
            ))}
          </ToggleGroup>

          {/* Divider */}
          <div className="w-px h-6 bg-border mx-1" />

          {/* Countries dropdown */}
          {countryStats.length > 0 && (
            <div className="relative">
              <Button
                variant={selectedCountry ? 'default' : 'outline'}
                size="sm"
                onClick={() => setShowCountries(!showCountries)}
                className="gap-1.5"
              >
                <Globe className="h-3.5 w-3.5" />
                {selectedCountry || `${uniqueCountries} countries`}
                <ChevronDown
                  className={cn('h-3.5 w-3.5 transition-transform', showCountries && 'rotate-180')}
                />
              </Button>

              {/* Countries dropdown panel */}
              {showCountries && (
                <div className="absolute top-full left-0 mt-1 z-50 w-64 max-h-80 overflow-auto rounded-md border border-border bg-card shadow-lg">
                  <div className="p-2 border-b border-border bg-muted/30">
                    <div className="flex items-center justify-between text-xs">
                      <span className="text-muted-foreground">{uniqueCountries} visited</span>
                      <span className="text-muted-foreground">{worldCoveragePct}% world</span>
                    </div>
                  </div>
                  <div className="p-1">
                    <button
                      onClick={() => handleCountrySelect('')}
                      className={cn(
                        'w-full px-3 py-2 text-left text-sm rounded-md transition-colors flex items-center gap-2',
                        !selectedCountry ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
                      )}
                    >
                      <MapPin className="h-3.5 w-3.5" />
                      All countries
                    </button>
                    {countryStats.slice(0, 20).map((c: HeatmapCountryStat) => (
                      <button
                        key={c.country}
                        onClick={() => handleCountrySelect(c.country)}
                        className={cn(
                          'w-full px-3 py-2 text-left text-sm rounded-md transition-colors flex items-center justify-between',
                          selectedCountry === c.country
                            ? 'bg-primary/10 text-primary'
                            : 'hover:bg-muted'
                        )}
                      >
                        <span className="flex items-center gap-2">
                          <span>{flagEmoji(c.iso2)}</span>
                          <span className="truncate">{c.country}</span>
                        </span>
                        <span className="text-muted-foreground text-xs">{c.count}</span>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* More filters button */}
          <Button
            variant={
              showAdvanced ||
              (hasActiveFilters && (after || before || commute !== 'all' || workoutType))
                ? 'default'
                : 'outline'
            }
            size="sm"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="gap-1.5"
          >
            <Filter className="h-3.5 w-3.5" />
            More
            <ChevronDown
              className={cn('h-3.5 w-3.5 transition-transform', showAdvanced && 'rotate-180')}
            />
          </Button>

          {/* Clear filters */}
          {hasActiveFilters && (
            <Button
              variant="ghost"
              size="sm"
              onClick={clearFilters}
              className="text-destructive hover:text-destructive hover:bg-destructive/10 gap-1"
            >
              <X className="h-3.5 w-3.5" />
              Clear
            </Button>
          )}
        </div>

        {/* Advanced filters (collapsible) */}
        {showAdvanced && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3 pt-2 border-t border-border">
            {/* Custom sport type */}
            <div>
              <Label className="text-xs text-muted-foreground mb-1">Custom sport</Label>
              <Input
                type="text"
                className="h-8"
                value={sportType}
                onChange={(e) => setSportType(e.target.value)}
                placeholder="e.g. Ride,VirtualRide"
              />
            </div>

            {/* Date from */}
            <div>
              <Label className="text-xs text-muted-foreground mb-1">From</Label>
              <Input
                type="date"
                className="h-8"
                value={after}
                onChange={(e) => setAfter(e.target.value)}
              />
            </div>

            {/* Date to */}
            <div>
              <Label className="text-xs text-muted-foreground mb-1">To</Label>
              <Input
                type="date"
                className="h-8"
                value={before}
                onChange={(e) => setBefore(e.target.value)}
              />
            </div>

            {/* Commute */}
            <div>
              <Label className="text-xs text-muted-foreground mb-1">Commute</Label>
              <Select
                value={commute}
                onValueChange={(value) => setCommute(value as 'all' | 'yes' | 'no')}
              >
                <SelectTrigger size="sm" className="h-8">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All</SelectItem>
                  <SelectItem value="yes">Commute only</SelectItem>
                  <SelectItem value="no">Non-commute</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        )}
      </div>

      {/* Map */}
      <div className="flex-1">
        {isLoading ? (
          <Skeleton className="h-full w-full" />
        ) : error ? (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <p className="mb-2 text-destructive">Failed to load heatmap data</p>
              <p className="text-sm text-muted-foreground">
                Make sure you are{' '}
                <Link to="/settings" className="underline">
                  connected to Strava
                </Link>
              </p>
            </div>
          </div>
        ) : data && data.activities.length > 0 ? (
          <Heatmap activities={data.activities} className="h-full" onMapReady={handleMapReady} />
        ) : (
          <div className="flex h-full items-center justify-center">
            <div className="text-center">
              <p className="mb-2 text-muted-foreground">No activities with GPS data found</p>
              <p className="text-sm text-muted-foreground">
                Import activities with GPS data to see them on the heatmap
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Click outside to close dropdowns */}
      {showCountries && (
        <div className="fixed inset-0 z-40" onClick={() => setShowCountries(false)} />
      )}
    </div>
  )
}

function flagEmoji(iso2?: string) {
  if (!iso2 || iso2.length !== 2) return ''
  const codePoints = iso2
    .toUpperCase()
    .split('')
    .map((c) => 127397 + c.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}
