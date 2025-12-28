import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState, useEffect } from 'react'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
})

interface HealthResponse {
  status: string
}

function App() {
  const [health, setHealth] = useState<HealthResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/v1/health')
      .then((res) => res.json())
      .then((data: HealthResponse) => setHealth(data))
      .catch((err: Error) => setError(err.message))
  }, [])

  return (
    <QueryClientProvider client={queryClient}>
      <div className="min-h-screen bg-background text-foreground">
        <header className="border-b border-border">
          <div className="container mx-auto px-4 py-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="h-8 w-8 rounded-lg bg-strava flex items-center justify-center">
                  <span className="text-white font-bold text-sm">S</span>
                </div>
                <h1 className="text-xl font-semibold">Stata</h1>
              </div>
              <nav className="flex items-center gap-6">
                <a href="#" className="text-sm text-muted-foreground hover:text-foreground">
                  Dashboard
                </a>
                <a href="#" className="text-sm text-muted-foreground hover:text-foreground">
                  Activities
                </a>
                <a href="#" className="text-sm text-muted-foreground hover:text-foreground">
                  Settings
                </a>
              </nav>
            </div>
          </div>
        </header>

        <main className="container mx-auto px-4 py-8">
          <div className="max-w-2xl mx-auto">
            <div className="rounded-lg border border-border bg-card p-6">
              <h2 className="text-lg font-semibold mb-4">Statistics for Strava</h2>
              <p className="text-muted-foreground mb-6">
                Your personal analytics dashboard for Strava activities. Import your activities and
                explore detailed statistics, visualizations, and insights.
              </p>

              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">API Status:</span>
                  {error ? (
                    <span className="text-sm text-destructive">{error}</span>
                  ) : health ? (
                    <span className="inline-flex items-center gap-1.5">
                      <span className="h-2 w-2 rounded-full bg-green-500" />
                      <span className="text-sm text-green-600">{health.status}</span>
                    </span>
                  ) : (
                    <span className="text-sm text-muted-foreground">Checking...</span>
                  )}
                </div>

                <div className="pt-4 border-t border-border">
                  <h3 className="text-sm font-medium mb-2">Getting Started</h3>
                  <ol className="text-sm text-muted-foreground space-y-1 list-decimal list-inside">
                    <li>Configure your Strava API credentials</li>
                    <li>Authenticate with Strava</li>
                    <li>Import your activities</li>
                    <li>Explore your statistics</li>
                  </ol>
                </div>
              </div>
            </div>
          </div>
        </main>

        <footer className="border-t border-border mt-auto">
          <div className="container mx-auto px-4 py-4">
            <p className="text-sm text-muted-foreground text-center">
              Stata - Statistics for Strava
            </p>
          </div>
        </footer>
      </div>
    </QueryClientProvider>
  )
}

export default App
