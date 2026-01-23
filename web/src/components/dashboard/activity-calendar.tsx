import { useMemo, useState } from 'react'
import { useCalendarData, useCalendarDataRange, useYearlyStats } from '@/lib/api'
import {
  ActivityCalendarChart,
  type CalendarMetric,
  type CalendarRange,
} from '@/components/charts/activity-charts'
import { getCalendarMaxValue } from '@/components/charts/calendar-utils'
import { CalendarLegend } from '@/components/charts/calendar-legend'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'

const METRIC_OPTIONS: { value: CalendarMetric; label: string }[] = [
  { value: 'count', label: 'Count' },
  { value: 'distance', label: 'Distance' },
  { value: 'time', label: 'Time' },
  { value: 'calories', label: 'Calories' },
  { value: 'intensity', label: 'Intensity' },
]

const RANGE_OPTIONS: { value: CalendarRange; label: string }[] = [
  { value: 'year', label: 'Year' },
  { value: 'rolling365', label: '365 Days' },
]

const formatLocalDate = (date: Date) => {
  const year = date.getFullYear()
  const month = `${date.getMonth() + 1}`.padStart(2, '0')
  const day = `${date.getDate()}`.padStart(2, '0')
  return `${year}-${month}-${day}`
}

export function ActivityCalendar() {
  const currentYear = new Date().getFullYear()
  const [year, setYear] = useState(currentYear)
  const [metric, setMetric] = useState<CalendarMetric>('count')
  const [rangeType, setRangeType] = useState<CalendarRange>('year')

  const { data: yearlyStats } = useYearlyStats()
  const { data: yearCalendarData, isLoading: yearLoading } = useCalendarData(year)

  // Calculate rolling 365 date range
  const rolling365Range = useMemo(() => {
    const now = new Date()
    const startDateObj = new Date(now)
    startDateObj.setDate(startDateObj.getDate() - 365)
    const endDate = formatLocalDate(now)
    const startDate = formatLocalDate(startDateObj)
    return { startDate, endDate }
  }, [])

  const { data: rollingCalendarData, isLoading: rollingLoading } = useCalendarDataRange(
    rolling365Range.startDate,
    rolling365Range.endDate
  )

  const availableYears = yearlyStats?.map((s) => s.year) ?? [currentYear]
  const minYear = Math.min(...availableYears)
  const maxYear = Math.max(...availableYears)

  // Select the appropriate data based on range type
  const calendarData = rangeType === 'year' ? yearCalendarData : rollingCalendarData
  const isLoading = rangeType === 'year' ? yearLoading : rollingLoading

  const chartData = useMemo(
    () =>
      calendarData?.map((d) => ({
        date: d.date,
        count: d.activity_count,
        distance: d.total_distance,
        time: d.total_time,
        calories: d.total_calories,
        intensity: d.total_intensity,
      })) ?? [],
    [calendarData]
  )

  const maxValue = useMemo(() => getCalendarMaxValue(chartData, metric), [chartData, metric])

  return (
    <WidgetWrapper
      title="Activity Calendar"
      action={
        <div className="flex items-center gap-2">
          {/* Range toggle */}
          <div className="flex rounded-sm border border-border overflow-hidden">
            {RANGE_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => setRangeType(opt.value)}
                className={cn(
                  'px-2 py-0.5 text-[10px] font-medium transition-colors',
                  rangeType === opt.value
                    ? 'bg-terminal-cyan/20 text-terminal-cyan'
                    : 'text-muted-foreground hover:text-foreground hover:bg-accent'
                )}
              >
                {opt.label}
              </button>
            ))}
          </div>
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
          {/* Year navigation - only show when in year mode */}
          {rangeType === 'year' && (
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
          )}
        </div>
      }
    >
      <div className="flex flex-col h-full">
        <div className="flex-1 min-h-0">
          <ActivityCalendarChart
            data={chartData}
            year={rangeType === 'year' ? year : undefined}
            rangeType={rangeType}
            dateRange={
              rangeType === 'rolling365'
                ? [rolling365Range.startDate, rolling365Range.endDate]
                : undefined
            }
            metric={metric}
            height="100%"
            loading={isLoading}
          />
        </div>
        <CalendarLegend metric={metric} maxValue={maxValue} className="mt-1 justify-center" />
      </div>
    </WidgetWrapper>
  )
}
