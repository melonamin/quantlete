import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the heatmap page
// provides selectors and helpers for interacting with the heatmap, filters, and country dropdown
export class HeatmapPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator
  readonly routeCount: Locator

  // quick sport filters (toggle group)
  readonly quickSportFilters: Locator
  readonly allSportButton: Locator
  readonly rideSportButton: Locator
  readonly runSportButton: Locator
  readonly walkSportButton: Locator
  readonly swimSportButton: Locator

  // country dropdown
  readonly countryDropdownButton: Locator
  readonly countryDropdownPanel: Locator
  readonly countryVisitedCount: Locator
  readonly allCountriesButton: Locator

  // more filters button and advanced filters panel
  readonly moreFiltersButton: Locator
  readonly advancedFiltersPanel: Locator
  readonly customSportInput: Locator
  readonly dateFromInput: Locator
  readonly dateToInput: Locator
  readonly commuteDropdown: Locator

  // clear filters button
  readonly clearFiltersButton: Locator

  // map container
  readonly mapContainer: Locator
  readonly mapSkeleton: Locator
  readonly leafletContainer: Locator
  readonly leafletTileContainer: Locator
  readonly leafletSvgOverlay: Locator
  readonly leafletZoomIn: Locator
  readonly leafletZoomOut: Locator

  // error and empty states
  readonly errorMessage: Locator
  readonly emptyState: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Heatmap' })
    this.pageSubtitle = page.locator('text=Visualize all your activities on a map')
    this.routeCount = page.locator('text=/\\d+ routes?/')

    // quick sport filter toggle group (uses data-slot="toggle-group")
    this.quickSportFilters = page.locator('[data-slot="toggle-group"]').first()
    this.allSportButton = page.locator('[data-slot="toggle-group"] button:has-text("All")').first()
    this.rideSportButton = page
      .locator('[data-slot="toggle-group"] button:has-text("Ride")')
      .first()
    this.runSportButton = page.locator('[data-slot="toggle-group"] button:has-text("Run")').first()
    this.walkSportButton = page
      .locator('[data-slot="toggle-group"] button:has-text("Walk")')
      .first()
    this.swimSportButton = page
      .locator('[data-slot="toggle-group"] button:has-text("Swim")')
      .first()

    // country dropdown
    this.countryDropdownButton = page.locator('button').filter({
      has: page.locator('[class*="lucide-globe"]'),
    })
    this.countryDropdownPanel = page.locator('.absolute.z-50').filter({
      has: page.locator('text=/\\d+ visited/'),
    })
    this.countryVisitedCount = page.locator('text=/\\d+ visited/')
    this.allCountriesButton = page.locator('button:has-text("All countries")')

    // more filters - button has "More" text
    this.moreFiltersButton = page.locator('button:has-text("More")').first()
    this.advancedFiltersPanel = page.locator('.grid.grid-cols-2.md\\:grid-cols-4')
    // use type selectors to distinguish text vs date inputs
    this.customSportInput = page.locator('input[type="text"][placeholder*="Ride"]')
    this.dateFromInput = page.locator('input[type="date"]').first()
    this.dateToInput = page.locator('input[type="date"]').last()
    this.commuteDropdown = page.locator('label:has-text("Commute")').locator('..').locator('button')

    // clear filters
    this.clearFiltersButton = page.locator('button:has-text("Clear")')

    // map
    this.mapContainer = page.locator('.flex-1').first()
    this.mapSkeleton = page.locator('[class*="animate-pulse"]').first()
    this.leafletContainer = page.locator('.leaflet-container').first()
    // leaflet-tile-pane is always present when map renders
    this.leafletTileContainer = page.locator('.leaflet-tile-pane').first()
    this.leafletSvgOverlay = page.locator('.leaflet-container svg').first()
    this.leafletZoomIn = page.locator('.leaflet-control-zoom-in').first()
    this.leafletZoomOut = page.locator('.leaflet-control-zoom-out').first()

    // error and empty states
    this.errorMessage = page.locator('text=Failed to load heatmap data')
    this.emptyState = page.locator('text=No activities with GPS data found')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/heatmap')
  }

  // wait for map to be loaded
  async waitForMapLoad(): Promise<void> {
    // wait for loading skeleton to disappear
    await this.mapSkeleton.waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {
      // skeleton may not exist if data loads fast
    })
    // then wait for either map or empty state
    await Promise.race([
      this.leafletContainer.waitFor({ state: 'visible', timeout: 15000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 15000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 15000 }),
    ])
  }

  // get the route count from header
  async getRouteCount(): Promise<number> {
    const text = await this.routeCount.textContent()
    const match = text?.match(/(\d+)\s+routes?/)
    return match ? parseInt(match[1]) : 0
  }

  // check if map has activity routes (SVG paths)
  async hasActivityRoutes(): Promise<boolean> {
    const svgVisible = await this.leafletSvgOverlay.isVisible()
    if (!svgVisible) return false
    // check for path elements (routes are rendered as SVG paths)
    const paths = this.leafletContainer.locator('svg path')
    const pathCount = await paths.count()
    return pathCount > 0
  }

  // select quick sport filter
  async selectQuickSportFilter(sport: 'All' | 'Ride' | 'Run' | 'Walk' | 'Swim'): Promise<void> {
    const button = this.page
      .locator(`[data-slot="toggle-group"] button:has-text("${sport}")`)
      .first()
    await button.click()
    await this.page.waitForTimeout(500)
    await this.waitForMapLoad()
  }

  // check if a sport filter button is selected (has data-state="on")
  async isSportFilterSelected(sport: string): Promise<boolean> {
    const button = this.page
      .locator(`[data-slot="toggle-group"] button:has-text("${sport}")`)
      .first()
    const state = await button.getAttribute('data-state')
    return state === 'on'
  }

  // open country dropdown
  async openCountryDropdown(): Promise<void> {
    const isVisible = await this.countryDropdownPanel.isVisible()
    if (!isVisible) {
      await this.countryDropdownButton.click()
      await this.countryDropdownPanel.waitFor({ state: 'visible', timeout: 5000 })
    }
  }

  // close country dropdown
  async closeCountryDropdown(): Promise<void> {
    const isVisible = await this.countryDropdownPanel.isVisible()
    if (isVisible) {
      // click outside to close
      await this.page.locator('.fixed.inset-0').click()
      await this.countryDropdownPanel.waitFor({ state: 'hidden', timeout: 5000 })
    }
  }

  // get list of countries from dropdown
  async getCountryList(): Promise<{ name: string; count: number }[]> {
    await this.openCountryDropdown()
    const countryButtons = this.countryDropdownPanel.locator('button').filter({
      hasNot: this.page.locator('text=All countries'),
    })
    const count = await countryButtons.count()
    const countries: { name: string; count: number }[] = []
    for (let i = 0; i < count; i++) {
      const button = countryButtons.nth(i)
      const text = await button.textContent()
      // text format: "🇺🇸 United States 42"
      const match = text?.match(/(.+?)\s+(\d+)$/)
      if (match) {
        countries.push({
          name: match[1].trim(),
          count: parseInt(match[2]),
        })
      }
    }
    return countries
  }

  // select a country from dropdown
  async selectCountry(countryName: string): Promise<void> {
    await this.openCountryDropdown()
    const countryButton = this.countryDropdownPanel.locator(`button:has-text("${countryName}")`)
    await countryButton.click()
    await this.page.waitForTimeout(1500) // wait for flyTo animation
  }

  // select "All countries" from dropdown
  async selectAllCountries(): Promise<void> {
    await this.openCountryDropdown()
    await this.allCountriesButton.click()
    await this.page.waitForTimeout(1500) // wait for flyTo animation
  }

  // open advanced filters
  async openAdvancedFilters(): Promise<void> {
    const isVisible = await this.advancedFiltersPanel.isVisible()
    if (!isVisible) {
      await this.moreFiltersButton.click()
      await this.advancedFiltersPanel.waitFor({ state: 'visible', timeout: 5000 })
    }
  }

  // close advanced filters
  async closeAdvancedFilters(): Promise<void> {
    const isVisible = await this.advancedFiltersPanel.isVisible()
    if (isVisible) {
      await this.moreFiltersButton.click()
      await this.advancedFiltersPanel.waitFor({ state: 'hidden', timeout: 5000 })
    }
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
    await this.page.waitForTimeout(600)
    await this.waitForMapLoad()
  }

  // set commute filter
  async setCommuteFilter(value: 'All' | 'Commute only' | 'Non-commute'): Promise<void> {
    await this.openAdvancedFilters()
    await this.commuteDropdown.click()
    await this.page.locator(`[role="option"]:has-text("${value}")`).click()
    await this.page.waitForTimeout(500)
    await this.waitForMapLoad()
  }

  // clear all filters
  async clearAllFilters(): Promise<void> {
    const clearVisible = await this.clearFiltersButton.isVisible()
    if (clearVisible) {
      await this.clearFiltersButton.click()
      await this.page.waitForTimeout(500)
      await this.waitForMapLoad()
    }
  }

  // check if map is interactive (has zoom controls)
  async hasZoomControls(): Promise<boolean> {
    const zoomInVisible = await this.leafletZoomIn.isVisible()
    const zoomOutVisible = await this.leafletZoomOut.isVisible()
    return zoomInVisible || zoomOutVisible
  }

  // check if tiles are loaded
  async hasTilesLoaded(): Promise<boolean> {
    return await this.leafletTileContainer.isVisible()
  }
}
