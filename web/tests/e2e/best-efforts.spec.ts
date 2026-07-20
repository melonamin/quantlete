import { test, expect } from './fixtures'
import { BestEffortsPage } from './pages/best-efforts.page'

test.describe('Best Efforts Page', () => {
  test.describe('Standard Distances List', () => {
    test('standard distances list renders with times or No data', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      // verify page header
      await expect(bestEffortsPage.pageTitle).toBeVisible()
      await expect(bestEffortsPage.pageSubtitle).toBeVisible()

      // check if PRs card or error is visible
      const hasCard = await bestEffortsPage.hasPrsCard()
      const hasError = await bestEffortsPage.hasError()

      // one of these should be true
      expect(hasCard || hasError).toBe(true)

      if (hasCard) {
        // PRs card should be visible
        await expect(bestEffortsPage.prsCard).toBeVisible()

        // get list of distances
        const distances = await bestEffortsPage.getDistanceList()

        // should have standard distances
        expect(distances.length).toBeGreaterThan(0)

        // each distance should have a label and time (or dash for no data)
        for (const dist of distances) {
          expect(dist.label).toBeTruthy()
          expect(dist.time).toBeTruthy()
        }
      }
    })

    test('distance rows show time and date when data exists', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      const distances = await bestEffortsPage.getDistanceList()
      const distancesWithData = distances.filter((d) => d.hasData)

      if (distancesWithData.length > 0) {
        // first distance with data should have a time (not dash)
        const first = distancesWithData[0]
        expect(first.time).not.toBe('—')
        // time should be in a duration format (e.g., "5:23" or "1:23:45")
        expect(first.time).toMatch(/\d/)
      }
    })

    test('distance list includes common distances', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      const distances = await bestEffortsPage.getDistanceList()
      const labels = distances.map((d) => d.label.toLowerCase())

      // should include some standard distances
      const standardDistances = ['1km', '5km', '10km', 'half marathon', 'marathon', '1 mile']
      const foundStandard = standardDistances.filter((std) =>
        labels.some((l) => l.includes(std.toLowerCase()))
      )

      // at least some standard distances should be present
      expect(foundStandard.length).toBeGreaterThan(0)
    })
  })

  test.describe('Sport Filter', () => {
    test('sport filter (All/Runs/Rides) updates distance list', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      // verify filter buttons are visible
      await expect(bestEffortsPage.allSportButton).toBeVisible()
      await expect(bestEffortsPage.runsSportButton).toBeVisible()
      await expect(bestEffortsPage.ridesSportButton).toBeVisible()

      // default should be runs (based on the component code)
      const initialFilter = await bestEffortsPage.getSelectedSportFilter()
      expect(initialFilter).toBe('runs')

      // switch to rides
      await bestEffortsPage.selectSportFilter('rides')

      // verify selection changed
      const ridesFilter = await bestEffortsPage.getSelectedSportFilter()
      expect(ridesFilter).toBe('rides')

      // switch to all
      await bestEffortsPage.selectSportFilter('all')
      const allFilter = await bestEffortsPage.getSelectedSportFilter()
      expect(allFilter).toBe('all')
    })

    test('filter buttons have correct styling when selected', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      // runs should be selected by default
      await expect(bestEffortsPage.runsSportButton).toHaveAttribute('data-variant', 'default')

      // all should not be selected
      await expect(bestEffortsPage.allSportButton).toHaveAttribute('data-variant', 'outline')
    })
  })

  test.describe('PR Progression Modal', () => {
    test('click distance row opens PR progression modal', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      // click first distance with data
      const clicked = await bestEffortsPage.clickFirstDistanceWithData()
      if (!clicked) {
        // no data - try clicking any row
        const distances = await bestEffortsPage.getDistanceList()
        if (distances.length > 0) {
          await bestEffortsPage.clickDistanceRow(distances[0].label)
        } else {
          return
        }
      }

      // modal should open
      const isOpen = await bestEffortsPage.isModalOpen()
      expect(isOpen).toBe(true)

      // modal title should be visible
      const title = await bestEffortsPage.getModalTitle()
      expect(title).toBeTruthy()
    })

    test('modal shows PR chart and efforts table', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      // click a distance with data
      const clicked = await bestEffortsPage.clickFirstDistanceWithData()
      if (!clicked) {
        return
      }

      // wait for modal to load
      await page.waitForTimeout(500)

      // modal should have chart section
      const hasChart = await bestEffortsPage.modalHasChart()
      expect(hasChart).toBe(true)

      // modal should have efforts table
      const hasTable = await bestEffortsPage.modalHasEffortsTable()
      expect(hasTable).toBe(true)
    })

    test('modal efforts table shows date, time, and activity link', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      const clicked = await bestEffortsPage.clickFirstDistanceWithData()
      if (!clicked) {
        return
      }

      // check if modal has efforts
      const hasEfforts = await bestEffortsPage.modalHasEfforts()
      if (hasEfforts) {
        const effortsCount = await bestEffortsPage.getModalEffortsCount()
        expect(effortsCount).toBeGreaterThan(0)

        // verify table headers are visible
        const table = bestEffortsPage.prModalEffortsTable
        await expect(table.locator('text=Date')).toBeVisible()
        await expect(table.locator('text=Time')).toBeVisible()
        await expect(table.locator('text=Activity')).toBeVisible()
      }
    })

    test('modal close button works', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      // click any distance
      const distances = await bestEffortsPage.getDistanceList()
      if (distances.length === 0) {
        return
      }
      await bestEffortsPage.clickDistanceRow(distances[0].label)

      // modal should be open
      await expect(bestEffortsPage.prModal).toBeVisible()

      // close button should be visible
      await expect(bestEffortsPage.prModalCloseButton).toBeVisible()

      // close modal
      await bestEffortsPage.closeModal()

      // modal should be closed
      const isOpen = await bestEffortsPage.isModalOpen()
      expect(isOpen).toBe(false)
    })
  })

  test.describe('Modal Chart', () => {
    test('modal shows PR progression chart with step line', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (hasError) {
        return
      }

      const clicked = await bestEffortsPage.clickFirstDistanceWithData()
      if (!clicked) {
        return
      }

      // chart container should be visible
      const hasChart = await bestEffortsPage.modalHasChart()
      expect(hasChart).toBe(true)

      // chart should render (canvas or svg)
      const chartElement = bestEffortsPage.prModalChart.locator('canvas, svg')
      await expect(chartElement.first()).toBeVisible({ timeout: 5000 })
    })
  })

  test.describe('Page Layout', () => {
    test('page displays title and subtitle', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      await expect(bestEffortsPage.pageTitle).toBeVisible()
      const titleText = await bestEffortsPage.pageTitle.textContent()
      expect(titleText).toContain('Best Efforts')

      await expect(bestEffortsPage.pageSubtitle).toBeVisible()
      const subtitleText = await bestEffortsPage.pageSubtitle.textContent()
      expect(subtitleText).toContain('personal records')
    })

    test('PRs card shows Personal Records title', async ({ page }) => {
      const bestEffortsPage = new BestEffortsPage(page)
      await bestEffortsPage.goto()
      await bestEffortsPage.waitForDataLoad()

      const hasError = await bestEffortsPage.hasError()
      if (!hasError) {
        const cardText = await bestEffortsPage.prsCard.textContent()
        expect(cardText).toContain('Personal Records')
      }
    })
  })
})
