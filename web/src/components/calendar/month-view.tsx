import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import {
  startOfMonth,
  endOfMonth,
  startOfWeek,
  endOfWeek,
  eachDayOfInterval,
  format,
  isSameMonth,
  isToday,
} from 'date-fns'
import type { CalendarActivity } from '@/lib/api'
import { cn } from '@/lib/utils'
import { getSportColor } from '@/lib/sport-types'
import { formatDistance, formatDuration, formatElevation } from '@/lib/format'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

interface MonthViewProps {
  year: number
  month: number
  activities: CalendarActivity[]
}

const WEEKDAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']

export function MonthView({ year, month, activities }: MonthViewProps) {
  const [selectedDate, setSelectedDate] = useState<string | null>(null)

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
          <div key={day} className="p-2 text-center text-sm font-medium text-muted-foreground">
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
              role="button"
              tabIndex={0}
              onClick={() => setSelectedDate(dateKey)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') setSelectedDate(dateKey)
              }}
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
                    onClick={(e) => e.stopPropagation()}
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

      {selectedDate && (
        <div className="fixed inset-0 z-50 bg-background/80 backdrop-blur-sm">
          <div className="mx-auto mt-10 w-[calc(100%-2rem)] max-w-2xl rounded-lg border border-border bg-background shadow-lg">
            <div className="flex items-center justify-between border-b border-border px-4 py-3">
              <div className="font-semibold">{selectedDate}</div>
              <Button variant="outline" size="sm" onClick={() => setSelectedDate(null)}>
                Close
              </Button>
            </div>
            <div className="p-4 space-y-4">
              <DaySummary
                dateKey={selectedDate}
                activities={activitiesByDate.get(selectedDate) || []}
              />
              <Card>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">Activities</CardTitle>
                </CardHeader>
                <CardContent>
                  {(activitiesByDate.get(selectedDate) || []).length === 0 ? (
                    <p className="text-sm text-muted-foreground">No activities</p>
                  ) : (
                    <div className="space-y-2">
                      {(activitiesByDate.get(selectedDate) || []).map((a) => (
                        <div
                          key={a.id}
                          className="flex items-center justify-between gap-4 rounded-md border border-border p-3"
                        >
                          <div className="min-w-0">
                            <div className="truncate font-medium">
                              <Link
                                to="/activities/$activityId"
                                params={{ activityId: String(a.id) }}
                                className="hover:underline"
                              >
                                {a.name}
                              </Link>
                            </div>
                            <div className="text-xs text-muted-foreground">{a.sport_type}</div>
                          </div>
                          <div className="flex shrink-0 items-center gap-4 text-sm">
                            <div className="text-right">
                              <div className="font-medium">{formatDistance(a.distance)}</div>
                              <div className="text-xs text-muted-foreground">Distance</div>
                            </div>
                            <div className="text-right">
                              <div className="font-medium">{formatDuration(a.moving_time)}</div>
                              <div className="text-xs text-muted-foreground">Time</div>
                            </div>
                            <div className="text-right">
                              <div className="font-medium">
                                {formatElevation(a.total_elevation_gain)}
                              </div>
                              <div className="text-xs text-muted-foreground">Elev</div>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function DaySummary({ dateKey, activities }: { dateKey: string; activities: CalendarActivity[] }) {
  const totals = useMemo(() => {
    let distance = 0
    let moving = 0
    let elev = 0
    for (const a of activities) {
      distance += a.distance || 0
      moving += a.moving_time || 0
      elev += a.total_elevation_gain || 0
    }
    return { distance, moving, elev }
  }, [activities])

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base">Totals</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid gap-3 sm:grid-cols-3">
          <div>
            <div className="text-xs text-muted-foreground">Distance</div>
            <div className="text-lg font-semibold">{formatDistance(totals.distance)}</div>
          </div>
          <div>
            <div className="text-xs text-muted-foreground">Time</div>
            <div className="text-lg font-semibold">{formatDuration(totals.moving)}</div>
          </div>
          <div>
            <div className="text-xs text-muted-foreground">Elevation</div>
            <div className="text-lg font-semibold">{formatElevation(totals.elev)}</div>
          </div>
        </div>
        <div className="mt-2 text-xs text-muted-foreground">
          {dateKey} • {activities.length} {activities.length === 1 ? 'activity' : 'activities'}
        </div>
      </CardContent>
    </Card>
  )
}
