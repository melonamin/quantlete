/**
 * Setup/configuration API types.
 */

/**
 * Response for credentials status endpoint.
 */
export interface CredentialsStatus {
  configured: boolean
  client_id?: string // Masked, e.g., "1234****"
  source: 'env' | 'database' | 'browser' // Where credentials come from
  redirect_uri: string
}

/**
 * Request to update credentials.
 */
export interface UpdateCredentialsRequest {
  client_id: string
  client_secret: string
}
