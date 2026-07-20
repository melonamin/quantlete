/* eslint-disable react-refresh/only-export-components */
/**
 * React context for DataProvider injection.
 *
 * This allows components to access data without knowing whether
 * the app is running in server mode or WASM mode.
 */

import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import type { DataProvider } from './provider'
import { isWasmMode, isDemoMode, getDemoDbUrl } from '@/lib/mode'

interface DataProviderState {
  provider: DataProvider | null
  initialized: boolean
  error: string | null
  loadingMessage: string | null
}

const DataProviderContext = createContext<DataProviderState>({
  provider: null,
  initialized: false,
  error: null,
  loadingMessage: null,
})

/**
 * Hook to access the DataProvider.
 * Throws if used before initialization is complete.
 */
export function useDataProvider(): DataProvider {
  const { provider, initialized, error } = useContext(DataProviderContext)

  if (error) {
    throw new Error(`DataProvider initialization failed: ${error}`)
  }

  if (!initialized || !provider) {
    throw new Error('DataProvider not initialized. Wrap your app with DataProviderWrapper.')
  }

  return provider
}

/**
 * Hook to check if DataProvider is ready.
 * Safe to call during initialization.
 */
export function useDataProviderStatus(): DataProviderState {
  return useContext(DataProviderContext)
}

interface DataProviderWrapperProps {
  children: ReactNode
}

/**
 * Loading screen for demo mode initialization.
 */
function DemoLoadingScreen({ message }: { message: string }) {
  return (
    <div className="fixed inset-0 z-50 flex flex-col items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-6">
        <div className="h-12 w-12 animate-spin rounded-full border-4 border-muted border-t-primary" />
        <div className="text-center">
          <h2 className="text-lg font-semibold">Loading Demo</h2>
          <p className="mt-1 text-sm text-muted-foreground">{message}</p>
        </div>
      </div>
    </div>
  )
}

/**
 * Wrapper component that initializes the appropriate DataProvider
 * based on the current app mode (server or wasm).
 */
export function DataProviderWrapper({ children }: DataProviderWrapperProps) {
  const [state, setState] = useState<DataProviderState>({
    provider: null,
    initialized: false,
    error: null,
    loadingMessage: null,
  })

  useEffect(() => {
    let disposed = false
    let providerToDispose: DataProvider | null = null

    async function initialize() {
      try {
        if (isDemoMode()) {
          // Demo mode - use GoWasmProvider with bundled database
          setState((s) => ({ ...s, loadingMessage: 'Downloading demo database...' }))
          const { GoWasmProvider } = await import('./wasm/go-provider')
          const provider = new GoWasmProvider({
            demoMode: true,
            demoDatabaseUrl: getDemoDbUrl(),
          })
          providerToDispose = provider
          setState((s) => ({ ...s, loadingMessage: 'Initializing...' }))
          await provider.initialize()
          if (disposed) {
            provider.dispose()
            return
          }
          setState({ provider, initialized: true, error: null, loadingMessage: null })
        } else if (isWasmMode()) {
          // Lazy load Go WASM provider to avoid bundling in server mode
          const { GoWasmProvider } = await import('./wasm/go-provider')
          const provider = new GoWasmProvider()
          providerToDispose = provider
          await provider.initialize()
          if (disposed) {
            provider.dispose()
            return
          }
          setState({ provider, initialized: true, error: null, loadingMessage: null })
        } else {
          // Server mode - use ServerProvider
          const { ServerProvider } = await import('./server/provider')
          const provider = new ServerProvider()
          providerToDispose = provider
          if (disposed) {
            return
          }
          setState({ provider, initialized: true, error: null, loadingMessage: null })
        }
      } catch (err) {
        if (disposed) {
          return
        }
        console.error('[DataProvider] Failed to initialize:', err)
        setState({
          provider: null,
          initialized: false,
          error: err instanceof Error ? err.message : 'Unknown error',
          loadingMessage: null,
        })
      }
    }

    void initialize()

    return () => {
      disposed = true
      providerToDispose?.dispose?.()
    }
  }, [])

  // Show loading screen for demo mode
  if (isDemoMode() && !state.initialized && !state.error && state.loadingMessage) {
    return (
      <>
        <DemoLoadingScreen message={state.loadingMessage} />
        <DataProviderContext.Provider value={state}>{children}</DataProviderContext.Provider>
      </>
    )
  }

  return <DataProviderContext.Provider value={state}>{children}</DataProviderContext.Provider>
}
