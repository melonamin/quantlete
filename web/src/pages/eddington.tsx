import { useMemo, useState } from 'react'
import {
  useAppSettings,
  useAuthStatus,
  useEddington,
  useEddingtonHistory,
  useEddingtonCompare,
  useSportGroups,
  type EddingtonResult,
  type EddingtonCompareItem,
} from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Skeleton } from '@/components/ui/skeleton'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { LineChart } from '@/components/charts'

const DEFAULT_DEFS = [
  { id: 'all', name: 'All activities', sport_types: undefined },
  {
    id: 'rides',
    name: 'Rides',
    sport_types: ['Ride', 'MountainBikeRide', 'GravelRide', 'EBikeRide', 'VirtualRide'],
  },
  { id: 'runs', name: 'Runs', sport_types: ['Run', 'TrailRun', 'VirtualRun'] },
]

type ViewMode = 'all' | 'sport-group' | 'custom'

export function EddingtonPage() {
  const { data: auth } = useAuthStatus()
  const isAuthenticated = auth?.authenticated
  const { data: settings } = useAppSettings({ enabled: !!isAuthenticated })
  const { data: sportGroups } = useSportGroups()
  const { data: compareData, isLoading: compareLoading } = useEddingtonCompare()

  const defs = (
    settings?.eddington_definitions && settings.eddington_definitions.length
      ? settings.eddington_definitions
      : DEFAULT_DEFS
  ) as {
    id: string
    name: string
    sport_types?: string[]
  }[]

  const [viewMode, setViewMode] = useState<ViewMode>('all')
  const [selectedSportGroup, setSelectedSportGroup] = useState<string>('')
  const [defId, setDefId] = useState(defs[0]?.id ?? 'all')

  // Derive effective defId for custom definitions
  const effectiveDefId = useMemo(() => {
    if (!defs.length) return 'all'
    if (defs.some((d) => d.id === defId)) return defId
    return defs[0].id
  }, [defs, defId])

  const def = defs.find((d) => d.id === effectiveDefId) ?? defs[0]

  // Determine sport type and sport group based on view mode
  const sportType =
    viewMode === 'custom' && def?.sport_types?.length ? def.sport_types.join(',') : undefined
  const sportGroup = viewMode === 'sport-group' ? selectedSportGroup : undefined

  const { data, isLoading, error } = useEddington(sportType, sportGroup)
  const { data: history, isLoading: historyLoading } = useEddingtonHistory(sportType, sportGroup)

  // Select first sport group with data when sport groups load
  const availableSportGroups = useMemo(() => {
    if (!compareData?.groups) return []
    return compareData.groups
  }, [compareData])

  // Auto-select first sport group when switching to sport-group mode
  const handleViewModeChange = (mode: string) => {
    setViewMode(mode as ViewMode)
    if (mode === 'sport-group' && availableSportGroups.length > 0 && !selectedSportGroup) {
      setSelectedSportGroup(availableSportGroups[0].sport_group)
    }
  }

  const historySeries = useMemo(() => {
    const pts = (history ?? []).map((p) => ({ x: p.date, y: p.number }))
    return [
      {
        name: 'Eddington',
        data: pts,
        smooth: false,
        step: 'end' as const,
      },
    ]
  }, [history])

  // Get display title based on view mode
  const displayTitle = useMemo(() => {
    if (viewMode === 'all') return 'All Activities'
    if (viewMode === 'sport-group') {
      const group = sportGroups?.find((g) => g.id === selectedSportGroup)
      return group?.name ?? 'Sport Group'
    }
    return def?.name ?? 'Custom'
  }, [viewMode, selectedSportGroup, sportGroups, def])

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl font-bold">Eddington Number</h1>
            <p className="text-muted-foreground">Track your Eddington number progress</p>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row sm:items-center gap-4">
          <Tabs value={viewMode} onValueChange={handleViewModeChange} className="w-full sm:w-auto">
            <TabsList>
              <TabsTrigger value="all">All</TabsTrigger>
              <TabsTrigger value="sport-group">By Sport</TabsTrigger>
              <TabsTrigger value="custom">Custom</TabsTrigger>
            </TabsList>
          </Tabs>

          {viewMode === 'sport-group' && availableSportGroups.length > 0 && (
            <Select value={selectedSportGroup} onValueChange={setSelectedSportGroup}>
              <SelectTrigger className="w-[180px]">
                <SelectValue placeholder="Select sport" />
              </SelectTrigger>
              <SelectContent>
                {availableSportGroups.map((g) => (
                  <SelectItem key={g.sport_group} value={g.sport_group}>
                    {g.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}

          {viewMode === 'custom' && (
            <Select value={effectiveDefId} onValueChange={setDefId}>
              <SelectTrigger className="w-[180px]">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {defs.map((d) => (
                  <SelectItem key={d.id} value={d.id}>
                    {d.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          )}
        </div>
      </div>

      {/* Sport Comparison Table - always visible */}
      {viewMode !== 'custom' && (
        <SportComparisonCard
          data={compareData?.groups ?? []}
          allNumber={compareData?.all_number ?? 0}
          isLoading={compareLoading}
          selectedSportGroup={viewMode === 'sport-group' ? selectedSportGroup : undefined}
          onSelectSportGroup={(id) => {
            setViewMode('sport-group')
            setSelectedSportGroup(id)
          }}
        />
      )}

      {isLoading ? (
        <EddingtonSkeleton />
      ) : error ? (
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">Failed to load Eddington data</p>
        </div>
      ) : data ? (
        <EddingtonDisplay
          data={data}
          title={displayTitle}
          historySeries={historySeries}
          historyLoading={historyLoading}
        />
      ) : null}
    </div>
  )
}

function SportComparisonCard({
  data,
  allNumber,
  isLoading,
  selectedSportGroup,
  onSelectSportGroup,
}: {
  data: EddingtonCompareItem[]
  allNumber: number
  isLoading: boolean
  selectedSportGroup?: string
  onSelectSportGroup: (id: string) => void
}) {
  if (isLoading) {
    return (
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Sport Comparison</CardTitle>
          <CardDescription>Eddington numbers across sport groups</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-20" />
            ))}
          </div>
        </CardContent>
      </Card>
    )
  }

  if (data.length === 0) {
    return null
  }

  return (
    <Card className="mb-6">
      <CardHeader>
        <CardTitle>Sport Comparison</CardTitle>
        <CardDescription>
          Eddington numbers across sport groups (click to view details)
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-6 gap-3">
          <button
            type="button"
            onClick={() => onSelectSportGroup('')}
            className={`p-4 rounded-lg border text-center transition-colors hover:bg-accent ${
              selectedSportGroup === '' ? 'border-primary bg-primary/5' : 'border-border'
            }`}
          >
            <div className="text-2xl font-bold tabular-nums">{allNumber}</div>
            <div className="text-sm text-muted-foreground">All</div>
          </button>
          {data.map((item) => (
            <button
              key={item.sport_group}
              type="button"
              onClick={() => onSelectSportGroup(item.sport_group)}
              className={`p-4 rounded-lg border text-center transition-colors hover:bg-accent ${
                selectedSportGroup === item.sport_group
                  ? 'border-primary bg-primary/5'
                  : 'border-border'
              }`}
            >
              <div className="text-2xl font-bold tabular-nums">{item.number}</div>
              <div className="text-sm text-muted-foreground truncate">{item.name}</div>
            </button>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

function EddingtonDisplay({
  data,
  title,
  historySeries,
  historyLoading,
}: {
  data: EddingtonResult
  title: string
  historySeries: { name: string; data: { x: string; y: number }[]; smooth: boolean; step: 'end' }[]
  historyLoading: boolean
}) {
  return (
    <div className="space-y-6">
      {/* Main number */}
      <Card>
        <CardContent className="pt-6">
          <div className="text-center">
            <div className="text-7xl font-bold text-strava mb-2">{data.number}</div>
            <p className="text-muted-foreground">
              {title}: at least {data.number} km on {data.number} different days
            </p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>History</CardTitle>
        </CardHeader>
        <CardContent>
          <LineChart
            series={historySeries}
            xAxisType="time"
            yAxisLabel="Eddington"
            height={260}
            loading={historyLoading}
            showLegend={false}
            showDataZoom
          />
        </CardContent>
      </Card>

      {/* Next steps */}
      {data.next_steps.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Next Goals</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Target</TableHead>
                  <TableHead>Rides Needed</TableHead>
                  <TableHead>Status</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.next_steps.map((step) => (
                  <TableRow key={step.target}>
                    <TableCell className="font-medium">E{step.target}</TableCell>
                    <TableCell>
                      {step.rides_needed} {step.rides_needed === 1 ? 'day' : 'days'} of{' '}
                      {step.target}+ km
                    </TableCell>
                    <TableCell>
                      <Badge variant={step.rides_needed === 1 ? 'default' : 'outline'}>
                        {step.rides_needed === 1 ? 'Almost there!' : 'In progress'}
                      </Badge>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      {/* Top days */}
      {data.distribution.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle>Top Distance Days</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Rank</TableHead>
                  <TableHead>Date</TableHead>
                  <TableHead className="text-right">Distance</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data.distribution.slice(0, 20).map((day, index) => (
                  <TableRow key={day.date}>
                    <TableCell className="font-medium">#{index + 1}</TableCell>
                    <TableCell>{day.date}</TableCell>
                    <TableCell className="text-right">{day.distance.toFixed(1)} km</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            {data.distribution.length > 20 && (
              <p className="text-sm text-muted-foreground text-center mt-4">
                Showing top 20 of {data.distribution.length} days
              </p>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  )
}

function EddingtonSkeleton() {
  return (
    <div className="space-y-6">
      <Card>
        <CardContent className="pt-6">
          <div className="text-center">
            <Skeleton className="h-20 w-32 mx-auto mb-2" />
            <Skeleton className="h-4 w-64 mx-auto" />
          </div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-24" />
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <Skeleton key={i} className="h-10 w-full" />
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
