import { useImportProgress, useCancelImport } from '@/lib/api'
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

export function SyncProgressModal() {
  const { open, closeModal } = useSyncModalStore()
  const { data: progress } = useImportProgress()
  const cancelImport = useCancelImport()

  const isRunning = progress?.status === 'running'
  const isCompleted = progress?.status === 'completed'
  const isFailed = progress?.status === 'failed'
  const isCanceled = progress?.status === 'canceled'

  const handleCancel = () => {
    cancelImport.mutate()
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && closeModal()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Syncing with Strava</DialogTitle>
          <DialogDescription>
            {isRunning
              ? 'Importing your activities. You can dismiss this and continue browsing.'
              : isCompleted
                ? 'Sync completed successfully.'
                : isFailed
                  ? 'Sync encountered an error.'
                  : isCanceled
                    ? 'Sync was canceled.'
                    : 'Sync progress'}
          </DialogDescription>
        </DialogHeader>

        {progress && progress.status !== 'idle' && (
          <ImportStatus progress={progress} compact />
        )}

        <DialogFooter className="gap-2 sm:gap-0">
          {isRunning && (
            <Button
              variant="outline"
              onClick={handleCancel}
              disabled={cancelImport.isPending}
            >
              Cancel Sync
            </Button>
          )}
          <Button variant={isRunning ? 'secondary' : 'default'} onClick={closeModal}>
            {isRunning ? 'Dismiss' : 'Close'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
