import type { EChartsOption } from 'echarts'
import { EChartsWrapper, chartColors, defaultGridConfig, defaultTooltipConfig } from './echarts-wrapper'
import { formatDistance } from '@/lib/format'

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
      case 'distance': return d.distance / 1000 // Convert to km
      case 'count': return d.count
      case 'time': return d.time / 3600 // Convert to hours
    }
  }

  const getLabel = () => {
    switch (metric) {
      case 'distance': return 'Distance (km)'
      case 'count': return 'Activities'
      case 'time': return 'Time (hours)'
    }
  }

  const formatValue = (value: number) => {
    switch (metric) {
      case 'distance': return `${value.toFixed(1)} km`
      case 'count': return `${value} activities`
      case 'time': return `${value.toFixed(1)} hours`
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
      data: data.map(d => d.month),
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
    series: [{
      type: 'bar',
      data: data.map(d => getValue(d)),
      itemStyle: {
        color: chartColors.primary,
        borderRadius: [4, 4, 0, 0],
      },
      emphasis: {
        itemStyle: {
          color: chartColors.secondary,
        },
      },
    }],
  }

  return (
    <EChartsWrapper
      option={option}
      height={height}
      loading={loading}
      className={className}
    />
  )
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
  const getValue = (d: SportData) => metric === 'count' ? d.count : d.distance

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

  const option: EChartsOption = {
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'item',
      formatter: (params: unknown) => {
        const p = params as { name: string; value: number; percent: number }
        const valueStr = metric === 'count'
          ? `${p.value} activities`
          : formatDistance(p.value)
        return `${p.name}<br/>${valueStr} (${p.percent.toFixed(1)}%)`
      },
    },
    legend: {
      orient: 'vertical',
      right: '5%',
      top: 'center',
      textStyle: { color: '#888' },
    },
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['35%', '50%'],
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
          show: true,
          fontSize: 14,
          fontWeight: 'bold',
        },
      },
      data: data.map((d, idx) => ({
        name: d.sport_type.replace(/([A-Z])/g, ' $1').trim(),
        value: getValue(d),
        itemStyle: {
          color: sportColorMap[d.sport_type] ?? Object.values(chartColors)[idx % Object.values(chartColors).length],
        },
      })),
    }],
  }

  return (
    <EChartsWrapper
      option={option}
      height={height}
      loading={loading}
      className={className}
    />
  )
}

// Activity calendar heatmap
interface CalendarData {
  date: string
  count: number
  distance?: number
}

interface ActivityCalendarChartProps {
  data: CalendarData[]
  year: number
  height?: number | string
  loading?: boolean
  className?: string
}

export function ActivityCalendarChart({
  data,
  year,
  height = 180,
  loading = false,
  className,
}: ActivityCalendarChartProps) {
  const maxCount = Math.max(...data.map(d => d.count), 1)

  const option: EChartsOption = {
    tooltip: {
      ...defaultTooltipConfig,
      formatter: (params: unknown) => {
        const p = params as { data: [string, number] }
        const date = p.data[0]
        const count = p.data[1]
        return `${date}<br/>${count} ${count === 1 ? 'activity' : 'activities'}`
      },
    },
    visualMap: {
      show: false,
      min: 0,
      max: maxCount,
      inRange: {
        color: ['#ebedf0', '#9be9a8', '#40c463', '#30a14e', '#216e39'],
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
    series: [{
      type: 'heatmap',
      coordinateSystem: 'calendar',
      data: data.map(d => [d.date, d.count]),
    }],
  }

  return (
    <EChartsWrapper
      option={option}
      height={height}
      loading={loading}
      className={className}
    />
  )
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
      data: data.map(d => d.day),
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888' },
    },
    yAxis: {
      type: 'value',
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#888' },
      splitLine: { lineStyle: { color: 'rgba(128, 128, 128, 0.2)' } },
    },
    series: [{
      type: 'bar',
      data: data.map(d => d.count),
      itemStyle: {
        color: chartColors.primary,
        borderRadius: [4, 4, 0, 0],
      },
    }],
  }

  return (
    <EChartsWrapper
      option={option}
      height={height}
      loading={loading}
      className={className}
    />
  )
}
