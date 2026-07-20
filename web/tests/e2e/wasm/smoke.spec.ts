import { ActivitiesPage } from '../pages/activities.page'
import { DashboardPage } from '../pages/dashboard.page'
import { expect, test } from './fixtures'

test.describe('WASM demo smoke tests', () => {
  test('app boots in demo/WASM mode and renders bundled dashboard data', async ({ page }) => {
    const dashboard = new DashboardPage(page)

    await expect(page.getByText('Demo mode with sample data.')).toBeVisible()
    await expect(page.getByText('Failed to load dashboard data')).toHaveCount(0)
    await expect(dashboard.getPageIdentifier()).toBeVisible()
    await expect(dashboard.totalActivitiesCard).toBeVisible()

    const totalActivities = await dashboard.getTotalActivities()
    expect(Number(totalActivities?.replaceAll(',', ''))).toBeGreaterThan(0)
    await expect(page.getByText('Recent Activities').first()).toBeVisible()
  })

  test('activities list renders rows from the bundled demo database', async ({ page }) => {
    const activities = new ActivitiesPage(page)

    await activities.goto()
    await activities.waitForTableLoad()

    await expect(activities.activitiesTable).toBeVisible()
    expect(await activities.getActivityRowCount()).toBeGreaterThan(0)
  })
})
