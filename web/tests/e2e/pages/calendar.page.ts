import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the calendar page
// provides selectors and helpers for interacting with the calendar grid, navigation, and day modals
export class CalendarPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // navigation controls
  readonly todayButton: Locator
  readonly prevMonthButton: Locator
  readonly nextMonthButton: Locator
  readonly monthYearDisplay: Locator

  // stats cards
  readonly distanceCard: Locator
  readonly elevationCard: Locator
  readonly timeCard: Locator
  readonly workoutsCard: Locator
  readonly caloriesCard: Locator
  readonly challengesCard: Locator

  // calendar grid
  readonly calendarContainer: Locator
  readonly weekdayHeaders: Locator
  readonly calendarGrid: Locator
  readonly calendarDays: Locator
  readonly calendarSkeleton: Locator

  // day modal
  readonly dayModal: Locator
  readonly dayModalTitle: Locator
  readonly dayModalCloseButton: Locator
  readonly dayModalTotalsCard: Locator
  readonly dayModalActivitiesCard: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Calendar' })
    this.pageSubtitle = page.locator('text=View your activities by date')

    // navigation - buttons are next to month display
    this.todayButton = page.locator('button:has-text("Today")')
    // the prev/next buttons are icon-only buttons before and after the month display
    // use the flex container containing the navigation
    const navContainer = page.locator('.flex.items-center.gap-1')
    this.prevMonthButton = navContainer.locator('button').first()
    this.nextMonthButton = navContainer.locator('button').last()
    this.monthYearDisplay = page.locator('.w-32.text-center.font-medium')

    // stats cards - use text matching within cards
    // the stats grid has 6 cards with label + value structure
    this.distanceCard = page.locator('[data-slot="card"]:has-text("Distance")').first()
    this.elevationCard = page.locator('[data-slot="card"]:has-text("Elevation")').first()
    this.timeCard = page.locator('[data-slot="card"]').filter({ hasText: 'Time' }).first()
    this.workoutsCard = page.locator('[data-slot="card"]:has-text("Workouts")').first()
    this.caloriesCard = page.locator('[data-slot="card"]:has-text("Calories")').first()
    this.challengesCard = page.locator('[data-slot="card"]:has-text("Challenges")').first()

    // calendar grid
    this.calendarContainer = page.locator('.rounded-lg.border.border-border.bg-card')
    this.weekdayHeaders = page.locator('.grid.grid-cols-7.border-b .text-center')
    this.calendarGrid = page.locator('.grid.grid-cols-7').last()
    this.calendarDays = page.locator('.min-h-\\[100px\\].border-b.border-r')
    this.calendarSkeleton = page.locator('[class*="animate-pulse"]')

    // day modal
    this.dayModal = page.locator('.fixed.inset-0.z-50')
    this.dayModalTitle = this.dayModal.locator('.font-semibold').first()
    this.dayModalCloseButton = this.dayModal.locator('button:has-text("Close")')
    this.dayModalTotalsCard = this.dayModal.locator('[data-slot="card"]').filter({
      hasText: 'Totals',
    })
    this.dayModalActivitiesCard = this.dayModal.locator('[data-slot="card"]').filter({
      hasText: 'Activities',
    })
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/calendar')
  }

  // wait for calendar to be loaded
  async waitForCalendarLoad(): Promise<void> {
    // wait for loading skeleton to disappear
    await this.calendarSkeleton.first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {
      // skeleton may not exist if data loads fast
    })
    // wait for calendar grid to be visible
    await this.calendarGrid.waitFor({ state: 'visible', timeout: 10000 })
  }

  // get current month/year from display
  async getCurrentMonthYear(): Promise<string> {
    return (await this.monthYearDisplay.textContent()) || ''
  }

  // navigate to previous month
  async goToPreviousMonth(): Promise<void> {
    await this.prevMonthButton.click()
    await this.page.waitForTimeout(300)
    await this.waitForCalendarLoad()
  }

  // navigate to next month
  async goToNextMonth(): Promise<void> {
    await this.nextMonthButton.click()
    await this.page.waitForTimeout(300)
    await this.waitForCalendarLoad()
  }

  // click Today button
  async goToToday(): Promise<void> {
    await this.todayButton.click()
    await this.page.waitForTimeout(300)
    await this.waitForCalendarLoad()
  }

  // get weekday header names
  async getWeekdayHeaders(): Promise<string[]> {
    const headers = await this.weekdayHeaders.allTextContents()
    return headers.map((h) => h.trim())
  }

  // get all calendar day cells
  async getDayCells(): Promise<Locator[]> {
    const count = await this.calendarDays.count()
    const cells: Locator[] = []
    for (let i = 0; i < count; i++) {
      cells.push(this.calendarDays.nth(i))
    }
    return cells
  }

  // get days that have activities (have colored activity links)
  async getDaysWithActivities(): Promise<number[]> {
    const cells = await this.getDayCells()
    const daysWithActivities: number[] = []
    for (let i = 0; i < cells.length; i++) {
      const cell = cells[i]
      const activityLinks = cell.locator('a')
      const linkCount = await activityLinks.count()
      if (linkCount > 0) {
        // get the day number from the cell
        const dayText = await cell.locator('.text-right.text-sm').first().textContent()
        if (dayText) {
          const day = parseInt(dayText.trim())
          if (!isNaN(day)) {
            daysWithActivities.push(day)
          }
        }
      }
    }
    return daysWithActivities
  }

  // click on a specific day cell (by day number)
  async clickDay(dayNumber: number): Promise<void> {
    // find the cell with this day number that's in the current month
    const cells = await this.getDayCells()
    for (const cell of cells) {
      const dayText = await cell.locator('.text-right.text-sm').first().textContent()
      if (dayText && parseInt(dayText.trim()) === dayNumber) {
        // check if it's in current month (not muted)
        const isMuted = await cell.evaluate((el) => el.classList.contains('bg-muted/30'))
        if (!isMuted) {
          await cell.click()
          await this.page.waitForTimeout(300)
          return
        }
      }
    }
  }

  // click on a day that has activities
  async clickDayWithActivities(): Promise<boolean> {
    const daysWithActivities = await this.getDaysWithActivities()
    if (daysWithActivities.length > 0) {
      await this.clickDay(daysWithActivities[0])
      return true
    }
    return false
  }

  // check if day modal is open
  async isDayModalOpen(): Promise<boolean> {
    return await this.dayModal.isVisible()
  }

  // close day modal
  async closeDayModal(): Promise<void> {
    if (await this.isDayModalOpen()) {
      await this.dayModalCloseButton.click()
      await this.dayModal.waitFor({ state: 'hidden', timeout: 5000 })
    }
  }

  // get day modal date
  async getDayModalDate(): Promise<string> {
    return (await this.dayModalTitle.textContent()) || ''
  }

  // get activity count in day modal
  async getDayModalActivityCount(): Promise<number> {
    const activitiesCard = this.dayModalActivitiesCard
    const activityItems = activitiesCard.locator('.rounded-md.border.border-border')
    return await activityItems.count()
  }

  // get stats card value
  async getStatsCardValue(cardLocator: Locator): Promise<string> {
    const value = cardLocator.locator('.text-lg.font-semibold')
    return (await value.textContent()) || ''
  }

  // get all stats values
  async getAllStatsValues(): Promise<{
    distance: string
    elevation: string
    time: string
    workouts: string
    calories: string
    challenges: string
  }> {
    return {
      distance: await this.getStatsCardValue(this.distanceCard),
      elevation: await this.getStatsCardValue(this.elevationCard),
      time: await this.getStatsCardValue(this.timeCard),
      workouts: await this.getStatsCardValue(this.workoutsCard),
      calories: await this.getStatsCardValue(this.caloriesCard),
      challenges: await this.getStatsCardValue(this.challengesCard),
    }
  }

  // check if today is highlighted
  async isTodayHighlighted(): Promise<boolean> {
    // today has text-strava and font-bold classes
    const todayCell = this.page.locator('.font-bold.text-strava')
    return await todayCell.isVisible()
  }

  // get activity links from a day cell
  async getActivityLinksFromDay(dayNumber: number): Promise<Locator[]> {
    const cells = await this.getDayCells()
    for (const cell of cells) {
      const dayText = await cell.locator('.text-right.text-sm').first().textContent()
      if (dayText && parseInt(dayText.trim()) === dayNumber) {
        const isMuted = await cell.evaluate((el) => el.classList.contains('bg-muted/30'))
        if (!isMuted) {
          const links = cell.locator('a')
          const count = await links.count()
          const result: Locator[] = []
          for (let i = 0; i < count; i++) {
            result.push(links.nth(i))
          }
          return result
        }
      }
    }
    return []
  }

  // check if "+N more" indicator is shown for a day
  async hasMoreActivitiesIndicator(dayNumber: number): Promise<boolean> {
    const cells = await this.getDayCells()
    for (const cell of cells) {
      const dayText = await cell.locator('.text-right.text-sm').first().textContent()
      if (dayText && parseInt(dayText.trim()) === dayNumber) {
        const moreIndicator = cell.locator('text=/\\+\\d+ more/')
        return await moreIndicator.isVisible()
      }
    }
    return false
  }
}
