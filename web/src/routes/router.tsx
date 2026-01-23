import {
  createRouter,
  createRootRoute,
  createRoute,
  Outlet,
  redirect,
} from '@tanstack/react-router'
import { RootLayout } from '@/components/layout/root-layout'
import { DashboardPage } from '@/pages/dashboard'
import { ActivitiesPage } from '@/pages/activities'
import { ActivityDetailPage } from '@/pages/activity-detail'
import { HeatmapPage } from '@/pages/heatmap'
import { CalendarPage } from '@/pages/calendar'
import { SegmentsPage } from '@/pages/segments'
import { GearPage } from '@/pages/gear'
import { EddingtonPage } from '@/pages/eddington'
import { BestEffortsPage } from '@/pages/best-efforts'
import { WrappedPage } from '@/pages/wrapped'
import { SettingsPage } from '@/pages/settings'
import { AthletePage } from '@/pages/athlete'
import { TrainingLoadPage } from '@/pages/training-load'
import { PowerPage } from '@/pages/power'
import { PhotosPage } from '@/pages/photos'
import { ChallengesPage } from '@/pages/challenges'
import { OAuthCallbackPage } from '@/pages/oauth-callback'
import { MonthlyStatsPage } from '@/pages/monthly-stats'
import { ExportPage } from '@/pages/export'
import { BadgesPage } from '@/pages/badges'

// Root route with layout
const rootRoute = createRootRoute({
  component: () => (
    <RootLayout>
      <Outlet />
    </RootLayout>
  ),
})

// Dashboard (home)
const dashboardRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: DashboardPage,
})

// Activities list
const activitiesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities',
  component: ActivitiesPage,
})

// Activity detail
const activityDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/activities/$activityId',
  component: ActivityDetailPage,
})

// Heatmap
const heatmapRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/heatmap',
  component: HeatmapPage,
})

// Calendar
const calendarRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/calendar',
  component: CalendarPage,
})

// Segments
const segmentsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/segments',
  component: SegmentsPage,
})

// Gear
const gearRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/gear',
  component: GearPage,
})

// Eddington
const eddingtonRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/eddington',
  component: EddingtonPage,
})

// Best Efforts
const bestEffortsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/best-efforts',
  component: BestEffortsPage,
})

// Wrapped
const wrappedRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/wrapped',
  component: WrappedPage,
})

// Redirect old /rewind to /wrapped for backwards compatibility
const rewindRedirectRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/rewind',
  beforeLoad: () => {
    throw redirect({ to: '/wrapped' })
  },
})

// Settings
const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: SettingsPage,
})

// Athlete Profile
const athleteRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/athlete',
  component: AthletePage,
})

// Training Load
const trainingLoadRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/training-load',
  component: TrainingLoadPage,
})

// Power
const powerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/power',
  component: PowerPage,
})

// Photos
const photosRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/photos',
  component: PhotosPage,
})

// Challenges
const challengesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/challenges',
  component: ChallengesPage,
})

// OAuth callback (WASM mode)
const oauthCallbackRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/oauth/callback',
  component: OAuthCallbackPage,
})

// Monthly Stats
const monthlyStatsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/monthly-stats',
  component: MonthlyStatsPage,
})

// Export
const exportRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/export',
  component: ExportPage,
})

// Badges
const badgesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/badges',
  component: BadgesPage,
})

// Route tree
const routeTree = rootRoute.addChildren([
  dashboardRoute,
  activitiesRoute,
  activityDetailRoute,
  heatmapRoute,
  calendarRoute,
  segmentsRoute,
  gearRoute,
  eddingtonRoute,
  bestEffortsRoute,
  wrappedRoute,
  rewindRedirectRoute,
  trainingLoadRoute,
  powerRoute,
  photosRoute,
  challengesRoute,
  athleteRoute,
  settingsRoute,
  oauthCallbackRoute,
  monthlyStatsRoute,
  exportRoute,
  badgesRoute,
])

// Create router
export const router = createRouter({ routeTree })

// Type declaration for router
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
