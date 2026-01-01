/**
 * Strava Credentials Form - allows users to configure their Strava app credentials.
 *
 * Each user must create their own Strava app at https://www.strava.com/settings/api
 * and enter their Client ID and Client Secret here.
 *
 * Works in both WASM mode (browser storage) and server mode (API storage).
 */

import { useState, useMemo } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Check, ExternalLink, Eye, EyeOff, Loader2, Settings } from 'lucide-react'
import { useCredentialsStatus, useUpdateCredentials } from '@/lib/data/hooks'

interface StravaCredentialsFormProps {
  onSave?: () => void
}

export function StravaCredentialsForm({ onSave }: StravaCredentialsFormProps) {
  const { data: credentials, isLoading } = useCredentialsStatus()
  const updateCredentials = useUpdateCredentials()

  const [clientSecret, setClientSecret] = useState('')
  const [showSecret, setShowSecret] = useState(false)
  const [isEditing, setIsEditing] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [clientIdInput, setClientIdInput] = useState('')

  // Derive clientId: use user input if editing, otherwise from credentials
  const credentialsClientId = useMemo(() => {
    if (credentials?.configured && credentials.client_id) {
      return credentials.client_id.replace(/\*+$/, '')
    }
    return ''
  }, [credentials])

  const clientId = isEditing ? clientIdInput : credentialsClientId
  const setClientId = (value: string) => setClientIdInput(value)

  // Start editing mode and initialize form with current credentials
  const startEditing = () => {
    setClientIdInput(credentialsClientId)
    setIsEditing(true)
  }

  const handleSave = async () => {
    if (!clientId.trim()) {
      setSaveError('Client ID is required')
      return
    }

    // If editing existing credentials, require secret only if not configured yet
    if (!clientSecret.trim() && !credentials?.configured) {
      setSaveError('Client Secret is required')
      return
    }

    // If updating existing credentials, we need both ID and secret
    if (!clientSecret.trim() && credentials?.configured) {
      setSaveError('Client Secret is required to update credentials')
      return
    }

    setSaveError(null)

    try {
      await updateCredentials.mutateAsync({
        client_id: clientId.trim(),
        client_secret: clientSecret.trim(),
      })

      setIsEditing(false)
      setClientSecret('')
      onSave?.()
    } catch (error) {
      setSaveError(error instanceof Error ? error.message : 'Failed to save credentials')
    }
  }

  const handleCancel = () => {
    setIsEditing(false)
    if (credentials?.configured && credentials.client_id) {
      const visiblePart = credentials.client_id.replace(/\*+$/, '')
      setClientId(visiblePart)
    } else {
      setClientId('')
    }
    setClientSecret('')
    setSaveError(null)
  }

  // Show loading state
  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Strava App Configuration</CardTitle>
          <CardDescription>Loading configuration...</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-2 text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            <span>Loading...</span>
          </div>
        </CardContent>
      </Card>
    )
  }

  // If credentials are from env vars in server mode, show read-only
  if (credentials?.configured && credentials.source === 'env') {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Strava App Configuration</CardTitle>
          <CardDescription>
            Your Strava API credentials for connecting to Strava
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center gap-3">
            <div className="h-8 w-8 rounded-full bg-green-100 flex items-center justify-center">
              <Settings className="h-4 w-4 text-green-600" />
            </div>
            <div>
              <p className="font-medium">Configured via environment</p>
              <p className="text-sm text-muted-foreground">
                Client ID: {credentials.client_id}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  // If credentials exist and we're not editing, show configured state
  if (credentials?.configured && !isEditing) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Strava App Configuration</CardTitle>
          <CardDescription>
            Your Strava API credentials for connecting to Strava
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="h-8 w-8 rounded-full bg-green-100 flex items-center justify-center">
                <Check className="h-4 w-4 text-green-600" />
              </div>
              <div>
                <p className="font-medium">Configured</p>
                <p className="text-sm text-muted-foreground">
                  Client ID: {credentials.client_id}
                </p>
              </div>
            </div>
            <Button variant="outline" onClick={startEditing}>
              Edit
            </Button>
          </div>
        </CardContent>
      </Card>
    )
  }

  // Show form for new credentials or editing
  return (
    <Card>
      <CardHeader>
        <CardTitle>Strava App Configuration</CardTitle>
        <CardDescription>
          Create your own Strava app to connect to your account. Each user needs their own app due to Strava API restrictions.
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        <Alert>
          <AlertDescription className="text-sm">
            <ol className="list-decimal list-inside space-y-1">
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
              <li>Create a new app (any name/website works)</li>
              <li>Set Authorization Callback Domain to: <code className="text-xs bg-muted px-1 py-0.5 rounded">{window.location.host}</code></li>
              <li>Copy your Client ID and Client Secret below</li>
            </ol>
          </AlertDescription>
        </Alert>

        {saveError && (
          <Alert variant="destructive">
            <AlertDescription>{saveError}</AlertDescription>
          </Alert>
        )}

        <div className="space-y-3">
          <div>
            <Label htmlFor="client-id" className="mb-1">
              Client ID
            </Label>
            <Input
              id="client-id"
              type="text"
              value={clientId}
              onChange={(e) => setClientId(e.target.value)}
              placeholder="e.g., 123456"
            />
          </div>

          <div>
            <Label htmlFor="client-secret" className="mb-1">
              Client Secret
            </Label>
            <div className="relative">
              <Input
                id="client-secret"
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

        <div className="flex gap-2">
          <Button onClick={handleSave} disabled={updateCredentials.isPending}>
            {updateCredentials.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              'Save Credentials'
            )}
          </Button>
          {credentials?.configured && (
            <Button variant="outline" onClick={handleCancel}>
              Cancel
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
