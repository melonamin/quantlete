import { useRef, useEffect } from 'react'
import ReactEChartsCore from 'echarts-for-react/lib/core'
import * as echarts from 'echarts/core'
import {
  LineChart,
  BarChart,
  PieChart,
  ScatterChart,
} from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  DataZoomComponent,
  ToolboxComponent,
  MarkLineComponent,
  MarkPointComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'
import { Skeleton } from '@/components/ui/skeleton'

// Register required components
echarts.use([
  LineChart,
  BarChart,
  PieChart,
  ScatterChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  DataZoomComponent,
  ToolboxComponent,
  MarkLineComponent,
  MarkPointComponent,
  CanvasRenderer,
])

interface EChartsWrapperProps {
  option: EChartsOption
  height?: number | string
  loading?: boolean
  className?: string
}

export function EChartsWrapper({
  option,
  height = 300,
  loading = false,
  className,
}: EChartsWrapperProps) {
  const chartRef = useRef<ReactEChartsCore>(null)

  useEffect(() => {
    const handleResize = () => {
      chartRef.current?.getEchartsInstance()?.resize()
    }

    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [])

  if (loading) {
    return (
      <Skeleton
        className={className}
        style={{ height: typeof height === 'number' ? `${height}px` : height }}
      />
    )
  }

  return (
    <ReactEChartsCore
      ref={chartRef}
      echarts={echarts}
      option={option}
      style={{ height: typeof height === 'number' ? `${height}px` : height }}
      className={className}
      notMerge={true}
      lazyUpdate={true}
    />
  )
}

// Theme colors that work in both light and dark modes
export const chartColors = {
  primary: '#3b82f6',
  secondary: '#8b5cf6',
  success: '#22c55e',
  warning: '#f59e0b',
  danger: '#ef4444',
  info: '#06b6d4',
  // Sport colors
  ride: '#f97316',
  run: '#22c55e',
  swim: '#3b82f6',
  walk: '#a855f7',
  winter: '#06b6d4',
  other: '#6b7280',
  // Data series
  heartRate: '#ef4444',
  power: '#f59e0b',
  cadence: '#8b5cf6',
  elevation: '#22c55e',
  speed: '#3b82f6',
  temperature: '#06b6d4',
}

// Common chart configurations
export const defaultGridConfig = {
  left: '3%',
  right: '4%',
  bottom: '3%',
  containLabel: true,
}

export const defaultTooltipConfig = {
  trigger: 'axis' as const,
  backgroundColor: 'rgba(0, 0, 0, 0.8)',
  borderColor: 'transparent',
  textStyle: {
    color: '#fff',
  },
}
