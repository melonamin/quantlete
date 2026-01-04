import { useImportProgress, useCancelImport, usePauseImport, useResumeImport } from '@/lib/api'
import { useSyncModalStore } from '@/stores'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { ImportStatus } from './import-status'
import { isWasmMode } from '@/lib/mode'

export function SyncProgressModal() {
  const { open, closeModal } = useSyncModalStore()
  const { data: progress } = useImportProgress()
  const cancelImport = useCancelImport()
  const pauseImport = usePauseImport()
  const resumeImport = useResumeImport()

  const isRunning = progress?.status === 'running'
  const isPaused = progress?.status === 'paused'
  const isCompleted = progress?.status === 'completed'
  const isFailed = progress?.status === 'failed'
  const isCanceled = progress?.status === 'canceled'

  // Pause/Resume is only available in WASM mode.
  // Server mode doesn't support true pause - only cancel with resumable state.
  const supportsPause = isWasmMode()

  const handleCancel = () => {
    cancelImport.mutate()
  }

  const handlePause = () => {
    pauseImport.mutate()
  }

  const handleResume = () => {
    resumeImport.mutate()
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && closeModal()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Syncing with Strava</DialogTitle>
          <DialogDescription>
            {isRunning
              ? 'Importing your activities. You can dismiss this and continue browsing.'
              : isPaused
                ? 'Sync is paused. Resume to continue importing.'
                : isCompleted
                  ? 'Sync completed successfully.'
                  : isFailed
                    ? 'Sync encountered an error.'
                    : isCanceled
                      ? 'Sync was canceled.'
                      : 'Sync progress'}
          </DialogDescription>
        </DialogHeader>

        {progress && progress.status !== 'idle' && <ImportStatus progress={progress} compact />}

        <DialogFooter className="gap-2 sm:gap-0">
          {isRunning && (
            <>
              {supportsPause && (
                <Button variant="outline" onClick={handlePause} disabled={pauseImport.isPending}>
                  Pause
                </Button>
              )}
              <Button variant="outline" onClick={handleCancel} disabled={cancelImport.isPending}>
                Cancel Sync
              </Button>
            </>
          )}
          {isPaused && supportsPause && (
            <>
              <Button variant="default" onClick={handleResume} disabled={resumeImport.isPending}>
                Resume
              </Button>
              <Button variant="outline" onClick={handleCancel} disabled={cancelImport.isPending}>
                Cancel Sync
              </Button>
            </>
          )}
          <Button variant={isRunning || isPaused ? 'secondary' : 'default'} onClick={closeModal}>
            {isRunning || isPaused ? 'Dismiss' : 'Close'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
