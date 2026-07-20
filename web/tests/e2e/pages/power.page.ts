import { expect, type Locator, type Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the power analytics page
// provides selectors and helpers for interacting with power charts and controls
export class PowerPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator
  readonly backButton: Locator

  // all-time best card
  readonly allTimeBestCard: Locator
  readonly allTimeBestChart: Locator

  // progression card
  readonly progressionCard: Locator
  readonly progressionChart: Locator
  readonly durationSelector: Locator
  readonly durationSelectorTrigger: Locator

  // power curve comparison card
  readonly powerCurveCard: Locator
  readonly powerCurveChart: Locator

  // power zones card
  readonly powerZonesCard: Locator
  readonly powerZonesChart: Locator
  readonly ftpDisplay: Locator

  // loading and error states
  readonly loadingState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Power' })
    this.pageSubtitle = page.locator('text=Peak power outputs and progression')
    this.backButton = page.locator('a:has-text("Back")')

    // all-time best card (first card with bar chart)
    this.allTimeBestCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=All-time Best'),
    })
    this.allTimeBestChart = this.allTimeBestCard.locator('canvas, svg').first()

    // progression card (has duration selector)
    this.progressionCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Progression'),
    })
    this.progressionChart = this.progressionCard.locator('canvas, svg').first()
    this.durationSelectorTrigger = this.progressionCard.locator('button[role="combobox"]')
    this.durationSelector = page.locator('[role="listbox"]')

    // power curve comparison card
    this.powerCurveCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Power Curve Comparison'),
    })
    this.powerCurveChart = this.powerCurveCard.locator('canvas, svg').first()

    // power zones card
    this.powerZonesCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Power Zones'),
    })
    this.powerZonesChart = this.powerZonesCard.locator('canvas, svg, [class*="zone"]').first()
    this.ftpDisplay = page.locator('text=/Based on FTP/')

    // loading and error states
    this.loadingState = page.locator('[data-slot="skeleton"], [class*="animate-pulse"]').first()
    this.errorMessage = page.locator('text=Failed to load power stats')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/power')
  }

  // wait for charts to render
  async waitForChartsLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for at least one chart to be visible
    await Promise.race([
      this.allTimeBestChart.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // check if all-time best chart has rendered
  async hasAllTimeBestChart(): Promise<boolean> {
    try {
      await this.allTimeBestChart.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if progression chart has rendered
  async hasProgressionChart(): Promise<boolean> {
    try {
      await this.progressionChart.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if power curve comparison chart has rendered
  async hasPowerCurveChart(): Promise<boolean> {
    try {
      await this.powerCurveChart.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if power zones chart has rendered
  async hasPowerZonesChart(): Promise<boolean> {
    try {
      await this.powerZonesChart.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // open duration selector dropdown
  async openDurationSelector(): Promise<void> {
    await this.durationSelectorTrigger.click()
    await this.durationSelector.waitFor({ state: 'visible', timeout: 5000 })
  }

  // get available duration options
  async getDurationOptions(): Promise<string[]> {
    await this.openDurationSelector()
    const options = this.page.locator('[role="option"]')
    const count = await options.count()
    const values: string[] = []
    for (let i = 0; i < count; i++) {
      const text = await options.nth(i).textContent()
      if (text) values.push(text.trim())
    }

    await this.page.keyboard.press('Escape')
    await this.durationSelector.waitFor({ state: 'hidden', timeout: 5000 })
    return values
  }

  // select a duration from the dropdown
  async selectDuration(duration: string): Promise<void> {
    await this.openDurationSelector()
    await this.page.getByRole('option', { name: duration, exact: true }).click()
    await this.durationSelector.waitFor({ state: 'hidden', timeout: 5000 })
    await expect(this.durationSelectorTrigger).toContainText(duration)
  }

  // get currently selected duration
  async getSelectedDuration(): Promise<string | null> {
    return await this.durationSelectorTrigger.textContent()
  }

  // check if FTP is displayed
  async hasFtpDisplay(): Promise<boolean> {
    try {
      await this.ftpDisplay.waitFor({ state: 'visible', timeout: 3000 })
      return true
    } catch {
      return false
    }
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }
}
