/// <reference types="vite/client" />

interface ImportMetaEnv {
  /**
   * Build mode: 'server' for Go backend, 'wasm' for browser-only
   */
  readonly VITE_BUILD_MODE: 'server' | 'wasm'

  /**
   * API base URL (server mode only)
   */
  readonly VITE_API_BASE?: string

  /**
   * Cloudflare Worker URL for OAuth token exchange (wasm mode only)
   */
  readonly VITE_CF_WORKER_URL?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
