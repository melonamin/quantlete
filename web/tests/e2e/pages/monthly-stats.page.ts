import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the monthly stats page
// provides selectors and helpers for interacting with year navigation, summary cards, and accordion table
export class MonthlyStatsPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // year navigation
  readonly prevYearButton: Locator
  readonly nextYearButton: Locator
  readonly yearDisplay: Locator
  readonly exportButton: Locator

  // yearly summary cards
  readonly summaryCards: Locator
  readonly totalActivitiesCard: Locator
  readonly totalDistanceCard: Locator
  readonly totalTimeCard: Locator
  readonly totalElevationCard: Locator

  // monthly breakdown table
  readonly monthlyTable: Locator
  readonly tableContainer: Locator
  readonly tableHeaders: Locator
  readonly tableRows: Locator

  // accordion table elements
  readonly accordionRows: Locator
  readonly expandButtons: Locator
  readonly expandedContent: Locator

  // loading and empty states
  readonly loadingSkeletons: Locator
  readonly emptyState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Monthly Statistics' })
    this.pageSubtitle = page.locator('text=Detailed monthly breakdown of your activities')

    // year navigation
    this.prevYearButton = page.locator('button').filter({ has: page.locator('svg') }).first()
    this.nextYearButton = page.locator('button').filter({ has: page.locator('svg') }).nth(1)
    this.yearDisplay = page.locator('.w-16.text-center.font-medium')
    this.exportButton = page.locator('button:has-text("Export")')

    // yearly summary cards
    this.summaryCards = page.locator('.mb-6.grid')
    this.totalActivitiesCard = page.locator('.rounded-lg.border').filter({
      has: page.locator('text=Total Activities'),
    })
    this.totalDistanceCard = page.locator('.rounded-lg.border').filter({
      has: page.locator('text=Total Distance'),
    })
    this.totalTimeCard = page.locator('.rounded-lg.border').filter({
      has: page.locator('text=Total Time'),
    })
    this.totalElevationCard = page.locator('.rounded-lg.border').filter({
      has: page.locator('text=Total Elevation'),
    })

    // monthly breakdown table
    this.monthlyTable = page.locator('.rounded-lg.border.border-border.bg-card')
    this.tableContainer = this.monthlyTable
    this.tableHeaders = page.locator('th, [role="columnheader"]')
    this.tableRows = page.locator('tbody tr, [role="row"]')

    // accordion table elements
    this.accordionRows = page.locator('[data-state], .accordion-item, tbody > tr')
    this.expandButtons = page.locator('button[aria-expanded], [role="button"]').filter({
      has: page.locator('svg'),
    })
    this.expandedContent = page.locator('[data-state="open"], .accordion-content')

    // loading and empty states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.emptyState = page.locator('text=No activities recorded')
    this.errorMessage = page.locator('text=Failed to load')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/monthly-stats')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either summary cards, empty state, or error
    await Promise.race([
      this.summaryCards.waitFor({ state: 'visible', timeout: 10000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ]).catch(() => {
      // page might load with different state
    })
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if page has data
  async hasData(): Promise<boolean> {
    try {
      await this.summaryCards.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // get current year
  async getCurrentYear(): Promise<number> {
    const yearText = await this.yearDisplay.textContent()
    return parseInt(yearText || '0', 10)
  }

  // navigate to previous year
  async goToPreviousYear(): Promise<void> {
    await this.prevYearButton.click()
    await this.page.waitForTimeout(500)
    await this.waitForLoadingComplete()
  }

  // navigate to next year
  async goToNextYear(): Promise<void> {
    await this.nextYearButton.click()
    await this.page.waitForTimeout(500)
    await this.waitForLoadingComplete()
  }

  // check if next year button is disabled
  async isNextYearDisabled(): Promise<boolean> {
    const disabled = await this.nextYearButton.isDisabled()
    return disabled
  }

  // get total activities value
  async getTotalActivitiesValue(): Promise<string | null> {
    try {
      const value = this.totalActivitiesCard.locator('.text-2xl.font-bold')
      return await value.textContent()
    } catch {
      return null
    }
  }

  // get total distance value
  async getTotalDistanceValue(): Promise<string | null> {
    try {
      const value = this.totalDistanceCard.locator('.text-2xl.font-bold')
      return await value.textContent()
    } catch {
      return null
    }
  }

  // get total time value
  async getTotalTimeValue(): Promise<string | null> {
    try {
      const value = this.totalTimeCard.locator('.text-2xl.font-bold')
      return await value.textContent()
    } catch {
      return null
    }
  }

  // get total elevation value
  async getTotalElevationValue(): Promise<string | null> {
    try {
      const value = this.totalElevationCard.locator('.text-2xl.font-bold')
      return await value.textContent()
    } catch {
      return null
    }
  }

  // check if monthly table is visible
  async isMonthlyTableVisible(): Promise<boolean> {
    return await this.tableContainer.isVisible()
  }

  // get row count
  async getRowCount(): Promise<number> {
    // find rows in the accordion table
    const rows = this.page.locator('table tbody tr, [role="row"]')
    return await rows.count()
  }

  // click a row to expand/collapse (if accordion style)
  async clickRow(index: number): Promise<void> {
    const rows = this.page.locator('table tbody tr')
    const row = rows.nth(index)
    // look for a clickable element in the row
    const clickable = row.locator('button, [role="button"]').first()
    if (await clickable.isVisible()) {
      await clickable.click()
    } else {
      // try clicking the row itself
      await row.click()
    }
    await this.page.waitForTimeout(300)
  }

  // check if a row is expanded
  async isRowExpanded(index: number): Promise<boolean> {
    const rows = this.page.locator('table tbody tr')
    const row = rows.nth(index)
    const ariaExpanded = await row.getAttribute('aria-expanded')
    const dataState = await row.getAttribute('data-state')
    return ariaExpanded === 'true' || dataState === 'open'
  }

  // click export button
  async clickExport(): Promise<void> {
    await this.exportButton.click()
  }

  // check if export button is visible
  async isExportButtonVisible(): Promise<boolean> {
    return await this.exportButton.isVisible()
  }

  // get month name from first row
  async getFirstMonthName(): Promise<string | null> {
    try {
      const rows = this.page.locator('table tbody tr')
      const firstRow = rows.first()
      const monthCell = firstRow.locator('td').first()
      return await monthCell.textContent()
    } catch {
      return null
    }
  }
}
