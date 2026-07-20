import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the Eddington number page
// provides selectors and helpers for interacting with eddington charts and view modes
export class EddingtonPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // view mode tabs
  readonly viewModeTabs: Locator
  readonly allTab: Locator
  readonly bySportTab: Locator
  readonly customTab: Locator

  // sport group selector (visible in By Sport mode)
  readonly sportGroupSelector: Locator
  readonly sportGroupTrigger: Locator

  // custom definition selector (visible in Custom mode)
  readonly customDefSelector: Locator
  readonly customDefTrigger: Locator

  // sport comparison card
  readonly sportComparisonCard: Locator
  readonly sportComparisonButtons: Locator
  readonly allComparisonButton: Locator

  // main eddington number display
  readonly eddingtonCard: Locator
  readonly eddingtonNumber: Locator
  readonly eddingtonDescription: Locator

  // history chart card
  readonly historyCard: Locator
  readonly historyChart: Locator

  // next goals table
  readonly nextGoalsCard: Locator
  readonly nextGoalsTable: Locator

  // top days table
  readonly topDaysCard: Locator
  readonly topDaysTable: Locator

  // loading and error states
  readonly loadingSkeletons: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Eddington Number' })
    this.pageSubtitle = page.locator('text=Track your Eddington number progress')

    // view mode tabs
    this.viewModeTabs = page.locator('[role="tablist"]')
    this.allTab = page.locator('[role="tab"]:has-text("All")')
    this.bySportTab = page.locator('[role="tab"]:has-text("By Sport")')
    this.customTab = page.locator('[role="tab"]:has-text("Custom")')

    // sport group selector (shows in By Sport mode)
    this.sportGroupSelector = page.locator('[role="listbox"]')
    this.sportGroupTrigger = page
      .locator('button[role="combobox"]')
      .filter({ has: page.locator('text=Select sport') })
      .or(page.locator('button[role="combobox"]').first())

    // custom definition selector (shows in Custom mode)
    this.customDefSelector = page.locator('[role="listbox"]')
    this.customDefTrigger = page.locator('button[role="combobox"]')

    // sport comparison card (with clickable buttons for each sport group)
    this.sportComparisonCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Sport Comparison'),
    })
    this.sportComparisonButtons = this.sportComparisonCard.locator('button[type="button"]')
    this.allComparisonButton = this.sportComparisonCard.locator('button:has-text("All")')

    // main eddington number card
    this.eddingtonCard = page
      .locator('[data-slot="card"]')
      .filter({
        has: page.locator('.text-7xl.font-bold'),
      })
      .first()
    this.eddingtonNumber = page.locator('.text-7xl.font-bold')
    this.eddingtonDescription = this.eddingtonCard.locator('p.text-muted-foreground')

    // history card
    this.historyCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=History'),
    })
    this.historyChart = this.historyCard.locator('canvas, svg').first()

    // next goals card
    this.nextGoalsCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Next Goals'),
    })
    this.nextGoalsTable = this.nextGoalsCard.locator('table')

    // top distance days card
    this.topDaysCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Top Distance Days'),
    })
    this.topDaysTable = this.topDaysCard.locator('table')

    // loading and error states
    this.loadingSkeletons = page.locator('[data-slot="skeleton"]')
    this.errorMessage = page.locator('text=Failed to load Eddington data')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/eddington')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either data or error
    await Promise.race([
      this.eddingtonNumber.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // get current eddington number
  async getEddingtonNumber(): Promise<number> {
    const text = await this.eddingtonNumber.textContent()
    return parseInt(text?.trim() ?? '0', 10)
  }

  // get description text
  async getEddingtonDescription(): Promise<string | null> {
    return await this.eddingtonDescription.textContent()
  }

  // get currently selected view mode
  async getSelectedViewMode(): Promise<'all' | 'sport-group' | 'custom' | null> {
    const allSelected = await this.allTab.getAttribute('data-state')
    const bySportSelected = await this.bySportTab.getAttribute('data-state')
    const customSelected = await this.customTab.getAttribute('data-state')

    if (allSelected === 'active') return 'all'
    if (bySportSelected === 'active') return 'sport-group'
    if (customSelected === 'active') return 'custom'
    return null
  }

  // select view mode
  async selectViewMode(mode: 'all' | 'sport-group' | 'custom'): Promise<void> {
    const tab =
      mode === 'all' ? this.allTab : mode === 'sport-group' ? this.bySportTab : this.customTab
    await tab.click()
    await this.page.waitForTimeout(500)
  }

  // check if sport comparison card is visible
  async hasSportComparisonCard(): Promise<boolean> {
    return await this.sportComparisonCard.isVisible()
  }

  // get sport groups from comparison card
  async getSportGroups(): Promise<{ name: string; number: number }[]> {
    const buttons = this.sportComparisonButtons
    const count = await buttons.count()
    const groups: { name: string; number: number }[] = []

    for (let i = 0; i < count; i++) {
      const button = buttons.nth(i)
      const numberEl = button.locator('.text-2xl.font-bold')
      const nameEl = button.locator('.text-sm.text-muted-foreground')

      const numberText = await numberEl.textContent()
      const nameText = await nameEl.textContent()

      if (numberText && nameText) {
        groups.push({
          name: nameText.trim(),
          number: parseInt(numberText.trim(), 10),
        })
      }
    }

    return groups
  }

  // click a sport group in comparison card
  async clickSportGroup(name: string): Promise<void> {
    const button = this.sportComparisonCard.locator(`button:has-text("${name}")`)
    await button.click()
    await this.page.waitForTimeout(500)
  }

  // check if history chart is visible
  async hasHistoryChart(): Promise<boolean> {
    try {
      await this.historyChart.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if next goals table is visible
  async hasNextGoalsTable(): Promise<boolean> {
    return await this.nextGoalsCard.isVisible()
  }

  // get next goals
  async getNextGoals(): Promise<{ target: string; ridesNeeded: string }[]> {
    if (!(await this.hasNextGoalsTable())) return []

    const rows = this.nextGoalsTable.locator('tbody tr')
    const count = await rows.count()
    const goals: { target: string; ridesNeeded: string }[] = []

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const cells = row.locator('td')
      const target = await cells.nth(0).textContent()
      const ridesNeeded = await cells.nth(1).textContent()

      if (target && ridesNeeded) {
        goals.push({
          target: target.trim(),
          ridesNeeded: ridesNeeded.trim(),
        })
      }
    }

    return goals
  }

  // check if top days table is visible
  async hasTopDaysTable(): Promise<boolean> {
    return await this.topDaysCard.isVisible()
  }

  // get visible sport group dropdown options
  async getSportGroupOptions(): Promise<string[]> {
    await this.sportGroupTrigger.click()
    await this.sportGroupSelector.waitFor({ state: 'visible', timeout: 5000 })

    const options = this.page.locator('[role="option"]')
    const count = await options.count()
    const values: string[] = []

    for (let i = 0; i < count; i++) {
      const text = await options.nth(i).textContent()
      if (text) values.push(text.trim())
    }

    // close dropdown
    await this.page.keyboard.press('Escape')
    return values
  }

  // select a sport group from dropdown
  async selectSportGroup(name: string): Promise<void> {
    await this.sportGroupTrigger.click()
    await this.sportGroupSelector.waitFor({ state: 'visible', timeout: 5000 })
    await this.page.locator(`[role="option"]:has-text("${name}")`).click()
    await this.page.waitForTimeout(500)
  }
}
