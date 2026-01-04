/**
 * Strava Credentials Storage - stores client_id and client_secret in the database.
 *
 * For WASM mode, each user must create their own Strava app and configure
 * their credentials here. The credentials are stored in the app_state table.
 */

import { getDatabase } from '../db'

const KEY_CLIENT_ID = 'strava_client_id'
const KEY_CLIENT_SECRET = 'strava_client_secret'

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

  const db = getDatabase()
  if (!db.isInitialized()) {
    return null
  }

  const clientIdRow = db.queryOne<{ value: string }>('SELECT value FROM app_state WHERE key = ?', [
    KEY_CLIENT_ID,
  ])
  const clientSecretRow = db.queryOne<{ value: string }>(
    'SELECT value FROM app_state WHERE key = ?',
    [KEY_CLIENT_SECRET]
  )

  if (clientIdRow && clientSecretRow) {
    cachedCredentials = {
      clientId: clientIdRow.value,
      clientSecret: clientSecretRow.value,
    }
    return cachedCredentials
  }

  return null
}

/**
 * Save Strava credentials to the database.
 */
export async function saveCredentials(clientId: string, clientSecret: string): Promise<void> {
  const db = getDatabase()
  if (!db.isInitialized()) {
    throw new Error('Database not initialized')
  }

  const now = new Date().toISOString()

  // Upsert client_id
  db.exec(
    `INSERT INTO app_state (key, value, updated_at)
     VALUES (?, ?, ?)
     ON CONFLICT (key) DO UPDATE SET
       value = EXCLUDED.value,
       updated_at = EXCLUDED.updated_at`,
    [KEY_CLIENT_ID, clientId, now]
  )

  // Upsert client_secret
  db.exec(
    `INSERT INTO app_state (key, value, updated_at)
     VALUES (?, ?, ?)
     ON CONFLICT (key) DO UPDATE SET
       value = EXCLUDED.value,
       updated_at = EXCLUDED.updated_at`,
    [KEY_CLIENT_SECRET, clientSecret, now]
  )

  // Persist to storage
  await db.persist()

  // Update cache
  cachedCredentials = { clientId, clientSecret }
}

/**
 * Check if credentials are configured.
 */
export function hasCredentials(): boolean {
  return getCredentials() !== null
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
  const db = getDatabase()
  if (!db.isInitialized()) {
    return
  }

  db.exec('DELETE FROM app_state WHERE key IN (?, ?)', [KEY_CLIENT_ID, KEY_CLIENT_SECRET])

  await db.persist()
  cachedCredentials = null
}
