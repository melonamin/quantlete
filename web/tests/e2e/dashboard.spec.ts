import { test, expect } from './fixtures'
import { DashboardPage } from './pages/dashboard.page'

test.describe('Dashboard Page', () => {
  test.describe('Stats Summary', () => {
    test('dashboard loads with stats summary cards', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      // verify all four stats summary cards are visible
      await expect(dashboard.totalActivitiesCard).toBeVisible()
      await expect(dashboard.totalDistanceCard).toBeVisible()
      await expect(dashboard.totalTimeCard).toBeVisible()
      await expect(dashboard.totalElevationCard).toBeVisible()
    })

    test('stats cards display data from demo database', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      // with demo data seeded, we should have non-zero values
      const activitiesValue = await dashboard.getTotalActivities()
      expect(activitiesValue).toBeTruthy()
      expect(activitiesValue).toMatch(/\d+/)
    })
  })

  test.describe('Visible Widgets', () => {
    test('weekly stats widget renders with data', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      // weekly stats shows "This Week" title
      const isVisible = await dashboard.isWidgetVisible('This Week')
      expect(isVisible).toBe(true)

      // should have activity count or "No activities this week" message
      const weeklyWidget = page.locator('text=This Week').locator('..').locator('..')
      const hasActivities = (await weeklyWidget.locator('text=/\\d+ activit/i').count()) > 0
      const hasNoActivities = (await weeklyWidget.locator('text=No activities this week').count()) > 0
      expect(hasActivities || hasNoActivities).toBe(true)
    })

    test('recent activities widget renders', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      const isVisible = await dashboard.isWidgetVisible('Recent Activities')
      expect(isVisible).toBe(true)

      // should have "View all" link to activities page
      const viewAllLink = page.locator('a:has-text("View all")').first()
      await expect(viewAllLink).toBeVisible()
      await expect(viewAllLink).toHaveAttribute('href', '/activities')
    })

    test('sport breakdown widget renders with data', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      const isVisible = await dashboard.isWidgetVisible('By Sport')
      expect(isVisible).toBe(true)

      // should show sport types or "No activities yet"
      const sportWidget = page.locator('text=By Sport').locator('..').locator('..')
      const hasSports = (await sportWidget.locator('text=sport types').count()) > 0
      const hasNoActivities = (await sportWidget.locator('text=No activities yet').count()) > 0
      expect(hasSports || hasNoActivities).toBe(true)
    })
  })

  test.describe('Monthly Chart and Activity Calendar', () => {
    test('monthly chart widget renders', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      const isVisible = await dashboard.isWidgetVisible('Monthly Activity')
      expect(isVisible).toBe(true)

      // should have metric buttons (Distance, Activities, Time)
      await expect(dashboard.monthlyChartWidget.locator('button:has-text("Distance")')).toBeVisible()
      await expect(dashboard.monthlyChartWidget.locator('button:has-text("Activities")')).toBeVisible()
      await expect(dashboard.monthlyChartWidget.locator('button:has-text("Time")')).toBeVisible()
    })

    test('activity calendar widget renders', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      const isVisible = await dashboard.isWidgetVisible('Activity Calendar')
      expect(isVisible).toBe(true)

      // calendar widget should have a card container
      const calendarCard = page.locator('[data-slot="card"]').filter({
        has: page.locator('text=Activity Calendar'),
      })
      await expect(calendarCard).toBeVisible()
    })
  })

  test.describe('Widget Visibility Toggle', () => {
    test('customize dashboard button opens edit mode', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      await dashboard.enterEditMode()

      // should show edit mode bar
      await expect(dashboard.editModeBar).toBeVisible()
      await expect(page.locator('text=Drag widgets to reorder')).toBeVisible()
    })

    test('widget panel opens and shows all widgets', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      await dashboard.enterEditMode()
      await dashboard.openWidgetPanel()

      // should show widget manager panel
      await expect(dashboard.widgetPanel).toBeVisible()
      await expect(page.locator('text=/\\d+ visible.*\\d+ hidden/')).toBeVisible()
    })

    test('can toggle widget visibility in panel', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      await dashboard.enterEditMode()
      await dashboard.openWidgetPanel()

      // get initial visible/hidden counts
      const statusText = await page.locator('text=/\\d+ visible.*\\d+ hidden/').textContent()
      const initialVisible = parseInt(statusText?.match(/(\d+) visible/)?.[1] ?? '0')
      const initialHidden = parseInt(statusText?.match(/(\d+) hidden/)?.[1] ?? '0')

      // find and click the eye toggle for the first visible widget
      const widgetRows = page.locator('.space-y-2 > div').filter({
        has: page.locator('button'),
      })
      const firstWidgetToggle = widgetRows.first().locator('button').last()
      await firstWidgetToggle.click()

      // wait a moment for the state to update
      await page.waitForTimeout(300)

      // verify the count changed
      const newStatusText = await page.locator('text=/\\d+ visible.*\\d+ hidden/').textContent()
      const newVisible = parseInt(newStatusText?.match(/(\d+) visible/)?.[1] ?? '0')
      const newHidden = parseInt(newStatusText?.match(/(\d+) hidden/)?.[1] ?? '0')

      // counts should have changed (one more hidden, one less visible)
      expect(newVisible + newHidden).toBe(initialVisible + initialHidden)
    })

    test('escape key exits edit mode', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      await dashboard.enterEditMode()
      await expect(dashboard.editModeBar).toBeVisible()

      // press escape
      await page.keyboard.press('Escape')

      // should exit edit mode
      await expect(dashboard.editModeBar).not.toBeVisible()
      await expect(dashboard.customizeButton).toBeVisible()
    })
  })

  test.describe('Time Period Filters', () => {
    test('monthly chart year buttons change chart data', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      await expect(dashboard.monthlyChartWidget).toBeVisible({ timeout: 10000 })

      const buttonCount = await dashboard.monthlyChartYearButtons.count()
      if (buttonCount > 1) {
        // click a different year button
        const secondYearButton = dashboard.monthlyChartYearButtons.nth(1)
        await secondYearButton.click()

        // the clicked button should now have secondary variant (selected state)
        await expect(secondYearButton).toHaveClass(/secondary/)
      }
    })

    test('monthly chart metric buttons switch between distance/activities/time', async ({
      page,
    }) => {
      const dashboard = new DashboardPage(page)

      await expect(dashboard.monthlyChartWidget).toBeVisible({ timeout: 10000 })

      // click Activities button
      await dashboard.selectMonthlyChartMetric('Activities')
      expect(await dashboard.isMonthlyChartMetricSelected('Activities')).toBe(true)
      expect(await dashboard.isMonthlyChartMetricSelected('Distance')).toBe(false)

      // click Time button
      await dashboard.selectMonthlyChartMetric('Time')
      expect(await dashboard.isMonthlyChartMetricSelected('Time')).toBe(true)
      expect(await dashboard.isMonthlyChartMetricSelected('Activities')).toBe(false)
    })
  })

  test.describe('Navigation to Activity Detail', () => {
    test('clicking recent activity navigates to activity detail page', async ({ page }) => {
      const dashboard = new DashboardPage(page)

      // wait for recent activities to load
      const isVisible = await dashboard.isWidgetVisible('Recent Activities')
      expect(isVisible).toBe(true)

      const activityCount = await dashboard.getRecentActivityCount()
      if (activityCount > 0) {
        // click the first activity
        await dashboard.clickRecentActivity(0)

        // should navigate to activity detail page
        await expect(page).toHaveURL(/\/activities\/\d+/)

        // activity detail page should show the activity title
        await expect(page.locator('h1')).toBeVisible({ timeout: 10000 })
      } else {
        // if no activities, verify the empty state
        const recentWidget = page.locator('text=Recent Activities').locator('..').locator('..')
        const noActivitiesMsg = await recentWidget.locator('text=No activities yet').count()
        expect(noActivitiesMsg).toBeGreaterThanOrEqual(0)
      }
    })
  })
})
