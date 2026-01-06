import type { EChartsOption } from 'echarts'
import type { WeeklyZoneDistribution } from '@/lib/wasm/types.gen'
import { EChartsWrapper } from './echarts-wrapper'
import { zoneColorsArray, defaultGridConfig, defaultTooltipConfig } from './chart-constants'

const zoneLabels = ['Z1 Recovery', 'Z2 Endurance', 'Z3 Tempo', 'Z4 Threshold', 'Z5 VO2max']

export function ZoneTrendChart({
  data,
  height = 360,
  loading = false,
}: {
  data: WeeklyZoneDistribution[]
  height?: number | string
  loading?: boolean
}) {
  const weeks = data.map((d) => d.week)

  const option: EChartsOption = {
    grid: { ...defaultGridConfig, bottom: '18%' },
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: unknown) => {
        const items = params as Array<{
          seriesName: string
          value: number
          color: string
          axisValue: string
        }>
        if (!items?.length) return ''

        const week = items[0].axisValue
        const lines = items.map(
          (item) =>
            `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${item.color};margin-right:5px"></span>${item.seriesName}: <b>${item.value.toFixed(1)}%</b>`
        )
        return `<div style="font-weight:600;margin-bottom:4px">${week}</div>${lines.join('<br/>')}`
      },
    },
    legend: {
      bottom: 0,
      textStyle: { color: '#888' },
      data: zoneLabels,
    },
    xAxis: {
      type: 'category',
      data: weeks,
      axisLabel: {
        color: '#888',
        rotate: 45,
        formatter: (value: string) => {
          // Format "2024-W01" to just "W01" for brevity
          const match = value.match(/W(\d+)$/)
          return match ? `W${match[1]}` : value
        },
      },
      axisLine: { lineStyle: { color: '#666' } },
    },
    yAxis: {
      type: 'value',
      min: 0,
      max: 100,
      axisLabel: {
        color: '#888',
        formatter: '{value}%',
      },
      axisLine: { lineStyle: { color: '#666' } },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    },
    dataZoom: [
      { type: 'inside', start: 0, end: 100 },
      { type: 'slider', start: 0, end: 100, height: 20, bottom: 30 },
    ],
    series: [
      {
        name: zoneLabels[0],
        type: 'line',
        stack: 'zones',
        areaStyle: { opacity: 0.8 },
        emphasis: { focus: 'series' },
        showSymbol: false,
        data: data.map((d) => d.percent_z1),
        itemStyle: { color: zoneColorsArray[0] },
        lineStyle: { width: 0 },
      },
      {
        name: zoneLabels[1],
        type: 'line',
        stack: 'zones',
        areaStyle: { opacity: 0.8 },
        emphasis: { focus: 'series' },
        showSymbol: false,
        data: data.map((d) => d.percent_z2),
        itemStyle: { color: zoneColorsArray[1] },
        lineStyle: { width: 0 },
      },
      {
        name: zoneLabels[2],
        type: 'line',
        stack: 'zones',
        areaStyle: { opacity: 0.8 },
        emphasis: { focus: 'series' },
        showSymbol: false,
        data: data.map((d) => d.percent_z3),
        itemStyle: { color: zoneColorsArray[2] },
        lineStyle: { width: 0 },
      },
      {
        name: zoneLabels[3],
        type: 'line',
        stack: 'zones',
        areaStyle: { opacity: 0.8 },
        emphasis: { focus: 'series' },
        showSymbol: false,
        data: data.map((d) => d.percent_z4),
        itemStyle: { color: zoneColorsArray[3] },
        lineStyle: { width: 0 },
      },
      {
        name: zoneLabels[4],
        type: 'line',
        stack: 'zones',
        areaStyle: { opacity: 0.8 },
        emphasis: { focus: 'series' },
        showSymbol: false,
        data: data.map((d) => d.percent_z5),
        itemStyle: { color: zoneColorsArray[4] },
        lineStyle: { width: 0 },
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} />
}
