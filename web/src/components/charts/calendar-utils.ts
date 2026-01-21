import type { CalendarMetric } from './activity-charts'

interface CalendarData {
  date: string
  count: number
  distance?: number
  time?: number
  calories?: number
  intensity?: number
}

// Calculate max value for calendar legend display
export function getCalendarMaxValue(data: CalendarData[], metric: CalendarMetric): number {
  const getValue = (d: CalendarData): number => {
    switch (metric) {
      case 'count':
        return d.count
      case 'distance':
        return (d.distance ?? 0) / 1000
      case 'time':
        return (d.time ?? 0) / 60
      case 'calories':
        return d.calories ?? 0
      case 'intensity':
        return d.intensity ?? 0
    }
  }
  return Math.max(...data.map((d) => getValue(d)), 1)
}
