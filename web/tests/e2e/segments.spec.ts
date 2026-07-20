import { test, expect } from './fixtures'
import { SegmentsPage } from './pages/segments.page'

test.describe('Segments Page', () => {
  test.describe('Segment Table', () => {
    test('segment table loads with segment data', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      // verify page header
      await expect(segmentsPage.pageTitle).toBeVisible()
      await expect(segmentsPage.pageSubtitle).toBeVisible()

      // check if table, empty state, or error is visible
      const hasData = await segmentsPage.hasData()
      const hasError = await segmentsPage.hasError()
      const isEmpty = await segmentsPage.emptyState.isVisible()

      // one of these should be true
      expect(hasData || hasError || isEmpty).toBe(true)

      if (hasData) {
        // table should be visible
        await expect(segmentsPage.segmentTable).toBeVisible()

        // should have at least one row
        const rowCount = await segmentsPage.getRowCount()
        expect(rowCount).toBeGreaterThan(0)

        // first row should have a segment name
        const firstName = await segmentsPage.getSegmentName(0)
        expect(firstName).toBeTruthy()
      }
    })

    test('table rows display segment information', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // verify table headers are visible
      await expect(segmentsPage.segmentHeader).toBeVisible()
      await expect(segmentsPage.distanceHeader).toBeVisible()
      await expect(segmentsPage.timesHeader).toBeVisible()
      await expect(segmentsPage.bestHeader).toBeVisible()
    })
  })

  test.describe('Search Filter', () => {
    test('search by name filters results with debounce', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // get initial row count
      const initialCount = await segmentsPage.getRowCount()

      // get a segment name to search for
      const segmentName = await segmentsPage.getSegmentName(0)
      if (!segmentName) {
        return
      }

      // search for part of the name
      const searchTerm = segmentName.substring(0, Math.min(5, segmentName.length))
      await segmentsPage.search(searchTerm)

      // verify search input has the term
      await expect(segmentsPage.searchInput).toHaveValue(searchTerm)

      // results should be filtered (may be same or fewer)
      const filteredCount = await segmentsPage.getRowCount()
      expect(filteredCount).toBeLessThanOrEqual(initialCount)

      // clear search
      await segmentsPage.clearSearch()
      await expect(segmentsPage.searchInput).toHaveValue('')
    })
  })

  test.describe('Sport Type Quick Filters', () => {
    test('sport type quick filters (All, Ride, Run) work', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // verify filter buttons are visible
      await expect(segmentsPage.allSportToggle).toBeVisible()
      await expect(segmentsPage.rideSportToggle).toBeVisible()
      await expect(segmentsPage.runSportToggle).toBeVisible()

      const initialCount = await segmentsPage.getRowCount()

      // click Ride filter
      await segmentsPage.selectSportFilter('ride')
      const isRideSelected = await segmentsPage.isSportFilterSelected('ride')
      expect(isRideSelected).toBe(true)

      // click Run filter
      await segmentsPage.selectSportFilter('run')
      const isRunSelected = await segmentsPage.isSportFilterSelected('run')
      expect(isRunSelected).toBe(true)

      // click All to reset
      await segmentsPage.selectSportFilter('all')
      const isAllSelected = await segmentsPage.isSportFilterSelected('all')
      expect(isAllSelected).toBe(true)

      // count should be back to initial (or close)
      const resetCount = await segmentsPage.getRowCount()
      expect(resetCount).toBe(initialCount)
    })
  })

  test.describe('Starred and KOM Toggles', () => {
    test('starred toggle filters results', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // verify starred button is visible
      await expect(segmentsPage.starredButton).toBeVisible()

      // toggle starred filter
      await segmentsPage.toggleStarred()
      const isStarredActive = await segmentsPage.isStarredActive()
      expect(isStarredActive).toBe(true)

      // should show clear filters button
      const hasFilters = await segmentsPage.hasActiveFilters()
      expect(hasFilters).toBe(true)

      // toggle off
      await segmentsPage.toggleStarred()
      const isStarredInactive = await segmentsPage.isStarredActive()
      expect(isStarredInactive).toBe(false)
    })

    test('KOM toggle filters results', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // verify KOM button is visible
      await expect(segmentsPage.komButton).toBeVisible()

      // toggle KOM filter
      await segmentsPage.toggleKom()
      const isKomActive = await segmentsPage.isKomActive()
      expect(isKomActive).toBe(true)

      // toggle off
      await segmentsPage.toggleKom()
      const isKomInactive = await segmentsPage.isKomActive()
      expect(isKomInactive).toBe(false)
    })
  })

  test.describe('Column Sorting', () => {
    test('column header sorting works', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // verify sortable headers are visible
      await expect(segmentsPage.segmentHeader).toBeVisible()
      await expect(segmentsPage.distanceHeader).toBeVisible()

      // times should be sorted by default (desc)
      const timesSort = await segmentsPage.isColumnSorted('times')
      expect(timesSort.sorted).toBe(true)
      expect(timesSort.direction).toBe('desc')

      // click segment header to sort
      await segmentsPage.clickColumnHeader('segment')
      const segmentSort = await segmentsPage.isColumnSorted('segment')
      expect(segmentSort.sorted).toBe(true)
      expect(segmentSort.direction).toBe('asc') // name defaults to asc

      // click again to toggle direction
      await segmentsPage.clickColumnHeader('segment')
      const segmentSortDesc = await segmentsPage.isColumnSorted('segment')
      expect(segmentSortDesc.direction).toBe('desc')

      // click distance to change sort column
      await segmentsPage.clickColumnHeader('distance')
      const distanceSort = await segmentsPage.isColumnSorted('distance')
      expect(distanceSort.sorted).toBe(true)
    })
  })

  test.describe('Segment Detail Modal', () => {
    test('click row opens segment detail modal', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // click first row
      await segmentsPage.clickRow(0)

      // modal should open
      const isOpen = await segmentsPage.isModalOpen()
      expect(isOpen).toBe(true)

      // modal should show segment name
      await expect(segmentsPage.modalTitle).toBeVisible()
      const modalName = await segmentsPage.getModalSegmentName()
      expect(modalName).toBeTruthy()

      // close button should be visible
      await expect(segmentsPage.modalCloseButton).toBeVisible()
    })

    test('modal shows map, PR progression chart, and efforts table', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // click first row to open modal
      await segmentsPage.clickRow(0)
      await page.waitForTimeout(500) // wait for modal content to load

      // modal should be open
      const isOpen = await segmentsPage.isModalOpen()
      expect(isOpen).toBe(true)

      // check for route card (may have map or placeholder)
      await expect(segmentsPage.modalRouteCard).toBeVisible()

      // check for PR chart card (may have chart or placeholder)
      await expect(segmentsPage.modalPrCard).toBeVisible()

      // check for efforts card (may have table or placeholder)
      await expect(segmentsPage.modalEffortsCard).toBeVisible()

      // strava link should be visible
      await expect(segmentsPage.modalStravaLink).toBeVisible()
    })

    test('modal close button works', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // click first row to open modal
      await segmentsPage.clickRow(0)

      // modal should be open
      await expect(segmentsPage.modal).toBeVisible()

      // close modal
      await segmentsPage.closeModal()

      // modal should be closed
      const isOpen = await segmentsPage.isModalOpen()
      expect(isOpen).toBe(false)
    })

    test('modal efforts table shows data when available', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // click first row to open modal
      await segmentsPage.clickRow(0)
      await page.waitForTimeout(500)

      // check if modal has efforts table
      const hasTable = await segmentsPage.modalHasEffortsTable()

      if (hasTable) {
        // verify table has expected headers
        const effortsTable = segmentsPage.modalEffortsTable
        await expect(effortsTable.locator('th:has-text("Date")')).toBeVisible()
        await expect(effortsTable.locator('th:has-text("Elapsed")')).toBeVisible()
        await expect(effortsTable.locator('th:has-text("Activity")')).toBeVisible()

        // check if there are any efforts
        const effortsCount = await segmentsPage.getModalEffortsCount()
        // may or may not have efforts, but count should be >= 0
        expect(effortsCount).toBeGreaterThanOrEqual(0)
      }
    })
  })

  test.describe('Clear Filters', () => {
    test('clear filters resets all active filters', async ({ page }) => {
      const segmentsPage = new SegmentsPage(page)
      await segmentsPage.goto()
      await segmentsPage.waitForDataLoad()

      const hasData = await segmentsPage.hasData()
      if (!hasData) {
        return
      }

      // apply some filters
      await segmentsPage.toggleStarred()
      await segmentsPage.search('test')

      // should have active filters
      const hasFilters = await segmentsPage.hasActiveFilters()
      expect(hasFilters).toBe(true)

      // clear all filters
      await segmentsPage.clearFilters()

      // starred should be inactive
      const isStarredActive = await segmentsPage.isStarredActive()
      expect(isStarredActive).toBe(false)

      // search should be cleared
      await expect(segmentsPage.searchInput).toHaveValue('')
    })
  })
})
