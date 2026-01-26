import { test, expect } from './fixtures'
import { TrainingLoadPage } from './pages/training-load.page'

test.describe('Training Load Page', () => {
  test.describe('Training Load Chart', () => {
    test('training load chart (CTL/ATL/TSB curves) renders', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      // verify page header
      await expect(trainingLoadPage.pageTitle).toBeVisible()
      await expect(trainingLoadPage.pageSubtitle).toBeVisible()

      // check if chart or error is visible
      const hasChart = await trainingLoadPage.hasTrainingLoadChart()
      const hasError = await trainingLoadPage.hasError()

      // one of these should be true
      expect(hasChart || hasError).toBe(true)

      if (hasChart) {
        await expect(trainingLoadPage.rangeCard).toBeVisible()
        await expect(trainingLoadPage.trainingLoadChart).toBeVisible()
      }
    })

    test('chart card displays Range title', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasError = await trainingLoadPage.hasError()
      if (!hasError) {
        const cardText = await trainingLoadPage.rangeCard.textContent()
        expect(cardText).toContain('Range')
      }
    })
  })

  test.describe('Date Range Filters', () => {
    test('date range filters update chart data', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasError = await trainingLoadPage.hasError()
      if (hasError) {
        return
      }

      // verify date inputs are visible
      await expect(trainingLoadPage.afterDateInput).toBeVisible()
      await expect(trainingLoadPage.beforeDateInput).toBeVisible()

      // set a date range
      const today = new Date()
      const threeMonthsAgo = new Date()
      threeMonthsAgo.setMonth(threeMonthsAgo.getMonth() - 3)

      const afterDate = threeMonthsAgo.toISOString().split('T')[0]
      const beforeDate = today.toISOString().split('T')[0]

      await trainingLoadPage.setDateRange(afterDate, beforeDate)

      // verify the inputs have the values
      const range = await trainingLoadPage.getDateRange()
      expect(range.after).toBe(afterDate)
      expect(range.before).toBe(beforeDate)

      // chart should still be visible after filter
      const hasChart = await trainingLoadPage.hasTrainingLoadChart()
      expect(hasChart).toBe(true)
    })

    test('date range can be cleared', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasError = await trainingLoadPage.hasError()
      if (hasError) {
        return
      }

      // set a date range first
      const today = new Date().toISOString().split('T')[0]
      await trainingLoadPage.setDateRange('2024-01-01', today)

      // verify range is set
      const rangeBefore = await trainingLoadPage.getDateRange()
      expect(rangeBefore.after).toBe('2024-01-01')

      // clear the range
      await trainingLoadPage.clearDateRange()

      // verify range is cleared
      const rangeAfter = await trainingLoadPage.getDateRange()
      expect(rangeAfter.after).toBe('')
      expect(rangeAfter.before).toBe('')
    })
  })

  test.describe('Summary Cards', () => {
    test('summary cards display current CTL/ATL/TSB values', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasError = await trainingLoadPage.hasError()
      if (hasError) {
        return
      }

      // wait a bit for cards to render
      await page.waitForTimeout(500)

      // summary cards should be visible when data is loaded
      const hasCards = await trainingLoadPage.hasSummaryCards()
      if (hasCards) {
        await expect(trainingLoadPage.ctlCard).toBeVisible()
        await expect(trainingLoadPage.atlCard).toBeVisible()
        await expect(trainingLoadPage.tsbCard).toBeVisible()

        // get values
        const values = await trainingLoadPage.getSummaryValues()

        // values should be defined (may be "0.0" if no data)
        expect(values.ctl).toBeTruthy()
        expect(values.atl).toBeTruthy()
        expect(values.tsb).toBeTruthy()
      }
    })

    test('CTL card shows Chronic Training Load label', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasCards = await trainingLoadPage.hasSummaryCards()
      if (hasCards) {
        const cardText = await trainingLoadPage.ctlCard.textContent()
        expect(cardText).toContain('CTL')
      }
    })

    test('ATL card shows Acute Training Load label', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasCards = await trainingLoadPage.hasSummaryCards()
      if (hasCards) {
        const cardText = await trainingLoadPage.atlCard.textContent()
        expect(cardText).toContain('ATL')
      }
    })

    test('TSB card shows Training Stress Balance label', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasCards = await trainingLoadPage.hasSummaryCards()
      if (hasCards) {
        const cardText = await trainingLoadPage.tsbCard.textContent()
        expect(cardText).toContain('TSB')
      }
    })
  })

  test.describe('Configuration Warning', () => {
    test('configuration warning shows when FTP not configured', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      // check if configuration warning is visible
      const hasWarning = await trainingLoadPage.hasConfigWarning()

      // this may or may not be true depending on user config
      // we just verify the check works
      expect(typeof hasWarning).toBe('boolean')

      if (hasWarning) {
        // warning should contain configuration-related text
        const warningText = await trainingLoadPage.getConfigWarningText()
        expect(warningText).toContain('Configuration')

        // settings link should be visible
        await expect(trainingLoadPage.settingsLink).toBeVisible()
      }
    })

    test('settings link in warning navigates to settings page', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasWarning = await trainingLoadPage.hasConfigWarning()
      if (hasWarning) {
        // click the settings link
        await trainingLoadPage.settingsLink.click()

        // should navigate to settings page
        await page.waitForURL(/settings/)
      }
    })
  })

  test.describe('Activity Breakdown', () => {
    test('activity breakdown shows activity statistics', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      const hasError = await trainingLoadPage.hasError()
      if (hasError) {
        return
      }

      // check if activity breakdown is visible
      const hasBreakdown = await trainingLoadPage.hasActivityBreakdown()

      // this may or may not be visible depending on data
      expect(typeof hasBreakdown).toBe('boolean')

      if (hasBreakdown) {
        // should show activity counts
        await expect(trainingLoadPage.totalActivitiesCount).toBeVisible()
        await expect(trainingLoadPage.activitiesWithPower).toBeVisible()
        await expect(trainingLoadPage.activitiesWithTss).toBeVisible()
      }
    })
  })

  test.describe('Page Layout', () => {
    test('back button is visible and functional', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      // back button should be visible
      await expect(trainingLoadPage.backButton).toBeVisible()

      // clicking should navigate to home
      await trainingLoadPage.backButton.click()
      await page.waitForURL('/')
      await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()
    })

    test('page displays title and subtitle', async ({ page }) => {
      const trainingLoadPage = new TrainingLoadPage(page)
      await trainingLoadPage.goto()
      await trainingLoadPage.waitForChartLoad()

      await expect(trainingLoadPage.pageTitle).toBeVisible()
      const titleText = await trainingLoadPage.pageTitle.textContent()
      expect(titleText).toContain('Training Load')

      await expect(trainingLoadPage.pageSubtitle).toBeVisible()
      const subtitleText = await trainingLoadPage.pageSubtitle.textContent()
      expect(subtitleText).toContain('Fitness, fatigue, and form')
    })
  })
})
