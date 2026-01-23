import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from './echarts-wrapper'
import {
  chartColors,
  defaultGridConfig,
  defaultTooltipConfig,
  inlineLegendConfig,
  maxInlineLegendItems,
  scrollLegendConfig,
} from './chart-constants'

interface DataPoint {
  x: number | string
  y: number
}

interface LineChartSeries {
  name: string
  data: DataPoint[]
  color?: string
  areaStyle?: boolean
  smooth?: boolean
  step?: 'start' | 'middle' | 'end' | false
}

interface LineChartProps {
  series: LineChartSeries[]
  xAxisType?: 'category' | 'value' | 'time'
  xAxisLabel?: string
  yAxisLabel?: string
  height?: number | string
  loading?: boolean
  showLegend?: boolean
  showDataZoom?: boolean
  className?: string
}

export function LineChart({
  series,
  xAxisType = 'category',
  xAxisLabel,
  yAxisLabel,
  height = 300,
  loading = false,
  showLegend = true,
  showDataZoom = false,
  className,
}: LineChartProps) {
  const hasLegend = showLegend && series.length > 1
  const needsScrollLegend = hasLegend && series.length > maxInlineLegendItems

  // Calculate proper grid bottom spacing for legend + dataZoom combo
  const gridBottom = (() => {
    if (hasLegend && showDataZoom) return '28%'
    if (hasLegend) return '15%'
    if (showDataZoom) return '15%'
    return '3%'
  })()

  const option: EChartsOption = {
    grid: { ...defaultGridConfig, bottom: gridBottom },
    tooltip: {
      ...defaultTooltipConfig,
      axisPointer: {
        type: 'cross',
      },
    },
    legend: hasLegend
      ? {
          data: series.map((s) => s.name),
          bottom: 0,
          ...(needsScrollLegend ? scrollLegendConfig : inlineLegendConfig),
        }
      : undefined,
    xAxis: {
      type: xAxisType,
      name: xAxisLabel,
      nameLocation: 'middle',
      nameGap: 30,
      data: xAxisType === 'category' ? series[0]?.data.map((d) => d.x) : undefined,
      axisLine: {
        lineStyle: { color: '#666' },
      },
      axisLabel: {
        color: '#888',
      },
    },
    yAxis: {
      type: 'value',
      name: yAxisLabel,
      nameLocation: 'middle',
      nameGap: 50,
      axisLine: {
        lineStyle: { color: '#666' },
      },
      axisLabel: {
        color: '#888',
      },
      splitLine: {
        lineStyle: { color: 'rgba(128, 128, 128, 0.2)' },
      },
    },
    dataZoom: showDataZoom
      ? [
          {
            type: 'inside',
            start: 0,
            end: 100,
          },
          {
            type: 'slider',
            start: 0,
            end: 100,
            height: 20,
            bottom: hasLegend ? (needsScrollLegend ? 65 : 45) : 10,
          },
        ]
      : undefined,
    series: series.map((s, idx) => ({
      name: s.name,
      type: 'line',
      data: xAxisType === 'category' ? s.data.map((d) => d.y) : s.data.map((d) => [d.x, d.y]),
      smooth: s.smooth ?? true,
      step: s.step,
      symbol: 'none',
      lineStyle: {
        color: s.color ?? Object.values(chartColors)[idx % Object.values(chartColors).length],
        width: 2,
      },
      areaStyle: s.areaStyle
        ? {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: (s.color ?? chartColors.primary) + '40' },
                { offset: 1, color: (s.color ?? chartColors.primary) + '05' },
              ],
            },
          }
        : undefined,
    })),
  }

  return <EChartsWrapper option={option} height={height} loading={loading} className={className} />
}
