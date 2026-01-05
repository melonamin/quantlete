/**
 * Check if we should show the onboarding modal.
 *
 * Shows onboarding when credentials are not configured (both modes)
 * or when not authenticated (both modes).
 *
 * The onboarding flow helps users:
 * 1. Configure Strava API credentials (required for both modes)
 * 2. Connect to Strava via OAuth
 */
export function shouldShowOnboarding(
  isAuthenticated: boolean,
  isAuthLoading: boolean,
  dismissed: boolean,
  credentialsConfigured?: boolean,
  isCredentialsLoading?: boolean
): boolean {
  // Wait for all loading to complete before deciding
  if (isAuthLoading) return false
  if (isCredentialsLoading) return false

  // Don't show if user dismissed
  if (dismissed) return false

  // Already authenticated, no need for onboarding
  if (isAuthenticated) return false

  // Don't show on OAuth callback page (let OAuth complete)
  if (typeof window !== 'undefined' && window.location.pathname === '/oauth/callback') return false

  // If credentialsConfigured is undefined, query hasn't completed yet
  // (either still loading or query disabled while provider initializes)
  if (credentialsConfigured === undefined) return false

  // Show if credentials not configured
  if (credentialsConfigured === false) return true

  // If not authenticated and credentials are configured, still show
  // so user can connect to Strava
  if (credentialsConfigured === true && !isAuthenticated) return true

  // Default: don't show while loading
  return false
}
