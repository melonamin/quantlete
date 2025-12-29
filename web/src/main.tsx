import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { router } from '@/routes/router'
import { DataProviderWrapper } from '@/lib/data/context'
import { WelcomeModal } from '@/components/onboarding/welcome-modal'
import { SyncProgressModal } from '@/components/sync/sync-progress-modal'
import { useSyncProtection } from '@/hooks/use-sync-protection'
import './index.css'

// Prevents accidental page refresh/close during active sync
function SyncProtection() {
  useSyncProtection()
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

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <DataProviderWrapper>
        <RouterProvider router={router} />
        <WelcomeModal />
        <SyncProgressModal />
        <SyncProtection />
      </DataProviderWrapper>
    </QueryClientProvider>
  </StrictMode>
)
