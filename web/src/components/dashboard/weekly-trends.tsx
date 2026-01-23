import { useState, useMemo } from 'react'
import type { EChartsOption } from 'echarts'
import { EChartsWrapper } from '@/components/charts/echarts-wrapper'
import {
  chartColors,
  defaultGridConfig,
  defaultTooltipConfig,
} from '@/components/charts/chart-constants'
import { useWeeklyTrends } from '@/lib/data/hooks'
import { useSportTypeStats } from '@/lib/data/hooks'
import { formatSportType } from '@/lib/sport-types'
import { useFormattedMetrics } from '@/hooks/use-formatted-metrics'
import { formatDuration, metersToDisplayUnit } from '@/lib/format'
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

export function WeeklyTrends() {
  const [metric, setMetric] = useState<Metric>('distance')
  const [sportType, setSportType] = useState<string>('all')
  const { formatDistance, formatElevation, distanceUnit, elevationUnit, unitSystem } =
    useFormattedMetrics()

  const { data: sportStats } = useSportTypeStats()
  const { data, isLoading, error } = useWeeklyTrends({
    weeks: 12,
    sport_type: sportType === 'all' ? '' : sportType,
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

  const chartData = useMemo(() => {
    if (!data?.weeks) return []
    return data.weeks.map((w) => ({
      week: w.week_start.slice(5), // MM-DD format
      distance: w.total_distance,
      time: w.total_time,
      elevation: w.total_elevation,
    }))
  }, [data])

  const getValue = (d: (typeof chartData)[0]) => {
    switch (metric) {
      case 'distance':
        return metersToDisplayUnit(d.distance, unitSystem)
      case 'time':
        return d.time / 3600 // Convert to hours
      case 'elevation':
        return unitSystem === 'imperial' ? d.elevation / 0.3048 : d.elevation // Convert to feet if imperial
    }
  }

  const formatValue = (value: number) => {
    switch (metric) {
      case 'distance':
        // Value is already in display units (km or mi), convert back to meters for formatting
        return formatDistance(
          unitSystem === 'imperial' ? value * 1609.344 : value * 1000
        )
      case 'time':
        return formatDuration(value * 3600)
      case 'elevation':
        // Value is already in display units (m or ft), convert back to meters for formatting
        return formatElevation(unitSystem === 'imperial' ? value * 0.3048 : value)
    }
  }

  const getYAxisLabel = () => {
    switch (metric) {
      case 'distance':
        return distanceUnit
      case 'time':
        return 'hours'
      case 'elevation':
        return elevationUnit
    }
  }

  const option: EChartsOption = {
    grid: {
      ...defaultGridConfig,
      top: 20,
      bottom: 30,
      left: 45,
      right: 10,
    },
    tooltip: {
      ...defaultTooltipConfig,
      trigger: 'axis',
      formatter: (params: unknown) => {
        const p = params as Array<{ name: string; value: number }>
        if (!p || p.length === 0) return ''
        const point = p[0]
        return `Week of ${point.name}<br/>${formatValue(point.value)}`
      },
    },
    xAxis: {
      type: 'category',
      data: chartData.map((d) => d.week),
      axisLine: { lineStyle: { color: 'rgba(100, 100, 120, 0.3)' } },
      axisLabel: {
        color: 'rgba(160, 160, 180, 0.8)',
        fontSize: 10,
        rotate: 45,
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
    series: [
      {
        type: 'line',
        data: chartData.map(getValue),
        smooth: true,
        symbol: 'circle',
        symbolSize: 6,
        lineStyle: {
          color: chartColors.primary,
          width: 2,
        },
        itemStyle: {
          color: chartColors.primary,
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(74, 222, 128, 0.3)' },
              { offset: 1, color: 'rgba(74, 222, 128, 0.05)' },
            ],
          },
        },
      },
    ],
  }

  if (error) {
    return (
      <WidgetWrapper title="Weekly Trends">
        <div className="flex items-center justify-center h-full text-sm text-muted-foreground">
          Failed to load weekly trends
        </div>
      </WidgetWrapper>
    )
  }

  return (
    <WidgetWrapper
      title="Weekly Trends"
      action={
        <Select value={sportType} onValueChange={setSportType}>
          <SelectTrigger className="h-6 w-[100px] text-xs">
            <SelectValue placeholder="All Sports" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Sports</SelectItem>
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
          ) : chartData.length === 0 ? (
            <div className="h-full flex items-center justify-center text-sm text-muted-foreground">
              No activity data for selected period
            </div>
          ) : (
            <EChartsWrapper option={option} height="100%" />
          )}
        </div>
      </div>
    </WidgetWrapper>
  )
}
