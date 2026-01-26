import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

// page object for the settings page
// provides selectors and helpers for interacting with connection, sync, and display settings
export class SettingsPage extends BasePage {
  // page header
  readonly pageTitle: Locator
  readonly pageSubtitle: Locator

  // connection section
  readonly connectionSection: Locator
  readonly stravaConnectionCard: Locator
  readonly connectionStatus: Locator
  readonly athleteName: Locator
  readonly connectButton: Locator

  // data sync section
  readonly dataSyncSection: Locator
  readonly importCard: Locator
  readonly startImportButton: Locator
  readonly resumeImportButton: Locator
  readonly cancelImportButton: Locator
  readonly advancedOptionsButton: Locator
  readonly advancedOptionsPanel: Locator

  // import options checkboxes
  readonly includeStreamsCheckbox: Locator
  readonly includeSegmentsCheckbox: Locator
  readonly includeBestEffortsCheckbox: Locator
  readonly includePhotosCheckbox: Locator

  // last sync card
  readonly lastSyncCard: Locator
  readonly historyButton: Locator

  // sync history modal
  readonly syncHistoryModal: Locator
  readonly syncHistoryCloseButton: Locator
  readonly syncHistoryTable: Locator

  // display section
  readonly displaySection: Locator
  readonly preferencesCard: Locator

  // unit system toggle
  readonly unitSystemLabel: Locator
  readonly metricButton: Locator
  readonly imperialButton: Locator

  // theme toggle
  readonly themeLabel: Locator
  readonly systemThemeButton: Locator
  readonly lightThemeButton: Locator
  readonly darkThemeButton: Locator

  // loading and error states
  readonly loadingSkeletons: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)

    // page header
    this.pageTitle = page.locator('h1').filter({ hasText: 'Settings' })
    this.pageSubtitle = page.locator('text=Configure your app preferences')

    // connection section
    this.connectionSection = page.locator('section').filter({
      has: page.locator('text=Connection'),
    })
    this.stravaConnectionCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Strava Connection'),
    })
    this.connectionStatus = this.stravaConnectionCard.locator('p.font-medium').first()
    this.athleteName = this.stravaConnectionCard.locator('.text-sm.text-muted-foreground').first()
    this.connectButton = page.locator('a:has-text("Connect Strava"), button:has-text("Connect Strava")')

    // data sync section
    this.dataSyncSection = page.locator('section').filter({
      has: page.locator('text=Data Sync'),
    })
    this.importCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Import'),
    }).first()
    this.startImportButton = page.locator('button:has-text("Start Import"), button:has-text("Start New Import")')
    this.resumeImportButton = page.locator('button:has-text("Resume Previous Sync")')
    this.cancelImportButton = page.locator('button:has-text("Cancel")')
    this.advancedOptionsButton = page.locator('button:has-text("Advanced options")')
    this.advancedOptionsPanel = page.locator('#import-advanced-options')

    // import options checkboxes
    this.includeStreamsCheckbox = page.locator('#include-streams')
    this.includeSegmentsCheckbox = page.locator('#include-segments')
    this.includeBestEffortsCheckbox = page.locator('#include-best-efforts')
    this.includePhotosCheckbox = page.locator('#include-photos')

    // last sync card
    this.lastSyncCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Last Sync'),
    })
    this.historyButton = page.locator('button:has-text("History")')

    // sync history modal
    this.syncHistoryModal = page.locator('[role="dialog"]').filter({
      has: page.locator('text=Sync History'),
    })
    this.syncHistoryCloseButton = this.syncHistoryModal.locator('button[aria-label="Close"], button:has-text("Close")').first()
    this.syncHistoryTable = this.syncHistoryModal.locator('table, .space-y-2')

    // display section
    this.displaySection = page.locator('section').filter({
      has: page.locator('text=Display'),
    })
    this.preferencesCard = page.locator('[class*="card"]').filter({
      has: page.locator('text=Preferences'),
    })

    // unit system toggle
    this.unitSystemLabel = page.locator('text=Unit System').locator('..')
    this.metricButton = page.locator('button:has-text("Metric")')
    this.imperialButton = page.locator('button:has-text("Imperial")')

    // theme toggle
    this.themeLabel = page.locator('p.font-medium:has-text("Theme")').locator('..')
    this.systemThemeButton = page.locator('button:has-text("System")')
    this.lightThemeButton = page.locator('button:has-text("Light")')
    this.darkThemeButton = page.locator('button:has-text("Dark")')

    // loading and error states
    this.loadingSkeletons = page.locator('[class*="skeleton"]')
    this.errorMessage = page.locator('text=Failed to load settings')
  }

  getPageIdentifier(): Locator {
    return this.pageTitle
  }

  async goto(): Promise<void> {
    await super.goto('/settings')
  }

  // wait for page to load
  async waitForDataLoad(): Promise<void> {
    await this.waitForLoadingComplete()
    // wait for preferences card (always visible) or error
    await Promise.race([
      this.preferencesCard.waitFor({ state: 'visible', timeout: 10000 }),
      this.errorMessage.waitFor({ state: 'visible', timeout: 10000 }),
    ]).catch(() => {
      // page might load without either if there's a different state
    })
  }

  // check if connected to Strava
  async isConnected(): Promise<boolean> {
    const statusText = await this.connectionStatus.textContent()
    return statusText?.toLowerCase().includes('connected') ?? false
  }

  // get athlete name if connected
  async getAthleteName(): Promise<string | null> {
    if (!(await this.isConnected())) return null
    return await this.athleteName.textContent()
  }

  // check if unit system button is selected
  async isUnitSystemSelected(unit: 'metric' | 'imperial'): Promise<boolean> {
    const button = unit === 'metric' ? this.metricButton : this.imperialButton
    const dataState = await button.getAttribute('data-state')
    const classes = await button.getAttribute('class')
    // check for various selected indicators
    return (
      dataState === 'on' ||
      dataState === 'active' ||
      classes?.includes('bg-primary') ||
      classes?.includes('text-primary-foreground') ||
      false
    )
  }

  // select unit system
  async selectUnitSystem(unit: 'metric' | 'imperial'): Promise<void> {
    const button = unit === 'metric' ? this.metricButton : this.imperialButton
    await button.click()
    await this.page.waitForTimeout(300)
  }

  // get selected unit system
  async getSelectedUnitSystem(): Promise<'metric' | 'imperial' | null> {
    if (await this.isUnitSystemSelected('metric')) return 'metric'
    if (await this.isUnitSystemSelected('imperial')) return 'imperial'
    return null
  }

  // check if theme button is selected
  async isThemeSelected(theme: 'system' | 'light' | 'dark'): Promise<boolean> {
    const button =
      theme === 'system'
        ? this.systemThemeButton
        : theme === 'light'
          ? this.lightThemeButton
          : this.darkThemeButton
    const dataState = await button.getAttribute('data-state')
    const classes = await button.getAttribute('class')
    return (
      dataState === 'on' ||
      dataState === 'active' ||
      classes?.includes('bg-primary') ||
      classes?.includes('text-primary-foreground') ||
      false
    )
  }

  // select theme
  async selectTheme(theme: 'system' | 'light' | 'dark'): Promise<void> {
    const button =
      theme === 'system'
        ? this.systemThemeButton
        : theme === 'light'
          ? this.lightThemeButton
          : this.darkThemeButton
    await button.click()
    await this.page.waitForTimeout(300)
  }

  // get selected theme
  async getSelectedTheme(): Promise<'system' | 'light' | 'dark' | null> {
    if (await this.isThemeSelected('system')) return 'system'
    if (await this.isThemeSelected('light')) return 'light'
    if (await this.isThemeSelected('dark')) return 'dark'
    return null
  }

  // check current theme (from document class)
  async getCurrentDocumentTheme(): Promise<'light' | 'dark'> {
    const isDark = await this.page.evaluate(() => {
      return document.documentElement.classList.contains('dark')
    })
    return isDark ? 'dark' : 'light'
  }

  // toggle advanced options
  async toggleAdvancedOptions(): Promise<void> {
    await this.advancedOptionsButton.click()
    await this.page.waitForTimeout(200)
  }

  // check if advanced options are visible
  async areAdvancedOptionsVisible(): Promise<boolean> {
    return await this.advancedOptionsPanel.isVisible()
  }

  // check if import checkbox is checked
  async isImportOptionChecked(option: 'streams' | 'segments' | 'bestEfforts' | 'photos'): Promise<boolean> {
    const checkbox =
      option === 'streams'
        ? this.includeStreamsCheckbox
        : option === 'segments'
          ? this.includeSegmentsCheckbox
          : option === 'bestEfforts'
            ? this.includeBestEffortsCheckbox
            : this.includePhotosCheckbox
    const dataState = await checkbox.getAttribute('data-state')
    const ariaChecked = await checkbox.getAttribute('aria-checked')
    return dataState === 'checked' || ariaChecked === 'true'
  }

  // toggle import checkbox
  async toggleImportOption(option: 'streams' | 'segments' | 'bestEfforts' | 'photos'): Promise<void> {
    const checkbox =
      option === 'streams'
        ? this.includeStreamsCheckbox
        : option === 'segments'
          ? this.includeSegmentsCheckbox
          : option === 'bestEfforts'
            ? this.includeBestEffortsCheckbox
            : this.includePhotosCheckbox
    await checkbox.click()
    await this.page.waitForTimeout(100)
  }

  // open sync history modal
  async openSyncHistory(): Promise<void> {
    await this.historyButton.click()
    await this.syncHistoryModal.waitFor({ state: 'visible', timeout: 5000 })
  }

  // close sync history modal
  async closeSyncHistory(): Promise<void> {
    // try to click close button, or press Escape
    try {
      await this.syncHistoryCloseButton.click({ timeout: 2000 })
    } catch {
      await this.page.keyboard.press('Escape')
    }
    await this.syncHistoryModal.waitFor({ state: 'hidden', timeout: 5000 })
  }

  // check if sync history modal is open
  async isSyncHistoryOpen(): Promise<boolean> {
    return await this.syncHistoryModal.isVisible()
  }

  // check if last sync card shows data
  async hasLastSyncData(): Promise<boolean> {
    try {
      await this.lastSyncCard.waitFor({ state: 'visible', timeout: 3000 })
      // check for status indicator (Completed, Failed, etc.)
      const statusText = await this.lastSyncCard.locator('.font-medium.capitalize').textContent()
      return !!statusText
    } catch {
      return false
    }
  }

  // get last sync status
  async getLastSyncStatus(): Promise<string | null> {
    try {
      const statusText = await this.lastSyncCard.locator('.font-medium.capitalize').textContent()
      return statusText
    } catch {
      return null
    }
  }
}
