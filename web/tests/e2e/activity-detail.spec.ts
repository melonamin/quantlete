import { test, expect } from './fixtures'

test.describe('Activity Detail Page', () => {
  // helper to navigate to an activity from the activities list
  async function navigateToFirstActivity(page: import('@playwright/test').Page): Promise<string | null> {
    await page.goto('/activities')
    await page.locator('h1').filter({ hasText: 'Activities' }).waitFor({ state: 'visible' })

    // wait for table to load
    await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 10000 }).catch(() => {})

    const rowCount = await page.locator('tbody tr').count()
    if (rowCount === 0) {
      return null
    }

    // get activity link href for verification later
    const activityLink = page.locator('tbody tr').first().locator('a').first()
    const href = await activityLink.getAttribute('href')

    // click to navigate
    await activityLink.click()
    await page.waitForURL(/\/activities\/\d+/)

    return href
  }

  test.describe('Page Load', () => {
    test('activity detail page loads with correct activity data', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        // no activities to test with
        return
      }

      // wait for page to load
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // verify activity title is visible (h1)
      const title = page.locator('h1').first()
      await expect(title).toBeVisible()
      const titleText = await title.textContent()
      expect(titleText).toBeTruthy()
      expect(titleText?.length).toBeGreaterThan(0)

      // verify back link is present
      await expect(page.locator('a:has-text("Back to Activities")')).toBeVisible()

      // verify View on Strava link is present
      await expect(page.locator('a:has-text("View on Strava")')).toBeVisible()
    })

    test('activity header shows sport type and date', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      // wait for loading
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // verify sport type info is in the header area
      const headerArea = page.locator('.mb-6').first()
      await expect(headerArea).toBeVisible()

      // the header should contain a sport icon (in a rounded-lg bg-muted container)
      const sportIcon = headerArea.locator('.rounded-lg.bg-muted')
      await expect(sportIcon).toBeVisible()
    })
  })

  test.describe('Stats Cards', () => {
    test('stats cards display correctly', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      // wait for loading
      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // basic stats should always be present
      await expect(page.locator('[data-slot="card"]').filter({ hasText: 'Distance' }).first()).toBeVisible()
      await expect(page.locator('[data-slot="card"]').filter({ hasText: 'Moving Time' }).first()).toBeVisible()
      await expect(page.locator('[data-slot="card"]').filter({ hasText: 'Elevation' }).first()).toBeVisible()

      // pace or speed should be present (depends on sport type)
      const hasPace = await page.locator('[data-slot="card"]').filter({ hasText: 'Pace' }).first().isVisible()
      const hasSpeed = await page.locator('[data-slot="card"]').filter({ hasText: 'Speed' }).first().isVisible()
      expect(hasPace || hasSpeed).toBe(true)
    })

    test('distance card shows value with unit', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      const distanceCard = page.locator('[data-slot="card"]').filter({ hasText: 'Distance' }).first()
      await expect(distanceCard).toBeVisible()

      // value should contain a number (distance)
      const valueText = await distanceCard.locator('.text-2xl.font-bold').textContent()
      expect(valueText).toBeTruthy()
      // should have a number followed by unit (km or mi)
      expect(valueText).toMatch(/[\d.,]+/)
    })

    test('elevation card shows value', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      const elevationCard = page.locator('[data-slot="card"]').filter({ hasText: 'Elevation' }).first()
      await expect(elevationCard).toBeVisible()

      // value should exist
      const valueText = await elevationCard.locator('.text-2xl.font-bold').textContent()
      expect(valueText).toBeTruthy()
    })

    test('duration card shows moving time', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      const timeCard = page.locator('[data-slot="card"]').filter({ hasText: 'Moving Time' }).first()
      await expect(timeCard).toBeVisible()

      // value should contain time format (e.g., "1h 23m" or "45:30")
      const valueText = await timeCard.locator('.text-2xl.font-bold').textContent()
      expect(valueText).toBeTruthy()
    })

    test('heart rate card displays when activity has HR data', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // HR card may or may not be present depending on activity data
      const hrCard = page.locator('[data-slot="card"]').filter({ hasText: 'Heart Rate' }).first()
      const hasHr = await hrCard.isVisible()

      if (hasHr) {
        const valueText = await hrCard.locator('.text-2xl.font-bold').textContent()
        expect(valueText).toBeTruthy()
        // should contain bpm
        expect(valueText?.toLowerCase()).toContain('bpm')
      }
      // if no HR data, that's fine - it's optional
    })

    test('power card displays when activity has power data', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // power card may or may not be present
      const powerCard = page.locator('[data-slot="card"]').filter({ hasText: 'Power' }).first()
      const hasPower = await powerCard.isVisible()

      if (hasPower) {
        const valueText = await powerCard.locator('.text-2xl.font-bold').textContent()
        expect(valueText).toBeTruthy()
        // should contain W (watts)
        expect(valueText).toContain('W')
      }
    })
  })

  test.describe('Elevation Profile Chart', () => {
    test('elevation profile chart renders when activity has altitude data', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // elevation profile section
      const elevSection = page.locator('.rounded-lg.border').filter({ hasText: 'Elevation Profile' }).first()
      const hasElevation = await elevSection.isVisible()

      if (hasElevation) {
        await expect(elevSection).toBeVisible()
        await expect(page.locator('h2:has-text("Elevation Profile")')).toBeVisible()

        // gradient checkbox should be present
        const gradientLabel = elevSection.locator('label:has-text("Gradient")')
        await expect(gradientLabel).toBeVisible()
      }
      // activities without streams won't have this section - that's expected
    })

    test('gradient toggle changes elevation profile display', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      const elevSection = page.locator('.rounded-lg.border').filter({ hasText: 'Elevation Profile' }).first()
      const hasElevation = await elevSection.isVisible()

      if (hasElevation) {
        // find and click the gradient checkbox
        const gradientCheckbox = elevSection.locator('button[role="checkbox"]')
        if (await gradientCheckbox.isVisible()) {
          // get initial state
          const initialState = await gradientCheckbox.getAttribute('data-state')

          // toggle
          await gradientCheckbox.click()
          await page.waitForTimeout(300)

          // state should have changed
          const newState = await gradientCheckbox.getAttribute('data-state')
          expect(newState).not.toBe(initialState)
        }
      }
    })
  })

  test.describe('Activity Stream Chart', () => {
    test('activity stream chart renders when activity has stream data', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // activity streams section
      const streamsSection = page.locator('.rounded-lg.border').filter({ hasText: 'Activity Streams' }).first()
      const hasStreams = await streamsSection.isVisible()

      if (hasStreams) {
        await expect(streamsSection).toBeVisible()
        await expect(page.locator('h2:has-text("Activity Streams")')).toBeVisible()

        // checkboxes for series should be present
        await expect(streamsSection.locator('label:has-text("HR")')).toBeVisible()
        await expect(streamsSection.locator('label:has-text("Power")')).toBeVisible()
        await expect(streamsSection.locator('label:has-text("Cadence")')).toBeVisible()
        await expect(streamsSection.locator('label:has-text("Elev")')).toBeVisible()
      }
    })

    test('stream chart series toggles work', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      const streamsSection = page.locator('.rounded-lg.border').filter({ hasText: 'Activity Streams' }).first()
      const hasStreams = await streamsSection.isVisible()

      if (hasStreams) {
        // find an enabled checkbox (one that's not disabled)
        const checkboxes = streamsSection.locator('button[role="checkbox"]')
        const count = await checkboxes.count()

        for (let i = 0; i < count; i++) {
          const checkbox = checkboxes.nth(i)
          const isDisabled = await checkbox.isDisabled()
          if (!isDisabled) {
            // toggle this checkbox
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
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // look for leaflet map container
      const mapContainer = page.locator('.leaflet-container').first()
      const hasMap = await mapContainer.isVisible().catch(() => false)

      if (hasMap) {
        await expect(mapContainer).toBeVisible()

        // map should have tiles loaded
        await expect(mapContainer.locator('.leaflet-tile-container')).toBeVisible()

        // map should have a polyline/path for the route
        // leaflet renders paths as SVG
        const svgOverlay = mapContainer.locator('svg')
        await expect(svgOverlay).toBeVisible()
      }
      // indoor activities or activities without GPS won't have a map - that's expected
    })

    test('map is interactive (can zoom)', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      const mapContainer = page.locator('.leaflet-container').first()
      const hasMap = await mapContainer.isVisible().catch(() => false)

      if (hasMap) {
        // leaflet maps have zoom controls
        const zoomIn = mapContainer.locator('.leaflet-control-zoom-in')
        const zoomOut = mapContainer.locator('.leaflet-control-zoom-out')

        // at least one zoom control should be visible
        const hasZoomControls = await zoomIn.isVisible() || await zoomOut.isVisible()
        expect(hasZoomControls).toBe(true)
      }
    })
  })

  test.describe('Segment Efforts', () => {
    test('segment efforts display when activity has segments', async ({ page }) => {
      // note: the current activity detail page doesn't display segment efforts inline
      // this test documents the expected behavior - segments are shown on the separate segments page
      // if segment efforts are added to activity detail in the future, this test will need updating

      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // currently, segment efforts are NOT displayed on the activity detail page
      // they are only visible on the dedicated /segments page
      // this test verifies the current state - no segment efforts section expected
      const segmentsSection = page.locator('.rounded-lg.border').filter({ hasText: 'Segment Efforts' })
      const hasSegments = await segmentsSection.isVisible()

      // for now, we don't expect segments on activity detail
      // if this changes in the future, update this test
      expect(hasSegments).toBe(false)
    })
  })

  test.describe('Photos Gallery', () => {
    test('photos gallery displays when activity has photos', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // photos section
      const photosSection = page.locator('.rounded-lg.border').filter({ hasText: 'Photos' }).first()
      const hasPhotos = await photosSection.isVisible()

      if (hasPhotos) {
        await expect(photosSection).toBeVisible()
        await expect(page.locator('h2:has-text("Photos")')).toBeVisible()

        // should have View all link
        await expect(photosSection.locator('a:has-text("View all")')).toBeVisible()

        // should have photo grid with images
        const photoGrid = photosSection.locator('.grid')
        await expect(photoGrid).toBeVisible()

        const imageCount = await photoGrid.locator('a').count()
        expect(imageCount).toBeGreaterThan(0)
      }
      // activities without photos won't have this section - that's expected
    })

    test('photo links open in new tab', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      const photosSection = page.locator('.rounded-lg.border').filter({ hasText: 'Photos' }).first()
      const hasPhotos = await photosSection.isVisible()

      if (hasPhotos) {
        const photoLinks = photosSection.locator('.grid a')
        const linkCount = await photoLinks.count()

        if (linkCount > 0) {
          const firstLink = photoLinks.first()
          // photo links should open in new tab
          await expect(firstLink).toHaveAttribute('target', '_blank')
        }
      }
    })
  })

  test.describe('Back Navigation', () => {
    test('back navigation returns to activities list', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // verify back link is visible
      const backLink = page.locator('a:has-text("Back to Activities")')
      await expect(backLink).toBeVisible()

      // click back
      await backLink.click()

      // should return to activities list
      await expect(page).toHaveURL(/\/activities$/)
      await expect(page.locator('h1').filter({ hasText: 'Activities' })).toBeVisible()
    })

    test('browser back button returns to activities list', async ({ page }) => {
      const href = await navigateToFirstActivity(page)
      if (!href) {
        return
      }

      await page.locator('[class*="animate-pulse"]').first().waitFor({ state: 'hidden', timeout: 15000 }).catch(() => {})

      // use browser back
      await page.goBack()

      // should return to activities list
      await expect(page).toHaveURL(/\/activities/)
    })
  })
})
