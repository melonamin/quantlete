import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  initGoStorage: vi.fn(async () => {}),
  isInitialized: vi.fn(() => true),
  persistDatabase: vi.fn(async () => {}),
  updateAppSettings: vi.fn(),
  loadAuth: vi.fn(async () => false),
}))

vi.mock('@/lib/wasm/go-storage', () => ({
  initGoStorage: mocks.initGoStorage,
  isInitialized: mocks.isInitialized,
  persistDatabase: mocks.persistDatabase,
  updateAppSettings: mocks.updateAppSettings,
}))

vi.mock('@/lib/wasm/strava', () => ({
  loadAuth: mocks.loadAuth,
  isAuthenticated: vi.fn(() => false),
  getAthlete: vi.fn(() => null),
  getAuthUrl: vi.fn(),
  exchangeCode: vi.fn(),
}))

vi.mock('@/lib/wasm/strava/credentials', () => ({
  getCredentials: vi.fn(() => null),
  saveCredentials: vi.fn(),
}))

let GoWasmProvider: typeof import('./go-provider').GoWasmProvider

beforeAll(async () => {
  const windowTarget = new EventTarget() as EventTarget & { location: { origin: string } }
  windowTarget.location = { origin: 'https://quantlete.test' }
  vi.stubGlobal('window', windowTarget)

  const documentTarget = new EventTarget() as EventTarget & { visibilityState: string }
  documentTarget.visibilityState = 'visible'
  vi.stubGlobal('document', documentTarget)
  ;({ GoWasmProvider } = await import('./go-provider'))
})

afterAll(() => {
  vi.unstubAllGlobals()
})

describe('GoWasmProvider persistence', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('persists when an import completes', () => {
    const provider = new GoWasmProvider()
    const unsubscribe = provider.subscribeToEvents(vi.fn())

    ;(
      window as Window & {
        onImportComplete?: (json: string) => void
      }
    ).onImportComplete?.(JSON.stringify({ success: true }))

    expect(mocks.persistDatabase).toHaveBeenCalledOnce()
    unsubscribe()
  })

  it('awaits persistence after a provider mutation', async () => {
    const provider = new GoWasmProvider()

    await provider.updateAppSettings({
      version: 1,
      virtual_world_tile_layers: {},
      scheduler: {
        version: 1,
        pull: { enabled: false, schedule: 'midnight' },
        push: { enabled: false },
      },
    })

    expect(mocks.updateAppSettings).toHaveBeenCalledOnce()
    expect(mocks.persistDatabase).toHaveBeenCalledOnce()
    expect(mocks.updateAppSettings.mock.invocationCallOrder[0]).toBeLessThan(
      mocks.persistDatabase.mock.invocationCallOrder[0]
    )
  })

  it('persists on pagehide and removes the listener on dispose', async () => {
    const provider = new GoWasmProvider()
    await provider.initialize()

    window.dispatchEvent(new Event('pagehide'))
    expect(mocks.persistDatabase).toHaveBeenCalledOnce()

    provider.dispose()
    window.dispatchEvent(new Event('pagehide'))
    expect(mocks.persistDatabase).toHaveBeenCalledOnce()
  })
})
