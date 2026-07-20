import { SettingsPage } from '../pages/settings.page'
import { expect, test } from './fixtures'

// Runs against the real WASM build (dist-wasm, empty OPFS database) where
// persistence is live — the demo build intentionally never persists. The
// unit-system toggle is a Go-side settings write that is reachable without
// any imported data, so it exercises the full dirty-mark → persist → OPFS
// path this spec guards.
test('unit system choice survives a reload', async ({ page }) => {
  const settings = new SettingsPage(page)
  const onboardingDialog = page.getByRole('dialog').filter({
    has: page.getByRole('heading', { name: 'Welcome to Quantlete' }),
  })

  await expect(onboardingDialog).toBeVisible()
  await onboardingDialog.getByRole('button', { name: 'Maybe later' }).click()
  await expect(onboardingDialog).toBeHidden()

  await settings.goto()
  await expect(settings.metricButton).toBeVisible()
  expect(await settings.isUnitSystemSelected('metric')).toBe(true)

  await settings.imperialButton.click()
  await expect.poll(() => settings.isUnitSystemSelected('imperial')).toBe(true)

  // The settings write is debounced/asynchronous before it reaches OPFS; give
  // the persist a moment to complete before tearing the page down.
  await page.waitForTimeout(1_000)
  await page.reload()
  // Still on /settings after reload, so wait for the WASM runtime directly
  // rather than the fixture's dashboard-specific readiness check.
  await page.waitForFunction(
    () => typeof (globalThis as unknown as { goStorage?: unknown }).goStorage === 'object',
    undefined,
    { timeout: 30_000 }
  )

  await expect(settings.imperialButton).toBeVisible()
  await expect.poll(() => settings.isUnitSystemSelected('imperial')).toBe(true)
  expect(await settings.isUnitSystemSelected('metric')).toBe(false)
})
