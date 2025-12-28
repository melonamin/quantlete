import { useEffect, useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { usePhotos, type PhotoListItem } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import { formatDate } from '@/lib/format'

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
  const perPage = 60

  const filtersKey = useMemo(
    () => JSON.stringify({ sportTypes: [...sportTypes].sort(), country }),
    [sportTypes, country]
  )
  useEffect(() => {
    setPage(1)
  }, [filtersKey])

  const sportParam = sportTypes.length > 0 ? sportTypes.join(',') : undefined
  const { data, isLoading, error } = usePhotos({
    sport_type: sportParam,
    country: country || undefined,
    page,
    per_page: perPage,
  })

  const [photos, setPhotos] = useState<PhotoListItem[]>([])
  useEffect(() => {
    if (!data) return
    if (page === 1) {
      setPhotos(data.data)
      return
    }
    setPhotos((prev) => {
      const seen = new Set(prev.map((p) => p.id))
      const next = [...prev]
      for (const p of data.data) {
        if (!seen.has(p.id)) next.push(p)
      }
      return next
    })
  }, [data, page])

  useEffect(() => {
    setPhotos([])
  }, [filtersKey])

  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null)
  const activePhoto = lightboxIndex != null ? photos[lightboxIndex] : null

  const sportFacet = data?.sport_types ?? []
  const countryFacet = data?.countries ?? []

  const toggleSport = (t: string) => {
    setSportTypes((prev) => (prev.includes(t) ? prev.filter((x) => x !== t) : [...prev, t]))
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Photos</h1>
          <p className="text-muted-foreground">Your activity photo gallery</p>
        </div>
        <div className="text-sm text-muted-foreground">
          {data ? `${data.total} photo${data.total === 1 ? '' : 's'}` : null}
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[18rem_1fr]">
        <div className="space-y-4">
          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-2 text-sm font-medium">Sport type</div>
            {sportFacet.length === 0 ? (
              <div className="text-sm text-muted-foreground">No photos yet.</div>
            ) : (
              <div className="space-y-2">
                {sportFacet.map((s) => (
                  <label key={s.value} className="flex items-center justify-between gap-2 text-sm">
                    <span className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={sportTypes.includes(s.value)}
                        onChange={() => toggleSport(s.value)}
                      />
                      {s.value}
                    </span>
                    <span className="text-xs text-muted-foreground">{s.count}</span>
                  </label>
                ))}
              </div>
            )}
          </div>

          <div className="rounded-lg border border-border bg-card p-4">
            <div className="mb-2 text-sm font-medium">Country</div>
            <div className="space-y-2">
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="radio"
                  name="photo-country"
                  checked={country === ''}
                  onChange={() => setCountry('')}
                />
                All
              </label>
              {countryFacet.map((c) => (
                <label key={c.value} className="flex items-center justify-between gap-2 text-sm">
                  <span className="flex min-w-0 items-center gap-2">
                    <input
                      type="radio"
                      name="photo-country"
                      checked={country === c.value}
                      onChange={() => setCountry(c.value)}
                    />
                    <span className="w-5 text-center">{flagEmoji(c.iso2)}</span>
                    <span className="truncate">{c.value}</span>
                  </span>
                  <span className="text-xs text-muted-foreground">{c.count}</span>
                </label>
              ))}
            </div>
          </div>

          <Button
            variant="outline"
            onClick={() => {
              setSportTypes([])
              setCountry('')
            }}
          >
            Clear filters
          </Button>
        </div>

        <div>
          {error && (
            <div className="mb-4 rounded-lg border border-destructive bg-destructive/10 p-4">
              <p className="text-destructive">Failed to load photos: {error.message}</p>
            </div>
          )}

          {isLoading && photos.length === 0 ? (
            <div className="columns-2 gap-3 md:columns-3 lg:columns-4">
              {Array.from({ length: 16 }).map((_, i) => (
                <div key={i} className="mb-3 break-inside-avoid">
                  <Skeleton className="h-40 w-full" />
                </div>
              ))}
            </div>
          ) : photos.length === 0 ? (
            <div className="rounded-lg border border-border bg-card p-8 text-center">
              <p className="text-muted-foreground">No photos found.</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Import with “Include photos” enabled to populate this page.
              </p>
            </div>
          ) : (
            <>
              <div className="columns-2 gap-3 md:columns-3 lg:columns-4">
                {photos.map((p, idx) => (
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
                        <div className="truncate text-xs font-medium text-white">
                          {p.activity_name}
                        </div>
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
                  <Button
                    variant="outline"
                    onClick={() => setPage((p) => p + 1)}
                    disabled={isLoading}
                  >
                    Load more
                  </Button>
                </div>
              )}
            </>
          )}
        </div>
      </div>

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
                  photos.length <= 1 && 'hidden'
                )}
                onClick={() =>
                  setLightboxIndex((i) =>
                    i == null ? null : (i - 1 + photos.length) % photos.length
                  )
                }
              >
                Prev
              </button>
              <button
                className={cn(
                  'absolute right-2 top-1/2 -translate-y-1/2 rounded-md bg-black/50 px-3 py-2 text-white hover:bg-black/70',
                  photos.length <= 1 && 'hidden'
                )}
                onClick={() =>
                  setLightboxIndex((i) => (i == null ? null : (i + 1) % photos.length))
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
