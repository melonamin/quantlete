import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { usePhotos, type PhotoListItem } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/format'
import { X, ChevronDown, Globe, Filter } from 'lucide-react'

function flagEmoji(iso2?: string) {
  if (!iso2 || iso2.length !== 2) return ''
  const codePoints = iso2
    .toUpperCase()
    .split('')
    .map((c) => 127397 + c.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

export function PhotosPage() {
  const [sportTypes, setSportTypes] = useState<string[]>([])
  const [country, setCountry] = useState('')
  const [page, setPage] = useState(1)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [showCountries, setShowCountries] = useState(false)
  const [photos, setPhotos] = useState<PhotoListItem[]>([])
  const perPage = 60

  const sportParam = sportTypes.length > 0 ? sportTypes.join(',') : undefined
  const { data, isLoading, error } = usePhotos({
    sport_type: sportParam,
    country: country || undefined,
    page,
    per_page: perPage,
  })

  // Accumulate photos from data for infinite scroll
  const displayPhotos = useMemo(() => {
    if (!data) return photos
    if (page === 1) return data.data
    // For subsequent pages, we need to merge with existing photos
    const seen = new Set(photos.map((p) => p.id))
    const newPhotos = data.data.filter((p) => !seen.has(p.id))
    return [...photos, ...newPhotos]
  }, [data, page, photos])

  // Update photos state when we get new data (for persistence across pages)
  // This is done as a side effect but in the event handler when loading more
  const handleLoadMore = () => {
    if (data) {
      const seen = new Set(photos.map((p) => p.id))
      const newPhotos = data.data.filter((p) => !seen.has(p.id))
      setPhotos([...photos, ...newPhotos])
    }
    setPage((p) => p + 1)
  }

  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null)
  const activePhoto = lightboxIndex != null ? displayPhotos[lightboxIndex] : null

  // Use Array.isArray for defensive check against unexpected data shapes
  const sportFacet = Array.isArray(data?.sport_types) ? data.sport_types : []
  const countryFacet = Array.isArray(data?.countries) ? data.countries : []

  // Filter change handlers that also reset pagination and photos
  const toggleSport = (t: string) => {
    setSportTypes((prev) => (prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]))
    setPage(1)
    setPhotos([])
  }

  const handleSetSportTypes = (types: string[]) => {
    setSportTypes(types)
    setPage(1)
    setPhotos([])
  }

  const handleSetCountry = (c: string) => {
    setCountry(c)
    setPage(1)
    setPhotos([])
  }

  const hasActiveFilters = sportTypes.length > 0 || country

  const clearFilters = () => {
    setSportTypes([])
    setCountry('')
    setPage(1)
    setPhotos([])
  }

  // Common sport types for quick filters
  const quickSports = ['Ride', 'Run', 'Walk', 'Hike']

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Photos</h1>
        <p className="text-muted-foreground">Your activity photo gallery</p>
      </div>

      {/* Filters Panel - Stacked Layout */}
      <div className="space-y-4 rounded-lg border border-border bg-card p-4 mb-6">
        {/* Quick filters row */}
        <div className="flex flex-wrap items-center gap-2">
          {/* All button */}
          <button
            onClick={() => handleSetSportTypes([])}
            className={cn(
              'px-3 py-1.5 rounded-md text-sm transition-colors',
              sportTypes.length === 0
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted text-muted-foreground hover:bg-muted/80'
            )}
          >
            All
          </button>

          {/* Quick sport buttons */}
          {quickSports.map((type) => (
            <button
              key={type}
              onClick={() => toggleSport(type)}
              className={cn(
                'px-3 py-1.5 rounded-md text-sm transition-colors',
                sportTypes.includes(type)
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              )}
            >
              {type}
            </button>
          ))}

          {/* Divider */}
          <div className="w-px h-6 bg-border mx-1" />

          {/* Countries dropdown */}
          {countryFacet.length > 0 && (
            <div className="relative">
              <button
                onClick={() => setShowCountries(!showCountries)}
                className={cn(
                  'px-3 py-1.5 rounded-md text-sm transition-colors flex items-center gap-1.5',
                  country
                    ? 'bg-primary text-primary-foreground'
                    : 'bg-muted text-muted-foreground hover:bg-muted/80'
                )}
              >
                <Globe className="h-3.5 w-3.5" />
                {country || `${countryFacet.length} countries`}
                <ChevronDown
                  className={cn('h-3.5 w-3.5 transition-transform', showCountries && 'rotate-180')}
                />
              </button>

              {/* Countries dropdown panel */}
              {showCountries && (
                <div className="absolute top-full left-0 mt-1 z-50 w-64 max-h-80 overflow-auto rounded-md border border-border bg-card shadow-lg">
                  <div className="p-1">
                    <button
                      onClick={() => {
                        handleSetCountry('')
                        setShowCountries(false)
                      }}
                      className={cn(
                        'w-full px-3 py-2 text-left text-sm rounded-md transition-colors flex items-center gap-2',
                        !country ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
                      )}
                    >
                      All countries
                    </button>
                    {countryFacet.map((c) => (
                      <button
                        key={c.value}
                        onClick={() => {
                          handleSetCountry(c.value)
                          setShowCountries(false)
                        }}
                        className={cn(
                          'w-full px-3 py-2 text-left text-sm rounded-md transition-colors flex items-center justify-between',
                          country === c.value ? 'bg-primary/10 text-primary' : 'hover:bg-muted'
                        )}
                      >
                        <span className="flex items-center gap-2">
                          <span>{flagEmoji(c.iso2)}</span>
                          <span className="truncate">{c.value}</span>
                        </span>
                        <span className="text-muted-foreground text-xs">{c.count}</span>
                      </button>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* More sports button */}
          {sportFacet.length > quickSports.length && (
            <button
              onClick={() => setShowAdvanced(!showAdvanced)}
              className={cn(
                'px-3 py-1.5 rounded-md text-sm transition-colors flex items-center gap-1.5',
                showAdvanced || sportTypes.some((t) => !quickSports.includes(t))
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-muted text-muted-foreground hover:bg-muted/80'
              )}
            >
              <Filter className="h-3.5 w-3.5" />
              More sports
              <ChevronDown
                className={cn('h-3.5 w-3.5 transition-transform', showAdvanced && 'rotate-180')}
              />
            </button>
          )}

          {/* Clear filters */}
          {hasActiveFilters && (
            <button
              onClick={clearFilters}
              className="px-3 py-1.5 rounded-md text-sm text-destructive hover:bg-destructive/10 transition-colors flex items-center gap-1"
            >
              <X className="h-3.5 w-3.5" />
              Clear
            </button>
          )}

          {/* Photo count */}
          {data && (
            <div className="ml-auto text-sm text-muted-foreground">
              {data.total} photo{data.total === 1 ? '' : 's'}
            </div>
          )}
        </div>

        {/* Advanced filters (all sport types) */}
        {showAdvanced && sportFacet.length > 0 && (
          <div className="pt-3 border-t border-border">
            <div className="text-xs text-muted-foreground mb-2">All sport types</div>
            <div className="flex flex-wrap gap-2">
              {sportFacet.map((s) => (
                <button
                  key={s.value}
                  onClick={() => toggleSport(s.value)}
                  className={cn(
                    'px-3 py-1.5 rounded-md text-sm transition-colors flex items-center gap-1.5',
                    sportTypes.includes(s.value)
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-muted-foreground hover:bg-muted/80'
                  )}
                >
                  {s.value}
                  <span
                    className={cn(
                      'text-xs',
                      sportTypes.includes(s.value)
                        ? 'text-primary-foreground/70'
                        : 'text-muted-foreground/70'
                    )}
                  >
                    ({s.count})
                  </span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Photo Gallery */}
      {error && (
        <div className="mb-4 rounded-lg border border-destructive bg-destructive/10 p-4">
          <p className="text-destructive">Failed to load photos: {error.message}</p>
        </div>
      )}

      {isLoading && displayPhotos.length === 0 ? (
        <div className="columns-2 gap-3 md:columns-3 lg:columns-4 xl:columns-5">
          {Array.from({ length: 20 }).map((_, i) => (
            <div key={i} className="mb-3 break-inside-avoid">
              <Skeleton className="h-40 w-full rounded-lg" />
            </div>
          ))}
        </div>
      ) : displayPhotos.length === 0 ? (
        <div className="rounded-lg border border-border bg-card p-8 text-center">
          <p className="text-muted-foreground">No photos found.</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Import with "Include photos" enabled to populate this page.
          </p>
        </div>
      ) : (
        <>
          <div className="columns-2 gap-3 md:columns-3 lg:columns-4 xl:columns-5">
            {displayPhotos.map((p, idx) => (
              <button
                key={p.id}
                className="mb-3 w-full break-inside-avoid overflow-hidden rounded-lg border border-border bg-card text-left"
                onClick={() => setLightboxIndex(idx)}
              >
                <div className="group relative">
                  <img
                    src={p.thumbnail_url || p.url}
                    alt={p.caption || p.activity_name}
                    loading="lazy"
                    className="w-full object-cover"
                  />
                  <div className="pointer-events-none absolute inset-0 bg-gradient-to-t from-black/70 via-transparent to-transparent opacity-0 transition-opacity group-hover:opacity-100" />
                  <div className="pointer-events-none absolute bottom-0 left-0 right-0 p-2 opacity-0 transition-opacity group-hover:opacity-100">
                    <div className="truncate text-xs font-medium text-white">{p.activity_name}</div>
                    <div className="text-[11px] text-white/80">
                      {formatDate(p.start_date_local)}
                    </div>
                  </div>
                </div>
              </button>
            ))}
          </div>

          {data && page < data.total_pages && (
            <div className="mt-6 flex justify-center">
              <Button variant="outline" onClick={handleLoadMore} disabled={isLoading}>
                Load more
              </Button>
            </div>
          )}
        </>
      )}

      {/* Click outside to close dropdowns */}
      {showCountries && (
        <div className="fixed inset-0 z-40" onClick={() => setShowCountries(false)} />
      )}

      {/* Lightbox */}
      {activePhoto && lightboxIndex != null && (
        <div className="fixed inset-0 z-50 bg-black/80">
          <div className="mx-auto flex h-full max-w-5xl flex-col p-4">
            <div className="flex items-center justify-between gap-2 pb-3 text-white">
              <div className="min-w-0">
                <div className="truncate font-medium">{activePhoto.activity_name}</div>
                <div className="text-xs text-white/70">
                  {formatDate(activePhoto.start_date_local)}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <Link
                  to="/activities/$activityId"
                  params={{ activityId: String(activePhoto.activity_id) }}
                  className="rounded-md border border-white/20 px-3 py-1 text-sm hover:bg-white/10"
                  onClick={() => setLightboxIndex(null)}
                >
                  View activity
                </Link>
                <button
                  className="rounded-md border border-white/20 px-3 py-1 text-sm hover:bg-white/10"
                  onClick={() => setLightboxIndex(null)}
                >
                  Close
                </button>
              </div>
            </div>

            <div className="relative flex-1 overflow-hidden rounded-lg bg-black">
              <img
                src={activePhoto.url}
                alt={activePhoto.caption || activePhoto.activity_name}
                className="h-full w-full object-contain"
              />

              <button
                className={cn(
                  'absolute left-2 top-1/2 -translate-y-1/2 rounded-md bg-black/50 px-3 py-2 text-white hover:bg-black/70',
                  displayPhotos.length <= 1 && 'hidden'
                )}
                onClick={() =>
                  setLightboxIndex((i) =>
                    i == null ? null : (i - 1 + displayPhotos.length) % displayPhotos.length
                  )
                }
              >
                Prev
              </button>
              <button
                className={cn(
                  'absolute right-2 top-1/2 -translate-y-1/2 rounded-md bg-black/50 px-3 py-2 text-white hover:bg-black/70',
                  displayPhotos.length <= 1 && 'hidden'
                )}
                onClick={() =>
                  setLightboxIndex((i) => (i == null ? null : (i + 1) % displayPhotos.length))
                }
              >
                Next
              </button>
            </div>

            {activePhoto.caption && (
              <div className="pt-3 text-sm text-white/80">{activePhoto.caption}</div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
