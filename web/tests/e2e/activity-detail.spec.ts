import { test, expect } from './fixtures'
import { ActivitiesPage } from './pages/activities.page'
import { ActivityDetailPage } from './pages/activity-detail.page'

test.describe('Activity Detail Page', () => {
  // helper to navigate to first activity and return page objects
  async function navigateToFirstActivity(
    page: import('@playwright/test').Page
  ): Promise<{ activitiesPage: ActivitiesPage; activityDetailPage: ActivityDetailPage } | null> {
    const activitiesPage = new ActivitiesPage(page)
    await activitiesPage.goto()
    await activitiesPage.waitForTableLoad()

    const rowCount = await activitiesPage.getActivityRowCount()
    if (rowCount === 0) {
      return null
    }

    await activitiesPage.clickActivityRow(0)
    await page.waitForURL(/\/activities\/\d+/)

    const activityDetailPage = new ActivityDetailPage(page)
    await activityDetailPage.waitForActivityLoad()

    return { activitiesPage, activityDetailPage }
  }

  test.describe('Page Load', () => {
    test('activity detail page loads with correct activity data', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      // verify activity title is visible
      await expect(activityDetailPage.activityTitle).toBeVisible()
      const titleText = await activityDetailPage.getActivityName()
      expect(titleText).toBeTruthy()
      expect(titleText?.length).toBeGreaterThan(0)

      // verify back link is present
      await expect(activityDetailPage.backLink).toBeVisible()

      // verify View on Strava link is present
      await expect(activityDetailPage.stravaLink).toBeVisible()
    })

    test('activity header shows sport type and date', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      // verify sport type info is in the header area
      const headerArea = page.locator('.mb-6').first()
      await expect(headerArea).toBeVisible()

      // the header should contain a sport icon
      const sportIcon = headerArea.locator('.rounded-lg.bg-muted')
      await expect(sportIcon).toBeVisible()
    })
  })

  test.describe('Stats Cards', () => {
    test('stats cards display correctly', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      // basic stats should always be present
      expect(await activityDetailPage.hasStatCard('Distance')).toBe(true)
      expect(await activityDetailPage.hasStatCard('Moving Time')).toBe(true)
      expect(await activityDetailPage.hasStatCard('Elevation')).toBe(true)

      // pace or speed should be present (depends on sport type)
      const hasPace = await activityDetailPage.hasStatCard('Pace')
      const hasSpeed = await activityDetailPage.hasStatCard('Speed')
      expect(hasPace || hasSpeed).toBe(true)
    })

    test('distance card shows value with unit', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      await expect(activityDetailPage.distanceCard).toBeVisible()
      const valueText = await activityDetailPage.getStatValue('Distance')
      expect(valueText).toBeTruthy()
      expect(valueText).toMatch(/[\d.,]+/)
    })

    test('elevation card shows value', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      await expect(activityDetailPage.elevationCard).toBeVisible()
      const valueText = await activityDetailPage.getStatValue('Elevation')
      expect(valueText).toBeTruthy()
    })

    test('duration card shows moving time', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      await expect(activityDetailPage.movingTimeCard).toBeVisible()
      const valueText = await activityDetailPage.getStatValue('Moving Time')
      expect(valueText).toBeTruthy()
    })

    test('heart rate card displays when activity has HR data', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasHr = await activityDetailPage.hasStatCard('Heart Rate')
      if (hasHr) {
        const valueText = await activityDetailPage.getStatValue('Heart Rate')
        expect(valueText).toBeTruthy()
        expect(valueText?.toLowerCase()).toContain('bpm')
      }
      // if no HR data, that's fine - it's optional
    })

    test('power card displays when activity has power data', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasPower = await activityDetailPage.hasStatCard('Power')
      if (hasPower) {
        const valueText = await activityDetailPage.getStatValue('Power')
        expect(valueText).toBeTruthy()
        expect(valueText).toContain('W')
      }
    })
  })

  test.describe('Elevation Profile Chart', () => {
    test('elevation profile chart renders when activity has altitude data', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasElevation = await activityDetailPage.hasElevationProfile()
      if (hasElevation) {
        await expect(activityDetailPage.elevationProfileSection).toBeVisible()
        await expect(activityDetailPage.elevationProfileTitle).toBeVisible()
      }
      // activities without streams won't have this section - that's expected
    })

    test('gradient toggle changes elevation profile display', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasElevation = await activityDetailPage.hasElevationProfile()
      if (hasElevation) {
        const gradientCheckbox = activityDetailPage.elevationProfileSection.locator(
          'button[role="checkbox"]'
        )
        if (await gradientCheckbox.isVisible()) {
          const initialState = await gradientCheckbox.getAttribute('data-state')
          await gradientCheckbox.click()
          await page.waitForTimeout(300)
          const newState = await gradientCheckbox.getAttribute('data-state')
          expect(newState).not.toBe(initialState)
        }
      }
    })
  })

  test.describe('Activity Stream Chart', () => {
    test('activity stream chart renders when activity has stream data', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasStreams = await activityDetailPage.hasActivityStreams()
      if (hasStreams) {
        await expect(activityDetailPage.activityStreamsSection).toBeVisible()
        await expect(activityDetailPage.streamChartTitle).toBeVisible()

        // checkboxes for series should be present
        await expect(activityDetailPage.activityStreamsSection.locator('label:has-text("HR")')).toBeVisible()
        await expect(
          activityDetailPage.activityStreamsSection.locator('label:has-text("Power")')
        ).toBeVisible()
        await expect(
          activityDetailPage.activityStreamsSection.locator('label:has-text("Cadence")')
        ).toBeVisible()
        await expect(
          activityDetailPage.activityStreamsSection.locator('label:has-text("Elev")')
        ).toBeVisible()
      }
    })

    test('stream chart series toggles work', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasStreams = await activityDetailPage.hasActivityStreams()
      if (hasStreams) {
        const checkboxes = activityDetailPage.activityStreamsSection.locator('button[role="checkbox"]')
        const count = await checkboxes.count()

        for (let i = 0; i < count; i++) {
          const checkbox = checkboxes.nth(i)
          const isDisabled = await checkbox.isDisabled()
          if (!isDisabled) {
            const initialState = await checkbox.getAttribute('data-state')
            await checkbox.click()
            await page.waitForTimeout(200)
            const newState = await checkbox.getAttribute('data-state')
            expect(newState).not.toBe(initialState)
            break
          }
        }
      }
    })
  })

  test.describe('GPS Map', () => {
    test('GPS map renders with route polyline when activity has GPS data', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasMap = await activityDetailPage.hasMap()
      if (hasMap) {
        await expect(activityDetailPage.activityMap).toBeVisible()

        // map should have tiles loaded
        await expect(activityDetailPage.activityMap.locator('.leaflet-tile-container')).toBeVisible()

        // map should have a polyline/path for the route
        const svgOverlay = activityDetailPage.activityMap.locator('svg')
        await expect(svgOverlay).toBeVisible()
      }
      // indoor activities or activities without GPS won't have a map - that's expected
    })

    test('map is interactive (can zoom)', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasMap = await activityDetailPage.hasMap()
      if (hasMap) {
        const zoomIn = activityDetailPage.activityMap.locator('.leaflet-control-zoom-in')
        const zoomOut = activityDetailPage.activityMap.locator('.leaflet-control-zoom-out')

        const hasZoomControls = (await zoomIn.isVisible()) || (await zoomOut.isVisible())
        expect(hasZoomControls).toBe(true)
      }
    })
  })

  test.describe('Photos Gallery', () => {
    test('photos gallery displays when activity has photos', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasPhotos = await activityDetailPage.hasPhotos()
      if (hasPhotos) {
        await expect(activityDetailPage.photosSection).toBeVisible()
        await expect(activityDetailPage.photosTitle).toBeVisible()
        await expect(activityDetailPage.viewAllPhotosLink).toBeVisible()

        const photoCount = await activityDetailPage.getPhotoCount()
        expect(photoCount).toBeGreaterThan(0)
      }
      // activities without photos won't have this section - that's expected
    })

    test('photo links open in new tab', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      const hasPhotos = await activityDetailPage.hasPhotos()
      if (hasPhotos) {
        const photoLinks = activityDetailPage.photoGrid.locator('a')
        const linkCount = await photoLinks.count()

        if (linkCount > 0) {
          const firstLink = photoLinks.first()
          await expect(firstLink).toHaveAttribute('target', '_blank')
        }
      }
    })
  })

  test.describe('Back Navigation', () => {
    test('back navigation returns to activities list', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      const { activityDetailPage } = result

      await expect(activityDetailPage.backLink).toBeVisible()
      await activityDetailPage.goBackToActivities()

      await expect(page).toHaveURL(/\/activities$/)
      await expect(page.locator('h1').filter({ hasText: 'Activities' })).toBeVisible()
    })

    test('browser back button returns to activities list', async ({ page }) => {
      const result = await navigateToFirstActivity(page)
      if (!result) return

      await page.goBack()
      await expect(page).toHaveURL(/\/activities/)
    })
  })
})
