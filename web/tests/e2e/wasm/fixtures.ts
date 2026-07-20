import { test as base, expect, type Page } from '@playwright/test'

type WasmFixtures = {
  waitForWasmAppReady: () => Promise<void>
}

export const test = base.extend<WasmFixtures>({
  page: async ({ page }, use) => {
    // Establish the localhost origin without booting the app, then remove any
    // OPFS data left by a prior run before sql.js initializes.
    await page.goto('/manifest.json')
    await clearOpfs(page)

    await page.goto('/')
    await waitForWasmAppReady(page)
    await use(page)
  },

  waitForWasmAppReady: async ({ page }, use) => {
    await use(async () => {
      await waitForWasmAppReady(page)
    })
  },
})

export { expect }

async function clearOpfs(page: Page): Promise<void> {
  await page.evaluate(async () => {
    const root = await navigator.storage.getDirectory()
    const entryNames: string[] = []

    for await (const [name] of root.entries()) {
      entryNames.push(name)
    }

    for (const name of entryNames) {
      await root.removeEntry(name, { recursive: true })
    }
  })
}

async function waitForWasmAppReady(page: Page): Promise<void> {
  await page.waitForFunction(
    () => typeof (globalThis as unknown as { goStorage?: unknown }).goStorage === 'object',
    undefined,
    { timeout: 30_000 }
  )
  await page.locator('h1:has-text("Dashboard")').waitFor({ state: 'visible', timeout: 30_000 })
  await page.getByText('Loading Demo').waitFor({ state: 'hidden', timeout: 30_000 })
}
