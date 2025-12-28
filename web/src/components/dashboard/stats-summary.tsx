import type { DashboardStats } from '@/lib/api/dashboard'
import { formatDistance, formatDuration, formatElevation } from '@/lib/format'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'

interface StatsSummaryProps {
  stats: DashboardStats | undefined
  isLoading: boolean
}

export function StatsSummary({ stats, isLoading }: StatsSummaryProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <Skeleton className="h-4 w-24" />
            </CardHeader>
            <CardContent>
              <Skeleton className="h-8 w-20 mb-1" />
              <Skeleton className="h-3 w-32" />
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  if (!stats) return null

  const cards = [
    {
      title: 'Total Activities',
      value: stats.total_activities.toLocaleString(),
      description: `${stats.year_activities} this year`,
    },
    {
      title: 'Total Distance',
      value: formatDistance(stats.total_distance),
      description: `${formatDistance(stats.year_distance)} this year`,
    },
    {
      title: 'Total Time',
      value: formatDuration(stats.total_moving_time),
      description: `${formatDuration(stats.year_moving_time)} this year`,
    },
    {
      title: 'Total Elevation',
      value: formatElevation(stats.total_elevation_gain),
      description: `${formatElevation(stats.year_elevation_gain)} this year`,
    },
  ]

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              {card.title}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{card.value}</div>
            <p className="text-xs text-muted-foreground">{card.description}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
