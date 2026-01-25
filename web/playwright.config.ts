import { defineConfig, devices } from '@playwright/test'

// allow overriding workers via environment variable
const workersOverride = process.env.PLAYWRIGHT_WORKERS
const parsedWorkersOverride = workersOverride
  ? Number.isNaN(Number(workersOverride))
    ? workersOverride
    : Number(workersOverride)
  : undefined

// in Docker, use host.docker.internal to reach host machine services
// E2E_IN_DOCKER is explicitly set by docker-playwright.sh
const isInDocker = process.env.E2E_IN_DOCKER === 'true'
const baseURL = isInDocker
  ? 'http://host.docker.internal:8081'
  : 'http://localhost:8081'

export default defineConfig({
  testDir: './tests/e2e',
  // fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  // retry on CI only
  retries: process.env.CI ? 2 : 0,
  // opt out of parallel tests on CI
  workers: parsedWorkersOverride ?? (process.env.CI ? 1 : '80%'),
  // reporter configuration
  reporter: process.env.CI
    ? [
        ['junit', { outputFile: './playwright-report/results.xml' }],
        ['html', { outputFolder: './playwright-report/html', open: 'never' }],
      ]
    : [['list', { printSteps: true }]],
  // shared settings for all projects
  use: {
    baseURL,
    // collect trace on first retry
    trace: 'on-first-retry',
    // take screenshot on failure
    screenshot: 'only-on-failure',
    // record video on failure
    video: 'on-first-retry',
  },
  // configure projects for major browsers
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  // set timeout for each test
  timeout: Number(process.env.PLAYWRIGHT_TIMEOUT ?? 30_000),
  // expect timeout
  expect: {
    timeout: 10_000,
  },
  // output directory for test artifacts
  outputDir: './test-results',
})
