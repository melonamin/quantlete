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

// API response envelope - matches shared.Response in Go
interface ApiEnvelope<T> {
  ok: boolean
  data?: T
  error?: string
  message?: string
}

// Check if a response is an API envelope
function isEnvelope<T>(body: unknown): body is ApiEnvelope<T> {
  return (
    typeof body === 'object' &&
    body !== null &&
    'ok' in body &&
    typeof (body as ApiEnvelope<T>).ok === 'boolean'
  )
}

async function handleResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null)

  // Handle HTTP errors first
  if (!response.ok) {
    // Try to extract error from envelope or plain error response
    const errorMsg = body?.error || `Request failed: ${response.status}`
    throw new ApiError(errorMsg, response.status, body)
  }

  // Unwrap envelope responses: {ok: true, data: T} -> T
  if (isEnvelope<T>(body)) {
    if (!body.ok) {
      throw new ApiError(body.error || 'Unknown error', response.status, {
        error: body.error || '',
      })
    }
    // Return data from envelope, or the whole response if data is undefined
    // Some endpoints return {ok: true, message: "..."} without data
    return (body.data !== undefined ? body.data : body) as T
  }

  // Legacy endpoints may return data directly without envelope
  return body as T
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

export async function put<T>(endpoint: string, data?: unknown): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: 'PUT',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: data ? JSON.stringify(data) : undefined,
  })

  return handleResponse<T>(response)
}

export async function del<T>(endpoint: string): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    method: 'DELETE',
    credentials: 'include',
    headers: {
      Accept: 'application/json',
    },
  })

  return handleResponse<T>(response)
}
