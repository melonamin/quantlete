/**
 * Hook to prevent accidental page refresh/close during active sync.
 * Shows browser's native "Leave site?" confirmation dialog.
 */

import { useEffect } from 'react'
import { useImportProgress } from '@/lib/api'
import { isWasmMode } from '@/lib/mode'

export function useSyncProtection() {
  const { data: progress } = useImportProgress()
  // Only show warning in WASM mode - server mode imports continue server-side
  const isRunning = isWasmMode() && progress?.status === 'running'

  useEffect(() => {
    if (!isRunning) return

    const handleBeforeUnload = (e: BeforeUnloadEvent) => {
      // Modern browsers ignore custom messages, but setting returnValue triggers the dialog
      e.preventDefault()
      e.returnValue = 'Sync is in progress. Are you sure you want to leave?'
      return e.returnValue
    }

    window.addEventListener('beforeunload', handleBeforeUnload)

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [isRunning])
}
