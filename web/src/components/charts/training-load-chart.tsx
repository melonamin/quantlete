import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from './echarts-wrapper'
import {
  chartColors,
  defaultGridConfig,
  defaultTooltipConfig,
} from './chart-constants'

export interface TrainingLoadPoint {
  day: string
  tss: number
  ctl: number
  atl: number
  tsb: number
}

export function TrainingLoadChart({
  data,
  height = 360,
  loading = false,
}: {
  data: TrainingLoadPoint[]
  height?: number | string
  loading?: boolean
}) {
  const days = data.map((d) => d.day)
  const option: EChartsOption = {
    grid: { ...defaultGridConfig, bottom: '18%' },
    tooltip: { ...defaultTooltipConfig, trigger: 'axis', axisPointer: { type: 'cross' } },
    legend: { bottom: 0, textStyle: { color: '#888' } },
    xAxis: {
      type: 'category',
      data: days,
      axisLabel: { color: '#888' },
      axisLine: { lineStyle: { color: '#666' } },
    },
    yAxis: {
      type: 'value',
      axisLabel: { color: '#888' },
      axisLine: { lineStyle: { color: '#666' } },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    },
    dataZoom: [
      { type: 'inside', start: 0, end: 100 },
      { type: 'slider', start: 0, end: 100, height: 20, bottom: 30 },
    ],
    series: [
      {
        name: 'Daily TSS',
        type: 'bar',
        data: data.map((d) => d.tss),
        itemStyle: { color: chartColors.info + '66' },
      },
      {
        name: 'CTL',
        type: 'line',
        data: data.map((d) => d.ctl),
        showSymbol: false,
        lineStyle: { color: chartColors.success, width: 2 },
      },
      {
        name: 'ATL',
        type: 'line',
        data: data.map((d) => d.atl),
        showSymbol: false,
        lineStyle: { color: chartColors.warning, width: 2 },
      },
      {
        name: 'TSB',
        type: 'line',
        data: data.map((d) => d.tsb),
        showSymbol: false,
        lineStyle: { color: chartColors.primary, width: 2 },
        areaStyle: { color: chartColors.primary + '11' },
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} />
}
