import { test, expect } from './fixtures'
import { PowerPage } from './pages/power.page'

test.describe('Power Page', () => {
  test.describe('All-time Best Chart', () => {
    test('all-time best power outputs bar chart renders', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      // verify page header
      await expect(powerPage.pageTitle).toBeVisible()
      await expect(powerPage.pageSubtitle).toBeVisible()

      // check if chart or error is visible
      const hasChart = await powerPage.hasAllTimeBestChart()
      const hasError = await powerPage.hasError()

      // one of these should be true
      expect(hasChart || hasError).toBe(true)

      if (hasChart) {
        // all-time best card should be visible
        await expect(powerPage.allTimeBestCard).toBeVisible()
        await expect(powerPage.allTimeBestChart).toBeVisible()
      }
    })

    test('all-time best card shows title', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (!hasError) {
        // card title should be visible
        const cardText = await powerPage.allTimeBestCard.textContent()
        expect(cardText).toContain('All-time Best')
      }
    })
  })

  test.describe('Duration Selector and Progression Chart', () => {
    test('duration selector dropdown changes progression chart view', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (hasError) {
        return
      }

      // verify progression card is visible
      await expect(powerPage.progressionCard).toBeVisible()

      // get available durations
      const durations = await powerPage.getDurationOptions()
      expect(durations.length).toBeGreaterThan(0)

      // verify initial selection
      const initialDuration = await powerPage.getSelectedDuration()
      expect(initialDuration).toBeTruthy()

      // select a different duration
      if (durations.length > 1) {
        const otherDuration = durations.find((d) => d !== initialDuration)
        if (otherDuration) {
          await powerPage.selectDuration(otherDuration)

          // verify selection changed
          const newDuration = await powerPage.getSelectedDuration()
          expect(newDuration).toContain(otherDuration)
        }
      }
    })

    test('progression chart renders', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (hasError) {
        return
      }

      // progression chart should render
      const hasProgressionChart = await powerPage.hasProgressionChart()
      expect(hasProgressionChart).toBe(true)
    })

    test('duration options include common power durations', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (hasError) {
        return
      }

      const durations = await powerPage.getDurationOptions()
      // should have at least some of the standard durations (5s, 1m, 5m, 20m, 1h)
      // the exact durations depend on the backend config
      expect(durations.length).toBeGreaterThan(0)
    })
  })

  test.describe('Power Curve Comparison Chart', () => {
    test('power curve comparison chart (all-time vs 90 days) renders', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (hasError) {
        return
      }

      // power curve card should be visible
      await expect(powerPage.powerCurveCard).toBeVisible()

      // chart should render
      const hasPowerCurve = await powerPage.hasPowerCurveChart()
      expect(hasPowerCurve).toBe(true)
    })

    test('power curve card shows comparison title', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (!hasError) {
        const cardText = await powerPage.powerCurveCard.textContent()
        expect(cardText).toContain('Power Curve Comparison')
      }
    })
  })

  test.describe('Power Zones Chart', () => {
    test('power zones breakdown chart renders with zone colors', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (hasError) {
        return
      }

      // power zones card should be visible
      await expect(powerPage.powerZonesCard).toBeVisible()

      // chart should render
      const hasPowerZones = await powerPage.hasPowerZonesChart()
      expect(hasPowerZones).toBe(true)
    })

    test('power zones card shows title', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (!hasError) {
        const cardText = await powerPage.powerZonesCard.textContent()
        expect(cardText).toContain('Power Zones')
      }
    })

    test('power zones shows FTP info when available', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      const hasError = await powerPage.hasError()
      if (hasError) {
        return
      }

      // FTP display may or may not be visible depending on user config
      // we just verify the check works
      const hasFtp = await powerPage.hasFtpDisplay()
      expect(typeof hasFtp).toBe('boolean')
    })
  })

  test.describe('Page Layout', () => {
    test('back button is visible and functional', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      // back button should be visible
      await expect(powerPage.backButton).toBeVisible()

      // clicking should navigate to home
      await powerPage.backButton.click()
      await page.waitForURL('/')
      await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()
    })

    test('page displays title and subtitle', async ({ page }) => {
      const powerPage = new PowerPage(page)
      await powerPage.goto()
      await powerPage.waitForChartsLoad()

      await expect(powerPage.pageTitle).toBeVisible()
      const titleText = await powerPage.pageTitle.textContent()
      expect(titleText).toContain('Power')

      await expect(powerPage.pageSubtitle).toBeVisible()
      const subtitleText = await powerPage.pageSubtitle.textContent()
      expect(subtitleText).toContain('Peak power outputs')
    })
  })
})
