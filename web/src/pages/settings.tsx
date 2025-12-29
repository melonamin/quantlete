import {
  useAppSettings,
  useAuthStatus,
  useCancelImport,
  useDeleteHrZoneDefinition,
  useFtpHistory,
  useHrZoneDefinitions,
  useImportProgress,
  useLatestSync,
  useStartImport,
  useUpdateAppSettings,
  useUpdateFtpHistory,
  useUpdateWeightHistory,
  useUpsertHrZoneDefinition,
  useWeightHistory,
} from '@/lib/data'
import { useQueryClient } from '@tanstack/react-query'
import { isWasmMode } from '@/lib/mode'
import { getAuthUrl } from '@/lib/wasm/strava/client'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { ImportStatus } from '@/components/sync/import-status'
import { SyncHistoryModal } from '@/components/sync/sync-history-modal'
import { Loader2, Check, X, ChevronDown, ChevronRight, History, Clock, AlertCircle } from 'lucide-react'
import { LineChart } from '@/components/charts'
import { useMemo, useState } from 'react'
import { formatDistance } from 'date-fns'

export function SettingsPage() {
  const queryClient = useQueryClient()
  const { data: authStatus } = useAuthStatus()
  const { data: progress } = useImportProgress()
  const { data: latestSync } = useLatestSync()
  const startImport = useStartImport()
  const cancelImport = useCancelImport()

  const isAuthenticated = authStatus?.authenticated
  const isImporting = progress?.status === 'running'
  // UI shows "include" checkboxes, but API uses "skip" flags (inverted)
  const [includeStreams, setIncludeStreams] = useState(true)
  const [includeSegments, setIncludeSegments] = useState(true)
  const [includeBestEfforts, setIncludeBestEfforts] = useState(true)
  const [includePhotos, setIncludePhotos] = useState(true)
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [showHistoryModal, setShowHistoryModal] = useState(false)

  const { data: ftpHistory } = useFtpHistory()
  const updateFtp = useUpdateFtpHistory()
  const { data: weightHistory } = useWeightHistory()
  const updateWeight = useUpdateWeightHistory()
  const { data: hrZoneDefs } = useHrZoneDefinitions({ enabled: !!isAuthenticated })
  const upsertHrZones = useUpsertHrZoneDefinition()
  const deleteHrZones = useDeleteHrZoneDefinition()
  const { data: appSettings } = useAppSettings({ enabled: !!isAuthenticated })
  const updateAppSettings = useUpdateAppSettings()

  const handleStartImport = () => {
    // API uses skip flags (inverted from UI include checkboxes)
    startImport.mutate({
      skip_streams: !includeStreams,
      skip_segments: !includeSegments,
      skip_best_efforts: !includeBestEfforts,
      skip_photos: !includePhotos,
    })
  }

  const handleCancelImport = () => {
    cancelImport.mutate()
  }

  const handleImportComplete = () => {
    // Invalidate activities data to refresh after import
    queryClient.invalidateQueries({ queryKey: ['data', 'activities'] })
    queryClient.invalidateQueries({ queryKey: ['data', 'dashboard'] })
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
                  <a
                    href={
                      isWasmMode()
                        ? getAuthUrl(`${window.location.origin}/oauth/callback`)
                        : '/api/v1/auth/strava'
                    }
                  >
                    Connect Strava
                  </a>
                </Button>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Data Import */}
        <Card>
          <CardHeader>
            <CardTitle>Data Import</CardTitle>
            <CardDescription>
              Import activities from Strava. Activities load first for immediate dashboard access,
              then streams, segments, and photos.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {!isAuthenticated ? (
              <p className="text-sm text-muted-foreground">
                Connect your Strava account first to import activities.
              </p>
            ) : (
              <>
                {progress && progress.status !== 'idle' && <ImportStatus progress={progress} />}

                <div className="flex gap-2">
                  <Button onClick={handleStartImport} disabled={isImporting || !isAuthenticated}>
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

                <div className="rounded-md border border-border bg-muted/30">
                  <button
                    type="button"
                    className="flex w-full items-center gap-2 p-3 text-sm text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => setShowAdvanced(!showAdvanced)}
                    disabled={isImporting}
                  >
                    {showAdvanced ? (
                      <ChevronDown className="h-4 w-4" />
                    ) : (
                      <ChevronRight className="h-4 w-4" />
                    )}
                    Advanced options
                  </button>
                  {showAdvanced && (
                    <div className="space-y-2 px-3 pb-3 border-t border-border pt-3">
                      <p className="text-xs text-muted-foreground mb-2">
                        All data is imported by default. Uncheck to skip specific data types.
                      </p>
                      <label className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={includeStreams}
                          onChange={(e) => setIncludeStreams(e.target.checked)}
                          disabled={isImporting}
                        />
                        Streams (GPS, heartrate, power - for maps & charts)
                      </label>
                      <label className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={includeSegments}
                          onChange={(e) => setIncludeSegments(e.target.checked)}
                          disabled={isImporting}
                        />
                        Segments (segment efforts & leaderboards)
                      </label>
                      <label className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={includeBestEfforts}
                          onChange={(e) => setIncludeBestEfforts(e.target.checked)}
                          disabled={isImporting}
                        />
                        Best efforts (PRs for standard distances)
                      </label>
                      <label className="flex items-center gap-2 text-sm">
                        <input
                          type="checkbox"
                          checked={includePhotos}
                          onChange={(e) => setIncludePhotos(e.target.checked)}
                          disabled={isImporting}
                        />
                        Photos (activity photos)
                      </label>
                    </div>
                  )}
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Last Sync Status */}
        {isAuthenticated && latestSync && (
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
                <div className="flex items-center justify-center h-10 w-10 rounded-full shrink-0"
                  style={{
                    backgroundColor: latestSync.status === 'completed' ? 'var(--green-100, #dcfce7)'
                      : latestSync.status === 'failed' ? 'var(--red-100, #fee2e2)'
                      : latestSync.status === 'running' ? 'var(--blue-100, #dbeafe)'
                      : 'var(--yellow-100, #fef3c7)'
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
                    {formatDistance(new Date(latestSync.started_at), new Date(), { addSuffix: true })}
                    {latestSync.duration_seconds && latestSync.status !== 'running' && (
                      <> • {latestSync.duration_seconds < 60
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
                        <span className="text-muted-foreground"> (+{latestSync.activities_skipped} skipped)</span>
                      )}
                    </div>
                    <div>
                      <span className="text-muted-foreground">Streams:</span>{' '}
                      {latestSync.streams_imported}
                    </div>
                    {latestSync.failed_count > 0 && (
                      <div className="text-red-600">
                        <span className="text-muted-foreground">Failed:</span> {latestSync.failed_count}
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

        <SyncHistoryModal open={showHistoryModal} onOpenChange={setShowHistoryModal} />

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

        {/* Athlete Metrics */}
        {isAuthenticated && (
          <Card>
            <CardHeader>
              <CardTitle>Athlete Metrics</CardTitle>
              <CardDescription>Used for training load and zone calculations</CardDescription>
            </CardHeader>
            <CardContent className="space-y-6">
              <FtpEditor
                value={ftpHistory}
                onSave={(v) => updateFtp.mutate(v)}
                saving={updateFtp.isPending}
              />
              <WeightEditor
                value={weightHistory}
                onSave={(v) => updateWeight.mutate(v)}
                saving={updateWeight.isPending}
              />
              <HrZonesEditor
                defs={hrZoneDefs}
                onSave={(def) => upsertHrZones.mutate(def)}
                onDelete={(p) => deleteHrZones.mutate(p)}
                saving={upsertHrZones.isPending}
              />
              <VirtualWorldTilesEditor
                settings={appSettings}
                onSave={(s) => updateAppSettings.mutate(s)}
                saving={updateAppSettings.isPending}
              />
              <EddingtonDefinitionsEditor
                settings={appSettings}
                onSave={(s) => updateAppSettings.mutate(s)}
                saving={updateAppSettings.isPending}
              />
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  )
}

function parseISODate(d: string) {
  const t = new Date(d + 'T00:00:00')
  return Number.isFinite(t.getTime()) ? t : null
}

function formatPaceFromMps(mps: number) {
  if (mps <= 0) return '–'
  const secPerKm = 1000 / mps
  const mm = Math.floor(secPerKm / 60)
  const ss = Math.round(secPerKm % 60)
  return `${mm}:${String(ss).padStart(2, '0')}/km`
}

function FtpEditor({
  value,
  onSave,
  saving,
}: {
  value:
    | {
        cycling: { recorded_at: string; value: number }[]
        running: { recorded_at: string; value: number }[]
      }
    | undefined
  onSave: (v: {
    cycling: { recorded_at: string; value: number }[]
    running: { recorded_at: string; value: number }[]
  }) => void
  saving: boolean
}) {
  const [cyclingDate, setCyclingDate] = useState('')
  const [cyclingValue, setCyclingValue] = useState('')
  const [runningDate, setRunningDate] = useState('')
  const [runningValue, setRunningValue] = useState('')

  const cyclingSeries = useMemo(() => {
    const pts = (value?.cycling ?? []).filter((p) => parseISODate(p.recorded_at))
    return [
      {
        name: 'Cycling FTP (W)',
        data: pts.map((p) => ({ x: p.recorded_at, y: p.value })),
      },
    ]
  }, [value])

  const runningSeries = useMemo(() => {
    const pts = (value?.running ?? []).filter((p) => parseISODate(p.recorded_at))
    return [
      {
        name: 'Running FTP (m/s)',
        data: pts.map((p) => ({ x: p.recorded_at, y: p.value })),
      },
    ]
  }, [value])

  const addPoint = (kind: 'cycling' | 'running') => {
    if (!value) return
    const date = kind === 'cycling' ? cyclingDate : runningDate
    const raw = kind === 'cycling' ? cyclingValue : runningValue
    const v = Number(raw)
    if (!date || !Number.isFinite(v) || v <= 0) return

    const next = structuredClone(value)
    next[kind] = [...next[kind], { recorded_at: date, value: v }]
      .sort((a, b) => a.recorded_at.localeCompare(b.recorded_at))
      .filter((p, idx, arr) => idx === 0 || p.recorded_at !== arr[idx - 1].recorded_at)

    onSave(next)
    if (kind === 'cycling') {
      setCyclingDate('')
      setCyclingValue('')
    } else {
      setRunningDate('')
      setRunningValue('')
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">FTP History</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>

      <div className="space-y-3">
        <div className="grid gap-2 md:grid-cols-3">
          <input
            type="date"
            className="h-9 rounded-md border border-border bg-background px-2"
            value={cyclingDate}
            onChange={(e) => setCyclingDate(e.target.value)}
          />
          <input
            inputMode="numeric"
            placeholder="Cycling FTP (W)"
            className="h-9 rounded-md border border-border bg-background px-2"
            value={cyclingValue}
            onChange={(e) => setCyclingValue(e.target.value)}
          />
          <Button onClick={() => addPoint('cycling')}>Add</Button>
        </div>
        <LineChart series={cyclingSeries} height={200} showLegend={false} />
      </div>

      <div className="space-y-3">
        <div className="grid gap-2 md:grid-cols-3">
          <input
            type="date"
            className="h-9 rounded-md border border-border bg-background px-2"
            value={runningDate}
            onChange={(e) => setRunningDate(e.target.value)}
          />
          <input
            inputMode="decimal"
            placeholder="Running FTP (m/s)"
            className="h-9 rounded-md border border-border bg-background px-2"
            value={runningValue}
            onChange={(e) => setRunningValue(e.target.value)}
          />
          <Button onClick={() => addPoint('running')}>Add</Button>
        </div>
        <div className="text-xs text-muted-foreground">
          Store running threshold speed in m/s (tooltip pace example: {formatPaceFromMps(3.5)}).
        </div>
        <LineChart series={runningSeries} height={200} showLegend={false} />
      </div>
    </div>
  )
}

function WeightEditor({
  value,
  onSave,
  saving,
}: {
  value: { points: { recorded_at: string; value: number }[] } | undefined
  onSave: (v: { points: { recorded_at: string; value: number }[] }) => void
  saving: boolean
}) {
  const [date, setDate] = useState('')
  const [weight, setWeight] = useState('')

  const series = useMemo(() => {
    const pts = (value?.points ?? []).filter((p) => parseISODate(p.recorded_at))
    return [
      {
        name: 'Weight (kg)',
        data: pts.map((p) => ({ x: p.recorded_at, y: p.value })),
      },
    ]
  }, [value])

  const addPoint = () => {
    if (!value) return
    const v = Number(weight)
    if (!date || !Number.isFinite(v) || v <= 0) return
    const next = structuredClone(value)
    next.points = [...next.points, { recorded_at: date, value: v }]
      .sort((a, b) => a.recorded_at.localeCompare(b.recorded_at))
      .filter((p, idx, arr) => idx === 0 || p.recorded_at !== arr[idx - 1].recorded_at)
    onSave(next)
    setDate('')
    setWeight('')
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Weight History</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <div className="grid gap-2 md:grid-cols-3">
        <input
          type="date"
          className="h-9 rounded-md border border-border bg-background px-2"
          value={date}
          onChange={(e) => setDate(e.target.value)}
        />
        <input
          inputMode="decimal"
          placeholder="Weight (kg)"
          className="h-9 rounded-md border border-border bg-background px-2"
          value={weight}
          onChange={(e) => setWeight(e.target.value)}
        />
        <Button onClick={addPoint}>Add</Button>
      </div>
      <LineChart series={series} height={200} showLegend={false} />
    </div>
  )
}

function HrZonesEditor({
  defs,
  onSave,
  onDelete,
  saving,
}: {
  defs:
    | {
        sport_type: string
        effective_from: string
        method: string
        zones: { bounds: number[]; hr_max?: number }
      }[]
    | undefined
  onSave: (def: {
    sport_type: string
    effective_from: string
    method: string
    zones: { bounds: number[]; hr_max?: number }
  }) => void
  onDelete: (p: { sport_type: string; effective_from: string }) => void
  saving: boolean
}) {
  const defaultDef = {
    sport_type: 'All',
    effective_from: '1970-01-01',
    method: 'percent_hrmax',
    zones: { hr_max: 190, bounds: [0.6, 0.7, 0.8, 0.9, 1.0] },
  }

  const [sportType, setSportType] = useState('All')
  const [effectiveFrom, setEffectiveFrom] = useState('1970-01-01')
  const [method, setMethod] = useState(defaultDef.method)
  const [hrMax, setHrMax] = useState(String(defaultDef.zones.hr_max ?? 190))
  const [bounds, setBounds] = useState(defaultDef.zones.bounds.map(String))

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Heart Rate Zones</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <div className="grid gap-2 md:grid-cols-3">
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Sport group</label>
          <select
            className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
            value={sportType}
            onChange={(e) => setSportType(e.target.value)}
          >
            <option value="All">All</option>
            <option value="Ride">Ride</option>
            <option value="Run">Run</option>
          </select>
        </div>
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Effective from</label>
          <input
            type="date"
            className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
            value={effectiveFrom}
            onChange={(e) => setEffectiveFrom(e.target.value)}
          />
        </div>
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Method</label>
          <select
            className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
            value={method}
            onChange={(e) => setMethod(e.target.value)}
          >
            <option value="percent_hrmax">% of HR max</option>
            <option value="absolute_bpm">Absolute BPM</option>
          </select>
        </div>
        {method === 'percent_hrmax' && (
          <div className="space-y-1">
            <label className="text-sm text-muted-foreground">HR max (bpm)</label>
            <input
              inputMode="numeric"
              className="h-9 w-full rounded-md border border-border bg-background px-2"
              value={hrMax}
              onChange={(e) => setHrMax(e.target.value)}
            />
          </div>
        )}
      </div>
      <div className="grid gap-2 md:grid-cols-5">
        {bounds.slice(0, 5).map((v, idx) => (
          <div key={idx} className="space-y-1">
            <label className="text-sm text-muted-foreground">Z{idx + 1} max</label>
            <input
              inputMode="decimal"
              className="h-9 w-full rounded-md border border-border bg-background px-2"
              value={v}
              onChange={(e) => {
                const next = bounds.slice()
                next[idx] = e.target.value
                setBounds(next)
              }}
            />
          </div>
        ))}
      </div>
      <div className="flex justify-end">
        <Button
          onClick={() => {
            const parsedBounds = bounds.map((b) => Number(b)).filter((n) => Number.isFinite(n))
            if (parsedBounds.length < 5) return
            onSave({
              sport_type: sportType,
              effective_from: effectiveFrom,
              method,
              zones:
                method === 'percent_hrmax'
                  ? { hr_max: Number(hrMax), bounds: parsedBounds }
                  : { bounds: parsedBounds },
            })
          }}
        >
          Save Definition
        </Button>
      </div>
      <p className="text-xs text-muted-foreground">
        For % zones, enter bounds as fractions (e.g. 0.6, 0.7, 0.8, 0.9, 1.0). For absolute zones,
        enter BPM caps.
      </p>

      {defs && defs.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-medium">Existing definitions</div>
          <div className="space-y-2">
            {defs.map((d) => (
              <div
                key={`${d.sport_type}:${d.effective_from}`}
                className="flex items-center justify-between rounded-md border border-border px-3 py-2"
              >
                <div className="text-sm">
                  <div className="font-medium">{d.sport_type}</div>
                  <div className="text-xs text-muted-foreground">
                    From {d.effective_from} • {d.method}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      setSportType(d.sport_type)
                      setEffectiveFrom(d.effective_from)
                      setMethod(d.method)
                      setHrMax(String(d.zones.hr_max ?? 190))
                      setBounds((d.zones.bounds ?? defaultDef.zones.bounds).slice(0, 5).map(String))
                    }}
                  >
                    Edit
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      onDelete({ sport_type: d.sport_type, effective_from: d.effective_from })
                    }
                  >
                    Delete
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

const VIRTUAL_WORLDS = [
  'Watopia',
  'London',
  'New York',
  'Innsbruck',
  'Richmond',
  'France',
  'Makuri Islands',
  'Scotland',
  'Yumezi',
  'Crit City',
] as const

const DEFAULT_EDDINGTON_DEFS = [
  {
    id: 'all',
    name: 'All activities',
    sport_types: undefined,
    show_in_nav: true,
    show_in_dashboard_widget: true,
  },
  {
    id: 'rides',
    name: 'Rides',
    sport_types: ['Ride', 'MountainBikeRide', 'GravelRide', 'EBikeRide', 'VirtualRide'],
    show_in_nav: true,
    show_in_dashboard_widget: true,
  },
  {
    id: 'runs',
    name: 'Runs',
    sport_types: ['Run', 'TrailRun', 'VirtualRun'],
    show_in_nav: true,
    show_in_dashboard_widget: true,
  },
]

function VirtualWorldTilesEditor({
  settings,
  onSave,
  saving,
}: {
  settings:
    | {
        version: number
        virtual_world_tile_layers: Record<
          string,
          { name: string; url: string; attribution?: string; max_zoom?: number }
        >
        eddington_definitions?: {
          id: string
          name: string
          sport_types?: string[]
          show_in_nav?: boolean
          show_in_dashboard_widget?: boolean
        }[]
      }
    | undefined
  onSave: (s: {
    version: number
    virtual_world_tile_layers: Record<
      string,
      { name: string; url: string; attribution?: string; max_zoom?: number }
    >
    eddington_definitions?: {
      id: string
      name: string
      sport_types?: string[]
      show_in_nav?: boolean
      show_in_dashboard_widget?: boolean
    }[]
  }) => void
  saving: boolean
}) {
  const current = useMemo(() => {
    const merged = settings ?? {
      version: 2,
      virtual_world_tile_layers: {},
      eddington_definitions: DEFAULT_EDDINGTON_DEFS,
    }
    return {
      ...merged,
      version: merged.version ?? 2,
      virtual_world_tile_layers: merged.virtual_world_tile_layers ?? {},
      eddington_definitions:
        merged.eddington_definitions && merged.eddington_definitions.length
          ? merged.eddington_definitions
          : DEFAULT_EDDINGTON_DEFS,
    }
  }, [settings])
  const [world, setWorld] = useState<string>(VIRTUAL_WORLDS[0])
  const [url, setUrl] = useState('')
  const [attribution, setAttribution] = useState('')
  const [maxZoom, setMaxZoom] = useState('18')

  const addOrUpdate = () => {
    if (!url) return
    const next = structuredClone(current)
    next.virtual_world_tile_layers[world] = {
      name: world,
      url,
      attribution: attribution || undefined,
      max_zoom: maxZoom ? Number(maxZoom) : undefined,
    }
    next.version = Math.max(next.version ?? 2, 2)
    onSave(next)
    setUrl('')
    setAttribution('')
  }

  const remove = (w: string) => {
    const next = structuredClone(current)
    delete next.virtual_world_tile_layers[w]
    next.version = Math.max(next.version ?? 2, 2)
    onSave(next)
  }

  const entries = Object.entries(current.virtual_world_tile_layers ?? {})

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Virtual World Maps</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <p className="text-xs text-muted-foreground">
        Add tile URLs for virtual worlds (Zwift/Rouvy/MyWhoosh). When a virtual activity is
        detected, the map switches to the matching tile layer.
      </p>
      <div className="grid gap-2 md:grid-cols-2">
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">World</label>
          <select
            className="h-9 w-full rounded-md border border-border bg-background px-2 text-sm"
            value={world}
            onChange={(e) => setWorld(e.target.value)}
          >
            {VIRTUAL_WORLDS.map((w) => (
              <option key={w} value={w}>
                {w}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Max zoom</label>
          <input
            className="h-9 w-full rounded-md border border-border bg-background px-2"
            value={maxZoom}
            onChange={(e) => setMaxZoom(e.target.value)}
          />
        </div>
        <div className="space-y-1 md:col-span-2">
          <label className="text-sm text-muted-foreground">Tile URL template</label>
          <input
            className="h-9 w-full rounded-md border border-border bg-background px-2"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="https://example.com/tiles/{z}/{x}/{y}.png"
          />
        </div>
        <div className="space-y-1 md:col-span-2">
          <label className="text-sm text-muted-foreground">Attribution (optional)</label>
          <input
            className="h-9 w-full rounded-md border border-border bg-background px-2"
            value={attribution}
            onChange={(e) => setAttribution(e.target.value)}
          />
        </div>
      </div>
      <div className="flex justify-end">
        <Button onClick={addOrUpdate}>Save Tile Layer</Button>
      </div>

      {entries.length > 0 && (
        <div className="space-y-2">
          <div className="text-sm font-medium">Configured worlds</div>
          <div className="space-y-2">
            {entries.map(([k, v]) => (
              <div key={k} className="rounded-md border border-border p-3">
                <div className="flex items-center justify-between gap-2">
                  <div className="min-w-0">
                    <div className="font-medium">{k}</div>
                    <div className="truncate text-xs text-muted-foreground">{v.url}</div>
                  </div>
                  <Button variant="outline" size="sm" onClick={() => remove(k)}>
                    Remove
                  </Button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

function EddingtonDefinitionsEditor({
  settings,
  onSave,
  saving,
}: {
  settings:
    | {
        version: number
        virtual_world_tile_layers: Record<
          string,
          { name: string; url: string; attribution?: string; max_zoom?: number }
        >
        eddington_definitions?: {
          id: string
          name: string
          sport_types?: string[]
          show_in_nav?: boolean
          show_in_dashboard_widget?: boolean
        }[]
      }
    | undefined
  onSave: (s: {
    version: number
    virtual_world_tile_layers: Record<
      string,
      { name: string; url: string; attribution?: string; max_zoom?: number }
    >
    eddington_definitions?: {
      id: string
      name: string
      sport_types?: string[]
      show_in_nav?: boolean
      show_in_dashboard_widget?: boolean
    }[]
  }) => void
  saving: boolean
}) {
  const current = settings ?? {
    version: 2,
    virtual_world_tile_layers: {},
    eddington_definitions: DEFAULT_EDDINGTON_DEFS,
  }
  const defs =
    current.eddington_definitions && current.eddington_definitions.length
      ? current.eddington_definitions
      : DEFAULT_EDDINGTON_DEFS

  const [name, setName] = useState('')
  const [sportTypes, setSportTypes] = useState('')
  const [showInNav, setShowInNav] = useState(true)
  const [showInWidget, setShowInWidget] = useState(true)

  const add = () => {
    if (!name.trim()) return
    const id = `e_${Math.random().toString(16).slice(2, 10)}`
    const types = sportTypes
      .split(',')
      .map((s) => s.trim())
      .filter(Boolean)
    const next = structuredClone(current)
    next.eddington_definitions = [
      ...(defs ?? []),
      {
        id,
        name: name.trim(),
        sport_types: types.length ? types : undefined,
        show_in_nav: showInNav,
        show_in_dashboard_widget: showInWidget,
      },
    ]
    next.version = Math.max(next.version ?? 2, 2)
    onSave(next)
    setName('')
    setSportTypes('')
  }

  const update = (
    id: string,
    patch: Partial<{
      name: string
      sport_types?: string[]
      show_in_nav?: boolean
      show_in_dashboard_widget?: boolean
    }>
  ) => {
    const next = structuredClone(current)
    next.eddington_definitions = defs.map((d) => (d.id === id ? { ...d, ...patch } : d))
    next.version = Math.max(next.version ?? 2, 2)
    onSave(next)
  }

  const remove = (id: string) => {
    const nextDefs = defs.filter((d) => d.id !== id)
    const next = structuredClone(current)
    next.eddington_definitions = nextDefs.length ? nextDefs : DEFAULT_EDDINGTON_DEFS
    next.version = Math.max(next.version ?? 2, 2)
    onSave(next)
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-medium">Eddington Definitions</h3>
        {saving && <span className="text-xs text-muted-foreground">Saving…</span>}
      </div>
      <p className="text-xs text-muted-foreground">
        Create multiple Eddington definitions (e.g. Rides, Runs). Leave sport types empty to include
        all activities.
      </p>

      <div className="grid gap-2 md:grid-cols-2">
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Name</label>
          <input
            className="h-9 w-full rounded-md border border-border bg-background px-2"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Commutes"
          />
        </div>
        <div className="space-y-1">
          <label className="text-sm text-muted-foreground">Sport types (optional)</label>
          <input
            className="h-9 w-full rounded-md border border-border bg-background px-2"
            value={sportTypes}
            onChange={(e) => setSportTypes(e.target.value)}
            placeholder="Ride,VirtualRide"
          />
        </div>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={showInNav}
            onChange={(e) => setShowInNav(e.target.checked)}
          />
          Show in navigation
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={showInWidget}
            onChange={(e) => setShowInWidget(e.target.checked)}
          />
          Show in dashboard widget
        </label>
      </div>

      <div className="flex justify-end">
        <Button onClick={add}>Add Definition</Button>
      </div>

      <div className="space-y-2">
        <div className="text-sm font-medium">Current definitions</div>
        <div className="space-y-2">
          {defs.map((d) => (
            <div key={d.id} className="rounded-md border border-border p-3">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="min-w-0">
                  <div className="font-medium">{d.name}</div>
                  <div className="text-xs text-muted-foreground">
                    {(d.sport_types && d.sport_types.length
                      ? d.sport_types.join(', ')
                      : 'All sport types') +
                      ' • ' +
                      (d.show_in_nav ? 'Nav' : 'Hidden') +
                      ' • ' +
                      (d.show_in_dashboard_widget ? 'Widget' : 'Hidden')}
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => update(d.id, { show_in_nav: !d.show_in_nav })}
                  >
                    {d.show_in_nav ? 'Hide Nav' : 'Show Nav'}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      update(d.id, { show_in_dashboard_widget: !d.show_in_dashboard_widget })
                    }
                  >
                    {d.show_in_dashboard_widget ? 'Hide Widget' : 'Show Widget'}
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => remove(d.id)}>
                    Delete
                  </Button>
                </div>
              </div>

              <div className="mt-3 grid gap-2 md:grid-cols-2">
                <div className="space-y-1">
                  <label className="text-sm text-muted-foreground">Rename</label>
                  <input
                    className="h-9 w-full rounded-md border border-border bg-background px-2"
                    value={d.name}
                    onChange={(e) => update(d.id, { name: e.target.value })}
                  />
                </div>
                <div className="space-y-1">
                  <label className="text-sm text-muted-foreground">Sport types</label>
                  <input
                    className="h-9 w-full rounded-md border border-border bg-background px-2"
                    value={(d.sport_types ?? []).join(',')}
                    onChange={(e) => {
                      const v = e.target.value
                      const types = v
                        .split(',')
                        .map((s) => s.trim())
                        .filter(Boolean)
                      update(d.id, { sport_types: types.length ? types : undefined })
                    }}
                    placeholder="All"
                  />
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
