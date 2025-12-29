import { useHeatmapData } from '@/lib/api'
import { Heatmap, getActivitiesBounds } from '@/components/maps'
import { Skeleton } from '@/components/ui/skeleton'
import { Link } from '@tanstack/react-router'
import { Button } from '@/components/ui/button'
import { useMemo, useState, useCallback, useRef } from 'react'
import type { Map as LeafletMap } from 'leaflet'

export function HeatmapPage() {
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [sportType, setSportType] = useState('')
  const [after, setAfter] = useState('')
  const [before, setBefore] = useState('')
  const [commute, setCommute] = useState<'all' | 'yes' | 'no'>('all')
  const [workoutType, setWorkoutType] = useState<string>('')
  const [selectedCountry, setSelectedCountry] = useState<string>('')
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

  const { data, isLoading, error } = useHeatmapData(apiFilters)

  const countryStats = data?.countries ?? []
  const uniqueCountries = countryStats.length
  const worldCoveragePct = uniqueCountries > 0 ? Math.round((uniqueCountries / 195) * 100) : 0

  // Handle map ready callback
  const handleMapReady = useCallback((map: LeafletMap) => {
    mapRef.current = map
  }, [])

  // Handle country selection with flyTo
  const handleCountrySelect = useCallback((country: string) => {
    setSelectedCountry(country)

    if (!mapRef.current || !data?.activities) return

    if (country === '') {
      // Reset to all activities
      const bounds = getActivitiesBounds(data.activities)
      if (bounds) {
        mapRef.current.flyToBounds(bounds, { padding: [20, 20], duration: 1 })
      }
      return
    }

    // This requires country info on activities - for now we'll use a simple approach
    // Filter activities that likely belong to the selected country based on start coordinates
    // A more accurate implementation would require geocoding data
    const bounds = getActivitiesBounds(data.activities)
    if (bounds) {
      mapRef.current.flyToBounds(bounds, { padding: [20, 20], duration: 1 })
    }
  }, [data?.activities])

  return (
    <div className="flex h-[calc(100vh-3.5rem)] flex-col">
      <div className="border-b border-border bg-background px-4 py-4">
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold">Heatmap</h1>
            <p className="text-muted-foreground">Visualize all your activities on a map</p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="md:hidden"
              onClick={() => setFiltersOpen(!filtersOpen)}
            >
              Filters
            </Button>
            {data && (
              <div className="text-sm text-muted-foreground">
                {data.total} {data.total === 1 ? 'route' : 'routes'}
              </div>
            )}
          </div>
        </div>
      </div>

      <div className="flex flex-1 overflow-hidden">
        {/* Sidebar */}
        <div className="hidden w-80 shrink-0 border-r border-border bg-background p-4 md:block">
          <FiltersPanel
            sportType={sportType}
            setSportType={setSportType}
            after={after}
            setAfter={setAfter}
            before={before}
            setBefore={setBefore}
            commute={commute}
            setCommute={setCommute}
            workoutType={workoutType}
            setWorkoutType={setWorkoutType}
            onClear={() => {
              setSportType('')
              setAfter('')
              setBefore('')
              setCommute('all')
              setWorkoutType('')
            }}
          />

          <CountryPanel
            countries={countryStats}
            worldCoveragePct={worldCoveragePct}
            selectedCountry={selectedCountry}
            onCountrySelect={handleCountrySelect}
          />
        </div>

        {/* Mobile filters overlay */}
        {filtersOpen && (
          <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm md:hidden">
            <div className="mx-auto mt-10 w-[calc(100%-2rem)] max-w-md rounded-lg border border-border bg-card p-4 shadow-lg">
              <div className="mb-3 flex items-center justify-between">
                <div className="font-semibold">Filters</div>
                <Button variant="ghost" size="sm" onClick={() => setFiltersOpen(false)}>
                  Close
                </Button>
              </div>
              <FiltersPanel
                sportType={sportType}
                setSportType={setSportType}
                after={after}
                setAfter={setAfter}
                before={before}
                setBefore={setBefore}
                commute={commute}
                setCommute={setCommute}
                workoutType={workoutType}
                setWorkoutType={setWorkoutType}
                onClear={() => {
                  setSportType('')
                  setAfter('')
                  setBefore('')
                  setCommute('all')
                  setWorkoutType('')
                }}
              />
              <div className="mt-4">
                <CountryPanel
                  countries={countryStats}
                  worldCoveragePct={worldCoveragePct}
                  selectedCountry={selectedCountry}
                  onCountrySelect={handleCountrySelect}
                />
              </div>
            </div>
          </div>
        )}

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
            <Heatmap
              activities={data.activities}
              className="h-full"
              onMapReady={handleMapReady}
            />
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
      </div>
    </div>
  )
}

function FiltersPanel({
  sportType,
  setSportType,
  after,
  setAfter,
  before,
  setBefore,
  commute,
  setCommute,
  workoutType,
  setWorkoutType,
  onClear,
}: {
  sportType: string
  setSportType: (v: string) => void
  after: string
  setAfter: (v: string) => void
  before: string
  setBefore: (v: string) => void
  commute: 'all' | 'yes' | 'no'
  setCommute: (v: 'all' | 'yes' | 'no') => void
  workoutType: string
  setWorkoutType: (v: string) => void
  onClear: () => void
}) {
  return (
    <div className="space-y-4">
      <div>
        <div className="mb-1 text-sm text-muted-foreground">Sport type</div>
        <div className="flex flex-wrap gap-2">
          {[
            { label: 'All', value: '' },
            { label: 'Ride', value: 'Ride' },
            { label: 'Run', value: 'Run' },
            { label: 'Walk', value: 'Walk' },
            { label: 'Swim', value: 'Swim' },
          ].map((o) => (
            <label key={o.label} className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                name="heatmap-sport"
                checked={sportType === o.value}
                onChange={() => setSportType(o.value)}
              />
              {o.label}
            </label>
          ))}
        </div>
        <div className="mt-2">
          <input
            className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
            value={sportType}
            onChange={(e) => setSportType(e.target.value)}
            placeholder="Or enter comma-separated types (e.g. Ride,VirtualRide)"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-2">
        <div>
          <div className="mb-1 text-sm text-muted-foreground">From</div>
          <input
            type="date"
            className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
            value={after}
            onChange={(e) => setAfter(e.target.value)}
          />
        </div>
        <div>
          <div className="mb-1 text-sm text-muted-foreground">To</div>
          <input
            type="date"
            className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
            value={before}
            onChange={(e) => setBefore(e.target.value)}
          />
        </div>
      </div>

      <div>
        <div className="mb-1 text-sm text-muted-foreground">Commute</div>
        <select
          className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
          value={commute}
          onChange={(e) => setCommute(e.target.value as 'all' | 'yes' | 'no')}
        >
          <option value="all">All</option>
          <option value="yes">Commute</option>
          <option value="no">Not commute</option>
        </select>
      </div>

      <div>
        <div className="mb-1 text-sm text-muted-foreground">Workout type</div>
        <input
          className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
          value={workoutType}
          onChange={(e) => setWorkoutType(e.target.value)}
          placeholder="e.g. 1"
        />
        <p className="mt-1 text-xs text-muted-foreground">
          Strava workout type integer (optional).
        </p>
      </div>

      <Button variant="outline" onClick={onClear}>
        Clear filters
      </Button>
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

function CountryPanel({
  countries,
  worldCoveragePct,
  selectedCountry,
  onCountrySelect,
}: {
  countries: { country: string; iso2?: string; count: number }[]
  worldCoveragePct: number
  selectedCountry: string
  onCountrySelect: (country: string) => void
}) {
  if (!countries.length) return null

  const top = countries.slice(0, 12)

  return (
    <div className="mt-6 space-y-3">
      <div className="rounded-lg border border-border bg-muted/30 p-3">
        <div className="flex items-center justify-between">
          <div>
            <div className="text-sm font-medium">Countries</div>
            <div className="text-xs text-muted-foreground">{countries.length} visited</div>
          </div>
          <div className="text-right">
            <div className="text-sm font-medium">{worldCoveragePct}%</div>
            <div className="text-xs text-muted-foreground">world coverage</div>
          </div>
        </div>
      </div>

      <div className="flex items-center justify-between">
        <div className="text-sm font-medium">Top countries</div>
        {selectedCountry && (
          <button
            onClick={() => onCountrySelect('')}
            className="text-xs text-terminal-green hover:underline"
          >
            Show all
          </button>
        )}
      </div>
      <div className="space-y-2">
        {top.map((c) => (
          <button
            key={c.country}
            onClick={() => onCountrySelect(c.country)}
            className={`flex w-full items-center justify-between rounded-md border px-3 py-2 text-sm transition-colors ${
              selectedCountry === c.country
                ? 'border-terminal-green bg-terminal-green/10 text-terminal-green'
                : 'border-border hover:border-terminal-green/50 hover:bg-muted/50'
            }`}
          >
            <span className="flex items-center gap-2">
              <span>{flagEmoji(c.iso2)}</span>
              <span className="truncate">{c.country}</span>
            </span>
            <span className="text-muted-foreground">{c.count}</span>
          </button>
        ))}
      </div>
    </div>
  )
}
