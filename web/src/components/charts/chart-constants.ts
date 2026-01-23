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

// Calendar metric palettes (5-step gradients for heatmap)
export const calendarPalettes = {
  count: [
    'rgba(74, 222, 128, 0.1)',
    'rgba(74, 222, 128, 0.3)',
    'rgba(74, 222, 128, 0.55)',
    'rgba(74, 222, 128, 0.8)',
    '#4ade80',
  ],
  distance: [
    'rgba(34, 211, 238, 0.1)',
    'rgba(34, 211, 238, 0.3)',
    'rgba(34, 211, 238, 0.55)',
    'rgba(34, 211, 238, 0.8)',
    '#22d3ee',
  ],
  time: [
    'rgba(251, 191, 36, 0.1)',
    'rgba(251, 191, 36, 0.3)',
    'rgba(251, 191, 36, 0.55)',
    'rgba(251, 191, 36, 0.8)',
    '#fbbf24',
  ],
  calories: [
    'rgba(251, 146, 60, 0.1)',
    'rgba(251, 146, 60, 0.3)',
    'rgba(251, 146, 60, 0.55)',
    'rgba(251, 146, 60, 0.8)',
    '#fb923c',
  ],
  // Multi-color intensity gradient (gray -> green -> amber -> orange -> red)
  // Represents workout intensity from rest/easy to very hard
  intensity: [
    'rgba(107, 114, 128, 0.3)', // gray - rest/very easy
    '#4ade80', // green - easy/moderate
    '#fbbf24', // amber - moderate/tempo
    '#fb923c', // orange - hard/threshold
    '#f87171', // red - very hard/VO2max
  ],
}

// Zone colors for training zones (Recovery → VO2max)
export const zoneColors = {
  z1: 'rgba(74, 222, 128, 0.4)', // Light terminal green - Recovery
  z2: '#4ade80', // Terminal green - Endurance
  z3: '#fbbf24', // Terminal amber - Tempo
  z4: '#fb923c', // Terminal orange - Threshold
  z5: '#f87171', // Terminal red - VO2max
}
export const zoneColorsArray = [
  zoneColors.z1,
  zoneColors.z2,
  zoneColors.z3,
  zoneColors.z4,
  zoneColors.z5,
]

// Zone labels for training zones
export const zoneLabels = ['Z1 Recovery', 'Z2 Endurance', 'Z3 Tempo', 'Z4 Threshold', 'Z5 VO2max']

// UI color tokens (semantic)
export const uiColors = {
  success: '#4ade80', // terminal green
  danger: '#f87171', // red
  warning: '#fbbf24', // amber
  info: '#22d3ee', // cyan
  accent: '#fb923c', // strava orange
  muted: '#6b7280', // gray
}

// Icon-specific colors for insights/stats
export const iconColors = {
  fire: '#fb923c', // orange - streak/intensity
  trend: '#4ade80', // green - positive trend
  trophy: '#fbbf24', // amber - achievements
  clock: '#22d3ee', // cyan - time-related
  location: '#14b8a6', // teal - maps/location
  power: '#fbbf24', // amber - power/energy
  calendar: '#3b82f6', // blue - schedule
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

// Legend configuration thresholds
// When a chart has more than this many categories, switch from inline to scrollable legend
export const maxInlineLegendItems = 6

// When a chart has more than this many categories, group small values into "Other"
export const maxChartCategories = 8

// Comprehensive sport colors for donut/pie charts
// Each sport type has a unique color for better differentiation
// Related sports use similar hues with different shades
export const sportColors: Record<string, string> = {
  // Cycling - Green family
  Ride: '#4ade80', // terminal green
  VirtualRide: '#22c55e', // slightly darker green
  MountainBikeRide: '#16a34a', // forest green
  GravelRide: '#15803d', // dark green
  EBikeRide: '#86efac', // light green
  Handcycle: '#bbf7d0', // pale green
  Velomobile: '#dcfce7', // very pale green

  // Running - Orange/Amber family
  Run: '#fb923c', // strava orange
  VirtualRun: '#f97316', // darker orange
  TrailRun: '#ea580c', // burnt orange
  Walk: '#fdba74', // light orange
  Hike: '#fed7aa', // pale orange

  // Water sports - Cyan/Blue family
  Swim: '#22d3ee', // terminal cyan
  OpenWaterSwim: '#06b6d4', // darker cyan
  Rowing: '#0891b2', // teal
  Kayaking: '#0e7490', // dark teal
  StandUpPaddling: '#67e8f9', // light cyan
  Surfing: '#a5f3fc', // pale cyan
  Canoeing: '#155e75', // deep teal
  Sailing: '#164e63', // navy teal

  // Winter sports - Blue family
  AlpineSki: '#38bdf8', // sky blue
  NordicSki: '#0ea5e9', // blue
  BackcountrySki: '#0284c7', // darker blue
  Snowboard: '#7dd3fc', // light blue
  Snowshoe: '#bae6fd', // pale blue
  IceSkate: '#e0f2fe', // very pale blue

  // Gym/Indoor - Purple family
  WeightTraining: '#a78bfa', // purple
  Workout: '#8b5cf6', // darker purple
  Crossfit: '#7c3aed', // violet
  Yoga: '#c4b5fd', // light purple
  Pilates: '#ddd6fe', // pale purple
  Elliptical: '#ede9fe', // very pale purple
  StairStepper: '#6d28d9', // deep purple

  // Racket sports - Pink/Rose family
  Tennis: '#f472b6', // pink
  Pickleball: '#ec4899', // darker pink
  Badminton: '#db2777', // magenta
  Squash: '#be185d', // dark pink
  TableTennis: '#fbcfe8', // light pink
  Racquetball: '#fce7f3', // pale pink

  // Team/Ball sports - Red/Coral family
  Soccer: '#f87171', // coral red
  Basketball: '#ef4444', // red
  Football: '#dc2626', // darker red
  Volleyball: '#fca5a5', // light coral
  Hockey: '#b91c1c', // deep red
  Lacrosse: '#991b1b', // very deep red
  Rugby: '#fecaca', // pale red

  // Golf - Lime family
  Golf: '#a3e635', // lime
  MiniGolf: '#84cc16', // lime green

  // Equestrian - Amber family
  Horseback: '#fbbf24', // amber
  HorsebackRiding: '#fbbf24', // amber

  // Skateboarding - Teal family
  Skateboard: '#14b8a6', // teal
  InlineSkate: '#0d9488', // darker teal
  RollerSki: '#2dd4bf', // light teal

  // Climbing - Stone/Slate family
  RockClimbing: '#64748b', // slate
  Climbing: '#475569', // dark slate

  // Other/Misc - Gray family
  Other: '#6b7280', // gray
  Wheelchair: '#9ca3af', // light gray
  Transition: '#d1d5db', // pale gray
}

// Weekday colors - Weekdays are cooler tones, weekends are warmer
export const weekdayColors: string[] = [
  '#fb923c', // Sunday - orange (weekend)
  '#4ade80', // Monday - green
  '#22d3ee', // Tuesday - cyan
  '#a78bfa', // Wednesday - purple
  '#fbbf24', // Thursday - amber
  '#38bdf8', // Friday - sky blue
  '#f472b6', // Saturday - pink (weekend)
]

// Map weekday labels to colors for DonutChart
export const weekdayColorMap: Record<string, string> = {
  Sun: weekdayColors[0],
  Sunday: weekdayColors[0],
  Mon: weekdayColors[1],
  Monday: weekdayColors[1],
  Tue: weekdayColors[2],
  Tuesday: weekdayColors[2],
  Wed: weekdayColors[3],
  Wednesday: weekdayColors[3],
  Thu: weekdayColors[4],
  Thursday: weekdayColors[4],
  Fri: weekdayColors[5],
  Friday: weekdayColors[5],
  Sat: weekdayColors[6],
  Saturday: weekdayColors[6],
}

// Time of day colors - representing the feel of each time period
export const daytimeColors: Record<string, string> = {
  Night: '#6366f1', // indigo - night sky
  Morning: '#fbbf24', // amber - sunrise
  Afternoon: '#fb923c', // orange - bright sun
  Evening: '#f472b6', // pink - sunset
}

// Year comparison colors - distinct colors for up to 10 years
// Uses a color sequence that remains distinguishable when overlaid
export const yearColors: string[] = [
  '#4ade80', // terminal green - current/latest year
  '#fb923c', // strava orange
  '#22d3ee', // terminal cyan
  '#a78bfa', // purple
  '#fbbf24', // amber
  '#f472b6', // pink
  '#38bdf8', // sky blue
  '#14b8a6', // teal
  '#f87171', // coral
  '#6366f1', // indigo
]

// Helper function to get year color by index or year value
export function getYearColor(yearOrIndex: number, years?: number[]): string {
  // If years array provided, find the index of the year
  if (years) {
    const index = years.indexOf(yearOrIndex)
    if (index !== -1) {
      return yearColors[index % yearColors.length]
    }
  }
  // Otherwise use the value directly as index
  return yearColors[yearOrIndex % yearColors.length]
}

// Helper function to get color for a value from a color map with fallback
export function getColorFromMap(
  value: string,
  colorMap: Record<string, string>,
  fallbackColors: string[] = Object.values(chartColors),
  index = 0
): string {
  return colorMap[value] ?? fallbackColors[index % fallbackColors.length]
}

// Helper function to group small slices into "Other" for cleaner charts
export interface ChartSlice {
  name: string
  value: number
  color?: string
}

export function groupSmallSlices(
  slices: ChartSlice[],
  maxCategories: number = maxChartCategories,
  colorMap?: Record<string, string>
): ChartSlice[] {
  if (slices.length <= maxCategories) {
    // Apply colors from map if provided
    if (colorMap) {
      return slices.map((s, idx) => ({
        ...s,
        color: getColorFromMap(s.name, colorMap, Object.values(chartColors), idx),
      }))
    }
    return slices
  }

  // Sort by value descending
  const sorted = [...slices].sort((a, b) => b.value - a.value)

  // Keep top (maxCategories - 1) and group the rest into "Other"
  const topSlices = sorted.slice(0, maxCategories - 1)
  const otherSlices = sorted.slice(maxCategories - 1)

  const otherTotal = otherSlices.reduce((sum, s) => sum + s.value, 0)

  const result: ChartSlice[] = topSlices.map((s, idx) => ({
    ...s,
    color: colorMap ? getColorFromMap(s.name, colorMap, Object.values(chartColors), idx) : s.color,
  }))

  if (otherTotal > 0) {
    result.push({
      name: 'Other',
      value: otherTotal,
      color: sportColors.Other,
    })
  }

  return result
}
