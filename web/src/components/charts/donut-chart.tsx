import type { EChartsOption } from 'echarts'
import { EChartsWrapper, defaultTooltipConfig, chartColors } from './echarts-wrapper'

export interface DonutSlice {
  name: string
  value: number
  color?: string
}

export function DonutChart({
  data,
  height = 260,
  loading = false,
  showLegend = true,
}: {
  data: DonutSlice[]
  height?: number | string
  loading?: boolean
  showLegend?: boolean
}) {
  const option: EChartsOption = {
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'item',
      formatter: (params: unknown) => {
        const p = params as { name: string; value: number; percent: number }
        return `${p.name}<br/>${p.value} (${p.percent.toFixed(1)}%)`
      },
    },
    legend: showLegend
      ? {
          orient: 'vertical',
          right: '5%',
          top: 'center',
          textStyle: { color: '#888' },
        }
      : undefined,
    series: [
      {
        type: 'pie',
        radius: ['40%', '70%'],
        center: showLegend ? ['35%', '50%'] : ['50%', '50%'],
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 4,
          borderColor: 'transparent',
          borderWidth: 2,
        },
        label: { show: false },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold',
          },
        },
        data: data.map((d, idx) => ({
          name: d.name,
          value: d.value,
          itemStyle: {
            color: d.color ?? Object.values(chartColors)[idx % Object.values(chartColors).length],
          },
        })),
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} />
}
