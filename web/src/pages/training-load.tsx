import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { AlertTriangle, Info } from 'lucide-react'
import { useTrainingLoad } from '@/lib/api'
import { TrainingLoadChart } from '@/components/charts'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'

export function TrainingLoadPage() {
  const [after, setAfter] = useState('')
  const [before, setBefore] = useState('')
  const { data, isLoading, error } = useTrainingLoad({
    after: after || undefined,
    before: before || undefined,
  })

  const series = useMemo(() => data?.series ?? [], [data])
  const diagnostics = data?.diagnostics
  const hasWarnings = (diagnostics?.missing_config_warnings?.length ?? 0) > 0

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

      {hasWarnings && (
        <Alert className="mb-6 border-amber-500/50 bg-amber-50 dark:bg-amber-950/20">
          <AlertTriangle className="h-4 w-4 text-amber-600" />
          <AlertTitle className="text-amber-700 dark:text-amber-400">
            Configuration Required
          </AlertTitle>
          <AlertDescription className="text-amber-600 dark:text-amber-300">
            <ul className="mt-2 list-inside list-disc space-y-1">
              {diagnostics?.missing_config_warnings?.map((warning, i) => (
                <li key={i}>{warning}</li>
              ))}
            </ul>
            <p className="mt-2">
              <Link to="/settings" className="font-medium underline hover:no-underline">
                Go to Settings
              </Link>{' '}
              to configure your thresholds.
            </p>
          </AlertDescription>
        </Alert>
      )}

      {diagnostics && (
        <Alert className="mb-6">
          <Info className="h-4 w-4" />
          <AlertTitle>Activity Breakdown</AlertTitle>
          <AlertDescription>
            <div className="mt-2 flex flex-wrap gap-4 text-sm">
              <span>Total activities: {diagnostics.total_activities}</span>
              <span>With power data: {diagnostics.activities_with_power}</span>
              <span>With pace data: {diagnostics.activities_with_speed}</span>
              <span>With HR data: {diagnostics.activities_with_hr}</span>
              <span className="font-medium">With computed TSS: {diagnostics.activities_with_tss}</span>
            </div>
          </AlertDescription>
        </Alert>
      )}

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

      {data?.summary && (
        <div className="grid gap-4 md:grid-cols-3">
          <Card>
            <CardHeader>
              <CardTitle className="text-sm text-muted-foreground">CTL</CardTitle>
            </CardHeader>
            <CardContent className="text-2xl font-bold">{data.summary.ctl.toFixed(1)}</CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="text-sm text-muted-foreground">ATL</CardTitle>
            </CardHeader>
            <CardContent className="text-2xl font-bold">{data.summary.atl.toFixed(1)}</CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle className="text-sm text-muted-foreground">TSB</CardTitle>
            </CardHeader>
            <CardContent className="text-2xl font-bold">{data.summary.tsb.toFixed(1)}</CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
