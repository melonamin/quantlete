import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the challenges page
// provides selectors and helpers for interacting with challenge badges grouped by month
export class ChallengesPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // import card
  readonly importCard: Locator
  readonly importFileInput: Locator
  readonly importButton: Locator
  readonly importMessage: Locator

  // month cards (grouped challenges)
  readonly monthCards: Locator

  // challenge badges
  readonly challengeBadges: Locator
  readonly badgeImages: Locator
  readonly badgeLinks: Locator

  // loading and empty states
  readonly loadingSkeletons: Locator
  readonly emptyState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Challenges' })
    this.pageSubtitle = page.locator('text=Completed Strava challenges')

    // import card
    this.importCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Import'),
    })
    this.importFileInput = this.importCard.locator('input[type="file"]')
    this.importButton = this.importCard.locator('button:has-text("Import")')
    this.importMessage = this.importCard.locator('.text-sm.text-muted-foreground').last()

    // month cards (each card has a month title)
    this.monthCards = page.locator('[class*="card"]').filter({
      has: page.locator('[class*="CardTitle"]'),
    })

    // challenge badges (links inside month cards, excluding import card)
    this.challengeBadges = page.locator('a.group.rounded-md.border')
    this.badgeImages = this.challengeBadges.locator('img')
    this.badgeLinks = this.challengeBadges

    // loading and empty states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.emptyState = page.locator('text=No challenges imported yet')
    this.errorMessage = page.locator('text=Failed to load challenges')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/challenges')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either challenges, empty state, or error
    await Promise.race([
      this.challengeBadges.first().waitFor({ state: 'visible', timeout: 10000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ])
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if page has challenges
  async hasChallenges(): Promise<boolean> {
    try {
      const count = await this.challengeBadges.count()
      return count > 0
    } catch {
      return false
    }
  }

  // get challenge count
  async getChallengeCount(): Promise<number> {
    return await this.challengeBadges.count()
  }

  // get month cards count (excluding import card)
  async getMonthCardsCount(): Promise<number> {
    // count cards that have month titles (e.g., "January 2024")
    const cards = this.page.locator('[class*="card"]').filter({
      has: this.page.locator('[class*="CardTitle"]:not(:has-text("Import"))'),
    })
    return await cards.count()
  }

  // get month labels
  async getMonthLabels(): Promise<string[]> {
    const titles = this.page.locator('[class*="CardTitle"]:not(:has-text("Import"))')
    const count = await titles.count()
    const labels: string[] = []
    for (let i = 0; i < count; i++) {
      const text = await titles.nth(i).textContent()
      if (text) {
        // extract month name (before count in parentheses)
        const match = text.match(/^([^(]+)/)
        if (match) {
          labels.push(match[1].trim())
        }
      }
    }
    return labels
  }

  // check if badge images loaded
  async badgeImagesLoaded(): Promise<boolean> {
    const images = this.badgeImages
    const count = await images.count()
    if (count === 0) return true // no images to check

    // check first few images have src
    const checkCount = Math.min(count, 5)
    for (let i = 0; i < checkCount; i++) {
      const src = await images.nth(i).getAttribute('src')
      if (!src) return false
    }
    return true
  }

  // get badge href by index
  async getBadgeHref(index: number): Promise<string | null> {
    return await this.badgeLinks.nth(index).getAttribute('href')
  }

  // check if badge links to Strava
  async badgeLinksToStrava(index: number): Promise<boolean> {
    const href = await this.getBadgeHref(index)
    return href?.includes('strava.com/challenges/') ?? false
  }

  // get badge name by index
  async getBadgeName(index: number): Promise<string | null> {
    const badge = this.challengeBadges.nth(index)
    const nameElement = badge.locator('.text-xs.font-medium')
    return await nameElement.textContent()
  }

  // click badge by index (will open external link)
  async clickBadge(index: number): Promise<void> {
    await this.challengeBadges.nth(index).click()
  }

  // check if import card is visible
  async isImportCardVisible(): Promise<boolean> {
    return await this.importCard.isVisible()
  }
}
