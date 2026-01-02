/* eslint-disable react-refresh/only-export-components */
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { router } from '@/routes/router'
import { DataProviderWrapper } from '@/lib/data/context'
import { WelcomeModal } from '@/components/onboarding/welcome-modal'
import { SyncProgressModal } from '@/components/sync/sync-progress-modal'
import { useSyncProtection } from '@/hooks/use-sync-protection'
import { useDataEvents } from '@/hooks/use-data-events'
import { useTheme } from '@/hooks/use-theme'
import { registerServiceWorker } from '@/lib/pwa'
import './index.css'

// Global sync-related hooks
function SyncEffects() {
  useTheme()
  useSyncProtection() // Prevents accidental page refresh during sync
  useDataEvents() // Event-driven UI updates when data changes
  return null
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      refetchOnWindowFocus: false,
    },
  },
})

registerServiceWorker()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <DataProviderWrapper>
        <RouterProvider router={router} />
        <WelcomeModal />
        <SyncProgressModal />
        <SyncEffects />
      </DataProviderWrapper>
    </QueryClientProvider>
  </StrictMode>
)
