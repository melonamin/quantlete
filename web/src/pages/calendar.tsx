import { useState } from 'react'
import { format, addMonths, subMonths } from 'date-fns'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useCalendarActivities, useCalendarSummary } from '@/lib/api'
import { MonthView } from '@/components/calendar'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent } from '@/components/ui/card'
import { formatDistance, formatDurationLong, formatNumber, formatElevation } from '@/lib/format'

export function CalendarPage() {
  const [currentDate, setCurrentDate] = useState(() => new Date())
  const year = currentDate.getFullYear()
  const month = currentDate.getMonth() + 1

  const { data: activities, isLoading } = useCalendarActivities(year, month)
  const { data: summary } = useCalendarSummary(year, month)

  const handlePrevMonth = () => {
    setCurrentDate((prev) => subMonths(prev, 1))
  }

  const handleNextMonth = () => {
    setCurrentDate((prev) => addMonths(prev, 1))
  }

  const handleToday = () => {
    setCurrentDate(new Date())
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Calendar</h1>
          <p className="text-muted-foreground">View your activities by date</p>
        </div>

        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={handleToday}>
            Today
          </Button>
          <div className="flex items-center gap-1">
            <Button variant="ghost" size="icon" onClick={handlePrevMonth}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <div className="w-32 text-center font-medium">{format(currentDate, 'MMMM yyyy')}</div>
            <Button variant="ghost" size="icon" onClick={handleNextMonth}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {summary && (
        <div className="mb-6 grid gap-3 sm:grid-cols-2 lg:grid-cols-6">
          <Card>
            <CardContent className="p-4">
              <div className="text-xs text-muted-foreground">Distance</div>
              <div className="text-lg font-semibold">{formatDistance(summary.total_distance)}</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-xs text-muted-foreground">Elevation</div>
              <div className="text-lg font-semibold">
                {formatElevation(summary.total_elevation_gain)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-xs text-muted-foreground">Time</div>
              <div className="text-lg font-semibold">
                {formatDurationLong(summary.total_moving_time)}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-xs text-muted-foreground">Workouts</div>
              <div className="text-lg font-semibold">{formatNumber(summary.workout_count)}</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-xs text-muted-foreground">Calories</div>
              <div className="text-lg font-semibold">{formatNumber(summary.total_calories)}</div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="p-4">
              <div className="text-xs text-muted-foreground">Challenges</div>
              <div className="text-lg font-semibold">
                {formatNumber(summary.challenges_completed)}
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {isLoading ? (
        <CalendarSkeleton />
      ) : (
        <MonthView year={year} month={month} activities={activities || []} />
      )}
    </div>
  )
}

function CalendarSkeleton() {
  return (
    <div className="rounded-lg border border-border bg-card">
      <div className="grid grid-cols-7 border-b border-border">
        {Array.from({ length: 7 }).map((_, i) => (
          <div key={i} className="p-2">
            <Skeleton className="h-4 w-full" />
          </div>
        ))}
      </div>
      <div className="grid grid-cols-7">
        {Array.from({ length: 35 }).map((_, i) => (
          <div key={i} className="min-h-[100px] border-b border-r border-border p-1">
            <Skeleton className="h-4 w-6 mb-2" />
          </div>
        ))}
      </div>
    </div>
  )
}
