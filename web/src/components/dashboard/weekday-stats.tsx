import { DonutChart } from '@/components/charts'
import { useWeekdayDistribution } from '@/lib/data'
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
      <DonutChart data={slices} height="100%" loading={isLoading} />
    </WidgetWrapper>
  )
}
