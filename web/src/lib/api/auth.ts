import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useDataProviderStatus } from '@/lib/data/context'
import type { AuthStatus } from './types'

export const authKeys = {
  status: ['auth', 'status'] as const,
}

export function useAuthStatus() {
  const { provider, initialized, error } = useDataProviderStatus()

  return useQuery({
    queryKey: authKeys.status,
    queryFn: async (): Promise<AuthStatus> => {
      if (!provider) throw new Error('Data provider not ready')
      return provider.getAuthStatus()
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
    retry: false,
    enabled: initialized && !error && !!provider,
  })
}

export function useRefreshToken() {
  const queryClient = useQueryClient()
  const { provider, initialized } = useDataProviderStatus()

  return useMutation({
    mutationFn: async () => {
      if (!provider || !initialized) throw new Error('Data provider not ready')
      return provider.refreshToken()
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.status })
    },
  })
}
