import { useState } from 'react'
import { useEddingtonData, type EddingtonResult } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Bike, Footprints, Activity } from 'lucide-react'

type SportFilter = 'all' | 'ride' | 'run'

const RIDE_TYPES = 'Ride,MountainBikeRide,GravelRide,EBikeRide,VirtualRide'
const RUN_TYPES = 'Run,TrailRun,VirtualRun'

export function EddingtonPage() {
  const [sportFilter, setSportFilter] = useState<SportFilter>('all')

  const sportType =
    sportFilter === 'ride'
      ? RIDE_TYPES
      : sportFilter === 'run'
        ? RUN_TYPES
        : undefined

  const { data, isLoading, error } = useEddingtonData(sportType)

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Eddington Number</h1>
          <p className="text-muted-foreground">
            Track your Eddington number progress
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant={sportFilter === 'all' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSportFilter('all')}
          >
            <Activity className="h-4 w-4 mr-1" />
            All
          </Button>
          <Button
            variant={sportFilter === 'ride' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSportFilter('ride')}
          >
            <Bike className="h-4 w-4 mr-1" />
            Rides
          </Button>
          <Button
            variant={sportFilter === 'run' ? 'default' : 'outline'}
            size="sm"
            onClick={() => setSportFilter('run')}
          >
            <Footprints className="h-4 w-4 mr-1" />
            Runs
          </Button>
        </div>
      </div>

      {isLoading ? (
        <EddingtonSkeleton />
      ) : error ? (
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">Failed to load Eddington data</p>
        </div>
      ) : data ? (
        <EddingtonDisplay data={data} />
      ) : null}
    </div>
  )
}

function EddingtonDisplay({ data }: { data: EddingtonResult }) {
  return (
    <div className="space-y-6">
      {/* Main number */}
      <Card>
        <CardContent className="pt-6">
          <div className="text-center">
            <div className="text-7xl font-bold text-strava mb-2">
              {data.number}
            </div>
            <p className="text-muted-foreground">
              You have ridden at least {data.number} km on {data.number}{' '}
              different days
            </p>
          </div>
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
                    <TableCell className="font-medium">
                      E{step.target}
                    </TableCell>
                    <TableCell>
                      {step.rides_needed}{' '}
                      {step.rides_needed === 1 ? 'ride' : 'rides'} of{' '}
                      {step.target}+ km
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={step.rides_needed === 1 ? 'default' : 'outline'}
                      >
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
                    <TableCell className="text-right">
                      {day.distance.toFixed(1)} km
                    </TableCell>
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
