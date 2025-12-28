import { useMemo } from 'react'
import { Link } from '@tanstack/react-router'
import {
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  eachDayOfInterval,
  format,
  isSameMonth,
  isSameDay,
  isToday,
} from 'date-fns'
import type { CalendarActivity } from '@/lib/api'
import { cn } from '@/lib/utils'
import { getSportColor, formatSportType } from '@/lib/sport-types'
import { formatDistance, formatDuration } from '@/lib/format'

interface MonthViewProps {
  year: number
  month: number
  activities: CalendarActivity[]
}

const WEEKDAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

export function MonthView({ year, month, activities }: MonthViewProps) {
  const { days, activitiesByDate } = useMemo(() => {
    const date = new Date(year, month - 1, 1)
    const monthStart = startOfMonth(date)
    const monthEnd = endOfMonth(date)
    const calendarStart = startOfWeek(monthStart, { weekStartsOn: 1 })
    const calendarEnd = endOfWeek(monthEnd, { weekStartsOn: 1 })

    const allDays = eachDayOfInterval({ start: calendarStart, end: calendarEnd })

    const byDate = new Map<string, CalendarActivity[]>()
    for (const activity of activities) {
      const dateKey = format(new Date(activity.start_date), 'yyyy-MM-dd')
      const existing = byDate.get(dateKey) || []
      existing.push(activity)
      byDate.set(dateKey, existing)
    }

    return { days: allDays, activitiesByDate: byDate }
  }, [year, month, activities])

  return (
    <div className="rounded-lg border border-border bg-card">
      {/* Weekday headers */}
      <div className="grid grid-cols-7 border-b border-border">
        {WEEKDAYS.map((day) => (
          <div
            key={day}
            className="p-2 text-center text-sm font-medium text-muted-foreground"
          >
            {day}
          </div>
        ))}
      </div>

      {/* Calendar grid */}
      <div className="grid grid-cols-7">
        {days.map((day) => {
          const dateKey = format(day, 'yyyy-MM-dd')
          const dayActivities = activitiesByDate.get(dateKey) || []
          const isCurrentMonth = isSameMonth(day, new Date(year, month - 1, 1))
          const isCurrentDay = isToday(day)

          return (
            <div
              key={dateKey}
              className={cn(
                'min-h-[100px] border-b border-r border-border p-1',
                !isCurrentMonth && 'bg-muted/30',
                isCurrentDay && 'bg-accent/20'
              )}
            >
              <div
                className={cn(
                  'text-right text-sm mb-1',
                  !isCurrentMonth && 'text-muted-foreground',
                  isCurrentDay && 'font-bold text-strava'
                )}
              >
                {format(day, 'd')}
              </div>
              <div className="space-y-1">
                {dayActivities.slice(0, 3).map((activity) => (
                  <Link
                    key={activity.id}
                    to="/activities/$activityId"
                    params={{ activityId: String(activity.id) }}
                    className={cn(
                      'block rounded px-1 py-0.5 text-xs truncate text-white hover:opacity-80 transition-opacity',
                      getSportColor(activity.sport_type)
                    )}
                    title={`${activity.name} - ${formatDistance(activity.distance)} - ${formatDuration(activity.moving_time)}`}
                  >
                    {activity.name}
                  </Link>
                ))}
                {dayActivities.length > 3 && (
                  <div className="text-xs text-muted-foreground px-1">
                    +{dayActivities.length - 3} more
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
