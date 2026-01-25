import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the activity detail page
// provides selectors and helpers for interacting with activity stats, charts, and map
export class ActivityDetailPage extends BasePage {
  // back navigation
  readonly backLink: Locator

  // header section
  readonly activityTitle: Locator
  readonly sportType: Locator
  readonly activityDate: Locator
  readonly stravaLink: Locator

  // badges
  readonly commuteBadge: Locator
  readonly indoorBadge: Locator
  readonly privateBadge: Locator

  // stats cards
  readonly statsGrid: Locator
  readonly distanceCard: Locator
  readonly movingTimeCard: Locator
  readonly elevationCard: Locator
  readonly paceSpeedCard: Locator
  readonly heartRateCard: Locator
  readonly powerCard: Locator
  readonly caloriesCard: Locator

  // map section
  readonly activityMap: Locator

  // stream charts section
  readonly activityStreamsSection: Locator
  readonly streamChartTitle: Locator
  readonly hrCheckbox: Locator
  readonly powerCheckbox: Locator
  readonly cadenceCheckbox: Locator
  readonly elevCheckbox: Locator

  // elevation profile section
  readonly elevationProfileSection: Locator
  readonly elevationProfileTitle: Locator
  readonly gradientCheckbox: Locator

  // activity analysis section
  readonly analysisSection: Locator
  readonly analysisToggle: Locator
  readonly splitsSection: Locator
  readonly hrZonesSection: Locator
  readonly paceDistributionSection: Locator

  // photos section
  readonly photosSection: Locator
  readonly photosTitle: Locator
  readonly photoGrid: Locator
  readonly viewAllPhotosLink: Locator

  // loading and error states
  readonly loadingSkeleton: Locator
  readonly errorState: Locator

  constructor(page: Page) {
    super(page)

    // back navigation
    this.backLink = page.locator('a:has-text("Back to Activities")')

    // header
    this.activityTitle = page.locator('h1').first()
    this.sportType = page.locator('.text-muted-foreground').filter({ hasText: /Ride|Run|Walk|Swim|Hike/ }).first()
    this.activityDate = page.locator('.text-muted-foreground').filter({ hasText: /\d{1,2}.*\d{4}/ }).first()
    this.stravaLink = page.locator('a:has-text("View on Strava")')

    // badges
    this.commuteBadge = page.locator('[class*="badge"]:has-text("Commute")')
    this.indoorBadge = page.locator('[class*="badge"]:has-text("Indoor")')
    this.privateBadge = page.locator('[class*="badge"]:has-text("Private")')

    // stats cards - using Card component structure
    this.statsGrid = page.locator('.grid.gap-4').first()
    this.distanceCard = page.locator('[data-slot="card"]').filter({ hasText: 'Distance' }).first()
    this.movingTimeCard = page.locator('[data-slot="card"]').filter({ hasText: 'Moving Time' }).first()
    this.elevationCard = page.locator('[data-slot="card"]').filter({ hasText: 'Elevation' }).first()
    this.paceSpeedCard = page.locator('[data-slot="card"]').filter({ hasText: /Pace|Speed/ }).first()
    this.heartRateCard = page.locator('[data-slot="card"]').filter({ hasText: 'Heart Rate' }).first()
    this.powerCard = page.locator('[data-slot="card"]').filter({ hasText: 'Power' }).first()
    this.caloriesCard = page.locator('[data-slot="card"]').filter({ hasText: 'Calories' }).first()

    // map
    this.activityMap = page.locator('.leaflet-container').first()

    // stream charts section
    this.activityStreamsSection = page.locator('.rounded-lg.border').filter({ hasText: 'Activity Streams' }).first()
    this.streamChartTitle = page.locator('h2:has-text("Activity Streams")')
    this.hrCheckbox = page.locator('label:has-text("HR") input[type="checkbox"], label:has-text("HR") button[role="checkbox"]')
    this.powerCheckbox = page.locator('label:has-text("Power") input[type="checkbox"], label:has-text("Power") button[role="checkbox"]')
    this.cadenceCheckbox = page.locator('label:has-text("Cadence") input[type="checkbox"], label:has-text("Cadence") button[role="checkbox"]')
    this.elevCheckbox = page.locator('label:has-text("Elev") input[type="checkbox"], label:has-text("Elev") button[role="checkbox"]')

    // elevation profile section
    this.elevationProfileSection = page.locator('.rounded-lg.border').filter({ hasText: 'Elevation Profile' }).first()
    this.elevationProfileTitle = page.locator('h2:has-text("Elevation Profile")')
    this.gradientCheckbox = page.locator('label:has-text("Gradient") input[type="checkbox"], label:has-text("Gradient") button[role="checkbox"]')

    // activity analysis section (collapsible)
    this.analysisSection = page.locator('.rounded-lg.border').filter({ hasText: 'Activity Analysis' })
    this.analysisToggle = page.locator('button:has-text("Activity Analysis")')
    this.splitsSection = page.locator('.rounded-lg.border').filter({ hasText: 'Splits' })
    this.hrZonesSection = page.locator('.rounded-lg.border').filter({ hasText: 'HR Zones' })
    this.paceDistributionSection = page.locator('.rounded-lg.border').filter({ hasText: 'Pace Distribution' })

    // photos section
    this.photosSection = page.locator('.rounded-lg.border').filter({ hasText: 'Photos' })
    this.photosTitle = page.locator('h2:has-text("Photos")')
    this.photoGrid = this.photosSection.locator('.grid')
    this.viewAllPhotosLink = page.locator('a:has-text("View all")')

    // loading and error
    this.loadingSkeleton = page.locator('[class*="animate-pulse"]')
    this.errorState = page.locator('.border-destructive')
  }

  getPageIdentifier(): Locator {
    return this.activityTitle
  }

  async goto(activityId: number): Promise<void> {
    await super.goto(`/activities/${activityId}`)
  }

  // wait for activity data to be loaded
  async waitForActivityLoad(): Promise<void> {
    // wait for loading skeletons to disappear
    await this.loadingSkeleton.first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {
      // skeletons may not exist if data loads fast
    })
    // then wait for activity title to be visible
    await this.activityTitle.waitFor({ state: 'visible', timeout: 10000 })
  }

  // get the activity name from the header
  async getActivityName(): Promise<string | null> {
    return await this.activityTitle.textContent()
  }

  // get the sport type text
  async getSportType(): Promise<string | null> {
    return await this.sportType.textContent()
  }

  // check if the map is visible (activity has GPS data)
  async hasMap(): Promise<boolean> {
    try {
      await this.activityMap.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  // check if activity streams section is visible
  async hasActivityStreams(): Promise<boolean> {
    return await this.activityStreamsSection.isVisible()
  }

  // check if elevation profile is visible
  async hasElevationProfile(): Promise<boolean> {
    return await this.elevationProfileSection.isVisible()
  }

  // check if photos section is visible
  async hasPhotos(): Promise<boolean> {
    return await this.photosSection.isVisible()
  }

  // get the number of photos displayed
  async getPhotoCount(): Promise<number> {
    if (!(await this.hasPhotos())) return 0
    return await this.photoGrid.locator('a').count()
  }

  // toggle analysis section visibility
  async toggleAnalysis(): Promise<void> {
    await this.analysisToggle.click()
    await this.page.waitForTimeout(300)
  }

  // click back to activities
  async goBackToActivities(): Promise<void> {
    await this.backLink.click()
    await this.page.waitForURL(/\/activities$/, { timeout: 10000 })
  }

  // get stat card value by label
  async getStatValue(label: string): Promise<string | null> {
    const card = this.page.locator('[data-slot="card"]').filter({ hasText: label }).first()
    const value = card.locator('.text-2xl.font-bold')
    return await value.textContent()
  }

  // check if a specific stat card exists
  async hasStatCard(label: string): Promise<boolean> {
    const card = this.page.locator('[data-slot="card"]').filter({ hasText: label }).first()
    return await card.isVisible()
  }

  // get all visible stat card labels
  async getVisibleStatLabels(): Promise<string[]> {
    const cards = this.page.locator('[data-slot="card"] .text-sm.text-muted-foreground')
    const count = await cards.count()
    const labels: string[] = []
    for (let i = 0; i < count; i++) {
      const text = await cards.nth(i).textContent()
      if (text) labels.push(text)
    }
    return labels
  }
}
