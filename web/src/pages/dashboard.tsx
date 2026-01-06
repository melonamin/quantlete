import { useEffect, useRef } from 'react'
import { Link } from '@tanstack/react-router'
import {
  useAuthStatus,
  useDashboard,
  useImportProgress,
  useCredentialsStatus,
} from '@/lib/data/hooks'
import { useQueryClient } from '@tanstack/react-query'
import { isWasmMode } from '@/lib/mode'
import { shouldShowOnboarding } from '@/components/onboarding/utils'
import { useOnboardingStore, useSyncModalStore } from '@/stores'
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
  ZoneTrend,
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
  KudosLeaders,
  SmartCoach,
} from '@/components/dashboard'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { getAuthUrl as getStravaOAuthUrl } from '@/lib/wasm/strava/client'

function getStravaAuthUrl(): string | null {
  if (isWasmMode()) {
    try {
      const redirectUri = `${window.location.origin}/oauth/callback`
      return getStravaOAuthUrl(redirectUri)
    } catch {
      // Credentials not configured yet
      return null
    }
  }
  // Server mode - use server-side OAuth
  return '/api/v1/auth/strava'
}

export function DashboardPage() {
  const queryClient = useQueryClient()
  const { data: authStatus, isLoading: authLoading } = useAuthStatus()
  const { data: credentials, isLoading: credentialsLoading } = useCredentialsStatus()
  const { data: dashboard, isLoading, error } = useDashboard()
  const { data: importProgress } = useImportProgress()

  const isAuthenticated = authStatus?.authenticated ?? false
  const credentialsConfigured = credentials?.configured ?? false
  const onboardingDismissed = useOnboardingStore((s) => s.dismissed)
  const syncModalOpen = useSyncModalStore((s) => s.open)

  // Track previous import status to detect completion
  const prevStatusRef = useRef(importProgress?.status)

  // Refresh dashboard data when sync completes
  useEffect(() => {
    const prevStatus = prevStatusRef.current
    const currentStatus = importProgress?.status

    // If status changed from 'running' to 'completed', refresh dashboard
    if (prevStatus === 'running' && currentStatus === 'completed') {
      queryClient.invalidateQueries({ queryKey: ['data', 'dashboard'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'activities'] })
    }

    prevStatusRef.current = currentStatus
  }, [importProgress?.status, queryClient])

  // If onboarding modal is showing, don't show the Get Started card
  // (the modal handles the login flow)
  const onboardingShowing = shouldShowOnboarding(
    isAuthenticated,
    authLoading,
    onboardingDismissed,
    credentialsConfigured,
    credentialsLoading
  )

  // Don't show Get Started if sync modal is open (user is already syncing)
  const syncInProgress = syncModalOpen || importProgress?.status === 'running'

  // Show getting started if not authenticated and onboarding isn't showing and sync isn't in progress
  if (!authLoading && !isAuthenticated && !onboardingShowing && !syncInProgress) {
    const stravaAuthUrl = getStravaAuthUrl()

    // In WASM mode, if credentials aren't configured, show setup message
    if (isWasmMode() && !stravaAuthUrl) {
      return (
        <div className="container mx-auto px-4 py-8">
          <div className="mb-8">
            <h1 className="text-2xl font-bold">Dashboard</h1>
            <p className="text-muted-foreground">Your activity statistics at a glance</p>
          </div>

          <Card className="max-w-md">
            <CardHeader>
              <CardTitle>Setup Required</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-sm text-muted-foreground">
                To get started, you need to configure your Strava API credentials. Go to Settings to
                enter your Client ID and Client Secret from your Strava API Application.
              </p>
              <Button asChild>
                <Link to="/settings">Go to Settings</Link>
              </Button>
              <p className="text-xs text-muted-foreground">
                Running in browser-only mode. Your data will be stored locally.
              </p>
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

        <Card className="max-w-md">
          <CardHeader>
            <CardTitle>Get Started</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Connect your Strava account to import your activities and view your statistics.
            </p>
            {stravaAuthUrl && (
              <Button asChild className="bg-strava hover:bg-strava/90">
                <a href={stravaAuthUrl}>Connect Strava</a>
              </Button>
            )}
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
          // Row 1: Welcome + Weekly progress + Insights
          {
            id: 'intro_text',
            title: 'Welcome',
            defaultWidth: 4,
            defaultHeight: 1,
            render: () => <IntroText />,
          },
          {
            id: 'weekly_stats',
            title: 'Weekly Stats',
            defaultWidth: 4,
            render: () => <WeeklyStats stats={dashboard?.weekly_stats} isLoading={isLoading} />,
          },
          {
            id: 'activity_insights',
            title: 'Insights',
            defaultWidth: 4,
            defaultHeight: 2,
            render: () => <ActivityInsights />,
          },
          {
            id: 'smart_coach',
            title: 'Smart Coach',
            defaultWidth: 4,
            defaultHeight: 2,
            render: () => <SmartCoach />,
          },
          // Row 2: Recent activities + Sport breakdown
          {
            id: 'recent_activities',
            title: 'Recent Activities',
            defaultWidth: 8,
            render: () => (
              <RecentActivities activities={dashboard?.recent_activities} isLoading={isLoading} />
            ),
          },
          {
            id: 'sport_breakdown',
            title: 'Sport Breakdown',
            defaultWidth: 4,
            render: () => (
              <SportBreakdown stats={dashboard?.sport_type_stats} isLoading={isLoading} />
            ),
          },
          // Row 3: Monthly chart + Goals
          {
            id: 'monthly_chart',
            title: 'Monthly Chart',
            defaultWidth: 8,
            defaultHeight: 2,
            render: () => <MonthlyChart />,
          },
          {
            id: 'training_goals',
            title: 'Training Goals',
            defaultWidth: 4,
            render: () => <TrainingGoals />,
          },
          // Row 4: Full-width calendar
          {
            id: 'activity_calendar',
            title: 'Activity Calendar',
            defaultWidth: 12,
            defaultHeight: 3,
            render: () => <ActivityCalendar />,
          },
          // Row 5: Yearly comparison + Timing patterns
          {
            id: 'yearly_stats',
            title: 'Yearly Comparison',
            defaultWidth: 8,
            defaultHeight: 2,
            render: () => <YearlyStats />,
          },
          {
            id: 'daytime_stats',
            title: 'Time of Day',
            defaultWidth: 4,
            defaultHeight: 2,
            render: () => <DaytimeStats />,
          },
          {
            id: 'weekday_stats',
            title: 'Weekday',
            defaultWidth: 4,
            defaultHeight: 2,
            render: () => <WeekdayStats />,
          },
          // Hidden by default: Niche or equipment-dependent widgets
          {
            id: 'sport_chart',
            title: 'Sport Distribution',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <SportChart />,
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
            defaultHeight: 1,
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
            id: 'zone_trend',
            title: 'HR Zone Trend',
            defaultWidth: 8,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <ZoneTrend />,
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
            id: 'goal_recommendations',
            title: 'Goal Suggestions',
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <GoalRecommendations />,
          },
          {
            id: 'kudos_leaders',
            title: "Most Kudos'd",
            defaultWidth: 4,
            defaultHeight: 2,
            defaultHidden: true,
            render: () => <KudosLeaders />,
          },
        ]}
      />
    </div>
  )
}
