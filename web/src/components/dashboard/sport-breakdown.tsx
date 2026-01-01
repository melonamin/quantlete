import type { SportTypeStat } from '@/lib/api'
import { SportIcon } from '@/lib/sport-icon'
import { formatSportType, getSportHexColor } from '@/lib/sport-types'
import { WidgetWrapper } from './widget-wrapper'
import { Skeleton } from '@/components/ui/skeleton'

interface SportBreakdownProps {
  stats: SportTypeStat[] | undefined
  isLoading: boolean
}

export function SportBreakdown({ stats, isLoading }: SportBreakdownProps) {
  if (isLoading) {
    return (
      <WidgetWrapper title="By Sport">
        <div className="space-y-3">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="flex items-center gap-3">
              <Skeleton className="h-8 w-8 rounded" />
              <div className="flex-1">
                <Skeleton className="h-4 w-20" />
              </div>
              <Skeleton className="h-4 w-16" />
            </div>
          ))}
        </div>
      </WidgetWrapper>
    )
  }

  const totalActivities = stats?.reduce((sum, s) => sum + s.activity_count, 0) ?? 0

  return (
    <WidgetWrapper title="By Sport">
      {!stats || stats.length === 0 ? (
        <p className="text-sm text-muted-foreground">No activities yet</p>
      ) : (
        <div className="space-y-3">
          {stats.slice(0, 6).map((stat) => {
            const percentage =
              totalActivities > 0 ? Math.round((stat.activity_count / totalActivities) * 100) : 0

            return (
              <div key={stat.sport_type} className="flex items-center gap-3">
                <div
                  className="h-8 w-8 rounded flex items-center justify-center flex-shrink-0"
                  style={{ backgroundColor: getSportHexColor(stat.sport_type) + '20' }}
                >
                  <SportIcon sportType={stat.sport_type} className="h-4 w-4" style={{ color: getSportHexColor(stat.sport_type) }} />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <p className="font-medium text-sm truncate">
                      {formatSportType(stat.sport_type)}
                    </p>
                    <span className="text-xs text-muted-foreground ml-2">{percentage}%</span>
                  </div>
                  <div className="h-1.5 bg-muted rounded-full mt-1 overflow-hidden">
                    <div
                      className="h-full rounded-full transition-all"
                      style={{
                        width: `${percentage}%`,
                        backgroundColor: getSportHexColor(stat.sport_type),
                      }}
                    />
                  </div>
                </div>
              </div>
            )
          })}

          {/* Summary row */}
          <div className="pt-2 border-t text-xs text-muted-foreground">
            {stats.length} sport types · {totalActivities.toLocaleString()} total activities
          </div>
        </div>
      )}
    </WidgetWrapper>
  )
}
