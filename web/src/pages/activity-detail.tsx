import { useParams, Link } from '@tanstack/react-router'
import { useActivity } from '@/lib/api'
import { ActivityHeader, ActivityStats } from '@/components/activities'
import { ActivityMap } from '@/components/maps'
import { Skeleton } from '@/components/ui/skeleton'
import { ChevronLeft } from 'lucide-react'

export function ActivityDetailPage() {
  const { activityId } = useParams({ from: '/activities/$activityId' })
  const { data: activity, isLoading, error } = useActivity(Number(activityId))

  if (isLoading) {
    return <ActivityDetailSkeleton />
  }

  if (error || !activity) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="mb-4">
          <Link
            to="/activities"
            className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground"
          >
            <ChevronLeft className="h-4 w-4 mr-1" />
            Back to Activities
          </Link>
        </div>
        <div className="rounded-lg border border-destructive bg-destructive/10 p-8 text-center">
          <p className="text-destructive">
            {error ? `Failed to load activity: ${error.message}` : 'Activity not found'}
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-4">
        <Link
          to="/activities"
          className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground"
        >
          <ChevronLeft className="h-4 w-4 mr-1" />
          Back to Activities
        </Link>
      </div>

      <ActivityHeader activity={activity} />
      <ActivityStats activity={activity} />

      {activity.summary_polyline && (
        <div className="mt-6">
          <ActivityMap
            polyline={activity.summary_polyline}
            className="h-[400px]"
          />
        </div>
      )}
    </div>
  )
}

function ActivityDetailSkeleton() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-4">
        <Skeleton className="h-4 w-32" />
      </div>
      <div className="mb-6 flex items-start gap-4">
        <Skeleton className="h-12 w-12 rounded-lg" />
        <div className="space-y-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-48" />
        </div>
      </div>
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <Skeleton key={i} className="h-24 rounded-lg" />
        ))}
      </div>
    </div>
  )
}
