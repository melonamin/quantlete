import {
  useAppSettings,
  useAuthStatus,
  useCancelImport,
  useCredentialsStatus,
  useHasResumableImport,
  useImportProgress,
  useLatestSync,
  useStartImport,
  useUpdateAppSettings,
} from '@/lib/data/hooks'
import { useQueryClient } from '@tanstack/react-query'
import { useSearch } from '@tanstack/react-router'
import { getAuthUrl } from '@/lib/wasm/strava/client'
import { isServerMode } from '@/lib/mode'
import { features } from '@/lib/features'
import {
  EddingtonDefinitionsEditor,
  NotificationsEditor,
  SchedulerEditor,
  SectionHeader,
  SegmentedButtons,
  StravaCredentialsForm,
  VirtualWorldTilesEditor,
} from '@/components/settings'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Checkbox } from '@/components/ui/checkbox'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { ImportStatus } from '@/components/sync/import-status'
import { SyncHistoryModal } from '@/components/sync/sync-history-modal'
import {
  Loader2,
  Check,
  X,
  ChevronDown,
  ChevronRight,
  History,
  Clock,
  AlertCircle,
  Link2,
  RefreshCw,
  Palette,
  Wrench,
  Bell,
} from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { formatDistance } from 'date-fns'
import { useSettingsStore, type Theme, type UnitSystem } from '@/stores/settings'

interface SettingsSearchParams {
  auth_error?: string
}

export function SettingsPage() {
  const queryClient = useQueryClient()
  const search = useSearch({ strict: false }) as SettingsSearchParams
  const { data: authStatus } = useAuthStatus()
  const { data: credentials } = useCredentialsStatus()
  const { data: progress } = useImportProgress()
  const { data: latestSync } = useLatestSync()
  const { data: hasResumable } = useHasResumableImport()
  const startImport = useStartImport()
  const cancelImport = useCancelImport()

  const isAuthenticated = authStatus?.authenticated ?? false
  const isDemoMode = authStatus?.demo_mode ?? false
  const credentialsConfigured = credentials?.configured ?? false
  const isImporting = progress?.status === 'running'
  const isPaused = progress?.status === 'paused'
  const isIdle = progress?.status === 'idle' || !progress

  const [includeStreams, setIncludeStreams] = useState(true)
  const [includeSegments, setIncludeSegments] = useState(true)
  const [includeBestEfforts, setIncludeBestEfforts] = useState(true)
  const [includePhotos, setIncludePhotos] = useState(true)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [showHistoryModal, setShowHistoryModal] = useState(false)
  const [authError, setAuthError] = useState<string | null>(search.auth_error || null)

  const unitSystem = useSettingsStore((s) => s.unitSystem)
  const setUnitSystem = useSettingsStore((s) => s.setUnitSystem)
  const theme = useSettingsStore((s) => s.theme)
  const setTheme = useSettingsStore((s) => s.setTheme)

  const { data: appSettings } = useAppSettings({ enabled: !!isAuthenticated })
  const updateAppSettings = useUpdateAppSettings()

  const handleStartImport = () => {
    startImport.mutate({
      skip_streams: !includeStreams,
      skip_segments: !includeSegments,
      skip_best_efforts: !includeBestEfforts,
      skip_photos: !includePhotos,
    })
  }

  const handleResumeImport = () => {
    startImport.mutate({ resume: true })
  }

  const handleCancelImport = () => {
    cancelImport.mutate()
  }

  const prevImportStatusRef = useRef<string | undefined>(undefined)

  useEffect(() => {
    const currentStatus = progress?.status
    const prevStatus = prevImportStatusRef.current

    if (currentStatus === 'completed' && prevStatus === 'running') {
      queryClient.invalidateQueries({ queryKey: ['data', 'activities'] })
      queryClient.invalidateQueries({ queryKey: ['data', 'dashboard'] })
    }

    prevImportStatusRef.current = currentStatus
  }, [progress?.status, queryClient])

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted-foreground">Configure your app preferences</p>
      </div>

      <div className="space-y-10 max-w-2xl">
        {/* Auth Error Alert */}
        {authError && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertTitle>Authentication Failed</AlertTitle>
            <AlertDescription className="flex items-center justify-between">
              <span>{authError}</span>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setAuthError(null)}
                className="h-6 px-2"
              >
                Dismiss
              </Button>
            </AlertDescription>
          </Alert>
        )}

        {/* CONNECTION SECTION */}
        <section>
          <SectionHeader icon={Link2} title="Connection" />
          <div className="space-y-4">
            {/* Strava App Configuration - hidden in demo mode */}
            {!isDemoMode && <StravaCredentialsForm />}

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
                  {!isAuthenticated &&
                    (!credentialsConfigured ? (
                      <Button disabled className="bg-strava/50">
                        Configure credentials first
                      </Button>
                    ) : credentials?.source === 'env' ? (
                      <Button asChild className="bg-strava hover:bg-strava/90">
                        <a href="/api/v1/auth/strava">Connect Strava</a>
                      </Button>
                    ) : (
                      <Button asChild className="bg-strava hover:bg-strava/90">
                        <a href={getAuthUrl(`${window.location.origin}/oauth/callback`)}>
                          Connect Strava
                        </a>
                      </Button>
                    ))}
                </div>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* DATA SYNC SECTION - disabled in demo mode */}
        <section>
          <SectionHeader icon={RefreshCw} title="Data Sync" />
          <div className="space-y-4">
            {/* Demo mode notice */}
            {isDemoMode && (
              <Card className="border-dashed opacity-75">
                <CardContent className="pt-6">
                  <p className="text-sm text-muted-foreground">
                    Data sync is disabled in demo mode. Demo data is pre-generated and cannot be
                    modified.
                  </p>
                </CardContent>
              </Card>
            )}

            {/* Data Import */}
            {!isDemoMode && (
              <Card>
                <CardHeader>
                  <CardTitle>Import</CardTitle>
                  <CardDescription>
                    Import activities from Strava. Activities load first for immediate dashboard
                    access, then streams, segments, and photos.
                  </CardDescription>
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
                        {hasResumable && isIdle && (
                          <Button onClick={handleResumeImport} disabled={startImport.isPending}>
                            {startImport.isPending ? (
                              <>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                Resuming...
                              </>
                            ) : (
                              'Resume Previous Sync'
                            )}
                          </Button>
                        )}
                        <Button
                          onClick={handleStartImport}
                          disabled={isImporting || isPaused || !isAuthenticated}
                          variant={hasResumable && isIdle ? 'outline' : 'default'}
                        >
                          {isImporting ? (
                            <>
                              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                              Importing...
                            </>
                          ) : hasResumable && isIdle ? (
                            'Start New Import'
                          ) : (
                            'Start Import'
                          )}
                        </Button>
                        {(isImporting || isPaused) && (
                          <Button variant="outline" onClick={handleCancelImport}>
                            Cancel
                          </Button>
                        )}
                      </div>

                      <div className="rounded-md border border-border bg-muted/30">
                        <Button
                          variant="ghost"
                          className="flex w-full items-center justify-start gap-2 p-3 text-sm text-muted-foreground hover:text-foreground"
                          onClick={() => setShowAdvanced(!showAdvanced)}
                          disabled={isImporting}
                          aria-expanded={showAdvanced}
                          aria-controls="import-advanced-options"
                        >
                          {showAdvanced ? (
                            <ChevronDown className="h-4 w-4" />
                          ) : (
                            <ChevronRight className="h-4 w-4" />
                          )}
                          Advanced options
                        </Button>
                        {showAdvanced && (
                          <div
                            id="import-advanced-options"
                            className="space-y-2 px-3 pb-3 border-t border-border pt-3"
                          >
                            <p className="text-xs text-muted-foreground mb-2">
                              All data is imported by default. Uncheck to skip specific data types.
                            </p>
                            <label
                              htmlFor="include-streams"
                              className="flex items-center gap-2 text-sm"
                            >
                              <Checkbox
                                id="include-streams"
                                checked={includeStreams}
                                onCheckedChange={(checked) => setIncludeStreams(!!checked)}
                                disabled={isImporting}
                                aria-describedby="include-streams-desc"
                              />
                              <span id="include-streams-desc">
                                Streams (GPS, heartrate, power - for maps & charts)
                              </span>
                            </label>
                            <label
                              htmlFor="include-segments"
                              className="flex items-center gap-2 text-sm"
                            >
                              <Checkbox
                                id="include-segments"
                                checked={includeSegments}
                                onCheckedChange={(checked) => setIncludeSegments(!!checked)}
                                disabled={isImporting}
                                aria-describedby="include-segments-desc"
                              />
                              <span id="include-segments-desc">
                                Segments (segment efforts & leaderboards)
                              </span>
                            </label>
                            <label
                              htmlFor="include-best-efforts"
                              className="flex items-center gap-2 text-sm"
                            >
                              <Checkbox
                                id="include-best-efforts"
                                checked={includeBestEfforts}
                                onCheckedChange={(checked) => setIncludeBestEfforts(!!checked)}
                                disabled={isImporting}
                                aria-describedby="include-best-efforts-desc"
                              />
                              <span id="include-best-efforts-desc">
                                Best efforts (PRs for standard distances)
                              </span>
                            </label>
                            <label
                              htmlFor="include-photos"
                              className="flex items-center gap-2 text-sm"
                            >
                              <Checkbox
                                id="include-photos"
                                checked={includePhotos}
                                onCheckedChange={(checked) => setIncludePhotos(!!checked)}
                                disabled={isImporting}
                                aria-describedby="include-photos-desc"
                              />
                              <span id="include-photos-desc">Photos (activity photos)</span>
                            </label>
                          </div>
                        )}
                      </div>
                    </>
                  )}
                </CardContent>
              </Card>
            )}

            {/* Scheduler - Server mode only */}
            {isServerMode() && !isDemoMode && isAuthenticated && appSettings && (
              <Card>
                <CardHeader>
                  <CardTitle>Schedule</CardTitle>
                  <CardDescription>Automate pull and/or push updates</CardDescription>
                </CardHeader>
                <CardContent>
                  <SchedulerEditor
                    key={`${appSettings.scheduler.pull.enabled}-${appSettings.scheduler.pull.schedule}-${appSettings.scheduler.push.enabled}`}
                    settings={appSettings}
                    onSave={(s) => updateAppSettings.mutate(s)}
                    saving={updateAppSettings.isPending}
                  />
                </CardContent>
              </Card>
            )}

            {/* Last Sync Status */}
            {!isDemoMode && isAuthenticated && latestSync && (
              <Card>
                <CardHeader className="flex-row items-center justify-between space-y-0">
                  <div>
                    <CardTitle>Last Sync</CardTitle>
                    <CardDescription>Your most recent data synchronization</CardDescription>
                  </div>
                  <Button variant="outline" size="sm" onClick={() => setShowHistoryModal(true)}>
                    <History className="h-4 w-4 mr-2" />
                    History
                  </Button>
                </CardHeader>
                <CardContent>
                  <div className="flex items-start gap-4">
                    <div
                      className="flex items-center justify-center h-10 w-10 rounded-full shrink-0"
                      style={{
                        backgroundColor:
                          latestSync.status === 'completed'
                            ? 'var(--green-100, #dcfce7)'
                            : latestSync.status === 'failed'
                              ? 'var(--red-100, #fee2e2)'
                              : latestSync.status === 'running'
                                ? 'var(--blue-100, #dbeafe)'
                                : 'var(--yellow-100, #fef3c7)',
                      }}
                    >
                      {latestSync.status === 'completed' ? (
                        <Check className="h-5 w-5 text-green-600" />
                      ) : latestSync.status === 'failed' ? (
                        <X className="h-5 w-5 text-red-600" />
                      ) : latestSync.status === 'running' ? (
                        <Clock className="h-5 w-5 text-blue-600 animate-pulse" />
                      ) : (
                        <AlertCircle className="h-5 w-5 text-yellow-600" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium capitalize">{latestSync.status}</span>
                        {latestSync.full_sync && (
                          <span className="text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                            Full
                          </span>
                        )}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {formatDistance(new Date(latestSync.started_at), new Date(), {
                          addSuffix: true,
                        })}
                        {latestSync.duration_seconds && latestSync.status !== 'running' && (
                          <>
                            {' '}
                            •{' '}
                            {latestSync.duration_seconds < 60
                              ? `${latestSync.duration_seconds}s`
                              : `${Math.floor(latestSync.duration_seconds / 60)}m ${latestSync.duration_seconds % 60}s`}
                          </>
                        )}
                      </div>
                      <div className="text-sm mt-2 grid grid-cols-2 gap-x-4 gap-y-1">
                        <div>
                          <span className="text-muted-foreground">Activities:</span>{' '}
                          {latestSync.activities_imported}
                          {latestSync.activities_skipped > 0 && (
                            <span className="text-muted-foreground">
                              {' '}
                              (+{latestSync.activities_skipped} skipped)
                            </span>
                          )}
                        </div>
                        <div>
                          <span className="text-muted-foreground">Streams:</span>{' '}
                          {latestSync.streams_imported}
                        </div>
                        {latestSync.failed_count > 0 && (
                          <div className="text-red-600">
                            <span className="text-muted-foreground">Failed:</span>{' '}
                            {latestSync.failed_count}
                          </div>
                        )}
                      </div>
                      {latestSync.error && (
                        <div className="text-sm text-red-600 mt-2 p-2 rounded bg-red-50 dark:bg-red-950">
                          {latestSync.error}
                        </div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            )}
          </div>
        </section>

        <SyncHistoryModal open={showHistoryModal} onOpenChange={setShowHistoryModal} />

        {/* DISPLAY SECTION */}
        <section>
          <SectionHeader icon={Palette} title="Display" />
          <div className="space-y-4">
            <Card>
              <CardHeader>
                <CardTitle>Preferences</CardTitle>
                <CardDescription>Customize how data is displayed</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div className="flex items-center justify-between gap-4">
                    <div>
                      <p className="font-medium">Unit System</p>
                      <p className="text-sm text-muted-foreground">
                        Choose metric or imperial units
                      </p>
                    </div>
                    <SegmentedButtons<UnitSystem>
                      value={unitSystem}
                      options={[
                        { value: 'metric', label: 'Metric' },
                        { value: 'imperial', label: 'Imperial' },
                      ]}
                      onChange={setUnitSystem}
                    />
                  </div>

                  <div className="flex items-center justify-between gap-4">
                    <div>
                      <p className="font-medium">Theme</p>
                      <p className="text-sm text-muted-foreground">Light, dark, or follow system</p>
                    </div>
                    <SegmentedButtons<Theme>
                      value={theme}
                      options={[
                        { value: 'system', label: 'System' },
                        { value: 'light', label: 'Light' },
                        { value: 'dark', label: 'Dark' },
                      ]}
                      onChange={setTheme}
                    />
                  </div>
                </div>
              </CardContent>
            </Card>
          </div>
        </section>

        {/* ADVANCED SECTION */}
        {isAuthenticated && (
          <section>
            <SectionHeader icon={Wrench} title="Advanced" />
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Virtual World Maps</CardTitle>
                  <CardDescription>
                    Configure tile layers for indoor cycling platforms
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <VirtualWorldTilesEditor
                    settings={appSettings}
                    onSave={(s) => updateAppSettings.mutate(s)}
                    saving={updateAppSettings.isPending}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Eddington Definitions</CardTitle>
                  <CardDescription>Customize Eddington number tracking categories</CardDescription>
                </CardHeader>
                <CardContent>
                  <EddingtonDefinitionsEditor
                    settings={appSettings}
                    onSave={(s) => updateAppSettings.mutate(s)}
                    saving={updateAppSettings.isPending}
                  />
                </CardContent>
              </Card>
            </div>
          </section>
        )}

        {/* NOTIFICATIONS SECTION - Server mode only */}
        {features.showNotificationSettings && isAuthenticated && appSettings && (
          <section>
            <SectionHeader icon={Bell} title="Notifications" />
            <div className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Notification Settings</CardTitle>
                  <CardDescription>
                    Configure alerts for import completion and maintenance reminders
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <NotificationsEditor
                    settings={appSettings}
                    onSave={(s) => updateAppSettings.mutate(s)}
                    saving={updateAppSettings.isPending}
                  />
                </CardContent>
              </Card>
            </div>
          </section>
        )}
      </div>
    </div>
  )
}
