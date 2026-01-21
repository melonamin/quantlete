import { useRef, useEffect, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from './echarts-wrapper'
import {
  defaultTooltipConfig,
  chartColors,
  maxInlineLegendItems,
  getColorFromMap,
} from './chart-constants'

export interface DonutSlice {
  name: string
  value: number
  color?: string
}

export interface DonutChartProps {
  data: DonutSlice[]
  height?: number | string
  loading?: boolean
  showLegend?: boolean
  className?: string
  /** Optional color map for semantic coloring based on slice name */
  colorMap?: Record<string, string>
}

export function DonutChart({
  data,
  height = 260,
  loading = false,
  showLegend = true,
  className,
  colorMap,
}: DonutChartProps) {
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
  const hasManyCategories = data.length > maxInlineLegendItems

  // Adapt chart configuration based on size and number of categories
  const chartConfig = isCompact
    ? {
        radius: ['35%', '60%'] as [string, string],
        center: ['50%', '50%'] as [string, string],
        legendPosition: {} as Record<string, unknown>,
        legendType: 'plain' as const,
      }
    : isNarrow || hasManyCategories
      ? {
          // Use bottom legend for narrow views or when there are many categories
          radius: ['35%', '60%'] as [string, string],
          center: ['50%', hasManyCategories ? '40%' : '40%'] as [string, string],
          legendPosition: {
            bottom: 0,
            left: 'center',
            orient: 'horizontal' as const,
          },
          // Use scrollable legend when there are many categories
          legendType: hasManyCategories ? ('scroll' as const) : ('plain' as const),
        }
      : {
          radius: ['40%', '70%'] as [string, string],
          center: effectiveShowLegend ? ['35%', '50%'] : ['50%', '50%'],
          legendPosition: {
            right: '5%',
            top: 'center',
            orient: 'vertical' as const,
          },
          legendType: 'plain' as const,
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
          type: chartConfig.legendType,
          textStyle: {
            color: '#888',
            fontSize: isNarrow || hasManyCategories ? 10 : 12,
          },
          itemWidth: isNarrow || hasManyCategories ? 10 : 14,
          itemHeight: isNarrow || hasManyCategories ? 10 : 14,
          itemGap: isNarrow || hasManyCategories ? 6 : 10,
          // Scroll legend controls
          pageButtonItemGap: 5,
          pageButtonGap: 5,
          pageIconColor: '#888',
          pageIconInactiveColor: '#444',
          pageTextStyle: {
            color: '#888',
            fontSize: 10,
          },
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
            color:
              d.color ??
              (colorMap
                ? getColorFromMap(d.name, colorMap, Object.values(chartColors), idx)
                : Object.values(chartColors)[idx % Object.values(chartColors).length]),
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
