import { useQuery } from '@tanstack/react-query'
import { Activity, TrendingUp, Clock, Mountain } from 'lucide-react'

interface HealthResponse {
  status: string
}

export function DashboardPage() {
  const { data: health, isLoading, error } = useQuery<HealthResponse>({
    queryKey: ['health'],
    queryFn: async () => {
      const res = await fetch('/api/v1/health')
      if (!res.ok) throw new Error('Failed to fetch health')
      return res.json()
    },
  })

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-muted-foreground">Your activity statistics at a glance</p>
      </div>

      {/* Stats overview */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard
          title="Total Activities"
          value="--"
          icon={Activity}
          description="All time"
        />
        <StatCard
          title="Total Distance"
          value="-- km"
          icon={TrendingUp}
          description="All time"
        />
        <StatCard
          title="Total Time"
          value="-- hours"
          icon={Clock}
          description="All time"
        />
        <StatCard
          title="Total Elevation"
          value="-- m"
          icon={Mountain}
          description="All time"
        />
      </div>

      {/* API Status */}
      <div className="rounded-lg border border-border bg-card p-6">
        <h2 className="text-lg font-semibold mb-4">System Status</h2>
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">API:</span>
          {isLoading ? (
            <span className="text-sm text-muted-foreground">Checking...</span>
          ) : error ? (
            <span className="text-sm text-destructive">Error connecting to API</span>
          ) : health ? (
            <span className="inline-flex items-center gap-1.5">
              <span className="h-2 w-2 rounded-full bg-green-500" />
              <span className="text-sm text-green-600">{health.status}</span>
            </span>
          ) : null}
        </div>

        <div className="mt-4 pt-4 border-t border-border">
          <h3 className="text-sm font-medium mb-2">Getting Started</h3>
          <ol className="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
            <li>
              <a href="/api/v1/auth/strava" className="text-strava hover:underline">
                Authenticate with Strava
              </a>
            </li>
            <li>Import your activities</li>
            <li>Explore your statistics</li>
          </ol>
        </div>
      </div>
    </div>
  )
}

interface StatCardProps {
  title: string
  value: string
  icon: React.ComponentType<{ className?: string }>
  description: string
}

function StatCard({ title, value, icon: Icon, description }: StatCardProps) {
  return (
    <div className="rounded-lg border border-border bg-card p-6">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-muted-foreground">{title}</span>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <div className="mt-2">
        <span className="text-2xl font-bold">{value}</span>
      </div>
      <p className="text-xs text-muted-foreground mt-1">{description}</p>
    </div>
  )
}
