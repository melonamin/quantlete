import { useMemo } from 'react'
import { useDashboard, useMonthlyStats, useActivities } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import {
  Lightbulb,
  TrendingUp,
  TrendingDown,
  Trophy,
  Flame,
  Calendar,
  Clock,
  MapPin,
  Zap,
} from 'lucide-react'
import { PAGINATION } from '@/lib/constants'
import { iconColors, uiColors } from '@/components/charts'

interface Insight {
  id: string
  type: 'achievement' | 'trend' | 'observation' | 'streak' | 'record'
  icon: React.ComponentType<{ className?: string; style?: React.CSSProperties }>
  iconColor: string
  title: string
  description: string
}

export function ActivityInsights() {
  const { data: dashboard, isLoading: dashboardLoading } = useDashboard()
  const { data: monthlyData, isLoading: monthlyLoading } = useMonthlyStats()
  const { data: recentActivities, isLoading: activitiesLoading } = useActivities({
    per_page: PAGINATION.DEFAULT_PER_PAGE,
    order_by: 'start_date',
    order_dir: 'desc',
  })
  const { formatDistance } = useFormattedMetrics()

  const isLoading = dashboardLoading || monthlyLoading || activitiesLoading

  const insights = useMemo((): Insight[] => {
    const result: Insight[] = []
    // Use Array.isArray for defensive check - recentActivities?.data could be null/undefined
    // during WASM initialization or if the query returns unexpected data
    const activities = Array.isArray(recentActivities?.data) ? recentActivities.data : []
    const stats = dashboard?.stats

    if (!stats || activities.length < 5) return result

    // Analyze streaks - consecutive days with activities
    const activityDates = new Set(
      activities.map((a) => new Date(a.start_date_local).toDateString())
    )
    let streak = 0
    const today = new Date()
    for (let i = 0; i < 30; i++) {
      const date = new Date(today)
      date.setDate(date.getDate() - i)
      if (activityDates.has(date.toDateString())) {
        streak++
      } else if (streak > 0) {
        break
      }
    }

    if (streak >= 3) {
      result.push({
        id: 'streak',
        type: 'streak',
        icon: Flame,
        iconColor: iconColors.fire,
        title: `${streak}-day streak!`,
        description:
          streak >= 7
            ? "You've been incredibly consistent!"
            : "Keep it up, you're building momentum!",
      })
    }

    // Monthly trend analysis
    if (monthlyData && monthlyData.length >= 2) {
      const thisMonth = monthlyData[0]
      const lastMonth = monthlyData[1]

      if (thisMonth && lastMonth && lastMonth.total_distance > 0) {
        const distanceChange =
          ((thisMonth.total_distance - lastMonth.total_distance) / lastMonth.total_distance) * 100
        if (distanceChange > 20) {
          result.push({
            id: 'distance-up',
            type: 'trend',
            icon: TrendingUp,
            iconColor: iconColors.trend,
            title: `${Math.round(distanceChange)}% more distance`,
            description: `You've covered ${formatDistance(thisMonth.total_distance)} this month, up from ${formatDistance(lastMonth.total_distance)} last month.`,
          })
        } else if (distanceChange < -20) {
          result.push({
            id: 'distance-down',
            type: 'observation',
            icon: TrendingDown,
            iconColor: uiColors.warning,
            title: 'Quieter month',
            description: `You're at ${formatDistance(thisMonth.total_distance)} so far. Last month you did ${formatDistance(lastMonth.total_distance)}.`,
          })
        }
      }
    }

    // Find longest recent activity
    if (activities.length > 0) {
      const longestRecent = activities.reduce((max, a) => (a.distance > max.distance ? a : max))
      const avgDistance = activities.reduce((sum, a) => sum + a.distance, 0) / activities.length

      if (longestRecent.distance > avgDistance * 2) {
        result.push({
          id: 'long-activity',
          type: 'achievement',
          icon: Trophy,
          iconColor: iconColors.trophy,
          title: 'Big effort!',
          description: `Your ${longestRecent.name} (${formatDistance(longestRecent.distance)}) was ${Math.round((longestRecent.distance / avgDistance - 1) * 100)}% longer than your average.`,
        })
      }
    }

    // Analyze activity timing patterns
    const morningActivities = activities.filter((a) => {
      const hour = new Date(a.start_date_local).getHours()
      return hour >= 5 && hour < 10
    }).length
    const eveningActivities = activities.filter((a) => {
      const hour = new Date(a.start_date_local).getHours()
      return hour >= 17 && hour < 21
    }).length

    if (morningActivities > eveningActivities * 2 && morningActivities >= 5) {
      result.push({
        id: 'morning-person',
        type: 'observation',
        icon: Clock,
        iconColor: iconColors.clock,
        title: 'Early bird',
        description: `${morningActivities} of your recent ${activities.length} activities started in the morning.`,
      })
    } else if (eveningActivities > morningActivities * 2 && eveningActivities >= 5) {
      result.push({
        id: 'evening-person',
        type: 'observation',
        icon: Clock,
        iconColor: '#a78bfa', // purple for evening
        title: 'Evening athlete',
        description: `You prefer evening workouts, with ${eveningActivities} recent sessions after 5pm.`,
      })
    }

    // Location variety
    const locations = new Set(
      activities.filter((a) => a.location_country).map((a) => a.location_country)
    )
    if (locations.size > 2) {
      result.push({
        id: 'locations',
        type: 'observation',
        icon: MapPin,
        iconColor: iconColors.location,
        title: `Active in ${locations.size} countries`,
        description: `You've recently recorded activities in ${Array.from(locations).slice(0, 3).join(', ')}.`,
      })
    }

    // Power/HR intensity analysis (if data available)
    const activitiesWithPower = activities.filter((a) => a.average_watts && a.average_watts > 0)
    if (activitiesWithPower.length >= 3) {
      const recentPower = activitiesWithPower.slice(0, 5)
      const avgPower =
        recentPower.reduce((sum, a) => sum + (a.average_watts ?? 0), 0) / recentPower.length
      const olderPower = activitiesWithPower.slice(5, 10)

      if (olderPower.length >= 3) {
        const olderAvgPower =
          olderPower.reduce((sum, a) => sum + (a.average_watts ?? 0), 0) / olderPower.length
        const powerChange = ((avgPower - olderAvgPower) / olderAvgPower) * 100

        if (powerChange > 5) {
          result.push({
            id: 'power-up',
            type: 'trend',
            icon: Zap,
            iconColor: iconColors.power,
            title: 'Power increasing',
            description: `Your recent average power is ${Math.round(powerChange)}% higher (${Math.round(avgPower)}W vs ${Math.round(olderAvgPower)}W).`,
          })
        }
      }
    }

    // Milestone check
    if (stats.total_distance > 0) {
      const totalKm = stats.total_distance / 1000
      const milestones = [10000, 5000, 2500, 1000, 500]
      for (const milestone of milestones) {
        if (totalKm >= milestone && totalKm < milestone * 1.1) {
          result.push({
            id: 'milestone',
            type: 'record',
            icon: Trophy,
            iconColor: iconColors.trophy,
            title: `${milestone.toLocaleString()} km milestone!`,
            description: `You've covered ${formatDistance(stats.total_distance)} in total. Amazing!`,
          })
          break
        }
      }
    }

    // Activity count milestone
    if (stats.total_activities > 0) {
      const milestones = [1000, 500, 250, 100]
      for (const milestone of milestones) {
        if (stats.total_activities >= milestone && stats.total_activities < milestone * 1.05) {
          result.push({
            id: 'activity-milestone',
            type: 'record',
            icon: Calendar,
            iconColor: iconColors.calendar,
            title: `${milestone} activities!`,
            description: `You've recorded ${stats.total_activities} total activities. Impressive dedication!`,
          })
          break
        }
      }
    }

    // Limit to 4 insights, prioritize by type
    const priority = { record: 0, achievement: 1, streak: 2, trend: 3, observation: 4 }
    return result.sort((a, b) => priority[a.type] - priority[b.type]).slice(0, 4)
  }, [dashboard, monthlyData, recentActivities, formatDistance])

  return (
    <WidgetWrapper title="Insights" isLoading={isLoading}>
      <div className="h-full flex flex-col">
        {insights.length === 0 ? (
          <div className="flex-1 flex items-center justify-center text-center">
            <div>
              <Lightbulb className="h-8 w-8 mx-auto mb-2 text-muted-foreground/50" />
              <p className="text-sm text-muted-foreground">
                {isLoading
                  ? 'Analyzing your activities...'
                  : 'Record more activities to unlock insights.'}
              </p>
            </div>
          </div>
        ) : (
          <div className="flex-1 space-y-2 overflow-auto">
            {insights.map((insight) => {
              const Icon = insight.icon
              return (
                <div key={insight.id} className="flex items-start gap-3 p-2 rounded-sm bg-muted/30">
                  <Icon className="h-4 w-4 mt-0.5 shrink-0" style={{ color: insight.iconColor }} />
                  <div className="min-w-0">
                    <div className="font-medium text-sm">{insight.title}</div>
                    <p className="text-xs text-muted-foreground">{insight.description}</p>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </WidgetWrapper>
  )
}
