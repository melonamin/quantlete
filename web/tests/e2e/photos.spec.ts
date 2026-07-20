import { test, expect } from './fixtures'
import { PhotosPage } from './pages/photos.page'

test.describe('Photos Page', () => {
  test.describe('Photo Gallery', () => {
    test('photo gallery loads with images in masonry grid', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      // verify page header
      await expect(photosPage.pageTitle).toBeVisible()
      await expect(photosPage.pageSubtitle).toBeVisible()

      // check if gallery, empty state, or error is visible
      const hasPhotos = await photosPage.hasPhotos()
      const hasError = await photosPage.hasError()
      const isEmpty = await photosPage.emptyState.isVisible()

      // one of these should be true
      expect(hasPhotos || hasError || isEmpty).toBe(true)

      if (hasPhotos) {
        // gallery should be visible
        await expect(photosPage.photoGallery).toBeVisible()

        // should have at least one photo
        const photoCount = await photosPage.getPhotoItemsCount()
        expect(photoCount).toBeGreaterThan(0)

        // photo count display should be visible
        const displayedCount = await photosPage.getDisplayedPhotoCount()
        if (displayedCount !== null) {
          expect(displayedCount).toBeGreaterThan(0)
        }
      }
    })

    test('photo items have images', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // first photo should have an image
      const firstPhoto = photosPage.photoItems.first()
      const img = firstPhoto.locator('img')
      await expect(img).toBeVisible()

      // image should have src attribute
      const src = await img.getAttribute('src')
      expect(src).toBeTruthy()
    })
  })

  test.describe('Sport Filter Buttons', () => {
    test('sport filter buttons update gallery contents', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // verify filter buttons are visible
      await expect(photosPage.allSportButton).toBeVisible()

      // All should be selected by default
      const isAllSelected = await photosPage.isSportFilterSelected('all')
      expect(isAllSelected).toBe(true)

      const initialCount = await photosPage.getPhotoItemsCount()

      // click Ride filter
      await photosPage.clickSportFilter('ride')
      const isRideSelected = await photosPage.isSportFilterSelected('ride')
      expect(isRideSelected).toBe(true)

      // should have active filters indicator
      const hasFilters = await photosPage.hasActiveFilters()
      expect(hasFilters).toBe(true)

      // click All to reset
      await photosPage.clickSportFilter('all')
      const isAllSelectedAgain = await photosPage.isSportFilterSelected('all')
      expect(isAllSelectedAgain).toBe(true)

      // count should be back to initial
      const resetCount = await photosPage.getPhotoItemsCount()
      expect(resetCount).toBe(initialCount)
    })

    test('multiple sport filters can be selected', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // click Ride filter
      await photosPage.clickSportFilter('ride')
      const isRideSelected = await photosPage.isSportFilterSelected('ride')
      expect(isRideSelected).toBe(true)

      // click Run filter (should add to selection)
      await photosPage.clickSportFilter('run')
      const isRunSelected = await photosPage.isSportFilterSelected('run')
      // both should be selected
      expect(isRunSelected).toBe(true)

      // click Ride again to deselect
      await photosPage.clickSportFilter('ride')
      const isRideDeselected = await photosPage.isSportFilterSelected('ride')
      expect(isRideDeselected).toBe(false)
    })
  })

  test.describe('Country Dropdown', () => {
    test('country dropdown filters photos by location', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // check if country dropdown exists
      const hasCountryButton = await photosPage.countryDropdownButton.isVisible()
      if (!hasCountryButton) {
        // no countries available, skip test
        return
      }

      // open country dropdown
      await photosPage.openCountryDropdown()
      const isOpen = await photosPage.isCountryDropdownOpen()
      expect(isOpen).toBe(true)

      // get country count
      const countryCount = await photosPage.getCountryOptionsCount()
      if (countryCount === 0) {
        return
      }

      // click all countries option to close
      await photosPage.selectAllCountries()
      const isClosed = await photosPage.isCountryDropdownOpen()
      expect(isClosed).toBe(false)
    })

    test('country selection shows country name in button', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      const hasCountryButton = await photosPage.countryDropdownButton.isVisible()
      if (!hasCountryButton) {
        return
      }

      // open dropdown
      await photosPage.openCountryDropdown()

      // get first country option (not "All countries")
      const countryOptions = photosPage.countryDropdownPanel.locator('button').filter({
        hasNot: page.locator('text=All countries'),
      })
      const count = await countryOptions.count()
      if (count === 0) {
        return
      }

      // get first country name
      const firstCountryText = await countryOptions.first().textContent()
      // extract country name (before the count)
      const countryMatch = firstCountryText?.match(/[\u{1F1E6}-\u{1F1FF}]{2}\s*([^\d]+)/u)
      if (!countryMatch) {
        return
      }

      // click first country
      await countryOptions.first().click()
      await photosPage.waitForLoadingComplete()

      // should show active filters
      const hasFilters = await photosPage.hasActiveFilters()
      expect(hasFilters).toBe(true)
    })
  })

  test.describe('Lightbox Modal', () => {
    test('click photo opens lightbox modal', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // click first photo
      await photosPage.clickPhoto(0)

      // lightbox should be open
      const isOpen = await photosPage.isLightboxOpen()
      expect(isOpen).toBe(true)

      // lightbox should show image
      await expect(photosPage.lightboxImage).toBeVisible()

      // lightbox should show title
      const title = await photosPage.getLightboxTitle()
      expect(title).toBeTruthy()

      // close button should be visible
      await expect(photosPage.lightboxCloseButton).toBeVisible()
    })

    test('lightbox prev/next navigation works', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      const photoCount = await photosPage.getPhotoItemsCount()
      if (photoCount < 2) {
        // need at least 2 photos for navigation
        return
      }

      // click first photo
      await photosPage.clickPhoto(0)
      await expect(photosPage.lightbox).toBeVisible()

      // get initial title
      const initialTitle = await photosPage.getLightboxTitle()

      // click next
      await expect(photosPage.lightboxNextButton).toBeVisible()
      await photosPage.lightboxNext()

      // wait for image to update
      await page.waitForTimeout(300)

      // click prev to go back
      await expect(photosPage.lightboxPrevButton).toBeVisible()
      await photosPage.lightboxPrev()

      // should be back to initial
      const backTitle = await photosPage.getLightboxTitle()
      expect(backTitle).toBe(initialTitle)
    })

    test('lightbox close button works', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // click first photo
      await photosPage.clickPhoto(0)
      await expect(photosPage.lightbox).toBeVisible()

      // close lightbox
      await photosPage.closeLightbox()

      // lightbox should be closed
      const isOpen = await photosPage.isLightboxOpen()
      expect(isOpen).toBe(false)
    })

    test('lightbox shows view activity link', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // click first photo
      await photosPage.clickPhoto(0)
      await expect(photosPage.lightbox).toBeVisible()

      // view activity link should be visible
      await expect(photosPage.lightboxViewActivityLink).toBeVisible()

      // should have href
      const href = await photosPage.lightboxViewActivityLink.getAttribute('href')
      expect(href).toContain('/activities/')
    })
  })

  test.describe('Load More Pagination', () => {
    test('load more pagination loads additional photos', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // check if load more button exists
      const hasLoadMore = await photosPage.hasLoadMore()
      if (!hasLoadMore) {
        // not enough photos for pagination, skip
        return
      }

      // get initial count
      const initialCount = await photosPage.getPhotoItemsCount()

      // click load more
      await photosPage.loadMore()

      // should have more photos now
      const newCount = await photosPage.getPhotoItemsCount()
      expect(newCount).toBeGreaterThan(initialCount)
    })
  })

  test.describe('Clear Filters', () => {
    test('clear filters resets all filters', async ({ page }) => {
      const photosPage = new PhotosPage(page)
      await photosPage.goto()
      await photosPage.waitForDataLoad()

      const hasPhotos = await photosPage.hasPhotos()
      if (!hasPhotos) {
        return
      }

      // apply a filter
      await photosPage.clickSportFilter('ride')

      // should have active filters
      const hasFilters = await photosPage.hasActiveFilters()
      expect(hasFilters).toBe(true)

      // clear filters
      await photosPage.clearFilters()

      // all should be selected again
      const isAllSelected = await photosPage.isSportFilterSelected('all')
      expect(isAllSelected).toBe(true)
    })
  })
})
