import { useState } from 'react'
import { useSportTypeStats } from '@/lib/data'
import { SportDistributionChart } from '@/components/charts/activity-charts'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'

type Metric = 'count' | 'distance'

export function SportChart() {
  const [metric, setMetric] = useState<Metric>('count')
  const { data: sportStats, isLoading } = useSportTypeStats()

  const chartData =
    sportStats?.map((s) => ({
      sport_type: s.sport_type,
      count: s.activity_count,
      distance: s.total_distance,
    })) ?? []

  return (
    <WidgetWrapper
      title="Sport Distribution"
      action={
        <div className="flex gap-1">
          <Button
            variant={metric === 'count' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-6 px-2 text-xs"
            onClick={() => setMetric('count')}
          >
            By Count
          </Button>
          <Button
            variant={metric === 'distance' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-6 px-2 text-xs"
            onClick={() => setMetric('distance')}
          >
            By Distance
          </Button>
        </div>
      }
    >
      <SportDistributionChart data={chartData} metric={metric} height="100%" loading={isLoading} />
    </WidgetWrapper>
  )
}
