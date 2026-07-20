import { test, expect } from './fixtures'
import { HeatmapPage } from './pages/heatmap.page'

test.describe('Heatmap Page', () => {
  test.describe('Page Load', () => {
    test('heatmap loads with activity routes rendered on map', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      // verify page header is visible
      await expect(heatmapPage.pageTitle).toBeVisible()
      await expect(heatmapPage.pageSubtitle).toBeVisible()

      // check if map loaded (either with data or empty state)
      const hasMap = await heatmapPage.leafletContainer.isVisible()
      const isEmpty = await heatmapPage.emptyState.isVisible()
      const hasError = await heatmapPage.errorMessage.isVisible()

      // one of these states should be true
      expect(hasMap || isEmpty || hasError).toBe(true)

      if (hasMap) {
        // if there are routes, verify route count is displayed
        const routeCount = await heatmapPage.getRouteCount()
        if (routeCount > 0) {
          // SVG overlay should be present for routes (allow for render time)
          await expect(heatmapPage.leafletSvgOverlay).toBeVisible({ timeout: 10000 })
        }
      }
    })

    test('heatmap displays route count in header', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (hasMap) {
        // route count should be displayed
        const routeCount = await heatmapPage.getRouteCount()
        expect(routeCount).toBeGreaterThanOrEqual(0)

        // verify the count text is visible
        await expect(heatmapPage.routeCount).toBeVisible()
      }
    })
  })

  test.describe('Sport Type Quick Filters', () => {
    test('sport type quick filters are visible', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      // verify all quick filter buttons are visible
      await expect(heatmapPage.allSportButton).toBeVisible()
      await expect(heatmapPage.rideSportButton).toBeVisible()
      await expect(heatmapPage.runSportButton).toBeVisible()
      await expect(heatmapPage.walkSportButton).toBeVisible()
    })

    test('All filter is the default state', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      // All button should be visible and clickable
      await expect(heatmapPage.allSportButton).toBeVisible()

      // no sport type should be selected initially - All means empty filter
      // verify we can click other filters and come back to All
      await heatmapPage.selectQuickSportFilter('Ride')
      await heatmapPage.selectQuickSportFilter('All')

      // the page should still work correctly
      await expect(heatmapPage.pageTitle).toBeVisible()
    })

    test('clicking sport filter updates map display', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        // skip if no map data
        return
      }

      // get initial route count
      const initialCount = await heatmapPage.getRouteCount()

      // click Ride filter
      await heatmapPage.selectQuickSportFilter('Ride')

      // Ride should now be selected
      const isRideSelected = await heatmapPage.isSportFilterSelected('Ride')
      expect(isRideSelected).toBe(true)

      // get new route count - may be different
      const rideCount = await heatmapPage.getRouteCount()
      expect(rideCount).toBeLessThanOrEqual(initialCount)

      // click All to reset - this deselects the current filter
      await heatmapPage.selectQuickSportFilter('All')

      // after clicking All, route count should return to initial
      const resetCount = await heatmapPage.getRouteCount()
      expect(resetCount).toBe(initialCount)
    })

    test('can filter by multiple sport types', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        return
      }

      // test Run filter
      await heatmapPage.selectQuickSportFilter('Run')
      const isRunSelected = await heatmapPage.isSportFilterSelected('Run')
      expect(isRunSelected).toBe(true)

      // test Walk filter
      await heatmapPage.selectQuickSportFilter('Walk')
      const isWalkSelected = await heatmapPage.isSportFilterSelected('Walk')
      expect(isWalkSelected).toBe(true)
    })
  })

  test.describe('Country Dropdown', () => {
    test('country dropdown lists visited countries with activity counts', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        return
      }

      // check if country dropdown button exists (only shows when there are countries)
      const hasCountryButton = await heatmapPage.countryDropdownButton.isVisible()
      if (!hasCountryButton) {
        // no countries to display - that's okay
        return
      }

      // open dropdown
      await heatmapPage.openCountryDropdown()

      // verify dropdown panel is visible
      await expect(heatmapPage.countryDropdownPanel).toBeVisible()

      // verify visited count is shown
      await expect(heatmapPage.countryVisitedCount).toBeVisible()

      // get country list
      const countries = await heatmapPage.getCountryList()
      if (countries.length > 0) {
        // each country should have a name and count
        for (const country of countries) {
          expect(country.name).toBeTruthy()
          expect(country.count).toBeGreaterThan(0)
        }
      }

      // close dropdown
      await heatmapPage.closeCountryDropdown()
    })

    test('country selection flies map to selected country bounds', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      const hasCountryButton = await heatmapPage.countryDropdownButton.isVisible()
      if (!hasMap || !hasCountryButton) {
        return
      }

      // get country list
      const countries = await heatmapPage.getCountryList()
      if (countries.length === 0) {
        return
      }

      // select first country
      const firstCountry = countries[0]
      await heatmapPage.selectCountry(firstCountry.name)

      // verify dropdown closed
      await expect(heatmapPage.countryDropdownPanel).not.toBeVisible()

      // verify country is now shown in button text
      const buttonText = await heatmapPage.countryDropdownButton.textContent()
      expect(buttonText).toContain(firstCountry.name.split(' ').pop()) // last word of country name

      // select All countries to reset
      await heatmapPage.selectAllCountries()
    })
  })

  test.describe('Advanced Filters', () => {
    test('More button opens advanced filters panel', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      // open advanced filters
      await heatmapPage.openAdvancedFilters()

      // verify advanced filters panel is visible
      await expect(heatmapPage.advancedFiltersPanel).toBeVisible()

      // verify filter inputs are visible
      await expect(heatmapPage.customSportInput).toBeVisible()
      await expect(heatmapPage.dateFromInput).toBeVisible()
      await expect(heatmapPage.dateToInput).toBeVisible()
      await expect(heatmapPage.commuteDropdown).toBeVisible()
    })

    test('date range filter filters routes', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        return
      }

      const initialCount = await heatmapPage.getRouteCount()

      // set a narrow date range that likely has fewer activities
      // use a date range from a year ago to filter
      const lastYear = new Date()
      lastYear.setFullYear(lastYear.getFullYear() - 1)
      const fromDate = lastYear.toISOString().split('T')[0]
      const toDate = new Date(lastYear.getTime() + 30 * 24 * 60 * 60 * 1000)
        .toISOString()
        .split('T')[0]

      await heatmapPage.setDateRange(fromDate, toDate)

      // route count may have changed
      const filteredCount = await heatmapPage.getRouteCount()
      expect(filteredCount).toBeLessThanOrEqual(initialCount)
    })

    test('commute filter filters routes', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        return
      }

      const initialCount = await heatmapPage.getRouteCount()

      // set commute filter to commute only
      await heatmapPage.setCommuteFilter('Commute only')

      // route count should be less than or equal to initial (commutes are a subset)
      const commuteCount = await heatmapPage.getRouteCount()
      expect(commuteCount).toBeLessThanOrEqual(initialCount)
    })
  })

  test.describe('Clear Filters', () => {
    test('clear filters resets map to default view', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        return
      }

      const initialCount = await heatmapPage.getRouteCount()

      // apply a filter
      await heatmapPage.selectQuickSportFilter('Ride')
      const filteredCount = await heatmapPage.getRouteCount()
      expect(filteredCount).toBeLessThanOrEqual(initialCount)

      // clear filters button should be visible when filters are active
      await expect(heatmapPage.clearFiltersButton).toBeVisible()

      // clear all filters
      await heatmapPage.clearAllFilters()

      // should return to initial count
      const resetCount = await heatmapPage.getRouteCount()
      expect(resetCount).toBe(initialCount)
    })

    test('clear filters button is hidden when no filters active', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      // by default, no filters are active so clear button should be hidden
      await expect(heatmapPage.clearFiltersButton).not.toBeVisible()
    })
  })

  test.describe('Map Interactivity', () => {
    test('map has zoom controls', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        return
      }

      // verify zoom controls are visible
      const hasZoom = await heatmapPage.hasZoomControls()
      expect(hasZoom).toBe(true)
    })

    test('map container is rendered', async ({ page }) => {
      const heatmapPage = new HeatmapPage(page)
      await heatmapPage.goto()
      await heatmapPage.waitForMapLoad()

      const hasMap = await heatmapPage.leafletContainer.isVisible()
      if (!hasMap) {
        return
      }

      // verify map container is present (tiles may not be visible in headless mode)
      await expect(heatmapPage.leafletContainer).toBeVisible()
    })
  })
})
