import { useState } from 'react'
import { useCalendarData, useYearlyStats } from '@/lib/api/dashboard'
import { ActivityCalendarChart } from '@/components/charts/activity-charts'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { ChevronLeft, ChevronRight } from 'lucide-react'

export function ActivityCalendar() {
  const currentYear = new Date().getFullYear()
  const [year, setYear] = useState(currentYear)

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
    })) ?? []

  return (
    <WidgetWrapper
      title="Activity Calendar"
      action={
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
      }
    >
      <ActivityCalendarChart data={chartData} year={year} height={160} loading={isLoading} />
    </WidgetWrapper>
  )
}
