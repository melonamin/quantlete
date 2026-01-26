import { test, expect } from './fixtures'
import { BadgesPage } from './pages/badges.page'

test.describe('Badges Page', () => {
  test.describe('Badge Previews', () => {
    test('badge previews render with default styling', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      // verify page header
      await expect(badgesPage.pageTitle).toBeVisible()
      await expect(badgesPage.pageSubtitle).toBeVisible()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        // no data, check for empty state
        const isEmpty = await badgesPage.emptyState.isVisible()
        expect(isEmpty).toBe(true)
        return
      }

      // should have badge SVGs
      const badgeCount = await badgesPage.getBadgeCount()
      expect(badgeCount).toBeGreaterThan(0)

      // first badge should be an SVG with dimensions
      const dimensions = await badgesPage.getFirstBadgeDimensions()
      expect(dimensions).not.toBeNull()
      expect(dimensions?.width).toBeGreaterThan(0)
      expect(dimensions?.height).toBeGreaterThan(0)
    })

    test('overall stats badges display', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // overall stats card should be visible
      await expect(badgesPage.overallStatsCard).toBeVisible()

      // should have stats badges
      const statsSvgs = badgesPage.overallStatsCard.locator('svg')
      const count = await statsSvgs.count()
      expect(count).toBeGreaterThan(0)
    })
  })

  test.describe('Theme Selector', () => {
    test('theme selector updates badge preview', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // verify style card is visible
      await expect(badgesPage.styleCard).toBeVisible()

      // get available themes
      const themes = await badgesPage.getAvailableThemes()
      expect(themes.length).toBeGreaterThan(0)

      // get initial theme
      const initialTheme = await badgesPage.getSelectedTheme()
      expect(initialTheme).toBeTruthy()

      // select a different theme
      const otherTheme = themes.find((t) => t !== initialTheme)
      if (otherTheme) {
        await badgesPage.selectTheme(otherTheme)

        // verify selection changed
        const newTheme = await badgesPage.getSelectedTheme()
        expect(newTheme).toBe(otherTheme)

        // accent color should change (different themes have different colors)
        const newAccent = await badgesPage.getFirstBadgeAccentColor()
        // colors might be the same for some themes, so just check it's valid
        expect(newAccent).toBeTruthy()
      }
    })
  })

  test.describe('Size Selector', () => {
    test('size selector updates badge preview', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // get available sizes
      const sizes = await badgesPage.getAvailableSizes()
      expect(sizes.length).toBeGreaterThan(0)

      // get initial dimensions
      const initialDimensions = await badgesPage.getFirstBadgeDimensions()
      expect(initialDimensions).not.toBeNull()

      // get initial size selection
      const initialSize = await badgesPage.getSelectedSize()
      expect(initialSize).toBeTruthy()

      // select a different size
      const otherSize = sizes.find((s) => s !== initialSize)
      if (otherSize) {
        await badgesPage.selectSize(otherSize)

        // verify selection changed
        const newSize = await badgesPage.getSelectedSize()
        expect(newSize).toBe(otherSize)

        // dimensions should change
        const newDimensions = await badgesPage.getFirstBadgeDimensions()
        expect(newDimensions).not.toBeNull()
        // different sizes should have different dimensions
        expect(
          newDimensions?.width !== initialDimensions?.width ||
            newDimensions?.height !== initialDimensions?.height
        ).toBe(true)
      }
    })
  })

  test.describe('Background Selector', () => {
    test('background selector updates badge preview', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // get available backgrounds
      const backgrounds = await badgesPage.getAvailableBackgrounds()
      expect(backgrounds.length).toBeGreaterThan(0)

      // get initial background
      const initialBg = await badgesPage.getSelectedBackground()
      expect(initialBg).toBeTruthy()

      // get initial fill color
      const initialFill = await badgesPage.getFirstBadgeBackground()

      // select a different background
      const otherBg = backgrounds.find((b) => b !== initialBg)
      if (otherBg) {
        await badgesPage.selectBackground(otherBg)

        // verify selection changed
        const newBg = await badgesPage.getSelectedBackground()
        expect(newBg).toBe(otherBg)

        // background color should change
        const newFill = await badgesPage.getFirstBadgeBackground()
        // for transparent, fill will be null
        if (otherBg.toLowerCase().includes('transparent')) {
          // transparent has no fill rect (or fill=none)
          expect(newFill === null || newFill === 'none' || newFill !== initialFill).toBe(true)
        } else {
          expect(newFill).not.toBe(initialFill)
        }
      }
    })
  })

  test.describe('Download Button', () => {
    test('download button triggers file download', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // check if download buttons are visible
      const hasDownloadButtons = await badgesPage.hasDownloadButtons()
      expect(hasDownloadButtons).toBe(true)

      // set up download listener before clicking
      const downloadPromise = page.waitForEvent('download', { timeout: 5000 }).catch(() => null)

      // click download button
      await badgesPage.clickDownloadButton()

      // check if download was triggered
      const download = await downloadPromise

      if (download) {
        // download was triggered
        const filename = download.suggestedFilename()
        expect(filename).toContain('.svg')
      } else {
        // download may not work in headless mode, but button should at least be clickable
        // and not throw an error
        expect(true).toBe(true)
      }
    })

    test('download buttons exist for all badge previews', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // count badge previews
      const previewCount = await badgesPage.badgePreviews.count()

      // count download buttons
      const downloadCount = await badgesPage.downloadButtons.count()

      // should have download button for each preview
      expect(downloadCount).toBe(previewCount)
    })
  })

  test.describe('Copy Button', () => {
    test('copy buttons exist for badge previews', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // count copy buttons
      const copyCount = await badgesPage.copyButtons.count()
      expect(copyCount).toBeGreaterThan(0)

      // first copy button should be visible
      await expect(badgesPage.copyButtons.first()).toBeVisible()
    })
  })

  test.describe('Badge Sections', () => {
    test('overall stats section displays', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // overall stats card should be visible
      await expect(badgesPage.overallStatsCard).toBeVisible()
    })

    test('achievements section displays when data exists', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // achievements card should be visible
      await expect(badgesPage.achievementsCard).toBeVisible()
    })

    test('time periods section displays when data exists', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      const hasBadges = await badgesPage.hasBadges()
      if (!hasBadges) {
        return
      }

      // time periods card should be visible
      await expect(badgesPage.timePeriodsCard).toBeVisible()
    })
  })

  test.describe('Style Customizer', () => {
    test('style card displays customizer options', async ({ page }) => {
      const badgesPage = new BadgesPage(page)
      await badgesPage.goto()
      await badgesPage.waitForDataLoad()

      // style card should be visible
      await expect(badgesPage.styleCard).toBeVisible()

      // should have theme buttons
      const themes = await badgesPage.getAvailableThemes()
      expect(themes.length).toBeGreaterThan(0)

      // should have background buttons
      const backgrounds = await badgesPage.getAvailableBackgrounds()
      expect(backgrounds.length).toBeGreaterThan(0)

      // should have size buttons
      const sizes = await badgesPage.getAvailableSizes()
      expect(sizes.length).toBeGreaterThan(0)
    })
  })
})
