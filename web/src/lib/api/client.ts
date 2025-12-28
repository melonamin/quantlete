import type { ErrorResponse } from './types'

const API_BASE = '/api/v1'

export class ApiError extends Error {
  status: number
  body?: ErrorResponse

  constructor(message: string, status: number, body?: ErrorResponse) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.body = body
  }
}

async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const body = await response.json().catch(() => null)
    throw new ApiError(
      body?.error || `Request failed: ${response.status}`,
      response.status,
      body
    )
  }
  return response.json()
}

export async function get<T>(
  endpoint: string,
  params?: Record<string, string | number | boolean | undefined>
): Promise<T> {
  const url = new URL(`${API_BASE}${endpoint}`, window.location.origin)

  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        url.searchParams.set(key, String(value))
      }
    })
  }

  const response = await fetch(url.toString(), {
    method: 'GET',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
    },
  })

  return handleResponse<T>(response)
}

export async function post<T>(endpoint: string, data?: unknown): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: 'POST',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: data ? JSON.stringify(data) : undefined,
  })

  return handleResponse<T>(response)
}
