import { Link } from '@tanstack/react-router'
import { useChallenges } from '@/lib/data'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { useRef } from 'react'

function formatMonthLabel(month?: string) {
  if (!month) return ''
  const [y, m] = month.split('-').map((x) => Number(x))
  if (!y || !m) return month
  const d = new Date(y, m - 1, 1)
  return d.toLocaleDateString(undefined, { year: 'numeric', month: 'short' })
}

export function RecentChallenges() {
  const { data, isLoading } = useChallenges()
  const items = (data ?? []).slice(0, 12)
  const scrollRef = useRef<HTMLDivElement>(null)

  const scroll = (direction: 'left' | 'right') => {
    if (!scrollRef.current) return
    const scrollAmount = 200
    scrollRef.current.scrollBy({
      left: direction === 'left' ? -scrollAmount : scrollAmount,
      behavior: 'smooth',
    })
  }

  return (
    <WidgetWrapper
      title="Recent Challenges"
      action={
        <div className="flex items-center gap-1">
          <button
            onClick={() => scroll('left')}
            className="p-1 rounded hover:bg-muted"
            aria-label="Scroll left"
          >
            <ChevronLeft className="h-4 w-4" />
          </button>
          <button
            onClick={() => scroll('right')}
            className="p-1 rounded hover:bg-muted"
            aria-label="Scroll right"
          >
            <ChevronRight className="h-4 w-4" />
          </button>
          <Button variant="ghost" size="sm" asChild>
            <Link to="/challenges">View all</Link>
          </Button>
        </div>
      }
      isLoading={isLoading}
    >
      {isLoading ? (
        <div className="flex gap-3 overflow-hidden">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="flex-shrink-0 w-[100px] space-y-2">
              <Skeleton className="h-[100px] w-full" />
              <Skeleton className="h-3 w-20" />
            </div>
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="text-sm text-muted-foreground">No challenges imported yet.</div>
      ) : (
        <div
          ref={scrollRef}
          className="flex gap-3 overflow-x-auto scrollbar-thin scrollbar-thumb-border scrollbar-track-transparent pb-2"
        >
          {items.map((c) => (
            <a
              key={c.id}
              href={c.slug ? `https://www.strava.com/challenges/${c.slug}` : undefined}
              target="_blank"
              rel="noreferrer"
              className="flex-shrink-0 w-[100px] rounded-md border border-border bg-background p-2 hover:bg-accent/30 transition-colors"
              title={c.name}
            >
              <div className="h-[80px] overflow-hidden rounded-md bg-muted">
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
              <div className="mt-2 line-clamp-2 text-xs font-medium leading-tight">{c.name}</div>
              <div className="mt-1 text-[10px] text-muted-foreground">
                {c.completion_date ?? formatMonthLabel(c.month)}
              </div>
            </a>
          ))}
        </div>
      )}
    </WidgetWrapper>
  )
}
