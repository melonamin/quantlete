# Plan: Comprehensive E2E Test Suite

Implement a full end-to-end test suite for Quantlete using Playwright in Docker. Tests run against Server Mode with a demo-seeded database. The suite uses a hybrid organization: page-centric tests for thorough coverage of each route, plus cross-page journey tests to catch integration issues.

## Validation Commands

- `just test-e2e`
- `just lint`
- `just typecheck`

### Task 1: Playwright Setup & Test Infrastructure (Docker-based)

Establish the foundation for running Playwright tests in Docker since Playwright cannot be installed locally. This includes the Docker wrapper script, Playwright configuration, test utilities, and the first smoke test to verify the setup works end-to-end.

- [x] Create `docker-playwright.sh` script adapted from pondpilot pattern (uses `mcr.microsoft.com/playwright` image)
- [x] Add Playwright dev dependency to `web/package.json`
- [x] Create `web/playwright.config.ts` configured for Go server on port 8081
- [x] Create `just test-e2e` command that seeds demo DB, starts Go server, and runs Playwright in Docker
- [x] Create `web/tests/e2e/fixtures/` directory with test fixtures and auth helpers
- [x] Create `web/tests/e2e/pages/base.page.ts` with common page object patterns
- [x] Create `web/tests/e2e/smoke.spec.ts` - app loads, dashboard renders with demo data

### Task 2: Dashboard Page Tests

The dashboard is the main entry point with 25+ configurable widgets displaying stats, charts, and activity summaries. Testing it ensures the core data display and widget system works correctly.

- [ ] Create `web/tests/e2e/pages/dashboard.page.ts` with widget and stats selectors
- [ ] Test: Dashboard loads with stats summary (activities, distance, elevation, time)
- [ ] Test: Visible widgets render with data (weekly stats, recent activities, sport breakdown)
- [ ] Test: Monthly chart and activity calendar widgets render
- [ ] Test: Widget visibility toggle works (show/hide widgets)
- [ ] Test: Time period filters update chart data
- [ ] Test: Navigation to activity detail from "recent activities" widget works

### Task 3: Activities Page Tests

The activities page is the main data table with extensive filtering, sorting, and pagination. This is one of the most interactive pages requiring thorough testing of all filter combinations.

- [ ] Create `web/tests/e2e/pages/activities.page.ts` with filter and table selectors
- [ ] Test: Activity table loads with paginated data
- [ ] Test: Sorting by columns (date, distance, elevation, duration) works
- [ ] Test: Filter by sport type narrows results
- [ ] Test: Filter by date range works
- [ ] Test: Filter by gear works
- [ ] Test: Commute filter toggle works
- [ ] Test: Combined filters work together correctly
- [ ] Test: Filter reset clears all filters and shows all data
- [ ] Test: Click row navigates to activity detail page
- [ ] Test: Pagination controls (next/prev/jump to page) work

### Task 4: Activity Detail Page Tests

Individual activity view showing maps, elevation profiles, stream charts, and segment efforts. This page combines multiple complex visualizations that need to render correctly with activity data.

- [ ] Create `web/tests/e2e/pages/activity-detail.page.ts` with stats, chart, and map selectors
- [ ] Test: Activity detail page loads with correct activity data
- [ ] Test: Stats cards display correctly (distance, elevation, duration, avg pace/HR/power)
- [ ] Test: Elevation profile chart renders
- [ ] Test: Activity stream chart renders (speed/HR/power over distance)
- [ ] Test: GPS map renders with route polyline
- [ ] Test: Segment efforts table displays when activity has segments
- [ ] Test: Photos gallery displays when activity has photos
- [ ] Test: Back navigation returns to activities list

### Task 5: Heatmap & Calendar Page Tests

Map visualization showing all activity routes and calendar view with activity dots. These pages test Leaflet map integration and date-based navigation.

- [ ] Create `web/tests/e2e/pages/heatmap.page.ts` with map and filter selectors
- [ ] Test: Heatmap loads with activity routes rendered on map
- [ ] Test: Sport type quick filters (All, Ride, Run, Walk) update map display
- [ ] Test: Country dropdown lists visited countries with activity counts
- [ ] Test: Country selection flies map to selected country bounds
- [ ] Test: Advanced filters (date range, commute) filter routes
- [ ] Test: Clear filters resets map to default view
- [ ] Create `web/tests/e2e/pages/calendar.page.ts` with calendar grid selectors
- [ ] Test: Calendar grid renders with current month
- [ ] Test: Activity dots appear on days with activities
- [ ] Test: Month navigation (prev/next buttons) works
- [ ] Test: "Today" button returns to current month
- [ ] Test: Stats cards update when navigating months
- [ ] Test: Click day with activities shows activity info

### Task 6: Analytics Pages Tests (Power, Training Load, Best Efforts, Eddington)

Performance analytics pages showing power curves, training stress, personal records, and Eddington numbers. These pages have complex charts and require data from multiple sources.

- [ ] Create `web/tests/e2e/pages/power.page.ts` with chart selectors
- [ ] Test: All-time best power outputs bar chart renders
- [ ] Test: Duration selector dropdown changes progression chart view
- [ ] Test: Power curve comparison chart (all-time vs 90 days) renders
- [ ] Test: Power zones breakdown chart renders with zone colors
- [ ] Create `web/tests/e2e/pages/training-load.page.ts` with chart and card selectors
- [ ] Test: Training load chart (CTL/ATL/TSB curves) renders
- [ ] Test: Date range filters update chart data
- [ ] Test: Summary cards display current CTL/ATL/TSB values
- [ ] Test: Configuration warning shows when FTP not configured
- [ ] Create `web/tests/e2e/pages/best-efforts.page.ts` with distance list and modal selectors
- [ ] Test: Standard distances list renders with times or "No data"
- [ ] Test: Sport filter (All/Runs/Rides) updates distance list
- [ ] Test: Click distance row opens PR progression modal
- [ ] Test: Modal shows PR chart and efforts table
- [ ] Test: Modal close button works
- [ ] Create `web/tests/e2e/pages/eddington.page.ts` with number and chart selectors
- [ ] Test: Current Eddington number displays prominently
- [ ] Test: History progression chart renders
- [ ] Test: View mode tabs (All/Sport group/Custom) switch correctly

### Task 7: Content Pages Tests (Segments, Photos, Challenges, Wrapped, Badges)

Content browsing pages for segments, photos, challenges, yearly wrapped summaries, and shareable badges. These pages test galleries, modals, external links, and download functionality.

- [ ] Create `web/tests/e2e/pages/segments.page.ts` with table and modal selectors
- [ ] Test: Segment table loads with segment data
- [ ] Test: Search by name filters results with debounce
- [ ] Test: Sport type quick filters (All, Ride, Run) work
- [ ] Test: Starred and KOM toggles filter results
- [ ] Test: Column header sorting works
- [ ] Test: Click row opens segment detail modal
- [ ] Test: Modal shows map, PR progression chart, and efforts table
- [ ] Create `web/tests/e2e/pages/photos.page.ts` with gallery and lightbox selectors
- [ ] Test: Photo gallery loads with images in masonry grid
- [ ] Test: Sport filter buttons update gallery contents
- [ ] Test: Country dropdown filters photos by location
- [ ] Test: Click photo opens lightbox modal
- [ ] Test: Lightbox prev/next navigation works
- [ ] Test: "Load more" pagination loads additional photos
- [ ] Create `web/tests/e2e/pages/challenges.page.ts` with badge grid selectors
- [ ] Test: Challenges display grouped by month
- [ ] Test: Badge images load correctly
- [ ] Test: Badge click opens external Strava link (verify href)
- [ ] Create `web/tests/e2e/pages/wrapped.page.ts` with year selector and chart selectors
- [ ] Test: Year selector dropdown changes displayed data
- [ ] Test: Metric cards display correct stats for selected year
- [ ] Test: All charts render (heatmap calendar, monthly bars, donut charts)
- [ ] Test: Comparison year selector shows delta comparison table
- [ ] Create `web/tests/e2e/pages/badges.page.ts` with customizer and preview selectors
- [ ] Test: Badge previews render with default styling
- [ ] Test: Theme selector updates badge preview
- [ ] Test: Size selector updates badge preview
- [ ] Test: Background selector updates badge preview
- [ ] Test: Download button triggers file download

### Task 8: Configuration Pages Tests (Settings, Gear, Athlete, Export, Monthly Stats)

User settings and data management pages for configuring the app, managing gear, editing athlete metrics, and exporting data.

- [ ] Create `web/tests/e2e/pages/settings.page.ts` with form and toggle selectors
- [ ] Test: Strava connection status displays correctly (connected with athlete name)
- [ ] Test: Unit system toggle (Metric/Imperial) changes and persists
- [ ] Test: Theme selector (System/Light/Dark) changes theme
- [ ] Test: Import options checkboxes toggle correctly
- [ ] Test: Sync history modal opens and displays past sync records
- [ ] Create `web/tests/e2e/pages/gear.page.ts` with list and modal selectors
- [ ] Test: Gear list displays bikes and shoes with usage statistics
- [ ] Test: Show/Hide Retired toggle filters gear list
- [ ] Test: Maintenance tab switch displays maintenance view
- [ ] Test: Gear usage chart renders
- [ ] Test: Add custom gear modal opens, accepts input, and submits
- [ ] Create `web/tests/e2e/pages/athlete.page.ts` with editor table selectors
- [ ] Test: Athlete info card displays profile information
- [ ] Test: FTP history table displays existing entries
- [ ] Test: Add FTP entry form works (date + watts)
- [ ] Test: Edit existing FTP entry works
- [ ] Test: Delete FTP entry works with confirmation
- [ ] Test: Weight history table displays existing entries
- [ ] Test: Add/edit/delete weight entries work
- [ ] Test: HR zones editor displays zone configuration
- [ ] Create `web/tests/e2e/pages/export.page.ts` with form selectors
- [ ] Test: Data stats card shows activity count and date range
- [ ] Test: Format selector toggles between CSV and JSON
- [ ] Test: Date range filter inputs accept values
- [ ] Test: Download button triggers file export
- [ ] Create `web/tests/e2e/pages/monthly-stats.page.ts` with table selectors
- [ ] Test: Year navigation (prev/next) changes displayed year
- [ ] Test: Yearly summary cards display correct totals
- [ ] Test: Accordion table rows expand and collapse
- [ ] Test: Export CSV button triggers download

### Task 9: Cross-Page User Journey Tests

End-to-end user flows that span multiple pages to catch integration issues and verify navigation state is preserved correctly across the application.

- [ ] Create `web/tests/e2e/journeys/` directory for journey test files
- [ ] Journey: Dashboard to Activity Deep Dive - dashboard → click recent activity → view detail → check segments → back to dashboard
- [ ] Journey: Activity Exploration Flow - activities list → apply filters → click activity → view segments → return to filtered list (verify filters preserved)
- [ ] Journey: Settings and Data Refresh - settings → change unit system → activities page → verify units changed → dashboard → verify widget units
- [ ] Journey: Analytics Navigation - dashboard → power page → verify charts → training load → best efforts → back to dashboard
- [ ] Journey: Content Browsing - heatmap → select country → calendar → navigate months → photos → browse gallery → wrapped → compare years
- [ ] Journey: Configuration Round Trip - athlete → edit FTP → gear → check usage stats → settings → verify connection → export → download data

### Task 10: CI/CD Integration & Documentation

Make E2E tests run automatically in GitHub Actions and document the test suite for future contributors.

- [ ] Add E2E test job to `.github/workflows/ci.yml` using Docker Playwright image
- [ ] Configure CI job to seed demo database before tests
- [ ] Configure CI job to start Go server and wait for readiness
- [ ] Configure CI job to upload test artifacts (screenshots, videos) on failure
- [ ] Add JUnit XML reporter for CI test result parsing
- [ ] Update `CLAUDE.md` with E2E testing section (commands, patterns)
- [ ] Update project `README.md` with E2E test commands
- [ ] Create `docs/testing.md` documenting how to run tests locally
- [ ] Document how to add new E2E tests with page object examples
- [ ] Document debugging strategies for failed tests
