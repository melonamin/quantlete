import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from './echarts-wrapper'
import { chartColors, defaultGridConfig, defaultTooltipConfig } from './chart-constants'

function toNumberArray(v: unknown): number[] {
  if (!Array.isArray(v)) return []
  return v.map((x) => (typeof x === 'number' ? x : Number(x))).filter((n) => Number.isFinite(n))
}

export function ElevationProfileChart({
  distance,
  altitude,
  gradient = false,
  height = 220,
  loading = false,
}: {
  distance: unknown
  altitude: unknown
  gradient?: boolean
  height?: number | string
  loading?: boolean
}) {
  const dist = toNumberArray(distance)
  const alt = toNumberArray(altitude)
  const x = dist.map((m) => m / 1000)

  const min = alt.length ? Math.min(...alt) : 0
  const max = alt.length ? Math.max(...alt) : 0
  const avg = alt.length ? alt.reduce((a, b) => a + b, 0) / alt.length : 0

  const option: EChartsOption = {
    grid: { ...defaultGridConfig, top: 20, bottom: 30 },
    tooltip: { ...defaultTooltipConfig, trigger: 'axis' },
    xAxis: {
      type: 'value',
      name: 'Distance (km)',
      nameLocation: 'middle',
      nameGap: 25,
      axisLabel: { color: '#888' },
      axisLine: { lineStyle: { color: '#666' } },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#888' },
      axisLine: { lineStyle: { color: '#666' } },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    },
    series: [
      {
        type: 'line',
        name: 'Elevation',
        showSymbol: false,
        data: x.map((xx, i) => [xx, alt[i] ?? null]),
        lineStyle: { color: chartColors.elevation, width: 2 },
        areaStyle: gradient
          ? {
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: chartColors.elevation + '55' },
                  { offset: 1, color: chartColors.elevation + '05' },
                ],
              },
            }
          : { color: chartColors.elevation + '22' },
      },
    ],
    title: {
      show: true,
      left: 'center',
      top: 0,
      textStyle: { fontSize: 12, color: '#888', fontWeight: 'normal' },
      text: alt.length
        ? `Min ${Math.round(min)}m • Avg ${Math.round(avg)}m • Max ${Math.round(max)}m`
        : '',
    },
  }

  return <EChartsWrapper option={option} height={height} loading={loading} />
}
