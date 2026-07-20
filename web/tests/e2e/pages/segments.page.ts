import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the segments page
// provides selectors and helpers for interacting with segment table, filters, and detail modal
export class SegmentsPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // search input
  readonly searchInput: Locator
  readonly searchClearButton: Locator

  // sport type quick filters (toggle group)
  readonly allSportToggle: Locator
  readonly rideSportToggle: Locator
  readonly runSportToggle: Locator

  // starred and KOM toggle buttons
  readonly starredButton: Locator
  readonly komButton: Locator

  // more filters button
  readonly moreFiltersButton: Locator

  // clear filters button
  readonly clearFiltersButton: Locator

  // table
  readonly segmentTable: Locator
  readonly tableRows: Locator
  readonly tableHeaders: Locator

  // sortable column headers
  readonly segmentHeader: Locator
  readonly distanceHeader: Locator
  readonly maxGradeHeader: Locator
  readonly timesHeader: Locator
  readonly lastEffortHeader: Locator
  readonly bestHeader: Locator

  // pagination
  readonly pagination: Locator

  // segment detail modal
  readonly modal: Locator
  readonly modalTitle: Locator
  readonly modalSubtitle: Locator
  readonly modalCloseButton: Locator
  readonly modalStravaLink: Locator
  readonly modalRouteCard: Locator
  readonly modalRouteMap: Locator
  readonly modalPrCard: Locator
  readonly modalPrChart: Locator
  readonly modalEffortsCard: Locator
  readonly modalEffortsTable: Locator

  // loading and empty states
  readonly loadingSkeletons: Locator
  readonly emptyState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Segments' })
    this.pageSubtitle = page.locator('text=Your segment efforts and PRs')

    // search input
    this.searchInput = page.locator('input[placeholder*="Search segments"]')
    this.searchClearButton = this.searchInput
      .locator('..')
      .locator('button')
      .filter({
        has: page.locator('svg'),
      })

    // sport type quick filters (in toggle group)
    this.allSportToggle = page.locator('[role="group"] button').filter({ hasText: /^All$/ })
    this.rideSportToggle = page.locator('[role="group"] button').filter({ hasText: 'Ride' })
    this.runSportToggle = page.locator('[role="group"] button').filter({ hasText: 'Run' })

    // starred and KOM buttons
    this.starredButton = page.locator('button').filter({ hasText: 'Starred' })
    this.komButton = page.locator('button').filter({ hasText: 'KOM' })

    // more filters button
    this.moreFiltersButton = page.locator('button').filter({ hasText: 'More' })

    // clear filters button
    this.clearFiltersButton = page.locator('button').filter({ hasText: 'Clear filters' })

    // table
    this.segmentTable = page.locator('table')
    this.tableRows = this.segmentTable.locator('tbody tr')
    this.tableHeaders = this.segmentTable.locator('thead th')

    // sortable column headers (buttons within th)
    this.segmentHeader = page.locator('th button:has-text("Segment")')
    this.distanceHeader = page.locator('th button:has-text("Distance")')
    this.maxGradeHeader = page.locator('th button:has-text("Max grade")')
    this.timesHeader = page.locator('th button:has-text("Times")')
    this.lastEffortHeader = page.locator('th button:has-text("Last effort")')
    this.bestHeader = page.locator('th button:has-text("Best")')

    // pagination
    this.pagination = page.locator('[class*="pagination"], .flex.items-center.justify-between')

    // segment detail modal (fixed overlay)
    this.modal = page.locator('.fixed.inset-0.z-50')
    this.modalTitle = this.modal.locator('.text-lg.font-semibold')
    this.modalSubtitle = this.modal.locator('.text-sm.text-muted-foreground').first()
    this.modalCloseButton = this.modal.locator('button:has-text("Close")')
    this.modalStravaLink = this.modal.locator('a:has-text("View on Strava")')
    this.modalRouteCard = this.modal.locator('[data-slot="card"]').filter({
      has: page.locator('text=Route'),
    })
    this.modalRouteMap = this.modalRouteCard.locator('.leaflet-container, [class*="map"]')
    this.modalPrCard = this.modal.locator('[data-slot="card"]').filter({
      has: page.locator('text=Best time progression'),
    })
    this.modalPrChart = this.modalPrCard.locator('canvas, svg').first()
    this.modalEffortsCard = this.modal.locator('[data-slot="card"]').filter({
      has: page.locator('text=Efforts'),
    })
    this.modalEffortsTable = this.modalEffortsCard.locator('table')

    // loading and empty states
    this.loadingSkeletons = page.locator('[data-slot="skeleton"]')
    this.emptyState = page.locator('text=No segments found')
    this.errorMessage = page.locator('text=Failed to load segments')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/segments')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either table, empty state, or error
    await Promise.race([
      this.segmentTable.waitFor({ state: 'visible', timeout: 10000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if table has data
  async hasData(): Promise<boolean> {
    try {
      await this.segmentTable.waitFor({ state: 'visible', timeout: 5000 })
      const rowCount = await this.tableRows.count()
      return rowCount > 0
    } catch {
      return false
    }
  }

  // get row count
  async getRowCount(): Promise<number> {
    return await this.tableRows.count()
  }

  // type in search input (with debounce simulation)
  async search(text: string): Promise<void> {
    await this.searchInput.fill(text)
    // wait for debounce (300ms in component)
    await this.page.waitForTimeout(400)
    await this.waitForLoadingComplete()
  }

  // clear search
  async clearSearch(): Promise<void> {
    await this.searchInput.clear()
    await this.page.waitForTimeout(400)
    await this.waitForLoadingComplete()
  }

  // select sport filter
  async selectSportFilter(filter: 'all' | 'ride' | 'run'): Promise<void> {
    const button =
      filter === 'all'
        ? this.allSportToggle
        : filter === 'ride'
          ? this.rideSportToggle
          : this.runSportToggle
    await button.click()
    await this.page.waitForTimeout(300)
    await this.waitForLoadingComplete()
  }

  // check if sport filter is selected
  async isSportFilterSelected(filter: 'all' | 'ride' | 'run'): Promise<boolean> {
    if (filter === 'all') {
      const rideState = await this.rideSportToggle.getAttribute('data-state')
      const runState = await this.runSportToggle.getAttribute('data-state')
      return rideState !== 'on' && runState !== 'on'
    }

    const button = filter === 'ride' ? this.rideSportToggle : this.runSportToggle
    const dataState = await button.getAttribute('data-state')
    return dataState === 'on'
  }

  // toggle starred filter
  async toggleStarred(): Promise<void> {
    await this.starredButton.click()
    await this.page.waitForTimeout(300)
    await this.waitForLoadingComplete()
  }

  // check if starred is active
  async isStarredActive(): Promise<boolean> {
    const classes = await this.starredButton.getAttribute('class')
    return classes?.includes('bg-primary') ?? false
  }

  // toggle KOM filter
  async toggleKom(): Promise<void> {
    await this.komButton.click()
    await this.page.waitForTimeout(300)
    await this.waitForLoadingComplete()
  }

  // check if KOM is active
  async isKomActive(): Promise<boolean> {
    const classes = await this.komButton.getAttribute('class')
    return classes?.includes('bg-primary') ?? false
  }

  // click column header to sort
  async clickColumnHeader(
    column: 'segment' | 'distance' | 'maxGrade' | 'times' | 'lastEffort' | 'best'
  ): Promise<void> {
    const headers: Record<string, Locator> = {
      segment: this.segmentHeader,
      distance: this.distanceHeader,
      maxGrade: this.maxGradeHeader,
      times: this.timesHeader,
      lastEffort: this.lastEffortHeader,
      best: this.bestHeader,
    }
    await headers[column].click()
    await this.page.waitForTimeout(300)
    await this.waitForLoadingComplete()
  }

  // check if column is sorted
  async isColumnSorted(
    column: 'segment' | 'distance' | 'maxGrade' | 'times' | 'lastEffort' | 'best'
  ): Promise<{ sorted: boolean; direction: 'asc' | 'desc' | null }> {
    const headers: Record<string, Locator> = {
      segment: this.segmentHeader,
      distance: this.distanceHeader,
      maxGrade: this.maxGradeHeader,
      times: this.timesHeader,
      lastEffort: this.lastEffortHeader,
      best: this.bestHeader,
    }
    const text = await headers[column].textContent()
    if (text?.includes('↑')) return { sorted: true, direction: 'asc' }
    if (text?.includes('↓')) return { sorted: true, direction: 'desc' }
    return { sorted: false, direction: null }
  }

  // click a table row by index
  async clickRow(index: number): Promise<void> {
    await this.tableRows.nth(index).click()
  }

  // get segment name from row
  async getSegmentName(index: number): Promise<string | null> {
    const row = this.tableRows.nth(index)
    const nameCell = row.locator('td').first()
    const name = nameCell.locator('.font-medium')
    return await name.textContent()
  }

  // check if modal is open
  async isModalOpen(): Promise<boolean> {
    return await this.modal.isVisible()
  }

  // get modal segment name
  async getModalSegmentName(): Promise<string | null> {
    return await this.modalTitle.textContent()
  }

  // close modal
  async closeModal(): Promise<void> {
    await this.modalCloseButton.click()
    await this.modal.waitFor({ state: 'hidden', timeout: 5000 })
  }

  // check if modal has map
  async modalHasMap(): Promise<boolean> {
    try {
      // wait for either a leaflet container or any map-like element
      await this.modalRouteCard.waitFor({ state: 'visible', timeout: 5000 })
      // check for any content in the route card
      const hasContent = await this.modalRouteCard
        .locator('.leaflet-container, svg, canvas')
        .count()
      return hasContent > 0
    } catch {
      return false
    }
  }

  // check if modal has PR chart
  async modalHasPrChart(): Promise<boolean> {
    try {
      await this.modalPrCard.waitFor({ state: 'visible', timeout: 5000 })
      const chart = this.modalPrCard.locator('canvas, svg').first()
      return await chart.isVisible()
    } catch {
      return false
    }
  }

  // check if modal has efforts table
  async modalHasEffortsTable(): Promise<boolean> {
    try {
      await this.modalEffortsCard.waitFor({ state: 'visible', timeout: 5000 })
      return await this.modalEffortsTable.isVisible()
    } catch {
      return false
    }
  }

  // get efforts count from modal
  async getModalEffortsCount(): Promise<number> {
    if (!(await this.modalHasEffortsTable())) return 0
    return await this.modalEffortsTable.locator('tbody tr').count()
  }

  // clear all filters
  async clearFilters(): Promise<void> {
    if (await this.clearFiltersButton.isVisible()) {
      await this.clearFiltersButton.click()
      await this.page.waitForTimeout(300)
      await this.waitForLoadingComplete()
    }
  }

  // check if clear filters button is visible
  async hasActiveFilters(): Promise<boolean> {
    return await this.clearFiltersButton.isVisible()
  }
}
