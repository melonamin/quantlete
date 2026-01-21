import { DonutChart } from '@/components/charts'
import { weekdayColorMap } from '@/components/charts/chart-constants'
import { useWeekdayDistribution } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'

export function WeekdayStats() {
  const { data, isLoading } = useWeekdayDistribution()

  const slices =
    data?.map((d) => ({
      name: d.label,
      value: d.count,
    })) ?? []

  return (
    <WidgetWrapper title="Weekday" isLoading={isLoading}>
      <DonutChart data={slices} height="100%" loading={isLoading} colorMap={weekdayColorMap} />
    </WidgetWrapper>
  )
}
