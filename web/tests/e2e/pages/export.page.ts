import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the export page
// provides selectors and helpers for interacting with export options and download
export class ExportPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // stats card (Your Data)
  readonly statsCard: Locator
  readonly statsCardTitle: Locator
  readonly totalActivitiesValue: Locator
  readonly firstActivityDate: Locator
  readonly lastActivityDate: Locator

  // export options card
  readonly exportOptionsCard: Locator
  readonly exportOptionsCardTitle: Locator

  // format selector buttons
  readonly formatLabel: Locator
  readonly csvButton: Locator
  readonly jsonButton: Locator
  readonly formatDescription: Locator

  // date range filters
  readonly dateRangeLabel: Locator
  readonly fromDateInput: Locator
  readonly toDateInput: Locator

  // sport type filter
  readonly sportTypeLabel: Locator
  readonly sportTypeInput: Locator

  // download button
  readonly downloadButton: Locator

  // what's included card
  readonly whatsIncludedCard: Locator
  readonly activityDetailsSection: Locator
  readonly performanceMetricsSection: Locator
  readonly locationGearSection: Locator

  // loading and error states
  readonly loadingSkeletons: Locator
  readonly errorMessage: Locator
  readonly noDataMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Export Data' })
    this.pageSubtitle = page.locator('text=Download your activity data in various formats')

    // stats card
    this.statsCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Your Data'),
    })
    this.statsCardTitle = this.statsCard.locator('[class*="CardTitle"]')
    this.totalActivitiesValue = this.statsCard.locator('.text-xl.font-bold')
    this.firstActivityDate = this.statsCard.locator('text=First Activity').locator('..').locator('.text-sm').last()
    this.lastActivityDate = this.statsCard.locator('text=Last Activity').locator('..').locator('.text-sm').last()

    // export options card
    this.exportOptionsCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Export Options'),
    })
    this.exportOptionsCardTitle = this.exportOptionsCard.locator('[class*="CardTitle"]')

    // format selector
    this.formatLabel = this.exportOptionsCard.locator('label:has-text("Format")')
    this.csvButton = this.exportOptionsCard.locator('button:has-text("CSV")')
    this.jsonButton = this.exportOptionsCard.locator('button:has-text("JSON")')
    this.formatDescription = this.exportOptionsCard.locator('.mt-1.text-xs')

    // date range filters
    this.dateRangeLabel = this.exportOptionsCard.locator('label').filter({
      has: page.locator('text=Date Range'),
    })
    this.fromDateInput = this.exportOptionsCard.locator('label:has-text("From")').locator('..').locator('input[type="date"]')
    this.toDateInput = this.exportOptionsCard.locator('label:has-text("To")').locator('..').locator('input[type="date"]')

    // sport type filter
    this.sportTypeLabel = this.exportOptionsCard.locator('label:has-text("Sport Type")')
    this.sportTypeInput = this.exportOptionsCard.locator('input[placeholder*="Ride"]')

    // download button
    this.downloadButton = this.exportOptionsCard.locator('button:has-text("Download")')

    // what's included card
    this.whatsIncludedCard = page.locator('[class*="card"]').filter({
      has: page.locator('text="What\'s Included"'),
    })
    this.activityDetailsSection = this.whatsIncludedCard.locator('h4:has-text("Activity Details")').locator('..')
    this.performanceMetricsSection = this.whatsIncludedCard.locator('h4:has-text("Performance Metrics")').locator('..')
    this.locationGearSection = this.whatsIncludedCard.locator('h4:has-text("Location")').locator('..')

    // loading and error states
    this.loadingSkeletons = page.locator('[class*="animate-pulse"]')
    this.errorMessage = page.locator('text=Failed to load')
    this.noDataMessage = page.locator('text=No data available')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/export')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for stats card or no data message
    await Promise.race([
      this.totalActivitiesValue.waitFor({ state: 'visible', timeout: 10000 }),
      this.noDataMessage.waitFor({ state: 'visible', timeout: 10000 }),
      this.loadingSkeletons.first().waitFor({ state: 'hidden', timeout: 10000 }),
    ]).catch(() => {
      // page might load with different state
    })
  }

  // check if stats are loaded
  async hasStats(): Promise<boolean> {
    try {
      await this.totalActivitiesValue.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // get total activities count
  async getTotalActivities(): Promise<string | null> {
    try {
      return await this.totalActivitiesValue.textContent()
    } catch {
      return null
    }
  }

  // check if CSV format is selected
  async isCsvSelected(): Promise<boolean> {
    const classes = await this.csvButton.getAttribute('class')
    const variant = await this.csvButton.getAttribute('data-state')
    return (
      !classes?.includes('outline') ||
      classes?.includes('bg-primary') ||
      variant === 'active' ||
      false
    )
  }

  // check if JSON format is selected
  async isJsonSelected(): Promise<boolean> {
    const classes = await this.jsonButton.getAttribute('class')
    const variant = await this.jsonButton.getAttribute('data-state')
    return (
      !classes?.includes('outline') ||
      classes?.includes('bg-primary') ||
      variant === 'active' ||
      false
    )
  }

  // select CSV format
  async selectCsvFormat(): Promise<void> {
    await this.csvButton.click()
    await this.page.waitForTimeout(200)
  }

  // select JSON format
  async selectJsonFormat(): Promise<void> {
    await this.jsonButton.click()
    await this.page.waitForTimeout(200)
  }

  // get selected format from download button text
  async getSelectedFormat(): Promise<'csv' | 'json' | null> {
    const buttonText = await this.downloadButton.textContent()
    if (buttonText?.includes('CSV')) return 'csv'
    if (buttonText?.includes('JSON')) return 'json'
    return null
  }

  // get format description text
  async getFormatDescription(): Promise<string | null> {
    try {
      return await this.formatDescription.textContent()
    } catch {
      return null
    }
  }

  // fill from date
  async fillFromDate(date: string): Promise<void> {
    await this.fromDateInput.fill(date)
  }

  // fill to date
  async fillToDate(date: string): Promise<void> {
    await this.toDateInput.fill(date)
  }

  // fill sport type
  async fillSportType(sportType: string): Promise<void> {
    await this.sportTypeInput.fill(sportType)
  }

  // get from date value
  async getFromDateValue(): Promise<string> {
    return await this.fromDateInput.inputValue()
  }

  // get to date value
  async getToDateValue(): Promise<string> {
    return await this.toDateInput.inputValue()
  }

  // click download button and return download info
  async clickDownload(): Promise<void> {
    await this.downloadButton.click()
  }

  // check if what's included card is visible
  async isWhatsIncludedVisible(): Promise<boolean> {
    return await this.whatsIncludedCard.isVisible()
  }
}
