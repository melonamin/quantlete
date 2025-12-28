import { useEffect, useMemo, useState } from 'react'
import {
  useAppSettings,
  useAuthStatus,
  useEddingtonData,
  useEddingtonHistory,
  type EddingtonResult,
} from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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

export function EddingtonPage() {
  const { data: auth } = useAuthStatus()
  const isAuthenticated = auth?.authenticated
  const { data: settings } = useAppSettings({ enabled: !!isAuthenticated })
  const defs = (
    settings?.eddington_definitions && settings.eddington_definitions.length
      ? settings.eddington_definitions
      : DEFAULT_DEFS
  ) as {
    id: string
    name: string
    sport_types?: string[]
  }[]

  const [defId, setDefId] = useState(defs[0]?.id ?? 'all')
  useEffect(() => {
    if (!defs.length) return
    if (defs.some((d) => d.id === defId)) return
    setDefId(defs[0].id)
  }, [defs, defId])

  const def = defs.find((d) => d.id === defId) ?? defs[0]
  const sportType = def?.sport_types?.length ? def.sport_types.join(',') : undefined

  const { data, isLoading, error } = useEddingtonData(sportType)
  const { data: history, isLoading: historyLoading } = useEddingtonHistory(sportType)

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

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Eddington Number</h1>
          <p className="text-muted-foreground">Track your Eddington number progress</p>
        </div>
        <div className="flex items-center gap-2">
          <select
            className="h-9 rounded-md border border-border bg-background px-2 text-sm"
            value={defId}
            onChange={(e) => setDefId(e.target.value)}
          >
            {defs.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </select>
        </div>
      </div>

      {isLoading ? (
        <EddingtonSkeleton />
      ) : error ? (
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">Failed to load Eddington data</p>
        </div>
      ) : data ? (
        <EddingtonDisplay
          data={data}
          title={def?.name ?? 'Eddington'}
          historySeries={historySeries}
          historyLoading={historyLoading}
        />
      ) : null}
    </div>
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
