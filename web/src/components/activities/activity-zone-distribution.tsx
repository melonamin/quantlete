import type { HRZonesOutput, ZoneItem } from '@/lib/api/activities'
import { formatDuration } from '@/lib/format'
import { cn } from '@/lib/utils'

interface ActivityZoneDistributionProps {
  hrZones: HRZonesOutput
  className?: string
}

// Zone colors following the standard 5-zone HR model
const zoneColors = [
  'bg-gray-400', // Zone 1 - Recovery
  'bg-blue-500', // Zone 2 - Aerobic
  'bg-green-500', // Zone 3 - Tempo
  'bg-yellow-500', // Zone 4 - Threshold
  'bg-red-500', // Zone 5 - VO2 Max
]

export function ActivityZoneDistribution({ hrZones, className }: ActivityZoneDistributionProps) {
  // Find the zone with most time for bar scaling
  const maxPercentage = Math.max(...hrZones.zones.map((z) => z.percentage))

  return (
    <div className={cn('rounded-lg border border-border bg-card p-4', className)}>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-medium">Heart Rate Zones</h3>
        <div className="text-xs text-muted-foreground">
          Total: {formatDuration(hrZones.total_seconds)}
        </div>
      </div>

      <div className="space-y-3">
        {hrZones.zones.map((zone, index) => (
          <ZoneRow
            key={zone.zone}
            zone={zone}
            color={zoneColors[index] || zoneColors[0]}
            maxPercentage={maxPercentage}
          />
        ))}
      </div>

      <div className="mt-4 flex justify-between border-t border-border pt-3 text-xs text-muted-foreground">
        <div>
          <span className="text-foreground">Avg HR:</span>{' '}
          <span className="font-medium">{Math.round(hrZones.avg_hr)} bpm</span>
        </div>
        <div>
          <span className="text-foreground">Max HR:</span>{' '}
          <span className="font-medium">{Math.round(hrZones.max_hr)} bpm</span>
        </div>
      </div>
    </div>
  )
}

interface ZoneRowProps {
  zone: ZoneItem
  color: string
  maxPercentage: number
}

function ZoneRow({ zone, color, maxPercentage }: ZoneRowProps) {
  // Scale bar width relative to the zone with most time
  const barWidth = maxPercentage > 0 ? (zone.percentage / maxPercentage) * 100 : 0

  return (
    <div className="flex items-center gap-3">
      <div className="flex w-20 items-center gap-2">
        <div className={cn('h-3 w-3 rounded-sm', color)} />
        <span className="text-xs text-muted-foreground">Z{zone.zone}</span>
      </div>

      <div className="relative h-5 flex-1 rounded bg-muted">
        <div
          className={cn('absolute inset-y-0 left-0 rounded', color)}
          style={{ width: `${barWidth}%` }}
        />
        {zone.percentage >= 10 && (
          <span className="absolute inset-y-0 left-2 flex items-center text-xs font-medium text-white">
            {zone.percentage.toFixed(0)}%
          </span>
        )}
      </div>

      <div className="flex w-32 items-center justify-end gap-2 text-xs">
        <span className="text-muted-foreground">{formatDuration(zone.seconds)}</span>
        <span className="w-12 text-right">
          {Math.round(zone.min_bpm)}-{Math.round(zone.max_bpm)}
        </span>
      </div>
    </div>
  )
}
