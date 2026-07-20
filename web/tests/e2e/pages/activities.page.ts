import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the activities page
// provides selectors and helpers for interacting with activity table, filters, and pagination
export class ActivitiesPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // filters panel
  readonly filtersPanel: Locator
  readonly searchInput: Locator
  readonly searchClearButton: Locator

  // quick sport filters (toggle group)
  readonly quickSportFilters: Locator
  readonly allSportButton: Locator

  // "More" button for advanced filters
  readonly moreFiltersButton: Locator

  // advanced filters section
  readonly advancedFiltersSection: Locator
  readonly sportTypeDropdown: Locator
  readonly dateFromInput: Locator
  readonly dateToInput: Locator
  readonly commuteCheckbox: Locator
  readonly trainerCheckbox: Locator
  readonly minDistanceInput: Locator
  readonly maxDistanceInput: Locator
  readonly minDurationInput: Locator
  readonly maxDurationInput: Locator

  // clear filters button
  readonly clearFiltersButton: Locator

  // activities table
  readonly activitiesTable: Locator
  readonly tableHeader: Locator
  readonly tableBody: Locator
  readonly tableRows: Locator
  readonly emptyState: Locator
  readonly loadingSkeleton: Locator

  // table column headers
  readonly activityHeader: Locator
  readonly typeHeader: Locator
  readonly dateHeader: Locator
  readonly distanceHeader: Locator
  readonly timeHeader: Locator
  readonly elevationHeader: Locator
  readonly tagsHeader: Locator

  // pagination
  readonly paginationContainer: Locator
  readonly showingInfo: Locator
  readonly previousButton: Locator
  readonly nextButton: Locator
  readonly pageInfo: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Activities' })
    this.pageSubtitle = page.locator('text=Browse and filter your activities')

    // filters panel
    this.filtersPanel = page.locator('.rounded-lg.border.border-border.bg-card.p-4.mb-6')
    this.searchInput = page.locator('input[placeholder*="Search activities"]')
    this.searchClearButton = this.filtersPanel.locator('button').filter({
      has: page.locator('[class*="lucide-x"]'),
    })

    // quick sport filter toggle group
    this.quickSportFilters = page.locator('[role="group"]').first()
    this.allSportButton = page.locator('[role="group"] button:has-text("All")').first()

    // "More" button
    this.moreFiltersButton = page.locator('button:has-text("More")')

    // advanced filters (visible after clicking "More")
    this.advancedFiltersSection = page.locator('.border-t.border-border')
    this.sportTypeDropdown = page.locator('button[role="combobox"]').first()
    this.dateFromInput = page.locator('input[type="date"]').first()
    this.dateToInput = page.locator('input[type="date"]').nth(1)
    this.commuteCheckbox = page.locator('label:has-text("Commute") input[type="checkbox"]')
    this.trainerCheckbox = page.locator('label:has-text("Trainer") input[type="checkbox"]')
    this.minDistanceInput = page
      .locator('label:has-text("Min Distance")')
      .locator('..')
      .locator('input')
    this.maxDistanceInput = page
      .locator('label:has-text("Max Distance")')
      .locator('..')
      .locator('input')
    this.minDurationInput = page
      .locator('label:has-text("Min Duration")')
      .locator('..')
      .locator('input')
    this.maxDurationInput = page
      .locator('label:has-text("Max Duration")')
      .locator('..')
      .locator('input')

    // clear filters
    this.clearFiltersButton = page.locator('button:has-text("Clear filters")')

    // activities table
    this.activitiesTable = page.locator('.rounded-lg.border.border-border').filter({
      has: page.locator('table'),
    })
    this.tableHeader = page.locator('thead')
    this.tableBody = page.locator('tbody')
    this.tableRows = page.locator('tbody tr')
    this.emptyState = page.locator('text=No activities found.')
    this.loadingSkeleton = page.locator('[class*="animate-pulse"]')

    // table column headers
    this.activityHeader = page.locator('th:has-text("Activity")')
    this.typeHeader = page.locator('th:has-text("Type")')
    this.dateHeader = page.locator('th:has-text("Date")')
    this.distanceHeader = page.locator('th:has-text("Distance")')
    this.timeHeader = page.locator('th:has-text("Time")')
    this.elevationHeader = page.locator('th:has-text("Elevation")')
    this.tagsHeader = page.locator('th:has-text("Tags")')

    // pagination
    this.paginationContainer = page.locator('.flex.items-center.justify-between.px-2.py-4')
    this.showingInfo = page.locator('text=Showing')
    this.previousButton = page.locator('button:has-text("Previous")')
    this.nextButton = page.locator('button:has-text("Next")')
    this.pageInfo = page.locator('text=/Page \\d+ of \\d+/')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/activities')
  }

  // wait for table to be loaded (not in loading state)
  async waitForTableLoad(): Promise<void> {
    // wait for loading skeletons to disappear
    await this.loadingSkeleton.first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {
      // skeletons may not exist if data loads fast
    })
    // then wait for either table rows or empty state
    await Promise.race([
      this.tableRows.first().waitFor({ state: 'visible', timeout: 10000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // get the number of visible activity rows
  async getActivityRowCount(): Promise<number> {
    return await this.tableRows.count()
  }

  // get activity name from a specific row
  async getActivityName(rowIndex: number): Promise<string | null> {
    const row = this.tableRows.nth(rowIndex)
    const link = row.locator('a').first()
    return await link.textContent()
  }

  // click on a specific activity row
  async clickActivityRow(rowIndex: number): Promise<void> {
    const row = this.tableRows.nth(rowIndex)
    const link = row.locator('a').first()
    await link.click()
  }

  // get the sport type for a specific row
  async getActivitySportType(rowIndex: number): Promise<string | null> {
    const row = this.tableRows.nth(rowIndex)
    const typeCell = row.locator('td').nth(1)
    return await typeCell.textContent()
  }

  // filter by quick sport type button
  async selectQuickSportFilter(sportType: string): Promise<void> {
    const button = this.page.locator(`[role="group"] button:has-text("${sportType}")`).first()
    await button.click()
    await this.waitForTableLoad()
  }

  // open advanced filters
  async openAdvancedFilters(): Promise<void> {
    // check if already open
    const isOpen = await this.advancedFiltersSection.isVisible()
    if (!isOpen) {
      await this.moreFiltersButton.click()
      await this.advancedFiltersSection.waitFor({ state: 'visible' })
    }
  }

  // set sport type via dropdown in advanced filters
  async selectSportTypeDropdown(sportType: string): Promise<void> {
    await this.openAdvancedFilters()
    await this.sportTypeDropdown.click()
    await this.page.locator(`[role="option"]:has-text("${sportType}")`).click()
    await this.waitForTableLoad()
  }

  // set date range filter
  async setDateRange(from?: string, to?: string): Promise<void> {
    await this.openAdvancedFilters()
    if (from) {
      await this.dateFromInput.fill(from)
    }
    if (to) {
      await this.dateToInput.fill(to)
    }
    // wait for debounce and reload
    await this.page.waitForTimeout(600)
    await this.waitForTableLoad()
  }

  // toggle commute filter
  async toggleCommuteFilter(): Promise<void> {
    await this.openAdvancedFilters()
    // the checkbox is in a label, we click the label
    await this.page.locator('label:has-text("Commute")').click()
    await this.page.waitForTimeout(300)
    await this.waitForTableLoad()
  }

  // toggle trainer filter
  async toggleTrainerFilter(): Promise<void> {
    await this.openAdvancedFilters()
    await this.page.locator('label:has-text("Trainer")').click()
    await this.page.waitForTimeout(300)
    await this.waitForTableLoad()
  }

  // set distance range
  async setDistanceRange(min?: string, max?: string): Promise<void> {
    await this.openAdvancedFilters()
    if (min) {
      await this.minDistanceInput.fill(min)
    }
    if (max) {
      await this.maxDistanceInput.fill(max)
    }
    // wait for debounce
    await this.page.waitForTimeout(600)
    await this.waitForTableLoad()
  }

  // set duration range
  async setDurationRange(min?: string, max?: string): Promise<void> {
    await this.openAdvancedFilters()
    if (min) {
      await this.minDurationInput.fill(min)
    }
    if (max) {
      await this.maxDurationInput.fill(max)
    }
    // wait for debounce
    await this.page.waitForTimeout(600)
    await this.waitForTableLoad()
  }

  // search for activities
  async searchActivities(query: string): Promise<void> {
    await this.searchInput.fill(query)
    // wait for debounce
    await this.page.waitForTimeout(400)
    await this.waitForTableLoad()
  }

  // clear search
  async clearSearch(): Promise<void> {
    await this.searchInput.clear()
    await this.page.waitForTimeout(400)
    await this.waitForTableLoad()
  }

  // clear all filters
  async clearAllFilters(): Promise<void> {
    const clearVisible = await this.clearFiltersButton.isVisible()
    if (clearVisible) {
      await this.clearFiltersButton.click()
      await this.waitForTableLoad()
    }
  }

  // pagination helpers
  async goToNextPage(): Promise<void> {
    const isDisabled = await this.nextButton.isDisabled()
    if (!isDisabled) {
      await this.nextButton.click()
      await this.waitForTableLoad()
    }
  }

  async goToPreviousPage(): Promise<void> {
    const isDisabled = await this.previousButton.isDisabled()
    if (!isDisabled) {
      await this.previousButton.click()
      await this.waitForTableLoad()
    }
  }

  async getCurrentPageNumber(): Promise<number> {
    const text = await this.pageInfo.textContent()
    const match = text?.match(/Page (\d+) of/)
    return match ? parseInt(match[1]) : 1
  }

  async getTotalPages(): Promise<number> {
    const text = await this.pageInfo.textContent()
    const match = text?.match(/of (\d+)/)
    return match ? parseInt(match[1]) : 1
  }

  async getTotalActivitiesCount(): Promise<number> {
    const text = await this.showingInfo.textContent()
    const match = text?.match(/of\s*(\d+)/)
    return match ? parseInt(match[1]) : 0
  }

  // check if showing info displays correct range
  async getShowingRange(): Promise<{ start: number; end: number; total: number }> {
    const text = await this.paginationContainer.locator('p').first().textContent()
    const match = text?.match(/Showing\s+(\d+)\s+to\s+(\d+)\s+of\s+(\d+)/)
    if (match) {
      return {
        start: parseInt(match[1]),
        end: parseInt(match[2]),
        total: parseInt(match[3]),
      }
    }
    return { start: 0, end: 0, total: 0 }
  }

  // check if row has commute badge
  async rowHasCommuteBadge(rowIndex: number): Promise<boolean> {
    const row = this.tableRows.nth(rowIndex)
    const badge = row.locator('text=Commute')
    return await badge.isVisible()
  }

  // check if row has indoor badge
  async rowHasIndoorBadge(rowIndex: number): Promise<boolean> {
    const row = this.tableRows.nth(rowIndex)
    const badge = row.locator('text=Indoor')
    return await badge.isVisible()
  }
}
