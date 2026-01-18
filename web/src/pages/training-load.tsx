import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { AlertTriangle, CheckCircle, Settings, XCircle } from 'lucide-react'
import { useTrainingLoad } from '@/lib/api'
import { TrainingLoadChart } from '@/components/charts'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Progress } from '@/components/ui/progress'

export function TrainingLoadPage() {
  const [after, setAfter] = useState('')
  const [before, setBefore] = useState('')
  const { data, isLoading, error } = useTrainingLoad({
    after: after || undefined,
    before: before || undefined,
  })

  const series = useMemo(() => data?.series ?? [], [data])
  const stats = data?.stats
  const weeklyMetrics = data?.weekly_metrics
  const polarized = data?.polarized

  // Check if FTP is configured
  const hasFTPConfigured = stats?.cycling_ftp_configured || stats?.running_ftp_configured
  const hasNoTSS = stats && stats.activities_with_tss === 0

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <p className="text-destructive">Failed to load training load.</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Training Load</h1>
          <p className="text-muted-foreground">Fitness, fatigue, and form over time</p>
        </div>
        <Button asChild variant="outline">
          <Link to="/">Back</Link>
        </Button>
      </div>

      {/* Configuration Warning Banner */}
      {stats && !hasFTPConfigured && (
        <Alert variant="destructive" className="mb-6">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>FTP Not Configured</AlertTitle>
          <AlertDescription className="flex items-center justify-between">
            <span>
              Configure your FTP (Functional Threshold Power/Pace) in Settings to enable training
              load calculations.
            </span>
            <Button asChild variant="outline" size="sm" className="ml-4">
              <Link to="/settings">
                <Settings className="mr-2 h-4 w-4" />
                Configure FTP
              </Link>
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {/* No TSS Warning */}
      {stats && hasFTPConfigured && hasNoTSS && (
        <Alert className="mb-6">
          <AlertTriangle className="h-4 w-4" />
          <AlertTitle>No Training Load Data</AlertTitle>
          <AlertDescription>
            No activities with computed TSS in this date range. This could be because:
            <ul className="mt-2 list-inside list-disc text-sm">
              <li>Activities don&apos;t have power or heart rate data</li>
              <li>FTP was configured after the activities were recorded</li>
              <li>The date range doesn&apos;t include any activities</li>
            </ul>
          </AlertDescription>
        </Alert>
      )}

      {/* Activity Stats Cards */}
      {stats && (
        <div className="mb-6 grid gap-4 md:grid-cols-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Total Activities</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.total_activities}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>With Power Data</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.activities_with_power}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>With TSS</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.activities_with_tss}</div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>Configuration</CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-1 text-sm">
              <div className="flex items-center gap-2">
                {stats.cycling_ftp_configured ? (
                  <CheckCircle className="h-4 w-4 text-green-500" />
                ) : (
                  <XCircle className="h-4 w-4 text-muted-foreground" />
                )}
                <span>Cycling FTP</span>
              </div>
              <div className="flex items-center gap-2">
                {stats.running_ftp_configured ? (
                  <CheckCircle className="h-4 w-4 text-green-500" />
                ) : (
                  <XCircle className="h-4 w-4 text-muted-foreground" />
                )}
                <span>Running FTP</span>
              </div>
              <div className="flex items-center gap-2">
                {stats.hr_zones_configured ? (
                  <CheckCircle className="h-4 w-4 text-green-500" />
                ) : (
                  <XCircle className="h-4 w-4 text-muted-foreground" />
                )}
                <span>HR Zones</span>
              </div>
            </CardContent>
          </Card>
        </div>
      )}

      {/* Chart */}
      <Card className="mb-6">
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Range</CardTitle>
          <div className="flex flex-wrap items-center gap-2">
            <Input
              type="date"
              className="w-auto"
              value={after}
              onChange={(e) => setAfter(e.target.value)}
            />
            <span className="text-sm text-muted-foreground">to</span>
            <Input
              type="date"
              className="w-auto"
              value={before}
              onChange={(e) => setBefore(e.target.value)}
            />
          </div>
        </CardHeader>
        <CardContent>
          <TrainingLoadChart data={series} height={360} loading={isLoading} />
        </CardContent>
      </Card>

      {/* Current Summary */}
      {data?.summary && (
        <div className="mb-6 grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">CTL (Fitness)</CardTitle>
            </CardHeader>
            <CardContent className="text-2xl font-bold">{data.summary.ctl.toFixed(1)}</CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">ATL (Fatigue)</CardTitle>
            </CardHeader>
            <CardContent className="text-2xl font-bold">{data.summary.atl.toFixed(1)}</CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm text-muted-foreground">TSB (Form)</CardTitle>
            </CardHeader>
            <CardContent
              className={`text-2xl font-bold ${
                data.summary.tsb > 0 ? 'text-green-500' : data.summary.tsb < -10 ? 'text-red-500' : ''
              }`}
            >
              {data.summary.tsb > 0 ? '+' : ''}
              {data.summary.tsb.toFixed(1)}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Weekly Metrics */}
      {weeklyMetrics && (
        <div className="mb-6">
          <h2 className="mb-4 text-lg font-semibold">Weekly Summary (Last 7 Days)</h2>
          <div className="grid gap-4 md:grid-cols-4">
            <Card>
              <CardHeader className="pb-2">
                <CardDescription>Rest Days</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{weeklyMetrics.rest_days}</div>
                <p className="text-xs text-muted-foreground">out of 7</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardDescription>Weekly Strain</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{weeklyMetrics.weekly_strain.toFixed(0)}</div>
                <p className="text-xs text-muted-foreground">Total TSS</p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardDescription>Monotony</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{weeklyMetrics.monotony.toFixed(2)}</div>
                <p className="text-xs text-muted-foreground">
                  {weeklyMetrics.monotony > 2 ? 'High (risk of overtraining)' : 'Normal'}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-2">
                <CardDescription>Weekly TRIMP</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {weeklyMetrics.weekly_trimp > 0 ? weeklyMetrics.weekly_trimp.toFixed(0) : '—'}
                </div>
                <p className="text-xs text-muted-foreground">Training Impulse</p>
              </CardContent>
            </Card>
          </div>
        </div>
      )}

      {/* Polarized Training Breakdown */}
      {polarized && polarized.total_seconds > 0 && (
        <div className="mb-6">
          <h2 className="mb-4 text-lg font-semibold">
            Polarized Training (Last {polarized.period_days} Days)
          </h2>
          <Card>
            <CardContent className="pt-6">
              <div className="space-y-4">
                <div>
                  <div className="mb-2 flex items-center justify-between text-sm">
                    <span className="text-blue-500">Low Intensity (Z1-Z2)</span>
                    <span className="font-medium">{polarized.low_percent.toFixed(1)}%</span>
                  </div>
                  <Progress value={polarized.low_percent} className="h-3 bg-muted" />
                </div>
                <div>
                  <div className="mb-2 flex items-center justify-between text-sm">
                    <span className="text-yellow-500">Moderate (Z3)</span>
                    <span className="font-medium">{polarized.moderate_percent.toFixed(1)}%</span>
                  </div>
                  <Progress value={polarized.moderate_percent} className="h-3 bg-muted" />
                </div>
                <div>
                  <div className="mb-2 flex items-center justify-between text-sm">
                    <span className="text-red-500">High Intensity (Z4-Z5)</span>
                    <span className="font-medium">{polarized.high_percent.toFixed(1)}%</span>
                  </div>
                  <Progress value={polarized.high_percent} className="h-3 bg-muted" />
                </div>
              </div>
              <p className="mt-4 text-xs text-muted-foreground">
                Polarized training typically aims for 80% low intensity, 0% moderate, and 20% high
                intensity. {formatDuration(polarized.total_seconds)} of training time analyzed.
              </p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }
  return `${minutes}m`
}
