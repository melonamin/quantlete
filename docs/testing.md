# Testing Guide

Quantlete uses a multi-layered testing strategy: Go unit tests, React unit tests with Vitest, and end-to-end tests with Playwright.

## Quick Reference

```bash
just test       # run all Go + React unit tests
just test-go    # Go tests only
just test-web   # React unit tests only
just test-e2e   # E2E tests (Playwright in Docker)
```

## End-to-End Tests

E2E tests use Playwright running in Docker against a demo-seeded database. The test suite runs in CI on every pull request.

### Running E2E Tests Locally

```bash
# run full E2E suite
just test-e2e

# run with sequential execution for debugging
PLAYWRIGHT_WORKERS=1 just test-e2e

# run specific test file (requires server running)
cd web && npx playwright test tests/e2e/dashboard.spec.ts
```

The `just test-e2e` command:
1. Builds the web frontend
2. Builds the Go binary
3. Seeds a demo database with 50 activities
4. Starts the Go server on port 8081
5. Runs Playwright tests in Docker
6. Cleans up on exit

### Test Organization

```
web/tests/e2e/
├── fixtures/           # test utilities, custom fixtures
│   └── index.ts        # exports test/expect with base URL
├── pages/              # page object classes
│   ├── base.page.ts    # common page functionality
│   ├── dashboard.page.ts
│   ├── activities.page.ts
│   └── ...
├── journeys/           # cross-page user flow tests
│   ├── dashboard-to-activity.spec.ts
│   ├── activity-exploration.spec.ts
│   └── ...
├── smoke.spec.ts       # basic smoke test
├── dashboard.spec.ts   # dashboard page tests
└── ...
```

## Adding New E2E Tests

### Step 1: Create or Update Page Object

Page objects encapsulate selectors and actions for a page. Create one in `web/tests/e2e/pages/`:

```typescript
import type { Locator, Page } from '@playwright/test'
import { BasePage } from './base.page'

export class ExamplePage extends BasePage {
  // define selectors as class properties
  readonly submitButton: Locator
  readonly titleInput: Locator
  readonly errorMessage: Locator

  constructor(page: Page) {
    super(page)
    this.submitButton = page.locator('button[type="submit"]')
    this.titleInput = page.locator('input[name="title"]')
    this.errorMessage = page.locator('[role="alert"]')
  }

  // required: unique element that identifies this page
  getPageIdentifier(): Locator {
    return this.page.locator('h1:has-text("Example Page")')
  }

  // navigation
  async goto(): Promise<void> {
    await super.goto('/example')
  }

  // page-specific actions
  async fillForm(title: string): Promise<void> {
    await this.titleInput.fill(title)
  }

  async submit(): Promise<void> {
    await this.submitButton.click()
  }
}
```

### Step 2: Write Tests

Create a spec file in `web/tests/e2e/`:

```typescript
import { test, expect } from './fixtures'
import { ExamplePage } from './pages/example.page'

test.describe('Example Page', () => {
  test.describe('Form Submission', () => {
    test('submits form successfully', async ({ page }) => {
      const examplePage = new ExamplePage(page)
      await examplePage.goto()

      await examplePage.fillForm('Test Title')
      await examplePage.submit()

      await expect(page.locator('text=Success')).toBeVisible()
    })

    test('shows error for empty title', async ({ page }) => {
      const examplePage = new ExamplePage(page)
      await examplePage.goto()

      await examplePage.submit()

      await expect(examplePage.errorMessage).toBeVisible()
      await expect(examplePage.errorMessage).toContainText('required')
    })
  })
})
```

### Step 3: For Cross-Page Flows

User journeys that span multiple pages go in `web/tests/e2e/journeys/`:

```typescript
import { test, expect } from '../fixtures'
import { DashboardPage } from '../pages/dashboard.page'
import { ActivityDetailPage } from '../pages/activity-detail.page'

test.describe('Dashboard to Activity Journey', () => {
  test('navigate from dashboard to activity detail and back', async ({ page }) => {
    const dashboard = new DashboardPage(page)
    const activityDetail = new ActivityDetailPage(page)

    await dashboard.goto()
    await dashboard.clickRecentActivity(0)

    await expect(activityDetail.getPageIdentifier()).toBeVisible()
    await activityDetail.goBack()

    await expect(dashboard.getPageIdentifier()).toBeVisible()
  })
})
```

## Debugging Failed Tests

### View Test Report

After test failure, check the HTML report:

```bash
cd web && npx playwright show-report playwright-report/html
```

### Check Artifacts

Failed tests generate:
- Screenshots: `web/test-results/<test-name>/`
- Traces: `web/playwright-report/` (on first retry)
- Videos: `web/test-results/<test-name>/` (on first retry)

### Run in Debug Mode

```bash
# headed browser with Playwright inspector
cd web && npx playwright test --headed --debug tests/e2e/dashboard.spec.ts
```

### Run Single Test

```bash
# run specific test by name
cd web && npx playwright test -g "dashboard loads with stats"

# run specific file
cd web && npx playwright test tests/e2e/dashboard.spec.ts
```

### Sequential Execution

Parallel tests can mask timing issues. Run sequentially:

```bash
PLAYWRIGHT_WORKERS=1 just test-e2e
```

### Verbose Output

```bash
cd web && DEBUG=pw:api npx playwright test tests/e2e/dashboard.spec.ts
```

## CI Integration

E2E tests run in GitHub Actions on every pull request. The CI job:

1. Builds the web frontend and Go binary
2. Seeds a demo database
3. Starts the Go server
4. Runs Playwright tests
5. Uploads artifacts (screenshots, traces, videos) on failure

### Viewing CI Artifacts

1. Go to the failed workflow run
2. Scroll to "Artifacts" section
3. Download `playwright-report`
4. Extract and open `html/index.html`

## Configuration

Playwright configuration is in `web/playwright.config.ts`:

- **testDir**: `./tests/e2e`
- **retries**: 2 in CI, 0 locally
- **workers**: 1 in CI, 80% of CPUs locally
- **baseURL**: `http://localhost:8081` (or `host.docker.internal:8081` in Docker)
- **timeout**: 30 seconds per test
- **artifacts**: screenshots on failure, traces/videos on first retry

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PLAYWRIGHT_WORKERS` | Number of parallel workers | 80% of CPUs |
| `PLAYWRIGHT_TIMEOUT` | Test timeout in ms | 30000 |
| `CI` | Enables CI mode (retries, JUnit output) | - |
| `E2E_IN_DOCKER` | Set automatically when running in Docker | - |

## Common Patterns

### Waiting for Loading States

```typescript
// wait for loading skeleton to disappear
await page.locator('[class*="animate-pulse"]').waitFor({ state: 'hidden' })

// wait for specific content
await expect(page.locator('text=Data loaded')).toBeVisible({ timeout: 10000 })
```

### Handling Conditional UI

```typescript
// check if element exists without failing
const hasData = await page.locator('text=Activities').count() > 0
if (hasData) {
  // test with data
} else {
  // test empty state
}
```

### Reliable Selectors

Prefer these selector strategies (in order):

1. Role-based: `page.getByRole('button', { name: 'Submit' })`
2. Test IDs: `page.locator('[data-testid="activity-row"]')`
3. Text content: `page.locator('text=Recent Activities')`
4. Class patterns: `page.locator('[class*="widget"]')`

Avoid:
- Brittle CSS selectors: `.container > div:nth-child(3) > span`
- Auto-generated class names: `.css-1a2b3c`
