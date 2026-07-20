import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the training load analytics page
// provides selectors and helpers for interacting with training load charts and summary cards
export class TrainingLoadPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator
  readonly backButton: Locator

  // date range filters
  readonly rangeCard: Locator
  readonly afterDateInput: Locator
  readonly beforeDateInput: Locator

  // training load chart
  readonly trainingLoadChart: Locator

  // summary cards
  readonly ctlCard: Locator
  readonly atlCard: Locator
  readonly tsbCard: Locator

  // configuration warning alert
  readonly configWarningAlert: Locator
  readonly settingsLink: Locator

  // activity breakdown alert
  readonly activityBreakdownAlert: Locator
  readonly totalActivitiesCount: Locator
  readonly activitiesWithPower: Locator
  readonly activitiesWithTss: Locator

  // loading and error states
  readonly loadingState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Training Load' })
    this.pageSubtitle = page.locator('text=Fitness, fatigue, and form over time')
    this.backButton = page.locator('a:has-text("Back")')

    // date range card and inputs
    this.rangeCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Range'),
    })
    this.afterDateInput = this.rangeCard.locator('input[type="date"]').first()
    this.beforeDateInput = this.rangeCard.locator('input[type="date"]').last()

    // training load chart (inside the range card)
    this.trainingLoadChart = this.rangeCard.locator('canvas, svg').first()

    // summary cards (CTL, ATL, TSB)
    this.ctlCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=CTL'),
    })
    this.atlCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=ATL'),
    })
    this.tsbCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=TSB'),
    })

    // configuration warning alert (amber colored)
    this.configWarningAlert = page.locator('.border-amber-500\\/50, [class*="bg-amber"]').first()
    this.settingsLink = this.configWarningAlert.locator('a:has-text("Go to Settings")')

    // activity breakdown alert
    this.activityBreakdownAlert = page.locator('text=Activity Breakdown').locator('..')
    this.totalActivitiesCount = page.locator('text=Total activities:')
    this.activitiesWithPower = page.locator('text=With power data:')
    this.activitiesWithTss = page.locator('text=With computed TSS:')

    // loading and error states
    this.loadingState = page.locator('[data-slot="skeleton"], [class*="animate-pulse"]').first()
    this.errorMessage = page.locator('text=Failed to load training load')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/training-load')
  }

  // wait for chart to render
  async waitForChartLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either chart or error
    await Promise.race([
      this.trainingLoadChart.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // check if training load chart has rendered
  async hasTrainingLoadChart(): Promise<boolean> {
    try {
      await this.trainingLoadChart.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if configuration warning is displayed
  async hasConfigWarning(): Promise<boolean> {
    return await this.configWarningAlert.isVisible()
  }

  // get configuration warning text
  async getConfigWarningText(): Promise<string | null> {
    if (await this.hasConfigWarning()) {
      return await this.configWarningAlert.textContent()
    }
    return null
  }

  // set date range filters
  async setDateRange(afterDate: string, beforeDate: string): Promise<void> {
    await this.afterDateInput.fill(afterDate)
    await this.beforeDateInput.fill(beforeDate)
    // wait for chart to update
    await this.page.waitForTimeout(500)
  }

  // get current date range values
  async getDateRange(): Promise<{ after: string; before: string }> {
    const after = await this.afterDateInput.inputValue()
    const before = await this.beforeDateInput.inputValue()
    return { after, before }
  }

  // clear date range filters
  async clearDateRange(): Promise<void> {
    await this.afterDateInput.clear()
    await this.beforeDateInput.clear()
    await this.page.waitForTimeout(500)
  }

  // check if summary cards are visible
  async hasSummaryCards(): Promise<boolean> {
    const ctlVisible = await this.ctlCard.isVisible()
    const atlVisible = await this.atlCard.isVisible()
    const tsbVisible = await this.tsbCard.isVisible()
    return ctlVisible && atlVisible && tsbVisible
  }

  // get CTL value from card
  async getCtlValue(): Promise<string | null> {
    if (!(await this.ctlCard.isVisible())) return null
    const valueEl = this.ctlCard.locator('.text-2xl.font-bold')
    return await valueEl.textContent()
  }

  // get ATL value from card
  async getAtlValue(): Promise<string | null> {
    if (!(await this.atlCard.isVisible())) return null
    const valueEl = this.atlCard.locator('.text-2xl.font-bold')
    return await valueEl.textContent()
  }

  // get TSB value from card
  async getTsbValue(): Promise<string | null> {
    if (!(await this.tsbCard.isVisible())) return null
    const valueEl = this.tsbCard.locator('.text-2xl.font-bold')
    return await valueEl.textContent()
  }

  // get all summary values
  async getSummaryValues(): Promise<{
    ctl: string | null
    atl: string | null
    tsb: string | null
  }> {
    return {
      ctl: await this.getCtlValue(),
      atl: await this.getAtlValue(),
      tsb: await this.getTsbValue(),
    }
  }

  // check if activity breakdown is visible
  async hasActivityBreakdown(): Promise<boolean> {
    return await this.activityBreakdownAlert.isVisible()
  }
}
