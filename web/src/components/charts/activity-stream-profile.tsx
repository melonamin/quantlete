import type { EChartsOption } from 'echarts'
import {
  EChartsWrapper,
  chartColors,
  defaultGridConfig,
  defaultTooltipConfig,
} from './echarts-wrapper'

type StreamMap = Record<string, unknown>

function toNumberArray(v: unknown): number[] {
  if (!Array.isArray(v)) return []
  return v.map((x) => (typeof x === 'number' ? x : Number(x))).filter((n) => Number.isFinite(n))
}

export type StreamProfileSeriesKey = 'heartrate' | 'watts' | 'cadence' | 'altitude'

export function ActivityStreamProfileChart({
  streams,
  enabledSeries,
  height = 320,
  loading = false,
}: {
  streams: StreamMap
  enabledSeries?: Partial<Record<StreamProfileSeriesKey, boolean>>
  height?: number | string
  loading?: boolean
}) {
  const distance = toNumberArray(streams.distance)
  const time = toNumberArray(streams.time)
  const x = distance.length ? distance.map((m) => m / 1000) : time
  const xLabel = distance.length ? 'Distance (km)' : 'Time (s)'

  const hr = toNumberArray(streams.heartrate)
  const watts = toNumberArray(streams.watts)
  const cadence = toNumberArray(streams.cadence)
  const altitude = toNumberArray(streams.altitude)

  const series: NonNullable<EChartsOption['series']> = []
  const yAxis: NonNullable<EChartsOption['yAxis']> = []

  const addAxis = (name: string, position: 'left' | 'right', offset: number, color: string) => {
    const idx = Array.isArray(yAxis) ? yAxis.length : 0
    ;(yAxis as any[]).push({
      type: 'value',
      name,
      position,
      offset,
      axisLine: { lineStyle: { color } },
      axisLabel: { color: '#888' },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    })
    return idx
  }

  if (enabledSeries?.heartrate !== false && hr.length && x.length) {
    const idx = addAxis('HR (bpm)', 'left', 0, chartColors.heartRate)
    series.push({
      name: 'Heart rate',
      type: 'line',
      yAxisIndex: idx,
      xAxisIndex: 0,
      data: x.map((xx, i) => [xx, hr[i] ?? null]),
      showSymbol: false,
      lineStyle: { color: chartColors.heartRate, width: 2 },
    })
  }
  if (enabledSeries?.watts !== false && watts.length && x.length) {
    const idx = addAxis('Power (W)', 'right', 0, chartColors.power)
    series.push({
      name: 'Power',
      type: 'line',
      yAxisIndex: idx,
      xAxisIndex: 0,
      data: x.map((xx, i) => [xx, watts[i] ?? null]),
      showSymbol: false,
      lineStyle: { color: chartColors.power, width: 2 },
    })
  }
  if (enabledSeries?.cadence !== false && cadence.length && x.length) {
    const idx = addAxis('Cadence', 'right', 50, chartColors.cadence)
    series.push({
      name: 'Cadence',
      type: 'line',
      yAxisIndex: idx,
      xAxisIndex: 0,
      data: x.map((xx, i) => [xx, cadence[i] ?? null]),
      showSymbol: false,
      lineStyle: { color: chartColors.cadence, width: 2 },
    })
  }
  if (enabledSeries?.altitude !== false && altitude.length && x.length) {
    const idx = addAxis('Elevation (m)', 'left', 50, chartColors.elevation)
    series.push({
      name: 'Elevation',
      type: 'line',
      yAxisIndex: idx,
      xAxisIndex: 0,
      data: x.map((xx, i) => [xx, altitude[i] ?? null]),
      showSymbol: false,
      lineStyle: { color: chartColors.elevation, width: 2 },
      areaStyle: {
        color: chartColors.elevation + '22',
      },
    })
  }

  const option: EChartsOption = {
    grid: {
      ...defaultGridConfig,
      left: '5%',
      right: '8%',
      top: 25,
      bottom: 35,
    },
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'axis',
      axisPointer: { type: 'cross' },
    },
    legend: {
      bottom: 0,
      textStyle: { color: '#888' },
    },
    xAxis: {
      type: 'value',
      name: xLabel,
      nameLocation: 'middle',
      nameGap: 28,
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888' },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.08)' } },
    },
    yAxis: yAxis.length ? yAxis : { type: 'value' },
    series,
  }

  return <EChartsWrapper option={option} height={height} loading={loading} />
}
