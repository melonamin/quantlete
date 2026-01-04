import { useMemo, useState, useEffect } from 'react'
import { Link } from '@tanstack/react-router'
import { useSegmentCountries, useSegmentDetail, useSegments } from '@/lib/api'
import type { SegmentsFilters } from '@/lib/api/segments'
import { formatDate, formatDistance, formatDuration } from '@/lib/format'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
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
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
  TableCell,
} from '@/components/ui/table'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { SegmentMap } from '@/components/maps'
import { SegmentPRChart } from '@/components/charts/segment-pr-chart'
import { Pagination } from '@/components/activities'
import { Search, X, Filter, Star, Crown } from 'lucide-react'
import { PAGINATION } from '@/lib/constants'

type SortKey = NonNullable<SegmentsFilters['order_by']>

function flagEmoji(iso2?: string) {
  if (!iso2 || iso2.length !== 2) return ''
  const codePoints = iso2
    .toUpperCase()
    .split('')
    .map((c) => 127397 + c.charCodeAt(0))
  return String.fromCodePoint(...codePoints)
}

export function SegmentsPage() {
  const [activityType, setActivityType] = useState('')
  const [country, setCountry] = useState('')
  const [starredOnly, setStarredOnly] = useState(false)
  const [komOnly, setKomOnly] = useState(false)
  const [localSearch, setLocalSearch] = useState('')
  const [search, setSearch] = useState('')
  const [showAdvanced, setShowAdvanced] = useState(false)

  const [sortKey, setSortKey] = useState<SortKey>('times_completed')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')

  const [selectedSegmentId, setSelectedSegmentId] = useState<number | null>(null)

  // Pagination state
  const [page, setPage] = useState(1)

  // Debounce search and reset page when search changes
  useEffect(() => {
    const timer = setTimeout(() => {
      if (search !== localSearch) {
        setSearch(localSearch)
        setPage(1)
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [localSearch, search])

  // Build API filters with pagination and sorting
  const apiFilters = useMemo((): SegmentsFilters => {
    return {
      activity_type: activityType || undefined,
      country: country || undefined,
      starred: starredOnly ? true : undefined,
      kom_only: komOnly ? true : undefined,
      search: search || undefined,
      page,
      per_page: PAGINATION.DEFAULT_PER_PAGE,
      order_by: sortKey,
      order_dir: sortDir,
    }
  }, [activityType, country, starredOnly, komOnly, search, page, sortKey, sortDir])

  const { data: segmentsResponse, isLoading, error } = useSegments(apiFilters)
  const { data: countries } = useSegmentCountries()

  // Extract data from response (use Array.isArray for defensive check)
  const segments = useMemo(
    () => (Array.isArray(segmentsResponse?.data) ? segmentsResponse.data : []),
    [segmentsResponse]
  )
  const total = segmentsResponse?.total ?? 0
  const totalPages = segmentsResponse?.total_pages ?? 1
  const currentPage = segmentsResponse?.page ?? 1
  const perPage = segmentsResponse?.per_page ?? PAGINATION.DEFAULT_PER_PAGE

  // Get sport options from first page for filter dropdown (or use a separate endpoint)
  const sportOptions = useMemo(() => {
    const types = new Set<string>()
    for (const s of segments) {
      if (s.activity_type) types.add(s.activity_type)
    }
    return Array.from(types).sort()
  }, [segments])

  const detailQuery = useSegmentDetail(selectedSegmentId, selectedSegmentId != null)

  // Filter change handlers that also reset pagination
  const handleActivityTypeChange = (value: string) => {
    setActivityType(value)
    setPage(1)
  }

  const handleCountryChange = (value: string) => {
    setCountry(value)
    setPage(1)
  }

  const handleStarredOnlyChange = (checked: boolean) => {
    setStarredOnly(checked)
    setPage(1)
  }

  const handleKomOnlyChange = (checked: boolean) => {
    setKomOnly(checked)
    setPage(1)
  }

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir(sortDir === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDir(key === 'name' ? 'asc' : 'desc')
    }
    setPage(1)
  }

  const hasActiveFilters = activityType || country || starredOnly || komOnly || search

  const clearFilters = () => {
    setActivityType('')
    setCountry('')
    setStarredOnly(false)
    setKomOnly(false)
    setLocalSearch('')
    setSearch('')
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Segments</h1>
        <p className="text-muted-foreground">Your segment efforts and PRs</p>
      </div>

      {/* Filters Panel - Stacked Layout */}
      <div className="space-y-4 rounded-lg border border-border bg-card p-4 mb-6">
        {/* Search input */}
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Search segments by name..."
            className="h-10 pl-10 pr-10"
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
          />
          {localSearch && (
            <Button
              variant="ghost"
              size="icon-sm"
              onClick={() => setLocalSearch('')}
              className="absolute right-1 top-1/2 -translate-y-1/2"
            >
              <X className="h-4 w-4" />
            </Button>
          )}
        </div>

        {/* Quick filters */}
        <div className="flex flex-wrap gap-2">
          {/* Sport type buttons */}
          <ToggleGroup
            type="single"
            value={activityType}
            onValueChange={handleActivityTypeChange}
            variant="outline"
            size="sm"
          >
            <ToggleGroupItem value="">All</ToggleGroupItem>
            <ToggleGroupItem value="Ride">Ride</ToggleGroupItem>
            <ToggleGroupItem value="Run">Run</ToggleGroupItem>
          </ToggleGroup>

          {/* Divider */}
          <div className="w-px bg-border mx-1" />

          {/* Toggle buttons */}
          <Button
            variant={starredOnly ? 'default' : 'outline'}
            size="sm"
            onClick={() => handleStarredOnlyChange(!starredOnly)}
            className="gap-1"
          >
            <Star className="h-3 w-3" />
            Starred
          </Button>
          <Button
            variant={komOnly ? 'default' : 'outline'}
            size="sm"
            onClick={() => handleKomOnlyChange(!komOnly)}
            className="gap-1"
          >
            <Crown className="h-3 w-3" />
            KOM
          </Button>

          {/* More filters button */}
          <Button
            variant={showAdvanced ? 'default' : 'outline'}
            size="sm"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="gap-1"
          >
            <Filter className="h-3 w-3" />
            More
          </Button>
        </div>

        {/* Advanced filters */}
        {showAdvanced && (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2 border-t border-border">
            {/* Sport type dropdown */}
            <div>
              <Label className="text-xs text-muted-foreground mb-1">Sport Type</Label>
              <Select value={activityType} onValueChange={handleActivityTypeChange}>
                <SelectTrigger size="sm">
                  <SelectValue placeholder="All types" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All types</SelectItem>
                  {sportOptions.map((type) => (
                    <SelectItem key={type} value={type}>
                      {type}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Country dropdown */}
            <div>
              <Label className="text-xs text-muted-foreground mb-1">Country</Label>
              <Select value={country} onValueChange={handleCountryChange}>
                <SelectTrigger size="sm">
                  <SelectValue placeholder="All countries" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="">All countries</SelectItem>
                  {(countries ?? []).map((c) => (
                    <SelectItem key={c.country} value={c.country}>
                      {flagEmoji(c.iso2)} {c.country} ({c.count})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Toggles */}
            <div>
              <Label className="text-xs text-muted-foreground mb-1">Options</Label>
              <div className="flex flex-wrap gap-3 h-9 items-center">
                <Label className="flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={starredOnly}
                    onCheckedChange={(checked) => handleStarredOnlyChange(!!checked)}
                  />
                  Starred only
                </Label>
                <Label className="flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={komOnly}
                    onCheckedChange={(checked) => handleKomOnlyChange(!!checked)}
                  />
                  KOM only
                </Label>
              </div>
            </div>
          </div>
        )}

        {/* Clear filters */}
        {hasActiveFilters && (
          <div className="flex justify-end">
            <Button variant="ghost" size="sm" onClick={clearFilters}>
              <X className="h-4 w-4 mr-1" />
              Clear filters
            </Button>
          </div>
        )}
      </div>

      {/* Error state */}
      {error && (
        <div className="mb-4 rounded-lg border border-destructive bg-destructive/10 p-4">
          <p className="text-destructive">Failed to load segments: {error.message}</p>
        </div>
      )}

      {/* Table */}
      {isLoading ? (
        <SegmentsSkeleton />
      ) : total === 0 ? (
        <div className="rounded-lg border border-border bg-card p-8 text-center">
          <p className="text-muted-foreground">No segments found.</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Import with "Include segments" enabled to populate this page.
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
              {segments.map((seg) => (
                <TableRow
                  key={seg.id}
                  className="cursor-pointer hover:bg-accent/30"
                  onClick={() => setSelectedSegmentId(seg.id)}
                >
                  <TableCell>
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
                  </TableCell>
                  <TableCell className="text-right">{formatDistance(seg.distance)}</TableCell>
                  <TableCell className="text-right">
                    {(seg.maximum_grade ?? 0).toFixed(1)}%
                  </TableCell>
                  <TableCell className="text-right">{seg.times_completed}</TableCell>
                  <TableCell>
                    {seg.last_effort_date ? formatDate(seg.last_effort_date) : '–'}
                  </TableCell>
                  <TableCell className="text-right">
                    {seg.best_elapsed_time ? formatDuration(seg.best_elapsed_time) : '–'}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Pagination */}
      {!isLoading && total > 0 && (
        <Pagination
          page={currentPage}
          totalPages={totalPages}
          total={total}
          perPage={perPage}
          onPageChange={setPage}
          itemLabel="segments"
        />
      )}

      {/* Segment Detail Modal */}
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
                              <TableCell>
                                {e.start_date_local ? formatDate(e.start_date_local) : '–'}
                              </TableCell>
                              <TableCell className="text-right">
                                {formatDuration(e.elapsed_time)}
                              </TableCell>
                              <TableCell className="text-right">
                                {formatDuration(e.moving_time)}
                              </TableCell>
                              <TableCell className="text-right">
                                {e.average_watts ? Math.round(e.average_watts) : '–'}
                              </TableCell>
                              <TableCell className="text-right">
                                {e.average_heartrate ? Math.round(e.average_heartrate) : '–'}
                              </TableCell>
                              <TableCell>{e.pr_rank ? `#${e.pr_rank}` : '–'}</TableCell>
                              <TableCell>
                                <Link
                                  to="/activities/$activityId"
                                  params={{ activityId: String(e.activity_id) }}
                                  className="text-primary hover:underline"
                                >
                                  View
                                </Link>
                              </TableCell>
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
      <Button
        variant="ghost"
        size="sm"
        className="h-auto p-0 font-medium uppercase tracking-wider text-[10px] hover:bg-transparent hover:underline"
        onClick={onClick}
      >
        {label}
        {active ? (
          <span className="ml-1 text-muted-foreground">{dir === 'asc' ? '↑' : '↓'}</span>
        ) : null}
      </Button>
    </TableHead>
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
              <TableCell>
                <Skeleton className="h-4 w-64" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="ml-auto h-4 w-16" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="ml-auto h-4 w-12" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="ml-auto h-4 w-10" />
              </TableCell>
              <TableCell>
                <Skeleton className="h-4 w-20" />
              </TableCell>
              <TableCell className="text-right">
                <Skeleton className="ml-auto h-4 w-14" />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
