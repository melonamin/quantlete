import { test, expect } from './fixtures'
import { WrappedPage } from './pages/wrapped.page'

test.describe('Wrapped Page', () => {
  test.describe('Year Selector', () => {
    test('year selector dropdown changes displayed data', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      // verify page header
      await expect(wrappedPage.pageTitle).toBeVisible()
      await expect(wrappedPage.pageSubtitle).toBeVisible()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // verify year selector is visible
      await expect(wrappedPage.yearSelectorTrigger).toBeVisible()

      // get current year
      const currentYear = await wrappedPage.getCurrentYear()
      expect(currentYear).toBeTruthy()

      // get available years
      const years = await wrappedPage.getAvailableYears()
      expect(years.length).toBeGreaterThan(0)

      // should have "All time" option
      expect(years).toContain('All time')

      // if we have more than one year, try selecting a different one
      if (years.length > 1) {
        const otherYear = years.find((y) => y !== currentYear && y !== 'All time')
        if (otherYear) {
          await wrappedPage.selectYear(otherYear)
          const newYear = await wrappedPage.getCurrentYear()
          expect(newYear).toContain(otherYear)
        }
      }
    })

    test('all time option shows aggregate data', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // select all time
      await wrappedPage.selectAllTime()

      // verify selection
      const currentYear = await wrappedPage.getCurrentYear()
      expect(currentYear).toContain('All time')

      // should still show metric cards
      const stillHasData = await wrappedPage.hasData()
      expect(stillHasData).toBe(true)
    })
  })

  test.describe('Metric Cards', () => {
    test('metric cards display correct stats for selected year', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // verify all metric cards are visible
      await expect(wrappedPage.activitiesCard).toBeVisible()
      await expect(wrappedPage.distanceCard).toBeVisible()
      await expect(wrappedPage.elevationCard).toBeVisible()
      await expect(wrappedPage.movingTimeCard).toBeVisible()
      await expect(wrappedPage.kudosCard).toBeVisible()
      await expect(wrappedPage.carbonCard).toBeVisible()

      // verify metric values are present
      const activities = await wrappedPage.getMetricValue('activities')
      expect(activities).toBeTruthy()

      const distance = await wrappedPage.getMetricValue('distance')
      expect(distance).toBeTruthy()

      const elevation = await wrappedPage.getMetricValue('elevation')
      expect(elevation).toBeTruthy()
    })

    test('metric cards show formatted values', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // distance should have unit (km or mi)
      const distance = await wrappedPage.getMetricValue('distance')
      expect(distance).toMatch(/km|mi/)

      // elevation should have unit (m or ft)
      const elevation = await wrappedPage.getMetricValue('elevation')
      expect(elevation).toMatch(/m|ft/)

      // moving time should be formatted as duration
      const movingTime = await wrappedPage.getMetricValue('movingTime')
      expect(movingTime).toMatch(/h|m|d/)
    })
  })

  test.describe('Charts', () => {
    test('all charts render (heatmap calendar, monthly bars, donut charts)', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // check for year heatmap (only visible for specific year, not all time)
      const currentYear = await wrappedPage.getCurrentYear()
      if (!currentYear?.includes('All time')) {
        const hasHeatmap = await wrappedPage.hasHeatmapCalendar()
        expect(hasHeatmap).toBe(true)

        // check for monthly bar charts
        const hasMonthlyCharts = await wrappedPage.hasMonthlyBarCharts()
        expect(hasMonthlyCharts).toBe(true)

        // verify monthly chart cards are visible
        await expect(wrappedPage.activitiesByMonthCard).toBeVisible()
        await expect(wrappedPage.distanceByMonthCard).toBeVisible()
        await expect(wrappedPage.elevationByMonthCard).toBeVisible()
        await expect(wrappedPage.prsByMonthCard).toBeVisible()
      }

      // check for donut charts (summary section)
      const hasDonutCharts = await wrappedPage.hasDonutCharts()
      expect(hasDonutCharts).toBe(true)

      // verify summary card is visible
      await expect(wrappedPage.summaryCard).toBeVisible()
    })

    test('start times chart renders', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // start times card should be visible
      await expect(wrappedPage.startTimesCard).toBeVisible()
    })

    test('locations chart renders', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // locations card should be visible
      await expect(wrappedPage.locationsCard).toBeVisible()
    })
  })

  test.describe('Comparison Year Selector', () => {
    test('comparison year selector shows delta comparison table', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // verify compare selector is visible
      await expect(wrappedPage.compareSelectorTrigger).toBeVisible()

      // comparison table should not be visible initially (no comparison)
      const isTableVisible = await wrappedPage.isComparisonTableVisible()
      expect(isTableVisible).toBe(false)

      // select a comparison year (e.g., "Compare to all time")
      await wrappedPage.openCompareSelector()

      // find compare to all time option
      const allTimeOption = page.locator('[role="option"]:has-text("Compare to all time")')
      if (await allTimeOption.isVisible()) {
        await allTimeOption.click()
        await page.waitForTimeout(500)
        await wrappedPage.waitForLoadingComplete()

        // comparison table should now be visible
        const isTableVisibleNow = await wrappedPage.isComparisonTableVisible()
        expect(isTableVisibleNow).toBe(true)

        // table should have rows
        const rowCount = await wrappedPage.getComparisonTableRows()
        expect(rowCount).toBeGreaterThan(0)
      }
    })

    test('comparison table shows current vs baseline values', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // select comparison
      await wrappedPage.openCompareSelector()
      const allTimeOption = page.locator('[role="option"]:has-text("Compare to all time")')
      if (await allTimeOption.isVisible()) {
        await allTimeOption.click()
        await page.waitForTimeout(500)
        await wrappedPage.waitForLoadingComplete()

        // verify table has expected columns
        const table = wrappedPage.comparisonTable
        await expect(table.locator('text=Current')).toBeVisible()
        await expect(table.locator('text=Baseline')).toBeVisible()
        await expect(table.locator('text=Δ')).toBeVisible()
      }
    })

    test('no comparison hides comparison table', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // select comparison first
      await wrappedPage.openCompareSelector()
      const allTimeOption = page.locator('[role="option"]:has-text("Compare to all time")')
      if (await allTimeOption.isVisible()) {
        await allTimeOption.click()
        await page.waitForTimeout(500)

        // verify table is visible
        let isVisible = await wrappedPage.isComparisonTableVisible()
        expect(isVisible).toBe(true)

        // select no comparison
        await wrappedPage.selectNoComparison()

        // table should be hidden
        isVisible = await wrappedPage.isComparisonTableVisible()
        expect(isVisible).toBe(false)
      }
    })
  })

  test.describe('Biggest Cards', () => {
    test('biggest activity cards display when data exists', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const hasData = await wrappedPage.hasData()
      if (!hasData) {
        return
      }

      // check if biggest cards are visible
      const longestDistanceVisible = await wrappedPage.longestDistanceCard.isVisible()
      const mostElevationVisible = await wrappedPage.mostElevationCard.isVisible()
      const longestDurationVisible = await wrappedPage.longestDurationCard.isVisible()

      // at least one should be visible (if there's data)
      expect(longestDistanceVisible || mostElevationVisible || longestDurationVisible).toBe(true)
    })
  })

  test.describe('Empty State', () => {
    test('empty state shows helpful message when no data', async ({ page }) => {
      const wrappedPage = new WrappedPage(page)
      await wrappedPage.goto()
      await wrappedPage.waitForDataLoad()

      const isEmpty = await wrappedPage.emptyState.isVisible()

      if (isEmpty) {
        await expect(wrappedPage.emptyState).toContainText('No activity data available')
      }
    })
  })
})
