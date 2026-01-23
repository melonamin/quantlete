import { Link } from '@tanstack/react-router'
import { useQueries } from '@tanstack/react-query'
import { useAppSettings, useEddingtonCompare, type EddingtonResult } from '@/lib/api'
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
  const { data: compareData, isLoading: compareLoading, isError: compareError } = useEddingtonCompare()

  // Get custom definitions that have custom sport types (not matching predefined groups)
  const customDefs = (
    settings?.eddington_definitions && settings.eddington_definitions.length
      ? settings.eddington_definitions
      : DEFAULT_DEFS
  ).filter((d) => {
    // Show in widget AND has custom sport types (not predefined sport groups)
    if (d.show_in_dashboard_widget === false) return false
    // Skip 'all' and predefined sport group IDs - these come from compareData
    if (d.id === 'all') return false
    // Only show truly custom definitions (not the default rides/runs which are covered by compare)
    const isDefaultDef = DEFAULT_DEFS.some(
      (def) => def.id === d.id ||
      (def.sport_types && d.sport_types &&
       def.sport_types.length === d.sport_types.length &&
       def.sport_types.every((t) => d.sport_types?.includes(t)))
    )
    return !isDefaultDef
  })

  // Fetch custom definitions that need individual queries
  const customQueries = useQueries({
    queries: customDefs.map((d) => {
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

  const isLoading = compareLoading || customQueries.some((q) => q.isLoading)
  const hasError = compareError || customQueries.some((q) => q.isError)

  // Combine sport groups from compare with custom definitions
  const sportGroupItems = compareData?.groups ?? []
  const hasData = compareData || customDefs.length > 0

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
      ) : !hasData ? (
        <p className="text-sm text-muted-foreground">No data available.</p>
      ) : (
        <div className="space-y-2">
          {/* All activities */}
          {compareData && (
            <div className="flex items-center justify-between rounded-md border border-border px-3 py-2">
              <div className="text-sm font-medium">All</div>
              <div className="text-sm tabular-nums text-muted-foreground">
                {compareData.all_number}
              </div>
            </div>
          )}

          {/* Sport groups from compare endpoint */}
          {sportGroupItems.map((item) => (
            <div
              key={item.sport_group}
              className="flex items-center justify-between rounded-md border border-border px-3 py-2"
            >
              <div className="text-sm font-medium">{item.name}</div>
              <div className="text-sm tabular-nums text-muted-foreground">
                {item.number}
              </div>
            </div>
          ))}

          {/* Custom definitions */}
          {customDefs.map((d, idx) => (
            <div
              key={d.id}
              className="flex items-center justify-between rounded-md border border-border px-3 py-2"
            >
              <div className="text-sm font-medium">{d.name}</div>
              <div className="text-sm tabular-nums text-muted-foreground">
                {customQueries[idx]?.data?.number ?? '—'}
              </div>
            </div>
          ))}

          {sportGroupItems.length === 0 && customDefs.length === 0 && !compareData && (
            <p className="text-sm text-muted-foreground">No definitions enabled.</p>
          )}
        </div>
      )}
    </WidgetWrapper>
  )
}
