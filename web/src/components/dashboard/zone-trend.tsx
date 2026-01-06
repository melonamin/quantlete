import { useZoneTrend } from '@/lib/data'
import { WidgetWrapper } from './widget-wrapper'
import { ZoneTrendChart } from '@/components/charts'

export function ZoneTrend() {
  const { data, isLoading } = useZoneTrend(52)

  return (
    <WidgetWrapper title="HR Zone Trend" isLoading={isLoading}>
      <div className="h-full min-h-0">
        <ZoneTrendChart data={data?.weeks ?? []} height="100%" loading={isLoading} />
      </div>
    </WidgetWrapper>
  )
}
