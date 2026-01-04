import { isDemoMode } from '@/lib/mode'
import { Sparkles, ExternalLink } from 'lucide-react'

export function DemoBanner() {
  if (!isDemoMode()) {
    return null
  }

  return (
    <div className="fixed top-0 left-0 right-0 z-50 flex items-center justify-center gap-2 bg-gradient-to-r from-purple-600 to-blue-600 px-4 py-1.5 text-sm text-white">
      <Sparkles className="h-4 w-4" />
      <span>
        Demo mode with sample data.{' '}
        <a
          href="https://quantlete.fit"
          className="inline-flex items-center gap-1 font-medium underline underline-offset-2 hover:no-underline"
          target="_blank"
          rel="noopener noreferrer"
        >
          Get Quantlete
          <ExternalLink className="h-3 w-3" />
        </a>
      </span>
      <span className="hidden sm:inline text-white/70">|</span>
      <span className="hidden sm:inline text-white/70">Data resets on reload</span>
    </div>
  )
}

export const DEMO_BANNER_HEIGHT = 36
