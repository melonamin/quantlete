import { useState } from 'react'
import { useAuthStatus, useCredentialsStatus, useUpdateCredentials } from '@/lib/data/hooks'
import { isWasmMode, isDemoMode } from '@/lib/mode'
import { getAuthUrl as getWasmAuthUrl } from '@/lib/wasm/strava/client'
import { useOnboardingStore } from '@/stores'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Lock,
  HardDrive,
  BarChart3,
  Clock,
  ExternalLink,
  Eye,
  EyeOff,
  Loader2,
  ArrowRight,
  Key,
  Server,
} from 'lucide-react'
import { shouldShowOnboarding } from './utils'

type OnboardingStep = 'welcome' | 'credentials' | 'connect'

function getAuthUrl(credentialsConfigured: boolean, source: string): string {
  // In server mode with env credentials, use server OAuth endpoint
  if (!isWasmMode() && source === 'env') {
    return '/api/v1/auth/strava'
  }

  // In WASM mode or server mode with database credentials, use WASM auth
  if (credentialsConfigured) {
    try {
      return getWasmAuthUrl(`${window.location.origin}/oauth/callback`)
    } catch (error) {
      console.error('Failed to generate WASM auth URL', error)
      return '#'
    }
  }

  return '#'
}

export function WelcomeModal() {
  const { data: auth, isLoading: authLoading } = useAuthStatus()
  const { data: credentials, isLoading: credentialsLoading } = useCredentialsStatus()
  const updateCredentials = useUpdateCredentials()
  const { dismissed, dismiss } = useOnboardingStore()

  // Multi-step flow: welcome -> credentials (if not configured) -> connect
  const [step, setStep] = useState<OnboardingStep>('welcome')

  // Credentials form state
  const [clientId, setClientId] = useState('')
  const [clientSecret, setClientSecret] = useState('')
  const [showSecret, setShowSecret] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  // Skip onboarding entirely in demo mode - the bundled database is pre-authenticated
  if (isDemoMode()) {
    return null
  }

  // Determine if we should show the modal
  const isAuthenticated = auth?.authenticated ?? false
  const credentialsConfigured = credentials?.configured ?? false
  const show = shouldShowOnboarding(
    isAuthenticated,
    authLoading,
    dismissed,
    credentialsConfigured,
    credentialsLoading
  )

  // Derive effective step - auto-advance from credentials if they're configured
  const effectiveStep = credentialsConfigured && step === 'credentials' ? 'connect' : step

  const handleContinue = () => {
    if (credentialsConfigured) {
      setStep('connect')
    } else {
      setStep('credentials')
    }
  }

  const handleSaveCredentials = async () => {
    if (!clientId.trim()) {
      setSaveError('Client ID is required')
      return
    }
    if (!clientSecret.trim()) {
      setSaveError('Client Secret is required')
      return
    }

    setSaveError(null)

    try {
      await updateCredentials.mutateAsync({
        client_id: clientId.trim(),
        client_secret: clientSecret.trim(),
      })
      setStep('connect')
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : 'Failed to save credentials')
    }
  }

  // Get auth URL based on mode and credentials source
  const authUrl = getAuthUrl(credentialsConfigured, credentials?.source ?? '')

  const wasmMode = isWasmMode()

  return (
    <Dialog open={show} onOpenChange={() => {}}>
      <DialogContent showCloseButton={false} className="sm:max-w-md">
        {effectiveStep === 'welcome' && (
          <>
            <DialogHeader>
              <DialogTitle className="text-2xl">Welcome to Quantlete</DialogTitle>
              <DialogDescription>Your private Strava analytics dashboard</DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              {wasmMode ? (
                <>
                  <div className="flex gap-3">
                    <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-green-100 dark:bg-green-950">
                      <Lock className="h-4 w-4 text-green-600 dark:text-green-400" />
                    </div>
                    <div>
                      <p className="font-medium">100% Private</p>
                      <p className="text-sm text-muted-foreground">
                        All data stays in your browser. Nothing is sent to any server.
                      </p>
                    </div>
                  </div>

                  <div className="flex gap-3">
                    <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-950">
                      <HardDrive className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                    </div>
                    <div>
                      <p className="font-medium">Stored Locally</p>
                      <p className="text-sm text-muted-foreground">
                        Data is tied to this device and browser using local storage.
                      </p>
                    </div>
                  </div>
                </>
              ) : (
                <>
                  <div className="flex gap-3">
                    <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-green-100 dark:bg-green-950">
                      <Server className="h-4 w-4 text-green-600 dark:text-green-400" />
                    </div>
                    <div>
                      <p className="font-medium">Self-Hosted</p>
                      <p className="text-sm text-muted-foreground">
                        Your data stays on your own server, fully under your control.
                      </p>
                    </div>
                  </div>

                  <div className="flex gap-3">
                    <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-950">
                      <Lock className="h-4 w-4 text-blue-600 dark:text-blue-400" />
                    </div>
                    <div>
                      <p className="font-medium">Private & Secure</p>
                      <p className="text-sm text-muted-foreground">
                        No third-party access to your activity data.
                      </p>
                    </div>
                  </div>
                </>
              )}

              <div className="flex gap-3">
                <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-purple-100 dark:bg-purple-950">
                  <BarChart3 className="h-4 w-4 text-purple-600 dark:text-purple-400" />
                </div>
                <div>
                  <p className="font-medium">Deep Analytics</p>
                  <p className="text-sm text-muted-foreground">
                    Training load, power curves, segments, and more.
                  </p>
                </div>
              </div>

              <div className="rounded-lg border border-border bg-muted/50 p-3">
                <div className="flex gap-2">
                  <Clock className="h-4 w-4 shrink-0 text-muted-foreground mt-0.5" />
                  <div className="text-sm text-muted-foreground">
                    <p className="font-medium text-foreground">First sync takes time</p>
                    <p>
                      Due to Strava rate limits, importing all your activities may take a while.
                    </p>
                  </div>
                </div>
              </div>
            </div>

            <DialogFooter className="flex-col gap-2 sm:flex-col">
              <Button className="w-full" onClick={handleContinue}>
                Get Started
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
              <Button variant="ghost" className="w-full" onClick={dismiss}>
                Maybe later
              </Button>
            </DialogFooter>
          </>
        )}

        {effectiveStep === 'credentials' && (
          <>
            <DialogHeader>
              <DialogTitle className="text-2xl">Configure Strava App</DialogTitle>
              <DialogDescription>
                Create your own Strava app to connect your account
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <Alert>
                <Key className="h-4 w-4" />
                <AlertDescription className="text-sm ml-2">
                  <p className="font-medium mb-2">Why do I need my own app?</p>
                  <p className="text-muted-foreground">
                    Strava limits API access to one athlete per app. Your credentials stay private
                    {wasmMode ? ' in your browser' : ' on your server'}.
                  </p>
                </AlertDescription>
              </Alert>

              <ol className="list-decimal list-inside space-y-2 text-sm">
                <li>
                  Go to{' '}
                  <a
                    href="https://www.strava.com/settings/api"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-primary hover:underline inline-flex items-center gap-1"
                  >
                    Strava API Settings
                    <ExternalLink className="h-3 w-3" />
                  </a>
                </li>
                <li>Create a new app (any name works)</li>
                <li>
                  Set callback domain to:{' '}
                  <code className="text-xs bg-muted px-1 py-0.5 rounded">
                    {window.location.host}
                  </code>
                </li>
                <li>Copy your credentials below</li>
              </ol>

              {saveError && (
                <Alert variant="destructive">
                  <AlertDescription>{saveError}</AlertDescription>
                </Alert>
              )}

              <div className="space-y-3">
                <div>
                  <label htmlFor="onboard-client-id" className="block text-sm font-medium mb-1">
                    Client ID
                  </label>
                  <Input
                    id="onboard-client-id"
                    type="text"
                    value={clientId}
                    onChange={(e) => setClientId(e.target.value)}
                    placeholder="e.g., 123456"
                  />
                </div>

                <div>
                  <label htmlFor="onboard-client-secret" className="block text-sm font-medium mb-1">
                    Client Secret
                  </label>
                  <div className="relative">
                    <Input
                      id="onboard-client-secret"
                      type={showSecret ? 'text' : 'password'}
                      value={clientSecret}
                      onChange={(e) => setClientSecret(e.target.value)}
                      placeholder="e.g., abc123..."
                      className="pr-10"
                    />
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      onClick={() => setShowSecret(!showSecret)}
                      className="absolute right-1 top-1/2 -translate-y-1/2"
                    >
                      {showSecret ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </Button>
                  </div>
                </div>
              </div>
            </div>

            <DialogFooter className="flex-col gap-2 sm:flex-col">
              <Button
                className="w-full"
                onClick={handleSaveCredentials}
                disabled={updateCredentials.isPending}
              >
                {updateCredentials.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    Save & Continue
                    <ArrowRight className="ml-2 h-4 w-4" />
                  </>
                )}
              </Button>
              <Button variant="ghost" className="w-full" onClick={() => setStep('welcome')}>
                Back
              </Button>
            </DialogFooter>
          </>
        )}

        {effectiveStep === 'connect' && (
          <>
            <DialogHeader>
              <DialogTitle className="text-2xl">Connect to Strava</DialogTitle>
              <DialogDescription>Authorize Quantlete to access your activities</DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-4">
              <div className="flex gap-3">
                <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-green-100 dark:bg-green-950">
                  <Lock className="h-4 w-4 text-green-600 dark:text-green-400" />
                </div>
                <div>
                  <p className="font-medium">Read-only access</p>
                  <p className="text-sm text-muted-foreground">
                    Quantlete only reads your activities. It cannot modify anything on Strava.
                  </p>
                </div>
              </div>

              <div className="rounded-lg border border-border bg-muted/50 p-3">
                <p className="text-sm text-muted-foreground">
                  You&apos;ll be redirected to Strava to authorize access. After approval,
                  you&apos;ll return here to start importing your activities.
                </p>
              </div>
            </div>

            <DialogFooter className="flex-col gap-2 sm:flex-col">
              <Button asChild className="w-full bg-[#FC4C02] hover:bg-[#FC4C02]/90">
                <a href={authUrl}>Connect with Strava</a>
              </Button>
              <Button variant="ghost" className="w-full" onClick={dismiss}>
                Maybe later
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
