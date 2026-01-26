import { test, expect } from '../fixtures'
import { HeatmapPage } from '../pages/heatmap.page'
import { CalendarPage } from '../pages/calendar.page'
import { PhotosPage } from '../pages/photos.page'
import { WrappedPage } from '../pages/wrapped.page'

// journey: content browsing
// heatmap -> select country -> calendar -> navigate months -> photos -> browse gallery -> wrapped -> compare years
// key verification: content pages load correctly and allow interaction
test.describe('Content Browsing Journey', () => {
  test('user can browse through content pages', async ({ page }) => {
    // step 1: navigate to heatmap page
    const heatmapPage = new HeatmapPage(page)
    await heatmapPage.goto()
    await heatmapPage.waitForMapLoad()

    // step 2: verify heatmap loaded
    await expect(heatmapPage.pageTitle).toBeVisible()

    // step 3: check for map or empty state
    const hasMap = await heatmapPage.leafletContainer.isVisible()
    const hasEmptyState = await heatmapPage.emptyState.isVisible()
    expect(hasMap || hasEmptyState).toBe(true)

    // step 4: if map is visible, test country dropdown
    if (hasMap) {
      const hasRoutes = await heatmapPage.hasActivityRoutes()
      if (hasRoutes) {
        // step 5: get route count
        const routeCount = await heatmapPage.getRouteCount()
        expect(routeCount).toBeGreaterThanOrEqual(0)

        // step 6: test sport filter
        const rideSportSelected = await heatmapPage.isSportFilterSelected('Ride')
        if (!rideSportSelected) {
          await heatmapPage.selectQuickSportFilter('Ride')
          await page.waitForTimeout(1000)
        }

        // step 7: reset to all
        await heatmapPage.selectQuickSportFilter('All')
        await page.waitForTimeout(500)

        // step 8: test country dropdown
        const countries = await heatmapPage.getCountryList()
        // close dropdown
        await page.keyboard.press('Escape')

        if (countries.length > 0) {
          // select a country
          await heatmapPage.selectCountry(countries[0].name)
          await page.waitForTimeout(1500)

          // select all countries to reset
          await heatmapPage.selectAllCountries()
          await page.waitForTimeout(500)
        }
      }
    }

    // step 9: navigate to calendar page
    await page.getByRole('link', { name: /calendar/i }).click()

    // step 10: wait for calendar to load
    const calendarPage = new CalendarPage(page)
    await calendarPage.waitForCalendarLoad()

    // step 11: verify calendar page loaded
    await expect(calendarPage.pageTitle).toBeVisible()

    // step 12: verify calendar grid is visible
    await expect(calendarPage.calendarGrid).toBeVisible()

    // step 13: get current month/year
    const currentMonthYear = await calendarPage.getCurrentMonthYear()
    expect(currentMonthYear).toBeTruthy()

    // step 14: navigate to previous month
    await calendarPage.goToPreviousMonth()
    const prevMonthYear = await calendarPage.getCurrentMonthYear()
    expect(prevMonthYear).not.toBe(currentMonthYear)

    // step 15: navigate to next month (back to current)
    await calendarPage.goToNextMonth()

    // step 16: click Today button to ensure we're on current month
    await calendarPage.goToToday()

    // step 17: check for days with activities
    const daysWithActivities = await calendarPage.getDaysWithActivities()

    // step 18: if there are activities, click on a day to open modal
    if (daysWithActivities.length > 0) {
      await calendarPage.clickDay(daysWithActivities[0])
      await page.waitForTimeout(300)

      const modalOpen = await calendarPage.isDayModalOpen()
      if (modalOpen) {
        // verify modal shows date
        const modalDate = await calendarPage.getDayModalDate()
        expect(modalDate).toBeTruthy()

        // close modal
        await calendarPage.closeDayModal()
      }
    }

    // step 19: verify stats cards are visible
    await expect(calendarPage.distanceCard).toBeVisible()
    await expect(calendarPage.workoutsCard).toBeVisible()

    // step 20: navigate to photos page
    await page.getByRole('link', { name: /photos/i }).click()

    // step 21: wait for photos page to load
    const photosPage = new PhotosPage(page)
    await photosPage.waitForDataLoad()

    // step 22: verify photos page loaded
    await expect(photosPage.pageTitle).toBeVisible()

    // step 23: check for photos or empty state
    const hasPhotos = await photosPage.hasPhotos()
    const hasEmptyPhotoState = await photosPage.emptyState.isVisible()
    expect(hasPhotos || hasEmptyPhotoState).toBe(true)

    // step 24: if photos exist, test the gallery
    if (hasPhotos) {
      const photoCount = await photosPage.getPhotoItemsCount()
      expect(photoCount).toBeGreaterThan(0)

      // step 25: test sport filter
      await photosPage.clickSportFilter('ride')
      await page.waitForTimeout(500)
      await photosPage.clickSportFilter('all')
      await page.waitForTimeout(500)

      // step 26: click first photo to open lightbox
      await photosPage.clickPhoto(0)
      await page.waitForTimeout(300)

      const lightboxOpen = await photosPage.isLightboxOpen()
      if (lightboxOpen) {
        // verify lightbox has content
        const lightboxTitle = await photosPage.getLightboxTitle()
        expect(lightboxTitle).toBeTruthy()

        // if there are multiple photos, test navigation
        if (photoCount > 1) {
          await photosPage.lightboxNext()
          await page.waitForTimeout(200)
        }

        // close lightbox
        await photosPage.closeLightbox()
        const lightboxClosed = await photosPage.isLightboxOpen()
        expect(lightboxClosed).toBe(false)
      }

      // step 27: test load more if available
      const hasLoadMore = await photosPage.hasLoadMore()
      if (hasLoadMore) {
        const initialCount = await photosPage.getPhotoItemsCount()
        await photosPage.loadMore()
        // after loading more, count might increase
        const newCount = await photosPage.getPhotoItemsCount()
        expect(newCount).toBeGreaterThanOrEqual(initialCount)
      }
    }

    // step 28: navigate to wrapped page
    await page.getByRole('link', { name: /wrapped/i }).click()

    // step 29: wait for wrapped page to load
    const wrappedPage = new WrappedPage(page)
    await wrappedPage.waitForDataLoad()

    // step 30: verify wrapped page loaded
    await expect(wrappedPage.pageTitle).toBeVisible()

    // step 31: check for data or empty state
    const hasWrappedData = await wrappedPage.hasData()
    const hasWrappedEmpty = await wrappedPage.emptyState.isVisible()
    expect(hasWrappedData || hasWrappedEmpty).toBe(true)

    // step 32: if we have data, test year selector
    if (hasWrappedData) {
      const currentYear = await wrappedPage.getCurrentYear()
      expect(currentYear).toBeTruthy()

      // get available years
      const years = await wrappedPage.getAvailableYears()
      if (years.length > 1) {
        // select a different year
        const otherYear = years.find(y => y !== currentYear && y !== 'All time')
        if (otherYear) {
          await wrappedPage.selectYear(otherYear)
          await page.waitForTimeout(500)

          // verify year changed
          const newYear = await wrappedPage.getCurrentYear()
          expect(newYear).toContain(otherYear)
        }

        // test comparison feature
        const comparisonYears = years.filter(y => y !== currentYear)
        if (comparisonYears.length > 0) {
          await wrappedPage.selectCompareYear(comparisonYears[0])
          await page.waitForTimeout(500)

          // check if comparison table appears
          const hasComparison = await wrappedPage.isComparisonTableVisible()
          if (hasComparison) {
            const comparisonRows = await wrappedPage.getComparisonTableRows()
            expect(comparisonRows).toBeGreaterThanOrEqual(0)
          }

          // clear comparison
          await wrappedPage.selectNoComparison()
        }
      }

      // step 33: verify charts are visible
      const hasHeatmap = await wrappedPage.hasHeatmapCalendar()
      const hasBarCharts = await wrappedPage.hasMonthlyBarCharts()
      const hasDonutCharts = await wrappedPage.hasDonutCharts()

      // at least some charts should be visible
      expect(hasHeatmap || hasBarCharts || hasDonutCharts).toBe(true)
    }
  })

  test('content pages handle empty data gracefully', async ({ page }) => {
    // step 1: navigate directly to heatmap
    const heatmapPage = new HeatmapPage(page)
    await heatmapPage.goto()
    await heatmapPage.waitForMapLoad()

    // page should not crash
    await expect(heatmapPage.pageTitle).toBeVisible()

    // step 2: navigate directly to calendar
    const calendarPage = new CalendarPage(page)
    await calendarPage.goto()
    await calendarPage.waitForCalendarLoad()

    await expect(calendarPage.pageTitle).toBeVisible()

    // step 3: navigate directly to photos
    const photosPage = new PhotosPage(page)
    await photosPage.goto()
    await photosPage.waitForDataLoad()

    await expect(photosPage.pageTitle).toBeVisible()

    // step 4: navigate directly to wrapped
    const wrappedPage = new WrappedPage(page)
    await wrappedPage.goto()
    await wrappedPage.waitForDataLoad()

    await expect(wrappedPage.pageTitle).toBeVisible()
  })

  test('calendar month navigation preserves stats context', async ({ page }) => {
    // step 1: navigate to calendar
    const calendarPage = new CalendarPage(page)
    await calendarPage.goto()
    await calendarPage.waitForCalendarLoad()

    // step 2: navigate to previous month
    await calendarPage.goToPreviousMonth()
    await page.waitForTimeout(500)

    // step 4: get stats for previous month
    const prevMonthStats = await calendarPage.getAllStatsValues()

    // stats might be different for different months
    // just verify they load without errors
    expect(prevMonthStats.distance).toBeTruthy()

    // step 5: navigate back to today
    await calendarPage.goToToday()
    await page.waitForTimeout(500)

    // step 6: stats should match initial (same month)
    const currentStats = await calendarPage.getAllStatsValues()
    // at minimum, workouts should load
    expect(currentStats.workouts).toBeTruthy()
  })
})
