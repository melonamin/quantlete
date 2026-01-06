import { useHrZones, useHrZoneDefinitions } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { zoneColorsArray, zoneLabels } from '@/components/charts'

function formatDuration(seconds: number) {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}

interface ZoneProgressRowProps {
  label: string
  color: string
  seconds: number
  total: number
}

function ZoneProgressRow({ label, color, seconds, total }: ZoneProgressRowProps) {
  const pct = total > 0 ? (seconds / total) * 100 : 0
  return (
    <div className="flex items-center gap-2 text-xs">
      <span
        className="h-2 w-2 rounded-full flex-shrink-0"
        style={{ backgroundColor: color }}
      />
      <span className="w-20 text-muted-foreground truncate flex-shrink-0">{label}</span>
      <div className="flex-1 h-1.5 bg-muted rounded-full overflow-hidden min-w-0">
        <div
          className="h-full rounded-full"
          style={{ width: `${pct}%`, backgroundColor: color }}
        />
      </div>
      <span className="w-16 text-right flex-shrink-0 tabular-nums whitespace-nowrap">
        {formatDuration(seconds)}
      </span>
      <span className="w-8 text-right text-muted-foreground flex-shrink-0 tabular-nums">
        {Math.round(pct)}%
      </span>
    </div>
  )
}

export function HeartRateZones() {
  const { data, isLoading } = useHrZones()
  const { data: defs } = useHrZoneDefinitions()

  const seconds = data?.seconds_by_zone ?? [0, 0, 0, 0, 0]
  const total = data?.total_seconds ?? 0

  return (
    <WidgetWrapper title="Heart Rate Zones" isLoading={isLoading}>
      <div className="h-full flex flex-col justify-center gap-1.5">
        {seconds.map((s, idx) => (
          <ZoneProgressRow
            key={idx}
            label={zoneLabels[idx]}
            color={zoneColorsArray[idx]}
            seconds={s}
            total={total}
          />
        ))}
        {defs && defs.length === 0 && (
          <p className="text-xs text-muted-foreground">
            Configure zones in Settings.
          </p>
        )}
      </div>
    </WidgetWrapper>
  )
}
