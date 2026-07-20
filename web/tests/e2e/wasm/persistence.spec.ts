import { DashboardPage } from '../pages/dashboard.page'
import { expect, test } from './fixtures'

test('dashboard configuration survives a reload', async ({ page, waitForWasmAppReady }) => {
  // TODO(Task 5): Go-side writes never reach OPFS — this documents the data-loss bug
  test.fixme()

  const dashboard = new DashboardPage(page)
  const widgetTitle = 'Recent Activities'
  const widgetCard = page.locator('[data-slot="card"]').filter({
    has: page.getByText(widgetTitle, { exact: true }),
  })

  await expect(widgetCard).toBeVisible()
  await dashboard.enterEditMode()
  await dashboard.openWidgetPanel()
  await dashboard.toggleWidgetVisibility(widgetTitle)
  await dashboard.closeWidgetPanel()
  await dashboard.exitEditMode()

  await expect(widgetCard).toHaveCount(0)

  // Dashboard saves are debounced by 650 ms. Wait for that Go WASM mutation
  // before reloading so this tests persistence, not the UI debounce.
  await page.waitForTimeout(1_000)
  await page.reload()
  await waitForWasmAppReady()

  await expect(widgetCard).toHaveCount(0)
})
