/**
 * Strava Credentials Storage - stores client_id and client_secret via Go WASM.
 *
 * For WASM mode, each user must create their own Strava app and configure
 * their credentials here. The credentials are stored in the app_state table
 * through the Go WASM bridge for consistency with native mode.
 */

import * as goStorage from '../go-storage'

export interface StravaCredentials {
  clientId: string
  clientSecret: string
}

// In-memory cache for fast access
let cachedCredentials: StravaCredentials | null = null

/**
 * Get stored Strava credentials from the database.
 * Returns null if credentials are not configured.
 */
export function getCredentials(): StravaCredentials | null {
  if (cachedCredentials) {
    return cachedCredentials
  }

  if (!goStorage.isInitialized()) {
    return null
  }

  const result = goStorage.getStravaCredentials()
  if (result) {
    cachedCredentials = {
      clientId: result.client_id,
      clientSecret: result.client_secret,
    }
    return cachedCredentials
  }

  return null
}

/**
 * Save Strava credentials to the database.
 */
export async function saveCredentials(clientId: string, clientSecret: string): Promise<void> {
  if (!goStorage.isInitialized()) {
    throw new Error('Go WASM storage not initialized')
  }

  goStorage.saveStravaCredentials(clientId, clientSecret)

  // Persist to OPFS
  await goStorage.persistDatabase()

  // Update cache
  cachedCredentials = { clientId, clientSecret }
}

/**
 * Check if credentials are configured.
 */
export function hasCredentials(): boolean {
  if (!goStorage.isInitialized()) {
    return false
  }
  return goStorage.hasStravaCredentials()
}

/**
 * Clear cached credentials (useful after logout or credential change).
 */
export function clearCredentialsCache(): void {
  cachedCredentials = null
}

/**
 * Delete stored credentials from the database.
 */
export async function deleteCredentials(): Promise<void> {
  if (!goStorage.isInitialized()) {
    return
  }

  goStorage.deleteStravaCredentials()

  await goStorage.persistDatabase()
  cachedCredentials = null
}
