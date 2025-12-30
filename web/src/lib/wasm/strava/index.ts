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
  getRateLimitInfo,
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

export {
  getCredentials,
  saveCredentials,
  hasCredentials,
  clearCredentialsCache,
  deleteCredentials,
  type StravaCredentials,
} from './credentials'
