import { Link } from '@tanstack/react-router'
import { useDashboard, useAuthStatus } from '@/lib/api'
import {
  StatsSummary,
  RecentActivities,
  WeeklyStats,
  SportBreakdown,
} from '@/components/dashboard'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function DashboardPage() {
  const { data: authStatus } = useAuthStatus()
  const { data: dashboard, isLoading, error } = useDashboard()

  const isAuthenticated = authStatus?.authenticated

  // Show getting started if not authenticated
  if (!isAuthenticated) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-2xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground">Your activity statistics at a glance</p>
        </div>

        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>Get Started</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Connect your Strava account to import your activities and view your statistics.
            </p>
            <Button asChild className="bg-strava hover:bg-strava/90">
              <a href="/api/v1/auth/strava">Connect Strava</a>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Show error state
  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-2xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground">Your activity statistics at a glance</p>
        </div>

        <Card>
          <CardContent className="py-8 text-center">
            <p className="text-destructive mb-4">Failed to load dashboard data</p>
            <Button onClick={() => window.location.reload()}>Retry</Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Show empty state if no activities
  const hasActivities = dashboard?.stats && dashboard.stats.total_activities > 0

  if (!isLoading && !hasActivities) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-2xl font-bold">Dashboard</h1>
          <p className="text-muted-foreground">Your activity statistics at a glance</p>
        </div>

        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>No Activities Yet</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Import your activities from Strava to see your statistics.
            </p>
            <Button asChild>
              <Link to="/settings">Go to Settings</Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">Your activity statistics at a glance</p>
      </div>

      {/* Stats overview */}
      <div className="mb-8">
        <StatsSummary stats={dashboard?.stats} isLoading={isLoading} />
      </div>

      {/* Widgets grid */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {/* Recent Activities - spans 2 columns on large screens */}
        <div className="lg:col-span-2">
          <RecentActivities
            activities={dashboard?.recent_activities}
            isLoading={isLoading}
          />
        </div>

        {/* Weekly Stats */}
        <WeeklyStats
          stats={dashboard?.weekly_stats}
          isLoading={isLoading}
        />

        {/* Sport Breakdown */}
        <SportBreakdown
          stats={dashboard?.sport_type_stats}
          isLoading={isLoading}
        />
      </div>
    </div>
  )
}
