import { useRef, useEffect, useState } from 'react'
import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from './echarts-wrapper'
import {
  chartColors,
  calendarPalettes,
  defaultGridConfig,
  defaultTooltipConfig,
  maxInlineLegendItems,
} from './chart-constants'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'

// Monthly activity summary chart
interface MonthlyData {
  month: string
  distance: number
  count: number
  time: number
}

interface MonthlyStatsChartProps {
  data: MonthlyData[]
  metric?: 'distance' | 'count' | 'time'
  height?: number | string
  loading?: boolean
  className?: string
}

export function MonthlyStatsChart({
  data,
  metric = 'distance',
  height = 250,
  loading = false,
  className,
}: MonthlyStatsChartProps) {
  const getValue = (d: MonthlyData) => {
    switch (metric) {
      case 'distance':
        return d.distance / 1000 // Convert to km
      case 'count':
        return d.count
      case 'time':
        return d.time / 3600 // Convert to hours
    }
  }

  const getLabel = () => {
    switch (metric) {
      case 'distance':
        return 'Distance (km)'
      case 'count':
        return 'Activities'
      case 'time':
        return 'Time (hours)'
    }
  }

  const formatValue = (value: number) => {
    const v = value ?? 0
    switch (metric) {
      case 'distance':
        return `${v.toFixed(1)} km`
      case 'count':
        return `${v} activities`
      case 'time':
        return `${v.toFixed(1)} hours`
    }
  }

  const option: EChartsOption = {
    grid: {
      ...defaultGridConfig,
      bottom: '15%',
    },
    tooltip: {
      ...defaultTooltipConfig,
      formatter: (params: unknown) => {
        const p = params as { name: string; value: number }
        return `${p.name}<br/>${formatValue(p.value)}`
      },
    },
    xAxis: {
      type: 'category',
      data: data.map((d) => d.month),
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888', rotate: 45 },
    },
    yAxis: {
      type: 'value',
      name: getLabel(),
      nameLocation: 'middle',
      nameGap: 40,
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888' },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.2)' } },
    },
    series: [
      {
        type: 'bar',
        data: data.map((d) => getValue(d)),
        itemStyle: {
          color: chartColors.primary,
          borderRadius: [4, 4, 0, 0],
        },
        emphasis: {
          itemStyle: {
            color: chartColors.secondary,
          },
        },
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} className={className} />
}

// Sport type distribution pie chart
interface SportData {
  sport_type: string
  count: number
  distance: number
}

interface SportDistributionChartProps {
  data: SportData[]
  metric?: 'count' | 'distance'
  height?: number | string
  loading?: boolean
  className?: string
}

export function SportDistributionChart({
  data,
  metric = 'count',
  height = 250,
  loading = false,
  className,
}: SportDistributionChartProps) {
  const { formatDistance } = useFormattedMetrics()
  const containerRef = useRef<HTMLDivElement>(null)
  const [dimensions, setDimensions] = useState({ width: 400, height: 250 })

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

  const getValue = (d: SportData) => (metric === 'count' ? d.count : d.distance)

  const sportColorMap: Record<string, string> = {
    Ride: chartColors.ride,
    VirtualRide: chartColors.ride,
    MountainBikeRide: chartColors.ride,
    GravelRide: chartColors.ride,
    Run: chartColors.run,
    VirtualRun: chartColors.run,
    TrailRun: chartColors.run,
    Walk: chartColors.walk,
    Hike: chartColors.walk,
    Swim: chartColors.swim,
    AlpineSki: chartColors.winter,
    NordicSki: chartColors.winter,
    Snowboard: chartColors.winter,
  }

  // Responsive layout logic based on container dimensions
  const isCompact = dimensions.height < 200 || dimensions.width < 280
  const isNarrow = dimensions.width < 350
  const hasManyCategories = data.length > maxInlineLegendItems

  // Adapt chart configuration based on size and number of categories
  const chartConfig = isCompact
    ? {
        // Compact mode: no legend, centered chart, smaller radius
        radius: ['35%', '60%'] as [string, string],
        center: ['50%', '50%'] as [string, string],
        showLegend: false,
        legendPosition: {} as Record<string, unknown>,
        legendType: 'plain' as const,
      }
    : isNarrow || hasManyCategories
      ? {
          // Use bottom legend for narrow views or when there are many categories
          radius: ['35%', '60%'] as [string, string],
          center: ['50%', '40%'] as [string, string],
          showLegend: true,
          legendPosition: {
            bottom: 0,
            left: 'center',
            orient: 'horizontal' as const,
          },
          // Use scrollable legend when there are many categories
          legendType: hasManyCategories ? ('scroll' as const) : ('plain' as const),
        }
      : {
          // Full mode: vertical legend on right side
          radius: ['40%', '70%'] as [string, string],
          center: ['40%', '50%'] as [string, string],
          showLegend: true,
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
        const valueStr = metric === 'count' ? `${p.value} activities` : formatDistance(p.value)
        return `${p.name}<br/>${valueStr} (${p.percent.toFixed(1)}%)`
      },
    },
    legend: chartConfig.showLegend
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
        label: {
          show: false,
        },
        emphasis: {
          label: {
            show: !isCompact,
            fontSize: isCompact ? 10 : 14,
            fontWeight: 'bold',
          },
        },
        data: data.map((d, idx) => ({
          name: d.sport_type.replace(/([A-Z])/g, ' $1').trim(),
          value: getValue(d),
          itemStyle: {
            color:
              sportColorMap[d.sport_type] ??
              Object.values(chartColors)[idx % Object.values(chartColors).length],
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

// Activity calendar heatmap
export type CalendarMetric = 'count' | 'distance' | 'time' | 'calories'

interface CalendarData {
  date: string
  count: number
  distance?: number
  time?: number // seconds
  calories?: number
}

interface ActivityCalendarChartProps {
  data: CalendarData[]
  year: number
  metric?: CalendarMetric
  height?: number | string
  loading?: boolean
  className?: string
}

export function ActivityCalendarChart({
  data,
  year,
  metric = 'count',
  height = 180,
  loading = false,
  className,
}: ActivityCalendarChartProps) {
  // Get value based on selected metric
  const getValue = (d: CalendarData): number => {
    switch (metric) {
      case 'count':
        return d.count
      case 'distance':
        return (d.distance ?? 0) / 1000 // Convert to km
      case 'time':
        return (d.time ?? 0) / 60 // Convert to minutes
      case 'calories':
        return d.calories ?? 0
    }
  }

  // Format tooltip based on metric
  const formatTooltip = (date: string, value: number): string => {
    switch (metric) {
      case 'count':
        return `${date}<br/>${Math.round(value)} ${value === 1 ? 'activity' : 'activities'}`
      case 'distance':
        return `${date}<br/>${value.toFixed(1)} km`
      case 'time': {
        const hours = Math.floor(value / 60)
        const mins = Math.round(value % 60)
        return `${date}<br/>${hours > 0 ? `${hours}h ` : ''}${mins}m`
      }
      case 'calories':
        return `${date}<br/>${Math.round(value)} kcal`
    }
  }

  const values = data.map((d) => getValue(d))
  const maxValue = Math.max(...values, 1)
  const palette = calendarPalettes[metric]

  const option: EChartsOption = {
    tooltip: {
      ...defaultTooltipConfig,
      formatter: (params: unknown) => {
        const p = params as { data: [string, number] }
        return formatTooltip(p.data[0], p.data[1])
      },
    },
    visualMap: {
      show: false,
      min: 0,
      max: maxValue,
      inRange: {
        color: palette,
      },
    },
    calendar: {
      top: 30,
      left: 30,
      right: 30,
      cellSize: ['auto', 13],
      range: year.toString(),
      itemStyle: {
        borderWidth: 2,
        borderColor: 'transparent',
      },
      yearLabel: { show: false },
      dayLabel: {
        firstDay: 1,
        nameMap: ['S', 'M', 'T', 'W', 'T', 'F', 'S'],
        color: '#888',
      },
      monthLabel: {
        color: '#888',
      },
      splitLine: {
        show: false,
      },
    },
    series: [
      {
        type: 'heatmap',
        coordinateSystem: 'calendar',
        data: data.map((d) => [d.date, getValue(d)]),
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} className={className} />
}

// Weekly activity distribution (day of week)
interface WeekdayData {
  day: string
  count: number
}

interface WeekdayDistributionChartProps {
  data: WeekdayData[]
  height?: number | string
  loading?: boolean
  className?: string
}

export function WeekdayDistributionChart({
  data,
  height = 200,
  loading = false,
  className,
}: WeekdayDistributionChartProps) {
  const option: EChartsOption = {
    grid: defaultGridConfig,
    tooltip: {
      ...defaultTooltipConfig,
      formatter: (params: unknown) => {
        const p = params as { name: string; value: number }
        return `${p.name}<br/>${p.value} activities`
      },
    },
    xAxis: {
      type: 'category',
      data: data.map((d) => d.day),
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888' },
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888' },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.2)' } },
    },
    series: [
      {
        type: 'bar',
        data: data.map((d) => d.count),
        itemStyle: {
          color: chartColors.primary,
          borderRadius: [4, 4, 0, 0],
        },
      },
    ],
  }

  return <EChartsWrapper option={option} height={height} loading={loading} className={className} />
}
