import { test, expect } from '../fixtures'
import { ActivitiesPage } from '../pages/activities.page'
import { ActivityDetailPage } from '../pages/activity-detail.page'

// journey: activity exploration flow
// activities list -> apply filters -> click activity -> view segments -> return to filtered list
// key verification: filters should be preserved when navigating back
test.describe('Activity Exploration Flow Journey', () => {
  test('filters are preserved when navigating to activity and back', async ({ page }) => {
    // step 1: navigate to activities page
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForTableLoad()

    // step 2: verify the page loaded
    await expect(activitiesPage.pageTitle).toBeVisible()

    // step 3: get initial activity count
    const initialRowCount = await activitiesPage.getActivityRowCount()
    if (initialRowCount === 0) {
      // no activities to test with
      return
    }

    // step 4: apply a sport type filter (Ride)
    await activitiesPage.selectQuickSportFilter('Ride')

    // step 5: verify the Ride filter is now selected
    const rideButton = page.locator('[role="group"] button:has-text("Ride")').first()
    await expect(rideButton).toHaveAttribute('data-state', 'on')

    // step 6: get filtered count
    const filteredRowCount = await activitiesPage.getActivityRowCount()

    // step 7: if we have filtered results, click on one
    if (filteredRowCount > 0) {
      // step 8: verify filtered results are rides
      const sportType = await activitiesPage.getActivitySportType(0)
      expect(sportType?.toLowerCase()).toContain('ride')

      // step 9: click on the first activity
      await activitiesPage.clickActivityRow(0)

      // step 11: verify navigation to activity detail
      await expect(page).toHaveURL(/\/activities\/\d+/)

      // step 12: wait for activity detail to load
      const activityDetailPage = new ActivityDetailPage(page)
      await activityDetailPage.waitForActivityLoad()

      // step 13: verify the activity detail shows
      await expect(activityDetailPage.activityTitle).toBeVisible()

      // step 14: go back using browser back button
      await page.goBack()

      // step 15: verify we're back on activities page
      await expect(page).toHaveURL(/\/activities/)
      await expect(activitiesPage.pageTitle).toBeVisible()

      // step 16: wait for table to reload
      await activitiesPage.waitForTableLoad()

      // step 17: verify the Ride filter is still applied
      await expect(rideButton).toHaveAttribute('data-state', 'on')

      // step 18: verify filter state is preserved
      const stillFiltered = await rideButton.getAttribute('data-state')
      expect(stillFiltered).toBe('on')
    }
  })

  test('multiple filters are preserved across navigation', async ({ page }) => {
    // step 1: navigate to activities page
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForTableLoad()

    // step 2: check if we have activities
    const rowCount = await activitiesPage.getActivityRowCount()
    if (rowCount === 0) {
      return
    }

    // step 3: apply sport filter
    await activitiesPage.selectQuickSportFilter('Run')
    await page.waitForTimeout(500)

    // step 4: open advanced filters and set a date range
    await activitiesPage.openAdvancedFilters()
    await activitiesPage.dateFromInput.fill('2024-01-01')
    await page.waitForTimeout(600)

    // step 5: verify clear filters button appears (filters are active)
    await expect(activitiesPage.clearFiltersButton).toBeVisible()

    // step 6: check row count after filtering
    await activitiesPage.waitForTableLoad()
    const filteredCount = await activitiesPage.getActivityRowCount()

    // step 7: if we have results, click an activity
    if (filteredCount > 0) {
      await activitiesPage.clickActivityRow(0)
      await expect(page).toHaveURL(/\/activities\/\d+/)

      // step 8: wait for detail page
      const activityDetailPage = new ActivityDetailPage(page)
      await activityDetailPage.waitForActivityLoad()

      // step 9: navigate back
      await page.goBack()

      // step 10: wait for activities page to reload
      await expect(activitiesPage.pageTitle).toBeVisible()
      await activitiesPage.waitForTableLoad()

      // step 11: verify filters are still showing as active
      // the clear filters button should still be visible
      await expect(activitiesPage.clearFiltersButton).toBeVisible()
    }
  })

  test('clearing filters resets view and allows fresh exploration', async ({ page }) => {
    // step 1: navigate to activities page
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForTableLoad()

    // step 2: verify initial state
    const initialCount = await activitiesPage.getActivityRowCount()
    if (initialCount === 0) {
      return
    }

    // step 3: apply a filter
    await activitiesPage.selectQuickSportFilter('Walk')
    await page.waitForTimeout(500)

    // step 4: verify filter is active
    const walkButton = page.locator('[role="group"] button:has-text("Walk")').first()
    await expect(walkButton).toHaveAttribute('data-state', 'on')

    // step 5: clear all filters
    await activitiesPage.clearAllFilters()

    // step 6: verify Walk filter is deselected
    await expect(walkButton).toHaveAttribute('data-state', 'off')

    // step 7: verify clear filters button is gone
    await expect(activitiesPage.clearFiltersButton).not.toBeVisible()

    // step 8: verify we can still navigate to an activity
    await activitiesPage.waitForTableLoad()
    const afterClearCount = await activitiesPage.getActivityRowCount()

    if (afterClearCount > 0) {
      await activitiesPage.clickActivityRow(0)
      await expect(page).toHaveURL(/\/activities\/\d+/)

      const activityDetailPage = new ActivityDetailPage(page)
      await activityDetailPage.waitForActivityLoad()
      await expect(activityDetailPage.activityTitle).toBeVisible()
    }
  })

  test('pagination state does not interfere with activity viewing', async ({ page }) => {
    // step 1: navigate to activities page
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForTableLoad()

    // step 2: check if we have multiple pages
    const pageInfoVisible = await activitiesPage.pageInfo.isVisible()
    if (!pageInfoVisible) {
      return
    }

    const totalPages = await activitiesPage.getTotalPages()
    if (totalPages <= 1) {
      // only one page, skip pagination test
      return
    }

    // step 3: go to page 2
    await activitiesPage.goToNextPage()
    await expect(page.locator('text=/Page 2 of/')).toBeVisible()

    // step 4: get activity from page 2
    const rowCount = await activitiesPage.getActivityRowCount()
    if (rowCount > 0) {
      // step 5: click on an activity from page 2
      await activitiesPage.clickActivityRow(0)
      await expect(page).toHaveURL(/\/activities\/\d+/)

      // step 6: wait for detail page
      const activityDetailPage = new ActivityDetailPage(page)
      await activityDetailPage.waitForActivityLoad()

      // step 7: activity detail should load correctly
      await expect(activityDetailPage.activityTitle).toBeVisible()

      // step 8: navigate back
      await page.goBack()

      // step 9: verify we're back on activities
      await expect(activitiesPage.pageTitle).toBeVisible()
    }
  })
})
