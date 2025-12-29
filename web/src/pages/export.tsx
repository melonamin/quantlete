import { useState } from 'react'
import { useExportStats, getExportCSVUrl, getExportJSONUrl } from '@/lib/api/export'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Download, FileJson, FileSpreadsheet, Calendar, Activity } from 'lucide-react'
import { formatDate } from '@/lib/format'

export function ExportPage() {
  const { data: stats, isLoading } = useExportStats()
  const [format, setFormat] = useState<'csv' | 'json'>('csv')
  const [after, setAfter] = useState('')
  const [before, setBefore] = useState('')
  const [sportType, setSportType] = useState('')

  const handleExport = () => {
    const params = {
      after: after || undefined,
      before: before || undefined,
      sport_type: sportType || undefined,
    }

    const url = format === 'csv' ? getExportCSVUrl(params) : getExportJSONUrl(params)

    // Trigger download
    const link = document.createElement('a')
    link.href = url
    link.download = `stata-activities.${format}`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Export Data</h1>
        <p className="text-muted-foreground">Download your activity data in various formats</p>
      </div>

      <div className="grid gap-6 md:grid-cols-2">
        {/* Stats Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Your Data
            </CardTitle>
            <CardDescription>Overview of available data for export</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-2">
                <div className="h-4 w-24 animate-pulse rounded bg-muted" />
                <div className="h-4 w-32 animate-pulse rounded bg-muted" />
              </div>
            ) : stats ? (
              <div className="space-y-4">
                <div className="flex items-center justify-between border-b border-border pb-3">
                  <span className="text-muted-foreground">Total Activities</span>
                  <span className="text-xl font-bold">{stats.total_activities.toLocaleString()}</span>
                </div>
                {stats.first_activity && (
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">First Activity</span>
                    <span className="text-sm">{formatDate(stats.first_activity)}</span>
                  </div>
                )}
                {stats.last_activity && (
                  <div className="flex items-center justify-between">
                    <span className="text-muted-foreground">Last Activity</span>
                    <span className="text-sm">{formatDate(stats.last_activity)}</span>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-muted-foreground">No data available</p>
            )}
          </CardContent>
        </Card>

        {/* Export Options Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5" />
              Export Options
            </CardTitle>
            <CardDescription>Choose format and filters for your export</CardDescription>
          </CardHeader>
          <CardContent className="space-y-6">
            {/* Format Selection */}
            <div>
              <label className="mb-2 block text-sm font-medium">Format</label>
              <div className="flex gap-2">
                <Button
                  variant={format === 'csv' ? 'default' : 'outline'}
                  onClick={() => setFormat('csv')}
                  className="flex-1"
                >
                  <FileSpreadsheet className="mr-2 h-4 w-4" />
                  CSV
                </Button>
                <Button
                  variant={format === 'json' ? 'default' : 'outline'}
                  onClick={() => setFormat('json')}
                  className="flex-1"
                >
                  <FileJson className="mr-2 h-4 w-4" />
                  JSON
                </Button>
              </div>
              <p className="mt-1 text-xs text-muted-foreground">
                {format === 'csv'
                  ? 'Best for spreadsheets (Excel, Google Sheets)'
                  : 'Best for developers and data analysis'}
              </p>
            </div>

            {/* Date Range */}
            <div>
              <label className="mb-2 block text-sm font-medium">
                <Calendar className="mr-1 inline h-4 w-4" />
                Date Range (Optional)
              </label>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <label className="mb-1 block text-xs text-muted-foreground">From</label>
                  <input
                    type="date"
                    className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                    value={after}
                    onChange={(e) => setAfter(e.target.value)}
                  />
                </div>
                <div>
                  <label className="mb-1 block text-xs text-muted-foreground">To</label>
                  <input
                    type="date"
                    className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                    value={before}
                    onChange={(e) => setBefore(e.target.value)}
                  />
                </div>
              </div>
            </div>

            {/* Sport Type */}
            <div>
              <label className="mb-2 block text-sm font-medium">Sport Type (Optional)</label>
              <input
                type="text"
                placeholder="e.g., Ride, Run, Swim"
                className="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"
                value={sportType}
                onChange={(e) => setSportType(e.target.value)}
              />
            </div>

            {/* Export Button */}
            <Button onClick={handleExport} className="w-full" size="lg">
              <Download className="mr-2 h-4 w-4" />
              Download {format.toUpperCase()}
            </Button>
          </CardContent>
        </Card>
      </div>

      {/* Export Info */}
      <Card className="mt-6">
        <CardHeader>
          <CardTitle>What's Included</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-3">
            <div>
              <h4 className="font-medium">Activity Details</h4>
              <ul className="mt-2 space-y-1 text-sm text-muted-foreground">
                <li>• Name and description</li>
                <li>• Sport type</li>
                <li>• Start date and timezone</li>
                <li>• Duration (moving/elapsed)</li>
              </ul>
            </div>
            <div>
              <h4 className="font-medium">Performance Metrics</h4>
              <ul className="mt-2 space-y-1 text-sm text-muted-foreground">
                <li>• Distance and elevation</li>
                <li>• Speed (avg/max)</li>
                <li>• Heart rate (avg/max)</li>
                <li>• Power (avg/max/weighted)</li>
              </ul>
            </div>
            <div>
              <h4 className="font-medium">Location & Gear</h4>
              <ul className="mt-2 space-y-1 text-sm text-muted-foreground">
                <li>• Start coordinates</li>
                <li>• City and country</li>
                <li>• Gear ID</li>
                <li>• Commute/Trainer flags</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
