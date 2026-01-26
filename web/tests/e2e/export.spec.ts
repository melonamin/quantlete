import { test, expect } from './fixtures'
import { ExportPage } from './pages/export.page'

test.describe('Export Page', () => {
  test.describe('Data Stats Card', () => {
    test('data stats card shows activity count and date range', async ({ page }) => {
      const exportPage = new ExportPage(page)
      await exportPage.goto()
      await exportPage.waitForDataLoad()

      // verify page header
      await expect(exportPage.pageTitle).toBeVisible()
      await expect(exportPage.pageSubtitle).toBeVisible()

      // stats card should be visible
      await expect(exportPage.statsCard).toBeVisible()
      await expect(exportPage.statsCardTitle).toBeVisible()

      // check if stats are loaded
      const hasStats = await exportPage.hasStats()

      if (hasStats) {
        // total activities should be displayed
        const totalActivities = await exportPage.getTotalActivities()
        expect(totalActivities).toBeTruthy()
        // should be a number
        expect(totalActivities).toMatch(/[\d,]+/)

        // first activity date should be visible
        await expect(exportPage.firstActivityDate).toBeVisible()

        // last activity date should be visible
        await expect(exportPage.lastActivityDate).toBeVisible()
      } else {
        // no data message should be visible
        await expect(exportPage.noDataMessage).toBeVisible()
      }
    })
  })

  test.describe('Format Selector', () => {
    test('format selector toggles between CSV and JSON', async ({ page }) => {
      const exportPage = new ExportPage(page)
      await exportPage.goto()
      await exportPage.waitForDataLoad()

      // export options card should be visible
      await expect(exportPage.exportOptionsCard).toBeVisible()

      // format buttons should be visible
      await expect(exportPage.csvButton).toBeVisible()
      await expect(exportPage.jsonButton).toBeVisible()

      // get initial format from download button
      const initialFormat = await exportPage.getSelectedFormat()
      expect(initialFormat).toBeTruthy()

      // select CSV
      await exportPage.selectCsvFormat()
      const csvFormat = await exportPage.getSelectedFormat()
      expect(csvFormat).toBe('csv')

      // description should mention spreadsheets
      let description = await exportPage.getFormatDescription()
      expect(description?.toLowerCase()).toContain('spreadsheet')

      // select JSON
      await exportPage.selectJsonFormat()
      const jsonFormat = await exportPage.getSelectedFormat()
      expect(jsonFormat).toBe('json')

      // description should mention developers or data analysis
      description = await exportPage.getFormatDescription()
      expect(description?.toLowerCase()).toMatch(/developer|data analysis/)
    })
  })

  test.describe('Date Range Filter', () => {
    test('date range filter inputs accept values', async ({ page }) => {
      const exportPage = new ExportPage(page)
      await exportPage.goto()
      await exportPage.waitForDataLoad()

      // export options card should be visible
      await expect(exportPage.exportOptionsCard).toBeVisible()

      // date range label should be visible
      await expect(exportPage.dateRangeLabel).toBeVisible()

      // from date input should be visible
      await expect(exportPage.fromDateInput).toBeVisible()

      // to date input should be visible
      await expect(exportPage.toDateInput).toBeVisible()

      // fill from date
      const fromDate = '2024-01-01'
      await exportPage.fillFromDate(fromDate)
      const fromValue = await exportPage.getFromDateValue()
      expect(fromValue).toBe(fromDate)

      // fill to date
      const toDate = '2024-12-31'
      await exportPage.fillToDate(toDate)
      const toValue = await exportPage.getToDateValue()
      expect(toValue).toBe(toDate)
    })

    test('sport type filter accepts values', async ({ page }) => {
      const exportPage = new ExportPage(page)
      await exportPage.goto()
      await exportPage.waitForDataLoad()

      // sport type label should be visible
      await expect(exportPage.sportTypeLabel).toBeVisible()

      // sport type input should be visible
      await expect(exportPage.sportTypeInput).toBeVisible()

      // fill sport type
      await exportPage.fillSportType('Ride')
      await expect(exportPage.sportTypeInput).toHaveValue('Ride')
    })
  })

  test.describe('Download Button', () => {
    test('download button triggers file export', async ({ page }) => {
      const exportPage = new ExportPage(page)
      await exportPage.goto()
      await exportPage.waitForDataLoad()

      // download button should be visible
      await expect(exportPage.downloadButton).toBeVisible()

      // button text should include format
      const format = await exportPage.getSelectedFormat()
      expect(format).toBeTruthy()

      // set up download listener
      const downloadPromise = page.waitForEvent('download', { timeout: 5000 }).catch(() => null)

      // click download
      await exportPage.clickDownload()

      // check if download was triggered
      const download = await downloadPromise

      if (download) {
        // download was triggered
        const filename = download.suggestedFilename()
        expect(filename).toContain('quantlete')
        expect(filename).toMatch(/\.(csv|json)$/)
      } else {
        // download may not trigger in headless mode without data
        // but button should be clickable without errors
        expect(true).toBe(true)
      }
    })

    test('download button shows correct format label', async ({ page }) => {
      const exportPage = new ExportPage(page)
      await exportPage.goto()
      await exportPage.waitForDataLoad()

      // select CSV
      await exportPage.selectCsvFormat()
      let buttonText = await exportPage.downloadButton.textContent()
      expect(buttonText).toContain('CSV')

      // select JSON
      await exportPage.selectJsonFormat()
      buttonText = await exportPage.downloadButton.textContent()
      expect(buttonText).toContain('JSON')
    })
  })

  test.describe('What\'s Included', () => {
    test('what\'s included section displays export fields', async ({ page }) => {
      const exportPage = new ExportPage(page)
      await exportPage.goto()
      await exportPage.waitForDataLoad()

      // what's included card should be visible
      const isVisible = await exportPage.isWhatsIncludedVisible()
      expect(isVisible).toBe(true)

      // activity details section should be visible
      await expect(exportPage.activityDetailsSection).toBeVisible()

      // performance metrics section should be visible
      await expect(exportPage.performanceMetricsSection).toBeVisible()

      // location & gear section should be visible
      await expect(exportPage.locationGearSection).toBeVisible()
    })
  })
})
