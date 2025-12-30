import { useRef, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import {
  BadgePreview,
  BadgeCustomizer,
  StatBadge,
  EddingtonBadge,
  YearlyBadge,
  MonthlyBadge,
  PRBadge,
} from '@/components/badges'
import { getTheme, getSize, getBackground } from '@/lib/badges/themes'
import type { BadgeSizeId, BadgeBackgroundId } from '@/lib/badges/types'
import { useSettingsStore } from '@/stores/settings'
import {
  useDashboardStats,
  useEddington,
  useYearlyStats,
  useMonthlyStats,
  useBestEffortPRs,
  useAppSettings,
  useUpdateAppSettings,
} from '@/lib/data/hooks'
import { features } from '@/lib/features'

export function BadgesPage() {
  const unitSystem = useSettingsStore((s) => s.unitSystem)
  const [themeId, setThemeId] = useState('terminal')
  const [sizeId, setSizeId] = useState<BadgeSizeId>('standard')
  const [backgroundId, setBackgroundId] = useState<BadgeBackgroundId>('dark')

  const theme = getTheme(themeId)
  const size = getSize(sizeId)
  const background = getBackground(backgroundId)

  // Data hooks
  const { data: settings } = useAppSettings()
  const updateSettings = useUpdateAppSettings()
  const { data: stats, isLoading: statsLoading } = useDashboardStats()
  const { data: eddington, isLoading: eddingtonLoading } = useEddington()
  const { data: yearlyStats, isLoading: yearlyLoading } = useYearlyStats()
  const { data: monthlyStats, isLoading: monthlyLoading } = useMonthlyStats()
  const { data: prs, isLoading: prsLoading } = useBestEffortPRs()

  const publicBadgesEnabled = settings?.enable_public_badges ?? false

  const handleTogglePublicBadges = () => {
    if (!settings) return
    updateSettings.mutate({
      ...settings,
      enable_public_badges: !publicBadgesEnabled,
    })
  }

  // Refs for each badge type
  const distanceRef = useRef<SVGSVGElement>(null)
  const timeRef = useRef<SVGSVGElement>(null)
  const elevationRef = useRef<SVGSVGElement>(null)
  const activitiesRef = useRef<SVGSVGElement>(null)
  const eddingtonRef = useRef<SVGSVGElement>(null)
  const yearlyRef = useRef<SVGSVGElement>(null)
  const monthlyRef = useRef<SVGSVGElement>(null)
  const prRef = useRef<SVGSVGElement>(null)

  // Get most recent year and month
  const currentYear = yearlyStats?.[0]
  const currentMonth = monthlyStats?.[0]
  const topPR = prs?.[0]

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Badges</h1>
        <p className="text-muted-foreground">
          Generate SVG badges from your stats
        </p>
      </div>

      {/* Public Badges Toggle - only shown in server mode (requires backend) */}
      {features.apiAccess && (
        <Card className="mb-6">
          <CardHeader>
            <CardTitle>Public Access</CardTitle>
            <CardDescription>
              Allow badge images to be accessed without authentication
            </CardDescription>
          </CardHeader>
          <CardContent>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="checkbox"
                checked={publicBadgesEnabled}
                onChange={handleTogglePublicBadges}
                disabled={!settings || updateSettings.isPending}
                className="h-4 w-4 rounded border-border"
              />
              <div>
                <p className="font-medium">Enable public badge URLs</p>
                <p className="text-sm text-muted-foreground">
                  When enabled, badge images can be embedded in external sites like GitHub READMEs
                </p>
              </div>
            </label>
          </CardContent>
        </Card>
      )}

      {/* Customizer */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>Style</CardTitle>
        </CardHeader>
        <CardContent>
          <BadgeCustomizer
            selectedTheme={themeId}
            selectedSize={sizeId}
            selectedBackground={backgroundId}
            onThemeChange={setThemeId}
            onSizeChange={setSizeId}
            onBackgroundChange={setBackgroundId}
          />
        </CardContent>
      </Card>

      <div className="space-y-6">
        {/* Overall Stats */}
        <Card>
          <CardHeader>
            <CardTitle>Overall Stats</CardTitle>
          </CardHeader>
          <CardContent>
            {statsLoading ? (
              <BadgeGridSkeleton count={4} />
            ) : stats ? (
              <div className="grid gap-6 md:grid-cols-2">
                <BadgePreview filename="distance" svgRef={distanceRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <StatBadge
                    ref={distanceRef}
                    statType="distance"
                    value={stats.total_distance}
                    theme={theme}
                    size={size}
                    background={background.color}
                    unitSystem={unitSystem}
                  />
                </BadgePreview>
                <BadgePreview filename="time" svgRef={timeRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <StatBadge
                    ref={timeRef}
                    statType="time"
                    value={stats.total_moving_time}
                    theme={theme}
                    size={size}
                    background={background.color}
                    unitSystem={unitSystem}
                  />
                </BadgePreview>
                <BadgePreview filename="elevation" svgRef={elevationRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <StatBadge
                    ref={elevationRef}
                    statType="elevation"
                    value={stats.total_elevation_gain}
                    theme={theme}
                    size={size}
                    background={background.color}
                    unitSystem={unitSystem}
                  />
                </BadgePreview>
                <BadgePreview filename="activities" svgRef={activitiesRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <StatBadge
                    ref={activitiesRef}
                    statType="activities"
                    value={stats.total_activities}
                    theme={theme}
                    size={size}
                    background={background.color}
                    unitSystem={unitSystem}
                  />
                </BadgePreview>
              </div>
            ) : (
              <EmptyState message="No stats data available" />
            )}
          </CardContent>
        </Card>

        {/* Achievements */}
        <Card>
          <CardHeader>
            <CardTitle>Achievements</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2">
              {/* Eddington */}
              {eddingtonLoading ? (
                <BadgeSkeleton />
              ) : eddington ? (
                <BadgePreview filename="eddington" svgRef={eddingtonRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <EddingtonBadge
                    ref={eddingtonRef}
                    number={eddington.number}
                    theme={theme}
                    size={size}
                    background={background.color}
                    unitSystem={unitSystem}
                  />
                </BadgePreview>
              ) : null}

              {/* Top PR */}
              {prsLoading ? (
                <BadgeSkeleton />
              ) : topPR ? (
                <BadgePreview filename={`pr-${topPR.distance_type}`} svgRef={prRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <PRBadge
                    ref={prRef}
                    distanceType={topPR.name}
                    elapsedTime={topPR.elapsed_time_s}
                    date={topPR.start_date_local}
                    theme={theme}
                    size={size}
                    background={background.color}
                  />
                </BadgePreview>
              ) : null}
            </div>
            {!eddingtonLoading && !eddington && !prsLoading && !topPR && (
              <EmptyState message="No achievement data available" />
            )}
          </CardContent>
        </Card>

        {/* Time Periods */}
        <Card>
          <CardHeader>
            <CardTitle>Time Periods</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-6 md:grid-cols-2">
              {/* Current Year */}
              {yearlyLoading ? (
                <BadgeSkeleton />
              ) : currentYear ? (
                <BadgePreview filename={`year-${currentYear.year}`} svgRef={yearlyRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <YearlyBadge
                    ref={yearlyRef}
                    year={currentYear.year}
                    activityCount={currentYear.activity_count}
                    totalDistance={currentYear.total_distance}
                    theme={theme}
                    size={size}
                    background={background.color}
                    unitSystem={unitSystem}
                  />
                </BadgePreview>
              ) : null}

              {/* Current Month */}
              {monthlyLoading ? (
                <BadgeSkeleton />
              ) : currentMonth ? (
                <BadgePreview filename={`month-${currentMonth.month}`} svgRef={monthlyRef} themeId={themeId} sizeId={sizeId} backgroundId={backgroundId} unitSystem={unitSystem}>
                  <MonthlyBadge
                    ref={monthlyRef}
                    month={currentMonth.month}
                    activityCount={currentMonth.activity_count}
                    totalDistance={currentMonth.total_distance}
                    theme={theme}
                    size={size}
                    background={background.color}
                    unitSystem={unitSystem}
                  />
                </BadgePreview>
              ) : null}
            </div>
            {!yearlyLoading && !currentYear && !monthlyLoading && !currentMonth && (
              <EmptyState message="No time period data available" />
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

function BadgeSkeleton() {
  return (
    <div className="flex flex-col gap-3">
      <div className="flex justify-center rounded-lg border border-border bg-background/50 p-4">
        <Skeleton className="h-[120px] w-[400px]" />
      </div>
      <div className="flex gap-2">
        <Skeleton className="h-8 w-16" />
        <Skeleton className="h-8 w-16" />
        <Skeleton className="h-8 w-20" />
        <Skeleton className="h-8 w-20" />
      </div>
    </div>
  )
}

function BadgeGridSkeleton({ count }: { count: number }) {
  return (
    <div className="grid gap-6 md:grid-cols-2">
      {Array.from({ length: count }).map((_, i) => (
        <BadgeSkeleton key={i} />
      ))}
    </div>
  )
}

function EmptyState({ message }: { message: string }) {
  return (
    <div className="rounded-lg border border-border bg-card p-8 text-center">
      <p className="text-muted-foreground">{message}</p>
    </div>
  )
}
