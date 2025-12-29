/**
 * Application mode detection utilities.
 *
 * The app can run in two modes:
 * - 'server': Uses Go backend API (default, for self-hosted deployment)
 * - 'wasm': Browser-only with SQLite-WASM (for static hosting)
 */

export type AppMode = 'server' | 'wasm'

/**
 * Get the current application mode.
 * Determined at build time via VITE_BUILD_MODE environment variable.
 */
export function getAppMode(): AppMode {
  const mode = import.meta.env.VITE_BUILD_MODE
  if (mode === 'wasm') {
    return 'wasm'
  }
  return 'server'
}

/**
 * Check if running in server mode (Go backend).
 */
export function isServerMode(): boolean {
  return getAppMode() === 'server'
}

/**
 * Check if running in WASM mode (browser-only).
 */
export function isWasmMode(): boolean {
  return getAppMode() === 'wasm'
}

/**
 * Get the API base URL.
 * Only valid in server mode.
 */
export function getApiBase(): string {
  if (isWasmMode()) {
    throw new Error('API calls are not available in WASM mode')
  }
  return import.meta.env.VITE_API_BASE || '/api/v1'
}

/**
 * Get the Cloudflare Worker URL for OAuth.
 * Only valid in WASM mode.
 */
export function getCfWorkerUrl(): string {
  const url = import.meta.env.VITE_CF_WORKER_URL
  if (!url) {
    throw new Error('VITE_CF_WORKER_URL not configured')
  }
  return url
}

// Export the current mode as a constant for use in conditional rendering
export const APP_MODE = getAppMode()
