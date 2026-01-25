import { test, expect } from './fixtures'

test.describe('Smoke Tests', () => {
  test('app loads and shows dashboard', async ({ page }) => {
    // verify we're on the dashboard
    await expect(page.locator('h1')).toContainText('Dashboard')

    // verify the page subtitle is visible
    await expect(page.locator('text=Your activity statistics at a glance')).toBeVisible()
  })

  test('dashboard renders with demo data', async ({ page }) => {
    // wait for stats summary to load - it should have activity data from demo
    // the stats summary section shows total activities, distance, elevation, time
    const statsSection = page.locator('text=Total Activities').first()
    await expect(statsSection).toBeVisible({ timeout: 10000 })

    // verify there are some stats cards with numbers
    // demo data should have generated activities, so we expect non-zero values
    const statsCards = page.locator('[class*="card"]').filter({
      has: page.locator('text=/\\d+/'),
    })
    await expect(statsCards.first()).toBeVisible()
  })

  test('dashboard widgets are visible', async ({ page }) => {
    // verify at least one widget is rendered
    // widgets have titles and content areas
    const widgetTitles = page.locator('text=Recent Activities')
    await expect(widgetTitles.first()).toBeVisible({ timeout: 10000 })

    // check for weekly stats widget
    await expect(page.locator('text=Weekly Stats').first()).toBeVisible()
  })

  test('navigation sidebar is functional', async ({ page }) => {
    // find and click on Activities link in navigation
    const activitiesLink = page.getByRole('link', { name: /activities/i }).first()
    await expect(activitiesLink).toBeVisible()

    // click and verify navigation works
    await activitiesLink.click()

    // wait for activities page to load
    await expect(page.locator('h1')).toContainText('Activities', { timeout: 10000 })
  })

  test('page has no console errors', async ({ page }) => {
    const errors: string[] = []

    // collect console errors
    page.on('console', (msg) => {
      if (msg.type() === 'error') {
        errors.push(msg.text())
      }
    })

    // navigate to dashboard and wait
    await page.goto('/')
    await page.waitForTimeout(2000)

    // filter out known acceptable errors
    const criticalErrors = errors.filter((err) => {
      // ignore favicon 404s
      if (err.includes('favicon') || err.includes('404')) return false
      // ignore CSP errors - these are expected in dev/test mode due to inline scripts and external fonts
      if (err.includes('Content Security Policy')) return false
      // ignore Google Fonts loading errors
      if (err.includes('fonts.googleapis.com') || err.includes('fonts.gstatic.com')) return false
      return true
    })

    expect(criticalErrors).toHaveLength(0)
  })
})
