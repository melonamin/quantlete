import { forwardRef, type RefObject } from 'react'
import { Button } from '@/components/ui/button'
import { Download, Copy, Check } from 'lucide-react'
import { useState } from 'react'
import type { BadgeTheme, BadgeSize, BadgeSizeId, BadgeBackgroundId } from '@/lib/badges/types'
import { themes, sizes, backgrounds, themeIds, sizeIds, backgroundIds } from '@/lib/badges/themes'
import { features } from '@/lib/features'

// Format utilities matching Go backend (internal/badges/format.go)
const METERS_PER_KM = 1000.0
const METERS_PER_MILE = 1609.34
const METERS_PER_FOOT = 0.3048

function formatDistance(meters: number, unitSystem: 'metric' | 'imperial'): string {
  const m = meters ?? 0
  if (unitSystem === 'imperial') {
    const miles = m / METERS_PER_MILE
    return `${miles.toFixed(1)} mi`
  }
  const km = m / METERS_PER_KM
  return `${km.toFixed(1)} km`
}

function formatDuration(seconds: number): string {
  const s = seconds ?? 0
  const hours = Math.floor(s / 3600)
  const minutes = Math.floor((s % 3600) / 60)
  return `${hours}h ${minutes}m`
}

function formatElevation(meters: number, unitSystem: 'metric' | 'imperial'): string {
  const m = meters ?? 0
  if (unitSystem === 'imperial') {
    const feet = m / METERS_PER_FOOT
    return `${formatNumber(Math.round(feet))} ft`
  }
  return `${formatNumber(Math.round(m))} m`
}

function formatNumber(n: number): string {
  if (n === undefined || n === null || Number.isNaN(n)) return '0'
  return n.toLocaleString('en-US')
}

function formatElapsedTime(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = seconds % 60

  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(secs).padStart(2, '0')}`
  }
  return `${minutes}:${String(secs).padStart(2, '0')}`
}

// Badge SVG component - matches Go template in internal/badges/render.go
interface BadgeSVGProps {
  title: string
  value: string
  subtitle?: string
  theme: BadgeTheme
  size: BadgeSize
  background?: string | null // null = transparent
}

const BadgeSVG = forwardRef<SVGSVGElement, BadgeSVGProps>(
  ({ title, value, subtitle, theme, size, background }, ref) => {
    const padding = 20
    const titleY = padding + size.titleSize
    const valueY = size.height / 2 + size.valueSize / 3 + (subtitle ? 0 : 4)
    const subtitleY = valueY + size.subtitleSize + 8

    // Use provided background, or fall back to theme default
    const bgColor = background === undefined ? theme.background : background

    return (
      <svg
        ref={ref}
        width={size.width}
        height={size.height}
        viewBox={`0 0 ${size.width} ${size.height}`}
        xmlns="http://www.w3.org/2000/svg"
      >
        {bgColor && <rect width={size.width} height={size.height} fill={bgColor} rx={8} ry={8} />}
        <rect
          x={1}
          y={1}
          width={size.width - 2}
          height={size.height - 2}
          fill="none"
          stroke={theme.muted}
          strokeWidth={1}
          strokeOpacity={0.3}
          rx={7}
          ry={7}
        />
        <text
          x={20}
          y={titleY}
          fill={theme.muted}
          fontSize={size.titleSize}
          fontFamily="'JetBrains Mono', monospace"
          fontWeight={500}
          style={{ letterSpacing: '0.05em' }}
        >
          {title.toUpperCase()}
        </text>
        <text
          x={20}
          y={valueY}
          fill={theme.accent}
          fontSize={size.valueSize}
          fontFamily="'JetBrains Mono', monospace"
          fontWeight={700}
        >
          {value}
        </text>
        {subtitle && (
          <text
            x={20}
            y={subtitleY}
            fill={theme.muted}
            fontSize={size.subtitleSize}
            fontFamily="'JetBrains Mono', monospace"
            fontWeight={400}
          >
            {subtitle}
          </text>
        )}
      </svg>
    )
  }
)
BadgeSVG.displayName = 'BadgeSVG'

// Badge Preview wrapper with copy/download buttons
interface BadgePreviewProps {
  children: React.ReactNode
  filename: string
  svgRef: RefObject<SVGSVGElement | null>
  themeId: string
  sizeId: string
  backgroundId: string
  unitSystem: 'metric' | 'imperial'
}

export function BadgePreview({
  children,
  filename,
  svgRef,
  themeId,
  sizeId,
  backgroundId,
  unitSystem,
}: BadgePreviewProps) {
  const [copiedType, setCopiedType] = useState<'svg' | 'link' | 'markdown' | null>(null)

  // Build URL with query params
  const buildUrl = () => {
    const params = new URLSearchParams()
    if (themeId !== 'terminal') params.set('theme', themeId)
    if (sizeId !== 'standard') params.set('size', sizeId)
    if (backgroundId !== 'dark') params.set('bg', backgroundId)
    if (unitSystem === 'imperial') params.set('unit', 'imperial')
    const query = params.toString()
    return `${window.location.origin}/badges/${filename}.svg${query ? `?${query}` : ''}`
  }

  const badgeUrl = buildUrl()

  const downloadSVG = () => {
    if (!svgRef.current) return
    const svgData = new XMLSerializer().serializeToString(svgRef.current)
    const blob = new Blob([svgData], { type: 'image/svg+xml' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${filename}.svg`
    a.click()
    URL.revokeObjectURL(url)
  }

  const copyToClipboard = async (text: string, type: 'svg' | 'link' | 'markdown') => {
    try {
      await navigator.clipboard.writeText(text)
      setCopiedType(type)
      setTimeout(() => setCopiedType(null), 2000)
    } catch {
      const textarea = document.createElement('textarea')
      textarea.value = text
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      setCopiedType(type)
      setTimeout(() => setCopiedType(null), 2000)
    }
  }

  const copySVG = () => {
    if (!svgRef.current) return
    const svgData = new XMLSerializer().serializeToString(svgRef.current)
    copyToClipboard(svgData, 'svg')
  }

  const copyLink = () => copyToClipboard(badgeUrl, 'link')
  const copyMarkdown = () => copyToClipboard(`![${filename}](${badgeUrl})`, 'markdown')

  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-center rounded-lg border border-border bg-background/50 p-4">
        {children}
      </div>
      <div className="flex flex-wrap gap-2">
        <Button variant="outline" size="sm" onClick={downloadSVG}>
          <Download className="mr-1.5 h-3.5 w-3.5" />
          SVG
        </Button>
        <Button variant="outline" size="sm" onClick={copySVG}>
          {copiedType === 'svg' ? (
            <Check className="mr-1.5 h-3.5 w-3.5 text-green-500" />
          ) : (
            <Copy className="mr-1.5 h-3.5 w-3.5" />
          )}
          Copy
        </Button>
        {/* Link/Markdown buttons only available in server mode (requires backend to serve badge URLs) */}
        {features.apiAccess && (
          <>
            <Button variant="outline" size="sm" onClick={copyLink}>
              {copiedType === 'link' ? (
                <Check className="mr-1.5 h-3.5 w-3.5 text-green-500" />
              ) : (
                <Copy className="mr-1.5 h-3.5 w-3.5" />
              )}
              Link
            </Button>
            <Button variant="outline" size="sm" onClick={copyMarkdown}>
              {copiedType === 'markdown' ? (
                <Check className="mr-1.5 h-3.5 w-3.5 text-green-500" />
              ) : (
                <Copy className="mr-1.5 h-3.5 w-3.5" />
              )}
              Markdown
            </Button>
          </>
        )}
      </div>
    </div>
  )
}

// Theme/Size/Background customizer
interface BadgeCustomizerProps {
  selectedTheme: string
  selectedSize: BadgeSizeId
  selectedBackground: BadgeBackgroundId
  onThemeChange: (theme: string) => void
  onSizeChange: (size: BadgeSizeId) => void
  onBackgroundChange: (bg: BadgeBackgroundId) => void
}

export function BadgeCustomizer({
  selectedTheme,
  selectedSize,
  selectedBackground,
  onThemeChange,
  onSizeChange,
  onBackgroundChange,
}: BadgeCustomizerProps) {
  return (
    <div className="flex flex-wrap gap-4">
      <div className="space-y-2">
        <p className="text-xs font-medium text-muted-foreground">Accent</p>
        <div className="flex gap-2">
          {themeIds.map((id) => (
            <button
              key={id}
              onClick={() => onThemeChange(id)}
              className={`flex h-8 items-center gap-2 rounded-md border px-3 text-sm transition-colors ${
                selectedTheme === id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50'
              }`}
            >
              <span
                className="h-3 w-3 rounded-full"
                style={{ backgroundColor: themes[id].accent }}
              />
              {themes[id].name}
            </button>
          ))}
        </div>
      </div>
      <div className="space-y-2">
        <p className="text-xs font-medium text-muted-foreground">Background</p>
        <div className="flex gap-2">
          {backgroundIds.map((id) => (
            <button
              key={id}
              onClick={() => onBackgroundChange(id)}
              className={`flex h-8 items-center gap-2 rounded-md border px-3 text-sm transition-colors ${
                selectedBackground === id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50'
              }`}
            >
              <span
                className="h-3 w-3 rounded border border-border"
                style={{
                  backgroundColor: backgrounds[id].color ?? 'transparent',
                  backgroundImage:
                    id === 'transparent'
                      ? 'linear-gradient(45deg, #ccc 25%, transparent 25%), linear-gradient(-45deg, #ccc 25%, transparent 25%), linear-gradient(45deg, transparent 75%, #ccc 75%), linear-gradient(-45deg, transparent 75%, #ccc 75%)'
                      : undefined,
                  backgroundSize: id === 'transparent' ? '6px 6px' : undefined,
                  backgroundPosition:
                    id === 'transparent' ? '0 0, 0 3px, 3px -3px, -3px 0px' : undefined,
                }}
              />
              {backgrounds[id].name}
            </button>
          ))}
        </div>
      </div>
      <div className="space-y-2">
        <p className="text-xs font-medium text-muted-foreground">Size</p>
        <div className="flex gap-2">
          {sizeIds.map((id) => (
            <button
              key={id}
              onClick={() => onSizeChange(id)}
              className={`h-8 rounded-md border px-3 text-sm transition-colors ${
                selectedSize === id
                  ? 'border-primary bg-primary/10 text-primary'
                  : 'border-border hover:border-primary/50'
              }`}
            >
              {sizes[id].name}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

// Stat badge for distance/time/elevation/activities
interface StatBadgeProps {
  statType: 'distance' | 'time' | 'elevation' | 'activities'
  value: number
  theme: BadgeTheme
  size: BadgeSize
  background?: string | null
  unitSystem: 'metric' | 'imperial'
}

export const StatBadge = forwardRef<SVGSVGElement, StatBadgeProps>(
  ({ statType, value, theme, size, background, unitSystem }, ref) => {
    let title: string
    let formattedValue: string

    switch (statType) {
      case 'distance':
        title = 'Total Distance'
        formattedValue = formatDistance(value, unitSystem)
        break
      case 'time':
        title = 'Total Time'
        formattedValue = formatDuration(value)
        break
      case 'elevation':
        title = 'Total Elevation'
        formattedValue = formatElevation(value, unitSystem)
        break
      case 'activities':
        title = 'Activities'
        formattedValue = formatNumber(value)
        break
    }

    return (
      <BadgeSVG
        ref={ref}
        title={title}
        value={formattedValue}
        theme={theme}
        size={size}
        background={background}
      />
    )
  }
)
StatBadge.displayName = 'StatBadge'

// Eddington badge
interface EddingtonBadgeProps {
  number: number
  theme: BadgeTheme
  size: BadgeSize
  background?: string | null
  unitSystem: 'metric' | 'imperial'
}

export const EddingtonBadge = forwardRef<SVGSVGElement, EddingtonBadgeProps>(
  ({ number, theme, size, background, unitSystem }, ref) => {
    const unit = unitSystem === 'imperial' ? 'mi' : 'km'
    return (
      <BadgeSVG
        ref={ref}
        title="Eddington Number"
        value={`E${number}`}
        subtitle={`${number} days with ${number}+ ${unit}`}
        theme={theme}
        size={size}
        background={background}
      />
    )
  }
)
EddingtonBadge.displayName = 'EddingtonBadge'

// Yearly badge
interface YearlyBadgeProps {
  year: number
  activityCount: number
  totalDistance: number
  theme: BadgeTheme
  size: BadgeSize
  background?: string | null
  unitSystem: 'metric' | 'imperial'
}

export const YearlyBadge = forwardRef<SVGSVGElement, YearlyBadgeProps>(
  ({ year, activityCount, totalDistance, theme, size, background, unitSystem }, ref) => {
    return (
      <BadgeSVG
        ref={ref}
        title={`Year ${year}`}
        value={formatDistance(totalDistance, unitSystem)}
        subtitle={`${activityCount} activities`}
        theme={theme}
        size={size}
        background={background}
      />
    )
  }
)
YearlyBadge.displayName = 'YearlyBadge'

// Monthly badge
interface MonthlyBadgeProps {
  month: string // "2024-12"
  activityCount: number
  totalDistance: number
  theme: BadgeTheme
  size: BadgeSize
  background?: string | null
  unitSystem: 'metric' | 'imperial'
}

function formatMonthName(monthStr: string): string {
  const months = [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December',
  ]
  const [year, month] = monthStr.split('-')
  const monthNum = parseInt(month, 10)
  if (monthNum >= 1 && monthNum <= 12) {
    return `${months[monthNum - 1]} ${year}`
  }
  return monthStr
}

export const MonthlyBadge = forwardRef<SVGSVGElement, MonthlyBadgeProps>(
  ({ month, activityCount, totalDistance, theme, size, background, unitSystem }, ref) => {
    return (
      <BadgeSVG
        ref={ref}
        title={formatMonthName(month)}
        value={formatDistance(totalDistance, unitSystem)}
        subtitle={`${activityCount} activities`}
        theme={theme}
        size={size}
        background={background}
      />
    )
  }
)
MonthlyBadge.displayName = 'MonthlyBadge'

// PR badge
interface PRBadgeProps {
  distanceType: string // "5K", "10K", etc.
  elapsedTime: number // seconds
  date: string // ISO date
  theme: BadgeTheme
  size: BadgeSize
  background?: string | null
}

export const PRBadge = forwardRef<SVGSVGElement, PRBadgeProps>(
  ({ distanceType, elapsedTime, date, theme, size, background }, ref) => {
    const formattedDate = new Date(date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    })
    return (
      <BadgeSVG
        ref={ref}
        title={`${distanceType} PR`}
        value={formatElapsedTime(elapsedTime)}
        subtitle={formattedDate}
        theme={theme}
        size={size}
        background={background}
      />
    )
  }
)
PRBadge.displayName = 'PRBadge'
