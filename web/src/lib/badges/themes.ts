import type { BadgeTheme, BadgeSize, BadgeThemeId, BadgeSizeId, BadgeBackgroundId, BadgeBackground } from './types'

// Themes match the Go backend definitions in internal/badges/themes.go
export const themes: Record<BadgeThemeId, BadgeTheme> = {
  terminal: {
    id: 'terminal',
    name: 'Terminal',
    background: '#1a1a1f',
    foreground: '#e5e5e5',
    accent: '#22c55e',
    muted: '#a1a1aa',
  },
  strava: {
    id: 'strava',
    name: 'Strava',
    background: '#1a1a1f',
    foreground: '#e5e5e5',
    accent: '#fc4c02',
    muted: '#a1a1aa',
  },
  amber: {
    id: 'amber',
    name: 'Amber',
    background: '#1a1a1f',
    foreground: '#e5e5e5',
    accent: '#f59e0b',
    muted: '#a1a1aa',
  },
  cyan: {
    id: 'cyan',
    name: 'Cyan',
    background: '#1a1a1f',
    foreground: '#e5e5e5',
    accent: '#06b6d4',
    muted: '#a1a1aa',
  },
  minimal: {
    id: 'minimal',
    name: 'Minimal',
    background: '#ffffff',
    foreground: '#1a1a1f',
    accent: '#1a1a1f',
    muted: '#71717a',
  },
}

// Sizes match the Go backend definitions in internal/badges/themes.go
export const sizes: Record<BadgeSizeId, BadgeSize> = {
  compact: {
    id: 'compact',
    name: 'Compact',
    width: 200,
    height: 80,
    titleSize: 10,
    valueSize: 24,
    subtitleSize: 10,
  },
  standard: {
    id: 'standard',
    name: 'Standard',
    width: 400,
    height: 120,
    titleSize: 12,
    valueSize: 36,
    subtitleSize: 12,
  },
  wide: {
    id: 'wide',
    name: 'Wide',
    width: 500,
    height: 100,
    titleSize: 12,
    valueSize: 32,
    subtitleSize: 12,
  },
}

// Backgrounds for badge rendering
export const backgrounds: Record<BadgeBackgroundId, BadgeBackground> = {
  dark: {
    id: 'dark',
    name: 'Dark',
    color: '#1a1a1f',
  },
  light: {
    id: 'light',
    name: 'Light',
    color: '#ffffff',
  },
  transparent: {
    id: 'transparent',
    name: 'Transparent',
    color: null,
  },
}

export const themeIds = Object.keys(themes) as BadgeThemeId[]
export const sizeIds = Object.keys(sizes) as BadgeSizeId[]
export const backgroundIds = Object.keys(backgrounds) as BadgeBackgroundId[]

export function getTheme(id: string): BadgeTheme {
  if (id in themes) {
    return themes[id as BadgeThemeId]
  }
  return themes.terminal
}

export function getSize(id: string): BadgeSize {
  if (id in sizes) {
    return sizes[id as BadgeSizeId]
  }
  return sizes.standard
}

export function getBackground(id: string): BadgeBackground {
  if (id in backgrounds) {
    return backgrounds[id as BadgeBackgroundId]
  }
  return backgrounds.dark
}
