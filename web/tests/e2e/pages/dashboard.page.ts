import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the dashboard page
// provides selectors and helpers for interacting with dashboard widgets and stats
export class DashboardPage extends BasePage {
  // stats summary section at the top
  readonly statsSummary: Locator
  readonly totalActivitiesCard: Locator
  readonly totalDistanceCard: Locator
  readonly totalTimeCard: Locator
  readonly totalElevationCard: Locator

  // widget grid and customize button
  readonly widgetGrid: Locator
  readonly customizeButton: Locator
  readonly editModeBar: Locator

  // widget panel (sidebar for show/hide widgets)
  readonly widgetPanel: Locator
  readonly widgetPanelCloseButton: Locator
  readonly widgetPanelDoneButton: Locator

  // specific widgets
  readonly recentActivitiesWidget: Locator
  readonly weeklyStatsWidget: Locator
  readonly sportBreakdownWidget: Locator
  readonly monthlyChartWidget: Locator
  readonly activityCalendarWidget: Locator

  // monthly chart controls
  readonly monthlyChartYearButtons: Locator
  readonly monthlyChartMetricButtons: Locator

  constructor(page: Page) {
    super(page)

    // stats summary cards at the top
    this.statsSummary = page.locator('.grid.gap-3.md\\:grid-cols-2.lg\\:grid-cols-4').first()
    this.totalActivitiesCard = this.statsSummary.locator('[data-slot="card"]').filter({
      has: page.locator('[data-slot="card-title"]').filter({ hasText: /^Total Activities$/ }),
    })
    this.totalDistanceCard = this.statsSummary.locator('[data-slot="card"]').filter({
      has: page.locator('[data-slot="card-title"]').filter({ hasText: /^Total Distance$/ }),
    })
    this.totalTimeCard = this.statsSummary.locator('[data-slot="card"]').filter({
      has: page.locator('[data-slot="card-title"]').filter({ hasText: /^Total Time$/ }),
    })
    this.totalElevationCard = this.statsSummary.locator('[data-slot="card"]').filter({
      has: page.locator('[data-slot="card-title"]').filter({ hasText: /^Total Elevation$/ }),
    })

    // widget grid
    this.widgetGrid = page.locator('.grid.gap-4.md\\:grid-cols-12')
    this.customizeButton = page.locator('button:has-text("Customize Dashboard")')
    this.editModeBar = page.locator('text=EDIT MODE').locator('..')

    // widget panel (sidebar)
    this.widgetPanel = page.locator('text=Widget Manager').locator('..').locator('..')
    this.widgetPanelCloseButton = this.widgetPanel.locator('button').first()
    this.widgetPanelDoneButton = page.locator('button:has-text("Done")')

    // widgets by title (these are within the widget grid)
    this.recentActivitiesWidget = this.widgetGrid.locator('[data-slot="card"]').filter({
      has: page.locator('text=Recent Activities'),
    })
    this.weeklyStatsWidget = this.widgetGrid.locator('[data-slot="card"]').filter({
      has: page.locator('text=This Week'),
    })
    this.sportBreakdownWidget = this.widgetGrid.locator('[data-slot="card"]').filter({
      has: page.locator('text=By Sport'),
    })
    this.monthlyChartWidget = this.widgetGrid.locator('[data-slot="card"]').filter({
      has: page.locator('text=Monthly Activity'),
    })
    this.activityCalendarWidget = this.widgetGrid.locator('[data-slot="card"]').filter({
      has: page.locator('text=Activity Calendar'),
    })

    // monthly chart controls
    this.monthlyChartYearButtons = this.monthlyChartWidget.locator('button').filter({
      hasText: /^\d{4}$/,
    })
    this.monthlyChartMetricButtons = this.monthlyChartWidget.locator(
      'button:has-text("Distance"), button:has-text("Activities"), button:has-text("Time")'
    )
  }

  getPageIdentifier(): Locator {
    return this.page.locator('h1:has-text("Dashboard")')
  }

  async goto(): Promise<void> {
    await super.goto('/')
  }

  // stats summary helpers
  async getStatValue(cardLocator: Locator): Promise<string | null> {
    const valueElement = cardLocator.locator('[class*="text-2xl"], [class*="font-bold"]').first()
    return await valueElement.textContent()
  }

  async getTotalActivities(): Promise<string | null> {
    return await this.getStatValue(this.totalActivitiesCard)
  }

  async getTotalDistance(): Promise<string | null> {
    return await this.getStatValue(this.totalDistanceCard)
  }

  async getTotalTime(): Promise<string | null> {
    return await this.getStatValue(this.totalTimeCard)
  }

  async getTotalElevation(): Promise<string | null> {
    return await this.getStatValue(this.totalElevationCard)
  }

  // widget helpers
  async isWidgetVisible(widgetTitle: string): Promise<boolean> {
    const widget = this.page.locator('text=' + widgetTitle).first()
    try {
      await widget.waitFor({ state: 'visible', timeout: 5000 })
      return true
    } catch {
      return false
    }
  }

  async clickRecentActivity(index: number = 0): Promise<void> {
    // recent activities are links in the Recent Activities widget
    const activityLinks = this.page.locator('a[href*="/activities/"]:has(.truncate)')
    await activityLinks.nth(index).click()
  }

  async getRecentActivityCount(): Promise<number> {
    const activityLinks = this.page.locator('a[href*="/activities/"]:has(.truncate)')
    return await activityLinks.count()
  }

  // edit mode helpers
  async enterEditMode(): Promise<void> {
    await this.customizeButton.click()
    await this.editModeBar.waitFor({ state: 'visible' })
  }

  async exitEditMode(): Promise<void> {
    await this.page.locator('button:has-text("EXIT")').click()
    await this.editModeBar.waitFor({ state: 'hidden' })
  }

  async openWidgetPanel(): Promise<void> {
    await this.page.locator('button:has-text("WIDGETS")').click()
    await this.widgetPanel.waitFor({ state: 'visible' })
  }

  async closeWidgetPanel(): Promise<void> {
    await this.widgetPanelDoneButton.click()
    await this.widgetPanel.waitFor({ state: 'hidden' })
  }

  async toggleWidgetVisibility(widgetTitle: string): Promise<void> {
    // find the widget row in the panel and click the eye toggle
    const widgetRow = this.page.locator('.space-y-2 > div').filter({
      has: this.page.locator(`text="${widgetTitle}"`),
    })
    await widgetRow.locator('button').last().click()
  }

  // monthly chart helpers
  async selectMonthlyChartYear(year: number): Promise<void> {
    await this.monthlyChartWidget.locator(`button:has-text("${year}")`).click()
  }

  async selectMonthlyChartMetric(metric: 'Distance' | 'Activities' | 'Time'): Promise<void> {
    await this.monthlyChartWidget.locator(`button:has-text("${metric}")`).click()
  }

  async isMonthlyChartMetricSelected(metric: 'Distance' | 'Activities' | 'Time'): Promise<boolean> {
    const button = this.monthlyChartWidget.locator(`button:has-text("${metric}")`)
    const classes = await button.getAttribute('class')
    // secondary variant indicates selected state
    return classes?.includes('secondary') ?? false
  }
}
