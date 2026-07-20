import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the wrapped (year in review) page
// provides selectors and helpers for interacting with year selector, stats, and charts
export class WrappedPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // year selector dropdown
  readonly yearSelector: Locator
  readonly yearSelectorTrigger: Locator
  readonly yearSelectorContent: Locator
  readonly allTimeOption: Locator

  // comparison year selector
  readonly compareSelector: Locator
  readonly compareSelectorTrigger: Locator
  readonly compareSelectorContent: Locator
  readonly noComparisonOption: Locator

  // metric cards
  readonly metricCards: Locator
  readonly activitiesCard: Locator
  readonly distanceCard: Locator
  readonly elevationCard: Locator
  readonly movingTimeCard: Locator
  readonly kudosCard: Locator
  readonly carbonCard: Locator

  // year activity heatmap
  readonly yearHeatmapCard: Locator
  readonly yearHeatmapChart: Locator

  // monthly bar charts
  readonly activitiesByMonthCard: Locator
  readonly distanceByMonthCard: Locator
  readonly elevationByMonthCard: Locator
  readonly prsByMonthCard: Locator

  // summary section with donut charts
  readonly summaryCard: Locator
  readonly activeVsRestDonut: Locator
  readonly movingTimeBySportDonut: Locator

  // start times chart
  readonly startTimesCard: Locator
  readonly startTimesChart: Locator

  // locations chart
  readonly locationsCard: Locator
  readonly locationsChart: Locator

  // biggest cards
  readonly longestDistanceCard: Locator
  readonly mostElevationCard: Locator
  readonly longestDurationCard: Locator

  // random photo card
  readonly randomPhotoCard: Locator

  // comparison table
  readonly comparisonCard: Locator
  readonly comparisonTable: Locator

  // loading and empty states
  readonly loadingSkeletons: Locator
  readonly emptyState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Wrapped' })
    this.pageSubtitle = page.locator('text=Year in review')

    // year selector
    this.yearSelector = page.locator('button[role="combobox"]').first()
    this.yearSelectorTrigger = this.yearSelector
    this.yearSelectorContent = page.locator('[role="listbox"]').first()
    this.allTimeOption = page.locator('[role="option"]:has-text("All time")')

    // comparison year selector (second combobox)
    this.compareSelectorTrigger = page.locator('button[role="combobox"]').nth(1)
    this.compareSelector = this.compareSelectorTrigger
    this.compareSelectorContent = page.locator('[role="listbox"]')
    this.noComparisonOption = page.locator('[role="option"]:has-text("No comparison")')

    // metric cards (in grid)
    this.metricCards = page
      .locator('.grid.gap-4.md\\:grid-cols-3')
      .first()
      .locator('[data-slot="card"]')
    this.activitiesCard = this.metricCards.filter({
      has: page.locator('text=Activities'),
    })
    this.distanceCard = this.metricCards.filter({
      has: page.locator('text=Distance'),
    })
    this.elevationCard = this.metricCards.filter({
      has: page.locator('text=Elevation'),
    })
    this.movingTimeCard = this.metricCards.filter({
      has: page.locator('text=Moving Time'),
    })
    this.kudosCard = this.metricCards.filter({
      has: page.locator('text=Kudos Received'),
    })
    this.carbonCard = this.metricCards.filter({
      has: page.locator('text=CO₂ Saved'),
    })

    // year activity heatmap
    this.yearHeatmapCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Year Activity Heatmap'),
    })
    this.yearHeatmapChart = this.yearHeatmapCard.locator('canvas, svg').first()

    // monthly bar charts
    this.activitiesByMonthCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Activities by Month'),
    })
    this.distanceByMonthCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Distance by Month'),
    })
    this.elevationByMonthCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Elevation by Month'),
    })
    this.prsByMonthCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Personal Records by Month'),
    })

    // summary section
    this.summaryCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Summary'),
    })
    this.activeVsRestDonut = page.locator('.rounded-md.border').filter({
      has: page.locator('text=Active vs rest days'),
    })
    this.movingTimeBySportDonut = page.locator('.rounded-md.border').filter({
      has: page.locator('text=Moving time by sport'),
    })

    // start times chart
    this.startTimesCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Start Times'),
    })
    this.startTimesChart = this.startTimesCard.locator('canvas, svg').first()

    // locations chart
    this.locationsCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Locations'),
    })
    this.locationsChart = this.locationsCard.locator('canvas, svg').first()

    // biggest cards
    this.longestDistanceCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Longest Distance'),
    })
    this.mostElevationCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Most Elevation'),
    })
    this.longestDurationCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Longest Duration'),
    })

    // random photo card
    this.randomPhotoCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Random Photo'),
    })

    // comparison card
    this.comparisonCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Comparison'),
    })
    this.comparisonTable = this.comparisonCard.locator('.rounded-md.border').filter({
      has: page.locator('.grid.grid-cols-12'),
    })

    // loading and empty states
    this.loadingSkeletons = page.locator('[data-slot="skeleton"]')
    this.emptyState = page.locator('text=No activity data available')
    this.errorMessage = page.locator('text=Failed to load')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/wrapped')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either metric cards, empty state, or error
    await Promise.race([
      this.metricCards.first().waitFor({ state: 'visible', timeout: 15000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 15000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 15000 }),
    ])
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if page has data
  async hasData(): Promise<boolean> {
    try {
      const count = await this.metricCards.count()
      return count > 0
    } catch {
      return false
    }
  }

  // open year selector
  async openYearSelector(): Promise<void> {
    await this.yearSelectorTrigger.click()
    await this.yearSelectorContent.waitFor({ state: 'visible', timeout: 5000 })
  }

  // get current year selection
  async getCurrentYear(): Promise<string | null> {
    return await this.yearSelectorTrigger.textContent()
  }

  // select year
  async selectYear(year: string): Promise<void> {
    await this.openYearSelector()
    await this.page.locator(`[role="option"]:has-text("${year}")`).click()
    await this.page.waitForTimeout(500)
    await this.waitForLoadingComplete()
  }

  // select all time
  async selectAllTime(): Promise<void> {
    await this.openYearSelector()
    await this.allTimeOption.click()
    await this.page.waitForTimeout(500)
    await this.waitForLoadingComplete()
  }

  // get available years
  async getAvailableYears(): Promise<string[]> {
    await this.openYearSelector()
    const options = this.page.locator('[role="option"]')
    const count = await options.count()
    const years: string[] = []
    for (let i = 0; i < count; i++) {
      const text = await options.nth(i).textContent()
      if (text) years.push(text.trim())
    }
    // close dropdown
    await this.page.keyboard.press('Escape')
    return years
  }

  // open comparison selector
  async openCompareSelector(): Promise<void> {
    await this.compareSelectorTrigger.click()
    await this.compareSelectorContent.waitFor({ state: 'visible', timeout: 5000 })
  }

  // select comparison year
  async selectCompareYear(year: string): Promise<void> {
    await this.openCompareSelector()
    await this.page.locator(`[role="option"]:has-text("${year}")`).click()
    await this.page.waitForTimeout(500)
    await this.waitForLoadingComplete()
  }

  // select no comparison
  async selectNoComparison(): Promise<void> {
    await this.openCompareSelector()
    await this.noComparisonOption.click()
    await this.page.waitForTimeout(300)
  }

  // check if comparison table is visible
  async isComparisonTableVisible(): Promise<boolean> {
    return await this.comparisonCard.isVisible()
  }

  // get metric card value
  async getMetricValue(
    metric: 'activities' | 'distance' | 'elevation' | 'movingTime' | 'kudos' | 'carbon'
  ): Promise<string | null> {
    const cards: Record<string, Locator> = {
      activities: this.activitiesCard,
      distance: this.distanceCard,
      elevation: this.elevationCard,
      movingTime: this.movingTimeCard,
      kudos: this.kudosCard,
      carbon: this.carbonCard,
    }
    const card = cards[metric]
    const valueElement = card.locator('.text-2xl.font-bold')
    return await valueElement.textContent()
  }

  // check if heatmap calendar is visible
  async hasHeatmapCalendar(): Promise<boolean> {
    try {
      await this.yearHeatmapCard.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if monthly bar charts are visible
  async hasMonthlyBarCharts(): Promise<boolean> {
    try {
      await this.activitiesByMonthCard.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if donut charts are visible
  async hasDonutCharts(): Promise<boolean> {
    try {
      await this.summaryCard.waitFor({ state: 'visible', timeout: 5000 })
      const activeRest = await this.activeVsRestDonut.isVisible()
      const sport = await this.movingTimeBySportDonut.isVisible()
      return activeRest && sport
    } catch {
      return false
    }
  }

  // get comparison table rows
  async getComparisonTableRows(): Promise<number> {
    if (!(await this.isComparisonTableVisible())) return 0
    const rows = this.comparisonTable.locator('.grid.grid-cols-12.gap-2.px-3.py-2.text-sm')
    return await rows.count()
  }
}
