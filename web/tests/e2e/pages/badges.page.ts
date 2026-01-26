import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the badges customizer page
// provides selectors and helpers for interacting with badge previews and customizer controls
export class BadgesPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // public access card (server mode only)
  readonly publicAccessCard: Locator
  readonly publicBadgesCheckbox: Locator

  // style customizer card
  readonly styleCard: Locator
  readonly customizer: Locator

  // theme selector buttons
  readonly themeButtons: Locator

  // background selector buttons
  readonly backgroundButtons: Locator

  // size selector buttons
  readonly sizeButtons: Locator

  // overall stats card
  readonly overallStatsCard: Locator
  readonly distanceBadge: Locator
  readonly timeBadge: Locator
  readonly elevationBadge: Locator
  readonly activitiesBadge: Locator

  // achievements card
  readonly achievementsCard: Locator
  readonly eddingtonBadge: Locator
  readonly prBadge: Locator

  // time periods card
  readonly timePeriodsCard: Locator
  readonly yearlyBadge: Locator
  readonly monthlyBadge: Locator

  // badge preview elements
  readonly badgePreviews: Locator
  readonly badgeSvgs: Locator

  // badge action buttons
  readonly downloadButtons: Locator
  readonly copyButtons: Locator

  // loading and empty states
  readonly loadingSkeletons: Locator
  readonly emptyState: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Badges' })
    this.pageSubtitle = page.locator('text=Generate SVG badges from your stats')

    // public access card
    this.publicAccessCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Public Access'),
    })
    this.publicBadgesCheckbox = this.publicAccessCard.locator('button[role="checkbox"]')

    // style customizer card
    this.styleCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Style'),
    })
    this.customizer = this.styleCard.locator('.flex.flex-wrap.gap-4')

    // theme selector (Accent section)
    this.themeButtons = page.locator('.space-y-2').filter({
      has: page.locator('text=Accent'),
    }).locator('button')

    // background selector
    this.backgroundButtons = page.locator('.space-y-2').filter({
      has: page.locator('text=Background'),
    }).locator('button')

    // size selector
    this.sizeButtons = page.locator('.space-y-2').filter({
      has: page.locator('text=Size'),
    }).locator('button')

    // overall stats card
    this.overallStatsCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Overall Stats'),
    })
    this.distanceBadge = this.overallStatsCard.locator('.flex.flex-col.gap-3').first()
    this.timeBadge = this.overallStatsCard.locator('.flex.flex-col.gap-3').nth(1)
    this.elevationBadge = this.overallStatsCard.locator('.flex.flex-col.gap-3').nth(2)
    this.activitiesBadge = this.overallStatsCard.locator('.flex.flex-col.gap-3').nth(3)

    // achievements card
    this.achievementsCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Achievements'),
    })
    this.eddingtonBadge = this.achievementsCard.locator('.flex.flex-col.gap-3').first()
    this.prBadge = this.achievementsCard.locator('.flex.flex-col.gap-3').nth(1)

    // time periods card
    this.timePeriodsCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Time Periods'),
    })
    this.yearlyBadge = this.timePeriodsCard.locator('.flex.flex-col.gap-3').first()
    this.monthlyBadge = this.timePeriodsCard.locator('.flex.flex-col.gap-3').nth(1)

    // badge previews (all svgs in preview containers)
    this.badgePreviews = page.locator('.flex.justify-center.rounded-lg.border')
    this.badgeSvgs = page.locator('.flex.justify-center.rounded-lg.border svg')

    // badge action buttons
    this.downloadButtons = page.locator('button:has-text("SVG")')
    this.copyButtons = page.locator('button:has-text("Copy")')

    // loading and empty states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.emptyState = page.locator('text=No stats data available')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/badges')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either badges or empty state
    await Promise.race([
      this.badgeSvgs.first().waitFor({ state: 'visible', timeout: 15000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 15000 }),
    ])
  }

  // check if page has badges
  async hasBadges(): Promise<boolean> {
    try {
      const count = await this.badgeSvgs.count()
      return count > 0
    } catch {
      return false
    }
  }

  // get badge count
  async getBadgeCount(): Promise<number> {
    return await this.badgeSvgs.count()
  }

  // get currently selected theme
  async getSelectedTheme(): Promise<string | null> {
    const buttons = this.themeButtons
    const count = await buttons.count()
    for (let i = 0; i < count; i++) {
      const classes = await buttons.nth(i).getAttribute('class')
      if (classes?.includes('border-primary')) {
        return await buttons.nth(i).textContent()
      }
    }
    return null
  }

  // select theme by name
  async selectTheme(themeName: string): Promise<void> {
    const button = this.themeButtons.filter({ hasText: themeName })
    await button.click()
    await this.page.waitForTimeout(200)
  }

  // get available themes
  async getAvailableThemes(): Promise<string[]> {
    const buttons = this.themeButtons
    const count = await buttons.count()
    const themes: string[] = []
    for (let i = 0; i < count; i++) {
      const text = await buttons.nth(i).textContent()
      if (text) themes.push(text.trim())
    }
    return themes
  }

  // get currently selected background
  async getSelectedBackground(): Promise<string | null> {
    const buttons = this.backgroundButtons
    const count = await buttons.count()
    for (let i = 0; i < count; i++) {
      const classes = await buttons.nth(i).getAttribute('class')
      if (classes?.includes('border-primary')) {
        return await buttons.nth(i).textContent()
      }
    }
    return null
  }

  // select background by name
  async selectBackground(bgName: string): Promise<void> {
    const button = this.backgroundButtons.filter({ hasText: bgName })
    await button.click()
    await this.page.waitForTimeout(200)
  }

  // get available backgrounds
  async getAvailableBackgrounds(): Promise<string[]> {
    const buttons = this.backgroundButtons
    const count = await buttons.count()
    const backgrounds: string[] = []
    for (let i = 0; i < count; i++) {
      const text = await buttons.nth(i).textContent()
      if (text) backgrounds.push(text.trim())
    }
    return backgrounds
  }

  // get currently selected size
  async getSelectedSize(): Promise<string | null> {
    const buttons = this.sizeButtons
    const count = await buttons.count()
    for (let i = 0; i < count; i++) {
      const classes = await buttons.nth(i).getAttribute('class')
      if (classes?.includes('border-primary')) {
        return await buttons.nth(i).textContent()
      }
    }
    return null
  }

  // select size by name
  async selectSize(sizeName: string): Promise<void> {
    const button = this.sizeButtons.filter({ hasText: sizeName })
    await button.click()
    await this.page.waitForTimeout(200)
  }

  // get available sizes
  async getAvailableSizes(): Promise<string[]> {
    const buttons = this.sizeButtons
    const count = await buttons.count()
    const sizes: string[] = []
    for (let i = 0; i < count; i++) {
      const text = await buttons.nth(i).textContent()
      if (text) sizes.push(text.trim())
    }
    return sizes
  }

  // get first badge SVG dimensions
  async getFirstBadgeDimensions(): Promise<{ width: number; height: number } | null> {
    const svg = this.badgeSvgs.first()
    const width = await svg.getAttribute('width')
    const height = await svg.getAttribute('height')
    if (width && height) {
      return { width: parseInt(width, 10), height: parseInt(height, 10) }
    }
    return null
  }

  // click download button for first badge
  async clickDownloadButton(): Promise<void> {
    await this.downloadButtons.first().click()
  }

  // check if download buttons are visible
  async hasDownloadButtons(): Promise<boolean> {
    const count = await this.downloadButtons.count()
    return count > 0
  }

  // get first badge's fill color (background)
  async getFirstBadgeBackground(): Promise<string | null> {
    const svg = this.badgeSvgs.first()
    const rect = svg.locator('rect').first()
    return await rect.getAttribute('fill')
  }

  // get first badge's accent color (from text)
  async getFirstBadgeAccentColor(): Promise<string | null> {
    const svg = this.badgeSvgs.first()
    // the accent color is used in the value text (second text element)
    const texts = svg.locator('text')
    const count = await texts.count()
    if (count >= 2) {
      return await texts.nth(1).getAttribute('fill')
    }
    return null
  }
}
