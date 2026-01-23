import type { SplitsOutput, SplitItem } from '@/lib/api/activities'
import { formatDuration } from '@/lib/format'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { cn } from '@/lib/utils'
import { TrendingUp, TrendingDown, Minus } from 'lucide-react'

interface ActivitySplitsProps {
  splits: SplitsOutput
  className?: string
}

export function ActivitySplits({ splits, className }: ActivitySplitsProps) {
  const { formatDistance } = useFormattedMetrics()

  // Calculate min/max pace for visualization scale
  const paces = splits.splits.map((s) => s.pace_sec_km)
  const minPace = Math.min(...paces)
  const maxPace = Math.max(...paces)
  const paceRange = maxPace - minPace || 1

  // Calculate average pace
  const avgPace = paces.reduce((a, b) => a + b, 0) / paces.length

  const getPaceBarWidth = (pace: number) => {
    // Invert the scale: faster (lower) pace = longer bar
    const normalized = 1 - (pace - minPace) / paceRange
    return 20 + normalized * 80 // 20-100% range
  }

  const getPaceColor = (_pace: number, index: number) => {
    if (index + 1 === splits.fastest_split) return 'bg-green-500'
    if (index + 1 === splits.slowest_split) return 'bg-orange-500'
    return 'bg-primary'
  }

  const getPaceIndicator = (pace: number) => {
    const diff = pace - avgPace
    const threshold = avgPace * 0.02 // 2% threshold

    if (diff < -threshold) return <TrendingUp className="h-3 w-3 text-green-500" />
    if (diff > threshold) return <TrendingDown className="h-3 w-3 text-orange-500" />
    return <Minus className="h-3 w-3 text-muted-foreground" />
  }

  return (
    <div className={cn('rounded-lg border border-border bg-card p-4', className)}>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-medium">
          Splits ({splits.total_splits} × {formatDistance(splits.split_length_m)})
        </h3>
        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          <span className="flex items-center gap-1">
            <div className="h-2 w-2 rounded-full bg-green-500" />
            Fastest
          </span>
          <span className="flex items-center gap-1">
            <div className="h-2 w-2 rounded-full bg-orange-500" />
            Slowest
          </span>
        </div>
      </div>

      <div className="space-y-1">
        {splits.splits.map((split, index) => (
          <SplitRow
            key={split.index}
            split={split}
            barWidth={getPaceBarWidth(split.pace_sec_km)}
            barColor={getPaceColor(split.pace_sec_km, index)}
            paceIndicator={getPaceIndicator(split.pace_sec_km)}
            isPartial={split.distance_m < splits.split_length_m * 0.95}
          />
        ))}
      </div>

      <div className="mt-4 border-t border-border pt-3 text-xs text-muted-foreground">
        <div className="flex justify-between">
          <span>Average Pace</span>
          <span className="font-medium text-foreground">{formatPaceFromSeconds(avgPace)}</span>
        </div>
      </div>
    </div>
  )
}

interface SplitRowProps {
  split: SplitItem
  barWidth: number
  barColor: string
  paceIndicator: React.ReactNode
  isPartial: boolean
}

function SplitRow({ split, barWidth, barColor, paceIndicator, isPartial }: SplitRowProps) {
  const { formatElevation } = useFormattedMetrics()
  const elevChange = split.elev_gain - split.elev_loss

  return (
    <div className="flex items-center gap-2 py-1 text-sm">
      <span className="w-6 text-right text-xs text-muted-foreground">{split.index}</span>

      <div className="relative h-6 flex-1">
        <div
          className={cn('absolute inset-y-0 left-0 rounded', barColor, isPartial && 'opacity-60')}
          style={{ width: `${barWidth}%` }}
        />
        <div className="absolute inset-y-0 left-2 flex items-center gap-2">{paceIndicator}</div>
      </div>

      <div className="flex w-36 items-center justify-end gap-2 tabular-nums">
        <span className="font-medium">{formatPaceFromSeconds(split.pace_sec_km)}</span>
        <span className="text-xs text-muted-foreground">({formatDuration(split.duration_s)})</span>
      </div>

      {split.avg_hr && split.avg_hr > 0 && (
        <span className="w-16 text-right text-xs text-muted-foreground">
          {Math.round(split.avg_hr)} bpm
        </span>
      )}

      {split.avg_watts && split.avg_watts > 0 && (
        <span className="w-14 text-right text-xs text-muted-foreground">
          {Math.round(split.avg_watts)} W
        </span>
      )}

      {(split.elev_gain > 0 || split.elev_loss > 0) && (
        <span
          className={cn(
            'w-16 text-right text-xs',
            elevChange > 5
              ? 'text-green-600'
              : elevChange < -5
                ? 'text-orange-600'
                : 'text-muted-foreground'
          )}
        >
          {elevChange > 0 ? '+' : ''}
          {formatElevation(elevChange)}
        </span>
      )}
    </div>
  )
}

function formatPaceFromSeconds(secondsPerKm: number): string {
  if (!secondsPerKm || secondsPerKm <= 0 || !isFinite(secondsPerKm)) return '-'
  const minutes = Math.floor(secondsPerKm / 60)
  const seconds = Math.round(secondsPerKm % 60)
  return `${minutes}:${seconds.toString().padStart(2, '0')}/km`
}
