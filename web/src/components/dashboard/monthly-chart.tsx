import { useState } from 'react'
import { useMonthlyStats, useYearlyStats } from '@/lib/api'
import { MonthlyStatsChart } from '@/components/charts/activity-charts'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'

type Metric = 'distance' | 'count' | 'time'

export function MonthlyChart() {
  const currentYear = new Date().getFullYear()
  const [year, setYear] = useState(currentYear)
  const [metric, setMetric] = useState<Metric>('distance')

  const { data: yearlyStats } = useYearlyStats()
  const { data: monthlyStats, isLoading } = useMonthlyStats(year)

  const availableYears = yearlyStats?.map((s) => s.year) ?? [currentYear]

  const chartData =
    monthlyStats?.map((s) => ({
      month: s.month.split('-')[1], // Just month number
      distance: s.total_distance,
      count: s.activity_count,
      time: s.total_time,
    })) ?? []

  return (
    <WidgetWrapper
      title="Monthly Activity"
      action={
        <div className="flex gap-1">
          {availableYears.slice(0, 3).map((y) => (
            <Button
              key={y}
              variant={year === y ? 'secondary' : 'ghost'}
              size="sm"
              className="h-6 px-2 text-xs"
              onClick={() => setYear(y)}
            >
              {y}
            </Button>
          ))}
        </div>
      }
    >
      <div className="h-full flex flex-col">
        <div className="mb-3 flex gap-1 flex-shrink-0">
          <Button
            variant={metric === 'distance' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-6 px-2 text-xs"
            onClick={() => setMetric('distance')}
          >
            Distance
          </Button>
          <Button
            variant={metric === 'count' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-6 px-2 text-xs"
            onClick={() => setMetric('count')}
          >
            Activities
          </Button>
          <Button
            variant={metric === 'time' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-6 px-2 text-xs"
            onClick={() => setMetric('time')}
          >
            Time
          </Button>
        </div>
        <div className="flex-1 min-h-0">
          <MonthlyStatsChart data={chartData} metric={metric} height="100%" loading={isLoading} />
        </div>
      </div>
    </WidgetWrapper>
  )
}
