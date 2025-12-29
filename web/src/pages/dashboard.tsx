import { Link } from '@tanstack/react-router'
import { useAuthStatus, useDashboard } from '@/lib/data'
import { isWasmMode } from '@/lib/mode'
import { shouldShowOnboarding } from '@/components/onboarding/welcome-modal'
import {
  StatsSummary,
  RecentActivities,
  WeeklyStats,
  SportBreakdown,
  MonthlyChart,
  SportChart,
  ActivityCalendar,
  TrainingGoals,
  WidgetGrid,
  PeakPowerOutputs,
  HeartRateZones,
  TrainingLoad,
  DaytimeStats,
  WeekdayStats,
  RecentChallenges,
  ChallengeConsistency,
  EddingtonWidget,
  YearlyStats,
  DistanceBreakdown,
  ChallengeConsistencyGrid,
  ZwiftStats,
  IntroText,
  GoalRecommendations,
  ActivityInsights,
} from '@/components/dashboard'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getAuthUrl as getStravaOAuthUrl } from '@/lib/wasm/strava/client'

function getStravaAuthUrl(): string {
  if (isWasmMode()) {
    const redirectUri = `${window.location.origin}/oauth/callback`
    return getStravaOAuthUrl(redirectUri)
  }
  // Server mode - use server-side OAuth
  return '/api/v1/auth/strava'
}

export function DashboardPage() {
  const { data: authStatus, isLoading: authLoading } = useAuthStatus()
  const { data: dashboard, isLoading, error } = useDashboard()

  const isAuthenticated = authStatus?.authenticated ?? false

  // In WASM mode, if onboarding modal is showing, don't show the Get Started card
  // (the modal handles the login flow)
  const onboardingShowing = shouldShowOnboarding(isAuthenticated, authLoading)

  // Show getting started if not authenticated and onboarding isn't showing
  if (!authLoading && !isAuthenticated && !onboardingShowing) {
    const stravaAuthUrl = getStravaAuthUrl()

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
              <a href={stravaAuthUrl}>Connect Strava</a>
            </Button>
            {isWasmMode() && (
              <p className="text-xs text-muted-foreground">
                Running in browser-only mode. Your data will be stored locally.
              </p>
            )}
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

      <WidgetGrid
        widgets={[
          {
            id: 'recent_activities',
            title: 'Recent Activities',
            defaultWidth: 8,
            render: () => (
              <RecentActivities activities={dashboard?.recent_activities} isLoading={isLoading} />
            ),
          },
          {
            id: 'weekly_stats',
            title: 'Weekly Stats',
            defaultWidth: 4,
            render: () => <WeeklyStats stats={dashboard?.weekly_stats} isLoading={isLoading} />,
          },
          {
            id: 'monthly_chart',
            title: 'Monthly Chart',
            defaultWidth: 8,
            defaultHeight: 2,
            render: () => <MonthlyChart />,
          },
          {
            id: 'sport_breakdown',
            title: 'Sport Breakdown',
            defaultWidth: 4,
            render: () => (
              <SportBreakdown stats={dashboard?.sport_type_stats} isLoading={isLoading} />
            ),
          },
          {
            id: 'sport_chart',
            title: 'Sport Distribution',
            defaultWidth: 4,
            defaultHeight: 2,
            render: () => <SportChart />,
          },
          {
            id: 'training_goals',
            title: 'Training Goals',
            defaultWidth: 4,
            render: () => <TrainingGoals />,
          },
          {
            id: 'activity_calendar',
            title: 'Activity Calendar',
            defaultWidth: 12,
            defaultHeight: 3,
            render: () => <ActivityCalendar />,
          },
          {
            id: 'peak_power_outputs',
            title: 'Peak Power Outputs',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <PeakPowerOutputs />,
          },
          {
            id: 'heart_rate_zones',
            title: 'Heart Rate Zones',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <HeartRateZones />,
          },
          {
            id: 'training_load',
            title: 'Training Load',
            defaultWidth: 8,
            defaultHeight: 3,
            defaultHidden: true,
            render: () => <TrainingLoad />,
          },
          {
            id: 'daytime_stats',
            title: 'Time of Day',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <DaytimeStats />,
          },
          {
            id: 'weekday_stats',
            title: 'Weekday',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <WeekdayStats />,
          },
          {
            id: 'recent_challenges',
            title: 'Recent Challenges',
            defaultWidth: 4,
            defaultHeight: 1,
            defaultHidden: true,
            render: () => <RecentChallenges />,
          },
          {
            id: 'challenge_consistency',
            title: 'Challenge Consistency',
            defaultWidth: 8,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <ChallengeConsistency />,
          },
          {
            id: 'eddington',
            title: 'Eddington',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <EddingtonWidget />,
          },
          {
            id: 'yearly_stats',
            title: 'Yearly Comparison',
            defaultWidth: 8,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <YearlyStats />,
          },
          {
            id: 'distance_breakdown',
            title: 'Distance Breakdown',
            defaultWidth: 6,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <DistanceBreakdown />,
          },
          {
            id: 'challenge_consistency_grid',
            title: 'Goal Consistency Grid',
            defaultWidth: 12,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <ChallengeConsistencyGrid />,
          },
          {
            id: 'zwift_stats',
            title: 'Virtual Riding',
            defaultWidth: 6,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <ZwiftStats />,
          },
          {
            id: 'intro_text',
            title: 'Welcome',
            defaultWidth: 4,
            defaultHeight: 1,
            defaultHidden: true,
            render: () => <IntroText />,
          },
          {
            id: 'goal_recommendations',
            title: 'Goal Suggestions',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <GoalRecommendations />,
          },
          {
            id: 'activity_insights',
            title: 'Insights',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <ActivityInsights />,
          },
        ]}
      />
    </div>
  )
}
