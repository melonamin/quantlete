import { WidgetWrapper } from './widget-wrapper'
import { useInsights } from '@/lib/data/hooks'
import { BatteryLow, Heart, TrendingUp, Lightbulb } from 'lucide-react'
import { uiColors } from '@/components/charts'
import type { Insight, InsightType, InsightSeverity } from '@/lib/api/stats'

// Valid insight types and severities for runtime validation.
const VALID_INSIGHT_TYPES: readonly InsightType[] = ['fatigue', 'recovery', 'fitness']
const VALID_SEVERITIES: readonly InsightSeverity[] = ['info', 'warning', 'success']

// Type guards for runtime validation of backend data.
function isValidInsightType(type: string): type is InsightType {
  return VALID_INSIGHT_TYPES.includes(type as InsightType)
}

function isValidInsightSeverity(severity: string): severity is InsightSeverity {
  return VALID_SEVERITIES.includes(severity as InsightSeverity)
}

// Known insight types with icons.
const iconMap: Record<InsightType, React.ComponentType<{ className?: string; style?: React.CSSProperties }>> = {
  fatigue: BatteryLow,
  recovery: Heart,
  fitness: TrendingUp,
}

// Known severity levels - using uiColors for consistency.
const colorMap: Record<InsightSeverity, string> = {
  warning: uiColors.warning,
  success: uiColors.success,
  info: uiColors.muted,
}

/** Normalized insight with guaranteed valid type and severity. */
interface NormalizedInsight extends Insight {
  normalizedType: InsightType
  normalizedSeverity: InsightSeverity
}

/**
 * Normalizes an insight from the backend, providing fallbacks for unknown types/severities.
 * Uses type guards for proper TypeScript narrowing while protecting against unexpected values.
 */
function normalizeInsight(insight: Insight): NormalizedInsight {
  return {
    ...insight,
    normalizedType: isValidInsightType(insight.type) ? insight.type : 'fitness',
    normalizedSeverity: isValidInsightSeverity(insight.severity) ? insight.severity : 'info',
  }
}

export function SmartCoach() {
  const { data, isLoading, isError } = useInsights()
  const insights = (data?.insights ?? []).map(normalizeInsight)

  // Show empty state for both error and no-data scenarios
  const showEmptyState = insights.length === 0 || isError

  return (
    <WidgetWrapper title="Smart Coach" isLoading={isLoading}>
      <div className="h-full flex flex-col">
        {showEmptyState ? (
          <div className="flex-1 flex items-center justify-center text-center">
            <div>
              <Lightbulb className="h-8 w-8 mx-auto mb-2 text-muted-foreground/50" />
              <p className="text-sm text-muted-foreground">
                {isLoading
                  ? 'Analyzing your training load...'
                  : isError
                    ? 'Unable to load insights. Try refreshing the page.'
                    : 'No training insights available yet. Record activities with power or heart rate data to get coaching tips.'}
              </p>
            </div>
          </div>
        ) : (
          <div className="flex-1 space-y-2 overflow-auto">
            {insights.map((insight) => {
              const Icon = iconMap[insight.normalizedType]
              const color = colorMap[insight.normalizedSeverity]
              return (
                <div key={insight.id} className="flex items-start gap-3 p-2 rounded-sm bg-muted/30">
                  <Icon className="h-4 w-4 mt-0.5 shrink-0" style={{ color }} />
                  <div className="min-w-0">
                    <div className="font-medium text-sm">{insight.title}</div>
                    <p className="text-xs text-muted-foreground">{insight.description}</p>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </div>
    </WidgetWrapper>
  )
}
