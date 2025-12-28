import { useRef, useEffect } from 'react'
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

// Terminal-themed chart colors
export const chartColors = {
  primary: '#4ade80', // terminal green
  secondary: '#fb923c', // strava orange
  success: '#4ade80',
  warning: '#fbbf24', // terminal amber
  danger: '#f87171',
  info: '#22d3ee', // terminal cyan
  // Sport colors
  ride: '#4ade80',
  run: '#fb923c',
  swim: '#22d3ee',
  walk: '#a78bfa',
  winter: '#38bdf8',
  other: '#6b7280',
  // Data series
  heartRate: '#f87171',
  power: '#fbbf24',
  cadence: '#a78bfa',
  elevation: '#4ade80',
  speed: '#22d3ee',
  temperature: '#fb923c',
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
  backgroundColor: 'rgba(20, 20, 30, 0.95)',
  borderColor: 'rgba(60, 60, 80, 0.5)',
  textStyle: {
    color: '#e5e5e5',
    fontFamily: "'JetBrains Mono', monospace",
    fontSize: 11,
  },
}

export const defaultAxisStyle = {
  axisLine: {
    lineStyle: { color: 'rgba(100, 100, 120, 0.3)' },
  },
  splitLine: {
    lineStyle: { color: 'rgba(100, 100, 120, 0.15)' },
  },
  axisLabel: {
    color: 'rgba(160, 160, 180, 0.8)',
    fontFamily: "'JetBrains Mono', monospace",
    fontSize: 10,
  },
}
