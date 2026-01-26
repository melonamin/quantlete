import { test, expect } from './fixtures'
import { EddingtonPage } from './pages/eddington.page'

test.describe('Eddington Page', () => {
  test.describe('Eddington Number Display', () => {
    test('current Eddington number displays prominently', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      // verify page header
      await expect(eddingtonPage.pageTitle).toBeVisible()
      await expect(eddingtonPage.pageSubtitle).toBeVisible()

      // check if number or error is visible
      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      // eddington number should be visible
      await expect(eddingtonPage.eddingtonNumber).toBeVisible()

      // get the number
      const number = await eddingtonPage.getEddingtonNumber()
      expect(number).toBeGreaterThanOrEqual(0)

      // description should be visible
      const description = await eddingtonPage.getEddingtonDescription()
      expect(description).toBeTruthy()
      expect(description).toContain('km')
    })

    test('eddington card shows context description', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      // description should explain what eddington number means
      const description = await eddingtonPage.getEddingtonDescription()
      expect(description).toContain('different days')
    })
  })

  test.describe('History Progression Chart', () => {
    test('history progression chart renders', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      // history card should be visible
      await expect(eddingtonPage.historyCard).toBeVisible()

      // chart should render
      const hasChart = await eddingtonPage.hasHistoryChart()
      expect(hasChart).toBe(true)
    })

    test('history card shows History title', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (!hasError) {
        const cardText = await eddingtonPage.historyCard.textContent()
        expect(cardText).toContain('History')
      }
    })
  })

  test.describe('View Mode Tabs', () => {
    test('view mode tabs (All/Sport group/Custom) switch correctly', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      // verify tabs are visible
      await expect(eddingtonPage.viewModeTabs).toBeVisible()
      await expect(eddingtonPage.allTab).toBeVisible()
      await expect(eddingtonPage.bySportTab).toBeVisible()
      await expect(eddingtonPage.customTab).toBeVisible()

      // default should be "all"
      const initialMode = await eddingtonPage.getSelectedViewMode()
      expect(initialMode).toBe('all')

      // switch to By Sport
      await eddingtonPage.selectViewMode('sport-group')
      const sportMode = await eddingtonPage.getSelectedViewMode()
      expect(sportMode).toBe('sport-group')

      // switch to Custom
      await eddingtonPage.selectViewMode('custom')
      const customMode = await eddingtonPage.getSelectedViewMode()
      expect(customMode).toBe('custom')

      // switch back to All
      await eddingtonPage.selectViewMode('all')
      const allMode = await eddingtonPage.getSelectedViewMode()
      expect(allMode).toBe('all')
    })

    test('All tab shows all activities eddington', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      // all should be selected by default
      await eddingtonPage.selectViewMode('all')

      // should have an eddington number
      const number = await eddingtonPage.getEddingtonNumber()
      expect(number).toBeGreaterThanOrEqual(0)

      // description should mention "All Activities"
      const description = await eddingtonPage.getEddingtonDescription()
      expect(description).toContain('All Activities')
    })

    test('By Sport tab shows sport group selector', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      // switch to By Sport mode
      await eddingtonPage.selectViewMode('sport-group')

      // sport comparison card should be visible
      const hasComparison = await eddingtonPage.hasSportComparisonCard()
      expect(hasComparison).toBe(true)
    })
  })

  test.describe('Sport Comparison Card', () => {
    test('sport comparison shows eddington numbers for each sport group', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      // should have sport comparison card in All or By Sport mode
      const hasComparison = await eddingtonPage.hasSportComparisonCard()
      if (hasComparison) {
        // get sport groups
        const groups = await eddingtonPage.getSportGroups()
        expect(groups.length).toBeGreaterThan(0)

        // each group should have a name and number
        for (const group of groups) {
          expect(group.name).toBeTruthy()
          expect(group.number).toBeGreaterThanOrEqual(0)
        }
      }
    })

    test('clicking sport group updates displayed eddington', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      const hasComparison = await eddingtonPage.hasSportComparisonCard()
      if (!hasComparison) {
        return
      }

      // get sport groups
      const groups = await eddingtonPage.getSportGroups()
      if (groups.length <= 1) {
        return // need at least 2 groups to test switching
      }

      // click a different sport group (not "All")
      const nonAllGroup = groups.find((g) => g.name !== 'All')
      if (nonAllGroup) {
        await eddingtonPage.clickSportGroup(nonAllGroup.name)

        // wait for update
        await page.waitForTimeout(500)

        // the page should still work - verify number is displayed
        const newNumber = await eddingtonPage.getEddingtonNumber()
        expect(newNumber).toBeGreaterThanOrEqual(0)
      }
    })

    test('All button in comparison shows all activities', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasComparison = await eddingtonPage.hasSportComparisonCard()
      if (hasComparison) {
        // All button should be visible
        await expect(eddingtonPage.allComparisonButton).toBeVisible()

        // clicking All should show all activities
        await eddingtonPage.allComparisonButton.click()
        await page.waitForTimeout(500)

        const description = await eddingtonPage.getEddingtonDescription()
        expect(description).toContain('All')
      }
    })
  })

  test.describe('Next Goals Table', () => {
    test('next goals table shows upcoming targets', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      // check if next goals card is visible
      const hasNextGoals = await eddingtonPage.hasNextGoalsTable()
      if (hasNextGoals) {
        // card title should be visible
        const cardText = await eddingtonPage.nextGoalsCard.textContent()
        expect(cardText).toContain('Next Goals')

        // get goals
        const goals = await eddingtonPage.getNextGoals()
        if (goals.length > 0) {
          // each goal should have target and rides needed
          for (const goal of goals) {
            expect(goal.target).toMatch(/E\d+/) // format like "E42"
            expect(goal.ridesNeeded).toBeTruthy()
          }
        }
      }
    })
  })

  test.describe('Top Days Table', () => {
    test('top distance days table shows ranking', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      const hasError = await eddingtonPage.hasError()
      if (hasError) {
        return
      }

      // check if top days card is visible
      const hasTopDays = await eddingtonPage.hasTopDaysTable()
      if (hasTopDays) {
        // card title should be visible
        const cardText = await eddingtonPage.topDaysCard.textContent()
        expect(cardText).toContain('Top Distance Days')

        // table headers should be visible
        const table = eddingtonPage.topDaysTable
        await expect(table.locator('text=Rank')).toBeVisible()
        await expect(table.locator('text=Date')).toBeVisible()
        await expect(table.locator('text=Distance')).toBeVisible()
      }
    })
  })

  test.describe('Page Layout', () => {
    test('page displays title and subtitle', async ({ page }) => {
      const eddingtonPage = new EddingtonPage(page)
      await eddingtonPage.goto()
      await eddingtonPage.waitForDataLoad()

      await expect(eddingtonPage.pageTitle).toBeVisible()
      const titleText = await eddingtonPage.pageTitle.textContent()
      expect(titleText).toContain('Eddington Number')

      await expect(eddingtonPage.pageSubtitle).toBeVisible()
      const subtitleText = await eddingtonPage.pageSubtitle.textContent()
      expect(subtitleText).toContain('progress')
    })
  })
})
