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
// E2E_PORT must match scripts/e2e-setup.sh (overridable so local runs can
// avoid a developer's own server on 8081)
const e2ePort = process.env.E2E_PORT ?? '8081'
const serverBaseURL = isInDocker
  ? `http://host.docker.internal:${e2ePort}`
  : `http://localhost:${e2ePort}`
const wasmBaseURL = 'http://localhost:4174'
const wasmPersistBaseURL = 'http://localhost:4175'

// Playwright runs every configured project when --project is omitted. Only
// expose the opt-in WASM project (and its static server) when it is requested,
// so the existing server-mode CI command keeps its current scope.
// Keyed on an env var (not argv): worker processes re-evaluate this config
// with different argv, and a project that exists only in the runner breaks
// with "Project not found in the worker process".
const wasmProjectRequested = process.env.WASM_E2E === '1'

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
    baseURL: serverBaseURL,
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
      testIgnore: '**/wasm/**',
      use: { ...devices['Desktop Chrome'] },
    },
    ...(wasmProjectRequested
      ? [
          {
            name: 'wasm',
            testDir: './tests/e2e/wasm',
            testIgnore: 'persistence.spec.ts',
            use: {
              ...devices['Desktop Chrome'],
              baseURL: wasmBaseURL,
            },
          },
          {
            name: 'wasm-persist',
            testDir: './tests/e2e/wasm',
            testMatch: 'persistence.spec.ts',
            use: {
              ...devices['Desktop Chrome'],
              baseURL: wasmPersistBaseURL,
            },
          },
        ]
      : []),
  ],
  webServer: wasmProjectRequested
    ? [
        {
          command: 'npx vite preview --outDir dist-demo --host localhost --port 4174 --strictPort',
          url: wasmBaseURL,
          reuseExistingServer: !process.env.CI,
          timeout: 120_000,
        },
        {
          command: 'npx vite preview --outDir dist-wasm --host localhost --port 4175 --strictPort',
          url: wasmPersistBaseURL,
          reuseExistingServer: !process.env.CI,
          timeout: 120_000,
        },
      ]
    : undefined,
  // set timeout for each test
  timeout: Number(process.env.PLAYWRIGHT_TIMEOUT ?? 30_000),
  // expect timeout
  expect: {
    timeout: 10_000,
  },
  // output directory for test artifacts
  outputDir: './test-results',
})
