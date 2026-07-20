import { test, expect } from '../fixtures'
import { DashboardPage } from '../pages/dashboard.page'
import { ActivityDetailPage } from '../pages/activity-detail.page'

// journey: dashboard to activity deep dive
// dashboard -> click recent activity -> view detail -> check segments -> back to dashboard
test.describe('Dashboard to Activity Deep Dive Journey', () => {
  test('user can navigate from dashboard to activity and back', async ({ page }) => {
    // step 1: start on dashboard and verify it loaded
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.waitForPageLoad()

    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()
    await expect(page.locator('text=Total Activities').first()).toBeVisible()

    // step 2: verify recent activities widget is visible
    await expect(page.locator('text=Recent Activities').first()).toBeVisible({ timeout: 10000 })

    // step 3: check if there are recent activities to click
    const activityCount = await dashboardPage.getRecentActivityCount()
    if (activityCount === 0) {
      // no activities to navigate to, skip the rest of the test
      return
    }

    // step 4: get the href of the first activity before clicking
    const activityLink = page.locator('a[href*="/activities/"]:has(.truncate)').first()
    const href = await activityLink.getAttribute('href')
    expect(href).toMatch(/\/activities\/\d+/)

    // step 5: click the first recent activity
    await dashboardPage.clickRecentActivity(0)

    // step 6: verify navigation to activity detail page
    await expect(page).toHaveURL(/\/activities\/\d+/)

    // step 7: wait for activity detail to load
    const activityDetailPage = new ActivityDetailPage(page)
    await activityDetailPage.waitForActivityLoad()

    // step 8: verify activity detail page elements are visible
    await expect(activityDetailPage.activityTitle).toBeVisible()

    // step 9: check for stats cards
    const hasDistanceCard = await activityDetailPage.hasStatCard('Distance')
    const hasMovingTimeCard = await activityDetailPage.hasStatCard('Moving Time')
    expect(hasDistanceCard || hasMovingTimeCard).toBe(true)

    // step 10: check for map if activity has GPS data
    const hasMap = await activityDetailPage.hasMap()
    if (hasMap) {
      await expect(activityDetailPage.activityMap).toBeVisible()
    }

    // step 11: check for elevation profile if available
    const hasElevation = await activityDetailPage.hasElevationProfile()
    if (hasElevation) {
      await expect(activityDetailPage.elevationProfileSection).toBeVisible()
    }

    // step 12: check for photos section if available
    const hasPhotos = await activityDetailPage.hasPhotos()
    if (hasPhotos) {
      await expect(activityDetailPage.photosSection).toBeVisible()
    }

    // step 13: verify activity streams section
    const hasStreams = await activityDetailPage.hasActivityStreams()
    if (hasStreams) {
      await expect(activityDetailPage.activityStreamsSection).toBeVisible()
    }

    // step 14: navigate back to dashboard using sidebar navigation
    await page.getByRole('link', { name: /dashboard/i }).click()

    // step 15: verify we're back on dashboard
    await expect(page).toHaveURL(/\/$/)
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()

    // step 16: verify dashboard still shows data correctly
    await expect(page.locator('text=Total Activities').first()).toBeVisible()
    await expect(page.locator('text=Recent Activities').first()).toBeVisible()
  })

  test('user can navigate back using the back link on activity detail', async ({ page }) => {
    // step 1: start on dashboard
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.waitForPageLoad()

    // step 2: check for recent activities
    const activityCount = await dashboardPage.getRecentActivityCount()
    if (activityCount === 0) {
      return
    }

    // step 3: navigate to an activity
    await dashboardPage.clickRecentActivity(0)
    await expect(page).toHaveURL(/\/activities\/\d+/)

    // step 4: wait for page to load
    const activityDetailPage = new ActivityDetailPage(page)
    await activityDetailPage.waitForActivityLoad()

    // step 5: use the "Back to Activities" link
    await activityDetailPage.goBackToActivities()

    // step 6: verify we're on the activities list page (not dashboard)
    await expect(page).toHaveURL(/\/activities$/)
    await expect(page.locator('h1').filter({ hasText: 'Activities' })).toBeVisible()
  })
})
