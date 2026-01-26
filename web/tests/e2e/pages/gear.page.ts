import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the gear page
// provides selectors and helpers for interacting with gear list, charts, and modals
export class GearPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // tab buttons
  readonly gearTabButton: Locator
  readonly maintenanceTabButton: Locator

  // filter controls
  readonly showRetiredButton: Locator
  readonly addCustomGearButton: Locator

  // gear cards grid
  readonly gearGrid: Locator
  readonly gearCards: Locator

  // individual gear card elements
  readonly gearCardTitle: Locator
  readonly gearCardDistance: Locator
  readonly gearCardActivities: Locator
  readonly gearCardEditButton: Locator
  readonly gearCardPriceButton: Locator

  // charts section
  readonly chartsSection: Locator
  readonly distancePerMonthChart: Locator
  readonly cumulativeDistanceChart: Locator
  readonly movingTimeShareChart: Locator

  // custom gear modal
  readonly customGearModal: Locator
  readonly customGearModalTitle: Locator
  readonly customGearNameInput: Locator
  readonly customGearHashtagInput: Locator
  readonly customGearPriceInput: Locator
  readonly customGearCurrencyInput: Locator
  readonly customGearRetiredCheckbox: Locator
  readonly customGearCreateButton: Locator
  readonly customGearSaveButton: Locator
  readonly customGearDeleteButton: Locator
  readonly customGearCloseButton: Locator

  // price edit modal
  readonly priceModal: Locator
  readonly priceInput: Locator
  readonly priceCurrencyInput: Locator
  readonly priceSaveButton: Locator
  readonly priceCancelButton: Locator

  // maintenance panel (when maintenance tab is active)
  readonly maintenanceDueCard: Locator
  readonly manageComponentsCard: Locator
  readonly componentList: Locator
  readonly addComponentButton: Locator
  readonly gearSelector: Locator

  // component modal
  readonly componentModal: Locator
  readonly componentNameInput: Locator
  readonly componentHashtagInput: Locator
  readonly componentCreateButton: Locator
  readonly componentCloseButton: Locator

  // loading and error states
  readonly loadingSkeletons: Locator
  readonly emptyState: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Gear' })
    this.pageSubtitle = page.locator('text=Manage your equipment and track usage')

    // tab buttons
    this.gearTabButton = page.locator('button').filter({ hasText: /^Gear$/ }).first()
    this.maintenanceTabButton = page.locator('button').filter({ hasText: 'Maintenance' }).first()

    // filter controls
    this.showRetiredButton = page.locator(
      'button:has-text("Show Retired"), button:has-text("Hide Retired")'
    )
    this.addCustomGearButton = page.locator('button:has-text("Add Custom Gear")')

    // gear cards grid
    this.gearGrid = page.locator('.grid.gap-4')
    this.gearCards = page.locator('[class*="card"]').filter({
      has: page.locator('text=Distance'),
    })

    // individual gear card elements (within cards)
    this.gearCardTitle = this.gearCards.locator('.text-lg, [class*="CardTitle"]')
    this.gearCardDistance = this.gearCards.locator('text=Distance').locator('..').locator('.font-medium')
    this.gearCardActivities = this.gearCards.locator('text=Activities').locator('..').locator('.font-medium')
    this.gearCardEditButton = this.gearCards.locator('button:has-text("Edit")')
    this.gearCardPriceButton = this.gearCards.locator(
      'button:has-text("Edit Price"), button:has-text("Add Price")'
    )

    // charts section
    this.chartsSection = page.locator('.mb-6.grid')
    this.distancePerMonthChart = page.locator('[class*="card"]').filter({
      has: page.locator('text=Distance per month'),
    })
    this.cumulativeDistanceChart = page.locator('[class*="card"]').filter({
      has: page.locator('text=Cumulative distance'),
    })
    this.movingTimeShareChart = page.locator('[class*="card"]').filter({
      has: page.locator('text=Moving time share'),
    })

    // custom gear modal
    this.customGearModal = page.locator('.fixed.inset-0.z-50').filter({
      has: page.locator('text=Custom Gear'),
    })
    this.customGearModalTitle = this.customGearModal.locator('.font-semibold').first()
    this.customGearNameInput = this.customGearModal.locator('input').first()
    this.customGearHashtagInput = this.customGearModal.locator('input[placeholder*="hashtag"], input[placeholder*="skateboard"]')
    this.customGearPriceInput = this.customGearModal.locator('input[placeholder*="499"]')
    this.customGearCurrencyInput = this.customGearModal.locator('input[placeholder="USD"]')
    this.customGearRetiredCheckbox = this.customGearModal.locator('[role="checkbox"]')
    this.customGearCreateButton = this.customGearModal.locator('button:has-text("Create")')
    this.customGearSaveButton = this.customGearModal.locator('button:has-text("Save")')
    this.customGearDeleteButton = this.customGearModal.locator('button:has-text("Delete")')
    this.customGearCloseButton = this.customGearModal.locator('button:has-text("Close")')

    // price edit modal
    this.priceModal = page.locator('.fixed.inset-0.z-50').filter({
      has: page.locator('text=Edit Price'),
    })
    this.priceInput = this.priceModal.locator('input[type="number"]')
    this.priceCurrencyInput = this.priceModal.locator('input[placeholder="USD"]')
    this.priceSaveButton = this.priceModal.locator('button:has-text("Save")')
    this.priceCancelButton = this.priceModal.locator('button:has-text("Cancel")')

    // maintenance panel
    this.maintenanceDueCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Maintenance due'),
    })
    this.manageComponentsCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Manage components'),
    })
    this.componentList = this.manageComponentsCard.locator('.space-y-2')
    this.addComponentButton = page.locator('button:has-text("Add component")')
    this.gearSelector = this.manageComponentsCard.locator('[role="combobox"]').first()

    // component modal
    this.componentModal = page.locator('.fixed.inset-0.z-50').filter({
      has: page.locator('text=component'),
    })
    this.componentNameInput = this.componentModal.locator('input').first()
    this.componentHashtagInput = this.componentModal.locator('input[placeholder*="chain"]')
    this.componentCreateButton = this.componentModal.locator('button:has-text("Create")')
    this.componentCloseButton = this.componentModal.locator('button:has-text("Close")')

    // loading and error states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.emptyState = page.locator('text=No gear found')
    this.errorMessage = page.locator('text=Failed to load gear')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/gear')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either gear cards, empty state, or error
    await Promise.race([
      this.gearCards.first().waitFor({ state: 'visible', timeout: 10000 }),
      this.emptyState.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ]).catch(() => {
      // page might show maintenance tab first
    })
  }

  // check if page has error
  async hasError(): Promise<boolean> {
    return await this.errorMessage.isVisible()
  }

  // check if gear list has data
  async hasGearData(): Promise<boolean> {
    try {
      await this.gearCards.first().waitFor({ state: 'visible', timeout: 5000 })
      const count = await this.gearCards.count()
      return count > 0
    } catch {
      return false
    }
  }

  // get gear card count
  async getGearCardCount(): Promise<number> {
    return await this.gearCards.count()
  }

  // get gear names from cards
  async getGearNames(): Promise<string[]> {
    const count = await this.gearCards.count()
    const names: string[] = []
    for (let i = 0; i < count; i++) {
      const card = this.gearCards.nth(i)
      const title = await card.locator('.text-lg, [class*="CardTitle"]').textContent()
      if (title) names.push(title.trim())
    }
    return names
  }

  // check if showing retired gear
  async isShowingRetired(): Promise<boolean> {
    const buttonText = await this.showRetiredButton.textContent()
    return buttonText?.includes('Hide Retired') ?? false
  }

  // toggle retired gear visibility
  async toggleRetired(): Promise<void> {
    await this.showRetiredButton.click()
    await this.page.waitForTimeout(500)
    await this.waitForLoadingComplete()
  }

  // switch to gear tab
  async switchToGearTab(): Promise<void> {
    await this.gearTabButton.click()
    await this.page.waitForTimeout(300)
  }

  // switch to maintenance tab
  async switchToMaintenanceTab(): Promise<void> {
    await this.maintenanceTabButton.click()
    await this.page.waitForTimeout(300)
  }

  // check if gear tab is active
  async isGearTabActive(): Promise<boolean> {
    const classes = await this.gearTabButton.getAttribute('class')
    const variant = await this.gearTabButton.getAttribute('data-state')
    return classes?.includes('bg-primary') || variant === 'active' || !classes?.includes('outline') || false
  }

  // check if maintenance tab is active
  async isMaintenanceTabActive(): Promise<boolean> {
    const classes = await this.maintenanceTabButton.getAttribute('class')
    const variant = await this.maintenanceTabButton.getAttribute('data-state')
    return classes?.includes('bg-primary') || variant === 'active' || !classes?.includes('outline') || false
  }

  // check if charts are visible
  async areChartsVisible(): Promise<boolean> {
    try {
      const distanceVisible = await this.distancePerMonthChart.isVisible()
      const cumulativeVisible = await this.cumulativeDistanceChart.isVisible()
      return distanceVisible || cumulativeVisible
    } catch {
      return false
    }
  }

  // open add custom gear modal
  async openAddCustomGearModal(): Promise<void> {
    await this.addCustomGearButton.click()
    await this.customGearModal.waitFor({ state: 'visible', timeout: 5000 })
  }

  // check if custom gear modal is open
  async isCustomGearModalOpen(): Promise<boolean> {
    return await this.customGearModal.isVisible()
  }

  // fill custom gear form
  async fillCustomGearForm(data: {
    name: string
    hashtag?: string
    price?: string
    currency?: string
    retired?: boolean
  }): Promise<void> {
    await this.customGearNameInput.fill(data.name)
    if (data.hashtag) {
      await this.customGearHashtagInput.fill(data.hashtag)
    }
    if (data.price) {
      await this.customGearPriceInput.fill(data.price)
    }
    if (data.currency) {
      await this.customGearCurrencyInput.fill(data.currency)
    }
    if (data.retired) {
      await this.customGearRetiredCheckbox.click()
    }
  }

  // close custom gear modal
  async closeCustomGearModal(): Promise<void> {
    await this.customGearCloseButton.click()
    await this.customGearModal.waitFor({ state: 'hidden', timeout: 5000 })
  }

  // check if maintenance due card is visible
  async isMaintenanceDueVisible(): Promise<boolean> {
    return await this.maintenanceDueCard.isVisible()
  }

  // check if manage components card is visible
  async isManageComponentsVisible(): Promise<boolean> {
    return await this.manageComponentsCard.isVisible()
  }

  // get a specific gear card by index
  getGearCard(index: number): Locator {
    return this.gearCards.nth(index)
  }

  // check if gear card has bike icon
  async gearCardHasBikeIcon(index: number): Promise<boolean> {
    const card = this.gearCards.nth(index)
    const bikeIcon = card.locator('svg.text-sport-ride, svg').filter({
      has: this.page.locator('[class*="sport-ride"]'),
    })
    return (await bikeIcon.count()) > 0
  }

  // check if gear card has shoes icon
  async gearCardHasShoesIcon(index: number): Promise<boolean> {
    const card = this.gearCards.nth(index)
    const shoesIcon = card.locator('svg.text-sport-run, svg').filter({
      has: this.page.locator('[class*="sport-run"]'),
    })
    return (await shoesIcon.count()) > 0
  }

  // get distance text from a gear card
  async getGearCardDistance(index: number): Promise<string | null> {
    const card = this.gearCards.nth(index)
    const distanceRow = card.locator('text=Distance').locator('..')
    const value = distanceRow.locator('.font-medium')
    return await value.textContent()
  }

  // get activities count from a gear card
  async getGearCardActivities(index: number): Promise<string | null> {
    const card = this.gearCards.nth(index)
    const activitiesRow = card.locator('text=Activities').locator('..')
    const value = activitiesRow.locator('.font-medium')
    return await value.textContent()
  }
}
