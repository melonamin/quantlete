import type { ReactNode } from 'react'
import { Sidebar, SIDEBAR_COLLAPSED_WIDTH, SIDEBAR_EXPANDED_WIDTH } from './sidebar'
import { useSidebarStore } from '@/stores/sidebar'
import { useMediaQuery } from '@/lib/hooks'
import { SyncRefresh } from '@/components/sync/sync-refresh'

interface RootLayoutProps {
  children: ReactNode
}

export function RootLayout({ children }: RootLayoutProps) {
  const { collapsed } = useSidebarStore()
  const isDesktop = useMediaQuery('(min-width: 768px)')

  const marginLeft = isDesktop ? (collapsed ? SIDEBAR_COLLAPSED_WIDTH : SIDEBAR_EXPANDED_WIDTH) : 0

  return (
    <div className="min-h-screen bg-background">
      <SyncRefresh />
      <Sidebar />

      {/* Main content area */}
      <main
        className="min-h-screen transition-[margin] duration-200 ease-in-out"
        style={{ marginLeft }}
      >
        {/* Mobile header */}
        <header className="sticky top-0 z-30 flex h-14 items-center border-b border-border bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 px-4 md:hidden">
          <div className="ml-12 flex items-center">
            <span className="font-semibold">Stata</span>
          </div>
        </header>

        <div className="flex-1">{children}</div>
      </main>
    </div>
  )
}
