import { useMemo, useState } from 'react'
import { useChallenges, useImportChallenges, type Challenge } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { features } from '@/lib/features'

function monthLabel(month?: string) {
  if (!month) return 'Unknown'
  const [y, m] = month.split('-').map((x) => Number(x))
  if (!y || !m) return month
  const d = new Date(y, m - 1, 1)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'long' })
}

export function ChallengesPage() {
  const { data, isLoading, error } = useChallenges()
  const importChallenges = useImportChallenges()
  const [file, setFile] = useState<File | null>(null)
  const canImport = features.challengeImport

  const grouped = useMemo(() => {
    const byMonth = new Map<string, Challenge[]>()
    for (const c of data ?? []) {
      const key = c.month || 'unknown'
      byMonth.set(key, [...(byMonth.get(key) ?? []), c])
    }
    const months = Array.from(byMonth.keys()).sort((a, b) => b.localeCompare(a))
    return { byMonth, months }
  }, [data])

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Challenges</h1>
          <p className="text-muted-foreground">Completed Strava challenges</p>
        </div>
      </div>

      <Card className="mb-6">
        <CardHeader className="pb-3">
          <CardTitle className="text-base">Import</CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="text-sm text-muted-foreground">
            Upload a Strava trophy case HTML export to import your completed challenges.
          </div>
          {!canImport && (
            <div className="text-sm text-muted-foreground">
              Challenge import is only available in self-hosted mode.
            </div>
          )}
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
            <input
              type="file"
              accept=".html,text/html"
              onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              disabled={!canImport}
            />
            <Button
              onClick={() => canImport && file && importChallenges.mutate({ file })}
              disabled={!canImport || !file || importChallenges.isPending}
            >
              {importChallenges.isPending ? 'Importing…' : 'Import'}
            </Button>
          </div>
          {importChallenges.error && (
            <div className="text-sm text-destructive">{importChallenges.error.message}</div>
          )}
          {importChallenges.data && (
            <div className="text-sm text-muted-foreground">
              Imported {importChallenges.data.imported} challenge
              {importChallenges.data.imported === 1 ? '' : 's'}.
            </div>
          )}
        </CardContent>
      </Card>

      {error && (
        <div className="mb-4 rounded-lg border border-destructive bg-destructive/10 p-4">
          <p className="text-destructive">Failed to load challenges: {error.message}</p>
        </div>
      )}

      {isLoading ? (
        <div className="space-y-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i}>
              <CardHeader className="pb-3">
                <Skeleton className="h-5 w-40" />
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-3 gap-3 sm:grid-cols-5 lg:grid-cols-8">
                  {Array.from({ length: 8 }).map((_, j) => (
                    <Skeleton key={j} className="h-24 w-full" />
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : (data ?? []).length === 0 ? (
        <div className="rounded-lg border border-border bg-card p-8 text-center">
          <p className="text-muted-foreground">No challenges imported yet.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {grouped.months.map((m) => {
            const items = grouped.byMonth.get(m) ?? []
            return (
              <Card key={m}>
                <CardHeader className="pb-3">
                  <CardTitle className="text-base">
                    {monthLabel(m === 'unknown' ? undefined : m)}{' '}
                    <span className="text-muted-foreground font-normal">({items.length})</span>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="grid grid-cols-3 gap-3 sm:grid-cols-5 lg:grid-cols-8">
                    {items.map((c) => (
                      <a
                        key={c.id}
                        href={c.slug ? `https://www.strava.com/challenges/${c.slug}` : undefined}
                        target="_blank"
                        rel="noreferrer"
                        className="group rounded-md border border-border bg-background p-2 hover:bg-accent/30"
                        title={c.name}
                      >
                        <div className="aspect-square overflow-hidden rounded-md bg-muted">
                          {c.badge_url ? (
                            <img
                              src={c.badge_url}
                              alt={c.name}
                              loading="lazy"
                              className="h-full w-full object-contain"
                            />
                          ) : (
                            <div className="flex h-full w-full items-center justify-center text-xs text-muted-foreground">
                              No badge
                            </div>
                          )}
                        </div>
                        <div className="mt-2 line-clamp-2 text-xs font-medium group-hover:underline">
                          {c.name}
                        </div>
                        {c.completion_date && (
                          <div className="mt-1 text-[11px] text-muted-foreground">
                            {c.completion_date}
                          </div>
                        )}
                      </a>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
