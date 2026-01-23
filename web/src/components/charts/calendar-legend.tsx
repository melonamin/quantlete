import { cn } from '@/lib/utils'
import type { CalendarMetric } from './activity-charts'
import { calendarPalettes } from './chart-constants'

interface CalendarLegendProps {
  metric: CalendarMetric
  maxValue: number
  className?: string
}

// Labels for each metric type
const metricLabels: Record<CalendarMetric, { unit: string; format: (v: number) => string }> = {
  count: { unit: '', format: (v) => `${Math.round(v)}` },
  distance: { unit: 'km', format: (v) => `${Math.round(v)}` },
  time: { unit: 'min', format: (v) => `${Math.round(v)}` },
  calories: { unit: 'kcal', format: (v) => `${Math.round(v)}` },
  intensity: { unit: '', format: (v) => `${Math.round(v)}` },
}

export function CalendarLegend({ metric, maxValue, className }: CalendarLegendProps) {
  const palette = calendarPalettes[metric]
  const { unit, format } = metricLabels[metric]

  // For single-color gradients, show min/max
  // For multi-color intensity gradient, show labeled scale
  const isIntensity = metric === 'intensity'

  if (isIntensity) {
    // Intensity legend shows labeled color scale
    const intensityLabels = ['Rest', 'Easy', 'Moderate', 'Hard', 'Max']
    return (
      <div className={cn('flex items-center gap-1 text-[9px] text-muted-foreground', className)}>
        <span className="mr-1">Intensity:</span>
        {palette.map((color, i) => (
          <div key={i} className="flex items-center gap-0.5">
            <div
              className="w-2.5 h-2.5 rounded-sm"
              style={{ backgroundColor: color }}
              title={intensityLabels[i]}
            />
            {i === 0 && <span className="text-[8px]">Low</span>}
            {i === palette.length - 1 && <span className="text-[8px]">High</span>}
          </div>
        ))}
      </div>
    )
  }

  // Standard legend shows gradient bar with min/max values
  return (
    <div className={cn('flex items-center gap-1.5 text-[9px] text-muted-foreground', className)}>
      <span>0</span>
      <div
        className="h-2 w-16 rounded-sm"
        style={{
          background: `linear-gradient(to right, ${palette.join(', ')})`,
        }}
      />
      <span>
        {format(maxValue)}
        {unit && ` ${unit}`}
      </span>
    </div>
  )
}
