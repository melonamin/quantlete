import { useAuthStatus } from '@/lib/data'
import { isWasmMode } from '@/lib/mode'
import { getAuthUrl } from '@/lib/wasm/strava/client'
import { useOnboardingStore } from '@/stores'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Lock, HardDrive, BarChart3, Clock } from 'lucide-react'

/**
 * Check if we should show the onboarding modal.
 * Returns true if: WASM mode + not authenticated + not dismissed + not on OAuth callback page.
 */
export function shouldShowOnboarding(isAuthenticated: boolean, isLoading: boolean, dismissed: boolean): boolean {
  if (!isWasmMode()) return false
  if (isLoading) return false
  if (isAuthenticated) return false
  if (dismissed) return false
  // Don't show on OAuth callback page
  if (typeof window !== 'undefined' && window.location.pathname === '/oauth/callback') return false
  return true
}

export function WelcomeModal() {
  const { data: auth, isLoading } = useAuthStatus()
  const { dismissed, dismiss } = useOnboardingStore()

  // Determine if we should show the modal
  const isAuthenticated = auth?.authenticated ?? false
  const show = shouldShowOnboarding(isAuthenticated, isLoading, dismissed)

  const authUrl = getAuthUrl(`${window.location.origin}/oauth/callback`)

  return (
    <Dialog open={show} onOpenChange={() => {}}>
      <DialogContent showCloseButton={false} className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="text-2xl">Welcome to Stata</DialogTitle>
          <DialogDescription>Your private Strava analytics dashboard</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="flex gap-3">
            <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-green-100 dark:bg-green-950">
              <Lock className="h-4 w-4 text-green-600 dark:text-green-400" />
            </div>
            <div>
              <p className="font-medium">100% Private</p>
              <p className="text-sm text-muted-foreground">
                All data stays in your browser. Nothing is sent to any server.
              </p>
            </div>
          </div>

          <div className="flex gap-3">
            <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-blue-100 dark:bg-blue-950">
              <HardDrive className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <p className="font-medium">Stored Locally</p>
              <p className="text-sm text-muted-foreground">
                Data is tied to this device and browser using local storage.
              </p>
            </div>
          </div>

          <div className="flex gap-3">
            <div className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-purple-100 dark:bg-purple-950">
              <BarChart3 className="h-4 w-4 text-purple-600 dark:text-purple-400" />
            </div>
            <div>
              <p className="font-medium">Deep Analytics</p>
              <p className="text-sm text-muted-foreground">
                Training load, power curves, segments, and more.
              </p>
            </div>
          </div>

          <div className="rounded-lg border border-border bg-muted/50 p-3">
            <div className="flex gap-2">
              <Clock className="h-4 w-4 shrink-0 text-muted-foreground mt-0.5" />
              <div className="text-sm text-muted-foreground">
                <p className="font-medium text-foreground">First sync takes time</p>
                <p>
                  Due to Strava rate limits, importing all your activities may take a while. The
                  dashboard becomes usable after activities load, and detailed data (streams, photos)
                  loads in the background.
                </p>
              </div>
            </div>
          </div>
        </div>

        <DialogFooter className="flex-col gap-2 sm:flex-col">
          <Button asChild className="w-full bg-[#FC4C02] hover:bg-[#FC4C02]/90">
            <a href={authUrl}>Connect with Strava</a>
          </Button>
          <Button variant="ghost" className="w-full" onClick={dismiss}>
            Maybe later
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
