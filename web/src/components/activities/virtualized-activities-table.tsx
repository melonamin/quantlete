import { useRef, useCallback } from 'react'
import { useVirtualizer } from '@tanstack/react-virtual'
import { Link } from '@tanstack/react-router'
import type { Activity } from '@/lib/api'
import { formatDuration, formatDate } from '@/lib/format'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { SportIcon } from '@/lib/sport-icon'
import { getSportTextColor, formatSportType } from '@/lib/sport-types'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'

interface VirtualizedActivitiesTableProps {
  activities: Activity[]
  isLoading?: boolean
  height?: number | string
}

const ROW_HEIGHT = 52 // Height of each row in pixels
const OVERSCAN = 5 // Number of rows to render outside visible area

export function VirtualizedActivitiesTable({
  activities,
  isLoading,
  height = 600,
}: VirtualizedActivitiesTableProps) {
  const parentRef = useRef<HTMLDivElement>(null)
  const { formatDistance, formatElevation } = useFormattedMetrics()

  // eslint-disable-next-line react-hooks/incompatible-library -- TanStack Virtual returns unstable functions by design
  const virtualizer = useVirtualizer({
    count: activities.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => ROW_HEIGHT,
    overscan: OVERSCAN,
  })

  const virtualRows = virtualizer.getVirtualItems()

  const renderRow = useCallback(
    (activity: Activity) => {
      const sportColor = getSportTextColor(activity.sport_type)

      return (
        <>
          {/* Activity Name */}
          <div className="flex-[3] min-w-[200px] px-4 py-3 font-medium truncate">
            <Link
              to="/activities/$activityId"
              params={{ activityId: String(activity.id) }}
              className="hover:underline"
            >
              {activity.name}
            </Link>
          </div>

          {/* Sport Type */}
          <div className="flex-[1.5] min-w-[120px] px-4 py-3">
            <div className="flex items-center gap-2">
              <SportIcon
                sportType={activity.sport_type}
                className={cn('h-4 w-4 shrink-0', sportColor)}
              />
              <span className="text-sm truncate">{formatSportType(activity.sport_type)}</span>
            </div>
          </div>

          {/* Date */}
          <div className="flex-[1.2] min-w-[100px] px-4 py-3 text-muted-foreground">
            {formatDate(activity.start_date_local)}
          </div>

          {/* Distance */}
          <div className="flex-1 min-w-[80px] px-4 py-3 text-right tabular-nums">
            {formatDistance(activity.distance)}
          </div>

          {/* Time */}
          <div className="flex-1 min-w-[80px] px-4 py-3 text-right tabular-nums">
            {formatDuration(activity.moving_time)}
          </div>

          {/* Elevation */}
          <div className="flex-1 min-w-[80px] px-4 py-3 text-right tabular-nums">
            {formatElevation(activity.total_elevation_gain)}
          </div>

          {/* Tags */}
          <div className="flex-1 min-w-[100px] px-4 py-3">
            <div className="flex gap-1">
              {activity.commute && (
                <Badge variant="secondary" className="text-xs">
                  Commute
                </Badge>
              )}
              {activity.trainer && (
                <Badge variant="secondary" className="text-xs">
                  Indoor
                </Badge>
              )}
            </div>
          </div>
        </>
      )
    },
    [formatDistance, formatElevation]
  )

  if (isLoading) {
    return <VirtualizedTableSkeleton height={height} />
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
    <div className="rounded-lg border border-border overflow-hidden">
      {/* Fixed Header */}
      <div className="flex bg-muted/50 border-b border-border text-sm font-medium text-muted-foreground">
        <div className="flex-[3] min-w-[200px] px-4 py-3">Activity</div>
        <div className="flex-[1.5] min-w-[120px] px-4 py-3">Type</div>
        <div className="flex-[1.2] min-w-[100px] px-4 py-3">Date</div>
        <div className="flex-1 min-w-[80px] px-4 py-3 text-right">Distance</div>
        <div className="flex-1 min-w-[80px] px-4 py-3 text-right">Time</div>
        <div className="flex-1 min-w-[80px] px-4 py-3 text-right">Elevation</div>
        <div className="flex-1 min-w-[100px] px-4 py-3">Tags</div>
      </div>

      {/* Virtualized Body */}
      <div
        ref={parentRef}
        className="overflow-auto"
        style={{ height: typeof height === 'number' ? `${height}px` : height }}
      >
        <div
          style={{
            height: `${virtualizer.getTotalSize()}px`,
            width: '100%',
            position: 'relative',
          }}
        >
          {virtualRows.map((virtualRow) => {
            const activity = activities[virtualRow.index]
            return (
              <div
                key={activity.id}
                data-index={virtualRow.index}
                ref={virtualizer.measureElement}
                className={cn(
                  'absolute left-0 right-0 flex items-center border-b border-border/50',
                  'hover:bg-muted/50 transition-colors',
                  virtualRow.index % 2 === 0 ? 'bg-background' : 'bg-muted/20'
                )}
                style={{
                  top: 0,
                  transform: `translateY(${virtualRow.start}px)`,
                  height: `${ROW_HEIGHT}px`,
                }}
              >
                {renderRow(activity)}
              </div>
            )
          })}
        </div>
      </div>

      {/* Footer with count */}
      <div className="flex items-center justify-between px-4 py-2 bg-muted/30 border-t border-border text-sm text-muted-foreground">
        <span>Showing {activities.length} activities</span>
        <span className="text-xs">Virtual scroll enabled</span>
      </div>
    </div>
  )
}

function VirtualizedTableSkeleton({ height }: { height: number | string }) {
  return (
    <div className="rounded-lg border border-border overflow-hidden">
      {/* Fixed Header */}
      <div className="flex bg-muted/50 border-b border-border text-sm font-medium text-muted-foreground">
        <div className="flex-[3] min-w-[200px] px-4 py-3">Activity</div>
        <div className="flex-[1.5] min-w-[120px] px-4 py-3">Type</div>
        <div className="flex-[1.2] min-w-[100px] px-4 py-3">Date</div>
        <div className="flex-1 min-w-[80px] px-4 py-3 text-right">Distance</div>
        <div className="flex-1 min-w-[80px] px-4 py-3 text-right">Time</div>
        <div className="flex-1 min-w-[80px] px-4 py-3 text-right">Elevation</div>
        <div className="flex-1 min-w-[100px] px-4 py-3">Tags</div>
      </div>

      {/* Skeleton Body */}
      <div
        className="overflow-hidden"
        style={{ height: typeof height === 'number' ? `${height}px` : height }}
      >
        {Array.from({ length: 12 }).map((_, i) => (
          <div
            key={i}
            className={cn(
              'flex items-center border-b border-border/50',
              i % 2 === 0 ? 'bg-background' : 'bg-muted/20'
            )}
            style={{ height: `${ROW_HEIGHT}px` }}
          >
            <div className="flex-[3] min-w-[200px] px-4 py-3">
              <Skeleton className="h-4 w-48" />
            </div>
            <div className="flex-[1.5] min-w-[120px] px-4 py-3">
              <Skeleton className="h-4 w-20" />
            </div>
            <div className="flex-[1.2] min-w-[100px] px-4 py-3">
              <Skeleton className="h-4 w-24" />
            </div>
            <div className="flex-1 min-w-[80px] px-4 py-3 flex justify-end">
              <Skeleton className="h-4 w-16" />
            </div>
            <div className="flex-1 min-w-[80px] px-4 py-3 flex justify-end">
              <Skeleton className="h-4 w-16" />
            </div>
            <div className="flex-1 min-w-[80px] px-4 py-3 flex justify-end">
              <Skeleton className="h-4 w-16" />
            </div>
            <div className="flex-1 min-w-[100px] px-4 py-3">
              <Skeleton className="h-4 w-16" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
