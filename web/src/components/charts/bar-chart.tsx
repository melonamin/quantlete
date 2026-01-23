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

interface BarChartData {
  label: string
  value: number
  color?: string
}

interface BarChartProps {
  data: BarChartData[]
  horizontal?: boolean
  height?: number | string
  loading?: boolean
  showValues?: boolean
  barWidth?: number | string
  className?: string
  /** When set, all bars use this single color instead of cycling through chartColors */
  uniformColor?: string
}

export function BarChart({
  data,
  horizontal = false,
  height = 300,
  loading = false,
  showValues = false,
  barWidth = '60%',
  className,
  uniformColor,
}: BarChartProps) {
  const categoryAxis = {
    type: 'category' as const,
    data: data.map((d) => d.label),
    axisLine: {
      lineStyle: { color: '#666' },
    },
    axisLabel: {
      color: '#888',
      rotate: horizontal ? 0 : data.length > 6 ? 45 : 0,
    },
  }

  const valueAxis = {
    type: 'value' as const,
    axisLine: {
      lineStyle: { color: '#666' },
    },
    axisLabel: {
      color: '#888',
    },
    splitLine: {
      lineStyle: { color: 'rgba(128, 128, 128, 0.2)' },
    },
  }

  const option: EChartsOption = {
    grid: {
      ...defaultGridConfig,
      left: horizontal ? '15%' : '3%',
    },
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'item',
    },
    xAxis: horizontal ? valueAxis : categoryAxis,
    yAxis: horizontal ? categoryAxis : valueAxis,
    series: [
      {
        type: 'bar',
        data: data.map((d, idx) => ({
          value: d.value,
          itemStyle: {
            color:
              uniformColor ??
              d.color ??
              Object.values(chartColors)[idx % Object.values(chartColors).length],
            borderRadius: horizontal ? [0, 4, 4, 0] : [4, 4, 0, 0],
          },
        })),
        barWidth,
        label: showValues
          ? {
              show: true,
              position: horizontal ? 'right' : 'top',
              color: '#888',
              formatter: '{c}',
            }
          : undefined,
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} className={className} />
}

// Stacked bar chart variant
interface StackedBarChartSeries {
  name: string
  data: number[]
  color?: string
}

interface StackedBarChartProps {
  categories: string[]
  series: StackedBarChartSeries[]
  horizontal?: boolean
  height?: number | string
  loading?: boolean
  showLegend?: boolean
  className?: string
}

export function StackedBarChart({
  categories,
  series,
  horizontal = false,
  height = 300,
  loading = false,
  showLegend = true,
  className,
}: StackedBarChartProps) {
  const needsScrollLegend = series.length > maxInlineLegendItems

  const categoryAxis = {
    type: 'category' as const,
    data: categories,
    axisLine: {
      lineStyle: { color: '#666' },
    },
    axisLabel: {
      color: '#888',
    },
  }

  const valueAxis = {
    type: 'value' as const,
    axisLine: {
      lineStyle: { color: '#666' },
    },
    axisLabel: {
      color: '#888',
    },
    splitLine: {
      lineStyle: { color: 'rgba(128, 128, 128, 0.2)' },
    },
  }

  const option: EChartsOption = {
    grid: {
      ...defaultGridConfig,
      bottom: showLegend ? (needsScrollLegend ? '18%' : '15%') : '3%',
    },
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
    },
    legend: showLegend
      ? {
          data: series.map((s) => s.name),
          bottom: 0,
          ...(needsScrollLegend ? scrollLegendConfig : inlineLegendConfig),
        }
      : undefined,
    xAxis: horizontal ? valueAxis : categoryAxis,
    yAxis: horizontal ? categoryAxis : valueAxis,
    series: series.map((s, idx) => ({
      name: s.name,
      type: 'bar',
      stack: 'total',
      data: s.data,
      itemStyle: {
        color: s.color ?? Object.values(chartColors)[idx % Object.values(chartColors).length],
      },
    })),
  }

  return <EChartsWrapper option={option} height={height} loading={loading} className={className} />
}
