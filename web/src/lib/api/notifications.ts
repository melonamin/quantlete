import { useMutation } from '@tanstack/react-query'

interface TestNotificationParams {
  serviceId: string
}

interface TestResult {
  success: boolean
  error?: string
}

export function useTestNotification() {
  return useMutation({
    mutationFn: async (params: TestNotificationParams): Promise<TestResult> => {
      const response = await fetch('/api/v1/notifications/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(params),
      })
      return response.json()
    },
  })
}

export function useTestAllNotifications() {
  return useMutation({
    mutationFn: async (): Promise<TestResult[]> => {
      const response = await fetch('/api/v1/notifications/test-all', {
        method: 'POST',
      })
      return response.json()
    },
  })
}
