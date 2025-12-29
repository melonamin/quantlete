import { Link } from '@tanstack/react-router'
import type { RecentActivity } from '@/lib/api'
import { formatDistance, formatDuration, formatRelativeDate } from '@/lib/format'
import { getSportIcon, getSportHexColor } from '@/lib/sport-types'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

interface RecentActivitiesProps {
  activities: RecentActivity[] | undefined
  isLoading: boolean
}

export function RecentActivities({ activities, isLoading }: RecentActivitiesProps) {
  if (isLoading) {
    return (
      <WidgetWrapper
        title="Recent Activities"
        action={
          <Button variant="ghost" size="sm" disabled>
            View all
          </Button>
        }
      >
        <div className="space-y-4">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="flex items-center gap-3">
              <Skeleton className="h-10 w-10 rounded-full" />
              <div className="flex-1 space-y-1">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-3 w-24" />
              </div>
              <Skeleton className="h-4 w-16" />
            </div>
          ))}
        </div>
      </WidgetWrapper>
    )
  }

  return (
    <WidgetWrapper
      title="Recent Activities"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/activities">View all</Link>
        </Button>
      }
    >
      {!activities || activities.length === 0 ? (
        <p className="text-sm text-muted-foreground">No activities yet</p>
      ) : (
        <div className="space-y-4">
          {activities.map((activity) => {
            const Icon = getSportIcon(activity.sport_type)
            return (
              <Link
                key={activity.id}
                to="/activities/$activityId"
                params={{ activityId: String(activity.id) }}
                className="flex items-center gap-3 hover:bg-muted/50 -mx-2 px-2 py-1 rounded-md transition-colors"
              >
                <div
                  className="h-10 w-10 rounded-full flex items-center justify-center text-lg"
                  style={{ backgroundColor: getSportHexColor(activity.sport_type) + '20' }}
                >
                  <Icon
                    className="h-5 w-5"
                    style={{ color: getSportHexColor(activity.sport_type) }}
                  />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="font-medium truncate">{activity.name}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatRelativeDate(activity.start_date)}
                  </p>
                </div>
                <div className="text-right text-sm">
                  <p className="font-medium">{formatDistance(activity.distance)}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatDuration(activity.moving_time)}
                  </p>
                </div>
              </Link>
            )
          })}
        </div>
      )}
    </WidgetWrapper>
  )
}
