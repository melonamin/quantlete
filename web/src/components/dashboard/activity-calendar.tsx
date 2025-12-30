import { useState } from 'react'
import { useCalendarData, useYearlyStats } from '@/lib/api'
import { ActivityCalendarChart, type CalendarMetric } from '@/components/charts/activity-charts'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

const METRIC_OPTIONS: { value: CalendarMetric; label: string }[] = [
  { value: 'count', label: 'Count' },
  { value: 'distance', label: 'Distance' },
  { value: 'time', label: 'Time' },
  { value: 'calories', label: 'Calories' },
]

export function ActivityCalendar() {
  const currentYear = new Date().getFullYear()
  const [year, setYear] = useState(currentYear)
  const [metric, setMetric] = useState<CalendarMetric>('count')

  const { data: yearlyStats } = useYearlyStats()
  const { data: calendarData, isLoading } = useCalendarData(year)

  const availableYears = yearlyStats?.map((s) => s.year) ?? [currentYear]
  const minYear = Math.min(...availableYears)
  const maxYear = Math.max(...availableYears)

  const chartData =
    calendarData?.map((d) => ({
      date: d.date,
      count: d.activity_count,
      distance: d.total_distance,
      time: d.total_time,
      calories: d.total_calories,
    })) ?? []

  return (
    <WidgetWrapper
      title="Activity Calendar"
      action={
        <div className="flex items-center gap-2">
          {/* Metric tabs */}
          <div className="flex rounded-sm border border-border overflow-hidden">
            {METRIC_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => setMetric(opt.value)}
                className={cn(
                  'px-2 py-0.5 text-[10px] font-medium transition-colors',
                  metric === opt.value
                    ? 'bg-terminal-green/20 text-terminal-green'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                )}
              >
                {opt.label}
              </button>
            ))}
          </div>
          {/* Year navigation */}
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0"
              onClick={() => setYear((y) => Math.max(minYear, y - 1))}
              disabled={year <= minYear}
            >
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="text-sm font-medium w-12 text-center">{year}</span>
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0"
              onClick={() => setYear((y) => Math.min(maxYear, y + 1))}
              disabled={year >= maxYear}
            >
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      }
    >
      <ActivityCalendarChart
        data={chartData}
        year={year}
        metric={metric}
        height="100%"
        loading={isLoading}
      />
    </WidgetWrapper>
  )
}
