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
export const zoneLabels = [
  'Z1 Recovery',
  'Z2 Endurance',
  'Z3 Tempo',
  'Z4 Threshold',
  'Z5 VO2max',
]

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
