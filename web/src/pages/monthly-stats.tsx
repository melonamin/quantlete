import { useState, useMemo } from 'react'
import { useMonthlyStats, useSportTypeStats } from '@/lib/api'
import { formatDistance, formatDuration } from '@/lib/format'
import { AccordionTable, type AccordionTableColumn, type AccordionTableGroup } from '@/components/ui/accordion-table'
import { Skeleton } from '@/components/ui/skeleton'
import { ChevronLeft, ChevronRight, Download } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface MonthlySummary {
  month: string
  label: string
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

interface SportBreakdown {
  sport_type: string
  activity_count: number
  total_distance: number
  total_time: number
  total_elevation: number
}

function formatMonthLabel(month: string): string {
  const [y, m] = month.split('-').map(Number)
  if (!y || !m) return month
  return new Date(y, m - 1, 1).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'long',
  })
}

export function MonthlyStatsPage() {
  const currentYear = new Date().getFullYear()
  const [year, setYear] = useState(currentYear)

  const { data: monthlyData, isLoading } = useMonthlyStats(year)
  const { data: _sportTypeData } = useSportTypeStats()

  // Build groups with mock sport breakdown (in production this would come from the API)
  const groups: AccordionTableGroup<MonthlySummary, SportBreakdown>[] = useMemo(() => {
    if (!monthlyData) return []

    return monthlyData
      .sort((a, b) => b.month.localeCompare(a.month)) // Most recent first
      .map((stat) => ({
        id: stat.month,
        label: formatMonthLabel(stat.month),
        summary: {
          ...stat,
          label: formatMonthLabel(stat.month),
        },
        // Sport breakdown would ideally come from the API
        // For now we show empty children
        children: [] as SportBreakdown[],
      }))
  }, [monthlyData])

  const summaryColumns: AccordionTableColumn<MonthlySummary>[] = [
    {
      key: 'month',
      header: 'Month',
      render: (row) => <span className="font-medium">{row.label}</span>,
    },
    {
      key: 'activities',
      header: 'Activities',
      render: (row) => row.activity_count,
      className: 'text-right',
      headerClassName: 'text-right',
    },
    {
      key: 'distance',
      header: 'Distance',
      render: (row) => formatDistance(row.total_distance),
      className: 'text-right',
      headerClassName: 'text-right',
    },
    {
      key: 'time',
      header: 'Time',
      render: (row) => formatDuration(row.total_time),
      className: 'text-right',
      headerClassName: 'text-right',
    },
    {
      key: 'elevation',
      header: 'Elevation',
      render: (row) => `${row.total_elevation.toLocaleString()}m`,
      className: 'text-right',
      headerClassName: 'text-right',
    },
  ]

  const childColumns: AccordionTableColumn<SportBreakdown>[] = [
    {
      key: 'sport',
      header: 'Sport',
      render: (row) => row.sport_type,
    },
    {
      key: 'activities',
      header: 'Activities',
      render: (row) => row.activity_count,
      className: 'text-right',
    },
    {
      key: 'distance',
      header: 'Distance',
      render: (row) => formatDistance(row.total_distance),
      className: 'text-right',
    },
    {
      key: 'time',
      header: 'Time',
      render: (row) => formatDuration(row.total_time),
      className: 'text-right',
    },
    {
      key: 'elevation',
      header: 'Elevation',
      render: (row) => `${row.total_elevation.toLocaleString()}m`,
      className: 'text-right',
    },
  ]

  // Calculate yearly totals
  const yearlyTotals = useMemo(() => {
    if (!monthlyData) return null
    return monthlyData.reduce(
      (acc, stat) => ({
        activity_count: acc.activity_count + stat.activity_count,
        total_distance: acc.total_distance + stat.total_distance,
        total_time: acc.total_time + stat.total_time,
        total_elevation: acc.total_elevation + stat.total_elevation,
      }),
      { activity_count: 0, total_distance: 0, total_time: 0, total_elevation: 0 }
    )
  }, [monthlyData])

  const handleExportCSV = () => {
    if (!monthlyData) return

    const headers = ['Month', 'Activities', 'Distance (m)', 'Time (s)', 'Elevation (m)']
    const rows = monthlyData.map((stat) => [
      stat.month,
      stat.activity_count,
      stat.total_distance,
      stat.total_time,
      stat.total_elevation,
    ])

    const csv = [headers.join(','), ...rows.map((r) => r.join(','))].join('\n')
    const blob = new Blob([csv], { type: 'text/csv' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `monthly-stats-${year}.csv`
    a.click()
    URL.revokeObjectURL(url)
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="mb-8 flex items-start justify-between">
        <div>
          <h1 className="text-2xl font-bold">Monthly Statistics</h1>
          <p className="text-muted-foreground">Detailed monthly breakdown of your activities</p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" onClick={() => setYear((y) => y - 1)}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <span className="w-16 text-center font-medium">{year}</span>
          <Button
            variant="outline"
            size="icon"
            onClick={() => setYear((y) => y + 1)}
            disabled={year >= currentYear}
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
          <Button variant="outline" size="sm" onClick={handleExportCSV} disabled={!monthlyData}>
            <Download className="h-4 w-4 mr-1" />
            Export
          </Button>
        </div>
      </div>

      {/* Yearly summary */}
      {yearlyTotals && (
        <div className="mb-6 grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="rounded-lg border border-border p-4">
            <div className="text-sm text-muted-foreground">Total Activities</div>
            <div className="text-2xl font-bold">{yearlyTotals.activity_count}</div>
          </div>
          <div className="rounded-lg border border-border p-4">
            <div className="text-sm text-muted-foreground">Total Distance</div>
            <div className="text-2xl font-bold">{formatDistance(yearlyTotals.total_distance)}</div>
          </div>
          <div className="rounded-lg border border-border p-4">
            <div className="text-sm text-muted-foreground">Total Time</div>
            <div className="text-2xl font-bold">{formatDuration(yearlyTotals.total_time)}</div>
          </div>
          <div className="rounded-lg border border-border p-4">
            <div className="text-sm text-muted-foreground">Total Elevation</div>
            <div className="text-2xl font-bold">
              {yearlyTotals.total_elevation.toLocaleString()}m
            </div>
          </div>
        </div>
      )}

      {/* Monthly breakdown table */}
      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 12 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      ) : (
        <div className="rounded-lg border border-border bg-card p-4">
          <AccordionTable
            columns={summaryColumns}
            childColumns={childColumns}
            groups={groups}
            emptyMessage="No activities recorded for this year."
          />
        </div>
      )}
    </div>
  )
}
