import { useAuthStatus, useImportProgress, useStartImport, useCancelImport } from '@/lib/api'
import { useQueryClient } from '@tanstack/react-query'
import { activityKeys } from '@/lib/api/activities'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Check, X, AlertCircle } from 'lucide-react'

export function SettingsPage() {
  const queryClient = useQueryClient()
  const { data: authStatus } = useAuthStatus()
  const { data: progress } = useImportProgress()
  const startImport = useStartImport()
  const cancelImport = useCancelImport()

  const isAuthenticated = authStatus?.authenticated
  const isImporting = progress?.status === 'running'

  const handleStartImport = () => {
    startImport.mutate({})
  }

  const handleCancelImport = () => {
    cancelImport.mutate()
  }

  const handleImportComplete = () => {
    queryClient.invalidateQueries({ queryKey: activityKeys.all })
  }

  // Invalidate activities when import completes
  if (progress?.status === 'completed' && startImport.isSuccess) {
    handleImportComplete()
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Configure your preferences</p>
      </div>

      <div className="space-y-6 max-w-2xl">
        {/* Strava Connection */}
        <Card>
          <CardHeader>
            <CardTitle>Strava Connection</CardTitle>
            <CardDescription>Connect your Strava account to import activities</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                {isAuthenticated ? (
                  <>
                    <div className="h-8 w-8 rounded-full bg-green-100 flex items-center justify-center">
                      <Check className="h-4 w-4 text-green-600" />
                    </div>
                    <div>
                      <p className="font-medium">Connected</p>
                      {authStatus?.athlete && (
                        <p className="text-sm text-muted-foreground">
                          {authStatus.athlete.firstname} {authStatus.athlete.lastname}
                        </p>
                      )}
                    </div>
                  </>
                ) : (
                  <>
                    <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center">
                      <X className="h-4 w-4 text-muted-foreground" />
                    </div>
                    <div>
                      <p className="font-medium">Not connected</p>
                      <p className="text-sm text-muted-foreground">
                        Connect to import your activities
                      </p>
                    </div>
                  </>
                )}
              </div>
              {!isAuthenticated && (
                <Button asChild className="bg-strava hover:bg-strava/90">
                  <a href="/api/v1/auth/strava">Connect Strava</a>
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Data Import */}
        <Card>
          <CardHeader>
            <CardTitle>Data Import</CardTitle>
            <CardDescription>Import activities from Strava</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {!isAuthenticated ? (
              <p className="text-sm text-muted-foreground">
                Connect your Strava account first to import activities.
              </p>
            ) : (
              <>
                {progress && progress.status !== 'idle' && (
                  <ImportStatus progress={progress} />
                )}

                <div className="flex gap-2">
                  <Button
                    onClick={handleStartImport}
                    disabled={isImporting || !isAuthenticated}
                  >
                    {isImporting ? (
                      <>
                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        Importing...
                      </>
                    ) : (
                      'Start Import'
                    )}
                  </Button>
                  {isImporting && (
                    <Button variant="outline" onClick={handleCancelImport}>
                      Cancel
                    </Button>
                  )}
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Display Settings */}
        <Card>
          <CardHeader>
            <CardTitle>Display</CardTitle>
            <CardDescription>Customize how data is displayed</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Unit System</p>
                <p className="text-sm text-muted-foreground">Choose metric or imperial units</p>
              </div>
              <span className="text-sm text-muted-foreground">Metric</span>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

interface ImportStatusProps {
  progress: {
    status: string
    total_activities: number
    imported_count: number
    failed_count: number
    current_page: number
    error?: string
  }
}

function ImportStatus({ progress }: ImportStatusProps) {
  const percentage = progress.total_activities > 0
    ? Math.round((progress.imported_count / progress.total_activities) * 100)
    : 0

  return (
    <div className="rounded-lg border border-border bg-muted/50 p-4">
      <div className="flex items-center gap-2 mb-2">
        {progress.status === 'running' && (
          <Loader2 className="h-4 w-4 animate-spin text-primary" />
        )}
        {progress.status === 'completed' && (
          <Check className="h-4 w-4 text-green-600" />
        )}
        {progress.status === 'failed' && (
          <AlertCircle className="h-4 w-4 text-destructive" />
        )}
        <span className="font-medium capitalize">{progress.status}</span>
      </div>

      <div className="space-y-1 text-sm text-muted-foreground">
        <p>
          Imported: {progress.imported_count} / {progress.total_activities} activities
          {progress.total_activities > 0 && ` (${percentage}%)`}
        </p>
        {progress.failed_count > 0 && (
          <p className="text-destructive">Failed: {progress.failed_count}</p>
        )}
        {progress.status === 'running' && (
          <p>Page: {progress.current_page}</p>
        )}
        {progress.error && (
          <p className="text-destructive">{progress.error}</p>
        )}
      </div>

      {progress.status === 'running' && progress.total_activities > 0 && (
        <div className="mt-3 h-2 rounded-full bg-muted overflow-hidden">
          <div
            className="h-full bg-primary transition-all duration-300"
            style={{ width: `${percentage}%` }}
          />
        </div>
      )}
    </div>
  )
}
