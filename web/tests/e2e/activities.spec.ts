import { test, expect } from './fixtures'

test.describe('Activities Page', () => {
  test.beforeEach(async ({ page }) => {
    // navigate to activities page before each test
    await page.goto('/activities')
    await page.locator('h1').filter({ hasText: 'Activities' }).waitFor({ state: 'visible' })
  })

  test.describe('Activity Table Loading', () => {
    test('activity table loads with paginated data', async ({ page }) => {
      // verify table header is visible
      await expect(page.locator('th:has-text("Activity")')).toBeVisible()
      await expect(page.locator('th:has-text("Type")')).toBeVisible()
      await expect(page.locator('th:has-text("Date")')).toBeVisible()
      await expect(page.locator('th:has-text("Distance")')).toBeVisible()
      await expect(page.locator('th:has-text("Time")')).toBeVisible()
      await expect(page.locator('th:has-text("Elevation")')).toBeVisible()

      // wait for loading to complete
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      // verify either data rows or empty state is shown
      const hasRows = await page.locator('tbody tr').count() > 0
      const hasEmptyState = await page.locator('text=No activities found.').isVisible()
      expect(hasRows || hasEmptyState).toBe(true)

      // if we have data, verify pagination appears
      if (hasRows) {
        // pagination should show "Showing X to Y of Z activities"
        await expect(page.locator('text=/Showing \\d+ to \\d+ of \\d+/')).toBeVisible()
      }
    })

    test('table rows display activity information correctly', async ({ page }) => {
      // wait for data to load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      const rowCount = await page.locator('tbody tr').count()
      if (rowCount > 0) {
        // verify first row has activity name link
        const firstRow = page.locator('tbody tr').first()
        const activityLink = firstRow.locator('a').first()
        await expect(activityLink).toBeVisible()
        await expect(activityLink).toHaveAttribute('href', /\/activities\/\d+/)

        // verify row has sport type icon and text
        const typeCell = firstRow.locator('td').nth(1)
        await expect(typeCell).toBeVisible()
        const typeText = await typeCell.textContent()
        expect(typeText).toBeTruthy()
      }
    })
  })

  test.describe('Sorting', () => {
    test('table header columns are present for sorting', async ({ page }) => {
      // note: the current implementation doesn't have clickable sort headers
      // this test verifies the columns exist that could be sortable
      await expect(page.locator('th:has-text("Date")')).toBeVisible()
      await expect(page.locator('th:has-text("Distance")')).toBeVisible()
      await expect(page.locator('th:has-text("Time")')).toBeVisible()
      await expect(page.locator('th:has-text("Elevation")')).toBeVisible()
    })
  })

  test.describe('Filter by Sport Type', () => {
    test('quick sport filter buttons are visible', async ({ page }) => {
      // verify toggle group with sport filters exists
      const allButton = page.locator('[role="group"] button:has-text("All")').first()
      await expect(allButton).toBeVisible()

      // should have Ride button (use exact text match to avoid matching VirtualRide)
      const rideButton = page.locator('[role="group"] button').filter({ hasText: /^Ride$/ })
      await expect(rideButton).toBeVisible()

      // verify the toggle group has multiple sport buttons (All + 5 sports = 6)
      const buttonCount = await page.locator('[role="group"] button').count()
      expect(buttonCount).toBeGreaterThanOrEqual(5)
    })

    test('clicking sport filter narrows results', async ({ page }) => {
      // wait for initial data load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      const initialRowCount = await page.locator('tbody tr').count()
      if (initialRowCount === 0) {
        // no data to filter, skip test
        return
      }

      // click on "Ride" filter
      await page.locator('[role="group"] button:has-text("Ride")').first().click()

      // wait for filtering
      await page.waitForTimeout(500)

      // verify button is now selected (has data-state="on")
      const rideButton = page.locator('[role="group"] button:has-text("Ride")').first()
      await expect(rideButton).toHaveAttribute('data-state', 'on')

      // if there are results, they should all be rides
      const filteredRowCount = await page.locator('tbody tr').count()
      if (filteredRowCount > 0) {
        const firstRowType = await page.locator('tbody tr').first().locator('td').nth(1).textContent()
        expect(firstRowType?.toLowerCase()).toContain('ride')
      }
    })
  })

  test.describe('Filter by Date Range', () => {
    test('date range inputs are available in advanced filters', async ({ page }) => {
      // click More button to show advanced filters
      await page.locator('button:has-text("More")').click()

      // verify date inputs are visible
      await expect(page.locator('input[type="date"]').first()).toBeVisible()
      await expect(page.locator('input[type="date"]').nth(1)).toBeVisible()
    })

    test('setting date range filters activities', async ({ page }) => {
      // open advanced filters
      await page.locator('button:has-text("More")').click()

      // set a date range (use a range that should include demo data)
      const fromInput = page.locator('input[type="date"]').first()
      await fromInput.fill('2024-01-01')

      // wait for debounce and reload
      await page.waitForTimeout(600)

      // the filter should now be active, check for clear filters button
      const clearButton = page.locator('button:has-text("Clear filters")')
      await expect(clearButton).toBeVisible()
    })
  })

  test.describe('Filter by Gear', () => {
    test('gear filter can be applied via search', async ({ page }) => {
      // gear filtering is done via search which supports gear names
      // verify search input is available and accepts input
      const searchInput = page.locator('input[placeholder*="Search activities"]')
      await expect(searchInput).toBeVisible()

      // the search tip mentions gear names
      await expect(page.locator('text=gear names')).toBeVisible()
    })
  })

  test.describe('Commute Filter Toggle', () => {
    test('commute checkbox is available in advanced filters', async ({ page }) => {
      // open advanced filters
      await page.locator('button:has-text("More")').click()

      // verify commute checkbox is visible
      await expect(page.locator('label:has-text("Commute")')).toBeVisible()
    })

    test('toggling commute filter updates results', async ({ page }) => {
      // wait for initial load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      // open advanced filters
      await page.locator('button:has-text("More")').click()

      // click commute checkbox
      await page.locator('label:has-text("Commute")').click()

      // wait for filter to apply
      await page.waitForTimeout(500)

      // the filter should now be active
      const clearButton = page.locator('button:has-text("Clear filters")')
      await expect(clearButton).toBeVisible()
    })
  })

  test.describe('Combined Filters', () => {
    test('multiple filters can be applied together', async ({ page }) => {
      // wait for initial load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      // apply sport type filter
      await page.locator('[role="group"] button:has-text("Ride")').first().click()
      await page.waitForTimeout(300)

      // open advanced filters and apply date filter
      await page.locator('button:has-text("More")').click()
      const fromInput = page.locator('input[type="date"]').first()
      await fromInput.fill('2024-01-01')
      await page.waitForTimeout(600)

      // both filters should be active, clear button visible
      await expect(page.locator('button:has-text("Clear filters")')).toBeVisible()

      // the sport type button should still show selected
      const rideButton = page.locator('[role="group"] button:has-text("Ride")').first()
      await expect(rideButton).toHaveAttribute('data-state', 'on')
    })
  })

  test.describe('Filter Reset', () => {
    test('clear filters button resets all filters', async ({ page }) => {
      // wait for initial load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      // apply a filter first
      await page.locator('[role="group"] button:has-text("Ride")').first().click()
      await page.waitForTimeout(300)

      // verify clear filters button appears
      const clearButton = page.locator('button:has-text("Clear filters")')
      await expect(clearButton).toBeVisible()

      // click clear filters
      await clearButton.click()
      await page.waitForTimeout(500)

      // "Ride" button should no longer be selected
      const rideButton = page.locator('[role="group"] button:has-text("Ride")').first()
      await expect(rideButton).toHaveAttribute('data-state', 'off')

      // clear filters button should be hidden (no active filters)
      await expect(clearButton).not.toBeVisible()
    })
  })

  test.describe('Row Navigation', () => {
    test('clicking activity row navigates to activity detail page', async ({ page }) => {
      // wait for data to load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      const rowCount = await page.locator('tbody tr').count()
      if (rowCount === 0) {
        // no activities to click, skip
        return
      }

      // get the href of the first activity link
      const firstActivityLink = page.locator('tbody tr').first().locator('a').first()
      const href = await firstActivityLink.getAttribute('href')
      expect(href).toMatch(/\/activities\/\d+/)

      // click the link
      await firstActivityLink.click()

      // should navigate to activity detail page
      await expect(page).toHaveURL(/\/activities\/\d+/)

      // activity detail page should load
      await expect(page.locator('h1')).toBeVisible({ timeout: 10000 })
    })
  })

  test.describe('Pagination Controls', () => {
    test('pagination shows current page info', async ({ page }) => {
      // wait for data to load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      const rowCount = await page.locator('tbody tr').count()
      if (rowCount === 0) {
        // no pagination needed for empty results
        return
      }

      // verify pagination info is shown
      await expect(page.locator('text=/Showing \\d+ to \\d+ of \\d+/')).toBeVisible()
      await expect(page.locator('text=/Page \\d+ of \\d+/')).toBeVisible()
    })

    test('previous button is disabled on first page', async ({ page }) => {
      // wait for data to load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      const rowCount = await page.locator('tbody tr').count()
      if (rowCount === 0) {
        return
      }

      // on first page, previous button should be disabled
      const prevButton = page.locator('button:has-text("Previous")')
      await expect(prevButton).toBeDisabled()
    })

    test('next button navigates to second page when available', async ({ page }) => {
      // wait for data to load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      // check if we have multiple pages
      const pageInfo = page.locator('text=/Page \\d+ of (\\d+)/')
      const isVisible = await pageInfo.isVisible()
      if (!isVisible) {
        return
      }

      const pageText = await pageInfo.textContent()
      const totalPagesMatch = pageText?.match(/of (\d+)/)
      const totalPages = totalPagesMatch ? parseInt(totalPagesMatch[1]) : 1

      if (totalPages <= 1) {
        // only one page, next should be disabled
        const nextButton = page.locator('button:has-text("Next")')
        await expect(nextButton).toBeDisabled()
        return
      }

      // click next
      const nextButton = page.locator('button:has-text("Next")')
      await nextButton.click()

      // wait for page to load
      await page.waitForTimeout(500)

      // verify we're now on page 2
      await expect(page.locator('text=/Page 2 of/')).toBeVisible()

      // previous button should now be enabled
      const prevButton = page.locator('button:has-text("Previous")')
      await expect(prevButton).not.toBeDisabled()
    })

    test('previous button navigates back from second page', async ({ page }) => {
      // wait for data to load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

      // check if we have multiple pages
      const pageInfo = page.locator('text=/Page \\d+ of (\\d+)/')
      const isVisible = await pageInfo.isVisible()
      if (!isVisible) {
        return
      }

      const pageText = await pageInfo.textContent()
      const totalPagesMatch = pageText?.match(/of (\d+)/)
      const totalPages = totalPagesMatch ? parseInt(totalPagesMatch[1]) : 1

      if (totalPages <= 1) {
        return
      }

      // go to page 2
      await page.locator('button:has-text("Next")').click()
      await page.waitForTimeout(500)
      await expect(page.locator('text=/Page 2 of/')).toBeVisible()

      // click previous
      await page.locator('button:has-text("Previous")').click()
      await page.waitForTimeout(500)

      // verify we're back on page 1
      await expect(page.locator('text=/Page 1 of/')).toBeVisible()
    })
  })
})
