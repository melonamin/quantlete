import { useRef, useEffect, useCallback } from 'react'
import ReactEChartsCore from 'echarts-for-react/lib/core'

// Suppress harmless echarts-for-react warnings in development mode:
// 1. Disposal errors during React 18 Strict Mode's double-invocation
// 2. Deprecated containLabel warning (types don't support the replacement yet)
if (import.meta.env.DEV) {
  const origError = console.error
  console.error = (...args: unknown[]) => {
    const msg = args[0]
    if (typeof msg === 'object' && msg instanceof TypeError) {
      if (msg.message?.includes('disconnect')) {
        return
      }
    }
    origError.apply(console, args)
  }

  const origWarn = console.warn
  console.warn = (...args: unknown[]) => {
    const msg = args[0]
    if (typeof msg === 'string' && msg.includes('containLabel')) {
      return
    }
    origWarn.apply(console, args)
  }
}
import * as echarts from 'echarts/core'
import {
  LineChart,
  BarChart,
  PieChart,
  ScatterChart,
  EffectScatterChart,
  HeatmapChart,
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
  VisualMapComponent,
  CalendarComponent,
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
  EffectScatterChart,
  HeatmapChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  DataZoomComponent,
  ToolboxComponent,
  MarkLineComponent,
  MarkPointComponent,
  VisualMapComponent,
  CalendarComponent,
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

  const handleResize = useCallback(() => {
    chartRef.current?.getEchartsInstance()?.resize()
  }, [])

  useEffect(() => {
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [handleResize])

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
