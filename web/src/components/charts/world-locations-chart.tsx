import type { EChartsOption, TooltipComponentFormatterCallbackParams } from 'echarts'
import { EChartsWrapper, defaultGridConfig, defaultTooltipConfig } from './echarts-wrapper'

type LocationValue = [number, number, number] // [lng, lat, count]

export function WorldLocationsChart({
  points,
  height = 360,
  loading = false,
}: {
  points: { lat: number; lng: number; count: number }[]
  height?: number | string
  loading?: boolean
}) {
  const data = points.map((p) => ({
    value: [p.lng, p.lat, p.count] as LocationValue,
  }))

  const option: EChartsOption = {
    grid: { ...defaultGridConfig, left: '6%', right: '6%', top: 20, bottom: 30 },
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'item',
      formatter: (params: TooltipComponentFormatterCallbackParams) => {
        if (Array.isArray(params)) return ''
        const v = params?.value as LocationValue | undefined
        if (!v) return ''
        return `Lat ${v[1].toFixed(2)}, Lng ${v[0].toFixed(2)}<br/>Activities: ${v[2]}`
      },
    },
    xAxis: {
      type: 'value',
      min: -180,
      max: 180,
      axisLabel: { show: false },
      axisLine: { lineStyle: { color: '#555' } },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    },
    yAxis: {
      type: 'value',
      min: -90,
      max: 90,
      axisLabel: { show: false },
      axisLine: { lineStyle: { color: '#555' } },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    },
    series: [
      {
        type: 'effectScatter',
        coordinateSystem: 'cartesian2d',
        data,
        symbolSize: (val: number | number[]) => {
          const c = Array.isArray(val) ? Number(val[2]) : 1
          const s = Math.sqrt(Math.max(1, c)) * 3
          return Math.max(4, Math.min(28, s))
        },
        rippleEffect: { scale: 2.5 },
        itemStyle: { color: '#3b82f6' },
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} />
}
