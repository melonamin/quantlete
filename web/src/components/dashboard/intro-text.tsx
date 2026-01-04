import { useMemo, useState } from 'react'
import { useDashboard } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { formatDuration } from '@/lib/format'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { Trophy, Activity, Route, Clock, Mountain, Flame } from 'lucide-react'
import { iconColors } from '@/components/charts'

interface StatHighlight {
  icon: React.ComponentType<{ className?: string }>
  label: string
  value: string
}

export function IntroText() {
  const { data, isLoading } = useDashboard()
  const { formatDistance, formatElevation } = useFormattedMetrics()

  const highlights = useMemo((): StatHighlight[] => {
    if (!data?.stats) return []

    const stats = data.stats
    const items: StatHighlight[] = []

    if (stats.total_activities > 0) {
      items.push({
        icon: Activity,
        label: 'Activities',
        value: stats.total_activities.toLocaleString(),
      })
    }

    if (stats.total_distance > 0) {
      items.push({
        icon: Route,
        label: 'Distance',
        value: formatDistance(stats.total_distance),
      })
    }

    if (stats.total_elevation_gain > 0) {
      items.push({
        icon: Mountain,
        label: 'Elevation',
        value: formatElevation(stats.total_elevation_gain),
      })
    }

    if (stats.total_moving_time > 0) {
      items.push({
        icon: Clock,
        label: 'Time',
        value: formatDuration(stats.total_moving_time),
      })
    }

    return items
  }, [data, formatDistance, formatElevation])

  const welcomeMessage = useMemo(() => {
    if (!data?.stats) return 'Welcome to Quantlete!'

    const activities = data.stats.total_activities
    if (activities === 0) {
      return "Let's get started!"
    }

    if (activities < 50) {
      return "You're building momentum!"
    }

    if (activities < 200) {
      return 'Great consistency!'
    }

    if (activities < 500) {
      return 'Impressive dedication!'
    }

    return 'Legendary athlete!'
  }, [data])

  // Use lazy state initializer since Date.now() is impure
  const [motivationalQuote] = useState(() => {
    const quotes = [
      'Every journey begins with a single step.',
      "The only bad workout is the one you didn't do.",
      'Consistency beats intensity.',
      'Progress, not perfection.',
      "Your body can stand almost anything. It's your mind you have to convince.",
      'The pain you feel today is the strength you feel tomorrow.',
      "Don't count the days, make the days count.",
    ]
    // Use date-based selection for variety
    const dayOfYear = Math.floor(
      (Date.now() - new Date(new Date().getFullYear(), 0, 0).getTime()) / (1000 * 60 * 60 * 24)
    )
    return quotes[dayOfYear % quotes.length]
  })

  return (
    <WidgetWrapper title="Welcome" isLoading={isLoading}>
      <div className="h-full flex flex-col">
        {/* Welcome Header */}
        <div className="mb-4 flex-shrink-0">
          <div className="flex items-center gap-2 mb-1">
            <Flame className="h-5 w-5" style={{ color: iconColors.fire }} />
            <span className="text-lg font-semibold">{welcomeMessage}</span>
          </div>
          <p className="text-sm text-muted-foreground italic">"{motivationalQuote}"</p>
        </div>

        {/* Lifetime Stats Grid */}
        {highlights.length > 0 && (
          <div className="flex-1 min-h-0">
            <div className="text-xs text-muted-foreground uppercase tracking-wider mb-2">
              Lifetime Stats
            </div>
            <div className="grid grid-cols-2 gap-3">
              {highlights.map((stat) => {
                const Icon = stat.icon
                return (
                  <div
                    key={stat.label}
                    className="flex items-center gap-2 p-2 rounded-sm bg-accent/30"
                  >
                    <Icon className="h-4 w-4 text-muted-foreground shrink-0" />
                    <div className="min-w-0">
                      <div className="text-[10px] text-muted-foreground uppercase tracking-wider truncate">
                        {stat.label}
                      </div>
                      <div className="font-semibold truncate">{stat.value}</div>
                    </div>
                  </div>
                )
              })}
            </div>
          </div>
        )}

        {/* Trophy icon for achievements */}
        {data?.stats && data.stats.total_activities >= 100 && (
          <div className="mt-3 pt-3 border-t border-border flex-shrink-0">
            <div className="flex items-center gap-2 text-sm">
              <Trophy className="h-4 w-4" style={{ color: iconColors.trophy }} />
              <span className="text-muted-foreground">
                {data.stats.total_activities >= 1000
                  ? 'Master Athlete'
                  : data.stats.total_activities >= 500
                    ? 'Elite Athlete'
                    : data.stats.total_activities >= 250
                      ? 'Dedicated Athlete'
                      : 'Century Club Member'}
              </span>
            </div>
          </div>
        )}
      </div>
    </WidgetWrapper>
  )
}
