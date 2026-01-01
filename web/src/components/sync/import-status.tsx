import { useState, useEffect, useMemo } from 'react'
import type { ImportProgress, ImportPhase } from '@/lib/api/import'
import { Loader2, Check, X, AlertCircle, Timer } from 'lucide-react'

// Phase order for display
const PHASE_ORDER: ImportPhase[] = [
  'activities',
  'gear',
  'streams',
  'activity_details',
  'segment_details',
  'photos',
]

const PHASE_LABELS: Record<ImportPhase, string> = {
  idle: 'Idle',
  activities: 'Activities',
  gear: 'Gear',
  streams: 'Streams',
  activity_details: 'Details',
  segment_details: 'Segments',
  photos: 'Photos',
  completed: 'Completed',
}

function getPhaseProgress(
  phase: ImportPhase,
  progress: ImportProgress
): { done: number; total: number } {
  switch (phase) {
    case 'activities':
      return { done: progress.activities_done, total: progress.activities_total }
    case 'gear':
      return { done: progress.gear_done, total: progress.gear_total }
    case 'streams':
      return { done: progress.streams_done, total: progress.streams_total }
    case 'activity_details':
      return { done: progress.details_done, total: progress.details_total }
    case 'segment_details':
      return { done: progress.segments_done, total: progress.segments_total }
    case 'photos':
      return { done: progress.photos_done, total: progress.photos_total }
    default:
      return { done: 0, total: 0 }
  }
}

type PhaseStatus = 'pending' | 'running' | 'completed' | 'skipped'

function getPhaseStatus(
  phase: ImportPhase,
  currentPhase: ImportPhase,
  progress: ImportProgress
): PhaseStatus {
  const phaseIdx = PHASE_ORDER.indexOf(phase)
  const currentIdx = PHASE_ORDER.indexOf(currentPhase)

  if (currentPhase === 'completed') return 'completed'

  const { total } = getPhaseProgress(phase, progress)

  // If total is 0, this phase is skipped
  if (total === 0 && phaseIdx < currentIdx) return 'skipped'
  if (total === 0 && phase !== currentPhase) return 'pending'

  if (phaseIdx < currentIdx) return 'completed'
  if (phaseIdx === currentIdx) return 'running'
  return 'pending'
}

function PhaseProgressBar({
  phase,
  progress,
  isRunning,
}: {
  phase: ImportPhase
  progress: ImportProgress
  isRunning: boolean
}) {
  const { done, total } = getPhaseProgress(phase, progress)
  const status = getPhaseStatus(phase, progress.phase, progress)
  const percentage = total > 0 ? Math.round((done / total) * 100) : 0

  // Don't show phases with 0 total (skipped)
  if (total === 0 && status !== 'running') return null

  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-xs">
        <div className="flex items-center gap-1.5">
          {status === 'completed' && <Check className="h-3 w-3 text-green-600" />}
          {status === 'running' && isRunning && (
            <Loader2 className="h-3 w-3 animate-spin text-primary" />
          )}
          {status === 'pending' && <div className="h-3 w-3" />}
          <span className={status === 'running' ? 'font-medium text-foreground' : 'text-muted-foreground'}>
            {PHASE_LABELS[phase]}
          </span>
        </div>
        <span className="text-muted-foreground tabular-nums">
          {done}/{total}
        </span>
      </div>
      <div className="h-1.5 rounded-full bg-muted overflow-hidden">
        <div
          className={`h-full transition-all duration-300 ${
            status === 'completed'
              ? 'bg-green-600'
              : status === 'running'
                ? 'bg-primary'
                : 'bg-muted-foreground/30'
          }`}
          style={{ width: `${status === 'pending' ? 0 : percentage}%` }}
        />
      </div>
    </div>
  )
}

function RateLimitCountdown({ waitingUntil }: { waitingUntil: string }) {
  const targetTime = useMemo(() => new Date(waitingUntil).getTime(), [waitingUntil])
  const isValidDate = !Number.isNaN(targetTime)
  const [timeRemaining, setTimeRemaining] = useState('')

  useEffect(() => {
    if (!isValidDate) return

    const updateCountdown = () => {
      const now = Date.now()
      const diff = targetTime - now

      if (diff <= 0) {
        setTimeRemaining('Resuming...')
        return
      }

      const minutes = Math.floor(diff / 60000)
      const seconds = Math.floor((diff % 60000) / 1000)

      if (minutes > 0) {
        setTimeRemaining(`${minutes}m ${seconds}s`)
      } else {
        setTimeRemaining(`${seconds}s`)
      }
    }

    updateCountdown()
    const interval = setInterval(updateCountdown, 1000)
    return () => clearInterval(interval)
  }, [isValidDate, targetTime])

  if (!isValidDate) {
    return <span className="tabular-nums font-medium text-amber-600">Unknown</span>
  }

  return (
    <span className="tabular-nums font-medium text-amber-600">{timeRemaining}</span>
  )
}

interface ImportStatusProps {
  progress: ImportProgress
  compact?: boolean
}

export function ImportStatus({ progress, compact = false }: ImportStatusProps) {
  const isRunning = progress.status === 'running'
  const isWaiting = progress.waiting_for_rate_limit && progress.waiting_until

  return (
    <div className="rounded-lg border border-border bg-muted/50 p-4">
      {/* Rate limit waiting banner */}
      {isWaiting && (
        <div className="mb-3 flex items-center gap-2 rounded-md bg-amber-100 dark:bg-amber-950/50 border border-amber-200 dark:border-amber-800 px-3 py-2">
          <Timer className="h-4 w-4 text-amber-600 dark:text-amber-500" />
          <span className="text-sm text-amber-800 dark:text-amber-300">
            Rate limit reached. Resuming in{' '}
            <RateLimitCountdown waitingUntil={progress.waiting_until!} />
          </span>
        </div>
      )}

      <div className="flex items-center justify-between mb-3">
        <div className="flex items-center gap-2">
          {progress.status === 'running' && !isWaiting && (
            <Loader2 className="h-4 w-4 animate-spin text-primary" />
          )}
          {progress.status === 'running' && isWaiting && (
            <Timer className="h-4 w-4 text-amber-600" />
          )}
          {progress.status === 'completed' && <Check className="h-4 w-4 text-green-600" />}
          {progress.status === 'failed' && <AlertCircle className="h-4 w-4 text-destructive" />}
          {progress.status === 'canceled' && <X className="h-4 w-4 text-muted-foreground" />}
          <span className="font-medium capitalize">
            {isWaiting ? 'Waiting for rate limit' : progress.status}
          </span>
        </div>
        {isRunning && !isWaiting && progress.estimated_eta && (
          <span className="text-sm text-muted-foreground">ETA: {progress.estimated_eta}</span>
        )}
      </div>

      {(isRunning || progress.status === 'completed') && (
        <div className="space-y-2">
          {PHASE_ORDER.map((phase) => (
            <PhaseProgressBar
              key={phase}
              phase={phase}
              progress={progress}
              isRunning={isRunning}
            />
          ))}
        </div>
      )}

      {progress.failed_count > 0 && (
        <p className="text-sm text-destructive mt-3">Failed: {progress.failed_count}</p>
      )}

      {progress.error && <p className="text-sm text-destructive mt-2">{progress.error}</p>}

      {!compact && isRunning && progress.rate_limit_limit_15min > 0 && (
        <div className="mt-3 text-xs text-muted-foreground">
          Rate limit: {progress.rate_limit_used_15min}/{progress.rate_limit_limit_15min} (15min),{' '}
          {progress.rate_limit_used_daily}/{progress.rate_limit_limit_daily} (daily)
        </div>
      )}
    </div>
  )
}
