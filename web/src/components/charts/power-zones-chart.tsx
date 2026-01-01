import { StackedBarChart } from './bar-chart'
import { zoneColorsArray } from './chart-constants'

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
    color: zoneColorsArray[idx],
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
