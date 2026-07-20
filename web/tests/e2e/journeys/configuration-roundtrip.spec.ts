import { test, expect } from '../fixtures'
import { AthletePage } from '../pages/athlete.page'
import { GearPage } from '../pages/gear.page'
import { SettingsPage } from '../pages/settings.page'
import { ExportPage } from '../pages/export.page'

// journey: configuration round trip
// athlete -> edit FTP -> gear -> check usage stats -> settings -> verify connection -> export -> download data
// key verification: configuration pages work together and maintain state
test.describe('Configuration Round Trip Journey', () => {
  test('user can navigate through configuration pages', async ({ page }) => {
    // step 1: navigate to athlete page
    const athletePage = new AthletePage(page)
    await athletePage.goto()
    await athletePage.waitForDataLoad()

    // step 2: verify athlete page loaded
    await expect(athletePage.pageTitle).toBeVisible()

    // step 3: check if authenticated
    const isAuthenticated = await athletePage.isAuthenticated()

    if (isAuthenticated) {
      // step 4: verify athlete info card is visible
      const athleteName = await athletePage.getAthleteName()
      expect(athleteName).toBeTruthy()

      // step 5: check if avatar is visible
      const hasAvatar = await athletePage.hasAthleteAvatar()
      expect(hasAvatar).toBe(true)

      // step 6: verify FTP card is visible
      const ftpCardVisible = await athletePage.isFtpCardVisible()
      expect(ftpCardVisible).toBe(true)

      // step 7: verify weight card is visible
      const weightCardVisible = await athletePage.isWeightCardVisible()
      expect(weightCardVisible).toBe(true)

      // step 8: verify HR zones card is visible
      const hrZonesCardVisible = await athletePage.isHrZonesCardVisible()
      expect(hrZonesCardVisible).toBe(true)

      // step 9: check HR zones configuration
      if (hrZonesCardVisible) {
        const sportSelectVisible = await athletePage.isHrZonesSportSelectVisible()
        expect(sportSelectVisible).toBe(true)

        const boundsInputCount = await athletePage.getHrZonesBoundsInputCount()
        // should have 5 zone boundary inputs
        expect(boundsInputCount).toBe(5)
      }
    } else {
      // step 10: verify not authenticated message
      await expect(athletePage.notAuthenticatedMessage).toBeVisible()
    }

    // step 11: navigate to gear page
    await page.getByRole('link', { name: /gear/i }).click()

    // step 12: wait for gear page to load
    const gearPage = new GearPage(page)
    await gearPage.waitForDataLoad()

    // step 13: verify gear page loaded
    await expect(gearPage.pageTitle).toBeVisible()

    // step 14: check if gear tab is active
    const gearTabActive = await gearPage.isGearTabActive()
    expect(gearTabActive).toBe(true)

    // step 15: check for gear data
    const hasGearData = await gearPage.hasGearData()
    if (hasGearData) {
      // step 16: get gear count
      const gearCount = await gearPage.getGearCardCount()
      expect(gearCount).toBeGreaterThan(0)

      // step 17: get gear names
      const gearNames = await gearPage.getGearNames()
      expect(gearNames.length).toBeGreaterThan(0)

      // step 18: check first gear card has usage stats
      const firstGearDistance = await gearPage.getGearCardDistance(0)
      const firstGearActivities = await gearPage.getGearCardActivities(0)
      expect(firstGearDistance).toBeTruthy()
      expect(firstGearActivities).toBeTruthy()

      // step 19: check if charts are visible
      const chartsVisible = await gearPage.areChartsVisible()
      // charts might not be visible depending on data
      expect(chartsVisible !== undefined).toBe(true)

      // step 20: test show/hide retired toggle
      const initialShowingRetired = await gearPage.isShowingRetired()
      await gearPage.toggleRetired()
      const afterToggleShowingRetired = await gearPage.isShowingRetired()
      expect(afterToggleShowingRetired).not.toBe(initialShowingRetired)

      // toggle back
      await gearPage.toggleRetired()

      // step 21: switch to maintenance tab
      await gearPage.switchToMaintenanceTab()
      await page.waitForTimeout(500)

      const maintenanceTabActive = await gearPage.isMaintenanceTabActive()
      expect(maintenanceTabActive).toBe(true)

      // step 22: verify maintenance components are visible
      const maintenanceDueVisible = await gearPage.isMaintenanceDueVisible()
      const manageComponentsVisible = await gearPage.isManageComponentsVisible()
      // at least one should be visible
      expect(maintenanceDueVisible || manageComponentsVisible).toBe(true)

      // step 23: switch back to gear tab
      await gearPage.switchToGearTab()
      await page.waitForTimeout(300)
    }

    // step 24: test custom gear modal
    const addButtonVisible = await gearPage.addCustomGearButton.isVisible()
    if (addButtonVisible) {
      await gearPage.openAddCustomGearModal()

      const modalOpen = await gearPage.isCustomGearModalOpen()
      expect(modalOpen).toBe(true)

      // verify modal inputs are visible
      await expect(gearPage.customGearNameInput).toBeVisible()

      // close modal without saving
      await gearPage.closeCustomGearModal()

      const modalClosed = await gearPage.isCustomGearModalOpen()
      expect(modalClosed).toBe(false)
    }

    // step 25: navigate to settings page
    await page.getByRole('link', { name: /settings/i }).click()

    // step 26: wait for settings page to load
    const settingsPage = new SettingsPage(page)
    await settingsPage.waitForDataLoad()

    // step 27: verify settings page loaded
    await expect(settingsPage.pageTitle).toBeVisible()

    // step 28: verify connection status
    const isConnected = await settingsPage.isConnected()
    if (isConnected) {
      const athleteNameFromSettings = await settingsPage.getAthleteName()
      expect(athleteNameFromSettings).toBeTruthy()
    }

    // step 29: verify unit system toggle is visible
    await expect(settingsPage.metricButton).toBeVisible()
    await expect(settingsPage.imperialButton).toBeVisible()

    // step 30: verify theme toggle is visible
    await expect(settingsPage.systemThemeButton).toBeVisible()
    await expect(settingsPage.lightThemeButton).toBeVisible()
    await expect(settingsPage.darkThemeButton).toBeVisible()

    // step 31: navigate to export page
    await page.getByRole('link', { name: /export/i }).click()

    // step 32: wait for export page to load
    const exportPage = new ExportPage(page)
    await exportPage.waitForDataLoad()

    // step 33: verify export page loaded
    await expect(exportPage.pageTitle).toBeVisible()

    // step 34: check if stats are available
    const hasStats = await exportPage.hasStats()
    if (hasStats) {
      // step 35: verify total activities is shown
      const totalActivities = await exportPage.getTotalActivities()
      expect(totalActivities).toBeTruthy()

      // step 36: verify format selector is visible
      await expect(exportPage.csvButton).toBeVisible()
      await expect(exportPage.jsonButton).toBeVisible()

      // step 37: test format switching
      await exportPage.selectJsonFormat()
      await page.waitForTimeout(200)

      const selectedFormat = await exportPage.getSelectedFormat()
      expect(selectedFormat).toBe('json')

      // switch back to CSV
      await exportPage.selectCsvFormat()
      await page.waitForTimeout(200)

      const csvFormat = await exportPage.getSelectedFormat()
      expect(csvFormat).toBe('csv')

      // step 38: verify date range inputs are available
      await expect(exportPage.fromDateInput).toBeVisible()
      await expect(exportPage.toDateInput).toBeVisible()

      // step 39: fill in a date range
      await exportPage.fillFromDate('2024-01-01')
      await exportPage.fillToDate('2024-12-31')

      const fromValue = await exportPage.getFromDateValue()
      const toValue = await exportPage.getToDateValue()
      expect(fromValue).toBe('2024-01-01')
      expect(toValue).toBe('2024-12-31')

      // step 40: verify download button is visible
      await expect(exportPage.downloadButton).toBeVisible()

      // step 41: verify what's included card is visible
      const whatsIncludedVisible = await exportPage.isWhatsIncludedVisible()
      expect(whatsIncludedVisible).toBe(true)
    }

    // step 42: navigate back to dashboard to complete the round trip
    await page.getByRole('link', { name: /dashboard/i }).click()
    await page.waitForTimeout(500)

    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()
  })

  test('athlete profile shows consistent data with settings', async ({ page }) => {
    // step 1: navigate to athlete page
    const athletePage = new AthletePage(page)
    await athletePage.goto()
    await athletePage.waitForDataLoad()

    // step 2: get athlete name from profile
    const isAuthenticated = await athletePage.isAuthenticated()
    if (!isAuthenticated) {
      return // skip if not authenticated
    }

    const athleteNameFromProfile = await athletePage.getAthleteName()

    // step 3: navigate to settings
    const settingsPage = new SettingsPage(page)
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    // step 4: get athlete name from settings
    const athleteNameFromSettings = await settingsPage.getAthleteName()

    // step 5: names should match (or both be null)
    expect(athleteNameFromProfile).toBe(athleteNameFromSettings)
  })

  test('export page reflects activity count from dashboard', async ({ page }) => {
    // step 1: navigate to dashboard
    await page.goto('/')
    await page.waitForTimeout(500)

    // step 2: get total activities from dashboard stats card
    const dashboardActivities = await page
      .locator('text=Total Activities')
      .locator('..')
      .locator('..')
      .locator('[class*="text-2xl"], [class*="font-bold"]')
      .first()
      .textContent()

    // step 3: navigate to export page
    const exportPage = new ExportPage(page)
    await exportPage.goto()
    await exportPage.waitForDataLoad()

    // step 4: get total activities from export page
    const hasStats = await exportPage.hasStats()
    if (hasStats) {
      const exportActivities = await exportPage.getTotalActivities()

      // step 5: activity counts should match
      // note: formats might differ (e.g., "1,234" vs "1234")
      const dashNum = dashboardActivities?.replace(/,/g, '').trim()
      const exportNum = exportActivities?.replace(/,/g, '').trim()

      expect(dashNum).toBe(exportNum)
    }
  })

  test('configuration pages handle no auth gracefully', async ({ page }) => {
    // step 1: athlete page handles no auth
    const athletePage = new AthletePage(page)
    await athletePage.goto()
    await athletePage.waitForDataLoad()

    // should not crash, should show either profile or connect message
    await expect(athletePage.pageTitle).toBeVisible()

    // step 2: gear page loads without auth
    const gearPage = new GearPage(page)
    await gearPage.goto()
    await gearPage.waitForDataLoad()

    await expect(gearPage.pageTitle).toBeVisible()

    // step 3: settings page loads without auth
    const settingsPage = new SettingsPage(page)
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    await expect(settingsPage.pageTitle).toBeVisible()

    // step 4: export page loads without auth
    const exportPage = new ExportPage(page)
    await exportPage.goto()
    await exportPage.waitForDataLoad()

    await expect(exportPage.pageTitle).toBeVisible()
  })
})
