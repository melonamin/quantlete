import { DonutChart } from '@/components/charts'
import { useDaytimeDistribution } from '@/lib/data'
import { WidgetWrapper } from './widget-wrapper'

export function DaytimeStats() {
  const { data, isLoading } = useDaytimeDistribution()

  const slices =
    data?.map((d) => ({
      name: d.label,
      value: d.count,
    })) ?? []

  return (
    <WidgetWrapper title="Time of Day" isLoading={isLoading}>
      <DonutChart data={slices} height="100%" loading={isLoading} />
    </WidgetWrapper>
  )
}
