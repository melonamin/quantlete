import { useRef, useEffect, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from './echarts-wrapper'
import { defaultTooltipConfig, chartColors } from './chart-constants'

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
  className,
}: {
  data: DonutSlice[]
  height?: number | string
  loading?: boolean
  showLegend?: boolean
  className?: string
}) {
  const containerRef = useRef<HTMLDivElement>(null)
  const [dimensions, setDimensions] = useState({ width: 400, height: 260 })

  // Track container dimensions for responsive layout
  useEffect(() => {
    const container = containerRef.current
    if (!container) return

    const observer = new ResizeObserver((entries) => {
      const entry = entries[0]
      if (entry) {
        setDimensions({
          width: entry.contentRect.width,
          height: entry.contentRect.height,
        })
      }
    })

    observer.observe(container)
    return () => observer.disconnect()
  }, [])

  // Responsive layout logic
  const isCompact = dimensions.height < 200 || dimensions.width < 280
  const isNarrow = dimensions.width < 350
  const effectiveShowLegend = showLegend && !isCompact

  // Adapt chart configuration based on size
  const chartConfig = isCompact
    ? {
        radius: ['35%', '60%'] as [string, string],
        center: ['50%', '50%'] as [string, string],
        legendPosition: {} as Record<string, unknown>,
      }
    : isNarrow
      ? {
          radius: ['40%', '65%'] as [string, string],
          center: ['50%', '40%'] as [string, string],
          legendPosition: {
            bottom: 0,
            left: 'center',
            orient: 'horizontal' as const,
          },
        }
      : {
          radius: ['40%', '70%'] as [string, string],
          center: effectiveShowLegend ? ['35%', '50%'] : ['50%', '50%'],
          legendPosition: {
            right: '5%',
            top: 'center',
            orient: 'vertical' as const,
          },
        }

  const option: EChartsOption = {
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'item',
      formatter: (params: unknown) => {
        const p = params as { name: string; value: number; percent: number }
        return `${p.name}<br/>${p.value} (${p.percent.toFixed(1)}%)`
      },
    },
    legend: effectiveShowLegend
      ? {
          ...chartConfig.legendPosition,
          textStyle: {
            color: '#888',
            fontSize: isNarrow ? 10 : 12,
          },
          itemWidth: isNarrow ? 10 : 14,
          itemHeight: isNarrow ? 10 : 14,
          itemGap: isNarrow ? 6 : 10,
        }
      : undefined,
    series: [
      {
        type: 'pie',
        radius: chartConfig.radius,
        center: chartConfig.center,
        avoidLabelOverlap: true,
        itemStyle: {
          borderRadius: 4,
          borderColor: 'transparent',
          borderWidth: 2,
        },
        label: { show: false },
        emphasis: {
          label: {
            show: !isCompact,
            fontSize: isCompact ? 10 : 14,
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

  return (
    <div ref={containerRef} className={className} style={{ height }}>
      <EChartsWrapper option={option} height="100%" loading={loading} />
    </div>
  )
}
