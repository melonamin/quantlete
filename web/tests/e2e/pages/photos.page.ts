import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the photos gallery page
// provides selectors and helpers for interacting with photo grid, filters, and lightbox
export class PhotosPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // sport filter buttons
  readonly allSportButton: Locator
  readonly rideSportButton: Locator
  readonly runSportButton: Locator
  readonly walkSportButton: Locator
  readonly hikeSportButton: Locator

  // country dropdown
  readonly countryDropdownButton: Locator
  readonly countryDropdownPanel: Locator
  readonly allCountriesOption: Locator

  // more sports button
  readonly moreSportsButton: Locator
  readonly moreSportsPanel: Locator

  // clear filters button
  readonly clearFiltersButton: Locator

  // photo count display
  readonly photoCount: Locator

  // photo gallery (masonry grid)
  readonly photoGallery: Locator
  readonly photoItems: Locator
  readonly loadMoreButton: Locator

  // lightbox
  readonly lightbox: Locator
  readonly lightboxImage: Locator
  readonly lightboxTitle: Locator
  readonly lightboxDate: Locator
  readonly lightboxPrevButton: Locator
  readonly lightboxNextButton: Locator
  readonly lightboxCloseButton: Locator
  readonly lightboxViewActivityLink: Locator

  // loading and empty states
  readonly loadingSkeletons: Locator
  readonly emptyState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Photos' })
    this.pageSubtitle = page.locator('text=Your activity photo gallery')

    // sport filter buttons
    this.allSportButton = page.locator('.rounded-lg.border.bg-card button').filter({ hasText: /^All$/ })
    this.rideSportButton = page.locator('.rounded-lg.border.bg-card button').filter({ hasText: 'Ride' })
    this.runSportButton = page.locator('.rounded-lg.border.bg-card button').filter({ hasText: 'Run' })
    this.walkSportButton = page.locator('.rounded-lg.border.bg-card button').filter({ hasText: 'Walk' })
    this.hikeSportButton = page.locator('.rounded-lg.border.bg-card button').filter({ hasText: 'Hike' })

    // country dropdown
    this.countryDropdownButton = page.locator('button').filter({ has: page.locator('text=/countries/') })
    this.countryDropdownPanel = page.locator('.absolute.top-full.z-50.w-64')
    this.allCountriesOption = this.countryDropdownPanel.locator('button:has-text("All countries")')

    // more sports button
    this.moreSportsButton = page.locator('button:has-text("More sports")')
    this.moreSportsPanel = page.locator('.pt-3.border-t')

    // clear filters button
    this.clearFiltersButton = page.locator('button:has-text("Clear")')

    // photo count
    this.photoCount = page.locator('text=/\\d+ photos?/')

    // photo gallery (masonry grid using columns-* classes)
    this.photoGallery = page.locator('.columns-2')
    this.photoItems = this.photoGallery.locator('button.mb-3')
    this.loadMoreButton = page.locator('button:has-text("Load more")')

    // lightbox
    this.lightbox = page.locator('.fixed.inset-0.z-50.bg-black\\/80')
    this.lightboxImage = this.lightbox.locator('img.h-full.w-full')
    this.lightboxTitle = this.lightbox.locator('.truncate.font-medium')
    this.lightboxDate = this.lightbox.locator('.text-xs.text-white\\/70')
    this.lightboxPrevButton = this.lightbox.locator('button:has-text("Prev")')
    this.lightboxNextButton = this.lightbox.locator('button:has-text("Next")')
    this.lightboxCloseButton = this.lightbox.locator('button:has-text("Close")')
    this.lightboxViewActivityLink = this.lightbox.locator('a:has-text("View activity")')

    // loading and empty states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.emptyState = page.locator('text=No photos found')
    this.errorMessage = page.locator('text=Failed to load photos')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/photos')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either gallery, empty state, or error
    await Promise.race([
      this.photoGallery.waitFor({ state: 'visible', timeout: 10000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if gallery has photos
  async hasPhotos(): Promise<boolean> {
    try {
      await this.photoGallery.waitFor({ state: 'visible', timeout: 5000 })
      const count = await this.photoItems.count()
      return count > 0
    } catch {
      return false
    }
  }

  // get photo count from display
  async getDisplayedPhotoCount(): Promise<number | null> {
    try {
      const text = await this.photoCount.textContent()
      const match = text?.match(/(\d+) photos?/)
      return match ? parseInt(match[1], 10) : null
    } catch {
      return null
    }
  }

  // get actual photo items count
  async getPhotoItemsCount(): Promise<number> {
    return await this.photoItems.count()
  }

  // check if sport filter is selected
  async isSportFilterSelected(sport: 'all' | 'ride' | 'run' | 'walk' | 'hike'): Promise<boolean> {
    const buttons: Record<string, Locator> = {
      all: this.allSportButton,
      ride: this.rideSportButton,
      run: this.runSportButton,
      walk: this.walkSportButton,
      hike: this.hikeSportButton,
    }
    const classes = await buttons[sport].getAttribute('class')
    return classes?.includes('bg-primary') ?? false
  }

  // click sport filter
  async clickSportFilter(sport: 'all' | 'ride' | 'run' | 'walk' | 'hike'): Promise<void> {
    const buttons: Record<string, Locator> = {
      all: this.allSportButton,
      ride: this.rideSportButton,
      run: this.runSportButton,
      walk: this.walkSportButton,
      hike: this.hikeSportButton,
    }
    await buttons[sport].click()
    await this.page.waitForTimeout(300)
    await this.waitForLoadingComplete()
  }

  // open country dropdown
  async openCountryDropdown(): Promise<void> {
    await this.countryDropdownButton.click()
    await this.countryDropdownPanel.waitFor({ state: 'visible', timeout: 5000 })
  }

  // check if country dropdown is visible
  async isCountryDropdownOpen(): Promise<boolean> {
    return await this.countryDropdownPanel.isVisible()
  }

  // get country options count
  async getCountryOptionsCount(): Promise<number> {
    if (!(await this.isCountryDropdownOpen())) {
      await this.openCountryDropdown()
    }
    // subtract 1 for "All countries" option
    const count = await this.countryDropdownPanel.locator('button').count()
    return count - 1
  }

  // select country
  async selectCountry(countryName: string): Promise<void> {
    if (!(await this.isCountryDropdownOpen())) {
      await this.openCountryDropdown()
    }
    await this.countryDropdownPanel.locator(`button:has-text("${countryName}")`).first().click()
    await this.page.waitForTimeout(300)
    await this.waitForLoadingComplete()
  }

  // select all countries
  async selectAllCountries(): Promise<void> {
    if (!(await this.isCountryDropdownOpen())) {
      await this.openCountryDropdown()
    }
    await this.allCountriesOption.click()
    await this.page.waitForTimeout(300)
    await this.waitForLoadingComplete()
  }

  // get currently selected country (from button text)
  async getSelectedCountry(): Promise<string | null> {
    const text = await this.countryDropdownButton.textContent()
    // if text contains "countries", no specific country is selected
    if (text?.includes('countries')) return null
    return text?.trim() ?? null
  }

  // click photo by index
  async clickPhoto(index: number): Promise<void> {
    await this.photoItems.nth(index).click()
  }

  // check if lightbox is open
  async isLightboxOpen(): Promise<boolean> {
    return await this.lightbox.isVisible()
  }

  // get lightbox activity name
  async getLightboxTitle(): Promise<string | null> {
    return await this.lightboxTitle.textContent()
  }

  // navigate lightbox prev
  async lightboxPrev(): Promise<void> {
    await this.lightboxPrevButton.click()
    await this.page.waitForTimeout(200)
  }

  // navigate lightbox next
  async lightboxNext(): Promise<void> {
    await this.lightboxNextButton.click()
    await this.page.waitForTimeout(200)
  }

  // close lightbox
  async closeLightbox(): Promise<void> {
    await this.lightboxCloseButton.click()
    await this.lightbox.waitFor({ state: 'hidden', timeout: 5000 })
  }

  // check if load more button is visible
  async hasLoadMore(): Promise<boolean> {
    return await this.loadMoreButton.isVisible()
  }

  // click load more
  async loadMore(): Promise<void> {
    await this.loadMoreButton.click()
    await this.page.waitForTimeout(500)
    await this.waitForLoadingComplete()
  }

  // clear filters
  async clearFilters(): Promise<void> {
    if (await this.clearFiltersButton.isVisible()) {
      await this.clearFiltersButton.click()
      await this.page.waitForTimeout(300)
      await this.waitForLoadingComplete()
    }
  }

  // check if clear filters is visible (has active filters)
  async hasActiveFilters(): Promise<boolean> {
    return await this.clearFiltersButton.isVisible()
  }
}
