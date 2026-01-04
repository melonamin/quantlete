import { useMemo } from 'react'
import { useMonthlyStats } from '@/lib/api'
import { useTrainingGoals, useUpdateTrainingGoals } from '@/lib/api/goals'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { formatDistance, formatDuration } from '@/lib/format'
import { Target, TrendingUp, Zap, Trophy } from 'lucide-react'
import { uiColors } from '@/components/charts'

type Difficulty = 'easy' | 'moderate' | 'stretch'

interface GoalRecommendation {
  id: string
  metric: 'distance' | 'elevation' | 'time'
  period: 'week' | 'month'
  target: number
  current: number
  difficulty: Difficulty
  reason: string
}

const DIFFICULTY_CONFIG: Record<Difficulty, { label: string; color: string; icon: typeof Target }> =
  {
    easy: { label: 'Easy', color: uiColors.success, icon: Target },
    moderate: { label: 'Moderate', color: uiColors.warning, icon: TrendingUp },
    stretch: { label: 'Stretch', color: uiColors.accent, icon: Zap },
  }

export function GoalRecommendations() {
  const { data: monthlyData, isLoading: monthlyLoading } = useMonthlyStats()
  const { data: goalsData, isLoading: goalsLoading } = useTrainingGoals()
  const updateGoals = useUpdateTrainingGoals()

  const isLoading = monthlyLoading || goalsLoading

  // Analyze historical data to generate recommendations
  const recommendations = useMemo((): GoalRecommendation[] => {
    if (!monthlyData?.length) return []

    const recs: GoalRecommendation[] = []

    // Get data from the last 6 months for analysis
    const recentMonths = monthlyData.slice(0, 6)
    if (recentMonths.length < 2) return []

    // Calculate averages
    const avgDistance =
      recentMonths.reduce((sum, m) => sum + m.total_distance, 0) / recentMonths.length
    const avgElevation =
      recentMonths.reduce((sum, m) => sum + m.total_elevation, 0) / recentMonths.length
    const avgTime = recentMonths.reduce((sum, m) => sum + m.total_time, 0) / recentMonths.length

    // Calculate best month
    const bestDistance = Math.max(...recentMonths.map((m) => m.total_distance))
    const bestElevation = Math.max(...recentMonths.map((m) => m.total_elevation))

    // Check if we already have goals set
    const hasDistanceGoal =
      goalsData?.config?.sports?.some((s) => s.targets?.month?.distance_m) ?? false
    const hasElevationGoal =
      goalsData?.config?.sports?.some((s) => s.targets?.month?.elevation_m) ?? false
    const hasTimeGoal =
      goalsData?.config?.sports?.some((s) => s.targets?.month?.moving_time_s) ?? false

    // Generate distance recommendations
    if (!hasDistanceGoal && avgDistance > 0) {
      // Easy: slightly above average
      const easyTarget = Math.ceil((avgDistance * 1.1) / 1000) * 1000
      recs.push({
        id: 'distance-easy',
        metric: 'distance',
        period: 'month',
        target: easyTarget,
        current: avgDistance,
        difficulty: 'easy',
        reason: `Slightly above your ${formatDistance(avgDistance)} monthly average`,
      })

      // Stretch: match your best month
      if (bestDistance > avgDistance * 1.15) {
        recs.push({
          id: 'distance-stretch',
          metric: 'distance',
          period: 'month',
          target: Math.ceil(bestDistance / 1000) * 1000,
          current: avgDistance,
          difficulty: 'stretch',
          reason: `Match your best month of ${formatDistance(bestDistance)}`,
        })
      }
    }

    // Generate elevation recommendations
    if (!hasElevationGoal && avgElevation > 100) {
      // Easy: round up to nearest 100m above average
      const easyTarget = Math.ceil((avgElevation * 1.1) / 100) * 100
      recs.push({
        id: 'elevation-easy',
        metric: 'elevation',
        period: 'month',
        target: easyTarget,
        current: avgElevation,
        difficulty: 'easy',
        reason: `10% above your ${Math.round(avgElevation)}m monthly average`,
      })

      // Moderate: 25% above average
      if (bestElevation > avgElevation * 1.2) {
        recs.push({
          id: 'elevation-moderate',
          metric: 'elevation',
          period: 'month',
          target: Math.ceil((avgElevation * 1.25) / 100) * 100,
          current: avgElevation,
          difficulty: 'moderate',
          reason: `Challenge yourself to 25% above average`,
        })
      }
    }

    // Generate time recommendations
    if (!hasTimeGoal && avgTime > 3600) {
      // in seconds, at least 1 hour average
      // Easy: round up to nearest hour above average
      const easyTarget = Math.ceil((avgTime * 1.1) / 3600) * 3600
      recs.push({
        id: 'time-easy',
        metric: 'time',
        period: 'month',
        target: easyTarget,
        current: avgTime,
        difficulty: 'easy',
        reason: `Just above your ${formatDuration(avgTime)} monthly average`,
      })
    }

    // Weekly recommendations based on monthly data
    const weeklyAvgDistance = avgDistance / 4
    if (!hasDistanceGoal && weeklyAvgDistance > 5000) {
      recs.push({
        id: 'weekly-distance',
        metric: 'distance',
        period: 'week',
        target: Math.ceil((weeklyAvgDistance * 1.15) / 1000) * 1000,
        current: weeklyAvgDistance,
        difficulty: 'moderate',
        reason: `15% above your typical weekly distance`,
      })
    }

    // Limit to top 4 recommendations
    return recs.slice(0, 4)
  }, [monthlyData, goalsData])

  const handleAcceptGoal = (rec: GoalRecommendation) => {
    if (!goalsData?.config) return

    const config = structuredClone(goalsData.config)

    // Find or create an "All Activities" sport config
    let sport = config.sports.find((s) => s.name === 'All Activities')
    if (!sport) {
      sport = { name: 'All Activities', sport_types: [], targets: {} }
      config.sports.push(sport)
    }

    // Initialize targets if needed
    sport.targets = sport.targets ?? {}
    sport.targets[rec.period] = sport.targets[rec.period] ?? {}

    // Set the target based on metric
    switch (rec.metric) {
      case 'distance':
        sport.targets[rec.period]!.distance_m = rec.target
        break
      case 'elevation':
        sport.targets[rec.period]!.elevation_m = rec.target
        break
      case 'time':
        sport.targets[rec.period]!.moving_time_s = rec.target
        break
    }

    updateGoals.mutate(config)
  }

  const formatTarget = (rec: GoalRecommendation): string => {
    switch (rec.metric) {
      case 'distance':
        return formatDistance(rec.target)
      case 'elevation':
        return `${Math.round(rec.target)}m`
      case 'time':
        return formatDuration(rec.target)
    }
  }

  const getMetricLabel = (metric: string): string => {
    switch (metric) {
      case 'distance':
        return 'Distance'
      case 'elevation':
        return 'Elevation'
      case 'time':
        return 'Time'
      default:
        return metric
    }
  }

  return (
    <WidgetWrapper title="Goal Suggestions" isLoading={isLoading}>
      <div className="h-full flex flex-col">
        {recommendations.length === 0 ? (
          <div className="flex-1 flex items-center justify-center text-center">
            <div>
              <Trophy className="h-8 w-8 mx-auto mb-2 text-muted-foreground/50" />
              <p className="text-sm text-muted-foreground">
                {isLoading
                  ? 'Analyzing your activity data...'
                  : 'Not enough activity data yet to suggest goals.'}
              </p>
              <p className="text-xs text-muted-foreground mt-1">
                Record a few more activities for personalized recommendations.
              </p>
            </div>
          </div>
        ) : (
          <div className="flex-1 space-y-2 overflow-auto">
            <p className="text-xs text-muted-foreground mb-3">
              Based on your recent performance, try these goals:
            </p>
            {recommendations.map((rec) => {
              const { label, color, icon: Icon } = DIFFICULTY_CONFIG[rec.difficulty]
              return (
                <div
                  key={rec.id}
                  className="flex items-start gap-3 p-2 rounded-sm bg-muted/30 hover:bg-muted/50 transition-colors"
                >
                  <Icon className="h-4 w-4 mt-0.5 shrink-0" style={{ color }} />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <span className="font-medium text-sm">
                        {formatTarget(rec)} / {rec.period}
                      </span>
                      <span className="text-[10px] uppercase" style={{ color }}>
                        {label}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground truncate">{rec.reason}</p>
                    <div className="text-[10px] text-muted-foreground/70">
                      {getMetricLabel(rec.metric)}
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    className="h-7 px-2 text-xs shrink-0"
                    onClick={() => handleAcceptGoal(rec)}
                    disabled={updateGoals.isPending}
                  >
                    Set Goal
                  </Button>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </WidgetWrapper>
  )
}
