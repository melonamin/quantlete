import { useState, useMemo } from 'react'
import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from '@/components/charts/echarts-wrapper'
import { defaultGridConfig, defaultTooltipConfig, yearColors } from '@/components/charts/chart-constants'
import { useMonthlyComparison } from '@/lib/data/hooks'
import { useSportTypeStats } from '@/lib/data/hooks'
import { formatSportType } from '@/lib/sport-types'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { formatDuration } from '@/lib/format'
import { WidgetWrapper } from './widget-wrapper'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'

type Metric = 'distance' | 'time' | 'elevation'

const MONTH_LABELS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

export function MonthlyComparison() {
  const [metric, setMetric] = useState<Metric>('distance')
  const [sportType, setSportType] = useState<string>('')
  const [hiddenYears, setHiddenYears] = useState<Set<number>>(new Set())
  const { formatDistance, formatElevation } = useFormattedMetrics()

  const { data: sportStats } = useSportTypeStats()
  const { data, isLoading, error } = useMonthlyComparison({
    sport_type: sportType,
  })

  const sportOptions = useMemo(() => {
    if (!sportStats) return []
    return sportStats
      .filter((s) => s.activity_count > 0)
      .map((s) => ({
        value: s.sport_type,
        label: formatSportType(s.sport_type),
      }))
  }, [sportStats])

  // Group data by year, creating a map of year -> month values (1-12)
  const yearData = useMemo(() => {
    if (!data?.months) return new Map<number, Map<number, number>>()

    const result = new Map<number, Map<number, number>>()
    for (const point of data.months) {
      if (!result.has(point.year)) {
        result.set(point.year, new Map())
      }
      const monthMap = result.get(point.year)!

      let value: number
      switch (metric) {
        case 'distance':
          value = point.total_distance / 1000 // km
          break
        case 'time':
          value = point.total_time / 3600 // hours
          break
        case 'elevation':
          value = point.total_elevation
          break
      }
      monthMap.set(point.month, value)
    }
    return result
  }, [data, metric])

  // Get years sorted in descending order (latest year first)
  const years = useMemo(() => {
    return data?.years ? [...data.years].sort((a, b) => b - a) : []
  }, [data])

  const formatValue = (value: number) => {
    switch (metric) {
      case 'distance':
        return formatDistance(value * 1000)
      case 'time':
        return formatDuration(value * 3600)
      case 'elevation':
        return formatElevation(value)
    }
  }

  const getYAxisLabel = () => {
    switch (metric) {
      case 'distance':
        return 'km'
      case 'time':
        return 'hours'
      case 'elevation':
        return 'm'
    }
  }

  // Toggle year visibility via legend click
  const handleLegendClick = (year: number) => {
    setHiddenYears((prev) => {
      const next = new Set(prev)
      if (next.has(year)) {
        next.delete(year)
      } else {
        next.add(year)
      }
      return next
    })
  }

  // Build series data for each year
  const series = useMemo(() => {
    return years.map((year, index) => {
      const monthMap = yearData.get(year) || new Map()
      // Create array of 12 values (one per month), null if no data
      const values = MONTH_LABELS.map((_, monthIndex) => {
        const value = monthMap.get(monthIndex + 1)
        return value !== undefined ? value : null
      })

      const color = yearColors[index % yearColors.length]
      const isHidden = hiddenYears.has(year)

      return {
        name: String(year),
        type: 'line' as const,
        data: isHidden ? [] : values,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        connectNulls: true,
        lineStyle: {
          color,
          width: index === 0 ? 2.5 : 1.5, // Latest year is thicker
          opacity: isHidden ? 0 : 1,
        },
        itemStyle: {
          color,
          opacity: isHidden ? 0 : 1,
        },
        emphasis: {
          focus: 'series' as const,
          lineStyle: {
            width: 3,
          },
        },
      }
    })
  }, [years, yearData, hiddenYears])

  const option: EChartsOption = {
    grid: {
      ...defaultGridConfig,
      top: 30,
      bottom: 30,
      left: 45,
      right: 10,
    },
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'axis',
      formatter: (params: unknown) => {
        const items = params as Array<{ seriesName: string; value: number | null; color: string }>
        if (!items || items.length === 0) return ''

        const month = MONTH_LABELS[(items[0] as { dataIndex: number }).dataIndex]
        const lines = items
          .filter((item) => item.value !== null && item.value !== undefined)
          .map((item) => {
            const colorDot = `<span style="display:inline-block;width:10px;height:10px;border-radius:50%;background:${item.color};margin-right:5px;"></span>`
            return `${colorDot}${item.seriesName}: ${formatValue(item.value as number)}`
          })

        return `<div style="font-weight:600;margin-bottom:4px;">${month}</div>${lines.join('<br/>')}`
      },
    },
    legend: {
      show: true,
      top: 0,
      right: 0,
      itemWidth: 16,
      itemHeight: 8,
      textStyle: {
        color: 'rgba(160, 160, 180, 0.8)',
        fontSize: 10,
      },
      // Custom selected state based on hiddenYears
      selected: Object.fromEntries(years.map((y) => [String(y), !hiddenYears.has(y)])),
    },
    xAxis: {
      type: 'category',
      data: MONTH_LABELS,
      axisLine: { lineStyle: { color: 'rgba(100, 100, 120, 0.3)' } },
      axisLabel: {
        color: 'rgba(160, 160, 180, 0.8)',
        fontSize: 10,
      },
      axisTick: { show: false },
    },
    yAxis: {
      type: 'value',
      name: getYAxisLabel(),
      nameLocation: 'middle',
      nameGap: 30,
      nameTextStyle: {
        color: 'rgba(160, 160, 180, 0.8)',
        fontSize: 10,
      },
      axisLine: { show: false },
      axisLabel: {
        color: 'rgba(160, 160, 180, 0.8)',
        fontSize: 10,
      },
      splitLine: { lineStyle: { color: 'rgba(100, 100, 120, 0.15)' } },
    },
    series,
  }

  // Handle ECharts legend select event
  const handleChartEvents = {
    legendselectchanged: (params: { name: string; selected: Record<string, boolean> }) => {
      const year = parseInt(params.name, 10)
      if (!isNaN(year)) {
        handleLegendClick(year)
      }
    },
  }

  if (error) {
    return (
      <WidgetWrapper title="Monthly Comparison">
        <div className="flex items-center justify-center h-full text-sm text-muted-foreground">
          Failed to load monthly comparison data
        </div>
      </WidgetWrapper>
    )
  }

  return (
    <WidgetWrapper
      title="Monthly Comparison"
      action={
        <Select value={sportType} onValueChange={setSportType}>
          <SelectTrigger className="h-6 w-[100px] text-xs">
            <SelectValue placeholder="All Sports" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All Sports</SelectItem>
            {sportOptions.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      }
    >
      <div className="h-full flex flex-col">
        <div className="mb-3 flex gap-1.5 flex-shrink-0">
          <Button
            variant={metric === 'distance' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-7 px-2.5 text-xs"
            onClick={() => setMetric('distance')}
          >
            Distance
          </Button>
          <Button
            variant={metric === 'time' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-7 px-2.5 text-xs"
            onClick={() => setMetric('time')}
          >
            Time
          </Button>
          <Button
            variant={metric === 'elevation' ? 'secondary' : 'ghost'}
            size="sm"
            className="h-7 px-2.5 text-xs"
            onClick={() => setMetric('elevation')}
          >
            Elevation
          </Button>
        </div>
        <div className="flex-1 min-h-0">
          {isLoading ? (
            <div className="h-full flex items-center justify-center">
              <Skeleton className="w-full h-[200px]" />
            </div>
          ) : years.length === 0 ? (
            <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
              No activity data available
            </div>
          ) : (
            <EChartsWrapper option={option} height="100%" onEvents={handleChartEvents} />
          )}
        </div>
      </div>
    </WidgetWrapper>
  )
}
