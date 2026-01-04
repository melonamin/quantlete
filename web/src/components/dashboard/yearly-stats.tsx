import { useYearlyStats } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { formatDistance, formatDuration } from '@/lib/format'
import { ArrowUp, ArrowDown, Minus } from 'lucide-react'
import { cn } from '@/lib/utils'
import { uiColors } from '@/components/charts'

interface YearlyStatWithDelta {
  year: number
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
  deltas: {
    activity_count: number | null
    distance: number | null
    time: number | null
    elevation: number | null
  }
}

function calculateDeltas(
  stats: {
    year: number
    activity_count: number
    total_distance: number
    total_time: number
    total_elevation: number
  }[]
): YearlyStatWithDelta[] {
  const sorted = [...stats].sort((a, b) => b.year - a.year)

  return sorted.map((stat, index) => {
    const prevYear = sorted[index + 1]

    const calcDelta = (current: number, prev: number | undefined): number | null => {
      if (prev === undefined || prev === 0) return null
      return ((current - prev) / prev) * 100
    }

    return {
      ...stat,
      deltas: {
        activity_count: prevYear ? calcDelta(stat.activity_count, prevYear.activity_count) : null,
        distance: prevYear ? calcDelta(stat.total_distance, prevYear.total_distance) : null,
        time: prevYear ? calcDelta(stat.total_time, prevYear.total_time) : null,
        elevation: prevYear ? calcDelta(stat.total_elevation, prevYear.total_elevation) : null,
      },
    }
  })
}

function DeltaIndicator({ value, className }: { value: number | null; className?: string }) {
  if (value === null) {
    return <span className={cn('text-muted-foreground', className)}>—</span>
  }

  const isPositive = value > 0
  const isNeutral = Math.abs(value) < 0.5

  if (isNeutral) {
    return (
      <span className={cn('inline-flex items-center gap-0.5 text-muted-foreground', className)}>
        <Minus className="h-3 w-3" />
        <span className="text-xs">0%</span>
      </span>
    )
  }

  const color = isPositive ? uiColors.success : uiColors.danger

  return (
    <span className={cn('inline-flex items-center gap-0.5', className)} style={{ color }}>
      {isPositive ? <ArrowUp className="h-3 w-3" /> : <ArrowDown className="h-3 w-3" />}
      <span className="text-xs">{Math.abs(value).toFixed(0)}%</span>
    </span>
  )
}

export function YearlyStats() {
  const { data, isLoading } = useYearlyStats()

  const statsWithDeltas = data ? calculateDeltas(data) : []
  // Show last 5 years for compact display
  const displayStats = statsWithDeltas.slice(0, 5)

  return (
    <WidgetWrapper title="Yearly Comparison" isLoading={isLoading}>
      {displayStats.length === 0 ? (
        <div className="text-sm text-muted-foreground">No yearly data available.</div>
      ) : (
        <div className="overflow-x-auto">
          <Table className="text-sm">
            <TableHeader>
              <TableRow className="border-b">
                <TableHead className="py-2 text-left font-medium">Year</TableHead>
                <TableHead className="py-2 text-right font-medium">Activities</TableHead>
                <TableHead className="py-2 text-right font-medium">Distance</TableHead>
                <TableHead className="py-2 text-right font-medium">Time</TableHead>
                <TableHead className="py-2 text-right font-medium">Elevation</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {displayStats.map((stat) => (
                <TableRow key={stat.year} className="border-b border-border/50 last:border-0">
                  <TableCell className="py-2 font-medium">{stat.year}</TableCell>
                  <TableCell className="py-2 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <span>{stat.activity_count}</span>
                      <DeltaIndicator value={stat.deltas.activity_count} />
                    </div>
                  </TableCell>
                  <TableCell className="py-2 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <span>{formatDistance(stat.total_distance)}</span>
                      <DeltaIndicator value={stat.deltas.distance} />
                    </div>
                  </TableCell>
                  <TableCell className="py-2 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <span>{formatDuration(stat.total_time)}</span>
                      <DeltaIndicator value={stat.deltas.time} />
                    </div>
                  </TableCell>
                  <TableCell className="py-2 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <span>{stat.total_elevation.toLocaleString()}m</span>
                      <DeltaIndicator value={stat.deltas.elevation} />
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </WidgetWrapper>
  )
}
