/**
 * Sync History Modal - Shows past sync runs with their status and stats.
 * When a sync is running, shows live progress at the top.
 */

import { useSyncHistory, useImportProgress, useCancelImport } from '@/lib/api'
import { formatDistance } from 'date-fns'
import { Clock, Check, X, AlertCircle, XCircle, ChevronDown, ChevronRight } from 'lucide-react'
import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { ImportStatus } from './import-status'

interface SyncHistoryModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SyncHistoryModal({ open, onOpenChange }: SyncHistoryModalProps) {
  const { data: history, isLoading } = useSyncHistory(20)
  const { data: progress } = useImportProgress()
  const cancelImport = useCancelImport()
  const [expandedId, setExpandedId] = useState<number | null>(null)

  const isRunning = progress?.status === 'running'

  const formatDuration = (seconds?: number) => {
    if (!seconds) return '–'
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    if (minutes < 60) return `${minutes}m ${remainingSeconds}s`
    const hours = Math.floor(minutes / 60)
    const remainingMinutes = minutes % 60
    return `${hours}h ${remainingMinutes}m`
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <Check className="h-4 w-4 text-green-600" />
      case 'failed':
        return <X className="h-4 w-4 text-red-600" />
      case 'canceled':
        return <XCircle className="h-4 w-4 text-yellow-600" />
      case 'running':
        return <Clock className="h-4 w-4 text-blue-600 animate-pulse" />
      default:
        return <AlertCircle className="h-4 w-4 text-muted-foreground" />
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'completed':
        return 'Completed'
      case 'failed':
        return 'Failed'
      case 'canceled':
        return 'Canceled'
      case 'running':
        return 'Running'
      default:
        return status
    }
  }

  // Filter out the running entry from history when showing live progress
  const filteredHistory = isRunning
    ? history?.filter((run) => run.status !== 'running')
    : history

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[80vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle>{isRunning ? 'Sync Progress' : 'Sync History'}</DialogTitle>
        </DialogHeader>

        <div className="overflow-y-auto flex-1 -mx-6 px-6">
          {/* Live progress when sync is running */}
          {isRunning && progress && (
            <div className="mb-4">
              <ImportStatus progress={progress} />
              <div className="flex justify-end mt-3">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => cancelImport.mutate()}
                  disabled={cancelImport.isPending}
                >
                  Cancel Sync
                </Button>
              </div>
            </div>
          )}

          {/* History list */}
          {isLoading ? (
            <div className="py-8 text-center text-muted-foreground">Loading...</div>
          ) : !filteredHistory || filteredHistory.length === 0 ? (
            !isRunning && (
              <div className="py-8 text-center text-muted-foreground">No sync history yet</div>
            )
          ) : (
            <div className="space-y-2">
              {isRunning && filteredHistory.length > 0 && (
                <div className="text-xs text-muted-foreground font-medium pt-2 pb-1">
                  Previous syncs
                </div>
              )}
              {filteredHistory.map((run) => (
                <div
                  key={run.id}
                  className="rounded-lg border border-border overflow-hidden"
                >
                  <button
                    type="button"
                    className="w-full flex items-center justify-between p-3 text-left hover:bg-muted/50 transition-colors"
                    onClick={() => setExpandedId(expandedId === run.id ? null : run.id)}
                  >
                    <div className="flex items-center gap-3">
                      {getStatusIcon(run.status)}
                      <div>
                        <div className="font-medium text-sm">
                          {getStatusText(run.status)}
                          {run.full_sync && (
                            <span className="ml-2 text-xs px-1.5 py-0.5 rounded bg-muted text-muted-foreground">
                              Full
                            </span>
                          )}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {formatDistance(new Date(run.started_at), new Date(), { addSuffix: true })}
                          {run.duration_seconds && ` • ${formatDuration(run.duration_seconds)}`}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="text-xs text-muted-foreground text-right">
                        {run.activities_imported > 0 && (
                          <span>{run.activities_imported} activities</span>
                        )}
                      </div>
                      {expandedId === run.id ? (
                        <ChevronDown className="h-4 w-4 text-muted-foreground" />
                      ) : (
                        <ChevronRight className="h-4 w-4 text-muted-foreground" />
                      )}
                    </div>
                  </button>

                  {expandedId === run.id && (
                    <div className="border-t border-border bg-muted/30 p-3 space-y-2 text-sm">
                      <div className="grid grid-cols-2 gap-2">
                        <div>
                          <span className="text-muted-foreground">Activities:</span>{' '}
                          {run.activities_imported} imported, {run.activities_skipped} skipped
                        </div>
                        <div>
                          <span className="text-muted-foreground">Streams:</span>{' '}
                          {run.streams_imported}
                        </div>
                        {run.failed_count > 0 && (
                          <div className="text-red-600">
                            <span className="text-muted-foreground">Failed:</span>{' '}
                            {run.failed_count}
                          </div>
                        )}
                        <div>
                          <span className="text-muted-foreground">Total:</span>{' '}
                          {run.activities_total}
                        </div>
                      </div>

                      {run.error && (
                        <div className="text-red-600 text-xs mt-2 p-2 rounded bg-red-50 dark:bg-red-950">
                          {run.error}
                        </div>
                      )}

                      <div className="text-xs text-muted-foreground pt-2 border-t border-border">
                        Started: {new Date(run.started_at).toLocaleString()}
                        {run.completed_at && (
                          <>
                            <br />
                            Completed: {new Date(run.completed_at).toLocaleString()}
                          </>
                        )}
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="flex justify-end pt-4 border-t border-border -mx-6 px-6">
          <Button variant="outline" onClick={() => onOpenChange(false)} autoFocus={false}>
            Close
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
