import type { Activity } from '@/lib/api'
import { ActivityRow } from './activity-row'
import {
  Table,
  TableBody,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Skeleton } from '@/components/ui/skeleton'

interface ActivitiesTableProps {
  activities: Activity[]
  isLoading?: boolean
}

export function ActivitiesTable({ activities, isLoading }: ActivitiesTableProps) {
  if (isLoading) {
    return <ActivitiesTableSkeleton />
  }

  if (activities.length === 0) {
    return (
      <div className="rounded-lg border border-border bg-card p-8 text-center">
        <p className="text-muted-foreground">No activities found.</p>
        <p className="text-sm text-muted-foreground mt-1">
          Import activities from Strava to get started.
        </p>
      </div>
    )
  }

  return (
    <div className="rounded-lg border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[300px]">Activity</TableHead>
            <TableHead className="w-[140px]">Type</TableHead>
            <TableHead className="w-[120px]">Date</TableHead>
            <TableHead className="w-[100px] text-right">Distance</TableHead>
            <TableHead className="w-[100px] text-right">Time</TableHead>
            <TableHead className="w-[100px] text-right">Elevation</TableHead>
            <TableHead className="w-[120px]">Tags</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {activities.map((activity) => (
            <ActivityRow key={activity.id} activity={activity} />
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

function ActivitiesTableSkeleton() {
  return (
    <div className="rounded-lg border border-border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-[300px]">Activity</TableHead>
            <TableHead className="w-[140px]">Type</TableHead>
            <TableHead className="w-[120px]">Date</TableHead>
            <TableHead className="w-[100px] text-right">Distance</TableHead>
            <TableHead className="w-[100px] text-right">Time</TableHead>
            <TableHead className="w-[100px] text-right">Elevation</TableHead>
            <TableHead className="w-[120px]">Tags</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 10 }).map((_, i) => (
            <TableRow key={i}>
              <TableHead><Skeleton className="h-4 w-48" /></TableHead>
              <TableHead><Skeleton className="h-4 w-20" /></TableHead>
              <TableHead><Skeleton className="h-4 w-24" /></TableHead>
              <TableHead><Skeleton className="h-4 w-16 ml-auto" /></TableHead>
              <TableHead><Skeleton className="h-4 w-16 ml-auto" /></TableHead>
              <TableHead><Skeleton className="h-4 w-16 ml-auto" /></TableHead>
              <TableHead><Skeleton className="h-4 w-16" /></TableHead>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
