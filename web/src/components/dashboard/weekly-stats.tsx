import type { WeeklyStat } from '@/lib/api'
import { formatDuration } from '@/lib/format'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { SportIcon } from '@/lib/sport-icon'
import { formatSportType, getSportHexColor } from '@/lib/sport-types'
import { WidgetWrapper } from './widget-wrapper'
import { Skeleton } from '@/components/ui/skeleton'

interface WeeklyStatsProps {
  stats: WeeklyStat[] | undefined
  isLoading: boolean
}

export function WeeklyStats({ stats, isLoading }: WeeklyStatsProps) {
  const { formatDistance } = useFormattedMetrics()
  if (isLoading) {
    return (
      <WidgetWrapper title="This Week">
        <div className="space-y-3">
          {[...Array(3)].map((_, i) => (
            <div key={i} className="flex items-center gap-3">
              <Skeleton className="h-8 w-8 rounded" />
              <div className="flex-1 space-y-1">
                <Skeleton className="h-4 w-20" />
                <Skeleton className="h-3 w-32" />
              </div>
            </div>
          ))}
        </div>
      </WidgetWrapper>
    )
  }

  const totalActivities = stats?.reduce((sum, s) => sum + s.activity_count, 0) ?? 0
  const totalDistance = stats?.reduce((sum, s) => sum + s.total_distance, 0) ?? 0
  const totalTime = stats?.reduce((sum, s) => sum + s.total_time, 0) ?? 0

  return (
    <WidgetWrapper title="This Week">
      {!stats || stats.length === 0 ? (
        <p className="text-sm text-muted-foreground">No activities this week</p>
      ) : (
        <div className="space-y-4">
          {/* Summary */}
          <div className="flex gap-4 text-sm pb-3 border-b">
            <div>
              <span className="font-medium">{totalActivities}</span>
              <span className="text-muted-foreground ml-1">
                {totalActivities === 1 ? 'activity' : 'activities'}
              </span>
            </div>
            <div>
              <span className="font-medium">{formatDistance(totalDistance)}</span>
            </div>
            <div>
              <span className="font-medium">{formatDuration(totalTime)}</span>
            </div>
          </div>

          {/* By sport type */}
          <div className="space-y-3">
            {stats.map((stat) => (
              <div key={stat.sport_type} className="flex items-center gap-3">
                <div
                  className="h-8 w-8 rounded flex items-center justify-center"
                  style={{ backgroundColor: getSportHexColor(stat.sport_type) + '20' }}
                >
                  <SportIcon
                    sportType={stat.sport_type}
                    className="h-4 w-4"
                    style={{ color: getSportHexColor(stat.sport_type) }}
                  />
                </div>
                <div className="flex-1">
                  <p className="font-medium text-sm">{formatSportType(stat.sport_type)}</p>
                  <p className="text-xs text-muted-foreground">
                    {stat.activity_count} {stat.activity_count === 1 ? 'activity' : 'activities'}
                    {' · '}
                    {formatDistance(stat.total_distance)}
                    {' · '}
                    {formatDuration(stat.total_time)}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </WidgetWrapper>
  )
}
