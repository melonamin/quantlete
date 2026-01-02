/**
 * OAuth callback page - handles Strava OAuth redirect.
 *
 * This page receives the authorization code from Strava and exchanges it
 * for access tokens via the worker proxy.
 */

import { useEffect, useState, useRef } from 'react'
import { useNavigate, useSearch } from '@tanstack/react-router'
import { useQueryClient } from '@tanstack/react-query'
import { useStartImport } from '@/lib/api'
import { useDataProviderStatus } from '@/lib/data'
import { isWasmMode } from '@/lib/mode'
import { exchangeCode } from '@/lib/wasm/strava/client'
import { useSyncModalStore, markOnboardingComplete } from '@/stores'
import { Loader2 } from 'lucide-react'

interface OAuthSearchParams {
  code?: string
  scope?: string
  error?: string
}

export function OAuthCallbackPage() {
  const navigate = useNavigate()
  const search = useSearch({ strict: false }) as OAuthSearchParams
  const queryClient = useQueryClient()
  const { provider, initialized, error: providerError } = useDataProviderStatus()
  const startImport = useStartImport()
  const openSyncModal = useSyncModalStore((s) => s.openModal)
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>('processing')
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [debugInfo, setDebugInfo] = useState<string>('Initializing...')

  // Track if we've already processed this callback to prevent double execution
  const processedRef = useRef(false)

  // Store callbacks in refs to avoid stale closures and prevent effect re-runs
  const startImportRef = useRef(startImport)
  const openSyncModalRef = useRef(openSyncModal)
  const navigateRef = useRef(navigate)

  // Keep refs up to date (must be in effect, not during render)
  useEffect(() => {
    startImportRef.current = startImport
    openSyncModalRef.current = openSyncModal
    navigateRef.current = navigate
  })

  useEffect(() => {
    async function handleCallback() {
      // Prevent double execution - check before any async work
      if (processedRef.current) return

      setDebugInfo(`initialized=${initialized}, provider=${!!provider}, error=${providerError || 'none'}`)

      // Check for provider initialization error
      if (providerError) {
        processedRef.current = true
        setStatus('error')
        setErrorMessage(`Provider error: ${providerError}`)
        return
      }

      // Wait for provider to initialize - don't mark as processed yet
      if (!initialized || !provider) {
        setDebugInfo(`Waiting for provider... initialized=${initialized}`)
        return
      }

      // Check for OAuth errors from Strava
      if (search.error) {
        processedRef.current = true
        setStatus('error')
        setErrorMessage(search.error)
        return
      }

      // Check for authorization code
      if (!search.code) {
        processedRef.current = true
        setStatus('error')
        setErrorMessage('No authorization code received')
        return
      }

      // Only handle OAuth in WASM mode
      if (!isWasmMode()) {
        processedRef.current = true
        setStatus('error')
        setErrorMessage('OAuth callback only works in WASM mode')
        return
      }

      // Mark as processed BEFORE starting async work to prevent re-entry
      // If async fails, the error state will be shown but we won't retry automatically
      processedRef.current = true

      try {
        setDebugInfo('Exchanging code for tokens...')

        // Exchange code for tokens via Strava client
        const redirectUri = `${window.location.origin}/oauth/callback`
        await exchangeCode(search.code, redirectUri)

        // Invalidate auth-related caches immediately to ensure all components
        // see the updated auth state (fixes race condition with Settings page)
        queryClient.invalidateQueries({ queryKey: ['data', 'auth', 'status'] })
        queryClient.invalidateQueries({ queryKey: ['data', 'setup', 'credentials'] })

        setStatus('success')

        // Mark onboarding as complete after successful auth
        markOnboardingComplete()

        // Start the initial sync
        setDebugInfo('Starting initial sync...')
        try {
          await startImportRef.current.mutateAsync({})
        } catch {
          // Don't fail the auth flow if import fails to start
        }

        // Redirect to dashboard and open sync modal after a short delay
        setTimeout(() => {
          // Open sync modal after navigation
          openSyncModalRef.current()
          navigateRef.current({ to: '/' })
        }, 1500)
      } catch (err) {
        console.error('OAuth callback error:', err)
        setStatus('error')
        setErrorMessage(err instanceof Error ? err.message : 'Failed to complete authentication')
      }
    }

    handleCallback()
    // Callbacks are stored in refs to avoid stale closures and prevent effect re-runs
    // queryClient is stable but included for exhaustive deps
  }, [initialized, provider, providerError, search.code, search.error, queryClient])

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        {status === 'processing' && (
          <>
            <Loader2 className="mx-auto h-8 w-8 animate-spin text-primary" />
            <p className="mt-4 text-muted-foreground">Completing authentication...</p>
            <p className="mt-2 text-xs text-muted-foreground/50">{debugInfo}</p>
          </>
        )}

        {status === 'success' && (
          <>
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <svg
                className="h-8 w-8 text-green-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <p className="mt-4 text-lg font-medium">Authentication successful!</p>
            <p className="text-muted-foreground">Starting sync and redirecting...</p>
          </>
        )}

        {status === 'error' && (
          <>
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-red-100">
              <svg
                className="h-8 w-8 text-red-600"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <p className="mt-4 text-lg font-medium">Authentication failed</p>
            <p className="text-muted-foreground">{errorMessage}</p>
            <button
              onClick={() => navigateRef.current({ to: '/' })}
              className="mt-4 rounded-md bg-primary px-4 py-2 text-primary-foreground hover:bg-primary/90"
            >
              Back to Home
            </button>
          </>
        )}
      </div>
    </div>
  )
}
