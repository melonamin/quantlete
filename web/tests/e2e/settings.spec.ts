import { test, expect } from './fixtures'
import { SettingsPage } from './pages/settings.page'

test.describe('Settings Page', () => {
  test.describe('Strava Connection', () => {
    test('Strava connection status displays correctly (connected with athlete name)', async ({
      page,
    }) => {
      const settingsPage = new SettingsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForDataLoad()

      // verify page header
      await expect(settingsPage.pageTitle).toBeVisible()
      await expect(settingsPage.pageSubtitle).toBeVisible()

      // strava connection card should be visible
      await expect(settingsPage.stravaConnectionCard).toBeVisible()

      // check connection status
      const isConnected = await settingsPage.isConnected()

      if (isConnected) {
        // should show "Connected" status
        await expect(settingsPage.connectionStatus).toContainText(/connected/i)

        // should show athlete name
        const athleteName = await settingsPage.getAthleteName()
        expect(athleteName).toBeTruthy()
      } else {
        // should show "Not connected" status
        await expect(settingsPage.connectionStatus).toContainText(/not connected/i)

        // connect button should be visible
        await expect(settingsPage.connectButton).toBeVisible()
      }
    })
  })

  test.describe('Display Preferences', () => {
    test('unit system toggle (Metric/Imperial) changes and persists', async ({ page }) => {
      const settingsPage = new SettingsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForDataLoad()

      // preferences card should be visible
      await expect(settingsPage.preferencesCard).toBeVisible()

      // unit system buttons should be visible
      await expect(settingsPage.metricButton).toBeVisible()
      await expect(settingsPage.imperialButton).toBeVisible()

      // get initial unit system
      const initialUnit = await settingsPage.getSelectedUnitSystem()
      expect(initialUnit).toBeTruthy()

      // toggle to opposite unit
      const newUnit = initialUnit === 'metric' ? 'imperial' : 'metric'
      await settingsPage.selectUnitSystem(newUnit)

      // verify selection changed
      const selectedUnit = await settingsPage.getSelectedUnitSystem()
      expect(selectedUnit).toBe(newUnit)

      // toggle back
      await settingsPage.selectUnitSystem(initialUnit!)

      // verify it changed back
      const finalUnit = await settingsPage.getSelectedUnitSystem()
      expect(finalUnit).toBe(initialUnit)
    })

    test('theme selector (System/Light/Dark) changes theme', async ({ page }) => {
      const settingsPage = new SettingsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForDataLoad()

      // preferences card should be visible
      await expect(settingsPage.preferencesCard).toBeVisible()

      // theme buttons should be visible
      await expect(settingsPage.systemThemeButton).toBeVisible()
      await expect(settingsPage.lightThemeButton).toBeVisible()
      await expect(settingsPage.darkThemeButton).toBeVisible()

      // get initial theme
      const initialTheme = await settingsPage.getSelectedTheme()
      expect(initialTheme).toBeTruthy()

      // select light theme explicitly
      await settingsPage.selectTheme('light')
      await page.waitForTimeout(500) // wait for theme transition

      // verify light theme is selected
      const isLightSelected = await settingsPage.isThemeSelected('light')
      expect(isLightSelected).toBe(true)

      // document should have light theme (no dark class)
      const themeAfterLight = await settingsPage.getCurrentDocumentTheme()
      expect(themeAfterLight).toBe('light')

      // select dark theme
      await settingsPage.selectTheme('dark')
      await page.waitForTimeout(500)

      // verify dark theme is selected
      const isDarkSelected = await settingsPage.isThemeSelected('dark')
      expect(isDarkSelected).toBe(true)

      // document should have dark theme
      const themeAfterDark = await settingsPage.getCurrentDocumentTheme()
      expect(themeAfterDark).toBe('dark')

      // restore initial theme
      if (initialTheme) {
        await settingsPage.selectTheme(initialTheme)
      }
    })
  })

  test.describe('Import Options', () => {
    test('import options checkboxes toggle correctly', async ({ page }) => {
      const settingsPage = new SettingsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForDataLoad()

      // check if connected and import card is visible
      const isConnected = await settingsPage.isConnected()
      if (!isConnected) {
        // skip if not connected - import options require authentication
        return
      }

      // import card should be visible
      await expect(settingsPage.importCard).toBeVisible()

      // advanced options button should be visible
      await expect(settingsPage.advancedOptionsButton).toBeVisible()

      // open advanced options
      await settingsPage.toggleAdvancedOptions()

      // advanced options panel should be visible
      const isVisible = await settingsPage.areAdvancedOptionsVisible()
      expect(isVisible).toBe(true)

      // all checkboxes should be visible
      await expect(settingsPage.includeStreamsCheckbox).toBeVisible()
      await expect(settingsPage.includeSegmentsCheckbox).toBeVisible()
      await expect(settingsPage.includeBestEffortsCheckbox).toBeVisible()
      await expect(settingsPage.includePhotosCheckbox).toBeVisible()

      // by default, all should be checked
      const streamsChecked = await settingsPage.isImportOptionChecked('streams')
      const segmentsChecked = await settingsPage.isImportOptionChecked('segments')
      const bestEffortsChecked = await settingsPage.isImportOptionChecked('bestEfforts')
      const photosChecked = await settingsPage.isImportOptionChecked('photos')

      // all should be checked initially
      expect(streamsChecked).toBe(true)
      expect(segmentsChecked).toBe(true)
      expect(bestEffortsChecked).toBe(true)
      expect(photosChecked).toBe(true)

      // toggle streams off
      await settingsPage.toggleImportOption('streams')
      const streamsAfterToggle = await settingsPage.isImportOptionChecked('streams')
      expect(streamsAfterToggle).toBe(false)

      // toggle streams back on
      await settingsPage.toggleImportOption('streams')
      const streamsRestored = await settingsPage.isImportOptionChecked('streams')
      expect(streamsRestored).toBe(true)

      // close advanced options
      await settingsPage.toggleAdvancedOptions()
      const isClosed = await settingsPage.areAdvancedOptionsVisible()
      expect(isClosed).toBe(false)
    })
  })

  test.describe('Sync History', () => {
    test('sync history modal opens and displays past sync records', async ({ page }) => {
      const settingsPage = new SettingsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForDataLoad()

      // check if connected
      const isConnected = await settingsPage.isConnected()
      if (!isConnected) {
        // sync history requires authentication
        return
      }

      // check if last sync card is visible (indicates there's sync data)
      const hasLastSync = await settingsPage.hasLastSyncData()
      if (!hasLastSync) {
        // no sync data yet, history button may not be visible
        return
      }

      // history button should be visible
      await expect(settingsPage.historyButton).toBeVisible()

      // open sync history modal
      await settingsPage.openSyncHistory()

      // modal should be open
      const isModalOpen = await settingsPage.isSyncHistoryOpen()
      expect(isModalOpen).toBe(true)

      // modal should have content (table or list)
      await expect(settingsPage.syncHistoryTable).toBeVisible()

      // close modal
      await settingsPage.closeSyncHistory()

      // modal should be closed
      const isModalClosed = await settingsPage.isSyncHistoryOpen()
      expect(isModalClosed).toBe(false)
    })

    test('last sync status displays correctly', async ({ page }) => {
      const settingsPage = new SettingsPage(page)
      await settingsPage.goto()
      await settingsPage.waitForDataLoad()

      // check if connected
      const isConnected = await settingsPage.isConnected()
      if (!isConnected) {
        return
      }

      // check if last sync card is visible
      const hasLastSync = await settingsPage.hasLastSyncData()
      if (!hasLastSync) {
        return
      }

      // last sync card should be visible
      await expect(settingsPage.lastSyncCard).toBeVisible()

      // should show a status (Completed, Failed, Running, etc.)
      const status = await settingsPage.getLastSyncStatus()
      expect(status).toBeTruthy()
      // status should be one of the known values
      expect(['completed', 'failed', 'running', 'paused']).toContainEqual(status?.toLowerCase())
    })
  })
})
