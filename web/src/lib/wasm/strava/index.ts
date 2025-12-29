/**
 * Browser Strava Integration
 *
 * Provides Strava API access via a Cloudflare Worker proxy.
 */

export {
  getAuthUrl,
  exchangeCode,
  refreshToken,
  getAccessToken,
  isAuthenticated,
  getAthlete,
  stravaFetch,
  loadAuth,
  clearAuth,
  type StravaToken,
  type StravaAthlete,
  type StravaAuthResponse,
} from './client'

export {
  startImport,
  cancelImport,
  getImportProgress,
  getSyncHistory,
  getLatestSync,
  type ImportOptions,
  type ImportProgress,
  type SyncRun,
} from './importer'
