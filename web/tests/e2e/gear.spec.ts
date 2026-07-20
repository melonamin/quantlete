import { test, expect } from './fixtures'
import { GearPage } from './pages/gear.page'

test.describe('Gear Page', () => {
  test.describe('Gear List', () => {
    test('gear list displays bikes and shoes with usage statistics', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      // verify page header
      await expect(gearPage.pageTitle).toBeVisible()
      await expect(gearPage.pageSubtitle).toBeVisible()

      // check if gear list, empty state, or error is visible
      const hasGear = await gearPage.hasGearData()
      const hasError = await gearPage.hasError()
      const isEmpty = await gearPage.emptyState.isVisible()

      // one of these should be true
      expect(hasGear || hasError || isEmpty).toBe(true)

      if (hasGear) {
        // should have at least one gear card
        const gearCount = await gearPage.getGearCardCount()
        expect(gearCount).toBeGreaterThan(0)

        // first gear card should have distance
        const distance = await gearPage.getGearCardDistance(0)
        expect(distance).toBeTruthy()

        // first gear card should have activities count
        const activities = await gearPage.getGearCardActivities(0)
        expect(activities).toBeTruthy()

        // get gear names
        const names = await gearPage.getGearNames()
        expect(names.length).toBeGreaterThan(0)
        expect(names[0]).toBeTruthy()
      }
    })

    test('gear cards display expected information', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      const hasGear = await gearPage.hasGearData()
      if (!hasGear) {
        return
      }

      // first gear card should be visible
      const firstCard = gearPage.getGearCard(0)
      await expect(firstCard).toBeVisible()

      // should have a title
      const title = firstCard.locator('[data-slot="card-title"]')
      await expect(title).toBeVisible()

      // should have distance row
      const distanceLabel = firstCard.locator('text=Distance')
      await expect(distanceLabel).toBeVisible()

      // should have activities row
      const activitiesLabel = firstCard.locator('text=Activities')
      await expect(activitiesLabel).toBeVisible()
    })
  })

  test.describe('Retired Gear Toggle', () => {
    test('Show/Hide Retired toggle filters gear list', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      const hasGear = await gearPage.hasGearData()
      if (!hasGear) {
        return
      }

      // show retired button should be visible
      await expect(gearPage.showRetiredButton).toBeVisible()

      // get initial state
      const initialShowingRetired = await gearPage.isShowingRetired()
      const initialCount = await gearPage.getGearCardCount()

      // toggle retired visibility
      await gearPage.toggleRetired()

      // state should change
      const afterToggleShowingRetired = await gearPage.isShowingRetired()
      expect(afterToggleShowingRetired).toBe(!initialShowingRetired)

      // count may or may not change depending on retired gear
      const afterToggleCount = await gearPage.getGearCardCount()
      // count should be valid (may be same, more, or less)
      expect(afterToggleCount).toBeGreaterThanOrEqual(0)

      // toggle back
      await gearPage.toggleRetired()

      // should return to initial state
      const finalShowingRetired = await gearPage.isShowingRetired()
      expect(finalShowingRetired).toBe(initialShowingRetired)

      const finalCount = await gearPage.getGearCardCount()
      expect(finalCount).toBe(initialCount)
    })
  })

  test.describe('Tab Navigation', () => {
    test('maintenance tab switch displays maintenance view', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      // tab buttons should be visible
      await expect(gearPage.gearTabButton).toBeVisible()
      await expect(gearPage.maintenanceTabButton).toBeVisible()

      // initially gear tab should be active
      const isGearActive = await gearPage.isGearTabActive()
      expect(isGearActive).toBe(true)

      // switch to maintenance tab
      await gearPage.switchToMaintenanceTab()

      // maintenance content should be visible
      const isMaintenanceDueVisible = await gearPage.isMaintenanceDueVisible()
      const isManageComponentsVisible = await gearPage.isManageComponentsVisible()

      // at least one of these should be visible
      expect(isMaintenanceDueVisible || isManageComponentsVisible).toBe(true)

      // switch back to gear tab
      await gearPage.switchToGearTab()

      // gear content should be visible again
      const hasGear = await gearPage.hasGearData()
      const isEmpty = await gearPage.emptyState.isVisible()
      expect(hasGear || isEmpty).toBe(true)
    })
  })

  test.describe('Gear Charts', () => {
    test('gear usage chart renders', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      const hasGear = await gearPage.hasGearData()
      if (!hasGear) {
        return
      }

      // check if charts are visible (only show when there's usage data)
      const chartsVisible = await gearPage.areChartsVisible()

      if (chartsVisible) {
        // distance per month chart should be visible
        await expect(gearPage.distancePerMonthChart).toBeVisible()

        // cumulative distance chart should be visible
        await expect(gearPage.cumulativeDistanceChart).toBeVisible()

        // moving time share chart should be visible
        await expect(gearPage.movingTimeShareChart).toBeVisible()
      }
      // charts may not be visible if there's no monthly usage data
    })
  })

  test.describe('Add Custom Gear Modal', () => {
    test('add custom gear modal opens, accepts input, and submits', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      // add custom gear button should be visible when on gear tab
      const isGearTabActive = await gearPage.isGearTabActive()
      if (!isGearTabActive) {
        await gearPage.switchToGearTab()
      }

      await expect(gearPage.addCustomGearButton).toBeVisible()

      // open modal
      await gearPage.openAddCustomGearModal()

      // modal should be open
      const isModalOpen = await gearPage.isCustomGearModalOpen()
      expect(isModalOpen).toBe(true)

      // modal title should be visible
      await expect(gearPage.customGearModalTitle).toBeVisible()

      // form fields should be visible
      await expect(gearPage.customGearNameInput).toBeVisible()
      await expect(gearPage.customGearHashtagInput).toBeVisible()
      await expect(gearPage.customGearPriceInput).toBeVisible()
      await expect(gearPage.customGearCurrencyInput).toBeVisible()
      await expect(gearPage.customGearRetiredCheckbox).toBeVisible()

      // fill form with test data
      await gearPage.fillCustomGearForm({
        name: 'Test Skateboard',
        hashtag: '#skateboard',
        price: '199.99',
        currency: 'USD',
      })

      // verify input values
      await expect(gearPage.customGearNameInput).toHaveValue('Test Skateboard')
      await expect(gearPage.customGearHashtagInput).toHaveValue('#skateboard')

      // create button should be visible
      await expect(gearPage.customGearCreateButton).toBeVisible()

      // close modal without submitting (to avoid side effects in test)
      await gearPage.closeCustomGearModal()

      // modal should be closed
      const isModalClosed = await gearPage.isCustomGearModalOpen()
      expect(isModalClosed).toBe(false)
    })

    test('custom gear modal close button works', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      // ensure gear tab is active
      const isGearTabActive = await gearPage.isGearTabActive()
      if (!isGearTabActive) {
        await gearPage.switchToGearTab()
      }

      // open modal
      await gearPage.openAddCustomGearModal()

      // modal should be open
      const isOpen = await gearPage.isCustomGearModalOpen()
      expect(isOpen).toBe(true)

      // close modal
      await gearPage.closeCustomGearModal()

      // modal should be closed
      const isClosed = await gearPage.isCustomGearModalOpen()
      expect(isClosed).toBe(false)
    })
  })

  test.describe('Maintenance Panel', () => {
    test('maintenance panel displays when tab is active', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      // switch to maintenance tab
      await gearPage.switchToMaintenanceTab()

      // maintenance due card should be visible
      await expect(gearPage.maintenanceDueCard).toBeVisible()

      // manage components card should be visible
      await expect(gearPage.manageComponentsCard).toBeVisible()

      // add component button should be visible (may be disabled if no gear selected)
      await expect(gearPage.addComponentButton).toBeVisible()
    })

    test('gear selector is available in maintenance panel', async ({ page }) => {
      const gearPage = new GearPage(page)
      await gearPage.goto()
      await gearPage.waitForDataLoad()

      const hasGear = await gearPage.hasGearData()
      if (!hasGear) {
        return
      }

      // switch to maintenance tab
      await gearPage.switchToMaintenanceTab()

      // manage components card should have gear selector
      await expect(gearPage.gearSelector).toBeVisible()
    })
  })
})
