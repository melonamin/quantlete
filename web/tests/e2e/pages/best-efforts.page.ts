import { expect, type Locator, type Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the best efforts (personal records) page
// provides selectors and helpers for interacting with distance PRs and modal
export class BestEffortsPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // sport filter buttons
  readonly allSportButton: Locator
  readonly runsSportButton: Locator
  readonly ridesSportButton: Locator

  // personal records card
  readonly prsCard: Locator
  readonly distanceRows: Locator

  // PR progression modal
  readonly prModal: Locator
  readonly prModalTitle: Locator
  readonly prModalSubtitle: Locator
  readonly prModalCloseButton: Locator
  readonly prModalChart: Locator
  readonly prModalEffortsTable: Locator

  // loading and error states
  readonly loadingSkeletons: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Best Efforts' })
    this.pageSubtitle = page.locator('text=Your personal records by distance')

    // sport filter buttons (in header area)
    this.allSportButton = page.getByRole('button', { name: 'All', exact: true })
    this.runsSportButton = page.getByRole('button', { name: 'Runs', exact: true })
    this.ridesSportButton = page.getByRole('button', { name: 'Rides', exact: true })

    // personal records card
    this.prsCard = page.locator('[data-slot="card"]').filter({
      has: page.locator('text=Personal Records'),
    })
    this.distanceRows = page.locator('button.flex.w-full.items-center.justify-between')

    // PR progression modal (fixed overlay)
    this.prModal = page.locator('.fixed.inset-0.z-50')
    this.prModalTitle = this.prModal.locator('.text-base.font-semibold')
    this.prModalSubtitle = this.prModal.locator('.text-xs.text-muted-foreground')
    this.prModalCloseButton = this.prModal.locator('button:has-text("Close")')
    this.prModalChart = this.prModal.locator('.rounded-md.border').filter({
      has: page.locator('canvas, svg'),
    })
    this.prModalEffortsTable = this.prModal.locator('.rounded-md.border').filter({
      has: page.locator('.grid-cols-12'),
    })

    // loading and error states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.errorMessage = page.locator('text=Failed to load best efforts')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/best-efforts')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either data or error
    await Promise.race([
      this.prsCard.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if PRs card is visible
  async hasPrsCard(): Promise<boolean> {
    return await this.prsCard.isVisible()
  }

  // get currently selected sport filter
  async getSelectedSportFilter(): Promise<'all' | 'runs' | 'rides' | null> {
    // The Button component exposes its semantic variant directly. Looking for the
    // word "outline" in class names also matches the shared "outline-none" class.
    if ((await this.allSportButton.getAttribute('data-variant')) === 'default') return 'all'
    if ((await this.runsSportButton.getAttribute('data-variant')) === 'default') return 'runs'
    if ((await this.ridesSportButton.getAttribute('data-variant')) === 'default') return 'rides'
    return null
  }

  // select sport filter
  async selectSportFilter(filter: 'all' | 'runs' | 'rides'): Promise<void> {
    const button =
      filter === 'all'
        ? this.allSportButton
        : filter === 'runs'
          ? this.runsSportButton
          : this.ridesSportButton
    await button.click()
    await expect(button).toHaveAttribute('data-variant', 'default')
  }

  // get list of distances with their PR times
  async getDistanceList(): Promise<{ label: string; time: string; hasData: boolean }[]> {
    const rows = this.distanceRows
    const count = await rows.count()
    const distances: { label: string; time: string; hasData: boolean }[] = []

    for (let i = 0; i < count; i++) {
      const row = rows.nth(i)
      const label = await row.locator('.text-sm.font-medium').first().textContent()
      const timeEl = row.locator('.tabular-nums')
      const time = await timeEl.textContent()
      const hasData = time !== '—'

      if (label) {
        distances.push({
          label: label.trim(),
          time: time?.trim() ?? '—',
          hasData,
        })
      }
    }

    return distances
  }

  // get count of distances with data
  async getDistancesWithDataCount(): Promise<number> {
    const distances = await this.getDistanceList()
    return distances.filter((d) => d.hasData).length
  }

  // click a distance row by label
  async clickDistanceRow(label: string): Promise<void> {
    const row = this.distanceRows.filter({
      has: this.page.locator(`.text-sm.font-medium:has-text("${label}")`),
    })
    await row.click()
  }

  // click the first distance row with data
  async clickFirstDistanceWithData(): Promise<boolean> {
    const distances = await this.getDistanceList()
    const firstWithData = distances.find((d) => d.hasData)
    if (firstWithData) {
      await this.clickDistanceRow(firstWithData.label)
      return true
    }
    return false
  }

  // check if modal is open
  async isModalOpen(): Promise<boolean> {
    return await this.prModal.isVisible()
  }

  // get modal title (distance label)
  async getModalTitle(): Promise<string | null> {
    return await this.prModalTitle.textContent()
  }

  // close modal
  async closeModal(): Promise<void> {
    await this.prModalCloseButton.click()
    await this.prModal.waitFor({ state: 'hidden', timeout: 5000 })
  }

  // check if modal has chart
  async modalHasChart(): Promise<boolean> {
    try {
      await this.prModalChart.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if modal has efforts table
  async modalHasEffortsTable(): Promise<boolean> {
    return await this.prModalEffortsTable.isVisible()
  }

  // get efforts count from modal table
  async getModalEffortsCount(): Promise<number> {
    const tableRows = this.prModalEffortsTable.locator('.grid.grid-cols-12.gap-2.px-3.py-2.text-sm')
    return await tableRows.count()
  }

  // check if modal has any efforts listed
  async modalHasEfforts(): Promise<boolean> {
    const count = await this.getModalEffortsCount()
    return count > 0
  }
}
