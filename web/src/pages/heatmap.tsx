import { useHeatmapData } from '@/lib/api'
import { Heatmap } from '@/components/maps'
import { Skeleton } from '@/components/ui/skeleton'
import { Link } from '@tanstack/react-router'

export function HeatmapPage() {
  const { data, isLoading, error } = useHeatmapData()

  return (
    <div className="flex flex-col h-[calc(100vh-3.5rem)]">
      <div className="px-4 py-4 border-b border-border bg-background">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Heatmap</h1>
            <p className="text-muted-foreground">
              Visualize all your activities on a map
            </p>
          </div>
          {data && (
            <div className="text-sm text-muted-foreground">
              {data.total} {data.total === 1 ? 'route' : 'routes'}
            </div>
          )}
        </div>
      </div>

      <div className="flex-1">
        {isLoading ? (
          <Skeleton className="h-full w-full" />
        ) : error ? (
          <div className="h-full flex items-center justify-center">
            <div className="text-center">
              <p className="text-destructive mb-2">Failed to load heatmap data</p>
              <p className="text-sm text-muted-foreground">
                Make sure you are{' '}
                <Link to="/settings" className="underline">
                  connected to Strava
                </Link>
              </p>
            </div>
          </div>
        ) : data && data.activities.length > 0 ? (
          <Heatmap activities={data.activities} className="h-full" />
        ) : (
          <div className="h-full flex items-center justify-center">
            <div className="text-center">
              <p className="text-muted-foreground mb-2">
                No activities with GPS data found
              </p>
              <p className="text-sm text-muted-foreground">
                Import activities with GPS data to see them on the heatmap
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
