import { useMemo } from 'react'
import { useTrainingGoals } from '@/lib/data'
import { WidgetWrapper } from './widget-wrapper'
import { Check, X } from 'lucide-react'
import { cn } from '@/lib/utils'

function getRecentMonths(count: number): { key: string; label: string }[] {
  const months: { key: string; label: string }[] = []
  const now = new Date()

  for (let i = count - 1; i >= 0; i--) {
    const date = new Date(now.getFullYear(), now.getMonth() - i, 1)
    const key = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`
    const label = date.toLocaleDateString(undefined, { month: 'short' })
    months.push({ key, label })
  }

  return months
}

interface GoalStatus {
  name: string
  months: Map<string, boolean>
}

export function ChallengeConsistencyGrid() {
  const { data: goalsData, isLoading } = useTrainingGoals()

  const months = useMemo(() => getRecentMonths(12), [])

  // Build goal status grid from the training goals progress
  const goalStatuses: GoalStatus[] = useMemo(() => {
    if (!goalsData?.progress) return []

    const statuses: GoalStatus[] = []
    const currentMonth = new Date().toISOString().slice(0, 7)

    // Progress is keyed by sport name, then by period
    for (const [sportName, periods] of Object.entries(goalsData.progress)) {
      const monthlyProgress = periods.month
      if (!monthlyProgress) continue

      // Check if any metric has a target set in config
      const sportConfig = goalsData.config?.sports.find((s) => s.name === sportName)
      if (!sportConfig?.targets?.month) continue

      const targets = sportConfig.targets.month
      const monthsMap = new Map<string, boolean>()

      // Check distance goal
      if (targets.distance_m && targets.distance_m > 0) {
        const isComplete = monthlyProgress.distance_m >= targets.distance_m
        monthsMap.set(currentMonth, isComplete)
        statuses.push({
          name: `${sportName} distance`,
          months: new Map(monthsMap),
        })
      }

      // Check elevation goal
      if (targets.elevation_m && targets.elevation_m > 0) {
        const isComplete = monthlyProgress.elevation_m >= targets.elevation_m
        statuses.push({
          name: `${sportName} elevation`,
          months: new Map([[currentMonth, isComplete]]),
        })
      }

      // Check time goal
      if (targets.moving_time_s && targets.moving_time_s > 0) {
        const isComplete = monthlyProgress.moving_time_s >= targets.moving_time_s
        statuses.push({
          name: `${sportName} time`,
          months: new Map([[currentMonth, isComplete]]),
        })
      }
    }

    return statuses.slice(0, 5) // Limit to 5 goals for display
  }, [goalsData])

  const hasGoals = goalStatuses.length > 0

  return (
    <WidgetWrapper title="Goal Consistency" isLoading={isLoading}>
      {!hasGoals ? (
        <div className="text-sm text-muted-foreground">
          No monthly goals configured. Set up goals in Training Goals widget.
        </div>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr>
                <th className="py-1 text-left font-medium text-muted-foreground">Goal</th>
                {months.map((m) => (
                  <th
                    key={m.key}
                    className="py-1 text-center font-medium text-muted-foreground"
                    style={{ writingMode: 'vertical-rl', textOrientation: 'mixed' }}
                  >
                    {m.label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {goalStatuses.map((goal, idx) => (
                <tr key={idx} className="border-t border-border/50">
                  <td className="py-1.5 pr-2 font-medium truncate max-w-[120px]" title={goal.name}>
                    {goal.name}
                  </td>
                  {months.map((m) => {
                    const status = goal.months.get(m.key)
                    return (
                      <td key={m.key} className="py-1.5 text-center">
                        {status === undefined ? (
                          <span className="text-muted-foreground/30">—</span>
                        ) : status ? (
                          <span
                            className={cn(
                              'inline-flex items-center justify-center w-5 h-5 rounded',
                              'bg-green-100 text-green-600 dark:bg-green-900/30 dark:text-green-400'
                            )}
                          >
                            <Check className="h-3 w-3" />
                          </span>
                        ) : (
                          <span
                            className={cn(
                              'inline-flex items-center justify-center w-5 h-5 rounded',
                              'bg-red-100 text-red-600 dark:bg-red-900/30 dark:text-red-400'
                            )}
                          >
                            <X className="h-3 w-3" />
                          </span>
                        )}
                      </td>
                    )
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </WidgetWrapper>
  )
}
