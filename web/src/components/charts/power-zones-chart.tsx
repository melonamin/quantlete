import { StackedBarChart } from './bar-chart'

const ZONE_COLORS = ['#cbd5e1', '#60a5fa', '#22c55e', '#f59e0b', '#ef4444']

export function PowerZonesChart({
  secondsByZone,
  height = 120,
  loading = false,
}: {
  secondsByZone: number[]
  height?: number | string
  loading?: boolean
}) {
  const series = secondsByZone.slice(0, 5).map((s, idx) => ({
    name: `Z${idx + 1}`,
    data: [s],
    color: ZONE_COLORS[idx],
  }))

  return (
    <StackedBarChart
      categories={['']}
      series={series}
      horizontal
      height={height}
      loading={loading}
    />
  )
}
