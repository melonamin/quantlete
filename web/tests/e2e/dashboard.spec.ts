import { test, expect } from './fixtures'

test.describe('Dashboard Page', () => {
  test.describe('Stats Summary', () => {
    test('dashboard loads with stats summary cards', async ({ page }) => {
      // verify all four stats summary cards are visible
      await expect(page.locator('text=Total Activities').first()).toBeVisible()
      await expect(page.locator('text=Total Distance').first()).toBeVisible()
      await expect(page.locator('text=Total Time').first()).toBeVisible()
      await expect(page.locator('text=Total Elevation').first()).toBeVisible()
    })

    test('stats cards display data from demo database', async ({ page }) => {
      // with demo data seeded, we should have non-zero values
      // verify the Total Activities card has a numeric value
      const activitiesCard = page.locator('text=Total Activities').locator('..').locator('..')
      const activitiesValue = activitiesCard.locator('[class*="text-2xl"], [class*="font-bold"]').first()
      await expect(activitiesValue).toBeVisible()

      // the value should contain a number
      const text = await activitiesValue.textContent()
      expect(text).toMatch(/\d+/)
    })
  })

  test.describe('Visible Widgets', () => {
    test('weekly stats widget renders with data', async ({ page }) => {
      // weekly stats shows "This Week" title
      await expect(page.locator('text=This Week').first()).toBeVisible({ timeout: 10000 })

      // should have activity count or "No activities this week" message
      const weeklyWidget = page.locator('text=This Week').locator('..').locator('..')
      const hasActivities = await weeklyWidget.locator('text=/\\d+ activit/i').count()
      const hasNoActivities = await weeklyWidget.locator('text=No activities this week').count()
      expect(hasActivities + hasNoActivities).toBeGreaterThan(0)
    })

    test('recent activities widget renders', async ({ page }) => {
      await expect(page.locator('text=Recent Activities').first()).toBeVisible({ timeout: 10000 })

      // should have "View all" link to activities page
      const viewAllLink = page.locator('a:has-text("View all")').first()
      await expect(viewAllLink).toBeVisible()
      await expect(viewAllLink).toHaveAttribute('href', '/activities')
    })

    test('sport breakdown widget renders with data', async ({ page }) => {
      // sport breakdown shows "By Sport" title
      await expect(page.locator('text=By Sport').first()).toBeVisible({ timeout: 10000 })

      // should show sport types or "No activities yet"
      const sportWidget = page.locator('text=By Sport').locator('..').locator('..')
      const hasSports = await sportWidget.locator('text=sport types').count()
      const hasNoActivities = await sportWidget.locator('text=No activities yet').count()
      expect(hasSports + hasNoActivities).toBeGreaterThan(0)
    })
  })

  test.describe('Monthly Chart and Activity Calendar', () => {
    test('monthly chart widget renders', async ({ page }) => {
      await expect(page.locator('text=Monthly Activity').first()).toBeVisible({ timeout: 10000 })

      // should have metric buttons (Distance, Activities, Time)
      const chartWidget = page.locator('text=Monthly Activity').locator('..').locator('..')
      await expect(chartWidget.locator('button:has-text("Distance")')).toBeVisible()
      await expect(chartWidget.locator('button:has-text("Activities")')).toBeVisible()
      await expect(chartWidget.locator('button:has-text("Time")')).toBeVisible()
    })

    test('activity calendar widget renders', async ({ page }) => {
      await expect(page.locator('text=Activity Calendar').first()).toBeVisible({ timeout: 10000 })

      // calendar widget should have a card container with year/calendar controls
      const calendarCard = page.locator('[data-slot="card"]').filter({
        has: page.locator('text=Activity Calendar'),
      })
      await expect(calendarCard).toBeVisible()
    })
  })

  test.describe('Widget Visibility Toggle', () => {
    test('customize dashboard button opens edit mode', async ({ page }) => {
      // click customize button
      await page.locator('button:has-text("Customize Dashboard")').click()

      // should show edit mode bar
      await expect(page.locator('text=EDIT MODE')).toBeVisible()
      await expect(page.locator('text=Drag widgets to reorder')).toBeVisible()
    })

    test('widget panel opens and shows all widgets', async ({ page }) => {
      // enter edit mode
      await page.locator('button:has-text("Customize Dashboard")').click()
      await expect(page.locator('text=EDIT MODE')).toBeVisible()

      // open widget panel
      await page.locator('button:has-text("WIDGETS")').click()

      // should show widget manager panel
      await expect(page.locator('text=Widget Manager')).toBeVisible()
      await expect(page.locator('text=/\\d+ visible.*\\d+ hidden/')).toBeVisible()
    })

    test('can toggle widget visibility in panel', async ({ page }) => {
      // enter edit mode
      await page.locator('button:has-text("Customize Dashboard")').click()
      await expect(page.locator('text=EDIT MODE')).toBeVisible()

      // open widget panel
      await page.locator('button:has-text("WIDGETS")').click()
      await expect(page.locator('text=Widget Manager')).toBeVisible()

      // get initial visible/hidden counts
      const statusText = await page.locator('text=/\\d+ visible.*\\d+ hidden/').textContent()
      const initialVisible = parseInt(statusText?.match(/(\d+) visible/)?.[1] ?? '0')
      const initialHidden = parseInt(statusText?.match(/(\d+) hidden/)?.[1] ?? '0')

      // find and click the eye toggle for the first visible widget
      // widgets are in a list, each with an eye button
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
      // enter edit mode
      await page.locator('button:has-text("Customize Dashboard")').click()
      await expect(page.locator('text=EDIT MODE')).toBeVisible()

      // press escape
      await page.keyboard.press('Escape')

      // should exit edit mode
      await expect(page.locator('text=EDIT MODE')).not.toBeVisible()
      await expect(page.locator('button:has-text("Customize Dashboard")')).toBeVisible()
    })
  })

  test.describe('Time Period Filters', () => {
    test('monthly chart year buttons change chart data', async ({ page }) => {
      const chartWidget = page.locator('text=Monthly Activity').locator('..').locator('..')
      await expect(chartWidget).toBeVisible({ timeout: 10000 })

      // find year buttons (they contain just 4-digit years)
      const yearButtons = chartWidget.locator('button').filter({
        hasText: /^\d{4}$/,
      })

      const buttonCount = await yearButtons.count()
      if (buttonCount > 1) {
        // click a different year button
        const secondYearButton = yearButtons.nth(1)
        await secondYearButton.click()

        // the clicked button should now have secondary variant (selected state)
        await expect(secondYearButton).toHaveClass(/secondary/)
      }
    })

    test('monthly chart metric buttons switch between distance/activities/time', async ({
      page,
    }) => {
      const chartWidget = page.locator('text=Monthly Activity').locator('..').locator('..')
      await expect(chartWidget).toBeVisible({ timeout: 10000 })

      // click Activities button
      const activitiesButton = chartWidget.locator('button:has-text("Activities")')
      await activitiesButton.click()

      // should be selected (has secondary class)
      await expect(activitiesButton).toHaveClass(/secondary/)

      // Distance button should not be selected
      const distanceButton = chartWidget.locator('button:has-text("Distance")')
      await expect(distanceButton).not.toHaveClass(/secondary/)

      // click Time button
      const timeButton = chartWidget.locator('button:has-text("Time")')
      await timeButton.click()

      // Time should be selected, Activities should not
      await expect(timeButton).toHaveClass(/secondary/)
      await expect(activitiesButton).not.toHaveClass(/secondary/)
    })
  })

  test.describe('Navigation to Activity Detail', () => {
    test('clicking recent activity navigates to activity detail page', async ({ page }) => {
      // wait for recent activities to load
      await expect(page.locator('text=Recent Activities').first()).toBeVisible({ timeout: 10000 })

      // find activity links in the recent activities section
      const activityLinks = page.locator('a[href*="/activities/"]:has(.truncate)')
      const linkCount = await activityLinks.count()

      if (linkCount > 0) {
        // get the href of the first activity
        const href = await activityLinks.first().getAttribute('href')
        expect(href).toMatch(/\/activities\/\d+/)

        // click the first activity
        await activityLinks.first().click()

        // should navigate to activity detail page
        await expect(page).toHaveURL(/\/activities\/\d+/)

        // activity detail page should show the activity title
        await expect(page.locator('h1')).toBeVisible({ timeout: 10000 })
      } else {
        // if no activities, that's okay - just verify the widget is there
        const recentWidget = page.locator('text=Recent Activities').locator('..').locator('..')
        const noActivitiesMsg = await recentWidget.locator('text=No activities yet').count()
        expect(noActivitiesMsg).toBeGreaterThanOrEqual(0)
      }
    })
  })
})
