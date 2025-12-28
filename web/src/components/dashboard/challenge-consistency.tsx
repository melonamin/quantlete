import { Link } from '@tanstack/react-router'
import { useChallenges } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { BarChart } from '@/components/charts'
import { Button } from '@/components/ui/button'

function monthLabel(month: string) {
  const [y, m] = month.split('-').map((x) => Number(x))
  if (!y || !m) return month
  return new Date(y, m - 1, 1).toLocaleDateString(undefined, { month: 'short', year: '2-digit' })
}

export function ChallengeConsistency() {
  const { data, isLoading } = useChallenges()

  const counts = new Map<string, number>()
  for (const c of data ?? []) {
    const month = c.month || ''
    if (!month) continue
    counts.set(month, (counts.get(month) ?? 0) + 1)
  }

  const months = Array.from(counts.keys()).sort()
  const recentMonths = months.slice(Math.max(0, months.length - 12))

  const chartData = recentMonths.map((m) => ({
    label: monthLabel(m),
    value: counts.get(m) ?? 0,
  }))

  // Simple streak: consecutive months with >=1 challenge, ending at the latest month.
  let streak = 0
  for (let i = recentMonths.length - 1; i >= 0; i--) {
    if ((counts.get(recentMonths[i]) ?? 0) > 0) streak++
    else break
  }

  return (
    <WidgetWrapper
      title="Challenge Consistency"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/challenges">View all</Link>
        </Button>
      }
      isLoading={isLoading}
    >
      {chartData.length === 0 ? (
        <div className="text-sm text-muted-foreground">No challenges imported yet.</div>
      ) : (
        <>
          <div className="mb-2 text-sm text-muted-foreground">
            {streak} month{streak === 1 ? '' : 's'} in a row with at least 1 challenge.
          </div>
          <BarChart data={chartData} height={220} showValues />
        </>
      )}
    </WidgetWrapper>
  )
}
