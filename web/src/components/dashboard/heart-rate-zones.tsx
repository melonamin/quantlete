import { Link } from '@tanstack/react-router'
import { useHrZones, useHrZoneDefinitions } from '@/lib/data'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { StackedBarChart } from '@/components/charts'

const ZONE_COLORS = ['#7dd3fc', '#22c55e', '#eab308', '#f97316', '#ef4444']

function formatDuration(seconds: number) {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

export function HeartRateZones() {
  const { data, isLoading } = useHrZones()
  const { data: defs } = useHrZoneDefinitions()

  const seconds = data?.seconds_by_zone ?? [0, 0, 0, 0, 0]
  const total = data?.total_seconds ?? 0

  const series = seconds.map((s, idx) => ({
    name: `Z${idx + 1}`,
    data: [s],
    color: ZONE_COLORS[idx],
  }))

  return (
    <WidgetWrapper
      title="Heart Rate Zones"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/training-load">View details</Link>
        </Button>
      }
      isLoading={isLoading}
    >
      <div className="h-full flex flex-col gap-3">
        <div className="flex-1 min-h-0">
          <StackedBarChart categories={['']} series={series} horizontal height="100%" />
        </div>
        <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground flex-shrink-0">
          {seconds.map((s, idx) => {
            const pct = total > 0 ? Math.round((s / total) * 100) : 0
            return (
              <div
                key={idx}
                className="flex items-center justify-between rounded-md border border-border px-2 py-1"
              >
                <span className="flex items-center gap-2">
                  <span
                    className="inline-block h-2 w-2 rounded-full"
                    style={{ backgroundColor: ZONE_COLORS[idx] }}
                  />
                  Z{idx + 1}
                </span>
                <span>
                  {formatDuration(s)} ({pct}%)
                </span>
              </div>
            )
          })}
        </div>
        {defs && defs.length === 0 && (
          <p className="text-xs text-muted-foreground flex-shrink-0">
            Set zone definitions in Settings (Zones section).
          </p>
        )}
      </div>
    </WidgetWrapper>
  )
}
