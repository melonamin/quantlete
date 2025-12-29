import { Link } from '@tanstack/react-router'
import { useQueries } from '@tanstack/react-query'
import { useAppSettings, type EddingtonResult } from '@/lib/api'
import { useDataProviderStatus } from '@/lib/data'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'

const DEFAULT_DEFS = [
  { id: 'all', name: 'All activities', sport_types: undefined, show_in_dashboard_widget: true },
  {
    id: 'rides',
    name: 'Rides',
    sport_types: ['Ride', 'MountainBikeRide', 'GravelRide', 'EBikeRide', 'VirtualRide'],
    show_in_dashboard_widget: true,
  },
  {
    id: 'runs',
    name: 'Runs',
    sport_types: ['Run', 'TrailRun', 'VirtualRun'],
    show_in_dashboard_widget: true,
  },
]

export function EddingtonWidget() {
  const { data: settings } = useAppSettings()
  const { provider, initialized } = useDataProviderStatus()
  const defs = (
    settings?.eddington_definitions && settings.eddington_definitions.length
      ? settings.eddington_definitions
      : DEFAULT_DEFS
  ).filter((d) => d.show_in_dashboard_widget !== false)

  const queries = useQueries({
    queries: defs.map((d) => {
      const sportType = d.sport_types?.length ? d.sport_types.join(',') : undefined
      return {
        queryKey: ['data', 'eddington', 'def', d.id, sportType] as const,
        queryFn: async (): Promise<EddingtonResult> => {
          if (!provider) throw new Error('Provider not ready')
          return provider.getEddingtonData(sportType)
        },
        enabled: initialized && !!provider,
        staleTime: 60_000,
      }
    }),
  })

  const isLoading = queries.some((q) => q.isLoading)
  const hasError = queries.some((q) => q.isError)

  return (
    <WidgetWrapper
      title="Eddington"
      action={
        <Button variant="ghost" size="sm" asChild>
          <Link to="/eddington">View details</Link>
        </Button>
      }
      isLoading={isLoading}
    >
      {hasError ? (
        <p className="text-sm text-destructive">Failed to load Eddington.</p>
      ) : (
        <div className="space-y-2">
          {defs.map((d, idx) => (
            <div
              key={d.id}
              className="flex items-center justify-between rounded-md border border-border px-3 py-2"
            >
              <div className="text-sm font-medium">{d.name}</div>
              <div className="text-sm tabular-nums text-muted-foreground">
                {queries[idx].data?.number ?? '—'}
              </div>
            </div>
          ))}
          {defs.length === 0 && (
            <p className="text-sm text-muted-foreground">No definitions enabled.</p>
          )}
        </div>
      )}
    </WidgetWrapper>
  )
}
