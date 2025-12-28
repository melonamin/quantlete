import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { ApiError, get } from './client'

export interface Challenge {
  id: string
  name: string
  slug?: string
  badge_url?: string
  completion_date?: string
  month?: string
}

export const challengeKeys = {
  all: ['challenges'] as const,
  list: (month?: string) => [...challengeKeys.all, 'list', month] as const,
}

export function useChallenges(month?: string) {
  return useQuery({
    queryKey: challengeKeys.list(month),
    queryFn: () => get<Challenge[]>('/challenges', { month: month || undefined }),
  })
}

async function postMultipart<T>(endpoint: string, formData: FormData): Promise<T> {
  const response = await fetch(`/api/v1${endpoint}`, {
    method: 'POST',
    credentials: 'include',
    body: formData,
    headers: {
      Accept: 'application/json',
    },
  })
  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new ApiError(body?.error || `Request failed: ${response.status}`, response.status, body)
  }
  return response.json()
}

export function useImportChallenges() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async ({ file }: { file: File }) => {
      const form = new FormData()
      form.append('file', file)
      return postMultipart<{ imported: number }>('/challenges/import', form)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: challengeKeys.all })
    },
  })
}
