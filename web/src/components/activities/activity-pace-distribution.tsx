import type { PaceDistributionOutput, PaceBucketItem } from '@/lib/api/activities'
import { formatDuration, formatPaceFromSecondsPerKm } from '@/lib/format'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { cn } from '@/lib/utils'

interface ActivityPaceDistributionProps {
  paceDistribution: PaceDistributionOutput
  className?: string
}

export function ActivityPaceDistribution({
  paceDistribution,
  className,
}: ActivityPaceDistributionProps) {
  const { unitSystem } = useFormattedMetrics()

  // Find the bucket with most time for bar scaling
  const maxPercentage = Math.max(...paceDistribution.buckets.map((b) => b.percentage))
  const formatSecondsPerKm = (secondsPerKm: number) =>
    formatPaceFromSecondsPerKm(secondsPerKm, unitSystem)

  return (
    <div className={cn('rounded-lg border border-border bg-card p-4', className)}>
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-sm font-medium">Pace Distribution</h3>
        <div className="text-xs text-muted-foreground">
          Total: {formatDuration(paceDistribution.total_seconds)}
        </div>
      </div>

      <div className="flex h-32 items-end gap-1">
        {paceDistribution.buckets.map((bucket, index) => (
          <PaceBucket
            key={index}
            bucket={bucket}
            maxPercentage={maxPercentage}
            avgPace={paceDistribution.avg_pace}
            formatPaceValue={formatSecondsPerKm}
          />
        ))}
      </div>

      <div className="mt-4 flex justify-between border-t border-border pt-3 text-xs">
        <StatBlock label="Fastest" value={formatSecondsPerKm(paceDistribution.fastest_pace)} />
        <StatBlock label="Average" value={formatSecondsPerKm(paceDistribution.avg_pace)} />
        <StatBlock label="Median" value={formatSecondsPerKm(paceDistribution.median_pace)} />
        <StatBlock label="Slowest" value={formatSecondsPerKm(paceDistribution.slowest_pace)} />
      </div>
    </div>
  )
}

interface PaceBucketProps {
  bucket: PaceBucketItem
  maxPercentage: number
  avgPace: number
  formatPaceValue: (secondsPerKm: number) => string
}

function PaceBucket({ bucket, maxPercentage, avgPace, formatPaceValue }: PaceBucketProps) {
  const barHeight = maxPercentage > 0 ? (bucket.percentage / maxPercentage) * 100 : 0
  const midPace = (bucket.min_pace + bucket.max_pace) / 2

  // Color based on relationship to average pace
  const isNearAvg = Math.abs(midPace - avgPace) < 15 // Within 15 sec/km of average
  const isFaster = midPace < avgPace - 15
  const isSlower = midPace > avgPace + 15

  const barColor = isFaster
    ? 'bg-green-500'
    : isSlower
      ? 'bg-orange-500'
      : isNearAvg
        ? 'bg-primary'
        : 'bg-primary/70'

  return (
    <div className="group relative flex flex-1 flex-col items-center">
      <div
        className={cn('w-full rounded-t transition-all', barColor)}
        style={{ height: `${barHeight}%`, minHeight: bucket.percentage > 0 ? '4px' : '0' }}
      />

      {/* Tooltip on hover */}
      <div className="pointer-events-none absolute bottom-full mb-2 hidden rounded bg-popover px-2 py-1 text-xs shadow-md group-hover:block">
        <div className="whitespace-nowrap font-medium">
          {formatPaceValue(bucket.min_pace)} - {formatPaceValue(bucket.max_pace)}
        </div>
        <div className="text-muted-foreground">
          {bucket.percentage.toFixed(1)}% ({formatDuration(bucket.seconds)})
        </div>
      </div>
    </div>
  )
}

interface StatBlockProps {
  label: string
  value: string
}

function StatBlock({ label, value }: StatBlockProps) {
  return (
    <div className="text-center">
      <div className="text-muted-foreground">{label}</div>
      <div className="font-medium text-foreground">{value}</div>
    </div>
  )
}
