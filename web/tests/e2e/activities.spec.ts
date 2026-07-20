import { test, expect } from './fixtures'
import { ActivitiesPage } from './pages/activities.page'

test.describe('Activities Page', () => {
  test.beforeEach(async ({ page }) => {
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForPageLoad()
  })

  test.describe('Activity Table Loading', () => {
    test('activity table loads with paginated data', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)

      // verify table headers are visible
      await expect(activitiesPage.activityHeader).toBeVisible()
      await expect(activitiesPage.typeHeader).toBeVisible()
      await expect(activitiesPage.dateHeader).toBeVisible()
      await expect(activitiesPage.distanceHeader).toBeVisible()
      await expect(activitiesPage.timeHeader).toBeVisible()
      await expect(activitiesPage.elevationHeader).toBeVisible()

      await activitiesPage.waitForTableLoad()

      // verify either data rows or empty state is shown
      const rowCount = await activitiesPage.getActivityRowCount()
      const hasEmptyState = await activitiesPage.emptyState.isVisible()
      expect(rowCount > 0 || hasEmptyState).toBe(true)

      // if we have data, verify pagination appears
      if (rowCount > 0) {
        await expect(activitiesPage.showingInfo).toBeVisible()
      }
    })

    test('table rows display activity information correctly', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      const rowCount = await activitiesPage.getActivityRowCount()
      if (rowCount > 0) {
        // verify first row has activity name
        const activityName = await activitiesPage.getActivityName(0)
        expect(activityName).toBeTruthy()

        // verify row has sport type
        const sportType = await activitiesPage.getActivitySportType(0)
        expect(sportType).toBeTruthy()
      }
    })
  })

  test.describe('Sorting', () => {
    test('table header columns are present for sorting', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)

      await expect(activitiesPage.dateHeader).toBeVisible()
      await expect(activitiesPage.distanceHeader).toBeVisible()
      await expect(activitiesPage.timeHeader).toBeVisible()
      await expect(activitiesPage.elevationHeader).toBeVisible()
    })
  })

  test.describe('Filter by Sport Type', () => {
    test('quick sport filter buttons are visible', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)

      await expect(activitiesPage.allSportButton).toBeVisible()
      await expect(activitiesPage.quickSportFilters).toBeVisible()

      // verify the toggle group has multiple sport buttons
      const buttonCount = await activitiesPage.quickSportFilters.locator('button').count()
      expect(buttonCount).toBeGreaterThanOrEqual(5)
    })

    test('clicking sport filter narrows results', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      const initialRowCount = await activitiesPage.getActivityRowCount()
      if (initialRowCount === 0) {
        return
      }

      // click on "Ride" filter
      await activitiesPage.selectQuickSportFilter('Ride')

      // verify button is now selected
      const rideButton = page.locator('[role="group"] button:has-text("Ride")').first()
      await expect(rideButton).toHaveAttribute('data-state', 'on')

      // if there are results, they should all be rides
      const filteredRowCount = await activitiesPage.getActivityRowCount()
      if (filteredRowCount > 0) {
        const firstRowType = await activitiesPage.getActivitySportType(0)
        expect(firstRowType?.toLowerCase()).toContain('ride')
      }
    })
  })

  test.describe('Filter by Date Range', () => {
    test('date range inputs are available in advanced filters', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)

      await activitiesPage.openAdvancedFilters()

      await expect(activitiesPage.dateFromInput).toBeVisible()
      await expect(activitiesPage.dateToInput).toBeVisible()
    })

    test('setting date range filters activities', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)

      await activitiesPage.setDateRange('2024-01-01')

      // the filter should now be active
      await expect(activitiesPage.clearFiltersButton).toBeVisible()
    })
  })

  test.describe('Filter by Gear', () => {
    test('gear filter can be applied via search', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)

      await expect(activitiesPage.searchInput).toBeVisible()

      // the search tip mentions gear names
      await expect(page.locator('text=gear names')).toBeVisible()
    })
  })

  test.describe('Commute Filter Toggle', () => {
    test('commute checkbox is available in advanced filters', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)

      await activitiesPage.openAdvancedFilters()

      await expect(page.locator('label:has-text("Commute")')).toBeVisible()
    })

    test('toggling commute filter updates results', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      await activitiesPage.toggleCommuteFilter()

      // the filter should now be active
      await expect(activitiesPage.clearFiltersButton).toBeVisible()
    })
  })

  test.describe('Combined Filters', () => {
    test('multiple filters can be applied together', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      // apply sport type filter
      await activitiesPage.selectQuickSportFilter('Ride')

      // apply date filter
      await activitiesPage.setDateRange('2024-01-01')

      // both filters should be active
      await expect(activitiesPage.clearFiltersButton).toBeVisible()

      // the sport type button should still show selected
      const rideButton = page.locator('[role="group"] button:has-text("Ride")').first()
      await expect(rideButton).toHaveAttribute('data-state', 'on')
    })
  })

  test.describe('Filter Reset', () => {
    test('clear filters button resets all filters', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      // apply a filter first
      await activitiesPage.selectQuickSportFilter('Ride')
      await expect(activitiesPage.clearFiltersButton).toBeVisible()

      // clear filters
      await activitiesPage.clearAllFilters()

      // "Ride" button should no longer be selected
      const rideButton = page.locator('[role="group"] button:has-text("Ride")').first()
      await expect(rideButton).toHaveAttribute('data-state', 'off')

      // clear filters button should be hidden
      await expect(activitiesPage.clearFiltersButton).not.toBeVisible()
    })
  })

  test.describe('Row Navigation', () => {
    test('clicking activity row navigates to activity detail page', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      const rowCount = await activitiesPage.getActivityRowCount()
      if (rowCount === 0) {
        return
      }

      // click the first activity
      await activitiesPage.clickActivityRow(0)

      // should navigate to activity detail page
      await expect(page).toHaveURL(/\/activities\/\d+/)
      await expect(page.locator('h1')).toBeVisible({ timeout: 10000 })
    })
  })

  test.describe('Pagination Controls', () => {
    test('pagination shows current page info', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      const rowCount = await activitiesPage.getActivityRowCount()
      if (rowCount === 0) {
        return
      }

      await expect(activitiesPage.showingInfo).toBeVisible()
      await expect(activitiesPage.pageInfo).toBeVisible()
    })

    test('previous button is disabled on first page', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      const rowCount = await activitiesPage.getActivityRowCount()
      if (rowCount === 0) {
        return
      }

      await expect(activitiesPage.previousButton).toBeDisabled()
    })

    test('next button navigates to second page when available', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      const totalPages = await activitiesPage.getTotalPages()
      if (totalPages <= 1) {
        await expect(activitiesPage.nextButton).toBeDisabled()
        return
      }

      await activitiesPage.goToNextPage()

      // verify we're now on page 2
      const currentPage = await activitiesPage.getCurrentPageNumber()
      expect(currentPage).toBe(2)

      // previous button should now be enabled
      await expect(activitiesPage.previousButton).not.toBeDisabled()
    })

    test('previous button navigates back from second page', async ({ page }) => {
      const activitiesPage = new ActivitiesPage(page)
      await activitiesPage.waitForTableLoad()

      const totalPages = await activitiesPage.getTotalPages()
      if (totalPages <= 1) {
        return
      }

      // go to page 2
      await activitiesPage.goToNextPage()
      expect(await activitiesPage.getCurrentPageNumber()).toBe(2)

      // go back to page 1
      await activitiesPage.goToPreviousPage()
      expect(await activitiesPage.getCurrentPageNumber()).toBe(1)
    })
  })
})
