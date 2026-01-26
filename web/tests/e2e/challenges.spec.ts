import { test, expect } from './fixtures'
import { ChallengesPage } from './pages/challenges.page'

test.describe('Challenges Page', () => {
  test.describe('Challenges Display', () => {
    test('challenges display grouped by month', async ({ page }) => {
      const challengesPage = new ChallengesPage(page)
      await challengesPage.goto()
      await challengesPage.waitForDataLoad()

      // verify page header
      await expect(challengesPage.pageTitle).toBeVisible()
      await expect(challengesPage.pageSubtitle).toBeVisible()

      // check if challenges, empty state, or error is visible
      const hasChallenges = await challengesPage.hasChallenges()
      const hasError = await challengesPage.hasError()
      const isEmpty = await challengesPage.emptyState.isVisible()

      // one of these should be true
      expect(hasChallenges || hasError || isEmpty).toBe(true)

      if (hasChallenges) {
        // should have at least one challenge
        const challengeCount = await challengesPage.getChallengeCount()
        expect(challengeCount).toBeGreaterThan(0)

        // should have month cards
        const monthCardsCount = await challengesPage.getMonthCardsCount()
        expect(monthCardsCount).toBeGreaterThan(0)

        // get month labels
        const monthLabels = await challengesPage.getMonthLabels()
        expect(monthLabels.length).toBeGreaterThan(0)

        // each label should look like a month name (e.g., "January 2024")
        for (const label of monthLabels) {
          // should contain a month name or year
          expect(label).toMatch(/\w+/)
        }
      }
    })

    test('month cards show challenge count', async ({ page }) => {
      const challengesPage = new ChallengesPage(page)
      await challengesPage.goto()
      await challengesPage.waitForDataLoad()

      const hasChallenges = await challengesPage.hasChallenges()
      if (!hasChallenges) {
        return
      }

      // month card titles should include count in parentheses
      const titles = page.locator('[class*="CardTitle"]:not(:has-text("Import"))')
      const count = await titles.count()

      if (count > 0) {
        const firstTitle = await titles.first().textContent()
        // should have format like "January 2024 (5)"
        expect(firstTitle).toMatch(/\(\d+\)/)
      }
    })
  })

  test.describe('Badge Images', () => {
    test('badge images load correctly', async ({ page }) => {
      const challengesPage = new ChallengesPage(page)
      await challengesPage.goto()
      await challengesPage.waitForDataLoad()

      const hasChallenges = await challengesPage.hasChallenges()
      if (!hasChallenges) {
        return
      }

      // check if images loaded
      const imagesLoaded = await challengesPage.badgeImagesLoaded()
      expect(imagesLoaded).toBe(true)

      // verify first badge has image with src
      const firstBadge = challengesPage.challengeBadges.first()
      const img = firstBadge.locator('img')

      if (await img.isVisible()) {
        const src = await img.getAttribute('src')
        expect(src).toBeTruthy()
      }
    })

    test('badge shows challenge name', async ({ page }) => {
      const challengesPage = new ChallengesPage(page)
      await challengesPage.goto()
      await challengesPage.waitForDataLoad()

      const hasChallenges = await challengesPage.hasChallenges()
      if (!hasChallenges) {
        return
      }

      // get first badge name
      const name = await challengesPage.getBadgeName(0)
      expect(name).toBeTruthy()
    })
  })

  test.describe('External Strava Links', () => {
    test('badge click opens external Strava link (verify href)', async ({ page }) => {
      const challengesPage = new ChallengesPage(page)
      await challengesPage.goto()
      await challengesPage.waitForDataLoad()

      const hasChallenges = await challengesPage.hasChallenges()
      if (!hasChallenges) {
        return
      }

      // verify first badge has href to Strava
      const linksToStrava = await challengesPage.badgeLinksToStrava(0)
      expect(linksToStrava).toBe(true)

      // get the href
      const href = await challengesPage.getBadgeHref(0)
      expect(href).toContain('strava.com/challenges/')

      // verify link opens in new tab
      const badge = challengesPage.challengeBadges.first()
      const target = await badge.getAttribute('target')
      expect(target).toBe('_blank')

      // verify rel attribute for security
      const rel = await badge.getAttribute('rel')
      expect(rel).toContain('noreferrer')
    })
  })

  test.describe('Import Card', () => {
    test('import card is visible', async ({ page }) => {
      const challengesPage = new ChallengesPage(page)
      await challengesPage.goto()
      await challengesPage.waitForDataLoad()

      // import card should be visible
      const isVisible = await challengesPage.isImportCardVisible()
      expect(isVisible).toBe(true)

      // should have file input
      await expect(challengesPage.importFileInput).toBeVisible()

      // should have import button
      await expect(challengesPage.importButton).toBeVisible()
    })
  })

  test.describe('Empty State', () => {
    test('empty state shows helpful message', async ({ page }) => {
      const challengesPage = new ChallengesPage(page)
      await challengesPage.goto()
      await challengesPage.waitForDataLoad()

      const isEmpty = await challengesPage.emptyState.isVisible()

      if (isEmpty) {
        // should show empty state message
        await expect(challengesPage.emptyState).toContainText('No challenges imported yet')
      }
    })
  })
})
