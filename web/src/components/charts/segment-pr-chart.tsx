import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from './echarts-wrapper'
import {
  chartColors,
  defaultGridConfig,
  defaultTooltipConfig,
} from './chart-constants'
import type { SegmentEffort } from '@/lib/api'

function formatElapsed(seconds: number) {
  if (!Number.isFinite(seconds) || seconds <= 0) return '–'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.round(seconds % 60)
  if (h > 0) return `${h}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`
  return `${m}:${String(s).padStart(2, '0')}`
}

export function SegmentPRChart({
  efforts,
  height = 260,
  loading = false,
  className,
}: {
  efforts: SegmentEffort[]
  height?: number | string
  loading?: boolean
  className?: string
}) {
  const data = (() => {
    const sorted = [...efforts]
      .filter((e) => e.start_date_local || e.start_date)
      .sort((a, b) => {
        const ad = new Date(a.start_date_local ?? a.start_date ?? '').getTime()
        const bd = new Date(b.start_date_local ?? b.start_date ?? '').getTime()
        return ad - bd
      })

    let best = Number.POSITIVE_INFINITY
    const points: [number, number][] = []
    for (const e of sorted) {
      const d = new Date(e.start_date_local ?? e.start_date ?? '').getTime()
      if (!Number.isFinite(d)) continue
      if (e.elapsed_time > 0 && e.elapsed_time < best) {
        best = e.elapsed_time
      }
      if (Number.isFinite(best)) {
        points.push([d, best])
      }
    }
    return points
  })()

  const option: EChartsOption = {
    grid: defaultGridConfig,
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'axis',
      formatter: (params: unknown) => {
        const ps = params as Array<{ value: [number, number] }>
        const v = ps?.[0]?.value
        if (!v) return ''
        const date = new Date(v[0]).toLocaleDateString(undefined, {
          year: 'numeric',
          month: 'short',
          day: 'numeric',
        })
        return `${date}<br/>Best: ${formatElapsed(v[1])}`
      },
    },
    xAxis: {
      type: 'time',
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888' },
    },
    yAxis: {
      type: 'value',
      inverse: true,
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: {
        color: '#888',
        formatter: (v: number) => formatElapsed(v),
      },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.2)' } },
    },
    series: [
      {
        name: 'Best time',
        type: 'line',
        step: 'end',
        data,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: { color: chartColors.primary, width: 2 },
        itemStyle: { color: chartColors.primary },
        emphasis: { focus: 'series' },
        encode: { x: 0, y: 1 },
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} className={className} />
}
