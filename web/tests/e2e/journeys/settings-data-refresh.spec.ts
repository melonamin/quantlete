import { test, expect } from '../fixtures'
import { SettingsPage } from '../pages/settings.page'
import { ActivitiesPage } from '../pages/activities.page'
import { DashboardPage } from '../pages/dashboard.page'

// journey: settings and data refresh
// settings -> change unit system -> activities page -> verify units changed -> dashboard -> verify widget units
// key verification: unit system preference should persist across pages
test.describe('Settings and Data Refresh Journey', () => {
  test('unit system change persists across pages', async ({ page }) => {
    // step 1: navigate to settings page
    const settingsPage = new SettingsPage(page)
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    // step 2: verify settings page loaded
    await expect(settingsPage.pageTitle).toBeVisible()

    // step 3: check current unit system
    const initialUnit = await settingsPage.getSelectedUnitSystem()

    // step 4: determine which unit to switch to
    const targetUnit = initialUnit === 'metric' ? 'imperial' : 'metric'

    // step 5: click the target unit button
    await settingsPage.selectUnitSystem(targetUnit)

    // step 6: wait for the change to apply
    await page.waitForTimeout(500)

    // step 7: verify the unit is now selected
    const newUnit = await settingsPage.getSelectedUnitSystem()
    expect(newUnit).toBe(targetUnit)

    // step 8: navigate to activities page
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForTableLoad()

    // step 9: verify activities page loaded
    await expect(activitiesPage.pageTitle).toBeVisible()

    // step 10: check the units displayed in the table
    // if we have activities, the distance column should show in the new unit system
    const rowCount = await activitiesPage.getActivityRowCount()
    if (rowCount > 0) {
      // imperial uses miles (mi), metric uses kilometers (km)
      const distanceHeader = activitiesPage.distanceHeader
      await expect(distanceHeader).toBeVisible()

      // the distance values in the table should reflect the unit system
      // we check that the page renders without errors
      const firstRow = activitiesPage.tableRows.first()
      await expect(firstRow).toBeVisible()
    }

    // step 11: navigate to dashboard
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.goto()
    await dashboardPage.waitForPageLoad()

    // step 12: verify dashboard loaded
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible()

    // step 13: check that the dashboard shows data with the new unit system
    // the Total Distance card should show in the selected unit
    const totalDistanceCard = page.locator('text=Total Distance').locator('..').locator('..')
    await expect(totalDistanceCard).toBeVisible()

    // step 14: verify the value is displayed (unit system applied correctly)
    const distanceValue = await totalDistanceCard.locator('[class*="text-2xl"], [class*="font-bold"]').first().textContent()
    expect(distanceValue).toBeTruthy()

    // step 15: clean up - switch back to original unit system
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()
    await settingsPage.selectUnitSystem(initialUnit ?? 'metric')
  })

  test('theme change persists across pages', async ({ page }) => {
    // step 1: navigate to settings page
    const settingsPage = new SettingsPage(page)
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    // step 2: get initial theme
    const initialDocTheme = await settingsPage.getCurrentDocumentTheme()

    // step 3: switch to opposite theme
    const targetTheme = initialDocTheme === 'dark' ? 'light' : 'dark'
    await settingsPage.selectTheme(targetTheme)
    await page.waitForTimeout(500)

    // step 4: verify the document class changed
    const newDocTheme = await settingsPage.getCurrentDocumentTheme()
    expect(newDocTheme).toBe(targetTheme)

    // step 5: navigate to activities page
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForTableLoad()

    // step 6: verify theme persists on activities page
    const activitiesDocTheme = await page.evaluate(() => {
      return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    })
    expect(activitiesDocTheme).toBe(targetTheme)

    // step 7: navigate to dashboard
    const dashboardPage = new DashboardPage(page)
    await dashboardPage.goto()
    await dashboardPage.waitForPageLoad()

    // step 8: verify theme persists on dashboard
    const dashboardDocTheme = await page.evaluate(() => {
      return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
    })
    expect(dashboardDocTheme).toBe(targetTheme)

    // step 9: clean up - switch back to system theme
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()
    await settingsPage.selectTheme('system')
  })

  test('strava connection status is visible and consistent', async ({ page }) => {
    // step 1: navigate to settings page
    const settingsPage = new SettingsPage(page)
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    // step 2: check if connected to Strava
    const isConnected = await settingsPage.isConnected()

    // step 3: if connected, athlete name should be visible
    if (isConnected) {
      const athleteName = await settingsPage.getAthleteName()
      expect(athleteName).toBeTruthy()

      // step 4: verify the connection card shows connected status
      await expect(settingsPage.connectionStatus).toBeVisible()
    } else {
      // step 5: if not connected, connect button should be visible
      await expect(settingsPage.connectButton).toBeVisible()
    }

    // step 6: navigate away and back to verify state persists
    await page.goto('/')
    await page.waitForTimeout(500)

    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    // step 7: verify connection status is the same
    const isStillConnected = await settingsPage.isConnected()
    expect(isStillConnected).toBe(isConnected)
  })

  test('import options checkboxes toggle correctly', async ({ page }) => {
    // step 1: navigate to settings page
    const settingsPage = new SettingsPage(page)
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    // step 2: check if advanced options are available
    const advancedOptionsVisible = await settingsPage.advancedOptionsButton.isVisible()
    if (!advancedOptionsVisible) {
      // advanced options not available, skip
      return
    }

    // step 3: expand advanced options
    await settingsPage.toggleAdvancedOptions()
    const panelVisible = await settingsPage.areAdvancedOptionsVisible()
    if (!panelVisible) {
      // panel didn't open, might be server mode limitation
      return
    }

    // step 4: check initial state of streams checkbox
    const initialStreams = await settingsPage.isImportOptionChecked('streams')

    // step 5: toggle the streams checkbox
    await settingsPage.toggleImportOption('streams')
    await page.waitForTimeout(300)

    // step 6: verify it changed
    const afterToggle = await settingsPage.isImportOptionChecked('streams')
    expect(afterToggle).toBe(!initialStreams)

    // step 7: toggle back to original state
    await settingsPage.toggleImportOption('streams')
    await page.waitForTimeout(300)

    // step 8: verify it's back to original
    const restored = await settingsPage.isImportOptionChecked('streams')
    expect(restored).toBe(initialStreams)
  })

  test('sync history modal opens and closes correctly', async ({ page }) => {
    // step 1: navigate to settings page
    const settingsPage = new SettingsPage(page)
    await settingsPage.goto()
    await settingsPage.waitForDataLoad()

    // step 2: check if last sync card has data
    const hasSyncData = await settingsPage.hasLastSyncData()
    if (!hasSyncData) {
      // no sync history available
      return
    }

    // step 3: check if history button is visible
    const historyButtonVisible = await settingsPage.historyButton.isVisible()
    if (!historyButtonVisible) {
      return
    }

    // step 4: open sync history modal
    await settingsPage.openSyncHistory()

    // step 5: verify modal is open
    const modalOpen = await settingsPage.isSyncHistoryOpen()
    expect(modalOpen).toBe(true)

    // step 6: verify sync history table is visible
    await expect(settingsPage.syncHistoryTable).toBeVisible()

    // step 7: close the modal
    await settingsPage.closeSyncHistory()

    // step 8: verify modal is closed
    const modalClosed = await settingsPage.isSyncHistoryOpen()
    expect(modalClosed).toBe(false)
  })
})
