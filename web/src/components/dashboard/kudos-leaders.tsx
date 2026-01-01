import { Link } from '@tanstack/react-router'
import { useActivities } from '@/lib/data/hooks'
import { formatRelativeDate } from '@/lib/format'
import { SportIcon } from '@/lib/sport-icon'
import { getSportHexColor } from '@/lib/sport-types'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ThumbsUp } from 'lucide-react'

export function KudosLeaders() {
  const { data, isLoading } = useActivities({
    per_page: 5,
    order_by: 'kudos_count',
    order_dir: 'desc',
  })

  const activities = data?.data

  if (isLoading) {
    return (
      <WidgetWrapper
        title="Most Kudos'd"
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
              <Skeleton className="h-4 w-12" />
            </div>
          ))}
        </div>
      </WidgetWrapper>
    )
  }

  // Filter to only activities with kudos
  const activitiesWithKudos = activities?.filter((a) => a.kudos_count > 0) ?? []

  return (
    <WidgetWrapper
      title="Most Kudos'd"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/activities">View all</Link>
        </Button>
      }
    >
      {activitiesWithKudos.length === 0 ? (
        <p className="text-sm text-muted-foreground">No activities with kudos yet</p>
      ) : (
        <div className="space-y-4">
          {activitiesWithKudos.map((activity) => (
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
                  <SportIcon
                    sportType={activity.sport_type}
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
                <div className="flex items-center gap-1 text-sm text-muted-foreground">
                  <ThumbsUp className="h-3.5 w-3.5" />
                  <span className="font-medium">{activity.kudos_count}</span>
                </div>
              </Link>
            ))}
        </div>
      )}
    </WidgetWrapper>
  )
}
