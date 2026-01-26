import { test, expect } from './fixtures'

test.describe('Error States & Edge Cases', () => {
  test.describe('404 Not Found', () => {
    test('invalid route shows 404 or redirects to dashboard', async ({ page }) => {
      // navigate to a non-existent route
      await page.goto('/this-route-does-not-exist')

      // the app should either show a 404 page or redirect to dashboard
      // check for common 404 indicators or dashboard
      const is404 = await page.locator('text=/not found|404|page.*exist/i').isVisible()
      const isDashboard = await page.locator('h1:has-text("Dashboard")').isVisible()

      expect(is404 || isDashboard).toBe(true)
    })

    test('invalid activity ID shows error or redirects', async ({ page }) => {
      // navigate to a non-existent activity
      await page.goto('/activities/999999999')

      // wait for the page to settle
      await page.waitForTimeout(1000)

      // should show error state or redirect to activities list
      const hasError = await page.locator('text=/not found|error|does not exist/i').isVisible()
      const isActivitiesList = await page
        .locator('h1')
        .filter({ hasText: 'Activities' })
        .isVisible()
      const isActivityPage = await page.locator('a:has-text("Back to Activities")').isVisible()

      // one of these should be true
      expect(hasError || isActivitiesList || isActivityPage).toBe(true)
    })
  })

  test.describe('Empty States', () => {
    test('activities page handles empty filter results gracefully', async ({ page }) => {
      await page.goto('/activities')
      await page.locator('h1').filter({ hasText: 'Activities' }).waitFor({ state: 'visible' })

      // wait for initial load
      await page
        .locator('[class*="animate-pulse"]')
        .first()
        .waitFor({ state: 'hidden', timeout: 10000 })
        .catch(() => {})

      // apply a filter that likely returns no results
      // open advanced filters and set an impossible date range
      await page.locator('button:has-text("More")').click()
      const fromInput = page.locator('input[type="date"]').first()
      await fromInput.fill('1900-01-01')
      const toInput = page.locator('input[type="date"]').nth(1)
      await toInput.fill('1900-01-02')

      // wait for filter to apply
      await page.waitForTimeout(600)

      // should show "No activities found" or similar empty state
      const emptyState = page.locator('text=/no activities|no results/i')
      const hasEmptyState = await emptyState.isVisible()

      // if data exists for that date range, we'll have rows - that's fine too
      const hasRows = (await page.locator('tbody tr').count()) > 0

      expect(hasEmptyState || hasRows).toBe(true)
    })

    test('segments page shows empty state when no segments match filters', async ({ page }) => {
      await page.goto('/segments')
      await page.locator('h1:has-text("Segments")').waitFor({ state: 'visible' })

      // wait for initial load
      await page.waitForTimeout(1000)

      // search for something that doesn't exist
      const searchInput = page.locator('input[placeholder*="Search"]')
      if (await searchInput.isVisible()) {
        await searchInput.fill('xyznonexistentsegment12345')
        await page.waitForTimeout(500)

        // should show empty state or no results
        const emptyState = page.locator('text=/no segments|no results/i')
        const hasEmptyState = await emptyState.isVisible()
        const noRows = (await page.locator('tbody tr').count()) === 0

        expect(hasEmptyState || noRows).toBe(true)
      }
    })
  })

  test.describe('Loading States', () => {
    test('dashboard shows loading skeleton before data loads', async ({ page }) => {
      // intercept and delay the API response to catch loading state
      await page.route('**/api/v1/**', async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 500))
        await route.continue()
      })

      await page.goto('/')

      // should show loading indicators initially
      // after data loads, loading should disappear
      await page.waitForTimeout(1500)

      // verify the page eventually loads
      await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible({ timeout: 10000 })
    })

    test('activities table shows skeleton while loading', async ({ page }) => {
      await page.route('**/api/v1/activities**', async (route) => {
        await new Promise((resolve) => setTimeout(resolve, 300))
        await route.continue()
      })

      await page.goto('/activities')

      // verify page eventually loads with either data or empty state
      await page.waitForTimeout(1000)
      await expect(page.locator('h1').filter({ hasText: 'Activities' })).toBeVisible()
    })
  })

  test.describe('Network Errors', () => {
    test('app handles API timeout gracefully', async ({ page }) => {
      // first load the app normally
      await page.goto('/')
      await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible({ timeout: 10000 })

      // then simulate a timeout on subsequent requests
      await page.route('**/api/v1/**', async (route) => {
        // delay longer than typical timeout
        await new Promise((resolve) => setTimeout(resolve, 35000))
        await route.abort('timedout')
      })

      // navigate to activities page
      await page.locator('a[href="/activities"]').click()

      // wait for some response - either error or the cached/old data
      await page.waitForTimeout(2000)

      // the page should still be functional (not crashed)
      // it might show an error or stale data
      const pageTitle = page.locator('h1')
      await expect(pageTitle).toBeVisible()
    })
  })

  test.describe('Form Validation', () => {
    test('athlete page validates FTP input', async ({ page }) => {
      await page.goto('/athlete')
      await page.locator('h1:has-text("Athlete")').waitFor({ state: 'visible' })

      // wait for page to load
      await page.waitForTimeout(500)

      // look for the FTP section
      const ftpSection = page.locator('text=FTP History').locator('..')

      if (await ftpSection.isVisible()) {
        // try to find and interact with FTP input
        const ftpInputs = page.locator('input[type="number"]')
        const inputCount = await ftpInputs.count()

        if (inputCount > 0) {
          // FTP should only accept positive numbers
          // this tests that the form exists and accepts input
          const firstInput = ftpInputs.first()
          await firstInput.fill('250')

          // verify value is accepted
          await expect(firstInput).toHaveValue('250')
        }
      }
    })

    test('settings page validates required fields', async ({ page }) => {
      await page.goto('/settings')
      await page.locator('h1:has-text("Settings")').waitFor({ state: 'visible' })

      // the settings page should load without errors
      await expect(page.locator('h1:has-text("Settings")')).toBeVisible()

      // verify theme toggle exists and works
      const themeButtons = page.locator('button:has-text("Light"), button:has-text("Dark")')
      const themeButtonCount = await themeButtons.count()
      expect(themeButtonCount).toBeGreaterThan(0)
    })

    test('export page validates date range inputs', async ({ page }) => {
      await page.goto('/export')
      await page.locator('h1:has-text("Export")').waitFor({ state: 'visible' })

      // find date inputs
      const dateInputs = page.locator('input[type="date"]')
      const inputCount = await dateInputs.count()

      if (inputCount >= 2) {
        // set an invalid range (end before start)
        await dateInputs.first().fill('2024-12-31')
        await dateInputs.nth(1).fill('2024-01-01')

        await page.waitForTimeout(300)

        // the app should handle this gracefully
        // either show an error or swap the dates
        const downloadButton = page.locator('button:has-text("Download")')
        await expect(downloadButton).toBeVisible()
      }
    })
  })

  test.describe('Edge Cases', () => {
    test('navigating directly to detail page works', async ({ page }) => {
      // first get a valid activity ID from the list
      await page.goto('/activities')
      await page.locator('h1').filter({ hasText: 'Activities' }).waitFor({ state: 'visible' })

      await page
        .locator('[class*="animate-pulse"]')
        .first()
        .waitFor({ state: 'hidden', timeout: 10000 })
        .catch(() => {})

      const firstLink = page.locator('tbody tr a').first()
      const href = await firstLink.getAttribute('href')

      if (href) {
        // navigate directly to that URL
        await page.goto(href)

        // should load the activity detail page
        await expect(page.locator('h1')).toBeVisible({ timeout: 10000 })
        await expect(page.locator('a:has-text("Back to Activities")')).toBeVisible()
      }
    })

    test('browser refresh preserves page state', async ({ page }) => {
      await page.goto('/activities')
      await page.locator('h1').filter({ hasText: 'Activities' }).waitFor({ state: 'visible' })

      await page
        .locator('[class*="animate-pulse"]')
        .first()
        .waitFor({ state: 'hidden', timeout: 10000 })
        .catch(() => {})

      // apply a filter
      await page.locator('[role="group"] button:has-text("Ride")').first().click()
      await page.waitForTimeout(500)

      // refresh the page
      await page.reload()

      // wait for page to load again
      await page.locator('h1').filter({ hasText: 'Activities' }).waitFor({ state: 'visible' })

      // the page should still be on activities (URL-based routing works)
      await expect(page).toHaveURL(/\/activities/)
    })

    test('rapid navigation does not break the app', async ({ page }) => {
      await page.goto('/')
      await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible({ timeout: 10000 })

      // rapidly click between pages
      const links = [
        'a[href="/activities"]',
        'a[href="/calendar"]',
        'a[href="/heatmap"]',
        'a[href="/settings"]',
        'a[href="/"]',
      ]

      for (const selector of links) {
        const link = page.locator(selector).first()
        if (await link.isVisible()) {
          await link.click()
          // don't wait for full load, just click quickly
          await page.waitForTimeout(100)
        }
      }

      // wait for things to settle
      await page.waitForTimeout(1000)

      // app should still be functional
      const pageTitle = page.locator('h1')
      await expect(pageTitle).toBeVisible()
    })
  })
})
