import { Link } from '@tanstack/react-router'
import { useChallenges } from '@/lib/api'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

function formatMonthLabel(month?: string) {
  if (!month) return ''
  const [y, m] = month.split('-').map((x) => Number(x))
  if (!y || !m) return month
  const d = new Date(y, m - 1, 1)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short' })
}

export function RecentChallenges() {
  const { data, isLoading } = useChallenges()
  const items = (data ?? []).slice(0, 6)

  return (
    <WidgetWrapper
      title="Recent Challenges"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/challenges">View all</Link>
        </Button>
      }
      isLoading={isLoading}
    >
      {isLoading ? (
        <div className="grid grid-cols-3 gap-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="space-y-2">
              <Skeleton className="h-20 w-full" />
              <Skeleton className="h-3 w-24" />
            </div>
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="text-sm text-muted-foreground">No challenges imported yet.</div>
      ) : (
        <div className="grid grid-cols-3 gap-3">
          {items.map((c) => (
            <a
              key={c.id}
              href={c.slug ? `https://www.strava.com/challenges/${c.slug}` : undefined}
              target="_blank"
              rel="noreferrer"
              className="rounded-md border border-border bg-background p-2 hover:bg-accent/30"
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
              <div className="mt-2 line-clamp-2 text-xs font-medium">{c.name}</div>
              <div className="mt-1 text-[11px] text-muted-foreground">
                {c.completion_date ?? formatMonthLabel(c.month)}
              </div>
            </a>
          ))}
        </div>
      )}
    </WidgetWrapper>
  )
}
