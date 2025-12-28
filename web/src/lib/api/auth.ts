import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { get, post } from './client'
import type { AuthStatus } from './types'

export const authKeys = {
  status: ['auth', 'status'] as const,
}

export function useAuthStatus() {
  return useQuery({
    queryKey: authKeys.status,
    queryFn: () => get<AuthStatus>('/auth/status'),
    staleTime: 1000 * 60 * 5, // 5 minutes
    retry: false,
  })
}

export function useRefreshToken() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: () => post<{ success: boolean }>('/auth/refresh'),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: authKeys.status })
    },
  })
}
