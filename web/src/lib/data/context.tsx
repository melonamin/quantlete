/* eslint-disable react-refresh/only-export-components */
/**
 * React context for DataProvider injection.
 *
 * This allows components to access data without knowing whether
 * the app is running in server mode or WASM mode.
 */

import { createContext, useContext, useState, useEffect, type ReactNode } from 'react'
import type { DataProvider } from './provider'
import { isWasmMode } from '@/lib/mode'

interface DataProviderState {
  provider: DataProvider | null
  initialized: boolean
  error: string | null
}

const DataProviderContext = createContext<DataProviderState>({
  provider: null,
  initialized: false,
  error: null,
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
 * Wrapper component that initializes the appropriate DataProvider
 * based on the current app mode (server or wasm).
 */
export function DataProviderWrapper({ children }: DataProviderWrapperProps) {
  const [state, setState] = useState<DataProviderState>({
    provider: null,
    initialized: false,
    error: null,
  })

  useEffect(() => {
    async function initialize() {
      try {
        if (isWasmMode()) {
          // Lazy load Go WASM provider to avoid bundling in server mode
          const { GoWasmProvider } = await import('./wasm/go-provider')
          const provider = new GoWasmProvider()
          await provider.initialize()
          setState({ provider, initialized: true, error: null })
        } else {
          // Server mode - use ServerProvider
          const { ServerProvider } = await import('./server/provider')
          const provider = new ServerProvider()
          setState({ provider, initialized: true, error: null })
        }
      } catch (err) {
        console.error('[DataProvider] Failed to initialize:', err)
        setState({
          provider: null,
          initialized: false,
          error: err instanceof Error ? err.message : 'Unknown error',
        })
      }
    }

    initialize()
  }, [])

  return <DataProviderContext.Provider value={state}>{children}</DataProviderContext.Provider>
}
