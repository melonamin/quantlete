import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the athlete page
// provides selectors and helpers for interacting with athlete info, FTP, weight, and HR zones
export class AthletePage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // athlete info card
  readonly athleteInfoCard: Locator
  readonly athleteAvatar: Locator
  readonly athleteName: Locator
  readonly athleteUsername: Locator
  readonly athleteLocation: Locator

  // not authenticated state
  readonly notAuthenticatedCard: Locator
  readonly notAuthenticatedMessage: Locator

  // FTP card
  readonly ftpCard: Locator
  readonly ftpCardTitle: Locator
  readonly ftpHistoryHeader: Locator

  // FTP cycling form
  readonly ftpCyclingDateInput: Locator
  readonly ftpCyclingValueInput: Locator
  readonly ftpCyclingAddButton: Locator
  readonly ftpCyclingChart: Locator

  // FTP running form
  readonly ftpRunningDateInput: Locator
  readonly ftpRunningValueInput: Locator
  readonly ftpRunningAddButton: Locator
  readonly ftpRunningChart: Locator

  // Weight card
  readonly weightCard: Locator
  readonly weightCardTitle: Locator
  readonly weightHistoryHeader: Locator
  readonly weightDateInput: Locator
  readonly weightValueInput: Locator
  readonly weightAddButton: Locator
  readonly weightChart: Locator

  // HR zones card
  readonly hrZonesCard: Locator
  readonly hrZonesCardTitle: Locator
  readonly hrZonesHeader: Locator
  readonly hrZonesSportSelect: Locator
  readonly hrZonesEffectiveFromInput: Locator
  readonly hrZonesMethodSelect: Locator
  readonly hrZonesHrMaxInput: Locator
  readonly hrZonesBoundsInputs: Locator
  readonly hrZonesSaveButton: Locator
  readonly hrZonesExistingDefs: Locator
  readonly hrZonesEditButtons: Locator
  readonly hrZonesDeleteButtons: Locator

  // saving indicators
  readonly savingIndicator: Locator

  // loading and error states
  readonly loadingSkeletons: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Athlete Profile' })
    this.pageSubtitle = page.locator('text=Your training metrics and zones')

    // athlete info card (first card without a CardTitle, just has avatar and name)
    this.athleteInfoCard = page.locator('[class*="card"]').first()
    this.athleteAvatar = this.athleteInfoCard.locator('img, .rounded-full').first()
    this.athleteName = this.athleteInfoCard.locator('.text-xl.font-semibold')
    this.athleteUsername = this.athleteInfoCard.locator('p:has-text("@")')
    this.athleteLocation = this.athleteInfoCard.locator('p').filter({
      has: page.locator('svg'),
    })

    // not authenticated state
    this.notAuthenticatedCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Connect your Strava account'),
    })
    this.notAuthenticatedMessage = page.locator('text=Connect your Strava account in Settings')

    // FTP card
    this.ftpCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Functional Threshold Power'),
    })
    this.ftpCardTitle = this.ftpCard.locator('[class*="CardTitle"]')
    this.ftpHistoryHeader = this.ftpCard.locator('h3:has-text("FTP History")')

    // FTP cycling form (first set of inputs in FTP card)
    const ftpContent = this.ftpCard.locator('[class*="CardContent"]')
    this.ftpCyclingDateInput = ftpContent.locator('input[type="date"]').first()
    this.ftpCyclingValueInput = ftpContent.locator('input[placeholder*="Cycling FTP"]')
    this.ftpCyclingAddButton = ftpContent.locator('button:has-text("Add")').first()
    this.ftpCyclingChart = ftpContent.locator('canvas, svg').first()

    // FTP running form (second set of inputs in FTP card)
    this.ftpRunningDateInput = ftpContent.locator('input[type="date"]').nth(1)
    this.ftpRunningValueInput = ftpContent.locator('input[placeholder*="Running FTP"]')
    this.ftpRunningAddButton = ftpContent.locator('button:has-text("Add")').nth(1)
    this.ftpRunningChart = ftpContent.locator('canvas, svg').nth(1)

    // Weight card
    this.weightCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Weight History').first(),
    })
    this.weightCardTitle = this.weightCard.locator('[class*="CardTitle"]')
    this.weightHistoryHeader = this.weightCard.locator('h3:has-text("Weight History")')
    const weightContent = this.weightCard.locator('[class*="CardContent"]')
    this.weightDateInput = weightContent.locator('input[type="date"]')
    this.weightValueInput = weightContent.locator('input[placeholder*="Weight"]')
    this.weightAddButton = weightContent.locator('button:has-text("Add")')
    this.weightChart = weightContent.locator('canvas, svg').first()

    // HR zones card
    this.hrZonesCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Heart Rate Zones').first(),
    })
    this.hrZonesCardTitle = this.hrZonesCard.locator('[class*="CardTitle"]')
    this.hrZonesHeader = this.hrZonesCard.locator('h3:has-text("Heart Rate Zones")')
    const hrZonesContent = this.hrZonesCard.locator('[class*="CardContent"]')
    this.hrZonesSportSelect = hrZonesContent.locator('[role="combobox"]').first()
    this.hrZonesEffectiveFromInput = hrZonesContent.locator('input[type="date"]')
    this.hrZonesMethodSelect = hrZonesContent.locator('[role="combobox"]').nth(1)
    this.hrZonesHrMaxInput = hrZonesContent.locator('input[inputmode="numeric"]')
    this.hrZonesBoundsInputs = hrZonesContent.locator('input[inputmode="decimal"]')
    this.hrZonesSaveButton = hrZonesContent.locator('button:has-text("Save Definition")')
    this.hrZonesExistingDefs = hrZonesContent.locator('.rounded-md.border')
    this.hrZonesEditButtons = this.hrZonesExistingDefs.locator('button:has-text("Edit")')
    this.hrZonesDeleteButtons = this.hrZonesExistingDefs.locator('button:has-text("Delete")')

    // saving indicator
    this.savingIndicator = page.locator('text=Saving…')

    // loading and error states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.errorMessage = page.locator('text=Failed to load')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/athlete')
  }

  // wait for data to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for either athlete info card, not authenticated message, or error
    await Promise.race([
      this.athleteInfoCard.waitFor({ state: 'visible', timeout: 10000 }),
      this.notAuthenticatedMessage.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ]).catch(() => {
      // page might have a different structure
    })
  }

  // check if user is authenticated
  async isAuthenticated(): Promise<boolean> {
    try {
      // if not authenticated message is visible, user is not authenticated
      const notAuthVisible = await this.notAuthenticatedMessage.isVisible()
      if (notAuthVisible) return false

      // if athlete name is visible, user is authenticated
      const nameVisible = await this.athleteName.isVisible()
      return nameVisible
    } catch {
      return false
    }
  }

  // get athlete name
  async getAthleteName(): Promise<string | null> {
    try {
      return await this.athleteName.textContent()
    } catch {
      return null
    }
  }

  // get athlete username
  async getAthleteUsername(): Promise<string | null> {
    try {
      return await this.athleteUsername.textContent()
    } catch {
      return null
    }
  }

  // check if athlete avatar is visible
  async hasAthleteAvatar(): Promise<boolean> {
    try {
      const img = this.athleteInfoCard.locator('img').first()
      return await img.isVisible()
    } catch {
      return false
    }
  }

  // check if FTP card is visible
  async isFtpCardVisible(): Promise<boolean> {
    return await this.ftpCard.isVisible()
  }

  // check if weight card is visible
  async isWeightCardVisible(): Promise<boolean> {
    return await this.weightCard.isVisible()
  }

  // check if HR zones card is visible
  async isHrZonesCardVisible(): Promise<boolean> {
    return await this.hrZonesCard.isVisible()
  }

  // fill FTP cycling form
  async fillFtpCyclingForm(date: string, value: string): Promise<void> {
    await this.ftpCyclingDateInput.fill(date)
    await this.ftpCyclingValueInput.fill(value)
  }

  // click add FTP cycling button
  async addFtpCycling(): Promise<void> {
    await this.ftpCyclingAddButton.click()
    await this.page.waitForTimeout(500)
  }

  // fill FTP running form
  async fillFtpRunningForm(date: string, value: string): Promise<void> {
    await this.ftpRunningDateInput.fill(date)
    await this.ftpRunningValueInput.fill(value)
  }

  // click add FTP running button
  async addFtpRunning(): Promise<void> {
    await this.ftpRunningAddButton.click()
    await this.page.waitForTimeout(500)
  }

  // fill weight form
  async fillWeightForm(date: string, value: string): Promise<void> {
    await this.weightDateInput.fill(date)
    await this.weightValueInput.fill(value)
  }

  // click add weight button
  async addWeight(): Promise<void> {
    await this.weightAddButton.click()
    await this.page.waitForTimeout(500)
  }

  // check if FTP cycling chart is visible
  async isFtpCyclingChartVisible(): Promise<boolean> {
    try {
      return await this.ftpCyclingChart.isVisible()
    } catch {
      return false
    }
  }

  // check if weight chart is visible
  async isWeightChartVisible(): Promise<boolean> {
    try {
      return await this.weightChart.isVisible()
    } catch {
      return false
    }
  }

  // get HR zones existing definitions count
  async getHrZonesDefinitionCount(): Promise<number> {
    return await this.hrZonesExistingDefs.count()
  }

  // check if HR zones sport select is visible
  async isHrZonesSportSelectVisible(): Promise<boolean> {
    return await this.hrZonesSportSelect.isVisible()
  }

  // check if HR zones method select is visible
  async isHrZonesMethodSelectVisible(): Promise<boolean> {
    return await this.hrZonesMethodSelect.isVisible()
  }

  // get count of HR zones bounds inputs (should be 5)
  async getHrZonesBoundsInputCount(): Promise<number> {
    return await this.hrZonesBoundsInputs.count()
  }

  // click HR zones save button
  async saveHrZonesDefinition(): Promise<void> {
    await this.hrZonesSaveButton.click()
    await this.page.waitForTimeout(500)
  }

  // click edit on first HR zones definition
  async editFirstHrZonesDefinition(): Promise<void> {
    await this.hrZonesEditButtons.first().click()
    await this.page.waitForTimeout(200)
  }

  // click delete on first HR zones definition
  async deleteFirstHrZonesDefinition(): Promise<void> {
    await this.hrZonesDeleteButtons.first().click()
    await this.page.waitForTimeout(500)
  }

  // check if saving indicator is visible
  async isSaving(): Promise<boolean> {
    return await this.savingIndicator.isVisible()
  }
}
