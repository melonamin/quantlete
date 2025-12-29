/**
 * Data layer abstraction module.
 *
 * This module provides a unified interface for data access that works
 * in both server mode (Go backend) and WASM mode (browser-only).
 *
 * Usage:
 * 1. Wrap your app with DataProviderWrapper
 * 2. Use useDataProvider() hook to access the provider
 * 3. Use useDataProviderStatus() to check initialization status
 */

// Provider types
export { type DataProvider } from './provider'

// React context
export { DataProviderWrapper, useDataProvider, useDataProviderStatus } from './context'

// React Query hooks (work with both server and WASM modes)
export * from './hooks'

// Re-export all data types
export * from './types'
