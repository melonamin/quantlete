import { createRouter, createRootRoute, createRoute, Outlet } from '@tanstack/react-router'
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
import { SettingsPage } from '@/pages/settings'

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

// Settings
const settingsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/settings',
  component: SettingsPage,
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
  settingsRoute,
])

// Create router
export const router = createRouter({ routeTree })

// Type declaration for router
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
