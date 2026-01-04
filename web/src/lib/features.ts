/**
 * Feature flags for conditional functionality based on app mode.
 *
 * Server mode (Go backend) has full functionality.
 * WASM mode (browser-only) has some features disabled or replaced.
 */

import { isServerMode, isWasmMode } from './mode'

export const features = {
  // Authentication & Data Import
  /** OAuth flow through Go backend */
  stravaOAuth: isServerMode(),
  /** Server-side data import from Strava */
  dataImport: isServerMode(),
  /** Scheduled background imports */
  scheduledImport: isServerMode(),
  /** Strava webhook support */
  webhooks: isServerMode(),
  /** Challenge scraping from Strava profile */
  challengeScraping: isServerMode(),
  /** Challenge import from Strava trophy case HTML */
  challengeImport: isServerMode(),

  // Data Sources
  /** REST API access to Go backend */
  apiAccess: isServerMode(),
  /** Local SQLite database in browser */
  localDatabase: isWasmMode(),

  // WASM-specific features
  /** Browser-based OAuth via Cloudflare Worker */
  browserOAuth: isWasmMode(),
  /** Browser-based activity import */
  browserImport: isWasmMode(),
  /** OPFS database persistence */
  opfsPersistence: isWasmMode(),

  // UI Sections to show/hide
  /** Show Strava connection section in settings (server uses OAuth redirect) */
  showStravaConnect: isServerMode(),
  /** Show data import section in settings */
  showImportSection: isServerMode(),
  /** Show webhook configuration in settings */
  showWebhookSettings: isServerMode(),
  /** Show WASM-specific settings (database management, file import) */
  showWasmSettings: isWasmMode(),

  // Features that work in both modes
  /** Weather data from Open-Meteo (has CORS support) */
  weatherEnrichment: true,
  /** All chart visualizations */
  charts: true,
  /** Map rendering */
  maps: true,
  /** Dashboard widgets */
  dashboard: true,
} as const

export type FeatureFlags = typeof features
