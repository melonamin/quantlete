import type { Page, Locator } from '@playwright/test'

// base page object that all page objects extend
// provides common functionality and navigation helpers
export abstract class BasePage {
  readonly page: Page

  // common selectors that appear on all pages
  protected readonly header: Locator
  protected readonly sidebar: Locator
  protected readonly mainContent: Locator

  constructor(page: Page) {
    this.page = page
    this.header = page.locator('header').first()
    this.sidebar = page.locator('nav').first()
    this.mainContent = page.locator('main').first()
  }

  // abstract method - each page must define its unique selector
  abstract getPageIdentifier(): Locator

  // wait for the page to be fully loaded
  async waitForPageLoad(): Promise<void> {
    await this.getPageIdentifier().waitFor({ state: 'visible', timeout: 10000 })
  }

  // navigate to a specific route
  async goto(path: string): Promise<void> {
    await this.page.goto(path)
    await this.waitForPageLoad()
  }

  // get the page title text
  async getPageTitle(): Promise<string | null> {
    return await this.page.locator('h1').first().textContent()
  }

  // check if page has loaded successfully
  async isLoaded(): Promise<boolean> {
    try {
      await this.getPageIdentifier().waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // navigation helpers
  async navigateToDashboard(): Promise<void> {
    await this.page.getByRole('link', { name: /dashboard/i }).click()
  }

  async navigateToActivities(): Promise<void> {
    await this.page.getByRole('link', { name: /activities/i }).click()
  }

  async navigateToSettings(): Promise<void> {
    await this.page.getByRole('link', { name: /settings/i }).click()
  }

  // wait for loading states to clear
  async waitForLoadingComplete(): Promise<void> {
    // wait for any loading skeletons to disappear
    const skeletons = this.page.locator('[class*="animate-pulse"]')
    await skeletons.first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {
      // skeletons may not exist if data loads quickly
    })
  }

  // helper to click and wait for navigation
  async clickAndWaitForNavigation(locator: Locator): Promise<void> {
    const currentURL = this.page.url()
    await Promise.all([
      this.page.waitForURL((url) => url.toString() !== currentURL, { timeout: 10000 }),
      locator.click(),
    ])
  }

  // helper to scroll element into view
  async scrollIntoView(locator: Locator): Promise<void> {
    await locator.scrollIntoViewIfNeeded()
  }

  // helper to take screenshot with context
  async screenshot(name: string): Promise<void> {
    await this.page.screenshot({ path: `test-results/${name}.png`, fullPage: true })
  }
}
