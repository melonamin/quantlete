/**
 * Browser Strava Client - Communicates with Strava API via Worker proxy.
 *
 * This client handles:
 * - OAuth flow (redirect to Strava, exchange code via worker)
 * - Token management (storage, refresh)
 * - API calls (via worker proxy)
 *
 * Credentials (client_id, client_secret) are stored locally in the database
 * and sent to the worker proxy for OAuth operations.
 */

import { getDatabase } from '../db'
import * as rateLimit from './ratelimit'
import { getCredentials } from './credentials'

// Remove trailing slash from worker URL to prevent double-slash in paths
const WORKER_URL = (import.meta.env.VITE_CF_WORKER_URL || import.meta.env.VITE_STRAVA_WORKER_URL || '').replace(/\/$/, '')
const STRAVA_AUTH_URL = 'https://www.strava.com/oauth/authorize'

export interface StravaToken {
  access_token: string
  refresh_token: string
  expires_at: number
  token_type: string
}

export interface StravaAthlete {
  id: number
  username: string
  firstname: string
  lastname: string
  profile_medium: string
  profile: string
}

export interface StravaAuthResponse {
  token_type: string
  expires_at: number
  expires_in: number
  refresh_token: string
  access_token: string
  athlete: StravaAthlete
}

// In-memory token cache
let currentToken: StravaToken | null = null
let currentAthlete: StravaAthlete | null = null

/**
 * Get the OAuth authorization URL to redirect the user to Strava.
 * Requires credentials to be configured first.
 */
export function getAuthUrl(redirectUri: string): string {
  const credentials = getCredentials()
  if (!credentials) {
    throw new Error('Strava credentials not configured. Please configure your Strava app first.')
  }

  const params = new URLSearchParams({
    client_id: credentials.clientId,
    redirect_uri: redirectUri,
    response_type: 'code',
    scope: 'read,activity:read_all,profile:read_all',
  })

  return `${STRAVA_AUTH_URL}?${params.toString()}`
}

/**
 * Exchange an authorization code for tokens via the worker.
 * Sends credentials to the worker for the token exchange.
 */
export async function exchangeCode(
  code: string,
  redirectUri: string
): Promise<StravaAuthResponse> {
  const credentials = getCredentials()
  if (!credentials) {
    throw new Error('Strava credentials not configured. Please configure your Strava app first.')
  }

  const response = await fetch(`${WORKER_URL}/oauth/exchange`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      code,
      redirect_uri: redirectUri,
      client_id: credentials.clientId,
      client_secret: credentials.clientSecret,
    }),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || 'Token exchange failed')
  }

  const data: StravaAuthResponse = await response.json()

  // Cache token and athlete
  currentToken = {
    access_token: data.access_token,
    refresh_token: data.refresh_token,
    expires_at: data.expires_at,
    token_type: data.token_type,
  }
  currentAthlete = data.athlete

  // Persist to database
  await persistAuth(data)

  return data
}

/**
 * Refresh the access token using the refresh token.
 * Sends credentials to the worker for the token refresh.
 */
export async function refreshToken(): Promise<StravaToken> {
  if (!currentToken?.refresh_token) {
    throw new Error('No refresh token available')
  }

  const credentials = getCredentials()
  if (!credentials) {
    throw new Error('Strava credentials not configured. Please configure your Strava app first.')
  }

  const response = await fetch(`${WORKER_URL}/oauth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      refresh_token: currentToken.refresh_token,
      client_id: credentials.clientId,
      client_secret: credentials.clientSecret,
    }),
  })

  if (!response.ok) {
    const error = await response.json()
    throw new Error(error.error || 'Token refresh failed')
  }

  const data = await response.json()

  currentToken = {
    access_token: data.access_token,
    refresh_token: data.refresh_token,
    expires_at: data.expires_at,
    token_type: data.token_type,
  }

  // Update database
  await updateToken(currentToken)

  return currentToken
}

/**
 * Get a valid access token, refreshing if necessary.
 */
export async function getAccessToken(): Promise<string> {
  if (!currentToken) {
    throw new Error('Not authenticated')
  }

  // Refresh if token expires within 5 minutes
  const now = Math.floor(Date.now() / 1000)
  if (currentToken.expires_at < now + 300) {
    await refreshToken()
  }

  return currentToken.access_token
}

/**
 * Check if we have valid authentication.
 */
export function isAuthenticated(): boolean {
  return currentToken !== null
}

/**
 * Get the current athlete.
 */
export function getAthlete(): StravaAthlete | null {
  return currentAthlete
}

/**
 * Make an authenticated API call to Strava via the worker.
 * Handles rate limiting with automatic waiting and retry.
 */
export async function stravaFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  // Wait if we're approaching rate limit
  await rateLimit.wait()

  const token = await getAccessToken()

  const makeRequest = async (): Promise<Response> => {
    const response = await fetch(`${WORKER_URL}/api${path}`, {
      ...options,
      headers: {
        ...options.headers,
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    })

    // Update rate limit state from headers (if forwarded by worker)
    rateLimit.updateFromHeaders(response.headers)

    // Also increment usage locally as fallback
    rateLimit.incrementUsage()

    return response
  }

  let response = await makeRequest()

  // Handle 401 - try to refresh token
  if (response.status === 401) {
    await refreshToken()
    const retryToken = await getAccessToken()

    response = await fetch(`${WORKER_URL}/api${path}`, {
      ...options,
      headers: {
        ...options.headers,
        Authorization: `Bearer ${retryToken}`,
        'Content-Type': 'application/json',
      },
    })

    rateLimit.updateFromHeaders(response.headers)
    rateLimit.incrementUsage()

    if (!response.ok) {
      throw new Error(`Strava API error: ${response.status}`)
    }

    return response.json()
  }

  // Handle 429 - rate limit hit, wait and retry
  if (response.status === 429) {
    console.log('[Strava] 429 rate limit hit, waiting for reset...')
    await rateLimit.waitForRetry()

    // Retry the request
    response = await makeRequest()

    if (!response.ok) {
      throw new Error(`Strava API error after retry: ${response.status}`)
    }

    return response.json()
  }

  if (!response.ok) {
    throw new Error(`Strava API error: ${response.status}`)
  }

  return response.json()
}

/**
 * Load authentication from database on startup.
 */
export async function loadAuth(): Promise<boolean> {
  try {
    const db = getDatabase()
    if (!db.isInitialized()) {
      return false
    }

    // Load rate limit state
    rateLimit.loadState()

    // Load token
    const tokenRow = db.queryOne<{
      access_token: string
      refresh_token: string
      expires_at: number
      token_type: string
    }>(`SELECT access_token, refresh_token, expires_at, token_type
        FROM auth_tokens LIMIT 1`)

    if (!tokenRow) {
      return false
    }

    currentToken = {
      access_token: tokenRow.access_token,
      refresh_token: tokenRow.refresh_token,
      expires_at: tokenRow.expires_at,
      token_type: tokenRow.token_type,
    }

    // Load athlete
    const athleteRow = db.queryOne<{
      id: number
      username: string
      firstname: string
      lastname: string
      profile_medium: string
      profile: string
    }>(`SELECT id, username, firstname, lastname, profile_medium, profile
        FROM athletes LIMIT 1`)

    if (athleteRow) {
      currentAthlete = athleteRow
    }

    return true
  } catch {
    return false
  }
}

/**
 * Get current rate limit info for UI display.
 */
export function getRateLimitInfo() {
  return rateLimit.getRateLimitInfo()
}

/**
 * Clear authentication (logout).
 */
export async function clearAuth(): Promise<void> {
  currentToken = null
  currentAthlete = null

  try {
    const db = getDatabase()
    db.exec('DELETE FROM auth_tokens')
    db.exec('DELETE FROM athletes')
    await db.persist()
  } catch {
    // Ignore database errors during logout
  }
}

/**
 * Persist authentication to database.
 */
async function persistAuth(data: StravaAuthResponse): Promise<void> {
  const db = getDatabase()

  // Upsert athlete
  db.exec(
    `INSERT INTO athletes (id, username, firstname, lastname, profile_medium, profile, updated_at)
     VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
     ON CONFLICT (id) DO UPDATE SET
       username = EXCLUDED.username,
       firstname = EXCLUDED.firstname,
       lastname = EXCLUDED.lastname,
       profile_medium = EXCLUDED.profile_medium,
       profile = EXCLUDED.profile,
       updated_at = EXCLUDED.updated_at`,
    [
      data.athlete.id,
      data.athlete.username,
      data.athlete.firstname,
      data.athlete.lastname,
      data.athlete.profile_medium,
      data.athlete.profile,
    ]
  )

  // Upsert token
  db.exec(
    `INSERT INTO auth_tokens (athlete_id, access_token, refresh_token, token_type, expires_at)
     VALUES (?, ?, ?, ?, ?)
     ON CONFLICT (athlete_id) DO UPDATE SET
       access_token = EXCLUDED.access_token,
       refresh_token = EXCLUDED.refresh_token,
       token_type = EXCLUDED.token_type,
       expires_at = EXCLUDED.expires_at`,
    [
      data.athlete.id,
      data.access_token,
      data.refresh_token,
      data.token_type,
      data.expires_at,
    ]
  )

  await db.persist()
}

/**
 * Update token in database.
 */
async function updateToken(token: StravaToken): Promise<void> {
  const db = getDatabase()

  db.exec(
    `UPDATE auth_tokens SET
       access_token = ?,
       refresh_token = ?,
       expires_at = ?`,
    [token.access_token, token.refresh_token, token.expires_at]
  )

  await db.persist()
}
