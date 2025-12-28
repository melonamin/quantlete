import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useSegmentCountries, useSegmentDetail, useSegments, type SegmentListItem } from '@/lib/api'
import { formatDate, formatDistance, formatDuration } from '@/lib/format'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { SegmentMap } from '@/components/maps'
import { SegmentPRChart } from '@/components/charts/segment-pr-chart'

type SortKey =
  | 'name'
  | 'distance'
  | 'maximum_grade'
  | 'times_completed'
  | 'last_effort_date'
  | 'best_elapsed_time'

function flagEmoji(iso2?: string) {
  if (!iso2 || iso2.length !== 2) return ''
  const codePoints = iso2
    .toUpperCase()
    .split('')
    .map((c) => 127397 + c.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

function sortValueForSegment(seg: SegmentListItem, key: SortKey) {
  switch (key) {
    case 'name':
      return seg.name.toLowerCase()
    case 'distance':
      return seg.distance ?? 0
    case 'maximum_grade':
      return seg.maximum_grade ?? 0
    case 'times_completed':
      return seg.times_completed ?? 0
    case 'last_effort_date':
      return seg.last_effort_date ? new Date(seg.last_effort_date).getTime() : 0
    case 'best_elapsed_time':
      return seg.best_elapsed_time ?? Number.POSITIVE_INFINITY
  }
}

export function SegmentsPage() {
  const [filtersOpen, setFiltersOpen] = useState(false)
  const [activityType, setActivityType] = useState('')
  const [country, setCountry] = useState('')
  const [starredOnly, setStarredOnly] = useState(false)
  const [komOnly, setKomOnly] = useState(false)
  const [search, setSearch] = useState('')

  const [sortKey, setSortKey] = useState<SortKey>('times_completed')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')

  const [selectedSegmentId, setSelectedSegmentId] = useState<number | null>(null)

  const apiFilters = useMemo(() => {
    return {
      activity_type: activityType || undefined,
      country: country || undefined,
      starred: starredOnly ? true : undefined,
      kom_only: komOnly ? true : undefined,
      search: search || undefined,
      limit: 5000,
    }
  }, [activityType, country, starredOnly, komOnly, search])

  const { data: segments, isLoading, error } = useSegments(apiFilters)
  const { data: countries } = useSegmentCountries(true)

  const sportOptions = useMemo(() => {
    const types = new Set<string>()
    for (const s of segments ?? []) {
      if (s.activity_type) types.add(s.activity_type)
    }
    const base = ['', 'Ride', 'Run']
    for (const t of types) {
      if (!base.includes(t)) base.push(t)
    }
    return base
  }, [segments])

  const sorted = useMemo(() => {
    const list = [...(segments ?? [])]
    list.sort((a, b) => {
      const av = sortValueForSegment(a, sortKey)
      const bv = sortValueForSegment(b, sortKey)
      if (av < bv) return sortDir === 'asc' ? -1 : 1
      if (av > bv) return sortDir === 'asc' ? 1 : -1
      return a.id - b.id
    })
    return list
  }, [segments, sortKey, sortDir])

  const detailQuery = useSegmentDetail(selectedSegmentId ?? 0, selectedSegmentId != null)

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc')
      return
    }
    setSortKey(key)
    setSortDir(key === 'name' ? 'asc' : 'desc')
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Segments</h1>
          <p className="text-muted-foreground">Your segment efforts and PRs</p>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="md:hidden"
          onClick={() => setFiltersOpen(!filtersOpen)}
        >
          Filters
        </Button>
      </div>

      <div className="flex gap-6">
        <div className="hidden w-80 shrink-0 md:block">
          <FiltersPanel
            sportOptions={sportOptions}
            activityType={activityType}
            setActivityType={setActivityType}
            countries={countries ?? []}
            country={country}
            setCountry={setCountry}
            starredOnly={starredOnly}
            setStarredOnly={setStarredOnly}
            komOnly={komOnly}
            setKomOnly={setKomOnly}
            search={search}
            setSearch={setSearch}
            onClear={() => {
              setActivityType('')
              setCountry('')
              setStarredOnly(false)
              setKomOnly(false)
              setSearch('')
            }}
          />
        </div>

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
                sportOptions={sportOptions}
                activityType={activityType}
                setActivityType={setActivityType}
                countries={countries ?? []}
                country={country}
                setCountry={setCountry}
                starredOnly={starredOnly}
                setStarredOnly={setStarredOnly}
                komOnly={komOnly}
                setKomOnly={setKomOnly}
                search={search}
                setSearch={setSearch}
                onClear={() => {
                  setActivityType('')
                  setCountry('')
                  setStarredOnly(false)
                  setKomOnly(false)
                  setSearch('')
                }}
              />
            </div>
          </div>
        )}

        <div className="min-w-0 flex-1">
          {error && (
            <div className="mb-4 rounded-lg border border-destructive bg-destructive/10 p-4">
              <p className="text-destructive">Failed to load segments: {error.message}</p>
            </div>
          )}

          {isLoading ? (
            <SegmentsSkeleton />
          ) : sorted.length === 0 ? (
            <div className="rounded-lg border border-border bg-card p-8 text-center">
              <p className="text-muted-foreground">No segments found.</p>
              <p className="mt-1 text-sm text-muted-foreground">
                Import with “Include segments” enabled to populate this page.
              </p>
              <p className="mt-3 text-sm">
                <Link to="/settings" className="text-primary hover:underline">
                  Go to Settings
                </Link>
              </p>
            </div>
          ) : (
            <div className="rounded-lg border border-border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <SortableHead
                      label="Segment"
                      active={sortKey === 'name'}
                      dir={sortDir}
                      onClick={() => toggleSort('name')}
                    />
                    <SortableHead
                      label="Distance"
                      align="right"
                      active={sortKey === 'distance'}
                      dir={sortDir}
                      onClick={() => toggleSort('distance')}
                    />
                    <SortableHead
                      label="Max grade"
                      align="right"
                      active={sortKey === 'maximum_grade'}
                      dir={sortDir}
                      onClick={() => toggleSort('maximum_grade')}
                    />
                    <SortableHead
                      label="Times"
                      align="right"
                      active={sortKey === 'times_completed'}
                      dir={sortDir}
                      onClick={() => toggleSort('times_completed')}
                    />
                    <SortableHead
                      label="Last effort"
                      active={sortKey === 'last_effort_date'}
                      dir={sortDir}
                      onClick={() => toggleSort('last_effort_date')}
                    />
                    <SortableHead
                      label="Best"
                      align="right"
                      active={sortKey === 'best_elapsed_time'}
                      dir={sortDir}
                      onClick={() => toggleSort('best_elapsed_time')}
                    />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {sorted.map((seg) => (
                    <TableRow
                      key={seg.id}
                      className="cursor-pointer hover:bg-accent/30"
                      onClick={() => setSelectedSegmentId(seg.id)}
                    >
                      <TableHead>
                        <div className="flex items-center gap-2">
                          <div className="min-w-0">
                            <div className="truncate font-medium">{seg.name}</div>
                            <div className="text-xs text-muted-foreground">
                              {seg.activity_type}
                              {seg.starred ? ' • ★' : ''}
                              {seg.athlete_kom_rank === 1 ? ' • KOM' : ''}
                            </div>
                          </div>
                        </div>
                      </TableHead>
                      <TableHead className="text-right">{formatDistance(seg.distance)}</TableHead>
                      <TableHead className="text-right">
                        {(seg.maximum_grade ?? 0).toFixed(1)}%
                      </TableHead>
                      <TableHead className="text-right">{seg.times_completed}</TableHead>
                      <TableHead>
                        {seg.last_effort_date ? formatDate(seg.last_effort_date) : '–'}
                      </TableHead>
                      <TableHead className="text-right">
                        {seg.best_elapsed_time ? formatDuration(seg.best_elapsed_time) : '–'}
                      </TableHead>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </div>
      </div>

      {selectedSegmentId != null && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
          <div className="mx-auto mt-8 h-[calc(100vh-4rem)] w-[calc(100%-2rem)] max-w-5xl overflow-auto rounded-lg border border-border bg-background shadow-lg">
            <div className="sticky top-0 z-10 flex items-center justify-between gap-4 border-b border-border bg-background/95 px-4 py-3">
              <div className="min-w-0">
                <div className="truncate text-lg font-semibold">
                  {detailQuery.data?.segment?.name ?? 'Segment'}
                </div>
                <div className="text-sm text-muted-foreground">
                  {detailQuery.data?.segment?.activity_type}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <a
                  className="text-sm text-primary hover:underline"
                  href={`https://www.strava.com/segments/${selectedSegmentId}`}
                  target="_blank"
                  rel="noreferrer"
                >
                  View on Strava
                </a>
                <Button variant="outline" size="sm" onClick={() => setSelectedSegmentId(null)}>
                  Close
                </Button>
              </div>
            </div>

            <div className="grid gap-4 p-4 lg:grid-cols-2">
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">Route</CardTitle>
                </CardHeader>
                <CardContent>
                  {detailQuery.isLoading ? (
                    <Skeleton className="h-64 w-full" />
                  ) : detailQuery.data?.segment?.polyline ? (
                    <SegmentMap
                      polyline={detailQuery.data.segment.polyline}
                      className="h-64"
                      averageGrade={detailQuery.data.segment.average_grade}
                      maximumGrade={detailQuery.data.segment.maximum_grade}
                      climbCategory={detailQuery.data.segment.climb_category}
                    />
                  ) : (
                    <div className="flex h-64 items-center justify-center rounded-lg bg-muted">
                      <p className="text-muted-foreground">No route data available</p>
                    </div>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">Best time progression</CardTitle>
                </CardHeader>
                <CardContent>
                  {detailQuery.isLoading ? (
                    <Skeleton className="h-64 w-full" />
                  ) : detailQuery.data?.efforts && detailQuery.data.efforts.length > 0 ? (
                    <SegmentPRChart efforts={detailQuery.data.efforts} height={260} />
                  ) : (
                    <div className="flex h-64 items-center justify-center rounded-lg bg-muted">
                      <p className="text-muted-foreground">No efforts yet</p>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>

            <div className="p-4 pt-0">
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">Efforts</CardTitle>
                </CardHeader>
                <CardContent>
                  {detailQuery.isLoading ? (
                    <Skeleton className="h-40 w-full" />
                  ) : detailQuery.data?.efforts && detailQuery.data.efforts.length > 0 ? (
                    <div className="rounded-lg border border-border">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Date</TableHead>
                            <TableHead className="text-right">Elapsed</TableHead>
                            <TableHead className="text-right">Moving</TableHead>
                            <TableHead className="text-right">Avg watts</TableHead>
                            <TableHead className="text-right">Avg HR</TableHead>
                            <TableHead>PR</TableHead>
                            <TableHead>Activity</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {detailQuery.data.efforts.map((e) => (
                            <TableRow key={e.id}>
                              <TableHead>
                                {e.start_date_local ? formatDate(e.start_date_local) : '–'}
                              </TableHead>
                              <TableHead className="text-right">
                                {formatDuration(e.elapsed_time)}
                              </TableHead>
                              <TableHead className="text-right">
                                {formatDuration(e.moving_time)}
                              </TableHead>
                              <TableHead className="text-right">
                                {e.average_watts ? Math.round(e.average_watts) : '–'}
                              </TableHead>
                              <TableHead className="text-right">
                                {e.average_heartrate ? Math.round(e.average_heartrate) : '–'}
                              </TableHead>
                              <TableHead>{e.pr_rank ? `#${e.pr_rank}` : '–'}</TableHead>
                              <TableHead>
                                <Link
                                  to="/activities/$activityId"
                                  params={{ activityId: String(e.activity_id) }}
                                  className="text-primary hover:underline"
                                >
                                  View
                                </Link>
                              </TableHead>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  ) : (
                    <div className="rounded-lg border border-border bg-card p-6 text-center">
                      <p className="text-muted-foreground">No efforts found.</p>
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function SortableHead({
  label,
  onClick,
  active,
  dir,
  align = 'left',
}: {
  label: string
  onClick: () => void
  active: boolean
  dir: 'asc' | 'desc'
  align?: 'left' | 'right'
}) {
  return (
    <TableHead className={align === 'right' ? 'text-right' : ''}>
      <button className="inline-flex items-center gap-1 hover:underline" onClick={onClick}>
        {label}
        {active ? (
          <span className="text-xs text-muted-foreground">{dir === 'asc' ? '↑' : '↓'}</span>
        ) : null}
      </button>
    </TableHead>
  )
}

function FiltersPanel({
  sportOptions,
  activityType,
  setActivityType,
  countries,
  country,
  setCountry,
  starredOnly,
  setStarredOnly,
  komOnly,
  setKomOnly,
  search,
  setSearch,
  onClear,
}: {
  sportOptions: string[]
  activityType: string
  setActivityType: (v: string) => void
  countries: { country: string; iso2?: string; count: number }[]
  country: string
  setCountry: (v: string) => void
  starredOnly: boolean
  setStarredOnly: (v: boolean) => void
  komOnly: boolean
  setKomOnly: (v: boolean) => void
  search: string
  setSearch: (v: string) => void
  onClear: () => void
}) {
  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-border bg-card p-4">
        <div className="mb-1 text-sm text-muted-foreground">Search</div>
        <input
          className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Segment name"
        />
      </div>

      <div className="rounded-lg border border-border bg-card p-4">
        <div className="mb-2 text-sm text-muted-foreground">Sport type</div>
        <div className="flex flex-col gap-2">
          {sportOptions.map((t) => (
            <label key={t || 'all'} className="flex items-center gap-2 text-sm">
              <input
                type="radio"
                name="segments-sport"
                checked={activityType === t}
                onChange={() => setActivityType(t)}
              />
              {t || 'All'}
            </label>
          ))}
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-4">
        <div className="mb-2 text-sm text-muted-foreground">Country</div>
        <div className="flex flex-col gap-2">
          <label className="flex items-center justify-between gap-2 text-sm">
            <span className="flex items-center gap-2">
              <input
                type="radio"
                name="segments-country"
                checked={country === ''}
                onChange={() => setCountry('')}
              />
              All
            </span>
          </label>
          {countries.map((c) => (
            <label key={c.country} className="flex items-center justify-between gap-2 text-sm">
              <span className="flex items-center gap-2">
                <input
                  type="radio"
                  name="segments-country"
                  checked={country === c.country}
                  onChange={() => setCountry(c.country)}
                />
                <span className="w-5 text-center">{flagEmoji(c.iso2)}</span>
                <span className="truncate">{c.country}</span>
              </span>
              <span className="text-xs text-muted-foreground">{c.count}</span>
            </label>
          ))}
        </div>
      </div>

      <div className="rounded-lg border border-border bg-card p-4 space-y-2">
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={starredOnly}
            onChange={(e) => setStarredOnly(e.target.checked)}
          />
          Starred only
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input type="checkbox" checked={komOnly} onChange={(e) => setKomOnly(e.target.checked)} />
          KOM only
        </label>
      </div>

      <Button variant="outline" onClick={onClear}>
        Clear filters
      </Button>
    </div>
  )
}

function SegmentsSkeleton() {
  return (
    <div className="rounded-lg border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Segment</TableHead>
            <TableHead className="text-right">Distance</TableHead>
            <TableHead className="text-right">Max grade</TableHead>
            <TableHead className="text-right">Times</TableHead>
            <TableHead>Last effort</TableHead>
            <TableHead className="text-right">Best</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 10 }).map((_, i) => (
            <TableRow key={i}>
              <TableHead>
                <Skeleton className="h-4 w-64" />
              </TableHead>
              <TableHead className="text-right">
                <Skeleton className="ml-auto h-4 w-16" />
              </TableHead>
              <TableHead className="text-right">
                <Skeleton className="ml-auto h-4 w-12" />
              </TableHead>
              <TableHead className="text-right">
                <Skeleton className="ml-auto h-4 w-10" />
              </TableHead>
              <TableHead>
                <Skeleton className="h-4 w-20" />
              </TableHead>
              <TableHead className="text-right">
                <Skeleton className="ml-auto h-4 w-14" />
              </TableHead>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
