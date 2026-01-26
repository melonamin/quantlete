import { test, expect } from '../fixtures'
import { DashboardPage } from '../pages/dashboard.page'
import { PowerPage } from '../pages/power.page'
import { TrainingLoadPage } from '../pages/training-load.page'
import { BestEffortsPage } from '../pages/best-efforts.page'

// journey: analytics navigation
// dashboard -> power page -> verify charts -> training load -> best efforts -> back to dashboard
// key verification: all analytics pages load correctly and display data/charts
test.describe('Analytics Navigation Journey', () => {
  test('user can navigate through all analytics pages', async ({ page }) => {
    // step 1: start on dashboard
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.waitForPageLoad()
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()

    // step 2: navigate to power page via sidebar
    await page.getByRole('link', { name: /power/i }).click()

    // step 3: wait for power page to load
    const powerPage = new PowerPage(page)
    await powerPage.waitForChartsLoad()

    // step 4: verify power page loaded
    await expect(powerPage.pageTitle).toBeVisible()

    // step 5: check if power charts rendered (or error if no power data)
    const hasAllTimeBest = await powerPage.hasAllTimeBestChart()
    const hasPowerCurve = await powerPage.hasPowerCurveChart()
    const hasError = await powerPage.hasError()

    // either charts should be visible or there should be an informative message
    expect(hasAllTimeBest || hasPowerCurve || hasError).toBe(true)

    // step 6: if progression chart exists, test the duration selector
    const hasProgression = await powerPage.hasProgressionChart()
    if (hasProgression) {
      const options = await powerPage.getDurationOptions()
      // close dropdown
      await page.keyboard.press('Escape')

      if (options.length > 1) {
        // select a different duration
        await powerPage.selectDuration(options[1])
        await page.waitForTimeout(500)
      }
    }

    // step 7: navigate to training load page
    await page.getByRole('link', { name: /training/i }).click()

    // step 8: wait for training load page to load
    const trainingLoadPage = new TrainingLoadPage(page)
    await trainingLoadPage.waitForChartLoad()

    // step 9: verify training load page loaded
    await expect(trainingLoadPage.pageTitle).toBeVisible()

    // step 10: check for training load chart or config warning
    const hasTrainingLoadChart = await trainingLoadPage.hasTrainingLoadChart()
    const hasConfigWarning = await trainingLoadPage.hasConfigWarning()
    const hasTrainingError = await trainingLoadPage.hasError()

    // page should show something meaningful
    expect(hasTrainingLoadChart || hasConfigWarning || hasTrainingError).toBe(true)

    // step 11: check summary cards if chart is visible
    if (hasTrainingLoadChart) {
      const hasSummaryCards = await trainingLoadPage.hasSummaryCards()
      expect(hasSummaryCards).toBe(true)

      // step 12: test date range filter
      const currentRange = await trainingLoadPage.getDateRange()
      if (currentRange.after && currentRange.before) {
        // just verify the inputs work
        await trainingLoadPage.setDateRange('2024-01-01', '2024-12-31')
        await page.waitForTimeout(500)

        // restore original range
        await trainingLoadPage.setDateRange(currentRange.after, currentRange.before)
      }
    }

    // step 13: navigate to best efforts page
    await page.getByRole('link', { name: /best efforts/i }).click()

    // step 14: wait for best efforts page to load
    const bestEffortsPage = new BestEffortsPage(page)
    await bestEffortsPage.waitForDataLoad()

    // step 15: verify best efforts page loaded
    await expect(bestEffortsPage.pageTitle).toBeVisible()

    // step 16: check for PRs card
    const hasPrsCard = await bestEffortsPage.hasPrsCard()
    const hasEffortsError = await bestEffortsPage.hasError()
    expect(hasPrsCard || hasEffortsError).toBe(true)

    // step 17: if we have PRs, try the sport filter
    if (hasPrsCard) {
      const distances = await bestEffortsPage.getDistanceList()

      // test sport filter buttons
      await bestEffortsPage.selectSportFilter('runs')
      await page.waitForTimeout(500)

      const selectedFilter = await bestEffortsPage.getSelectedSportFilter()
      expect(selectedFilter).toBe('runs')

      // switch to rides
      await bestEffortsPage.selectSportFilter('rides')
      await page.waitForTimeout(500)

      // switch back to all
      await bestEffortsPage.selectSportFilter('all')
      await page.waitForTimeout(500)

      // step 18: if we have distances with data, test the modal
      const distancesWithData = distances.filter(d => d.hasData)
      if (distancesWithData.length > 0) {
        await bestEffortsPage.clickFirstDistanceWithData()
        await page.waitForTimeout(500)

        // verify modal is open
        const modalOpen = await bestEffortsPage.isModalOpen()
        expect(modalOpen).toBe(true)

        // check modal has chart or efforts table
        const modalHasChart = await bestEffortsPage.modalHasChart()
        const modalHasEfforts = await bestEffortsPage.modalHasEfforts()
        expect(modalHasChart || modalHasEfforts).toBe(true)

        // close modal
        await bestEffortsPage.closeModal()
        const modalClosed = await bestEffortsPage.isModalOpen()
        expect(modalClosed).toBe(false)
      }
    }

    // step 19: navigate back to dashboard
    await page.getByRole('link', { name: /dashboard/i }).click()

    // step 20: verify we're back on dashboard
    await dashboardPage.waitForPageLoad()
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()

    // step 21: verify dashboard still displays correctly
    await expect(page.locator('text=Total Activities').first()).toBeVisible()
  })

  test('analytics pages handle no data gracefully', async ({ page }) => {
    // step 1: navigate directly to power page
    const powerPage = new PowerPage(page)
    await powerPage.goto()

    // step 2: wait for page to load
    await powerPage.waitForChartsLoad()

    // step 3: page should not crash, should show either data or a message
    await expect(powerPage.pageTitle).toBeVisible()

    // step 4: navigate directly to training load
    const trainingLoadPage = new TrainingLoadPage(page)
    await trainingLoadPage.goto()
    await trainingLoadPage.waitForChartLoad()

    await expect(trainingLoadPage.pageTitle).toBeVisible()

    // step 5: navigate directly to best efforts
    const bestEffortsPage = new BestEffortsPage(page)
    await bestEffortsPage.goto()
    await bestEffortsPage.waitForDataLoad()

    await expect(bestEffortsPage.pageTitle).toBeVisible()
  })

  test('back button navigation works across analytics pages', async ({ page }) => {
    // step 1: start on dashboard
    await page.goto('/')
    await page.waitForTimeout(500)

    // step 2: navigate to power
    await page.getByRole('link', { name: /power/i }).click()
    await expect(page).toHaveURL(/\/power/)

    // step 3: navigate to training load
    await page.getByRole('link', { name: /training/i }).click()
    await expect(page).toHaveURL(/\/training-load/)

    // step 4: navigate to best efforts
    await page.getByRole('link', { name: /best efforts/i }).click()
    await expect(page).toHaveURL(/\/best-efforts/)

    // step 5: go back to training load
    await page.goBack()
    await expect(page).toHaveURL(/\/training-load/)

    // step 6: go back to power
    await page.goBack()
    await expect(page).toHaveURL(/\/power/)

    // step 7: go back to dashboard
    await page.goBack()
    await expect(page).toHaveURL(/\/$/)
  })
})
