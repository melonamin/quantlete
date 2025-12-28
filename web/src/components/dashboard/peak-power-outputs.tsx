import { Link } from '@tanstack/react-router'
import { usePowerStats } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { BarChart } from '@/components/charts'
import { Button } from '@/components/ui/button'

function formatDurationLabel(seconds: number) {
  if (seconds < 60) return `${seconds}s`
  if (seconds % 60 === 0 && seconds < 3600) return `${seconds / 60}m`
  if (seconds === 3600) return '1h'
  return `${Math.round(seconds / 60)}m`
}

export function PeakPowerOutputs() {
  const { data, isLoading } = usePowerStats()

  const bestByDuration = new Map<number, number>()
  for (const b of data?.best ?? []) bestByDuration.set(b.duration_s, b.watts)

  const chartData =
    data?.durations_s?.map((d) => ({
      label: formatDurationLabel(d),
      value: Math.round(bestByDuration.get(d) ?? 0),
    })) ?? []

  return (
    <WidgetWrapper
      title="Peak Power Outputs"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/power">View details</Link>
        </Button>
      }
      isLoading={isLoading}
    >
      <BarChart data={chartData} height="100%" showValues />
    </WidgetWrapper>
  )
}
