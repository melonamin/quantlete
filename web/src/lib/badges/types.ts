export interface BadgeTheme {
  id: string
  name: string
  background: string
  foreground: string
  accent: string
  muted: string
}

export interface BadgeSize {
  id: string
  name: string
  width: number
  height: number
  titleSize: number
  valueSize: number
  subtitleSize: number
}

export type BadgeThemeId = 'terminal' | 'strava' | 'amber' | 'cyan' | 'minimal'
export type BadgeSizeId = 'compact' | 'standard' | 'wide'
export type BadgeBackgroundId = 'dark' | 'light' | 'transparent'

export interface BadgeBackground {
  id: BadgeBackgroundId
  name: string
  color: string | null // null = transparent
}
