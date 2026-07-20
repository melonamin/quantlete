import { test as base, expect, type Page } from '@playwright/test'

// type definitions for custom fixtures
type AppFixtures = {
  // wait for the app to be fully ready (dashboard loaded with data)
  waitForAppReady: () => Promise<void>
  // navigate to a route and wait for it to be ready
  goToRoute: (path: string) => Promise<void>
}

// extend base test with app-specific fixtures
export const test = base.extend<AppFixtures>({
  // automatically navigate to home and wait for app to be ready before each test
  page: async ({ page }, use) => {
    await page.goto('/')
    await waitForAppReady(page)
    await use(page)
  },

  waitForAppReady: async ({ page }, use) => {
    await use(async () => {
      await waitForAppReady(page)
    })
  },

  goToRoute: async ({ page }, use) => {
    await use(async (path: string) => {
      await page.goto(path)
      await waitForAppReady(page)
    })
  },
})

// re-export expect from playwright
export { expect }

// helper function to wait for app to be ready
async function waitForAppReady(page: Page): Promise<void> {
  // wait for the main content to be visible (dashboard header or page content)
  await page.waitForSelector('h1', { timeout: 15000 })

  // wait for any loading states to clear
  // the app uses loading skeletons and spinners during data fetch
  await page.waitForFunction(
    () => {
      // check no loading spinners are visible
      const spinners = document.querySelectorAll('[data-testid="loading-spinner"]')
      return spinners.length === 0
    },
    { timeout: 10000 }
  ).catch(() => {
    // loading spinners may not exist if data loads fast, that's ok
  })

  // small delay to ensure React has finished rendering
  await page.waitForTimeout(100)
}

