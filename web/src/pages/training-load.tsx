import { useMemo, useState } from 'react'
import { Link } from '@tanstack/react-router'
import { useTrainingLoad } from '@/lib/api'
import { TrainingLoadChart } from '@/components/charts'
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
